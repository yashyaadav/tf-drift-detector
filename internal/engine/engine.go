// Package engine compares desired (Terraform state) against actual (live cloud)
// resources and classifies each as managed, unmanaged, or missing by
// set-membership over the canonical identity key. Attribute-level "changed"
// detection is reserved for --deep (a later milestone).
//
// Joins are scoped by (partition, account, region) via resource.Key(). In M0 a
// scan targets a single region, and desired regional resources are stamped with
// that region; global types (IAM, S3) use resource.GlobalRegion on both sides.
//
// See docs/adr/0006-set-membership-default-deep-opt-in.md.
package engine

import (
	"math"
	"sort"

	"github.com/yashyaadav/tf-drift-detector/internal/registry"
	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

// Options carries non-data inputs to Compare so the engine stays deterministic
// and testable (no clock or environment access).
type Options struct {
	GeneratedAt  string
	ProviderName string
	// Suppress reports whether an otherwise-unmanaged live resource is
	// AWS-managed noise to be excluded from findings and the coverage
	// denominator. Typically normalize.IsAWSManaged.
	Suppress func(resource.Resource) bool
}

// Compare classifies desired vs actual and returns a report.
func Compare(
	desired, actual []resource.Resource,
	reg *registry.Registry,
	scope provider.ScanScope,
	statuses []provider.ScanStatus,
	opts Options,
) schema.Report {
	scanned := scope.Types
	if len(scanned) == 0 {
		scanned = reg.Types()
	}
	scannedSet := make(map[string]bool, len(scanned))
	for _, t := range scanned {
		scannedSet[t] = true
	}

	regionForDesired := ""
	if len(scope.Regions) > 0 {
		regionForDesired = scope.Regions[0]
	}

	// Key the desired side, restricting to types actually being scanned (we can
	// only verify what we enumerate) and normalizing identity + scope.
	desiredByKey := make(map[string]resource.Resource)
	for _, r := range desired {
		if !scannedSet[r.Type] {
			continue
		}
		r.Source = resource.SourceDesired
		r.ID = reg.CanonicalID(r.Type, r.ID)
		r.Partition = scope.Partition
		r.Account = scope.Account
		if reg.Global(r.Type) {
			r.Region = resource.GlobalRegion
		} else {
			r.Region = regionForDesired
		}
		desiredByKey[r.Key()] = r
	}

	actualByKey := make(map[string]resource.Resource, len(actual))
	for _, r := range actual {
		r.Source = resource.SourceActual
		actualByKey[r.Key()] = r
	}

	findings := []schema.Finding{}
	var managed, unmanaged, missing, suppressed int

	// Actual side: present in desired => managed; otherwise unmanaged (unless
	// suppressed as AWS-managed noise).
	for _, key := range sortedKeys(actualByKey) {
		ar := actualByKey[key]
		if _, ok := desiredByKey[key]; ok {
			managed++
			continue
		}
		if opts.Suppress != nil && opts.Suppress(ar) {
			suppressed++
			continue
		}
		unmanaged++
		findings = append(findings, toFinding(schema.ClassUnmanaged, ar))
	}

	// Desired side: absent from actual => missing (deleted out-of-band).
	for _, key := range sortedKeys(desiredByKey) {
		if _, ok := actualByKey[key]; !ok {
			missing++
			findings = append(findings, toFinding(schema.ClassMissing, desiredByKey[key]))
		}
	}

	live := managed + unmanaged
	coverage := 100.0
	if live > 0 {
		coverage = 100.0 * float64(managed) / float64(live)
	}

	sortFindings(findings)

	return schema.Report{
		SchemaVersion: schema.Version,
		GeneratedAt:   opts.GeneratedAt,
		Provider:      opts.ProviderName,
		Scope: schema.Scope{
			Partition: scope.Partition,
			Account:   scope.Account,
			Regions:   scope.Regions,
			Types:     scanned,
		},
		Summary: schema.Summary{
			CoveragePercent: round1(coverage),
			Total:           managed + unmanaged + missing,
			Managed:         managed,
			Unmanaged:       unmanaged,
			Missing:         missing,
			Suppressed:      suppressed,
		},
		Findings:     findings,
		ScanStatuses: toSchemaStatuses(statuses),
	}
}

func toFinding(class schema.DriftClass, r resource.Resource) schema.Finding {
	return schema.Finding{
		Class:   class,
		Type:    r.Type,
		ID:      r.ID,
		Address: r.Address,
		ARN:     r.ARN,
		Region:  r.Region,
		Account: r.Account,
	}
}

func toSchemaStatuses(in []provider.ScanStatus) []schema.ScanStatus {
	out := make([]schema.ScanStatus, 0, len(in))
	for _, s := range in {
		out = append(out, schema.ScanStatus{
			Type:   s.Type,
			Region: s.Region,
			State:  s.State,
			Detail: s.Detail,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return out[i].Region < out[j].Region
	})
	return out
}

func sortedKeys(m map[string]resource.Resource) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// classOrder gives drift classes a stable, severity-ish display order.
var classOrder = map[schema.DriftClass]int{
	schema.ClassUnmanaged: 0,
	schema.ClassMissing:   1,
	schema.ClassChanged:   2,
	schema.ClassManaged:   3,
}

func sortFindings(f []schema.Finding) {
	sort.Slice(f, func(i, j int) bool {
		if classOrder[f[i].Class] != classOrder[f[j].Class] {
			return classOrder[f[i].Class] < classOrder[f[j].Class]
		}
		if f[i].Type != f[j].Type {
			return f[i].Type < f[j].Type
		}
		return f[i].ID < f[j].ID
	})
}

func round1(x float64) float64 {
	return math.Round(x*10) / 10
}

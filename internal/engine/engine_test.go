package engine

import (
	"testing"

	"github.com/yashyaadav/tf-drift-detector/internal/normalize"
	"github.com/yashyaadav/tf-drift-detector/internal/registry"
	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

func noop(id string) (string, error) { return id, nil }

func testReg() *registry.Registry {
	return registry.FromDescriptors([]provider.Descriptor{
		{TerraformType: "aws_security_group", Identity: noop},
		{TerraformType: "aws_iam_role", Global: true, Identity: noop},
	})
}

func live(typ, id, region string, attrs map[string]any) resource.Resource {
	return resource.Resource{
		Type: typ, ID: id, Partition: "aws", Account: "123456789012",
		Region: region, Source: resource.SourceActual, Attributes: attrs,
	}
}

func TestCompareClassification(t *testing.T) {
	reg := testReg()
	scope := provider.ScanScope{
		Partition: "aws", Account: "123456789012",
		Regions: []string{"us-east-1"},
		Types:   []string{"aws_security_group", "aws_iam_role"},
	}

	desired := []resource.Resource{
		{Type: "aws_security_group", ID: "sg-111"},     // managed
		{Type: "aws_security_group", ID: "sg-removed"}, // missing
		{Type: "aws_iam_role", ID: "role-A"},           // managed
		{Type: "aws_iam_role", ID: "role-B"},           // missing
	}
	actual := []resource.Resource{
		live("aws_security_group", "sg-111", "us-east-1", nil),                                                          // managed
		live("aws_security_group", "sg-999", "us-east-1", nil),                                                          // unmanaged
		live("aws_security_group", "sg-def", "us-east-1", map[string]any{"name": "default"}),                            // suppressed
		live("aws_iam_role", "role-A", resource.GlobalRegion, nil),                                                      // managed
		live("aws_iam_role", "AWSServiceRoleForX", resource.GlobalRegion, map[string]any{"path": "/aws-service-role/"}), // suppressed
	}

	rep := Compare(desired, actual, reg, scope, nil, Options{
		GeneratedAt: "t", ProviderName: "aws", Suppress: normalize.IsAWSManaged,
	})

	s := rep.Summary
	if s.Managed != 2 || s.Unmanaged != 1 || s.Missing != 2 || s.Suppressed != 2 {
		t.Fatalf("summary = %+v", s)
	}
	if s.CoveragePercent != 66.7 {
		t.Errorf("coverage = %v, want 66.7", s.CoveragePercent)
	}
	if len(rep.Findings) != 3 {
		t.Fatalf("findings = %d, want 3 (1 unmanaged + 2 missing)", len(rep.Findings))
	}
	// Unmanaged sorts before missing.
	if rep.Findings[0].Class != schema.ClassUnmanaged || rep.Findings[0].ID != "sg-999" {
		t.Errorf("first finding = %+v, want unmanaged sg-999", rep.Findings[0])
	}
}

// TestCompareCleanAccount asserts the headline trust property: an account that
// contains only AWS-managed/default resources (and nothing in state) reports
// zero unmanaged drift. See docs/adr/0007-default-suppression-heuristic.md.
func TestCompareCleanAccount(t *testing.T) {
	reg := testReg()
	scope := provider.ScanScope{
		Partition: "aws", Account: "123456789012",
		Regions: []string{"us-east-1"},
		Types:   []string{"aws_security_group", "aws_iam_role"},
	}
	actual := []resource.Resource{
		live("aws_security_group", "sg-def", "us-east-1", map[string]any{"name": "default"}),
		live("aws_iam_role", "AWSServiceRoleForOrg", resource.GlobalRegion, map[string]any{"path": "/aws-service-role/"}),
	}

	rep := Compare(nil, actual, reg, scope, nil, Options{Suppress: normalize.IsAWSManaged})

	if rep.Summary.Unmanaged != 0 {
		t.Errorf("clean account: unmanaged = %d, want 0 (findings: %+v)", rep.Summary.Unmanaged, rep.Findings)
	}
	if rep.Summary.Suppressed != 2 {
		t.Errorf("clean account: suppressed = %d, want 2", rep.Summary.Suppressed)
	}
	if rep.Summary.CoveragePercent != 100 {
		t.Errorf("clean account: coverage = %v, want 100", rep.Summary.CoveragePercent)
	}
}

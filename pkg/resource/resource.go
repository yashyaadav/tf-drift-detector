// Package resource defines the canonical, provider-agnostic resource model that
// both the desired (Terraform state) and actual (live cloud) sides reduce to, so
// a single drift engine can compare them.
//
// See docs/adr/0003-canonical-resource-model.md.
package resource

import "strings"

// Source identifies which side of the comparison a Resource came from.
type Source string

const (
	// SourceDesired is a resource read from Terraform state.
	SourceDesired Source = "desired"
	// SourceActual is a resource enumerated from the live cloud.
	SourceActual Source = "actual"
)

// GlobalRegion is the Region value used for resources that are not bound to a
// single region (e.g. IAM, the S3 bucket namespace). A sentinel keeps the
// identity key uniform across global and regional resources.
const GlobalRegion = "global"

// Resource is the canonical representation of one infrastructure resource.
//
// Identity is scoped by (Partition, Account, Region) from day one so that
// set-membership joins remain correct under future multi-account / multi-region
// scanning, without a breaking change to the output schema.
// See docs/adr/0008-identity-scoped-partition-account-region.md.
type Resource struct {
	Type       string            // Terraform type, e.g. "aws_security_group"
	ID         string            // canonical identifier (Terraform id == cloud primary id after identity mapping)
	Address    string            // Terraform address when known, e.g. module.x.aws_s3_bucket.this
	ARN        string            // cloud ARN when available
	Partition  string            // cloud partition, e.g. "aws"
	Account    string            // cloud account id
	Region     string            // region, or GlobalRegion for global resources
	Source     Source            // desired | actual
	Tags       map[string]string // resource tags, when available
	Attributes map[string]any    // normalized attributes (used by --deep attribute diffing)
}

// Key returns the identity key used to join desired against actual resources.
// Resources on both sides that share a Key are the "same" resource.
func (r Resource) Key() string {
	return strings.Join([]string{r.Partition, r.Account, r.Region, r.Type, r.ID}, "|")
}

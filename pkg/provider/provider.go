// Package provider defines the cloud abstraction. A Provider enumerates the
// actual resources in a scope by delegating to per-type Scanners. AWS implements
// this first; Azure (Resource Graph) and GCP (Cloud Asset Inventory) can be added
// behind the same interface later.
//
// See docs/adr/0002-aws-only-v1-provider-interface.md and
// docs/adr/0004-scanner-registry-bespoke-primary.md.
package provider

import (
	"context"

	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

// ScanScope bounds a scan.
type ScanScope struct {
	Partition      string
	Account        string
	Regions        []string // regions to scan; empty means "the provider's default region"
	Types          []string // Terraform types to scan; empty means "all registered types"
	MaxConcurrency int      // upper bound on concurrent (type,region) scans
}

// IdentityFn maps a Terraform resource id to its canonical cloud primary
// identifier. For many types this is the identity function; for composite-id
// types (e.g. aws_route53_record, aws_iam_role_policy) it is not.
//
// See docs/adr/0005-per-type-identity-join.md.
type IdentityFn func(terraformID string) (cloudID string, err error)

// Descriptor is the metadata for a resource type. It is deliberately free of any
// cloud client so the registry can answer identity/global questions for the
// desired (state) side without credentials.
type Descriptor struct {
	TerraformType string
	CloudType     string // e.g. CloudFormation type AWS::EC2::SecurityGroup
	Global        bool   // not bound to a single region (IAM, S3 namespace)
	Identity      IdentityFn
}

// ScanStatus records the outcome of scanning one (type, region) scope.
type ScanStatus struct {
	Type   string
	Region string
	State  schema.ScanState
	Detail string
}

// ScanResult is partial-failure aware: it carries whatever resources were
// collected plus a per-scope status list, so a throttle or AccessDenied in one
// scope never aborts the whole scan.
//
// See docs/adr/0009-partial-failure-scanstatus.md.
type ScanResult struct {
	Resources []resource.Resource
	Statuses  []ScanStatus
	// Scope is the resolved scope after the provider filled in account,
	// partition, and the default region(s). The engine uses it to scope joins.
	Scope ScanScope
}

// Provider enumerates live cloud resources for a scope.
type Provider interface {
	Name() string
	Scan(ctx context.Context, scope ScanScope) (ScanResult, error)
}

// Scanner enumerates the live resources of a single Terraform type in one region.
type Scanner interface {
	Descriptor() Descriptor
	// Scan returns the live resources of this type in the given region. For
	// global scanners the region argument is resource.GlobalRegion.
	Scan(ctx context.Context, scope ScanScope, region string) ([]resource.Resource, error)
}

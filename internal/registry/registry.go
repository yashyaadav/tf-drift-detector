// Package registry maps Terraform resource types to their metadata
// (descriptors), so both the engine (identity/global lookups for the desired
// side) and the CLI (listing supported types) can reason about types without
// holding any cloud client.
//
// See docs/adr/0004-scanner-registry-bespoke-primary.md.
package registry

import (
	"sort"

	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
)

// Registry is a set of resource-type descriptors keyed by Terraform type.
type Registry struct {
	byTF    map[string]provider.Descriptor
	byCloud map[string]string // cloud type -> terraform type
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{
		byTF:    map[string]provider.Descriptor{},
		byCloud: map[string]string{},
	}
}

// FromDescriptors builds a registry from a slice of descriptors.
func FromDescriptors(ds []provider.Descriptor) *Registry {
	r := New()
	for _, d := range ds {
		r.Register(d)
	}
	return r
}

// Register adds (or replaces) a descriptor.
func (r *Registry) Register(d provider.Descriptor) {
	r.byTF[d.TerraformType] = d
	if d.CloudType != "" {
		r.byCloud[d.CloudType] = d.TerraformType
	}
}

// Get returns the descriptor for a Terraform type.
func (r *Registry) Get(tfType string) (provider.Descriptor, bool) {
	d, ok := r.byTF[tfType]
	return d, ok
}

// Has reports whether a Terraform type is registered.
func (r *Registry) Has(tfType string) bool {
	_, ok := r.byTF[tfType]
	return ok
}

// Types returns the registered Terraform types, sorted.
func (r *Registry) Types() []string {
	out := make([]string, 0, len(r.byTF))
	for t := range r.byTF {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// TerraformTypeForCloud resolves a cloud type to its Terraform type.
func (r *Registry) TerraformTypeForCloud(cloudType string) (string, bool) {
	t, ok := r.byCloud[cloudType]
	return t, ok
}

// CanonicalID applies the type's IdentityFn to a Terraform id to produce the
// canonical cloud identifier used for joining. If the type has no IdentityFn, or
// the mapping fails, the original id is returned unchanged.
//
// See docs/adr/0005-per-type-identity-join.md.
func (r *Registry) CanonicalID(tfType, id string) string {
	d, ok := r.byTF[tfType]
	if !ok || d.Identity == nil {
		return id
	}
	cloudID, err := d.Identity(id)
	if err != nil || cloudID == "" {
		return id
	}
	return cloudID
}

// Global reports whether a type is region-independent.
func (r *Registry) Global(tfType string) bool {
	d, ok := r.byTF[tfType]
	return ok && d.Global
}

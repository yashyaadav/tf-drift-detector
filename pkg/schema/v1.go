// Package schema defines the versioned JSON output contract for tfdrift.
//
// It is intentionally dependency-free: this is the stable serialization surface
// that CI consumers and dashboards depend on. Changes within v1 must be additive;
// a breaking change starts schema/v2.
//
// See docs/adr/0011 (forthcoming) — versioned JSON output as a contract.
package schema

// Version is the current output schema version.
const Version = "1"

// DriftClass is how the engine classifies a resource.
type DriftClass string

const (
	// ClassManaged: present in Terraform state and in the cloud.
	ClassManaged DriftClass = "managed"
	// ClassUnmanaged: in the cloud but not in any state (orphan / ClickOps).
	ClassUnmanaged DriftClass = "unmanaged"
	// ClassMissing: in state but not in the cloud (deleted out-of-band).
	ClassMissing DriftClass = "missing"
	// ClassChanged: present on both sides but attributes differ (--deep).
	ClassChanged DriftClass = "changed"
)

// ScanState is the outcome of scanning a single (type, region) scope.
// Denied/Throttled/Skipped/Unsupported scopes are reported and excluded from the
// coverage denominator — never silently treated as "no resources".
// See docs/adr/0009-partial-failure-scanstatus.md.
type ScanState string

const (
	ScanScanned     ScanState = "scanned"
	ScanThrottled   ScanState = "throttled"
	ScanDenied      ScanState = "denied"
	ScanSkipped     ScanState = "skipped"
	ScanUnsupported ScanState = "unsupported"
)

// Report is the top-level tfdrift output document.
type Report struct {
	SchemaVersion string       `json:"schemaVersion"`
	GeneratedAt   string       `json:"generatedAt"`
	Provider      string       `json:"provider"`
	Scope         Scope        `json:"scope"`
	Summary       Summary      `json:"summary"`
	Findings      []Finding    `json:"findings"`
	ScanStatuses  []ScanStatus `json:"scanStatuses"`
}

// Scope echoes back what was scanned.
type Scope struct {
	Partition string   `json:"partition,omitempty"`
	Account   string   `json:"account,omitempty"`
	Regions   []string `json:"regions,omitempty"`
	Types     []string `json:"types,omitempty"`
}

// Summary is the headline rollup. CoveragePercent is the share of live cloud
// resources (managed + unmanaged) that Terraform manages.
type Summary struct {
	CoveragePercent float64 `json:"coveragePercent"`
	Total           int     `json:"total"`
	Managed         int     `json:"managed"`
	Unmanaged       int     `json:"unmanaged"`
	Missing         int     `json:"missing"`
	Changed         int     `json:"changed"`
	Suppressed      int     `json:"suppressed"`
}

// Finding is one drifted resource.
type Finding struct {
	Class   DriftClass `json:"class"`
	Type    string     `json:"type"`
	ID      string     `json:"id"`
	Address string     `json:"address,omitempty"`
	ARN     string     `json:"arn,omitempty"`
	Region  string     `json:"region,omitempty"`
	Account string     `json:"account,omitempty"`
	Changes []Change   `json:"changes,omitempty"` // populated for ClassChanged
}

// Change is a single attribute difference (used by --deep).
type Change struct {
	Path    string `json:"path"`
	Desired any    `json:"desired"`
	Actual  any    `json:"actual"`
}

// ScanStatus is the per-scope outcome of enumeration.
type ScanStatus struct {
	Type   string    `json:"type"`
	Region string    `json:"region"`
	State  ScanState `json:"state"`
	Detail string    `json:"detail,omitempty"`
}

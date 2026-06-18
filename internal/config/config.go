// Package config holds the resolved, decoupled scan configuration. It is a plain
// value object so the rest of the app does not depend on the CLI framework.
//
// M0 populates it from cobra flags; flag>env>file precedence (viper) is an M1
// addition behind this same struct.
package config

// Config is the fully-resolved configuration for a scan.
type Config struct {
	StatePath      string   // path to terraform.tfstate
	Output         string   // table | json
	Regions        []string // AWS regions to scan (empty = provider default)
	Types          []string // Terraform types to scan (empty = all supported)
	MaxConcurrency int      // upper bound on concurrent scans
	FailOn         string   // none | unmanaged | missing | any
	FailOnCount    int      // exit 2 when findings exceed this count (0 = disabled)
}

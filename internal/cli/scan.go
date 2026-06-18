package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	awscloud "github.com/yashyaadav/tf-drift-detector/internal/cloud/aws"
	"github.com/yashyaadav/tf-drift-detector/internal/config"
	"github.com/yashyaadav/tf-drift-detector/internal/engine"
	"github.com/yashyaadav/tf-drift-detector/internal/normalize"
	"github.com/yashyaadav/tf-drift-detector/internal/registry"
	"github.com/yashyaadav/tf-drift-detector/internal/report"
	"github.com/yashyaadav/tf-drift-detector/internal/state"
	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

func newScanCmd() *cobra.Command {
	cfg := config.Config{}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for drift between Terraform state and live AWS",
		Example: "  tfdrift scan --state ./terraform.tfstate --types aws_security_group --region us-east-1\n" +
			"  tfdrift scan --state ./terraform.tfstate --output json | jq '.summary'\n" +
			"  tfdrift scan --state ./terraform.tfstate --fail-on unmanaged",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runScan(cmd.Context(), cmd.OutOrStdout(), cfg)
		},
	}
	f := cmd.Flags()
	f.StringVar(&cfg.StatePath, "state", "", "path to terraform.tfstate (required)")
	f.StringVarP(&cfg.Output, "output", "o", "table", "output format: table|json")
	f.StringSliceVar(&cfg.Regions, "region", nil, "AWS region(s) to scan (default: AWS_REGION / profile)")
	f.StringSliceVar(&cfg.Types, "types", nil, "Terraform type(s) to scan (default: all supported)")
	f.IntVar(&cfg.MaxConcurrency, "max-concurrency", 8, "max concurrent (type,region) scans")
	f.StringVar(&cfg.FailOn, "fail-on", "none", "exit 2 on drift class: none|unmanaged|missing|any")
	f.IntVar(&cfg.FailOnCount, "fail-on-count", 0, "exit 2 when findings exceed this count (0 = disabled)")
	_ = cmd.MarkFlagRequired("state")
	return cmd
}

func runScan(ctx context.Context, out io.Writer, cfg config.Config) error {
	// --- validate inputs (friendly, single-line preflight errors) ---
	if cfg.Output != "table" && cfg.Output != "json" {
		return fmt.Errorf("invalid --output %q (want table|json)", cfg.Output)
	}
	if !validFailOn(cfg.FailOn) {
		return fmt.Errorf("invalid --fail-on %q (want none|unmanaged|missing|any)", cfg.FailOn)
	}
	if fi, err := os.Stat(cfg.StatePath); err != nil || fi.IsDir() {
		return fmt.Errorf("cannot read state file %q (point --state at a terraform.tfstate file)", cfg.StatePath)
	}

	reg := registry.FromDescriptors(awscloud.Descriptors())
	for _, t := range cfg.Types {
		if !reg.Has(t) {
			return fmt.Errorf("unsupported type %q; supported types: %s", t, strings.Join(reg.Types(), ", "))
		}
	}

	// --- desired side: Terraform state ---
	loader := state.LocalLoader{}
	rc, err := loader.Load(ctx, cfg.StatePath)
	if err != nil {
		return err
	}
	defer rc.Close()
	desired, err := state.Read(rc)
	if err != nil {
		return err
	}

	// --- actual side: live AWS ---
	region := ""
	if len(cfg.Regions) > 0 {
		region = cfg.Regions[0]
	}
	awsCfg, err := awscloud.LoadConfig(ctx, region)
	if err != nil {
		return fmt.Errorf("load AWS config: %w", err)
	}
	prov := awscloud.NewProvider(awsCfg)

	scope := provider.ScanScope{
		Regions:        cfg.Regions,
		Types:          cfg.Types,
		MaxConcurrency: cfg.MaxConcurrency,
	}
	slog.Info("scanning", "provider", prov.Name(), "types", scanTypes(cfg.Types, reg), "regions", cfg.Regions)
	result, err := prov.Scan(ctx, scope)
	if err != nil {
		return err
	}

	// --- compare + report ---
	rep := engine.Compare(desired, result.Resources, reg, result.Scope, result.Statuses, engine.Options{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		ProviderName: prov.Name(),
		Suppress:     normalize.IsAWSManaged,
	})
	if err := report.Render(out, rep, cfg.Output); err != nil {
		return err
	}

	if code := gateExitCode(rep, cfg); code != 0 {
		return &exitError{code: code}
	}
	return nil
}

func validFailOn(v string) bool {
	switch v {
	case "none", "unmanaged", "missing", "any":
		return true
	}
	return false
}

// gateExitCode returns 2 when drift trips the configured CI gate, else 0.
func gateExitCode(rep schema.Report, cfg config.Config) int {
	s := rep.Summary
	trip := false
	switch cfg.FailOn {
	case "unmanaged":
		trip = s.Unmanaged > 0
	case "missing":
		trip = s.Missing > 0
	case "any":
		trip = s.Unmanaged > 0 || s.Missing > 0 || s.Changed > 0
	}
	if cfg.FailOnCount > 0 && len(rep.Findings) > cfg.FailOnCount {
		trip = true
	}
	if trip {
		return 2
	}
	return 0
}

func scanTypes(types []string, reg *registry.Registry) []string {
	if len(types) > 0 {
		return types
	}
	return reg.Types()
}

// Package report renders a schema.Report. Both the human (table) and machine
// (JSON) renderers are driven by the same Report so they never diverge. Logs go
// to stderr elsewhere; renderers write only the report to the given writer, so
// `--output json` stdout stays pipeable to jq.
//
// M0 uses stdlib text/tabwriter for aligned columns to keep the dependency
// surface minimal; a richer table library can be swapped in later.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

// Render writes rep to w in the requested format ("table" or "json").
func Render(w io.Writer, rep schema.Report, format string) error {
	switch format {
	case "json":
		return renderJSON(w, rep)
	case "table", "":
		return renderTable(w, rep)
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}

func renderJSON(w io.Writer, rep schema.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

func renderTable(w io.Writer, rep schema.Report) error {
	s := rep.Summary
	fmt.Fprintf(w, "%.1f%% of live resources managed by Terraform  ·  %d managed  ·  %d unmanaged  ·  %d missing",
		s.CoveragePercent, s.Managed, s.Unmanaged, s.Missing)
	if s.Suppressed > 0 {
		fmt.Fprintf(w, "  (%d AWS-managed suppressed)", s.Suppressed)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	if len(rep.Findings) == 0 {
		fmt.Fprintln(w, "No drift detected.")
	} else {
		tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, "CLASS\tTYPE\tID\tREGION")
		for _, f := range rep.Findings {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", f.Class, f.Type, f.ID, f.Region)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	}

	// Surface scopes that could not be fully scanned (denied/throttled/etc.).
	var warn []schema.ScanStatus
	for _, st := range rep.ScanStatuses {
		if st.State != schema.ScanScanned {
			warn = append(warn, st)
		}
	}
	if len(warn) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Scan warnings:")
		for _, st := range warn {
			region := st.Region
			if region == "" {
				region = "-"
			}
			fmt.Fprintf(w, "  ! %s [%s] %s: %s\n", st.State, region, st.Type, st.Detail)
		}
	}
	return nil
}

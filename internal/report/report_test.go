package report

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

var update = flag.Bool("update", false, "update golden files")

func sampleReport() schema.Report {
	return schema.Report{
		SchemaVersion: schema.Version,
		GeneratedAt:   "2026-06-18T00:00:00Z",
		Provider:      "aws",
		Scope: schema.Scope{
			Partition: "aws",
			Account:   "123456789012",
			Regions:   []string{"us-east-1"},
			Types:     []string{"aws_iam_role", "aws_security_group"},
		},
		Summary: schema.Summary{
			CoveragePercent: 66.7,
			Total:           4,
			Managed:         2,
			Unmanaged:       1,
			Missing:         1,
			Suppressed:      2,
		},
		Findings: []schema.Finding{
			{Class: schema.ClassUnmanaged, Type: "aws_security_group", ID: "sg-999", Region: "us-east-1", Account: "123456789012"},
			{Class: schema.ClassMissing, Type: "aws_iam_role", ID: "role-B", Region: "global", Account: "123456789012"},
		},
		ScanStatuses: []schema.ScanStatus{
			{Type: "aws_iam_role", Region: "global", State: schema.ScanScanned},
			{Type: "aws_instance", Region: "us-west-2", State: schema.ScanDenied, Detail: "AccessDenied"},
			{Type: "aws_security_group", Region: "us-east-1", State: schema.ScanScanned},
		},
	}
}

func TestRenderGolden(t *testing.T) {
	rep := sampleReport()
	cases := []struct{ name, format, golden string }{
		{"table", "table", "report.table.golden"},
		{"json", "json", "report.json.golden"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Render(&buf, rep, tc.format); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", tc.golden)
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden (run `make golden` to generate): %v", err)
			}
			if buf.String() != string(want) {
				t.Errorf("%s output mismatch\n--- got ---\n%s\n--- want ---\n%s", tc.format, buf.String(), want)
			}
		})
	}
}

func TestRenderUnknownFormat(t *testing.T) {
	if err := Render(&bytes.Buffer{}, sampleReport(), "yaml"); err == nil {
		t.Error("expected error for unknown format")
	}
}

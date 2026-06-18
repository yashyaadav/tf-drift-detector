package aws

import "testing"

// TestIdentityConformance asserts every registered descriptor has an IdentityFn
// that round-trips the canonical ids of its type. For M0 types the mapping is
// the identity function (Terraform id == cloud primary id), but the harness is
// the gate that future composite-id types must pass before they ship.
// See docs/adr/0005-per-type-identity-join.md.
func TestIdentityConformance(t *testing.T) {
	ids := map[string]string{
		"aws_s3_bucket":      "my-data-bucket",
		"aws_security_group": "sg-0abc123",
		"aws_instance":       "i-0abc123",
		"aws_iam_role":       "app-role",
	}
	for _, d := range Descriptors() {
		if d.Identity == nil {
			t.Errorf("%s: nil IdentityFn", d.TerraformType)
			continue
		}
		id, ok := ids[d.TerraformType]
		if !ok {
			t.Errorf("%s: no conformance fixture id", d.TerraformType)
			continue
		}
		got, err := d.Identity(id)
		if err != nil || got != id {
			t.Errorf("%s: Identity(%q) = %q, %v; want %q, nil", d.TerraformType, id, got, err, id)
		}
	}
}

func TestParseRoute53RecordID(t *testing.T) {
	zone, name, typ, err := ParseRoute53RecordID("Z123_www.example.com_A")
	if err != nil || zone != "Z123" || name != "www.example.com" || typ != "A" {
		t.Fatalf("ParseRoute53RecordID = (%q, %q, %q, %v)", zone, name, typ, err)
	}
	if _, _, _, err := ParseRoute53RecordID("not-a-composite"); err == nil {
		t.Error("expected error for malformed route53 record id")
	}
}

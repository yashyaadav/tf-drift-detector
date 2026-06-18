package resource

import "testing"

func TestKey(t *testing.T) {
	r := Resource{Partition: "aws", Account: "123456789012", Region: "us-east-1", Type: "aws_instance", ID: "i-1"}
	if got, want := r.Key(), "aws|123456789012|us-east-1|aws_instance|i-1"; got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

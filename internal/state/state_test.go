package state

import (
	"os"
	"strings"
	"testing"

	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
)

func TestReadSimpleState(t *testing.T) {
	f, err := os.Open("testdata/simple_state.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	res, err := Read(f)
	if err != nil {
		t.Fatal(err)
	}

	// 7 managed instances; the data source is skipped.
	if len(res) != 7 {
		t.Fatalf("got %d managed resources, want 7", len(res))
	}

	byID := map[string]resource.Resource{}
	for _, r := range res {
		byID[r.Type+"/"+r.ID] = r
	}

	if b, ok := byID["aws_s3_bucket/my-data-bucket"]; !ok {
		t.Error("missing aws_s3_bucket/my-data-bucket")
	} else if b.Tags["env"] != "prod" {
		t.Errorf("bucket tags = %v, want env=prod", b.Tags)
	}

	// Sensitive values must be redacted in the desired model.
	if db, ok := byID["aws_db_instance/db-1"]; !ok {
		t.Error("missing aws_db_instance/db-1")
	} else if db.Attributes["password"] != "<sensitive>" {
		t.Errorf("password = %v, want <sensitive>", db.Attributes["password"])
	}

	// Module + for_each index key must reconstruct into the address.
	var subnetA *resource.Resource
	for i := range res {
		if res[i].Type == "aws_subnet" && res[i].ID == "subnet-a" {
			subnetA = &res[i]
		}
	}
	if subnetA == nil {
		t.Fatal("missing module.network.aws_subnet.private[\"a\"]")
	}
	if want := `module.network.aws_subnet.private["a"]`; subnetA.Address != want {
		t.Errorf("address = %q, want %q", subnetA.Address, want)
	}
}

func TestReadRejectsUnsupportedVersion(t *testing.T) {
	const v3 = `{"version":3,"resources":[]}`
	if _, err := Read(strings.NewReader(v3)); err == nil {
		t.Error("expected error for version 3 state")
	}
}

package aws

import (
	"fmt"
	"strings"

	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
)

// identityNoop is the IdentityFn for types whose Terraform id already equals the
// cloud primary identifier — the case for every M0 type (sg-id, bucket name,
// instance-id, role name). See docs/adr/0005-per-type-identity-join.md.
func identityNoop(id string) (string, error) { return id, nil }

var (
	descS3Bucket = provider.Descriptor{
		TerraformType: "aws_s3_bucket",
		CloudType:     "AWS::S3::Bucket",
		Global:        true, // bucket namespace is global; keyed without a region
		Identity:      identityNoop,
	}
	descSecurityGroup = provider.Descriptor{
		TerraformType: "aws_security_group",
		CloudType:     "AWS::EC2::SecurityGroup",
		Global:        false,
		Identity:      identityNoop,
	}
	descInstance = provider.Descriptor{
		TerraformType: "aws_instance",
		CloudType:     "AWS::EC2::Instance",
		Global:        false,
		Identity:      identityNoop,
	}
	descIAMRole = provider.Descriptor{
		TerraformType: "aws_iam_role",
		CloudType:     "AWS::IAM::Role",
		Global:        true,
		Identity:      identityNoop,
	}
)

// Descriptors returns the metadata for every AWS type tfdrift can scan. Used to
// build the registry for the desired (state) side without needing credentials.
func Descriptors() []provider.Descriptor {
	return []provider.Descriptor{descS3Bucket, descSecurityGroup, descInstance, descIAMRole}
}

// ParseRoute53RecordID demonstrates a non-trivial identity mapping: an
// aws_route53_record Terraform id is the composite "ZONEID_NAME_TYPE" (optionally
// with a trailing "_set-identifier"), which is NOT the cloud primary identifier.
// Types like this register an IdentityFn plus a conformance test; see
// docs/adr/0005-per-type-identity-join.md. Not scanned in M0, but it exercises
// the identity-join harness so the pattern is proven before such a type ships.
func ParseRoute53RecordID(id string) (zoneID, name, recordType string, err error) {
	parts := strings.SplitN(id, "_", 3)
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("invalid route53 record id %q: want ZONEID_NAME_TYPE", id)
	}
	return parts[0], parts[1], parts[2], nil
}

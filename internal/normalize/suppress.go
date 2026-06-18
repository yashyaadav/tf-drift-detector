// Package normalize decides which live resources are AWS-managed noise that
// should never be reported as drift. Without this, even a clean account scans to
// dozens of false "unmanaged" findings (default VPC/SG, service-linked roles,
// AWS-managed tags), which is the difference between a credible drift tool and a
// naive diff script.
//
// Suppression is expressed as a maintained heuristic over ownership signals
// rather than a static list of resource ids, which would rot immediately.
//
// See docs/adr/0007-default-suppression-heuristic.md.
package normalize

import (
	"strings"

	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
)

// IsAWSManaged reports whether a live resource is owned/created by AWS itself
// (and therefore not something the user is expected to manage in Terraform).
//
// The signals are deliberately ownership-based and generalize across types:
//   - IAM service-linked roles (path /aws-service-role/ or AWSServiceRoleFor* name)
//   - the default security group of a VPC (group name "default")
//   - default VPCs / subnets (is_default attribute)
//   - any resource carrying a reserved "aws:" tag (set by AWS services, e.g.
//     aws:cloudformation:*) which signals AWS/CloudFormation ownership.
func IsAWSManaged(r resource.Resource) bool {
	// Reserved AWS tag prefix => managed by an AWS service.
	for k := range r.Tags {
		if strings.HasPrefix(strings.ToLower(k), "aws:") {
			return true
		}
	}

	switch r.Type {
	case "aws_iam_role":
		if s, _ := r.Attributes["path"].(string); strings.HasPrefix(s, "/aws-service-role/") {
			return true
		}
		if strings.HasPrefix(r.ID, "AWSServiceRoleFor") {
			return true
		}
		if strings.Contains(r.ARN, ":role/aws-service-role/") {
			return true
		}
	case "aws_security_group":
		// The default SG of every VPC is created by AWS, not Terraform.
		if s, _ := r.Attributes["name"].(string); s == "default" {
			return true
		}
	case "aws_vpc", "aws_subnet":
		if b, _ := r.Attributes["is_default"].(bool); b {
			return true
		}
	}
	return false
}

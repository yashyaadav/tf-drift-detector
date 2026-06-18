package normalize

import (
	"testing"

	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
)

func TestIsAWSManaged(t *testing.T) {
	cases := []struct {
		name string
		r    resource.Resource
		want bool
	}{
		{"service-linked role by path", resource.Resource{Type: "aws_iam_role", ID: "X", Attributes: map[string]any{"path": "/aws-service-role/"}}, true},
		{"service-linked role by name", resource.Resource{Type: "aws_iam_role", ID: "AWSServiceRoleForFoo"}, true},
		{"service-linked role by arn", resource.Resource{Type: "aws_iam_role", ID: "X", ARN: "arn:aws:iam::1:role/aws-service-role/foo"}, true},
		{"default security group", resource.Resource{Type: "aws_security_group", ID: "sg-1", Attributes: map[string]any{"name": "default"}}, true},
		{"aws reserved tag", resource.Resource{Type: "aws_s3_bucket", ID: "b", Tags: map[string]string{"aws:cloudformation:stack-name": "x"}}, true},
		{"default vpc", resource.Resource{Type: "aws_vpc", ID: "vpc-1", Attributes: map[string]any{"is_default": true}}, true},
		{"normal security group", resource.Resource{Type: "aws_security_group", ID: "sg-2", Attributes: map[string]any{"name": "web"}}, false},
		{"normal role", resource.Resource{Type: "aws_iam_role", ID: "app", Attributes: map[string]any{"path": "/"}}, false},
		{"normal bucket", resource.Resource{Type: "aws_s3_bucket", ID: "b", Tags: map[string]string{"env": "prod"}}, false},
	}
	for _, c := range cases {
		if got := IsAWSManaged(c.r); got != c.want {
			t.Errorf("%s: IsAWSManaged = %v, want %v", c.name, got, c.want)
		}
	}
}

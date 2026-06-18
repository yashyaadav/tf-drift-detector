package aws

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
)

// newScanners builds the bespoke per-type scanners from a base AWS config.
func newScanners(cfg awssdk.Config) map[string]provider.Scanner {
	list := []provider.Scanner{
		s3Scanner{cfg},
		securityGroupScanner{cfg},
		instanceScanner{cfg},
		iamRoleScanner{cfg},
	}
	m := make(map[string]provider.Scanner, len(list))
	for _, sc := range list {
		m[sc.Descriptor().TerraformType] = sc
	}
	return m
}

// regionalConfig returns a copy of the base config pinned to a region.
func regionalConfig(base awssdk.Config, region string) awssdk.Config {
	c := base.Copy()
	c.Region = region
	return c
}

// --- aws_s3_bucket (global namespace) ---

type s3Scanner struct{ cfg awssdk.Config }

func (s s3Scanner) Descriptor() provider.Descriptor { return descS3Bucket }

func (s s3Scanner) Scan(ctx context.Context, _ provider.ScanScope, _ string) ([]resource.Resource, error) {
	client := s3.NewFromConfig(s.cfg)
	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	res := make([]resource.Resource, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		name := awssdk.ToString(b.Name)
		res = append(res, resource.Resource{
			Type:       descS3Bucket.TerraformType,
			ID:         name,
			ARN:        "arn:aws:s3:::" + name,
			Region:     resource.GlobalRegion,
			Source:     resource.SourceActual,
			Attributes: map[string]any{"bucket": name},
		})
	}
	return res, nil
}

// --- aws_security_group (regional) ---

type securityGroupScanner struct{ cfg awssdk.Config }

func (s securityGroupScanner) Descriptor() provider.Descriptor { return descSecurityGroup }

func (s securityGroupScanner) Scan(ctx context.Context, _ provider.ScanScope, region string) ([]resource.Resource, error) {
	client := ec2.NewFromConfig(regionalConfig(s.cfg, region))
	p := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{})
	var res []resource.Resource
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, sg := range out.SecurityGroups {
			res = append(res, resource.Resource{
				Type:   descSecurityGroup.TerraformType,
				ID:     awssdk.ToString(sg.GroupId),
				Region: region,
				Source: resource.SourceActual,
				Tags:   ec2Tags(sg.Tags),
				Attributes: map[string]any{
					"name":        awssdk.ToString(sg.GroupName),
					"vpc_id":      awssdk.ToString(sg.VpcId),
					"description": awssdk.ToString(sg.Description),
				},
			})
		}
	}
	return res, nil
}

// --- aws_instance (regional) ---

type instanceScanner struct{ cfg awssdk.Config }

func (s instanceScanner) Descriptor() provider.Descriptor { return descInstance }

func (s instanceScanner) Scan(ctx context.Context, _ provider.ScanScope, region string) ([]resource.Resource, error) {
	client := ec2.NewFromConfig(regionalConfig(s.cfg, region))
	p := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	var res []resource.Resource
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, rv := range out.Reservations {
			for _, inst := range rv.Instances {
				// Terminated instances are gone; do not report them as live.
				if inst.State != nil && inst.State.Name == ec2types.InstanceStateNameTerminated {
					continue
				}
				res = append(res, resource.Resource{
					Type:   descInstance.TerraformType,
					ID:     awssdk.ToString(inst.InstanceId),
					Region: region,
					Source: resource.SourceActual,
					Tags:   ec2Tags(inst.Tags),
					Attributes: map[string]any{
						"instance_type": string(inst.InstanceType),
						"ami":           awssdk.ToString(inst.ImageId),
					},
				})
			}
		}
	}
	return res, nil
}

// --- aws_iam_role (global) ---

type iamRoleScanner struct{ cfg awssdk.Config }

func (s iamRoleScanner) Descriptor() provider.Descriptor { return descIAMRole }

func (s iamRoleScanner) Scan(ctx context.Context, _ provider.ScanScope, _ string) ([]resource.Resource, error) {
	client := iam.NewFromConfig(s.cfg) // IAM is global
	p := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})
	var res []resource.Resource
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, role := range out.Roles {
			res = append(res, resource.Resource{
				Type:   descIAMRole.TerraformType,
				ID:     awssdk.ToString(role.RoleName),
				ARN:    awssdk.ToString(role.Arn),
				Region: resource.GlobalRegion,
				Source: resource.SourceActual,
				Attributes: map[string]any{
					"path": awssdk.ToString(role.Path),
					"name": awssdk.ToString(role.RoleName),
				},
			})
		}
	}
	return res, nil
}

func ec2Tags(tags []ec2types.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[awssdk.ToString(t.Key)] = awssdk.ToString(t.Value)
	}
	return m
}

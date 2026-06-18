// Package aws implements the cloud Provider for AWS: it resolves the account and
// partition, fans out the per-type Scanners across regions with bounded
// concurrency, and isolates per-scope failures (a throttle or AccessDenied in one
// scope never aborts the scan).
//
// See docs/adr/0002, 0004, 0009 (partial failure) and 0010 (throttling).
package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithy "github.com/aws/smithy-go"

	"github.com/yashyaadav/tf-drift-detector/pkg/provider"
	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
	"github.com/yashyaadav/tf-drift-detector/pkg/schema"
)

const defaultMaxConcurrency = 8

// Provider scans AWS for live resources.
type Provider struct {
	cfg      awssdk.Config
	scanners map[string]provider.Scanner
}

// LoadConfig loads the default AWS config with adaptive retry as the throttling
// backstop (in-process pacing is layered on top later). See ADR-0010.
func LoadConfig(ctx context.Context, region string) (awssdk.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRetryMode(awssdk.RetryModeAdaptive),
		awsconfig.WithRetryMaxAttempts(5),
	}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	return awsconfig.LoadDefaultConfig(ctx, opts...)
}

// NewProvider builds an AWS provider from a loaded config.
func NewProvider(cfg awssdk.Config) *Provider {
	return &Provider{cfg: cfg, scanners: newScanners(cfg)}
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return "aws" }

// Scan enumerates live AWS resources for the scope.
func (p *Provider) Scan(ctx context.Context, scope provider.ScanScope) (provider.ScanResult, error) {
	// Resolve account + partition (also validates credentials early).
	ident, err := sts.NewFromConfig(p.cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return provider.ScanResult{}, fmt.Errorf("resolve AWS identity (check credentials/region): %w", err)
	}
	scope.Partition = partitionFromARN(awssdk.ToString(ident.Arn))
	scope.Account = awssdk.ToString(ident.Account)

	if len(scope.Regions) == 0 {
		if p.cfg.Region == "" {
			return provider.ScanResult{}, errors.New("no region configured; set --region or AWS_REGION")
		}
		scope.Regions = []string{p.cfg.Region}
	}

	types := scope.Types
	if len(types) == 0 {
		for t := range p.scanners {
			types = append(types, t)
		}
	}

	type task struct {
		sc     provider.Scanner
		region string
	}
	var tasks []task
	var statuses []provider.ScanStatus
	for _, t := range types {
		sc, ok := p.scanners[t]
		if !ok {
			statuses = append(statuses, provider.ScanStatus{
				Type:   t,
				State:  schema.ScanUnsupported,
				Detail: "no scanner registered for type",
			})
			continue
		}
		if sc.Descriptor().Global {
			tasks = append(tasks, task{sc, resource.GlobalRegion})
		} else {
			for _, rg := range scope.Regions {
				tasks = append(tasks, task{sc, rg})
			}
		}
	}

	maxConc := scope.MaxConcurrency
	if maxConc <= 0 {
		maxConc = defaultMaxConcurrency
	}

	var (
		mu        sync.Mutex
		resources []resource.Resource
		wg        sync.WaitGroup
		sem       = make(chan struct{}, maxConc)
	)
	for _, tk := range tasks {
		wg.Add(1)
		sem <- struct{}{}
		go func(tk task) {
			defer wg.Done()
			defer func() { <-sem }()

			res, scanErr := tk.sc.Scan(ctx, scope, tk.region)
			tfType := tk.sc.Descriptor().TerraformType

			mu.Lock()
			defer mu.Unlock()
			if scanErr != nil {
				st, detail := classifyErr(scanErr)
				statuses = append(statuses, provider.ScanStatus{Type: tfType, Region: tk.region, State: st, Detail: detail})
				return
			}
			for i := range res {
				res[i].Partition = scope.Partition
				res[i].Account = scope.Account
			}
			resources = append(resources, res...)
			statuses = append(statuses, provider.ScanStatus{Type: tfType, Region: tk.region, State: schema.ScanScanned})
		}(tk)
	}
	wg.Wait()

	return provider.ScanResult{Resources: resources, Statuses: statuses, Scope: scope}, nil
}

func partitionFromARN(arn string) string {
	// arn:PARTITION:service:region:account:resource
	parts := strings.SplitN(arn, ":", 3)
	if len(parts) >= 2 && parts[0] == "arn" && parts[1] != "" {
		return parts[1]
	}
	return "aws"
}

// classifyErr maps an AWS API error to a scan status so a permissions gap or a
// throttle is reported distinctly and never mistaken for "no resources".
func classifyErr(err error) (schema.ScanState, string) {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		switch {
		case strings.Contains(code, "AccessDenied"), code == "UnauthorizedOperation", strings.Contains(code, "Forbidden"):
			return schema.ScanDenied, code
		case strings.Contains(code, "Throttl"), code == "RequestLimitExceeded", code == "TooManyRequestsException":
			return schema.ScanThrottled, code
		}
		return schema.ScanSkipped, code
	}
	return schema.ScanSkipped, err.Error()
}

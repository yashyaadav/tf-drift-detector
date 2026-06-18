// Package state reads Terraform state into the canonical resource model — the
// "desired" side of the comparison.
//
// M0 ships a direct parser for the on-disk v4 format behind the Loader/Reader
// seam; a fujiwara/tfstate-lookup adapter for remote backends (S3, GCS, TFC)
// drops in behind the same interface in M1.
//
// See docs/adr/0013 (forthcoming) — tfstate parsing strategy.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/yashyaadav/tf-drift-detector/pkg/resource"
)

// Loader fetches raw Terraform state bytes from a location.
type Loader interface {
	Load(ctx context.Context, location string) (io.ReadCloser, error)
}

// LocalLoader reads state from a local file path.
type LocalLoader struct{}

// Load opens a local state file.
func (LocalLoader) Load(_ context.Context, path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open state file %q: %w", path, err)
	}
	return f, nil
}

// Read parses Terraform state (on-disk v4 format) into canonical resources.
// Only "managed" resources are returned; data sources are not real infrastructure.
// Numbers decode as json.Number to avoid float64 corruption of 64-bit ids/ports.
func Read(r io.Reader) ([]resource.Resource, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var f v4File
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("decode tfstate: %w", err)
	}
	if f.Version != 0 && f.Version != 4 {
		return nil, fmt.Errorf("unsupported tfstate version %d (expected 4); upgrade it with terraform first", f.Version)
	}

	var out []resource.Resource
	for _, res := range f.Resources {
		if res.Mode != "managed" {
			continue
		}
		for _, inst := range res.Instances {
			out = append(out, toResource(res, inst))
		}
	}
	return out, nil
}

type v4File struct {
	Version   int          `json:"version"`
	Resources []v4Resource `json:"resources"`
}

type v4Resource struct {
	Module    string       `json:"module"`
	Mode      string       `json:"mode"`
	Type      string       `json:"type"`
	Name      string       `json:"name"`
	Provider  string       `json:"provider"`
	Instances []v4Instance `json:"instances"`
}

type v4Instance struct {
	IndexKey            any               `json:"index_key"`
	Attributes          map[string]any    `json:"attributes"`
	SensitiveAttributes []json.RawMessage `json:"sensitive_attributes"`
}

func toResource(res v4Resource, inst v4Instance) resource.Resource {
	attrs := inst.Attributes
	redactSensitive(attrs, inst.SensitiveAttributes)
	return resource.Resource{
		Type:       res.Type,
		ID:         attrString(attrs, "id"),
		Address:    address(res, inst),
		ARN:        attrString(attrs, "arn"),
		Source:     resource.SourceDesired,
		Tags:       extractTags(attrs),
		Attributes: attrs,
	}
}

// address reconstructs the Terraform address (state v4 stores module/type/name/
// index_key separately rather than a pre-built address string).
func address(res v4Resource, inst v4Instance) string {
	addr := res.Type + "." + res.Name
	switch k := inst.IndexKey.(type) {
	case nil:
		// singleton; no suffix
	case string:
		addr += fmt.Sprintf("[%q]", k)
	case json.Number:
		addr += "[" + k.String() + "]"
	case float64:
		addr += "[" + strconv.Itoa(int(k)) + "]"
	}
	if res.Module != "" {
		addr = res.Module + "." + addr
	}
	return addr
}

func attrString(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

func extractTags(m map[string]any) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"tags", "tags_all"} {
		if raw, ok := m[key].(map[string]any); ok {
			for k, v := range raw {
				if s, ok := v.(string); ok {
					out[k] = s
				}
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// redactSensitive replaces values flagged in sensitive_attributes with a
// placeholder. State stores secrets in plaintext; we never surface or log them.
// Best-effort for M0: handles single-step get_attr paths (top-level attributes).
func redactSensitive(attrs map[string]any, paths []json.RawMessage) {
	for _, p := range paths {
		var steps []struct {
			Type  string `json:"type"`
			Value any    `json:"value"`
		}
		if err := json.Unmarshal(p, &steps); err != nil || len(steps) != 1 {
			continue
		}
		if steps[0].Type != "get_attr" {
			continue
		}
		if key, ok := steps[0].Value.(string); ok {
			if _, exists := attrs[key]; exists {
				attrs[key] = "<sensitive>"
			}
		}
	}
}

package main

import (
	"reflect"
	"testing"

	"github.com/stephan/rinku/internal/types"
)

func TestBuildIndexes(t *testing.T) {
	libs := map[string]types.Library{
		"go:spf13/cobra": {
			URL:  "https://github.com/spf13/cobra",
			Lang: "go",
		},
		"go:golang/net": {
			URL:    "https://github.com/golang/net",
			Lang:   "go",
			Unsafe: "14 vulns",
		},
		"rust:clap-rs/clap": {
			URL:  "https://github.com/clap-rs/clap",
			Lang: "rust",
		},
		"rust:hyperium/hyper": {
			URL:  "https://github.com/hyperium/hyper",
			Lang: "rust",
		},
	}

	mappings := []types.Mapping{
		{
			Source:   "go:spf13/cobra",
			Targets:  []string{"rust:clap-rs/clap"},
			Category: "cli",
		},
		{
			Source:   "go:golang/net",
			Targets:  []string{"rust:hyperium/hyper"},
			Category: "http_client",
		},
	}

	result := BuildIndexes(libs, mappings)

	// Check counts
	if result.UnsafeCount != 1 {
		t.Errorf("UnsafeCount = %d, want 1", result.UnsafeCount)
	}
	if result.LibrariesCount != 4 {
		t.Errorf("LibrariesCount = %d, want 4", result.LibrariesCount)
	}
	if result.MappingsCount != 2 {
		t.Errorf("MappingsCount = %d, want 2", result.MappingsCount)
	}

	// Check forward index (safe only)
	wantForward := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
	}
	if !reflect.DeepEqual(result.Forward, wantForward) {
		t.Errorf("Forward = %v, want %v", result.Forward, wantForward)
	}

	// Check forward index (all including unsafe)
	wantForwardAll := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
		"rust:github.com/golang/net":  {"https://github.com/hyperium/hyper"},
	}
	if !reflect.DeepEqual(result.ForwardAll, wantForwardAll) {
		t.Errorf("ForwardAll = %v, want %v", result.ForwardAll, wantForwardAll)
	}

	// Check reverse index (safe only)
	wantReverse := map[string][]string{
		"go:github.com/clap-rs/clap": {"https://github.com/spf13/cobra"},
	}
	if !reflect.DeepEqual(result.Reverse, wantReverse) {
		t.Errorf("Reverse = %v, want %v", result.Reverse, wantReverse)
	}

	// Check reverse index (all including unsafe)
	wantReverseAll := map[string][]string{
		"go:github.com/clap-rs/clap":   {"https://github.com/spf13/cobra"},
		"go:github.com/hyperium/hyper": {"https://github.com/golang/net"},
	}
	if !reflect.DeepEqual(result.ReverseAll, wantReverseAll) {
		t.Errorf("ReverseAll = %v, want %v", result.ReverseAll, wantReverseAll)
	}
}

func TestBuildIndexes_NormalizesURLs(t *testing.T) {
	libs := map[string]types.Library{
		"go:Foo/Bar": {
			URL:  "HTTPS://GitHub.com/Foo/Bar/",
			Lang: "go",
		},
		"rust:example/lib": {
			URL:  "https://example.com",
			Lang: "rust",
		},
	}

	mappings := []types.Mapping{
		{
			Source:  "go:Foo/Bar",
			Targets: []string{"rust:example/lib"},
		},
	}

	result := BuildIndexes(libs, mappings)

	// Should normalize to lowercase, no prefix, no trailing slash
	if _, ok := result.Forward["rust:github.com/foo/bar"]; !ok {
		t.Errorf("expected normalized key 'rust:github.com/foo/bar', got keys: %v", result.Forward)
	}
}

func TestBuildIndexes_SkipsNonePlaceholder(t *testing.T) {
	libs := map[string]types.Library{
		"go:foo/bar": {
			URL:  "https://github.com/foo/bar",
			Lang: "go",
		},
	}

	mappings := []types.Mapping{
		{
			Source:  "go:foo/bar",
			Targets: []string{"<None>"},
		},
	}

	result := BuildIndexes(libs, mappings)

	// Should not have any forward or reverse entries for <None>
	if len(result.Forward) != 0 {
		t.Errorf("Forward should be empty, got: %v", result.Forward)
	}
	if len(result.Reverse) != 0 {
		t.Errorf("Reverse should be empty, got: %v", result.Reverse)
	}
}

func TestBuildIndexes_MultipleTargets(t *testing.T) {
	libs := map[string]types.Library{
		"go:foo/bar": {
			URL:  "https://github.com/foo/bar",
			Lang: "go",
		},
		"rust:target1/lib": {
			URL:  "https://github.com/target1/lib",
			Lang: "rust",
		},
		"rust:target2/lib": {
			URL:  "https://github.com/target2/lib",
			Lang: "rust",
		},
	}

	mappings := []types.Mapping{
		{
			Source:  "go:foo/bar",
			Targets: []string{"rust:target1/lib", "rust:target2/lib"},
		},
	}

	result := BuildIndexes(libs, mappings)

	// Forward index should have both targets
	wantForward := []string{
		"https://github.com/target1/lib",
		"https://github.com/target2/lib",
	}
	if got := result.Forward["rust:github.com/foo/bar"]; !reflect.DeepEqual(got, wantForward) {
		t.Errorf("Forward = %v, want %v", got, wantForward)
	}

	// Reverse index should have entries for both targets
	if got := result.Reverse["go:github.com/target1/lib"]; !reflect.DeepEqual(got, []string{"https://github.com/foo/bar"}) {
		t.Errorf("Reverse[target1] = %v, want [https://github.com/foo/bar]", got)
	}
	if got := result.Reverse["go:github.com/target2/lib"]; !reflect.DeepEqual(got, []string{"https://github.com/foo/bar"}) {
		t.Errorf("Reverse[target2] = %v, want [https://github.com/foo/bar]", got)
	}
}

func TestBuildIndexes_UnsafeTarget(t *testing.T) {
	libs := map[string]types.Library{
		"go:safe/source": {
			URL:  "https://github.com/safe/source",
			Lang: "go",
		},
		"rust:unsafe/target": {
			URL:    "https://github.com/unsafe/target",
			Lang:   "rust",
			Unsafe: "has vulnerabilities",
		},
	}

	mappings := []types.Mapping{
		{
			Source:  "go:safe/source",
			Targets: []string{"rust:unsafe/target"},
		},
	}

	result := BuildIndexes(libs, mappings)

	// Safe source to unsafe target should only appear in *All indexes
	if len(result.Forward) != 0 {
		t.Errorf("Forward should be empty (target is unsafe), got: %v", result.Forward)
	}
	if len(result.Reverse) != 0 {
		t.Errorf("Reverse should be empty (target is unsafe), got: %v", result.Reverse)
	}

	// But should appear in All indexes
	if len(result.ForwardAll) != 1 {
		t.Errorf("ForwardAll should have 1 entry, got: %v", result.ForwardAll)
	}
	if len(result.ReverseAll) != 1 {
		t.Errorf("ReverseAll should have 1 entry, got: %v", result.ReverseAll)
	}
}

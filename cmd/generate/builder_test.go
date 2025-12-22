package main

import (
	"reflect"
	"testing"
)

func TestBuildIndexes(t *testing.T) {
	mappings := []Mapping{
		{
			SourceLang: "go",
			TargetLang: "rust",
			Category:   "cli",
			Source:     "https://github.com/spf13/cobra",
			Target:     []string{"https://github.com/clap-rs/clap"},
		},
		{
			SourceLang: "go",
			TargetLang: "rust",
			Category:   "http_client",
			Source:     "https://github.com/golang/net",
			Target:     []string{"https://github.com/hyperium/hyper"},
			Disabled:   "14 vulns",
		},
	}

	index, indexAll, disabledCount := BuildIndexes(mappings)

	// Check disabled count
	if disabledCount != 1 {
		t.Errorf("disabledCount = %d, want 1", disabledCount)
	}

	// Check index (enabled only)
	wantIndex := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
	}
	if !reflect.DeepEqual(index, wantIndex) {
		t.Errorf("index = %v, want %v", index, wantIndex)
	}

	// Check indexAll (includes disabled)
	wantIndexAll := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
		"rust:github.com/golang/net":  {"https://github.com/hyperium/hyper"},
	}
	if !reflect.DeepEqual(indexAll, wantIndexAll) {
		t.Errorf("indexAll = %v, want %v", indexAll, wantIndexAll)
	}
}

func TestBuildIndexes_NormalizesURLs(t *testing.T) {
	mappings := []Mapping{
		{
			TargetLang: "rust",
			Source:     "HTTPS://GitHub.com/Foo/Bar/",
			Target:     []string{"https://example.com"},
		},
	}

	index, _, _ := BuildIndexes(mappings)

	// Should normalize to lowercase, no prefix, no trailing slash
	if _, ok := index["rust:github.com/foo/bar"]; !ok {
		t.Errorf("expected normalized key 'rust:github.com/foo/bar', got keys: %v", index)
	}
}

func TestBuildIndexes_ConsolidationAndCase(t *testing.T) {
	mappings := []Mapping{
		{
			TargetLang: "Rust",
			Source:     "github.com/foo/bar",
			Target:     []string{"target1"},
		},
		{
			TargetLang: "rust",
			Source:     "https://github.com/foo/bar",
			Target:     []string{"target2"},
		},
	}

	index, _, _ := BuildIndexes(mappings)

	// Should lowercase "Rust" to "rust" and combine targets
	want := []string{"target1", "target2"}
	if got := index["rust:github.com/foo/bar"]; !reflect.DeepEqual(got, want) {
		t.Errorf("index[\"rust:github.com/foo/bar\"] = %v, want %v", got, want)
	}
}

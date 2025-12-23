package gomod

import (
	"strings"
	"testing"
)

func TestParseReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *ParseResult
		wantErr bool
	}{
		{
			name: "single line require",
			input: `module example.com/test
go 1.21
require github.com/spf13/cobra v1.8.0`,
			want: &ParseResult{
				Module:    "example.com/test",
				GoVersion: "1.21",
				Dependencies: []Dependency{
					{Path: "github.com/spf13/cobra", Version: "v1.8.0", Indirect: false},
				},
			},
		},
		{
			name: "single line require with indirect",
			input: `module example.com/test
go 1.21
require github.com/spf13/cobra v1.8.0 // indirect`,
			want: &ParseResult{
				Module:    "example.com/test",
				GoVersion: "1.21",
				Dependencies: []Dependency{
					{Path: "github.com/spf13/cobra", Version: "v1.8.0", Indirect: true},
				},
			},
		},
		{
			name: "block require",
			input: `module test
go 1.22
require (
	github.com/foo/bar v1.0.0
	github.com/baz/qux v2.0.0
)`,
			want: &ParseResult{
				Module:    "test",
				GoVersion: "1.22",
				Dependencies: []Dependency{
					{Path: "github.com/foo/bar", Version: "v1.0.0", Indirect: false},
					{Path: "github.com/baz/qux", Version: "v2.0.0", Indirect: false},
				},
			},
		},
		{
			name: "block require with indirect",
			input: `module test
go 1.22
require (
	github.com/foo/bar v1.0.0
	github.com/baz/qux v2.0.0 // indirect
)`,
			want: &ParseResult{
				Module:    "test",
				GoVersion: "1.22",
				Dependencies: []Dependency{
					{Path: "github.com/foo/bar", Version: "v1.0.0", Indirect: false},
					{Path: "github.com/baz/qux", Version: "v2.0.0", Indirect: true},
				},
			},
		},
		{
			name: "mixed single and block require",
			input: `module test
go 1.22
require github.com/single/dep v1.0.0

require (
	github.com/block/dep v2.0.0
)`,
			want: &ParseResult{
				Module:    "test",
				GoVersion: "1.22",
				Dependencies: []Dependency{
					{Path: "github.com/single/dep", Version: "v1.0.0", Indirect: false},
					{Path: "github.com/block/dep", Version: "v2.0.0", Indirect: false},
				},
			},
		},
		{
			name: "versioned module path",
			input: `module test
go 1.22
require github.com/foo/bar/v2 v2.5.0`,
			want: &ParseResult{
				Module:    "test",
				GoVersion: "1.22",
				Dependencies: []Dependency{
					{Path: "github.com/foo/bar/v2", Version: "v2.5.0", Indirect: false},
				},
			},
		},
		{
			name: "empty go.mod",
			input: `module test
go 1.22`,
			want: &ParseResult{
				Module:       "test",
				GoVersion:    "1.22",
				Dependencies: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReader(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Module != tt.want.Module {
				t.Errorf("Module = %v, want %v", got.Module, tt.want.Module)
			}

			if got.GoVersion != tt.want.GoVersion {
				t.Errorf("GoVersion = %v, want %v", got.GoVersion, tt.want.GoVersion)
			}

			if len(got.Dependencies) != len(tt.want.Dependencies) {
				t.Errorf("Dependencies count = %d, want %d", len(got.Dependencies), len(tt.want.Dependencies))
				return
			}

			for i, dep := range got.Dependencies {
				wantDep := tt.want.Dependencies[i]
				if dep.Path != wantDep.Path {
					t.Errorf("Dependency[%d].Path = %v, want %v", i, dep.Path, wantDep.Path)
				}
				if dep.Version != wantDep.Version {
					t.Errorf("Dependency[%d].Version = %v, want %v", i, dep.Version, wantDep.Version)
				}
				if dep.Indirect != wantDep.Indirect {
					t.Errorf("Dependency[%d].Indirect = %v, want %v", i, dep.Indirect, wantDep.Indirect)
				}
			}
		})
	}
}

func TestDirectDependencies(t *testing.T) {
	result := &ParseResult{
		Dependencies: []Dependency{
			{Path: "github.com/foo/bar", Version: "v1.0.0", Indirect: false},
			{Path: "github.com/baz/qux", Version: "v2.0.0", Indirect: true},
			{Path: "github.com/direct/dep", Version: "v3.0.0", Indirect: false},
		},
	}

	direct := result.DirectDependencies()

	if len(direct) != 2 {
		t.Errorf("DirectDependencies() count = %d, want 2", len(direct))
	}

	if direct[0].Path != "github.com/foo/bar" {
		t.Errorf("DirectDependencies()[0].Path = %v, want github.com/foo/bar", direct[0].Path)
	}

	if direct[1].Path != "github.com/direct/dep" {
		t.Errorf("DirectDependencies()[1].Path = %v, want github.com/direct/dep", direct[1].Path)
	}
}

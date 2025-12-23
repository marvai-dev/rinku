// Package cargo provides Cargo.toml generation functionality.
package cargo

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/stephan/rinku/internal/gomod"
	urlpkg "github.com/stephan/rinku/internal/url"
)

// Lookup is an interface for looking up Rust equivalents.
type Lookup interface {
	Lookup(sourceURL, targetLang string, unsafe bool) []string
}

// MappedDependency represents a Go dependency mapped to Rust.
type MappedDependency struct {
	GoDep       gomod.Dependency
	RustTargets []string // GitHub URLs of Rust equivalents
	CrateNames  []string // Corresponding crate names
}

// UnmappedDependency represents a Go dependency with no Rust mapping.
type UnmappedDependency struct {
	GoDep gomod.Dependency
}

// GenerateResult contains the categorized dependencies.
type GenerateResult struct {
	Mapped   []MappedDependency
	Unmapped []UnmappedDependency
}

// MapDependencies maps Go dependencies to Rust equivalents.
func MapDependencies(deps []gomod.Dependency, lookup Lookup, unsafe bool) *GenerateResult {
	result := &GenerateResult{}

	for _, dep := range deps {
		// Convert module path to GitHub URL
		ghURL := ModulePathToGitHubURL(dep.Path)

		// Look up Rust equivalents
		rustURLs := lookup.Lookup(ghURL, "rust", unsafe)

		if len(rustURLs) > 0 {
			mapped := MappedDependency{
				GoDep:       dep,
				RustTargets: rustURLs,
			}

			// Extract crate names
			for _, rustURL := range rustURLs {
				crateName := ExtractCrateName(rustURL)
				mapped.CrateNames = append(mapped.CrateNames, crateName)
			}

			result.Mapped = append(result.Mapped, mapped)
		} else {
			result.Unmapped = append(result.Unmapped, UnmappedDependency{
				GoDep: dep,
			})
		}
	}

	return result
}

// ModulePathToGitHubURL converts a Go module path to a GitHub URL.
func ModulePathToGitHubURL(path string) string {
	// Handle golang.org/x/... -> github.com/golang/...
	if strings.HasPrefix(path, "golang.org/x/") {
		pkg := strings.TrimPrefix(path, "golang.org/x/")
		// Remove version suffix like /v2
		if idx := strings.Index(pkg, "/"); idx != -1 {
			pkg = pkg[:idx]
		}
		return "https://github.com/golang/" + pkg
	}

	// Handle github.com paths
	if strings.HasPrefix(path, "github.com/") {
		// Remove version suffix /vN at the end
		path = stripVersionSuffix(path)
		return "https://" + path
	}

	// For other paths, try as-is with https://
	return "https://" + path
}

// stripVersionSuffix removes /vN version suffix from module paths.
func stripVersionSuffix(path string) string {
	// Match /v2, /v3, etc. at the end
	re := regexp.MustCompile(`/v\d+$`)
	return re.ReplaceAllString(path, "")
}

// ExtractCrateName extracts a Rust crate name from a GitHub URL.
func ExtractCrateName(githubURL string) string {
	normalized := urlpkg.Normalize(githubURL)

	// Check known mappings first
	if name, ok := knownCrateNames[normalized]; ok {
		return name
	}

	// Parse: github.com/owner/repo or github.com/owner/repo/tree/...
	parts := strings.Split(normalized, "/")
	if len(parts) < 3 || parts[0] != "github.com" {
		return ""
	}

	repoName := parts[2]

	// Handle subpaths like github.com/tokio-rs/tracing/tree/master/tracing-appender
	if len(parts) >= 5 && parts[3] == "tree" {
		// Use the last path component as crate name
		repoName = parts[len(parts)-1]
	}

	// Common transformations:
	// - Remove -rs suffix (common in Rust repos)
	// - Replace hyphens with underscores for crate names
	crateName := repoName
	crateName = strings.TrimSuffix(crateName, "-rs")
	crateName = strings.ReplaceAll(crateName, "-", "_")

	return crateName
}

// knownCrateNames maps GitHub URLs to known crate names.
var knownCrateNames = map[string]string{
	"github.com/serde-rs/json":               "serde_json",
	"github.com/serde-rs/serde":              "serde",
	"github.com/dtolnay/serde-yaml":          "serde_yaml",
	"github.com/dtolnay/anyhow":              "anyhow",
	"github.com/tokio-rs/tokio":              "tokio",
	"github.com/tokio-rs/axum":               "axum",
	"github.com/tokio-rs/tracing":            "tracing",
	"github.com/clap-rs/clap":                "clap",
	"github.com/hyperium/hyper":              "hyper",
	"github.com/rust-lang/regex":             "regex",
	"github.com/chronotope/chrono":           "chrono",
	"github.com/uuid-rs/uuid":                "uuid",
	"github.com/rayon-rs/rayon":              "rayon",
	"github.com/crossbeam-rs/crossbeam":      "crossbeam",
	"github.com/rusqlite/rusqlite":           "rusqlite",
	"github.com/launchbadge/sqlx":            "sqlx",
	"github.com/serenity-rs/serenity":        "serenity",
	"github.com/actix/actix-web":             "actix_web",
	"github.com/rustls/rustls":               "rustls",
	"github.com/image-rs/image":              "image",
	"github.com/burntsushi/toml":             "toml",
	"github.com/toml-rs/toml":                "toml",
	"github.com/rust-lang/log":               "log",
	"github.com/seaorm/sea-orm":              "sea_orm",
	"github.com/hyperium/tonic":              "tonic",
	"github.com/tokio-rs/prost":              "prost",
	"github.com/sfackler/rust-postgres":      "postgres",
	"github.com/rust-lang/hashbrown":         "hashbrown",
	"github.com/tower-rs/tower":              "tower",
	"github.com/tower-rs/tower-http":         "tower_http",
	"github.com/rust-random/rand":            "rand",
	"github.com/bytecodealliance/wasmtime":   "wasmtime",
	"github.com/rust-rocksdb/rust-rocksdb":   "rocksdb",
	"github.com/redis-rs/redis-rs":           "redis",
	"github.com/mongodb/mongo-rust-driver":   "mongodb",
	"github.com/awslabs/aws-sdk-rust":        "aws_sdk_config",
	"github.com/azure/azure-sdk-for-rust":    "azure_core",
	"github.com/googleapis/google-cloud-rust": "google_cloud_storage",
}

// GenerateCargoToml writes a Cargo.toml to the provided writer.
func GenerateCargoToml(w io.Writer, moduleName string, result *GenerateResult) error {
	// Write header
	fmt.Fprintln(w, "# Generated by rinku - https://github.com/marvai-dev/rinku")
	fmt.Fprintf(w, "# Original Go module: %s\n", moduleName)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "[package]")
	fmt.Fprintln(w, `name = "converted_project"`)
	fmt.Fprintln(w, `version = "0.1.0"`)
	fmt.Fprintln(w, `edition = "2021"`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "[dependencies]")

	// Sort mapped dependencies by crate name
	sort.Slice(result.Mapped, func(i, j int) bool {
		if len(result.Mapped[i].CrateNames) > 0 && len(result.Mapped[j].CrateNames) > 0 {
			return result.Mapped[i].CrateNames[0] < result.Mapped[j].CrateNames[0]
		}
		return false
	})

	// Write mapped dependencies
	for _, mapped := range result.Mapped {
		for i, crateName := range mapped.CrateNames {
			fmt.Fprintf(w, "%s = \"*\"  # from %s -> %s\n",
				crateName, mapped.GoDep.Path, mapped.RustTargets[i])
		}
	}

	// Write unmapped dependencies as TODO comments
	if len(result.Unmapped) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "# TODO: Find equivalents for these Go dependencies:")
		for _, unmapped := range result.Unmapped {
			fmt.Fprintf(w, "# TODO: find equivalent for %s\n", unmapped.GoDep.Path)
		}
	}

	return nil
}

// WriteCargoTomlFS writes a Cargo.toml file to the given filesystem.
func WriteCargoTomlFS(fs afero.Fs, path string, moduleName string, result *GenerateResult) error {
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return GenerateCargoToml(file, moduleName, result)
}

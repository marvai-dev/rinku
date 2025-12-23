package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/stephan/rinku/internal/cargo"
	"github.com/stephan/rinku/internal/gomod"
	"github.com/stephan/rinku/internal/rinku"
)

//go:generate go run ../generate

const description = `Rinku finds equivalent Rust libraries for Go dependencies.

COMMANDS:
  scan <go.mod>                 Parse go.mod and show Rust equivalents for each dependency
  convert <go.mod> [-o file]    Generate a Cargo.toml file from go.mod
  <url>                         Look up Rust equivalent for a single GitHub URL

EXAMPLES:
  # Look up a single library
  rinku https://github.com/spf13/cobra
  # Output: https://github.com/clap-rs/clap

  # Scan a go.mod file and list all mappings
  rinku scan go.mod
  # Output: Lists each dependency with its Rust equivalent(s)

  # Generate Cargo.toml from go.mod (prints to stdout)
  rinku convert go.mod

  # Generate Cargo.toml to a file
  rinku convert go.mod -o Cargo.toml

  # Include libraries with known vulnerabilities
  rinku scan go.mod --unsafe
  rinku convert go.mod --unsafe

OUTPUT FORMATS:
  lookup:   Prints GitHub URL(s) of Rust equivalent(s), one per line
  scan:     Prints each Go dependency followed by its Rust mapping(s)
  convert:  Prints valid Cargo.toml with [dependencies] section

EXIT CODES:
  0  Success
  1  Error (invalid input, file not found, etc.)

MORE INFO:
  Repository: https://github.com/marvai-dev/rinku
  Database:   160+ Go-to-Rust library mappings across 140+ categories`

// CLI defines the command-line interface with subcommands.
var CLI struct {
	Scan    ScanCmd    `cmd:"" help:"Parse go.mod and show Rust equivalents for each dependency. Outputs: module name, go version, each dependency with its Rust crate name(s) and URL(s), summary of mapped/total count."`
	Convert ConvertCmd `cmd:"" help:"Generate a Cargo.toml file from go.mod. Mapped dependencies use version \"*\". Unmapped dependencies are listed as TODO comments."`
	Lookup  LookupCmd  `cmd:"" default:"withargs" help:"Look up Rust equivalent for a single GitHub URL. Outputs the Rust library URL(s), one per line. Returns empty if no mapping exists."`
}

// LookupCmd handles the original URL lookup behavior.
type LookupCmd struct {
	URL      string `arg:"" help:"GitHub URL of the Go library (e.g., https://github.com/spf13/cobra)."`
	Language string `arg:"" optional:"" default:"rust" help:"Target language (default: rust). Currently only 'rust' is supported."`
	Unsafe   bool   `help:"Include libraries with known security vulnerabilities in results."`
}

// ScanCmd handles scanning go.mod files.
type ScanCmd struct {
	Path   string `arg:"" type:"existingfile" help:"Path to go.mod file to scan."`
	Unsafe bool   `help:"Include libraries with known security vulnerabilities in results."`
}

// ConvertCmd handles generating Cargo.toml from go.mod.
type ConvertCmd struct {
	Path   string `arg:"" type:"existingfile" help:"Path to go.mod file to convert."`
	Output string `short:"o" default:"-" help:"Output file path. Use '-' for stdout (default: -)."`
	Unsafe bool   `help:"Include libraries with known security vulnerabilities in results."`
}

// Run executes the lookup command.
func (c *LookupCmd) Run(r *rinku.Rinku) error {
	results := r.Lookup(c.URL, c.Language, c.Unsafe)
	for _, result := range results {
		fmt.Println(result)
	}
	return nil
}

// Run executes the scan command.
func (c *ScanCmd) Run(r *rinku.Rinku) error {
	result, err := gomod.Parse(c.Path)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod: %w", err)
	}

	fmt.Printf("Module: %s\n", result.Module)
	fmt.Printf("Go version: %s\n", result.GoVersion)

	deps := result.DirectDependencies()
	fmt.Printf("Direct dependencies: %d\n\n", len(deps))

	mapped := 0
	for _, dep := range deps {
		ghURL := cargo.ModulePathToGitHubURL(dep.Path)
		rustURLs := r.Lookup(ghURL, "rust", c.Unsafe)

		fmt.Printf("%s\n", dep.Path)
		if len(rustURLs) > 0 {
			mapped++
			for _, rustURL := range rustURLs {
				crateName := cargo.ExtractCrateName(rustURL)
				fmt.Printf("  -> %s (%s)\n", crateName, rustURL)
			}
		} else {
			fmt.Printf("  -> (no mapping found)\n")
		}
	}

	fmt.Printf("\nMapped %d/%d direct dependencies\n", mapped, len(deps))
	return nil
}

// Run executes the convert command.
func (c *ConvertCmd) Run(r *rinku.Rinku) error {
	result, err := gomod.Parse(c.Path)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod: %w", err)
	}

	deps := result.DirectDependencies()
	genResult := cargo.MapDependencies(deps, r, c.Unsafe)

	// Determine output writer
	var w *os.File
	if c.Output == "-" {
		w = os.Stdout
	} else {
		w, err = os.Create(c.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer w.Close()
	}

	if err := cargo.GenerateCargoToml(w, result.Module, genResult); err != nil {
		return fmt.Errorf("failed to generate Cargo.toml: %w", err)
	}

	if c.Output != "-" {
		fmt.Fprintf(os.Stderr, "Generated %s with %d dependencies (%d mapped, %d unmapped)\n",
			c.Output, len(deps), len(genResult.Mapped), len(genResult.Unmapped))
	}

	return nil
}

func main() {
	// Show comprehensive help if no arguments provided
	if len(os.Args) == 1 {
		fmt.Println(description)
		os.Exit(0)
	}

	r := rinku.New(index, indexAll, reverseIndex, reverseIndexAll)

	ctx := kong.Parse(&CLI,
		kong.Name("rinku"),
		kong.Description("Find equivalent Rust libraries for Go dependencies."),
		kong.UsageOnError(),
		kong.Bind(r),
	)

	err := ctx.Run(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

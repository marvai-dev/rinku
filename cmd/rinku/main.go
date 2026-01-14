package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var CLI struct {
	Scan    ScanCmd    `cmd:"" help:"Parse go.mod and show Rust equivalents for each dependency."`
	Convert ConvertCmd `cmd:"" help:"Generate a Cargo.toml file from go.mod."`
	Lookup  LookupCmd  `cmd:"" default:"withargs" help:"Look up equivalent for a single GitHub URL."`
}

type LookupCmd struct {
	URL      string `arg:"" help:"GitHub URL of the library."`
	Language string `arg:"" optional:"" default:"rust" help:"Target language (default: rust)."`
	Unsafe   bool   `help:"Include libraries with known vulnerabilities."`
}

type ScanCmd struct {
	Path   string `arg:"" type:"existingfile" help:"Path to go.mod file."`
	Unsafe bool   `help:"Include libraries with known vulnerabilities."`
}

type ConvertCmd struct {
	Path   string `arg:"" type:"existingfile" help:"Path to go.mod file."`
	Output string `short:"o" default:"-" help:"Output file (- for stdout)."`
	Unsafe bool   `help:"Include libraries with known vulnerabilities."`
}

func (c *LookupCmd) Run(r *rinku.Rinku) error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		return fmt.Errorf("invalid URL: must start with http:// or https://")
	}
	for _, result := range r.Lookup(c.URL, c.Language, c.Unsafe) {
		fmt.Println(result)
	}
	return nil
}

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
				crateName := r.CrateName(rustURL)
				if crateName == "" {
					crateName = cargo.ExtractCrateName(rustURL)
				}
				fmt.Printf("  -> %s (%s)\n", crateName, rustURL)
			}
		} else {
			fmt.Printf("  -> (no mapping found)\n")
		}
	}

	fmt.Printf("\nMapped %d/%d direct dependencies\n", mapped, len(deps))
	return nil
}

func (c *ConvertCmd) Run(r *rinku.Rinku) (err error) {
	result, err := gomod.Parse(c.Path)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod: %w", err)
	}

	deps := result.DirectDependencies()
	genResult := cargo.MapDependencies(deps, r, c.Unsafe)

	var w *os.File
	if c.Output == "-" {
		w = os.Stdout
	} else {
		if err := validateOutputPath(c.Output); err != nil {
			return err
		}
		w, err = os.Create(c.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := w.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("failed to close output file: %w", cerr)
			}
		}()
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

func validateOutputPath(path string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed: %s", path)
	}
	if strings.HasPrefix(filepath.Clean(path), "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}
	return nil
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(description)
		os.Exit(0)
	}

	r := rinku.New(index, indexAll, reverseIndex, reverseIndexAll, knownCrateNames)

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

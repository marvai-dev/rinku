package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/stephan/rinku/internal/cargo"
	"github.com/stephan/rinku/internal/gomod"
	"github.com/stephan/rinku/internal/progress"
	"github.com/stephan/rinku/internal/prompt"
	"github.com/stephan/rinku/internal/requirements"
	"github.com/stephan/rinku/internal/rinku"
	"github.com/stephan/rinku/internal/types"
)

//go:generate go run ../generate

const description = `Rinku - Go to Rust library mapper

USAGE:
  rinku <command> [arguments] [flags]

COMMANDS:
  rinku <github-url>                    Look up Rust equivalent for a Go library
  rinku scan <path-to-go.mod>           List Rust equivalents for all dependencies
  rinku convert <path-to-go.mod>        Generate Cargo.toml from go.mod

FLAGS:
  --unsafe    Include libraries with known security vulnerabilities
  -o <file>   Output file for convert command (default: stdout)
  --help      Show this help message

EXAMPLES:
  rinku https://github.com/spf13/cobra
  Output: https://github.com/clap-rs/clap

  rinku https://github.com/sirupsen/logrus
  Output: https://github.com/rust-lang/log

  rinku scan go.mod
  Output:
    Module: myproject
    Go version: 1.21
    Direct dependencies: 5

    github.com/spf13/cobra
      -> clap (https://github.com/clap-rs/clap)
    github.com/sirupsen/logrus
      -> log (https://github.com/rust-lang/log)
    ...

  rinku convert go.mod
  Output: [Cargo.toml content to stdout]

  rinku convert go.mod -o Cargo.toml
  Output: Writes Cargo.toml file

EXIT CODES:
  0  Success
  1  Error (invalid input, file not found, no mapping found)

NOTES:
  - Input URLs must be full GitHub URLs (https://github.com/owner/repo)
  - The database contains 160+ Go-to-Rust mappings across 140+ categories
  - Use --unsafe only if you need libraries flagged for vulnerabilities

Repository: https://github.com/marvai-dev/rinku`

var CLI struct {
	Scan    ScanCmd    `cmd:"" help:"Parse go.mod and show Rust equivalents for each dependency."`
	Convert ConvertCmd `cmd:"" help:"Generate a Cargo.toml file from go.mod."`
	Analyze AnalyzeCmd `cmd:"" help:"Analyze go.mod and output detected project type tags."`
	Migrate MigrateCmd `cmd:"" help:"Output migration workflow steps."`
	Req     ReqCmd     `cmd:"" help:"Manage migration requirements."`
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

type AnalyzeCmd struct {
	Path string `arg:"" type:"existingfile" help:"Path to go.mod file."`
}

type ConvertCmd struct {
	Path   string `arg:"" type:"existingfile" help:"Path to go.mod file."`
	Output string `short:"o" default:"-" help:"Output file (- for stdout)."`
	Unsafe bool   `help:"Include libraries with known vulnerabilities."`
}

type MigrateCmd struct {
	Step   string `arg:"" optional:"" help:"Step ID to retrieve."`
	Start  string `help:"Mark step as in_progress."`
	Finish string `help:"Mark step as completed."`
	Status bool   `help:"Show current migration status."`
	Reset  bool   `help:"Reset migration progress."`
	Note   string `help:"Add note when finishing a step."`
}

type ReqCmd struct {
	Set  ReqSetCmd  `cmd:"" help:"Set a requirement."`
	Get  ReqGetCmd  `cmd:"" help:"Get a requirement."`
	List ReqListCmd `cmd:"" help:"List requirements."`
	Done ReqDoneCmd `cmd:"" help:"Mark a requirement as done."`
}

type ReqSetCmd struct {
	Path    string `arg:"" help:"Requirement path (e.g., api/cli)."`
	Content string `arg:"" optional:"" help:"Requirement content (reads from stdin if omitted)."`
}

type ReqGetCmd struct {
	Path string `arg:"" help:"Requirement path."`
}

type ReqListCmd struct {
	Prefix string `arg:"" optional:"" help:"Optional prefix filter."`
}

type ReqDoneCmd struct {
	Path string `arg:"" help:"Requirement path."`
}

func (c *ReqSetCmd) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	content := c.Content
	if content == "" {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading from stdin: %w", err)
		}
		content = strings.TrimSpace(string(data))
	}

	if content == "" {
		return fmt.Errorf("content is required (provide as argument or via stdin)")
	}

	if err := requirements.Set(cwd, c.Path, content); err != nil {
		return fmt.Errorf("setting requirement: %w", err)
	}
	fmt.Printf("Set %s\n", c.Path)
	return nil
}

func (c *ReqGetCmd) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	req, err := requirements.Get(cwd, c.Path)
	if err != nil {
		return fmt.Errorf("getting requirement: %w", err)
	}
	if req == nil {
		return fmt.Errorf("requirement '%s' not found", c.Path)
	}
	fmt.Println(req.Content)
	return nil
}

func (c *ReqListCmd) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	paths, err := requirements.List(cwd, c.Prefix)
	if err != nil {
		return fmt.Errorf("listing requirements: %w", err)
	}

	if len(paths) == 0 {
		fmt.Println("No requirements found.")
		return nil
	}

	for _, p := range paths {
		req, _ := requirements.Get(cwd, p)
		if req != nil && req.Done {
			fmt.Printf("[x] %s\n", p)
		} else {
			fmt.Printf("[ ] %s\n", p)
		}
	}
	return nil
}

func (c *ReqDoneCmd) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if err := requirements.Done(cwd, c.Path); err != nil {
		return err
	}
	fmt.Printf("Marked %s as done\n", c.Path)
	return nil
}

func (c *LookupCmd) Run(r *rinku.Rinku) error {
	if c.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if !isValidURL(c.URL) {
		return fmt.Errorf("invalid URL: must start with http:// or https://")
	}
	for _, result := range r.Lookup(c.URL, c.Language, c.Unsafe) {
		fmt.Println(result)
	}
	return nil
}

func (c *MigrateCmd) Run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	p, err := prompt.Migration()
	if err != nil {
		return fmt.Errorf("failed to load migration prompt: %w", err)
	}

	// Handle --reset first
	if c.Reset {
		if err := progress.Delete(cwd); err != nil {
			return fmt.Errorf("resetting progress: %w", err)
		}
		fmt.Println("Migration progress reset.")
		return nil
	}

	// Load or create progress
	m, err := progress.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading progress: %w", err)
	}
	if m == nil {
		m = progress.New(cwd, p.Steps())
		if err := m.Save(cwd); err != nil {
			return fmt.Errorf("saving initial progress: %w", err)
		}
	}

	// Handle --status
	if c.Status {
		showMigrationStatus(m)
		return nil
	}

	// Handle --start <step>
	if c.Start != "" {
		if err := m.StartStep(c.Start); err != nil {
			return err
		}
		if err := m.Save(cwd); err != nil {
			return fmt.Errorf("saving progress: %w", err)
		}
		content, ok := p.GetStep(c.Start)
		if !ok {
			return fmt.Errorf("step '%s' not found", c.Start)
		}
		// Show Before section if present
		if before := p.Before(); before != "" {
			fmt.Println(before)
			fmt.Println()
		}
		fmt.Println(content)
		// Show After section if present
		if after := p.After(); after != "" {
			fmt.Println()
			fmt.Println(after)
		}
		return nil
	}

	// Handle --finish <step>
	if c.Finish != "" {
		if err := m.CompleteStep(c.Finish, c.Note); err != nil {
			return err
		}
		if err := m.Save(cwd); err != nil {
			return fmt.Errorf("saving progress: %w", err)
		}
		fmt.Printf("Completed step %s\n", c.Finish)
		return nil
	}

	// Default: show step content
	// No args = show introduction (entry point)
	// Explicit step = show that step
	if c.Step == "" {
		if intro := p.Introduction(); intro != "" {
			fmt.Println(intro)
			return nil
		}
		// Fallback to first step if no introduction
		c.Step = p.FirstStep()
	}

	content, ok := p.GetStep(c.Step)
	if !ok {
		return fmt.Errorf("step '%s' not found", c.Step)
	}
	fmt.Println(content)
	return nil
}

func showMigrationStatus(m *progress.Migration) {
	completed, total := m.Progress()
	fmt.Printf("Migration Progress: %d/%d steps\n", completed, total)
	fmt.Printf("Current step: %s\n", m.CurrentStep)
	fmt.Printf("Started: %s\n\n", m.StartedAt.Format("2006-01-02 15:04:05"))

	for _, id := range m.StepOrder {
		step := m.Steps[id]
		symbol := statusSymbol(step.Status)
		fmt.Printf("  %s Step %s", symbol, id)
		if step.Status == progress.StepCompleted && step.CompletedAt != nil {
			fmt.Printf(" (completed %s)", step.CompletedAt.Format("Jan 2 15:04"))
		}
		if step.Notes != "" {
			fmt.Printf("\n      Note: %s", step.Notes)
		}
		fmt.Println()
	}
}

func statusSymbol(s progress.StepStatus) string {
	switch s {
	case progress.StepCompleted:
		return "[x]"
	case progress.StepInProgress:
		return "[>]"
	case progress.StepSkipped:
		return "[-]"
	default:
		return "[ ]"
	}
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

func (c *AnalyzeCmd) Run(r *rinku.Rinku) error {
	result, err := gomod.Parse(c.Path)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod: %w", err)
	}

	deps := result.DirectDependencies()
	tagSet := make(map[string]struct{})

	for _, dep := range deps {
		ghURL := cargo.ModulePathToGitHubURL(dep.Path)
		for _, tag := range r.Tags(ghURL) {
			tagSet[tag] = struct{}{}
		}
	}

	// Output unique tags, one per line
	for tag := range tagSet {
		fmt.Println(tag)
	}

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

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func shouldShowHelp(args []string) bool {
	if len(args) == 1 {
		return true
	}
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		return true
	}
	return false
}

func convertRequiredDeps(m map[string][]requiredDep) map[string][]types.RequiredDep {
	result := make(map[string][]types.RequiredDep, len(m))
	for k, deps := range m {
		converted := make([]types.RequiredDep, len(deps))
		for i, d := range deps {
			converted[i] = types.RequiredDep{
				Crate:    d.Crate,
				Features: d.Features,
				Reason:   d.Reason,
			}
		}
		result[k] = converted
	}
	return result
}

func main() {
	if shouldShowHelp(os.Args) {
		fmt.Println(description)
		os.Exit(0)
	}

	r := rinku.New(index, indexAll, reverseIndex, reverseIndexAll, knownCrateNames, tags, convertRequiredDeps(requiredDeps))

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

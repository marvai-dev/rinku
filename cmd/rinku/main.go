package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/stephan/rinku/internal/rinku"
)

//go:generate go run ../generate

// CLI defines the command-line interface
var CLI struct {
	URL      string `arg:"" help:"GitHub URL of the Go library to look up."`
	Language string `arg:"" default:"rust" help:"Target language to find equivalents in (currently only rust is supported)."`
	Unsafe   bool   `help:"Include libraries with known vulnerabilities in results."`
}

func main() {
	kong.Parse(&CLI,
		kong.Name("rinku"),
		kong.Description("Find equivalent Rust libraries for Go dependencies."),
		kong.UsageOnError(),
	)

	r := rinku.New(index, indexAll, reverseIndex, reverseIndexAll)
	results := r.Lookup(CLI.URL, CLI.Language, CLI.Unsafe)

	for _, result := range results {
		fmt.Println(result)
	}
}

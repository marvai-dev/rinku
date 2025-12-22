package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/stephan/rinku/internal/rinku"
)

//go:generate go run ../generate

// CLI defines the command-line interface
var CLI struct {
	GithubURL      string `arg:"" help:"GitHub URL of the library to look up."`
	TargetLanguage string `arg:"" help:"Target language (go, rust)."`
	Unsafe         bool   `help:"Include libraries with known vulnerabilities in results."`
}

func main() {
	kong.Parse(&CLI,
		kong.Name("rinku"),
		kong.Description("Look up equivalent libraries across programming languages."),
		kong.UsageOnError(),
	)

	r := rinku.New(index, indexAll)
	targets := r.Lookup(CLI.GithubURL, CLI.TargetLanguage, CLI.Unsafe)
	for _, target := range targets {
		fmt.Println(target)
	}
}

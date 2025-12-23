// Package gomod provides parsing functionality for go.mod files.
package gomod

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// Dependency represents a Go module dependency.
type Dependency struct {
	Path     string // e.g., "github.com/spf13/cobra"
	Version  string // e.g., "v1.10.2"
	Indirect bool   // true if marked "// indirect"
}

// ParseResult contains parsed go.mod data.
type ParseResult struct {
	Module       string       // e.g., "github.com/stephan/rinku"
	GoVersion    string       // e.g., "1.25.5"
	Dependencies []Dependency // all require dependencies
}

// Parse reads and parses a go.mod file from the given path.
func Parse(path string) (*ParseResult, error) {
	return ParseFS(afero.NewOsFs(), path)
}

// ParseFS reads and parses a go.mod file from the given filesystem.
func ParseFS(fs afero.Fs, path string) (*ParseResult, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseReader(file)
}

// ParseReader parses go.mod content from an io.Reader.
func ParseReader(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}
	scanner := bufio.NewScanner(r)
	inRequireBlock := false

	// Regex patterns
	moduleRe := regexp.MustCompile(`^module\s+(\S+)`)
	goVersionRe := regexp.MustCompile(`^go\s+(\S+)`)
	requireSingleRe := regexp.MustCompile(`^require\s+(\S+)\s+(\S+)(.*)`)
	requireBlockStartRe := regexp.MustCompile(`^require\s*\(`)
	depLineRe := regexp.MustCompile(`^\s*(\S+)\s+(\S+)(.*)`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		// Handle require block end
		if inRequireBlock && trimmedLine == ")" {
			inRequireBlock = false
			continue
		}

		// Parse module declaration
		if matches := moduleRe.FindStringSubmatch(trimmedLine); matches != nil {
			result.Module = matches[1]
			continue
		}

		// Parse go version
		if matches := goVersionRe.FindStringSubmatch(trimmedLine); matches != nil {
			result.GoVersion = matches[1]
			continue
		}

		// Parse require block start
		if requireBlockStartRe.MatchString(trimmedLine) {
			inRequireBlock = true
			continue
		}

		// Parse single-line require
		if matches := requireSingleRe.FindStringSubmatch(trimmedLine); matches != nil {
			dep := Dependency{
				Path:     matches[1],
				Version:  matches[2],
				Indirect: strings.Contains(matches[3], "indirect"),
			}
			result.Dependencies = append(result.Dependencies, dep)
			continue
		}

		// Parse dependency line in block
		if inRequireBlock {
			if matches := depLineRe.FindStringSubmatch(trimmedLine); matches != nil {
				dep := Dependency{
					Path:     matches[1],
					Version:  matches[2],
					Indirect: strings.Contains(matches[3], "indirect"),
				}
				result.Dependencies = append(result.Dependencies, dep)
			}
		}
	}

	return result, scanner.Err()
}

// DirectDependencies returns only the non-indirect dependencies.
func (p *ParseResult) DirectDependencies() []Dependency {
	var direct []Dependency
	for _, dep := range p.Dependencies {
		if !dep.Indirect {
			direct = append(direct, dep)
		}
	}
	return direct
}

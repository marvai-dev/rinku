// Package gomod parses go.mod files.
package gomod

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

const MaxDependencies = 10000

var (
	ErrTooManyDependencies = errors.New("too many dependencies (limit: 10000)")
	ErrUnclosedRequireBlock = errors.New("unclosed require block")

	moduleRe          = regexp.MustCompile(`^module\s+(\S+)`)
	goVersionRe       = regexp.MustCompile(`^go\s+(\S+)`)
	requireSingleRe   = regexp.MustCompile(`^require\s+(\S+)\s+(\S+)(.*)`)
	requireBlockStart = regexp.MustCompile(`^require\s*\(`)
	depLineRe         = regexp.MustCompile(`^\s*(\S+)\s+(\S+)(.*)`)
)

type Dependency struct {
	Path     string
	Version  string
	Indirect bool
}

type ParseResult struct {
	Module       string
	GoVersion    string
	Dependencies []Dependency
}

func Parse(path string) (*ParseResult, error) {
	return ParseFS(afero.NewOsFs(), path)
}

// ParseFS parses a go.mod from a filesystem (useful for testing).
func ParseFS(fs afero.Fs, path string) (*ParseResult, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ParseReader(file)
}

func ParseReader(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}
	scanner := bufio.NewScanner(r)
	inBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if inBlock && line == ")" {
			inBlock = false
			continue
		}

		if m := moduleRe.FindStringSubmatch(line); m != nil {
			result.Module = m[1]
			continue
		}

		if m := goVersionRe.FindStringSubmatch(line); m != nil {
			result.GoVersion = m[1]
			continue
		}

		if requireBlockStart.MatchString(line) {
			inBlock = true
			continue
		}

		if m := requireSingleRe.FindStringSubmatch(line); m != nil {
			result.Dependencies = append(result.Dependencies, Dependency{
				Path:     m[1],
				Version:  m[2],
				Indirect: strings.Contains(m[3], "// indirect"),
			})
			if len(result.Dependencies) >= MaxDependencies {
				return nil, ErrTooManyDependencies
			}
			continue
		}

		if inBlock {
			if m := depLineRe.FindStringSubmatch(line); m != nil {
				result.Dependencies = append(result.Dependencies, Dependency{
					Path:     m[1],
					Version:  m[2],
					Indirect: strings.Contains(m[3], "// indirect"),
				})
				if len(result.Dependencies) >= MaxDependencies {
					return nil, ErrTooManyDependencies
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inBlock {
		return nil, ErrUnclosedRequireBlock
	}
	return result, nil
}

func (p *ParseResult) DirectDependencies() []Dependency {
	var direct []Dependency
	for _, dep := range p.Dependencies {
		if !dep.Indirect {
			direct = append(direct, dep)
		}
	}
	return direct
}

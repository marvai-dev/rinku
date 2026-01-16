// Package verify checks requirement coverage against expected project tags.
package verify

import (
	"path/filepath"
	"strings"

	"github.com/stephan/rinku/internal/requirements"
)

// TagToCategoryMap maps tags to requirement path prefixes.
// A wildcard * matches any binary name.
var TagToCategoryMap = map[string][]string{
	"cli":              {"*/cli"},
	"web":              {"*/api"},
	"templating":       {"*/templates"},
	"sql":              {"db"},
	"orm":              {"db"},
	"codegen:protobuf": {"codegen/protobuf"},
	"codegen:ent":      {"codegen/ent"},
	"codegen:templ":    {"codegen/templ"},
	"codegen:wire":     {"codegen/wire"},
	"codegen:sqlc":     {"codegen/sqlc"},
	"codegen:gqlgen":   {"codegen/gqlgen"},
}

// CategoryStatus represents coverage for a requirement category.
type CategoryStatus struct {
	Category        string
	Pattern         string
	Expected        bool
	HasRequirements bool
	Count           int
	DoneCount       int
	Paths           []string
}

// CheckCoverage compares expected tags against captured requirements.
func CheckCoverage(projectDir string, tags []string) ([]CategoryStatus, error) {
	// Get all requirements
	allReqs, err := requirements.List(projectDir, "")
	if err != nil {
		return nil, err
	}

	// Build set of expected categories from tags
	expectedPatterns := make(map[string]string) // pattern -> tag
	for _, tag := range tags {
		if patterns, ok := TagToCategoryMap[tag]; ok {
			for _, pattern := range patterns {
				expectedPatterns[pattern] = tag
			}
		}
	}

	// Check each expected pattern
	var results []CategoryStatus
	for pattern, tag := range expectedPatterns {
		matching := filterByPattern(allReqs, pattern)

		// Count done requirements
		doneCount := 0
		for _, path := range matching {
			req, _ := requirements.Get(projectDir, path)
			if req != nil && req.Done {
				doneCount++
			}
		}

		results = append(results, CategoryStatus{
			Category:        tag,
			Pattern:         pattern,
			Expected:        true,
			HasRequirements: len(matching) > 0,
			Count:           len(matching),
			DoneCount:       doneCount,
			Paths:           matching,
		})
	}

	return results, nil
}

// CheckImplementation returns done and pending requirement paths.
func CheckImplementation(projectDir string) (done, pending []string, err error) {
	paths, err := requirements.List(projectDir, "")
	if err != nil {
		return nil, nil, err
	}

	for _, path := range paths {
		req, err := requirements.Get(projectDir, path)
		if err != nil {
			return nil, nil, err
		}
		if req == nil {
			continue
		}
		if req.Done {
			done = append(done, path)
		} else {
			pending = append(pending, path)
		}
	}

	return done, pending, nil
}

// filterByPattern returns requirement paths matching the pattern.
// Pattern supports * as a wildcard for a single path segment.
func filterByPattern(reqs []string, pattern string) []string {
	var matches []string
	for _, req := range reqs {
		if matchPattern(pattern, req) {
			matches = append(matches, req)
		}
	}
	return matches
}

// matchPattern checks if a requirement path matches the pattern.
// * matches any single path segment (not including /).
func matchPattern(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	// Pattern must be a prefix match
	if len(patternParts) > len(pathParts) {
		return false
	}

	for i, pp := range patternParts {
		if pp == "*" {
			continue // Wildcard matches anything
		}
		if pp != pathParts[i] {
			return false
		}
	}

	return true
}

// FilterByPattern is exported for use in gating.
func FilterByPattern(reqs []string, pattern string) []string {
	return filterByPattern(reqs, pattern)
}

// MatchPattern is exported for use in gating.
func MatchPattern(pattern, path string) bool {
	return matchPattern(pattern, path)
}

// GetRequirementStatus returns whether all requirements matching a pattern are done.
func GetRequirementStatus(projectDir, pattern string) (allDone bool, pending []string, err error) {
	paths, err := requirements.List(projectDir, "")
	if err != nil {
		return false, nil, err
	}

	matching := filterByPattern(paths, pattern)
	if len(matching) == 0 {
		// No requirements for this pattern - consider it satisfied
		return true, nil, nil
	}

	for _, path := range matching {
		req, err := requirements.Get(projectDir, path)
		if err != nil {
			return false, nil, err
		}
		if req != nil && !req.Done {
			pending = append(pending, path)
		}
	}

	return len(pending) == 0, pending, nil
}

// GetRequirementsByPattern returns all requirement paths matching the pattern.
func GetRequirementsByPattern(projectDir, pattern string) ([]string, error) {
	paths, err := requirements.List(projectDir, "")
	if err != nil {
		return nil, err
	}
	return filterByPattern(paths, pattern), nil
}

// ExpandWildcardPattern expands a pattern with * to match actual requirement paths.
// Returns the pattern as-is if it contains no wildcard.
func ExpandWildcardPattern(projectDir string, pattern string) ([]string, error) {
	if !strings.Contains(pattern, "*") {
		return []string{pattern}, nil
	}

	paths, err := requirements.List(projectDir, "")
	if err != nil {
		return nil, err
	}

	// Find unique prefixes that match the pattern
	prefixes := make(map[string]struct{})
	patternParts := strings.Split(pattern, "/")

	for _, path := range paths {
		pathParts := strings.Split(path, "/")
		if len(pathParts) < len(patternParts) {
			continue
		}

		// Check if path matches pattern and extract the expanded prefix
		match := true
		expanded := make([]string, len(patternParts))
		for i, pp := range patternParts {
			if pp == "*" {
				expanded[i] = pathParts[i]
			} else if pp == pathParts[i] {
				expanded[i] = pp
			} else {
				match = false
				break
			}
		}

		if match {
			prefixes[filepath.Join(expanded...)] = struct{}{}
		}
	}

	result := make([]string, 0, len(prefixes))
	for p := range prefixes {
		result = append(result, p)
	}
	return result, nil
}

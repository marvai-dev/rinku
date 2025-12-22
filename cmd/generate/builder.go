package main

import (
	"strings"

	"github.com/stephan/rinku/internal/url"
)

// LibsFile represents the libs.json structure
type LibsFile struct {
	Libs map[string]Library `json:"libs"`
}

// Library represents a library entry
type Library struct {
	URL    string `json:"url"`
	Lang   string `json:"lang"`
	Unsafe string `json:"unsafe,omitempty"`
}

// MappingsFile represents the mappings.json structure
type MappingsFile struct {
	Mappings []Mapping `json:"mappings"`
}

// Mapping represents a directional library mapping entry
type Mapping struct {
	Source   string   `json:"source"`
	Targets  []string `json:"targets"`
	Category string   `json:"category"`
}

// IndexResult contains all generated indexes
type IndexResult struct {
	Forward        map[string][]string // target_lang:source_url -> target_urls (safe only)
	ForwardAll     map[string][]string // target_lang:source_url -> target_urls (including unsafe)
	Reverse        map[string][]string // source_lang:target_url -> source_urls (safe only)
	ReverseAll     map[string][]string // source_lang:target_url -> source_urls (including unsafe)
	UnsafeCount    int
	MappingsCount  int
	LibrariesCount int
}

// BuildIndexes creates lookup indexes from libs and mappings.
// Returns forward and reverse indexes for both safe and all entries.
func BuildIndexes(libs map[string]Library, mappings []Mapping) IndexResult {
	result := IndexResult{
		Forward:        make(map[string][]string),
		ForwardAll:     make(map[string][]string),
		Reverse:        make(map[string][]string),
		ReverseAll:     make(map[string][]string),
		LibrariesCount: len(libs),
		MappingsCount:  len(mappings),
	}

	// Count unsafe libraries
	for _, lib := range libs {
		if lib.Unsafe != "" {
			result.UnsafeCount++
		}
	}

	for _, mapping := range mappings {
		sourceLib, sourceExists := libs[mapping.Source]
		if !sourceExists {
			continue // Skip if source lib not found
		}

		sourceURL := sourceLib.URL
		sourceLang := sourceLib.Lang
		sourceUnsafe := sourceLib.Unsafe != ""

		for _, targetID := range mapping.Targets {
			if targetID == "<None>" {
				continue // Skip placeholder targets
			}

			targetLib, targetExists := libs[targetID]
			if !targetExists {
				continue // Skip if target lib not found
			}

			targetURL := targetLib.URL
			targetLang := targetLib.Lang
			targetUnsafe := targetLib.Unsafe != ""

			// Forward index: given source URL, find targets in target language
			// Key: target_lang:normalized_source_url
			forwardKey := strings.ToLower(targetLang) + ":" + url.Normalize(sourceURL)
			result.ForwardAll[forwardKey] = append(result.ForwardAll[forwardKey], targetURL)
			if !sourceUnsafe && !targetUnsafe {
				result.Forward[forwardKey] = append(result.Forward[forwardKey], targetURL)
			}

			// Reverse index: given target URL, find sources in source language
			// Key: source_lang:normalized_target_url
			reverseKey := strings.ToLower(sourceLang) + ":" + url.Normalize(targetURL)
			result.ReverseAll[reverseKey] = append(result.ReverseAll[reverseKey], sourceURL)
			if !sourceUnsafe && !targetUnsafe {
				result.Reverse[reverseKey] = append(result.Reverse[reverseKey], sourceURL)
			}
		}
	}

	return result
}

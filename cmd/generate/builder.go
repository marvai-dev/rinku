package main

import (
	"strings"

	"github.com/stephan/rinku/internal/types"
	"github.com/stephan/rinku/internal/url"
)

// IndexResult contains all generated indexes
type IndexResult struct {
	Forward         map[string][]string            // target_lang:source_url -> target_urls (safe only)
	ForwardAll      map[string][]string            // target_lang:source_url -> target_urls (including unsafe)
	Reverse         map[string][]string            // source_lang:target_url -> source_urls (safe only)
	ReverseAll      map[string][]string            // source_lang:target_url -> source_urls (including unsafe)
	KnownCrateNames map[string]string              // normalized_url -> crate_name (for Rust libraries)
	Tags            map[string][]string            // normalized_url -> tags (for all libraries)
	RequiredDeps    map[string][]types.RequiredDep // target_lang:source_url -> required deps
	UnsafeCount     int
	MappingsCount   int
	LibrariesCount  int
}

func BuildIndexes(libs map[string]types.Library, mappings []types.Mapping) IndexResult {
	result := IndexResult{
		Forward:         make(map[string][]string),
		ForwardAll:      make(map[string][]string),
		Reverse:         make(map[string][]string),
		ReverseAll:      make(map[string][]string),
		KnownCrateNames: make(map[string]string),
		Tags:            make(map[string][]string),
		RequiredDeps:    make(map[string][]types.RequiredDep),
		LibrariesCount:  len(libs),
		MappingsCount:   len(mappings),
	}

	// Count unsafe libraries and build crate names and tags maps
	for _, lib := range libs {
		if lib.Unsafe != "" {
			result.UnsafeCount++
		}
		normalizedURL := url.Normalize(lib.URL)
		// Build crate names map for Rust libraries with explicit crate names
		if lib.Lang == "rust" && lib.CrateName != "" {
			result.KnownCrateNames[normalizedURL] = lib.CrateName
		}
		// Build tags map for all libraries with tags
		if len(lib.Tags) > 0 {
			result.Tags[normalizedURL] = lib.Tags
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

		// Required dependencies: key by forward key (target_lang:source_url)
		if len(mapping.Requires) > 0 {
			for _, targetID := range mapping.Targets {
				if targetID == "<None>" {
					continue
				}
				targetLib, targetExists := libs[targetID]
				if !targetExists {
					continue
				}
				reqKey := strings.ToLower(targetLib.Lang) + ":" + url.Normalize(sourceURL)
				result.RequiredDeps[reqKey] = append(result.RequiredDeps[reqKey], mapping.Requires...)
			}
		}
	}

	return result
}

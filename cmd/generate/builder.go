package main

import (
	"strings"

	"github.com/stephan/rinku/internal/url"
)

// BuildIndexes creates lookup indexes from mappings.
// Returns the enabled-only index, all-entries index, and count of disabled entries.
func BuildIndexes(mappings []Mapping) (index, indexAll map[string][]string, disabledCount int) {
	index = make(map[string][]string)
	indexAll = make(map[string][]string)

	for _, mapping := range mappings {
		key := strings.ToLower(mapping.TargetLang) + ":" + url.Normalize(mapping.Source)
		indexAll[key] = append(indexAll[key], mapping.Target...)
		if mapping.Disabled == "" {
			index[key] = append(index[key], mapping.Target...)
		} else {
			disabledCount++
		}
	}

	return index, indexAll, disabledCount
}

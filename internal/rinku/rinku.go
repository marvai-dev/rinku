package rinku

import (
	"strings"

	"github.com/stephan/rinku/internal/url"
)

// Rinku provides library lookup functionality.
type Rinku struct {
	index           map[string][]string
	indexAll        map[string][]string
	reverseIndex    map[string][]string
	reverseIndexAll map[string][]string
}

// New creates a new Rinku instance with the given indexes.
func New(index, indexAll, reverseIndex, reverseIndexAll map[string][]string) *Rinku {
	return &Rinku{
		index:           index,
		indexAll:        indexAll,
		reverseIndex:    reverseIndex,
		reverseIndexAll: reverseIndexAll,
	}
}

// Lookup finds equivalent libraries for the given source URL and target language.
// If unsafe is true, includes libraries with known vulnerabilities.
func (r *Rinku) Lookup(sourceURL, targetLang string, unsafe bool) []string {
	key := strings.ToLower(targetLang) + ":" + url.Normalize(sourceURL)
	if unsafe {
		return r.indexAll[key]
	}
	return r.index[key]
}

// ReverseLookup finds source libraries that map to the given target URL in the specified source language.
// If unsafe is true, includes libraries with known vulnerabilities.
func (r *Rinku) ReverseLookup(targetURL, sourceLang string, unsafe bool) []string {
	key := strings.ToLower(sourceLang) + ":" + url.Normalize(targetURL)
	if unsafe {
		return r.reverseIndexAll[key]
	}
	return r.reverseIndex[key]
}

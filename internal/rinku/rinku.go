package rinku

import (
	"strings"

	"github.com/stephan/rinku/internal/url"
)

// Rinku provides cross-language library lookup.
type Rinku struct {
	safe        map[string][]string // excludes vulnerable libraries
	all         map[string][]string // includes vulnerable libraries
	reverseSafe map[string][]string
	reverseAll  map[string][]string
	crateNames  map[string]string // normalized_url -> crate_name
}

func New(safe, all, reverseSafe, reverseAll map[string][]string, crateNames map[string]string) *Rinku {
	return &Rinku{safe: safe, all: all, reverseSafe: reverseSafe, reverseAll: reverseAll, crateNames: crateNames}
}

// CrateName returns the known crate name for a Rust library URL.
// Returns empty string if no explicit crate name is configured.
func (r *Rinku) CrateName(rustURL string) string {
	return r.crateNames[url.Normalize(rustURL)]
}

func (r *Rinku) Lookup(sourceURL, targetLang string, includeUnsafe bool) []string {
	key := strings.ToLower(targetLang) + ":" + url.Normalize(sourceURL)
	if includeUnsafe {
		return r.all[key]
	}
	return r.safe[key]
}

func (r *Rinku) ReverseLookup(targetURL, sourceLang string, includeUnsafe bool) []string {
	key := strings.ToLower(sourceLang) + ":" + url.Normalize(targetURL)
	if includeUnsafe {
		return r.reverseAll[key]
	}
	return r.reverseSafe[key]
}

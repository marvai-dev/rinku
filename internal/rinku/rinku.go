package rinku

import (
	"strings"

	"github.com/stephan/rinku/internal/types"
	"github.com/stephan/rinku/internal/url"
)

// Rinku provides cross-language library lookup.
type Rinku struct {
	safe         map[string][]string             // excludes vulnerable libraries
	all          map[string][]string             // includes vulnerable libraries
	reverseSafe  map[string][]string
	reverseAll   map[string][]string
	crateNames   map[string]string               // normalized_url -> crate_name
	tags         map[string][]string             // normalized_url -> tags
	requiredDeps map[string][]types.RequiredDep  // target_lang:source_url -> required deps
}

func New(safe, all, reverseSafe, reverseAll map[string][]string, crateNames map[string]string, tags map[string][]string, requiredDeps map[string][]types.RequiredDep) *Rinku {
	return &Rinku{
		safe:         safe,
		all:          all,
		reverseSafe:  reverseSafe,
		reverseAll:   reverseAll,
		crateNames:   crateNames,
		tags:         tags,
		requiredDeps: requiredDeps,
	}
}

// CrateName returns the known crate name for a Rust library URL.
// Returns empty string if no explicit crate name is configured.
func (r *Rinku) CrateName(rustURL string) string {
	return r.crateNames[url.Normalize(rustURL)]
}

// Tags returns the tags for a library URL.
// Returns nil if no tags are configured for this library.
func (r *Rinku) Tags(libURL string) []string {
	return r.tags[url.Normalize(libURL)]
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

// RequiredDeps returns the required dependencies for a lookup.
// Uses the same key format as Lookup: targetLang:sourceURL
func (r *Rinku) RequiredDeps(sourceURL, targetLang string) []types.RequiredDep {
	key := strings.ToLower(targetLang) + ":" + url.Normalize(sourceURL)
	return r.requiredDeps[key]
}

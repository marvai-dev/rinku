package rinku

import (
	"strings"

	"github.com/stephan/rinku/internal/url"
)

// Rinku provides library lookup functionality.
type Rinku struct {
	index    map[string][]string
	indexAll map[string][]string
}

// New creates a new Rinku instance with the given indexes.
func New(index, indexAll map[string][]string) *Rinku {
	return &Rinku{
		index:    index,
		indexAll: indexAll,
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

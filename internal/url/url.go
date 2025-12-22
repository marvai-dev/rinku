package url

import "strings"

// Normalize converts a URL to a canonical form for lookup.
// It lowercases, strips http(s):// prefix, and removes trailing slash.
func Normalize(url string) string {
	url = strings.ToLower(url)
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	return url
}

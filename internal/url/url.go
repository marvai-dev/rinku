package url

import "strings"

// Normalize converts a URL to a canonical form for lookup.
// It lowercases, strips http(s):// prefix, removes www. from host only,
// and removes trailing slash. Returns empty string for invalid input.
func Normalize(inputURL string) string {
	if inputURL == "" {
		return ""
	}

	url := strings.ToLower(inputURL)
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Handle empty result after stripping protocol (e.g., "https://")
	if url == "" {
		return ""
	}

	// Only strip www. from the host portion, not from paths
	// Find the first slash to separate host from path
	slashIdx := strings.Index(url, "/")
	if slashIdx == -1 {
		// No path, entire string is host
		url = strings.TrimPrefix(url, "www.")
	} else {
		// Split into host and path, only strip www. from host
		host := url[:slashIdx]
		path := url[slashIdx:]
		host = strings.TrimPrefix(host, "www.")
		url = host + path
	}

	url = strings.TrimSuffix(url, "/")
	return url
}

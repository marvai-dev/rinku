package url

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"strips https prefix", "https://github.com/foo/bar", "github.com/foo/bar"},
		{"strips http prefix", "http://github.com/foo/bar", "github.com/foo/bar"},
		{"strips trailing slash", "github.com/foo/bar/", "github.com/foo/bar"},
		{"lowercases URL", "GitHub.com/Foo/Bar", "github.com/foo/bar"},
		{"strips www prefix", "www.github.com/foo/bar", "github.com/foo/bar"},
		{"handles all transformations", "HTTPS://www.GitHub.com/Foo/Bar/", "github.com/foo/bar"},
		{"no changes needed", "github.com/foo/bar", "github.com/foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

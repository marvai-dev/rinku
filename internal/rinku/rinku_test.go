package rinku

import (
	"reflect"
	"testing"
)

func TestLookup(t *testing.T) {
	// Create test indexes
	index := map[string][]string{
		"rust:github.com/spf13/cobra":   {"https://github.com/clap-rs/clap"},
		"rust:github.com/gin-gonic/gin": {"https://github.com/tokio-rs/axum"},
	}
	indexAll := map[string][]string{
		"rust:github.com/spf13/cobra":   {"https://github.com/clap-rs/clap"},
		"rust:github.com/gin-gonic/gin": {"https://github.com/tokio-rs/axum"},
		"rust:github.com/golang/net":    {"https://github.com/hyperium/hyper"}, // disabled in index
	}
	reverseIndex := map[string][]string{
		"go:github.com/clap-rs/clap":   {"https://github.com/spf13/cobra"},
		"go:github.com/tokio-rs/axum":  {"https://github.com/gin-gonic/gin"},
	}
	reverseIndexAll := map[string][]string{
		"go:github.com/clap-rs/clap":    {"https://github.com/spf13/cobra"},
		"go:github.com/tokio-rs/axum":   {"https://github.com/gin-gonic/gin"},
		"go:github.com/hyperium/hyper":  {"https://github.com/golang/net"}, // disabled in reverseIndex
	}

	r := New(index, indexAll, reverseIndex, reverseIndexAll, nil)

	tests := []struct {
		name       string
		sourceURL  string
		targetLang string
		unsafe     bool
		want       []string
	}{
		{
			name:       "finds mapping",
			sourceURL:  "https://github.com/spf13/cobra",
			targetLang: "rust",
			unsafe:     false,
			want:       []string{"https://github.com/clap-rs/clap"},
		},
		{
			name:       "normalizes URL with http prefix",
			sourceURL:  "http://github.com/spf13/cobra",
			targetLang: "rust",
			unsafe:     false,
			want:       []string{"https://github.com/clap-rs/clap"},
		},
		{
			name:       "normalizes URL with trailing slash",
			sourceURL:  "https://github.com/spf13/cobra/",
			targetLang: "rust",
			unsafe:     false,
			want:       []string{"https://github.com/clap-rs/clap"},
		},
		{
			name:       "case insensitive",
			sourceURL:  "https://GitHub.com/SPF13/Cobra",
			targetLang: "rust",
			unsafe:     false,
			want:       []string{"https://github.com/clap-rs/clap"},
		},
		{
			name:       "not found returns nil",
			sourceURL:  "https://github.com/nonexistent/lib",
			targetLang: "rust",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "wrong language returns nil",
			sourceURL:  "https://github.com/spf13/cobra",
			targetLang: "python",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "disabled entry not in safe mode",
			sourceURL:  "https://github.com/golang/net",
			targetLang: "rust",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "disabled entry available in unsafe mode",
			sourceURL:  "https://github.com/golang/net",
			targetLang: "rust",
			unsafe:     true,
			want:       []string{"https://github.com/hyperium/hyper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Lookup(tt.sourceURL, tt.targetLang, tt.unsafe)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lookup(%q, %q, %v) = %v, want %v",
					tt.sourceURL, tt.targetLang, tt.unsafe, got, tt.want)
			}
		})
	}
}

func TestReverseLookup(t *testing.T) {
	// Create test indexes
	index := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
	}
	indexAll := map[string][]string{
		"rust:github.com/spf13/cobra": {"https://github.com/clap-rs/clap"},
		"rust:github.com/golang/net":  {"https://github.com/hyperium/hyper"},
	}
	reverseIndex := map[string][]string{
		"go:github.com/clap-rs/clap": {"https://github.com/spf13/cobra"},
	}
	reverseIndexAll := map[string][]string{
		"go:github.com/clap-rs/clap":   {"https://github.com/spf13/cobra"},
		"go:github.com/hyperium/hyper": {"https://github.com/golang/net"},
	}

	r := New(index, indexAll, reverseIndex, reverseIndexAll, nil)

	tests := []struct {
		name       string
		targetURL  string
		sourceLang string
		unsafe     bool
		want       []string
	}{
		{
			name:       "finds reverse mapping",
			targetURL:  "https://github.com/clap-rs/clap",
			sourceLang: "go",
			unsafe:     false,
			want:       []string{"https://github.com/spf13/cobra"},
		},
		{
			name:       "normalizes URL",
			targetURL:  "HTTPS://GitHub.com/clap-rs/clap/",
			sourceLang: "go",
			unsafe:     false,
			want:       []string{"https://github.com/spf13/cobra"},
		},
		{
			name:       "not found returns nil",
			targetURL:  "https://github.com/nonexistent/lib",
			sourceLang: "go",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "wrong language returns nil",
			targetURL:  "https://github.com/clap-rs/clap",
			sourceLang: "python",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "disabled entry not in safe mode",
			targetURL:  "https://github.com/hyperium/hyper",
			sourceLang: "go",
			unsafe:     false,
			want:       nil,
		},
		{
			name:       "disabled entry available in unsafe mode",
			targetURL:  "https://github.com/hyperium/hyper",
			sourceLang: "go",
			unsafe:     true,
			want:       []string{"https://github.com/golang/net"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.ReverseLookup(tt.targetURL, tt.sourceLang, tt.unsafe)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReverseLookup(%q, %q, %v) = %v, want %v",
					tt.targetURL, tt.sourceLang, tt.unsafe, got, tt.want)
			}
		})
	}
}

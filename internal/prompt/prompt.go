package prompt

import (
	_ "embed"

	"github.com/stephan/rinku/internal/multistep"
)

//go:embed migration-prompt.md
var migrationPrompt string

// Migration returns the parsed migration workflow prompt.
func Migration() (*multistep.Prompt, error) {
	return multistep.Parse(migrationPrompt)
}

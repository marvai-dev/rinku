package multistep

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Prompt holds parsed steps from a markdown prompt file.
type Prompt struct {
	steps        map[string]string
	order        []string
	introduction string // Content from "# Introduction" section, entry point
	before       string // Content from "# Before" section, shown before each step
	after        string // Content from "# After" section, shown after each step
}

// Parse parses steps from markdown content.
// Steps are identified by headers like "# Step 1" or "# Step Find Tests".
// Special "# Before" and "# After" sections are shown before/after each step when using --start.
func Parse(content string) (*Prompt, error) {
	p := &Prompt{
		steps: make(map[string]string),
		order: []string{},
	}

	var currentSection string // "before" or step ID
	var currentContent strings.Builder

	for _, line := range strings.Split(content, "\n") {
		if isIntroductionHeader(line) {
			// Save previous section if any
			if currentSection != "" {
				saveSection(p, currentSection, currentContent.String())
			}
			currentSection = "introduction"
			currentContent.Reset()
		} else if isBeforeHeader(line) {
			// Save previous section if any
			if currentSection != "" {
				saveSection(p, currentSection, currentContent.String())
			}
			currentSection = "before"
			currentContent.Reset()
		} else if isAfterHeader(line) {
			// Save previous section if any
			if currentSection != "" {
				saveSection(p, currentSection, currentContent.String())
			}
			currentSection = "after"
			currentContent.Reset()
		} else if id, ok := parseStepHeader(line); ok {
			// Save previous section if any
			if currentSection != "" {
				saveSection(p, currentSection, currentContent.String())
			}
			currentSection = id
			p.order = append(p.order, id)
			currentContent.Reset()
		} else if currentSection != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save last section
	if currentSection != "" {
		saveSection(p, currentSection, currentContent.String())
	}

	if len(p.steps) == 0 {
		return nil, errors.New("no steps found")
	}

	return p, nil
}

func saveSection(p *Prompt, section, content string) {
	content = strings.TrimSpace(content)
	switch section {
	case "introduction":
		p.introduction = content
	case "before":
		p.before = content
	case "after":
		p.after = content
	default:
		p.steps[section] = content
	}
}

func isIntroductionHeader(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	return strings.EqualFold(rest, "introduction")
}

func isBeforeHeader(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	return strings.EqualFold(rest, "before")
}

func isAfterHeader(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	return strings.EqualFold(rest, "after")
}

// parseStepHeader checks if a line is a step header and returns the step ID.
// Matches "# Step <id>" or "# step <id>" (case insensitive for "step").
func parseStepHeader(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return "", false
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	lower := strings.ToLower(rest)

	// Must match "step " (with space) to avoid matching "steps", "stepping", etc.
	if !strings.HasPrefix(lower, "step ") {
		return "", false
	}

	// Extract everything after "step " as the ID
	afterStep := strings.TrimSpace(rest[5:])
	if afterStep == "" {
		return "", false
	}

	return afterStep, true
}

// ParseFile parses steps from a file path.
func ParseFile(path string) (*Prompt, error) {
	content, err := os.ReadFile(path) //#nosec G304 -- caller provides trusted path
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return Parse(string(content))
}

// GetStep returns the content for a step ID.
func (p *Prompt) GetStep(id string) (string, bool) {
	content, ok := p.steps[id]
	return content, ok
}

// Introduction returns the content of the "# Introduction" section.
func (p *Prompt) Introduction() string {
	return p.introduction
}

// Before returns the content of the "# Before" section.
func (p *Prompt) Before() string {
	return p.before
}

// After returns the content of the "# After" section.
func (p *Prompt) After() string {
	return p.after
}

// FirstStep returns the ID of the first step.
func (p *Prompt) FirstStep() string {
	if len(p.order) == 0 {
		return ""
	}
	return p.order[0]
}

// Steps returns all step IDs in order.
func (p *Prompt) Steps() []string {
	result := make([]string, len(p.order))
	copy(result, p.order)
	return result
}

// Bootstrap returns the initial instruction for an LLM.
func (p *Prompt) Bootstrap(command string) string {
	first := p.FirstStep()
	if first == "" {
		return ""
	}
	return fmt.Sprintf("Execute '%s %s'. This will return instructions. Execute those instructions.", command, first)
}

package requirements

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/natefinch/atomic"
	"github.com/stephan/rinku/internal/progress"
)

const (
	RequirementsDir = "requirements"
)

// requirementsPath returns the full path for a requirement file.
func requirementsPath(projectDir, reqPath string) string {
	return filepath.Join(projectDir, progress.ProgressDir, RequirementsDir, reqPath+".json")
}

// Set creates or updates a requirement.
func Set(projectDir, reqPath, content string) error {
	now := time.Now()

	// Try to load existing requirement to preserve created_at
	existing, _ := Get(projectDir, reqPath)

	req := &Requirement{
		Path:      reqPath,
		Content:   content,
		Step:      getCurrentStep(projectDir),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if existing != nil {
		req.CreatedAt = existing.CreatedAt
	}

	return save(projectDir, req)
}

// Get retrieves a requirement by path.
func Get(projectDir, reqPath string) (*Requirement, error) {
	path := requirementsPath(projectDir, reqPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading requirement: %w", err)
	}

	var req Requirement
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("parsing requirement: %w", err)
	}
	return &req, nil
}

// List returns all requirement paths, optionally filtered by prefix.
func List(projectDir, prefix string) ([]string, error) {
	baseDir := filepath.Join(projectDir, progress.ProgressDir, RequirementsDir)

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, nil
	}

	var paths []string
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Get relative path and remove .json extension
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		reqPath := strings.TrimSuffix(relPath, ".json")

		// Filter by prefix if provided
		if prefix != "" && !strings.HasPrefix(reqPath, prefix) {
			return nil
		}

		paths = append(paths, reqPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("listing requirements: %w", err)
	}

	sort.Strings(paths)
	return paths, nil
}

// Delete removes a requirement.
func Delete(projectDir, reqPath string) error {
	path := requirementsPath(projectDir, reqPath)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Done marks a requirement as done.
func Done(projectDir, reqPath string) error {
	req, err := Get(projectDir, reqPath)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("requirement '%s' not found\nHint: Requirements are stored in .rinku/ - are you in the correct project directory?", reqPath)
	}

	now := time.Now()
	req.Done = true
	req.DoneAt = &now
	req.UpdatedAt = now

	return save(projectDir, req)
}

// save writes a requirement to disk atomically.
func save(projectDir string, req *Requirement) error {
	path := requirementsPath(projectDir, req.Path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling requirement: %w", err)
	}

	return atomic.WriteFile(path, bytes.NewReader(append(data, '\n')))
}

// getCurrentStep reads the current step from progress.json.
func getCurrentStep(projectDir string) string {
	m, err := progress.Load(projectDir)
	if err != nil || m == nil {
		return ""
	}
	return m.GetCurrentStep()
}

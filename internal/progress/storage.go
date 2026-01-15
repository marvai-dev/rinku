package progress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natefinch/atomic"
)

const (
	ProgressDir  = ".rinku"
	ProgressFile = "progress.json"
)

// ProgressPath returns the path to progress.json for a project directory.
func ProgressPath(projectDir string) string {
	return filepath.Join(projectDir, ProgressDir, ProgressFile)
}

// Load reads progress from disk. Returns nil, nil if no progress file exists.
func Load(projectDir string) (*Migration, error) {
	path := ProgressPath(projectDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading progress: %w", err)
	}

	var m Migration
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing progress: %w", err)
	}
	return &m, nil
}

// Save atomically writes progress to disk.
func (m *Migration) Save(projectDir string) error {
	dir := filepath.Join(projectDir, ProgressDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s directory: %w", ProgressDir, err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling progress: %w", err)
	}

	path := ProgressPath(projectDir)
	return atomic.WriteFile(path, bytes.NewReader(append(data, '\n')))
}

// Delete removes the progress file.
func Delete(projectDir string) error {
	path := ProgressPath(projectDir)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists checks if a progress file exists.
func Exists(projectDir string) bool {
	_, err := os.Stat(ProgressPath(projectDir))
	return err == nil
}

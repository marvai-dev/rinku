package progress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_InitializesAllStepsPending(t *testing.T) {
	steps := []string{"1", "2", "3"}
	m := New("/test/project", steps)

	if m.Version != currentVersion {
		t.Errorf("Version = %d, want %d", m.Version, currentVersion)
	}
	if m.ProjectPath != "/test/project" {
		t.Errorf("ProjectPath = %q, want %q", m.ProjectPath, "/test/project")
	}
	if m.CurrentStep != "1" {
		t.Errorf("CurrentStep = %q, want %q", m.CurrentStep, "1")
	}
	if len(m.Steps) != 3 {
		t.Errorf("len(Steps) = %d, want 3", len(m.Steps))
	}
	for _, id := range steps {
		step := m.Steps[id]
		if step.Status != StepPending {
			t.Errorf("step %q status = %q, want %q", id, step.Status, StepPending)
		}
	}
}

func TestNew_EmptySteps(t *testing.T) {
	m := New("/test", []string{})

	if m.CurrentStep != "" {
		t.Errorf("CurrentStep = %q, want empty", m.CurrentStep)
	}
	if len(m.Steps) != 0 {
		t.Errorf("len(Steps) = %d, want 0", len(m.Steps))
	}
}

func TestStartStep(t *testing.T) {
	m := New("/test", []string{"1", "2", "3"})

	err := m.StartStep("2")
	if err != nil {
		t.Fatalf("StartStep failed: %v", err)
	}

	if m.CurrentStep != "2" {
		t.Errorf("CurrentStep = %q, want %q", m.CurrentStep, "2")
	}
	step := m.Steps["2"]
	if step.Status != StepInProgress {
		t.Errorf("status = %q, want %q", step.Status, StepInProgress)
	}
	if step.StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}

func TestStartStep_NotFound(t *testing.T) {
	m := New("/test", []string{"1", "2"})

	err := m.StartStep("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent step")
	}
}

func TestCompleteStep(t *testing.T) {
	m := New("/test", []string{"1", "2"})

	err := m.CompleteStep("1", "done")
	if err != nil {
		t.Fatalf("CompleteStep failed: %v", err)
	}

	step := m.Steps["1"]
	if step.Status != StepCompleted {
		t.Errorf("status = %q, want %q", step.Status, StepCompleted)
	}
	if step.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
	if step.Notes != "done" {
		t.Errorf("Notes = %q, want %q", step.Notes, "done")
	}
}

func TestCompleteStep_EmptyNotes(t *testing.T) {
	m := New("/test", []string{"1"})

	err := m.CompleteStep("1", "")
	if err != nil {
		t.Fatalf("CompleteStep failed: %v", err)
	}

	if m.Steps["1"].Notes != "" {
		t.Errorf("Notes should be empty, got %q", m.Steps["1"].Notes)
	}
}

func TestCompleteStep_NotFound(t *testing.T) {
	m := New("/test", []string{"1"})

	err := m.CompleteStep("nonexistent", "")
	if err == nil {
		t.Error("expected error for nonexistent step")
	}
}

func TestProgress(t *testing.T) {
	m := New("/test", []string{"1", "2", "3", "4"})

	completed, total := m.Progress()
	if completed != 0 || total != 4 {
		t.Errorf("Progress = (%d, %d), want (0, 4)", completed, total)
	}

	m.Steps["1"].Status = StepCompleted
	m.Steps["2"].Status = StepSkipped

	completed, total = m.Progress()
	if completed != 2 || total != 4 {
		t.Errorf("Progress = (%d, %d), want (2, 4)", completed, total)
	}
}

func TestIsComplete(t *testing.T) {
	m := New("/test", []string{"1", "2"})

	if m.IsComplete() {
		t.Error("IsComplete should be false initially")
	}

	m.Steps["1"].Status = StepCompleted
	m.Steps["2"].Status = StepCompleted

	if !m.IsComplete() {
		t.Error("IsComplete should be true when all steps completed")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	steps := []string{"1", "2", "3"}
	m := New(dir, steps)
	m.StartStep("1")
	m.CompleteStep("1", "first done")
	m.StartStep("2")

	err := m.Save(dir)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := ProgressPath(dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("progress file not created at %s", path)
	}

	// Load and verify
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != m.Version {
		t.Errorf("Version = %d, want %d", loaded.Version, m.Version)
	}
	if loaded.CurrentStep != "2" {
		t.Errorf("CurrentStep = %q, want %q", loaded.CurrentStep, "2")
	}
	if loaded.Steps["1"].Status != StepCompleted {
		t.Errorf("step 1 status = %q, want %q", loaded.Steps["1"].Status, StepCompleted)
	}
	if loaded.Steps["1"].Notes != "first done" {
		t.Errorf("step 1 notes = %q, want %q", loaded.Steps["1"].Notes, "first done")
	}
	if loaded.Steps["2"].Status != StepInProgress {
		t.Errorf("step 2 status = %q, want %q", loaded.Steps["2"].Status, StepInProgress)
	}
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()

	m, err := Load(dir)
	if err != nil {
		t.Fatalf("Load should not error for missing file: %v", err)
	}
	if m != nil {
		t.Error("Load should return nil for missing file")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, []string{"1"})
	m.Save(dir)

	if !Exists(dir) {
		t.Fatal("progress file should exist after save")
	}

	err := Delete(dir)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if Exists(dir) {
		t.Error("progress file should not exist after delete")
	}
}

func TestDelete_NoFile(t *testing.T) {
	dir := t.TempDir()

	err := Delete(dir)
	if err != nil {
		t.Errorf("Delete should not error for missing file: %v", err)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if Exists(dir) {
		t.Error("Exists should return false before save")
	}

	m := New(dir, []string{"1"})
	m.Save(dir)

	if !Exists(dir) {
		t.Error("Exists should return true after save")
	}
}

func TestProgressPath(t *testing.T) {
	got := ProgressPath("/my/project")
	want := filepath.Join("/my/project", ".rinku", "progress.json")
	if got != want {
		t.Errorf("ProgressPath = %q, want %q", got, want)
	}
}

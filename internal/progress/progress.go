package progress

import (
	"fmt"
	"time"
)

// StepStatus represents the state of a migration step.
type StepStatus string

const (
	StepPending    StepStatus = "pending"
	StepInProgress StepStatus = "in_progress"
	StepCompleted  StepStatus = "completed"
	StepSkipped    StepStatus = "skipped"
)

// StepRecord captures the state of a single step.
type StepRecord struct {
	ID          string     `json:"id"`
	Status      StepStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Notes       string     `json:"notes,omitempty"`
}

// Migration represents the full migration progress state.
type Migration struct {
	Version     int                    `json:"version"`
	StartedAt   time.Time              `json:"started_at"`
	ProjectPath string                 `json:"project_path"`
	CurrentStep string                 `json:"current_step"`
	Steps       map[string]*StepRecord `json:"steps"`
	StepOrder   []string               `json:"step_order"`
}

const currentVersion = 1

// New creates a new Migration with all steps initialized as pending.
func New(projectDir string, stepOrder []string) *Migration {
	now := time.Now()
	m := &Migration{
		Version:     currentVersion,
		StartedAt:   now,
		ProjectPath: projectDir,
		CurrentStep: "",
		Steps:       make(map[string]*StepRecord),
		StepOrder:   stepOrder,
	}

	for _, id := range stepOrder {
		m.Steps[id] = &StepRecord{
			ID:     id,
			Status: StepPending,
		}
	}

	if len(stepOrder) > 0 {
		m.CurrentStep = stepOrder[0]
	}

	return m
}

// StartStep marks a step as in_progress and sets it as the current step.
func (m *Migration) StartStep(id string) error {
	step, ok := m.Steps[id]
	if !ok {
		return fmt.Errorf("step '%s' not found", id)
	}

	now := time.Now()
	step.Status = StepInProgress
	step.StartedAt = &now
	m.CurrentStep = id
	return nil
}

// CompleteStep marks a step as completed with optional notes.
func (m *Migration) CompleteStep(id string, notes string) error {
	step, ok := m.Steps[id]
	if !ok {
		return fmt.Errorf("step '%s' not found", id)
	}

	now := time.Now()
	step.Status = StepCompleted
	step.CompletedAt = &now
	if notes != "" {
		step.Notes = notes
	}
	return nil
}

// GetCurrentStep returns the current step ID.
func (m *Migration) GetCurrentStep() string {
	return m.CurrentStep
}

// Progress returns the count of completed steps and total steps.
func (m *Migration) Progress() (completed int, total int) {
	for _, step := range m.Steps {
		if step.Status == StepCompleted || step.Status == StepSkipped {
			completed++
		}
	}
	return completed, len(m.Steps)
}

// IsComplete returns true if all steps are completed or skipped.
func (m *Migration) IsComplete() bool {
	completed, total := m.Progress()
	return completed == total
}

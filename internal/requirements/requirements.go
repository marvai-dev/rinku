package requirements

import "time"

// Requirement represents a captured requirement during migration.
type Requirement struct {
	Path      string     `json:"path"`
	Content   string     `json:"content"`
	Step      string     `json:"step"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Done      bool       `json:"done"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

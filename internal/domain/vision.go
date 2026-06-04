package domain

import "time"

// Vision captures the high-level strategic direction for a project:
// what problem is being solved, what success looks like, and whether
// the vision has been approved by stakeholders.
type Vision struct {
	ID              string
	ProjectID       string
	Title           string
	Summary         string
	ExpectedOutcome string
	Approved        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

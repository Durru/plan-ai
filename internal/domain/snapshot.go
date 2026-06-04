package domain

import "time"

// Snapshot captures the state of project decisions, plans, and
// context at a point in time. Snapshots enable rollback,
// comparison, and audit of how the project evolved.
type Snapshot struct {
	ID        string
	ProjectID string
	Reason    string
	Summary   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

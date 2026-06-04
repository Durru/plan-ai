package domain

import "time"

// ConstraintType categorises the kind of limitation.
type ConstraintType string

const (
	ConstraintBudget     ConstraintType = "budget"
	ConstraintStack      ConstraintType = "stack"
	ConstraintTime       ConstraintType = "time"
	ConstraintCompliance ConstraintType = "compliance"
	ConstraintResource   ConstraintType = "resource"
	ConstraintOther      ConstraintType = "other"
)

// Constraint represents a limitation or boundary the project must
// operate within — budget, approved tech stack, timeline, compliance
// requirements, or resource availability.
type Constraint struct {
	ID          string
	ProjectID   string
	Type        ConstraintType
	Description string
	Approved    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

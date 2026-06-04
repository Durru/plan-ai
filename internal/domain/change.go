package domain

import "time"

// ChangeRequestStatus represents the lifecycle state of a change request.
type ChangeRequestStatus string

const (
	ChangeRequestDraft     ChangeRequestStatus = "draft"
	ChangeRequestSubmitted ChangeRequestStatus = "submitted"
	ChangeRequestApproved  ChangeRequestStatus = "approved"
	ChangeRequestRejected  ChangeRequestStatus = "rejected"
	ChangeRequestApplied   ChangeRequestStatus = "applied"
)

// ChangeRequest is a proposal to modify the project's plans,
// decisions, or scope. It captures the reason for the change,
// who requested it, and its current status.
type ChangeRequest struct {
	ID          string
	ProjectID   string
	Reason      string
	Description string
	Status      ChangeRequestStatus
	Requester   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ImpactReport details the entities, plans, and decisions that a
// ChangeRequest would affect if applied. It is always derived from
// a specific ChangeRequest and never exists independently.
type ImpactReport struct {
	ID                string
	ChangeRequestID   string
	AffectedPlans     []string // plan IDs
	AffectedPhases    []string // phase IDs
	AffectedTasks     []string // task IDs
	AffectedDecisions []string // decision IDs
	AffectedEntities  []string // arbitrary entity IDs
	Summary           string
	CreatedAt         time.Time
}

package domain

import "time"

// ChangeRequestStatus represents the lifecycle state of a change request.
// Phase 14 extended from 5 to 10 states.
type ChangeRequestStatus string

const (
	ChangeRequestDraft         ChangeRequestStatus = "draft"
	ChangeRequestSubmitted     ChangeRequestStatus = "submitted"
	ChangeRequestAnalyzing     ChangeRequestStatus = "analyzing"
	ChangeRequestResearchReq   ChangeRequestStatus = "research_required"
	ChangeRequestProposalReady ChangeRequestStatus = "proposal_ready"
	ChangeRequestWaitApproval  ChangeRequestStatus = "waiting_approval"
	ChangeRequestApproved      ChangeRequestStatus = "approved"
	ChangeRequestApplied       ChangeRequestStatus = "applied"
	ChangeRequestValidated     ChangeRequestStatus = "validated"
	ChangeRequestRejected      ChangeRequestStatus = "rejected"
	ChangeRequestCancelled     ChangeRequestStatus = "cancelled"
)

// ValidChangeRequestTransitions returns the allowed status transitions
// for a ChangeRequest. The canonical lifecycle follows the v4 spec:
// draft→submitted→analyzing→(research_required)→proposal_ready→waiting_approval→approved→applied→validated→(done)
func ValidChangeRequestTransitions() map[ChangeRequestStatus][]ChangeRequestStatus {
	return map[ChangeRequestStatus][]ChangeRequestStatus{
		ChangeRequestDraft:         {ChangeRequestSubmitted, ChangeRequestCancelled},
		ChangeRequestSubmitted:     {ChangeRequestAnalyzing, ChangeRequestRejected},
		ChangeRequestAnalyzing:     {ChangeRequestResearchReq, ChangeRequestProposalReady, ChangeRequestRejected},
		ChangeRequestResearchReq:   {ChangeRequestProposalReady, ChangeRequestRejected},
		ChangeRequestProposalReady: {ChangeRequestWaitApproval, ChangeRequestRejected},
		ChangeRequestWaitApproval:  {ChangeRequestApproved, ChangeRequestRejected},
		ChangeRequestApproved:      {ChangeRequestApplied, ChangeRequestCancelled},
		ChangeRequestApplied:       {ChangeRequestValidated, ChangeRequestRejected},
		ChangeRequestValidated:     {},
		ChangeRequestRejected:      {ChangeRequestDraft},
		ChangeRequestCancelled:     {ChangeRequestDraft},
	}
}

// IsValidChangeRequestTransition checks if a transition is allowed.
func IsValidChangeRequestTransition(from, to ChangeRequestStatus) bool {
	for _, allowed := range ValidChangeRequestTransitions()[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

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

// ChangeAudit records a state transition in a change request lifecycle.
type ChangeAudit struct {
	ID              string
	ChangeRequestID string
	ProjectID       string
	Actor           string
	Action          string
	FromState       ChangeRequestStatus
	ToState         ChangeRequestStatus
	Note            string
	CreatedAt       time.Time
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

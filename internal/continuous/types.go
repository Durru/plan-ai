package continuous

import "time"

// ──────────────────────────────────────────────
// Event Types
// ──────────────────────────────────────────────

// EventKind represents the type of event detected.
type EventKind string

const (
	EventNewApprovedContext     EventKind = "new_approved_context"
	EventNewResearch            EventKind = "new_research"
	EventNewKnowledge           EventKind = "new_knowledge"
	EventDecisionChanged        EventKind = "decision_changed"
	EventPlanOutdated           EventKind = "plan_outdated"
	EventImplementationFeedback EventKind = "implementation_feedback"
	EventTaskCompleted          EventKind = "task_completed"
	EventValidationFailed       EventKind = "validation_failed"
	EventChangeRequestCreated   EventKind = "change_request_created"
)

// ContinuousEvent represents a detected event for continuous planning.
type ContinuousEvent struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	EventType EventKind `json:"event_type"`
	Summary   string    `json:"summary"`
	Details   string    `json:"details"` // JSON
	Source    string    `json:"source"`
	CreatedAt string    `json:"created_at"`
}

// ──────────────────────────────────────────────
// Plan Update Proposal Types
// ──────────────────────────────────────────────

// ProposalStatus represents the lifecycle status of a plan update proposal.
type ProposalStatus string

const (
	ProposalDraft           ProposalStatus = "draft"
	ProposalPendingApproval ProposalStatus = "pending_approval"
	ProposalApproved        ProposalStatus = "approved"
	ProposalRejected        ProposalStatus = "rejected"
	ProposalApplied         ProposalStatus = "applied"
)

// PlanUpdateProposal represents a proposed update to plans based on detected events.
type PlanUpdateProposal struct {
	ID                string         `json:"id"`
	ProjectID         string         `json:"project_id"`
	Reason            string         `json:"reason"`
	AffectedPlans     []string       `json:"affected_plans"`
	AffectedTasks     []string       `json:"affected_tasks"`
	AffectedDecisions []string       `json:"affected_decisions"`
	SuggestedUpdates  string         `json:"suggested_updates"`
	RequiresResearch  bool           `json:"requires_research"`
	RequiresApproval  bool           `json:"requires_approval"`
	Status            ProposalStatus `json:"status"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

// ──────────────────────────────────────────────
// Continuous Status Types
// ──────────────────────────────────────────────

// ProjectStatus represents the continuous status of a project.
type ProjectStatus struct {
	ProjectID        string   `json:"project_id"`
	ActivePlan       string   `json:"active_plan"`
	ActivePhase      string   `json:"active_phase"`
	NextTask         string   `json:"next_task"`
	BlockedItems     []string `json:"blocked_items"`
	ApprovalsNeeded  []string `json:"approvals_needed"`
	OutdatedPlans    []string `json:"outdated_plans"`
	RecentEvents     int      `json:"recent_events"`
	PendingProposals int      `json:"pending_proposals"`
}

// ──────────────────────────────────────────────
// Context Levels
// ──────────────────────────────────────────────

// ContextLevel represents the detail level for context serving.
type ContextLevel string

const (
	ContextL0Executive      ContextLevel = "L0_Executive"
	ContextL1Planning       ContextLevel = "L1_Planning"
	ContextL2Plan           ContextLevel = "L2_Specific_Plan"
	ContextL3Task           ContextLevel = "L3_Task"
	ContextL4Implementation ContextLevel = "L4_Implementation"
)

// ContextDelivery represents a delivered context response.
type ContextDelivery struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"project_id"`
	Level     ContextLevel `json:"level"`
	Content   string       `json:"content"`
	CreatedAt string       `json:"created_at"`
}

// ──────────────────────────────────────────────
// Repository Contracts
// ──────────────────────────────────────────────

// ContinuousEventRepository persists continuous events.
type ContinuousEventRepository interface {
	CreateEvent(ev ContinuousEvent) (ContinuousEvent, error)
	ListEvents(projectID string, limit int) ([]ContinuousEvent, error)
}

// PlanUpdateProposalRepository persists plan update proposals.
type PlanUpdateProposalRepository interface {
	CreateProposal(p PlanUpdateProposal) (PlanUpdateProposal, error)
	GetProposal(id string) (PlanUpdateProposal, error)
	ListProposals(projectID string) ([]PlanUpdateProposal, error)
	UpdateProposalStatus(id string, status ProposalStatus) error
}

// ContextDeliveryRepository persists context deliveries.
type ContextDeliveryRepository interface {
	CreateDelivery(d ContextDelivery) (ContextDelivery, error)
	ListDeliveries(projectID string, level ContextLevel, limit int) ([]ContextDelivery, error)
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

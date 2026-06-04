package domain

import "time"

// Plan-specific status values extend the shared Status enum.
const (
	PlanStatusReview    Status = "review"
	PlanStatusPending   Status = "pending" // for phases and tasks
	PlanStatusDone      Status = "done"    // for tasks
	PlanStatusValidated Status = "validated"
	PlanStatusActive    Status = "active"
	PlanStatusCompleted Status = "completed"
)

// MasterPlan is the top-level strategic plan for a project. It defines
// the overall objectives and high-level phases.
type MasterPlan struct {
	ID        string
	ProjectID string
	Title     string
	Summary   string
	Status    Status
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SpecificPlan is a detailed implementation plan derived from a
// MasterPlan. It contains concrete phases and tasks.
type SpecificPlan struct {
	ID           string
	ProjectID    string
	MasterPlanID string
	ParentPlanID string // legacy alias for MasterPlanID; used by store
	Title        string
	Summary      string
	Status       Status
	Version      int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Plan is a unified representation used by the store layer for
// persistence. It maps to the plans table.
type Plan struct {
	ID           string
	Type         PlanType
	Title        string
	Summary      string
	Status       Status
	Version      int
	ParentPlanID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Phase is a major stage within a SpecificPlan. Phases group tasks
// and represent a coherent unit of work.
type Phase struct {
	ID        string
	PlanID    string
	Title     string
	Summary   string
	Status    Status
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Task is the smallest unit of planned work. It belongs to a Phase
// and carries execution metadata.
type Task struct {
	ID          string
	PhaseID     string
	PlanID      string
	Title       string
	Summary     string
	Status      Status
	Position    int
	ContextSize ContextSize
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ImplementationDocument is a deliverable guide derived from a
// SpecificPlan. It captures the implementation instructions,
// architecture decisions, and validation criteria that an implementer
// follows. It is NOT a plan — it is a downstream artifact.
type ImplementationDocument struct {
	ID             string
	SpecificPlanID string
	Title          string
	Content        string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ValidMasterPlanTransitions defines the canonical lifecycle for
// MasterPlan status. The definitive statuses are:
//
//	draft → review → approved → archived
func ValidMasterPlanTransitions(from, to Status) bool {
	switch from {
	case StatusDraft:
		return to == PlanStatusReview
	case PlanStatusReview:
		return to == StatusApproved
	case StatusApproved:
		return to == StatusArchived
	case StatusArchived:
		return false
	default:
		return false
	}
}

// ValidSpecificPlanTransitions defines the canonical lifecycle for
// SpecificPlan status. The definitive statuses are:
//
//	draft → review → approved → blocked → archived
//	blocked → draft (re-plan)
func ValidSpecificPlanTransitions(from, to Status) bool {
	switch from {
	case StatusDraft:
		return to == PlanStatusReview
	case PlanStatusReview:
		return to == StatusApproved || to == StatusDraft
	case StatusApproved:
		return to == StatusBlocked || to == StatusArchived
	case StatusBlocked:
		return to == StatusDraft || to == StatusArchived
	case StatusArchived:
		return false
	default:
		return false
	}
}

// ValidPhaseTransitions defines the canonical lifecycle for
// Phase status. The definitive statuses are:
//
//	pending → active → completed
//	active → blocked
//	blocked → pending
func ValidPhaseTransitions(from, to Status) bool {
	switch from {
	case PlanStatusPending:
		return to == PlanStatusActive
	case PlanStatusActive:
		return to == PlanStatusCompleted || to == StatusBlocked
	case PlanStatusCompleted:
		return false
	case StatusBlocked:
		return to == PlanStatusPending
	default:
		return false
	}
}

// ValidTaskTransitions defines the canonical lifecycle for
// Task status. The definitive statuses are:
//
//	pending → active → done → validated
//	active → blocked
//	blocked → pending
func ValidTaskTransitions(from, to Status) bool {
	switch from {
	case PlanStatusPending:
		return to == PlanStatusActive
	case PlanStatusActive:
		return to == PlanStatusDone || to == StatusBlocked
	case PlanStatusDone:
		return to == PlanStatusValidated
	case PlanStatusValidated:
		return false
	case StatusBlocked:
		return to == PlanStatusPending
	default:
		return false
	}
}

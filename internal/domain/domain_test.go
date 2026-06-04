package domain

import (
	"strings"
	"testing"
	"time"
)

func TestStatusConstantsAreOfficialValues(t *testing.T) {
	tests := []struct {
		name string
		got  Status
		want string
	}{
		{name: "draft", got: StatusDraft, want: "draft"},
		{name: "in review", got: StatusInReview, want: "in_review"},
		{name: "approved", got: StatusApproved, want: "approved"},
		{name: "rejected", got: StatusRejected, want: "rejected"},
		{name: "blocked", got: StatusBlocked, want: "blocked"},
		{name: "implemented", got: StatusImplemented, want: "implemented"},
		{name: "validated", got: StatusValidated, want: "validated"},
		{name: "archived", got: StatusArchived, want: "archived"},
		{name: "completed", got: StatusCompleted, want: "completed"},
		{name: "proposed", got: DecisionProposed, want: "proposed"},
		{name: "deprecated", got: DecisionDeprecated, want: "deprecated"},
		{name: "review", got: PlanStatusReview, want: "review"},
		{name: "pending", got: PlanStatusPending, want: "pending"},
		{name: "done", got: PlanStatusDone, want: "done"},
		{name: "active", got: PlanStatusActive, want: "active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("status = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestResearchStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ResearchStatus
		want string
	}{
		{name: "draft", got: ResearchStatusDraft, want: "draft"},
		{name: "in review", got: ResearchStatusInReview, want: "in_review"},
		{name: "approved", got: ResearchStatusApproved, want: "approved"},
		{name: "rejected", got: ResearchStatusRejected, want: "rejected"},
		{name: "archived", got: ResearchStatusArchived, want: "archived"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("research status = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestKnowledgeTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  KnowledgeType
		want string
	}{
		{name: "research", got: KnowledgeTypeResearch, want: "research"},
		{name: "decision", got: KnowledgeTypeDecision, want: "decision"},
		{name: "requirement", got: KnowledgeTypeRequirement, want: "requirement"},
		{name: "constraint", got: KnowledgeTypeConstraint, want: "constraint"},
		{name: "reference", got: KnowledgeTypeReference, want: "reference"},
		{name: "pattern", got: KnowledgeTypePattern, want: "pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("knowledge type = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestProjectStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ProjectStatus
		want string
	}{
		{name: "draft", got: ProjectStatusDraft, want: "draft"},
		{name: "active", got: ProjectStatusActive, want: "active"},
		{name: "paused", got: ProjectStatusPaused, want: "paused"},
		{name: "completed", got: ProjectStatusCompleted, want: "completed"},
		{name: "archived", got: ProjectStatusArchived, want: "archived"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("project status = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestTypedConstantsAreOfficialValues(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "master plan", got: string(PlanTypeMaster), want: "master"},
		{name: "specific plan", got: string(PlanTypeSpecific), want: "specific"},
		{name: "short context", got: string(ContextSizeShort), want: "short"},
		{name: "medium context", got: string(ContextSizeMedium), want: "medium"},
		{name: "full context", got: string(ContextSizeFull), want: "full"},
		{name: "plan validation", got: string(ValidationTargetPlan), want: "plan"},
		{name: "phase validation", got: string(ValidationTargetPhase), want: "phase"},
		{name: "task validation", got: string(ValidationTargetTask), want: "task"},
		{name: "decision validation", got: string(ValidationTargetDecision), want: "decision"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("value = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestNewIDGeneratesNonEmptyDistinctIDs(t *testing.T) {
	first := NewID("plan")
	second := NewID("plan")

	if first == "" || second == "" {
		t.Fatalf("ids should be non-empty: %q %q", first, second)
	}
	if first == second {
		t.Fatalf("ids should be distinct: %q", first)
	}
	if !strings.HasPrefix(first, "plan_") {
		t.Fatalf("id %q should include prefix", first)
	}
}

func TestDomainEntitiesExposeIdentityAndTimestamps(t *testing.T) {
	now := time.Now().UTC()
	type entityCase struct {
		name string
		id   string
		crt  time.Time
		upd  time.Time
	}
	tests := []entityCase{
		{name: "project", id: Project{ID: "project_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Project{CreatedAt: now}.CreatedAt, upd: Project{UpdatedAt: now}.UpdatedAt},
		{name: "master plan", id: MasterPlan{ID: "plan_1", CreatedAt: now, UpdatedAt: now}.ID, crt: MasterPlan{CreatedAt: now}.CreatedAt, upd: MasterPlan{UpdatedAt: now}.UpdatedAt},
		{name: "specific plan", id: SpecificPlan{ID: "plan_2", CreatedAt: now, UpdatedAt: now}.ID, crt: SpecificPlan{CreatedAt: now}.CreatedAt, upd: SpecificPlan{UpdatedAt: now}.UpdatedAt},
		{name: "phase", id: Phase{ID: "phase_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Phase{CreatedAt: now}.CreatedAt, upd: Phase{UpdatedAt: now}.UpdatedAt},
		{name: "task", id: Task{ID: "task_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Task{CreatedAt: now}.CreatedAt, upd: Task{UpdatedAt: now}.UpdatedAt},
		{name: "decision", id: Decision{ID: "decision_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Decision{CreatedAt: now}.CreatedAt, upd: Decision{UpdatedAt: now}.UpdatedAt},
		{name: "research", id: Research{ID: "research_1", Category: KnowledgeCategoryGeneral, Status: ResearchStatusDraft, CreatedAt: now, UpdatedAt: now}.ID, crt: Research{CreatedAt: now}.CreatedAt, upd: Research{UpdatedAt: now}.UpdatedAt},
		{name: "knowledge object", id: KnowledgeObject{ID: "knowledge_1", CreatedAt: now, UpdatedAt: now}.ID, crt: KnowledgeObject{CreatedAt: now}.CreatedAt, upd: KnowledgeObject{UpdatedAt: now}.UpdatedAt},
		{name: "validation", id: Validation{ID: "validation_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Validation{CreatedAt: now}.CreatedAt, upd: Validation{UpdatedAt: now}.UpdatedAt},
		{name: "snapshot", id: Snapshot{ID: "snapshot_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Snapshot{CreatedAt: now}.CreatedAt, upd: Snapshot{UpdatedAt: now}.UpdatedAt},
		{name: "vision", id: Vision{ID: "vision_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Vision{CreatedAt: now}.CreatedAt, upd: Vision{UpdatedAt: now}.UpdatedAt},
		{name: "requirement", id: Requirement{ID: "req_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Requirement{CreatedAt: now}.CreatedAt, upd: Requirement{UpdatedAt: now}.UpdatedAt},
		{name: "constraint", id: Constraint{ID: "c_1", CreatedAt: now, UpdatedAt: now}.ID, crt: Constraint{CreatedAt: now}.CreatedAt, upd: Constraint{UpdatedAt: now}.UpdatedAt},
		{name: "change request", id: ChangeRequest{ID: "cr_1", CreatedAt: now, UpdatedAt: now}.ID, crt: ChangeRequest{CreatedAt: now}.CreatedAt, upd: ChangeRequest{UpdatedAt: now}.UpdatedAt},
		{name: "impl document", id: ImplementationDocument{ID: "doc_1", CreatedAt: now, UpdatedAt: now}.ID, crt: ImplementationDocument{CreatedAt: now}.CreatedAt, upd: ImplementationDocument{UpdatedAt: now}.UpdatedAt},
	}
	// ImpactReport has no UpdatedAt — add separately with no upd check.
	impactReportCase := entityCase{
		name: "impact report",
		id:   ImpactReport{ID: "ir_1", CreatedAt: now}.ID,
		crt:  ImpactReport{CreatedAt: now}.CreatedAt,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id == "" {
				t.Fatalf("id should be present")
			}
			if tt.crt.IsZero() {
				t.Fatalf("created at should be present")
			}
			if tt.upd.IsZero() {
				t.Fatalf("updated at should be present")
			}
		})
	}

	// ImpactReport has only CreatedAt.
	t.Run("impact report", func(t *testing.T) {
		if impactReportCase.id == "" {
			t.Fatalf("id should be present")
		}
		if impactReportCase.crt.IsZero() {
			t.Fatalf("created at should be present")
		}
	})
}

func TestValidProjectTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ProjectStatus
		to   ProjectStatus
		want bool
	}{
		// Valid transitions
		{name: "draft to active", from: ProjectStatusDraft, to: ProjectStatusActive, want: true},
		{name: "draft to archived", from: ProjectStatusDraft, to: ProjectStatusArchived, want: true},
		{name: "active to paused", from: ProjectStatusActive, to: ProjectStatusPaused, want: true},
		{name: "active to completed", from: ProjectStatusActive, to: ProjectStatusCompleted, want: true},
		{name: "active to archived", from: ProjectStatusActive, to: ProjectStatusArchived, want: true},
		{name: "paused to active", from: ProjectStatusPaused, to: ProjectStatusActive, want: true},
		{name: "paused to completed", from: ProjectStatusPaused, to: ProjectStatusCompleted, want: true},
		{name: "completed to archived", from: ProjectStatusCompleted, to: ProjectStatusArchived, want: true},
		// Prohibited transitions
		{name: "archived to draft", from: ProjectStatusArchived, to: ProjectStatusDraft, want: false},
		{name: "archived to active", from: ProjectStatusArchived, to: ProjectStatusActive, want: false},
		{name: "completed to draft", from: ProjectStatusCompleted, to: ProjectStatusDraft, want: false},
		{name: "completed to active", from: ProjectStatusCompleted, to: ProjectStatusActive, want: false},
		{name: "draft to paused", from: ProjectStatusDraft, to: ProjectStatusPaused, want: false},
		{name: "draft to completed", from: ProjectStatusDraft, to: ProjectStatusCompleted, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidProjectTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidProjectTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidDecisionTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		// Valid transitions
		{name: "proposed to approved", from: DecisionProposed, to: StatusApproved, want: true},
		{name: "proposed to rejected", from: DecisionProposed, to: StatusRejected, want: true},
		{name: "approved to deprecated", from: StatusApproved, to: DecisionDeprecated, want: true},
		{name: "rejected to proposed", from: StatusRejected, to: DecisionProposed, want: true},
		// Prohibited transitions
		{name: "deprecated to anything", from: DecisionDeprecated, to: StatusApproved, want: false},
		{name: "approved to rejected", from: StatusApproved, to: StatusRejected, want: false},
		{name: "approved to proposed", from: StatusApproved, to: DecisionProposed, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidDecisionTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidDecisionTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidMasterPlanTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "draft to review", from: StatusDraft, to: PlanStatusReview, want: true},
		{name: "review to approved", from: PlanStatusReview, to: StatusApproved, want: true},
		{name: "approved to archived", from: StatusApproved, to: StatusArchived, want: true},
		{name: "draft to approved (skip review)", from: StatusDraft, to: StatusApproved, want: false},
		{name: "archived to draft", from: StatusArchived, to: StatusDraft, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidMasterPlanTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidMasterPlanTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidSpecificPlanTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "draft to review", from: StatusDraft, to: PlanStatusReview, want: true},
		{name: "review to approved", from: PlanStatusReview, to: StatusApproved, want: true},
		{name: "review back to draft", from: PlanStatusReview, to: StatusDraft, want: true},
		{name: "approved to blocked", from: StatusApproved, to: StatusBlocked, want: true},
		{name: "approved to archived", from: StatusApproved, to: StatusArchived, want: true},
		{name: "blocked to draft", from: StatusBlocked, to: StatusDraft, want: true},
		{name: "blocked to archived", from: StatusBlocked, to: StatusArchived, want: true},
		{name: "draft to approved (skip review)", from: StatusDraft, to: StatusApproved, want: false},
		{name: "approved to draft (skip blocked)", from: StatusApproved, to: StatusDraft, want: false},
		{name: "archived to draft", from: StatusArchived, to: StatusDraft, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidSpecificPlanTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidSpecificPlanTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidPhaseTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "pending to active", from: PlanStatusPending, to: PlanStatusActive, want: true},
		{name: "active to completed", from: PlanStatusActive, to: PlanStatusCompleted, want: true},
		{name: "active to blocked", from: PlanStatusActive, to: StatusBlocked, want: true},
		{name: "blocked to pending", from: StatusBlocked, to: PlanStatusPending, want: true},
		{name: "completed to active", from: PlanStatusCompleted, to: PlanStatusActive, want: false},
		{name: "pending to completed", from: PlanStatusPending, to: PlanStatusCompleted, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPhaseTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidPhaseTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidTaskTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "pending to active", from: PlanStatusPending, to: PlanStatusActive, want: true},
		{name: "active to done", from: PlanStatusActive, to: PlanStatusDone, want: true},
		{name: "active to blocked", from: PlanStatusActive, to: StatusBlocked, want: true},
		{name: "done to validated", from: PlanStatusDone, to: PlanStatusValidated, want: true},
		{name: "blocked to pending", from: StatusBlocked, to: PlanStatusPending, want: true},
		{name: "validated to done", from: PlanStatusValidated, to: PlanStatusDone, want: false},
		{name: "pending to done", from: PlanStatusPending, to: PlanStatusDone, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidTaskTransitions(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidTaskTransitions(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidationTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ValidationType
		want string
	}{
		{name: "manual", got: ValidationTypeManual, want: "manual"},
		{name: "automatic", got: ValidationTypeAutomatic, want: "automatic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("validation type = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestValidationStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ValidationStatus
		want string
	}{
		{name: "pending", got: ValidationPending, want: "pending"},
		{name: "passed", got: ValidationPassed, want: "passed"},
		{name: "failed", got: ValidationFailed, want: "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("validation status = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestRequirementTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  RequirementType
		want string
	}{
		{name: "functional", got: RequirementTypeFunctional, want: "functional"},
		{name: "ux", got: RequirementTypeUX, want: "ux"},
		{name: "technical", got: RequirementTypeTechnical, want: "technical"},
		{name: "business", got: RequirementTypeBusiness, want: "business"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("requirement type = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestConstraintTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ConstraintType
		want string
	}{
		{name: "budget", got: ConstraintBudget, want: "budget"},
		{name: "stack", got: ConstraintStack, want: "stack"},
		{name: "time", got: ConstraintTime, want: "time"},
		{name: "compliance", got: ConstraintCompliance, want: "compliance"},
		{name: "resource", got: ConstraintResource, want: "resource"},
		{name: "other", got: ConstraintOther, want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("constraint type = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestChangeRequestStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		got  ChangeRequestStatus
		want string
	}{
		{name: "draft", got: ChangeRequestDraft, want: "draft"},
		{name: "submitted", got: ChangeRequestSubmitted, want: "submitted"},
		{name: "approved", got: ChangeRequestApproved, want: "approved"},
		{name: "rejected", got: ChangeRequestRejected, want: "rejected"},
		{name: "applied", got: ChangeRequestApplied, want: "applied"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("change request status = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestRepositoryInterfaceCompileSafety(t *testing.T) {
	// Compile-time assertions that the domain repository interfaces
	// are well-formed. These pass as long as the code compiles.
	var _ ProjectRepository = nil
	var _ VisionRepository = nil
	var _ RequirementRepository = nil
	var _ DecisionRepository = nil
	var _ ResearchRepository = nil
	var _ KnowledgeRepository = nil
	var _ PlanRepository = nil
	var _ TaskRepository = nil
	var _ SnapshotRepository = nil
	var _ ChangeRepository = nil

	_ = t // silence unused lint
}

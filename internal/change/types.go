package change

import "time"

// ChangeType represents the kind of change event.
type ChangeType string

const (
	VisionChanged          ChangeType = "vision_changed"
	RequirementAdded       ChangeType = "requirement_added"
	RequirementRemoved     ChangeType = "requirement_removed"
	ConstraintChanged      ChangeType = "constraint_changed"
	DecisionChanged        ChangeType = "decision_changed"
	ResearchUpdated        ChangeType = "research_updated"
	KnowledgeUpdated       ChangeType = "knowledge_updated"
	PlanChanged            ChangeType = "plan_changed"
	TechnologyChanged      ChangeType = "technology_changed"
	ImplementationFeedback ChangeType = "implementation_feedback"
)

// AllChangeTypes lists every known change type.
var AllChangeTypes = []ChangeType{
	VisionChanged, RequirementAdded, RequirementRemoved,
	ConstraintChanged, DecisionChanged, ResearchUpdated,
	KnowledgeUpdated, PlanChanged, TechnologyChanged, ImplementationFeedback,
}

// Severity indicates how disruptive a change is.
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// ChangeEvent represents a single discrete change event.
type ChangeEvent struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	ChangeType    ChangeType `json:"change_type"`
	Severity      Severity   `json:"severity"`
	Summary       string     `json:"summary"`
	Description   string     `json:"description"`
	EntityType    string     `json:"entity_type"`
	EntityID      string     `json:"entity_id"`
	PreviousState string     `json:"previous_state,omitempty"`
	NewState      string     `json:"new_state,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// EntityStatus represents the invalidation status of a planning entity.
type EntityStatus string

const (
	EntityCurrent     EntityStatus = "current"
	EntityOutdated    EntityStatus = "outdated"
	EntityNeedsReview EntityStatus = "needs_review"
	EntityBlocked     EntityStatus = "blocked"
)

// EntityState tracks the validity state of a planning entity.
type EntityState struct {
	ID           string       `json:"id"`
	EntityType   string       `json:"entity_type"`
	EntityID     string       `json:"entity_id"`
	Status       EntityStatus `json:"status"`
	LastChangeID string       `json:"last_change_id,omitempty"`
	Reason       string       `json:"reason,omitempty"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// InvalidationRule maps entity types to their invalidation behavior.
type InvalidationRule struct {
	EntityType   string       `json:"entity_type"`
	AffectedBy   []ChangeType `json:"affected_by"`
	ResultStatus EntityStatus `json:"result_status"`
}

// DefaultInvalidationRules defines the invalidation matrix.
var DefaultInvalidationRules = []InvalidationRule{
	{EntityType: "vision", AffectedBy: []ChangeType{VisionChanged, ConstraintChanged, DecisionChanged}, ResultStatus: EntityNeedsReview},
	{EntityType: "requirement", AffectedBy: []ChangeType{RequirementAdded, RequirementRemoved, ConstraintChanged, VisionChanged}, ResultStatus: EntityNeedsReview},
	{EntityType: "constraint", AffectedBy: []ChangeType{ConstraintChanged, VisionChanged, DecisionChanged}, ResultStatus: EntityNeedsReview},
	{EntityType: "decision", AffectedBy: []ChangeType{DecisionChanged, ResearchUpdated, KnowledgeUpdated}, ResultStatus: EntityNeedsReview},
	{EntityType: "research", AffectedBy: []ChangeType{ResearchUpdated, KnowledgeUpdated}, ResultStatus: EntityOutdated},
	{EntityType: "knowledge", AffectedBy: []ChangeType{KnowledgeUpdated}, ResultStatus: EntityOutdated},
	{EntityType: "master_plan", AffectedBy: []ChangeType{VisionChanged, RequirementAdded, RequirementRemoved, ConstraintChanged, DecisionChanged, ResearchUpdated, KnowledgeUpdated, PlanChanged}, ResultStatus: EntityNeedsReview},
	{EntityType: "specific_plan", AffectedBy: []ChangeType{RequirementAdded, RequirementRemoved, ConstraintChanged, DecisionChanged, ResearchUpdated, KnowledgeUpdated, PlanChanged, ImplementationFeedback}, ResultStatus: EntityNeedsReview},
	{EntityType: "phase", AffectedBy: []ChangeType{PlanChanged, RequirementAdded, RequirementRemoved, ImplementationFeedback}, ResultStatus: EntityNeedsReview},
	{EntityType: "task", AffectedBy: []ChangeType{PlanChanged, RequirementRemoved, DecisionChanged, ImplementationFeedback}, ResultStatus: EntityBlocked},
	{EntityType: "validation", AffectedBy: []ChangeType{PlanChanged, RequirementRemoved, DecisionChanged, ImplementationFeedback}, ResultStatus: EntityNeedsReview},
}

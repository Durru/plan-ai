package domain

import "time"

// RequirementType classifies requirements by domain concern.
type RequirementType string

const (
	RequirementTypeFunctional RequirementType = "functional"
	RequirementTypeUX         RequirementType = "ux"
	RequirementTypeTechnical  RequirementType = "technical"
	RequirementTypeBusiness   RequirementType = "business"
)

// Requirement is a single desired capability, constraint, or quality
// that the project must satisfy. Requirements are typed so the planner
// can group, prioritize, and trace them through implementation.
type Requirement struct {
	ID        string
	ProjectID string
	Type      RequirementType
	Statement string
	Approved  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

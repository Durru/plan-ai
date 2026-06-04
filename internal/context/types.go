package context

import "time"

type ApprovedType string

const (
	TypeRequirement ApprovedType = "requirement"
	TypeConstraint  ApprovedType = "constraint"
	TypeDecision    ApprovedType = "decision"
	TypePreference  ApprovedType = "preference"
	TypeGoal        ApprovedType = "goal"
	TypeReference   ApprovedType = "reference"
)

type State string

const StateApproved State = "approved"

type ApprovedItem struct {
	ID        string
	ProjectID string
	Type      ApprovedType
	SourceID  string
	Content   string
	State     State
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ApprovedRequirement = ApprovedItem
type ApprovedConstraint = ApprovedItem
type ApprovedDecision = ApprovedItem
type ApprovedPreference = ApprovedItem
type ApprovedReference = ApprovedItem
type ApprovedGoal = ApprovedItem

type Repository interface {
	StoreApproved(ApprovedItem) (ApprovedItem, error)
	GetApproved(ApprovedType, string) (ApprovedItem, error)
	ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error)
	FindApproved(projectID string, itemType ApprovedType, query string) ([]ApprovedItem, error)
}

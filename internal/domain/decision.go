package domain

import "time"

// Decision-specific status values extend the shared Status enum.
// They represent the canonical lifecycle for architectural and
// design decisions.
const (
	DecisionProposed   Status = "proposed"
	DecisionDeprecated Status = "deprecated"
)

// Decision records a design or architectural choice with its
// rationale, alternatives, and current approval status.
type Decision struct {
	ID           string
	ProjectID    string
	Title        string
	Context      string
	Decision     string
	Rationale    string
	Alternatives string
	Status       Status
	Impact       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ValidDecisionTransitions returns the allowed status transitions
// for a Decision. The canonical lifecycle is:
//
//	proposed → approved | rejected
//	approved → deprecated
//	rejected → proposed (reconsider)
//	deprecated → (terminal)
func ValidDecisionTransitions(from, to Status) bool {
	switch from {
	case DecisionProposed:
		return to == StatusApproved || to == StatusRejected
	case StatusApproved:
		return to == DecisionDeprecated
	case StatusRejected:
		return to == DecisionProposed
	case DecisionDeprecated:
		return false // terminal
	default:
		return false
	}
}

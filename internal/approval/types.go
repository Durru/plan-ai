package approval

import "time"

type State string

const (
	StateDraft         State = "draft"
	StateReview        State = "review"
	StateClarification State = "clarification"
	StateApproved      State = "approved"
	StateRejected      State = "rejected"
)

type Record struct {
	ID         string
	ProjectID  string
	TargetType string
	TargetID   string
	State      State
	Reason     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Repository interface {
	SaveRecord(Record) (Record, error)
	ListRecords(projectID string) ([]Record, error)
	ApproveRecord(id string) (Record, error)
	RejectRecord(id, reason string) (Record, error)
}

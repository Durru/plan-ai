package vision

import "time"

type Draft struct {
	ID                 string
	ProjectID          string
	Title              string
	Summary            string
	TargetUsers        []string
	ExpectedOutcome    string
	FunctionalGoals    []string
	UXGoals            []string
	BusinessGoals      []string
	Constraints        []string
	Assumptions        []string
	MissingInformation []string
	VisualReferences   []string
	SuccessCriteria    []string
	Approved           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Repository interface {
	SaveVision(Draft) (Draft, error)
	GetVision(id string) (Draft, error)
	ListVisions(projectID string) ([]Draft, error)
	ApproveVision(id string) (Draft, error)
}

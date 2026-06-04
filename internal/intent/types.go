package intent

import "time"

type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
)

type SignalState string

const (
	SignalCandidate SignalState = "candidate"
	SignalApproved  SignalState = "approved"
)

type Intent struct {
	Name       string      `json:"name"`
	Confidence int         `json:"confidence"`
	State      SignalState `json:"state"`
}

type Goal struct {
	Name  string      `json:"name"`
	State SignalState `json:"state"`
}

type UserExpectation struct {
	Name  string      `json:"name"`
	State SignalState `json:"state"`
}

type SuccessCriteria struct {
	Name  string      `json:"name"`
	State SignalState `json:"state"`
}

type UserPriority struct {
	Name  string      `json:"name"`
	Rank  int         `json:"rank"`
	State SignalState `json:"state"`
}

type Profile struct {
	ID              string
	ProjectID       string
	Source          string
	PrimaryIntent   Intent
	SecondaryGoals  []Goal
	Constraints     []string
	Expectations    []UserExpectation
	SuccessCriteria []SuccessCriteria
	Priorities      []UserPriority
	Status          Status
	Approved        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Repository interface {
	SaveProfile(Profile) (Profile, error)
	GetProfile(id string) (Profile, error)
	LatestProfile(projectID string) (Profile, error)
	ApproveProfile(id string) (Profile, error)
}

package requirements

import (
	"strings"
	"time"
)

type State string

const (
	StateCandidate State = "candidate"
	StateApproved  State = "approved"
)

type Candidate struct {
	ID           string
	ProjectID    string
	Source       string
	Name         string
	Description  string
	Reason       string
	Dependencies []string
	Ambiguities  []string
	State        State
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	SaveCandidates([]Candidate) ([]Candidate, error)
	ListCandidates(projectID string) ([]Candidate, error)
	ApproveCandidate(id string) (Candidate, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) Service { return Service{repo: repo} }

func (s Service) Discover(projectID, content string) ([]Candidate, error) {
	return s.repo.SaveCandidates(Discover(projectID, content))
}

func (s Service) List(projectID string) ([]Candidate, error) { return s.repo.ListCandidates(projectID) }

func (s Service) Approve(id string) (Candidate, error) { return s.repo.ApproveCandidate(id) }

func Discover(projectID, content string) []Candidate {
	normalized := strings.ToLower(content)
	names := []string{"roles and permissions", "analytics", "audit trail"}
	if strings.Contains(normalized, "ecommerce") || strings.Contains(normalized, "tienda") {
		names = []string{"cart", "checkout", "coupons", "SEO", "blog", "analytics", "inventory", "payments", "tax", "fulfillment"}
	}
	if strings.Contains(normalized, "crm") {
		names = []string{"contacts", "pipeline", "admin panel", "reports", "automations", "subscriptions", "team roles"}
	}
	out := make([]Candidate, 0, len(names))
	for _, name := range names {
		out = append(out, Candidate{ProjectID: projectID, Source: content, Name: name, Description: "Candidate requirement requiring approval: " + name, Reason: "Discovered from user intent, not yet approved", State: StateCandidate})
	}
	return out
}

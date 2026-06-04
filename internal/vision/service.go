package vision

import "github.com/plan-ai/plan-ai/internal/ingestion"

type Service struct{ repo Repository }

func NewService(repo Repository) Service { return Service{repo: repo} }

func (s Service) CreateDraft(projectID string, sources []ingestion.IngestedSource) (Draft, error) {
	draft := Extract(projectID, sources)
	draft.Approved = false
	return s.repo.SaveVision(draft)
}

func (s Service) Approve(id string) (Draft, error) { return s.repo.ApproveVision(id) }

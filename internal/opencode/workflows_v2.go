package opencode

import (
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type WorkflowCommand struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Description string   `json:"description"`
	ReadOnly    bool     `json:"read_only"`
	Args        []string `json:"args"`
}

type WorkflowRegistration struct {
	ID        string
	ProjectID string
	Commands  []WorkflowCommand
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkflowRepository interface {
	SaveWorkflowRegistration(WorkflowRegistration) (WorkflowRegistration, error)
	ListWorkflowRegistrations(projectID string) ([]WorkflowRegistration, error)
}

type WorkflowService struct{ repo WorkflowRepository }

func NewWorkflowService(repo WorkflowRepository) WorkflowService { return WorkflowService{repo: repo} }

func (s WorkflowService) Register(projectID string) (WorkflowRegistration, error) {
	return s.repo.SaveWorkflowRegistration(BuildWorkflowRegistration(projectID))
}

func (s WorkflowService) List(projectID string) ([]WorkflowRegistration, error) {
	return s.repo.ListWorkflowRegistrations(projectID)
}

func BuildWorkflowRegistration(projectID string) WorkflowRegistration {
	return WorkflowRegistration{ID: domain.NewID("ocwf"), ProjectID: projectID, Commands: DefaultWorkflowCommands(), Status: "synced"}
}

func DefaultWorkflowCommands() []WorkflowCommand {
	return []WorkflowCommand{
		{Name: "status", Command: "plan-ai status", Description: "Read Plan-AI project status", ReadOnly: true},
		{Name: "next", Command: "plan-ai next", Description: "Read the next actionable task", ReadOnly: true},
		{Name: "context", Command: "plan-ai context", Description: "Read executive context", ReadOnly: true, Args: []string{"level"}},
		{Name: "plans", Command: "plan-ai plan blueprints", Description: "Read Plan Evolution V3 blueprints", ReadOnly: true},
		{Name: "changes", Command: "plan-ai impact reports-v2", Description: "Read Change Impact V2 reports", ReadOnly: true},
	}
}

package workflows

import "time"

type WorkflowType string
type RunStatus string

const (
	WorkflowTypeVision   WorkflowType = "vision"
	WorkflowTypeResearch WorkflowType = "research"
	WorkflowTypePlanning WorkflowType = "planning"
	WorkflowTypeApproval WorkflowType = "approval"

	StatusRunning   RunStatus = "running"
	StatusCompleted RunStatus = "completed"
	StatusFailed    RunStatus = "failed"
	StatusRejected  RunStatus = "rejected"
)

type Workflow struct {
	Type  WorkflowType
	Name  string
	Steps []string
}

type WorkflowRun struct {
	ID           string
	WorkflowType WorkflowType
	Status       RunStatus
	StartedAt    time.Time
	FinishedAt   time.Time
}

type RunRepository interface {
	CreateWorkflowRun(WorkflowRun) (WorkflowRun, error)
	UpdateWorkflowRun(WorkflowRun) (WorkflowRun, error)
	GetWorkflowRun(string) (WorkflowRun, error)
}

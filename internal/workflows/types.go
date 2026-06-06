package workflows

import (
	"database/sql"
	"time"
)

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

type Step struct {
	Name        string     `json:"name"`
	Status      RunStatus  `json:"status"`
	Error       string     `json:"error,omitempty"`
	Output      string     `json:"output,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type WorkflowRun struct {
	ID           string
	WorkflowType WorkflowType
	Status       RunStatus
	Steps        []Step
	StartedAt    time.Time
	FinishedAt   time.Time
}

type RunRepository interface {
	CreateWorkflowRun(WorkflowRun) (WorkflowRun, error)
	UpdateWorkflowRun(WorkflowRun) (WorkflowRun, error)
	GetWorkflowRun(string) (WorkflowRun, error)
}

type StepExecutor func(run *WorkflowRun, stepName string, db *sql.DB) error

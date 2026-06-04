package orchestrator

import "time"

// ──────────────────────────────────────────────
// Workflow / job types
// ──────────────────────────────────────────────

type WorkflowType string

const (
	WorkflowVision   WorkflowType = "vision"
	WorkflowResearch WorkflowType = "research"
	WorkflowPlanning WorkflowType = "planning"
	WorkflowApproval WorkflowType = "approval"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// ──────────────────────────────────────────────
// Job model
// ──────────────────────────────────────────────

type Job struct {
	ID           string       `json:"id"`
	WorkflowType WorkflowType `json:"workflow_type"`
	Capability   string       `json:"capability"`
	Strategy     string       `json:"strategy"`
	Status       JobStatus    `json:"status"`
	ProjectID    string       `json:"project_id"`
	Error        string       `json:"error,omitempty"`
	StartedAt    time.Time    `json:"started_at"`
	FinishedAt   time.Time    `json:"finished_at,omitempty"`
}

// ──────────────────────────────────────────────
// JobRun — individual execution record for a job
// ──────────────────────────────────────────────

type JobRun struct {
	ID         string    `json:"id"`
	JobID      string    `json:"job_id"`
	Step       string    `json:"step"`
	Status     JobStatus `json:"status"`
	Output     string    `json:"output,omitempty"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

// ──────────────────────────────────────────────
// Repository interfaces for persistence
// ──────────────────────────────────────────────

type JobRepository interface {
	CreateJob(Job) (Job, error)
	GetJob(id string) (Job, error)
	UpdateJob(Job) (Job, error)
	ListJobs(projectID string) ([]Job, error)
}

type JobRunRepository interface {
	CreateJobRun(JobRun) (JobRun, error)
	UpdateJobRun(JobRun) (JobRun, error)
	ListJobRuns(jobID string) ([]JobRun, error)
}

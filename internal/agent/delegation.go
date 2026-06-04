package agent

import (
	"database/sql"
	"fmt"
	"time"
)

// StoreDelegator creates and manages delegated jobs via the store.
type StoreDelegator struct {
	db   *sql.DB
	repo DelegatedJobRepository
}

// NewDelegator creates a new StoreDelegator.
func NewDelegator(db *sql.DB, repo DelegatedJobRepository) *StoreDelegator {
	return &StoreDelegator{db: db, repo: repo}
}

// CreateJob creates a new delegated job.
func (d *StoreDelegator) CreateJob(job DelegatedJob) (DelegatedJob, error) {
	job.Status = JobStatusPending
	job.CreatedAt = nowUTC()
	return d.repo.CreateJob(job)
}

// GetJob retrieves a delegated job by ID.
func (d *StoreDelegator) GetJob(id string) (DelegatedJob, error) {
	return d.repo.GetJob(id)
}

// ListJobs lists delegated jobs for a project.
func (d *StoreDelegator) ListJobs(projectID string) ([]DelegatedJob, error) {
	return d.repo.ListJobs(projectID)
}

// UpdateJobStatus updates a job's status and optional summary.
func (d *StoreDelegator) UpdateJobStatus(id string, status JobStatus, summary string) error {
	completedAt := ""
	if status == JobStatusCompleted || status == JobStatusFailed || status == JobStatusCancelled {
		completedAt = time.Now().UTC().Format(time.RFC3339)
		_ = completedAt // we update via the repo
	}
	return d.repo.UpdateJob(id, string(status), summary)
}

// JobForIntent creates a DelegatedJob for the given intent.
func JobForIntent(projectID string, intent IntentKind, workflowType string) DelegatedJob {
	jobType := jobTypeForIntent(intent)
	return DelegatedJob{
		ID:           fmt.Sprintf("job_%d", time.Now().UnixNano()),
		ProjectID:    projectID,
		Intent:       intent,
		Capability:   string(capabilityForIntent(intent)),
		WorkflowType: workflowType,
		JobType:      jobType,
	}
}

func jobTypeForIntent(intent IntentKind) DelegatedJobType {
	switch intent {
	case IntentCreateMasterPlan, IntentCreateSpecificPlan:
		return JobTypePlanning
	case IntentResearchTopic:
		return JobTypeResearch
	case IntentChangeRequest:
		return JobTypeImpact
	case IntentApprove, IntentReject, IntentValidate:
		return JobTypeValidation
	default:
		return JobTypeContext
	}
}

func capabilityForIntent(intent IntentKind) string {
	switch intent {
	case IntentCreateMasterPlan, IntentCreateSpecificPlan:
		return CapabilityPlanning
	case IntentResearchTopic:
		return CapabilityResearch
	case IntentChangeRequest:
		return CapabilityChange
	case IntentProjectStatus, IntentNextTask:
		return CapabilityContext
	default:
		return CapabilityContext
	}
}

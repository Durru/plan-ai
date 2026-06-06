package orchestrator

import (
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/capabilities"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/modelstrategy"
)

// Orchestrator coordinates job creation, capability selection,
// model strategy selection, and workflow dispatch.
type Orchestrator struct {
	capRegistry   *capabilities.Registry
	modelStrategy *modelstrategy.Service
	jobs          JobRepository
	runs          JobRunRepository
	now           func() time.Time
}

// NewOrchestrator creates a new orchestrator with the given dependencies.
func NewOrchestrator(
	capRegistry *capabilities.Registry,
	modelStrategy *modelstrategy.Service,
	jobs JobRepository,
	runs JobRunRepository,
) *Orchestrator {
	return &Orchestrator{
		capRegistry:   capRegistry,
		modelStrategy: modelStrategy,
		jobs:          jobs,
		runs:          runs,
		now:           time.Now().UTC,
	}
}

// CreateJob creates a new job record.
func (o *Orchestrator) CreateJob(projectID string, wfType WorkflowType, capability string) (Job, error) {
	if projectID == "" {
		return Job{}, fmt.Errorf("project id is required")
	}
	if wfType == "" {
		return Job{}, fmt.Errorf("workflow type is required")
	}

	job := Job{
		ID:           domain.NewID("job"),
		WorkflowType: wfType,
		Capability:   capability,
		Status:       JobStatusPending,
		ProjectID:    projectID,
		StartedAt:    o.now(),
	}

	// If a capability was given, validate it
	if capability != "" {
		if _, err := o.capRegistry.GetCapability(capabilities.CapabilityType(capability)); err != nil {
			return Job{}, fmt.Errorf("capability check: %w", err)
		}
	}

	return o.jobs.CreateJob(job)
}

// GetJob returns a job by ID.
func (o *Orchestrator) GetJob(id string) (Job, error) {
	return o.jobs.GetJob(id)
}

// ExecuteJob transitions a job to running, executes it, and updates status.
func (o *Orchestrator) ExecuteJob(jobID string) (Job, error) {
	job, err := o.jobs.GetJob(jobID)
	if err != nil {
		return Job{}, fmt.Errorf("get job: %w", err)
	}

	job.Status = JobStatusRunning
	job.StartedAt = o.now()
	if _, err := o.jobs.UpdateJob(job); err != nil {
		return Job{}, err
	}

	return job, nil
}

// CompleteJob marks a job as completed.
func (o *Orchestrator) CompleteJob(jobID string) (Job, error) {
	job, err := o.jobs.GetJob(jobID)
	if err != nil {
		return Job{}, err
	}
	job.Status = JobStatusCompleted
	job.FinishedAt = o.now()
	return o.jobs.UpdateJob(job)
}

// FailJob marks a job as failed with an error message.
func (o *Orchestrator) FailJob(jobID, errMsg string) (Job, error) {
	job, err := o.jobs.GetJob(jobID)
	if err != nil {
		return Job{}, err
	}
	job.Status = JobStatusFailed
	job.Error = errMsg
	job.FinishedAt = o.now()
	return o.jobs.UpdateJob(job)
}

// DispatchWorkflow creates and executes a job for the given workflow type.
func (o *Orchestrator) DispatchWorkflow(projectID string, wfType WorkflowType) (Job, error) {
	capType := workflowToCapability(wfType)

	job, err := o.CreateJob(projectID, wfType, string(capType))
	if err != nil {
		return Job{}, err
	}

	job, err = o.ExecuteJob(job.ID)
	if err != nil {
		return Job{}, err
	}

	return job, nil
}

// SelectCapability finds the appropriate capability for a workflow type.
func (o *Orchestrator) SelectCapability(wfType WorkflowType) (capabilities.Capability, error) {
	capType := workflowToCapability(wfType)
	return o.capRegistry.GetCapability(capType)
}

// SelectModelStrategy selects the appropriate model strategy for a task classification.
func (o *Orchestrator) SelectModelStrategy(classification modelstrategy.TaskClassification) (modelstrategy.ModelStrategy, error) {
	return o.modelStrategy.Selector.Select(classification)
}

// LoadContext is a placeholder for context loading (Phase 17 integration).
func (o *Orchestrator) LoadContext(_ string, _ string) string {
	return "context loading delegated to context engine"
}

// PersistResults creates a job run record for the given job and output.
func (o *Orchestrator) PersistResults(jobID, step, output string) (JobRun, error) {
	run := JobRun{
		ID:        domain.NewID("run"),
		JobID:     jobID,
		Step:      step,
		Status:    JobStatusCompleted,
		Output:    output,
		StartedAt: o.now(),
	}
	return o.runs.CreateJobRun(run)
}

// workflowToCapability maps a workflow type to its primary capability.
func workflowToCapability(wfType WorkflowType) capabilities.CapabilityType {
	switch wfType {
	case WorkflowVision:
		return capabilities.CapVision
	case WorkflowResearch:
		return capabilities.CapResearch
	case WorkflowPlanning:
		return capabilities.CapPlanning
	case WorkflowApproval:
		return capabilities.CapValidation
	default:
		return capabilities.CapPlanning
	}
}

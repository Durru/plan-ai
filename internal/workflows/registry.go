package workflows

import (
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type Registry struct {
	repo      RunRepository
	workflows map[WorkflowType]Workflow
	now       func() time.Time
}

func NewRegistry(repo RunRepository) *Registry {
	r := &Registry{repo: repo, workflows: map[WorkflowType]Workflow{}, now: time.Now().UTC}
	_ = r.RegisterWorkflow(VisionWorkflow())
	_ = r.RegisterWorkflow(ResearchWorkflow())
	_ = r.RegisterWorkflow(PlanningWorkflow())
	_ = r.RegisterWorkflow(ApprovalWorkflow())
	return r
}

func (r *Registry) RegisterWorkflow(workflow Workflow) error {
	if workflow.Type == "" {
		return fmt.Errorf("workflow type is required")
	}
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	r.workflows[workflow.Type] = workflow
	return nil
}

func (r *Registry) GetWorkflow(typ WorkflowType) (Workflow, error) {
	w, ok := r.workflows[typ]
	if !ok {
		return Workflow{}, fmt.Errorf("workflow %q not registered", typ)
	}
	return w, nil
}

func (r *Registry) ExecuteWorkflow(typ WorkflowType) (WorkflowRun, error) {
	if _, err := r.GetWorkflow(typ); err != nil {
		return WorkflowRun{}, err
	}
	now := r.now()
	run := WorkflowRun{ID: domain.NewID("workflow"), WorkflowType: typ, Status: StatusRunning, StartedAt: now}
	created, err := r.repo.CreateWorkflowRun(run)
	if err != nil {
		return WorkflowRun{}, err
	}
	created.Status = StatusCompleted
	created.FinishedAt = r.now()
	return r.repo.UpdateWorkflowRun(created)
}

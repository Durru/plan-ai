package workflows

import "testing"

type memoryRunRepository struct{ runs []WorkflowRun }

func (r *memoryRunRepository) CreateWorkflowRun(run WorkflowRun) (WorkflowRun, error) {
	r.runs = append(r.runs, run)
	return run, nil
}
func (r *memoryRunRepository) UpdateWorkflowRun(run WorkflowRun) (WorkflowRun, error) {
	r.runs[len(r.runs)-1] = run
	return run, nil
}
func (r *memoryRunRepository) GetWorkflowRun(id string) (WorkflowRun, error) { return r.runs[0], nil }

func TestRegistryRegistersAndExecutesWorkflow(t *testing.T) {
	repo := &memoryRunRepository{}
	registry := NewRegistry(repo)
	workflow := Workflow{Type: WorkflowTypeResearch, Name: "Research Workflow", Steps: []string{"Topic", "Research", "Knowledge"}}
	if err := registry.RegisterWorkflow(workflow); err != nil {
		t.Fatalf("RegisterWorkflow error: %v", err)
	}
	got, err := registry.GetWorkflow(WorkflowTypeResearch)
	if err != nil {
		t.Fatalf("GetWorkflow error: %v", err)
	}
	if got.Name != workflow.Name {
		t.Fatalf("unexpected workflow: %+v", got)
	}

	run, err := registry.ExecuteWorkflow(WorkflowTypeResearch)
	if err != nil {
		t.Fatalf("ExecuteWorkflow error: %v", err)
	}
	if run.WorkflowType != WorkflowTypeResearch || run.Status != StatusCompleted || run.FinishedAt.IsZero() {
		t.Fatalf("unexpected run: %+v", run)
	}
	if len(repo.runs) != 1 {
		t.Fatalf("expected persisted run, got %d", len(repo.runs))
	}
}

func TestNewRegistryIncludesCoreWorkflows(t *testing.T) {
	registry := NewRegistry(&memoryRunRepository{})
	for _, typ := range []WorkflowType{WorkflowTypeVision, WorkflowTypeResearch, WorkflowTypePlanning, WorkflowTypeApproval} {
		if _, err := registry.GetWorkflow(typ); err != nil {
			t.Fatalf("missing workflow %s: %v", typ, err)
		}
	}
}

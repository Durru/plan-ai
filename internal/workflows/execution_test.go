package workflows

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestExecuteStepsRunsAllSteps(t *testing.T) {
	run := &WorkflowRun{
		ID:           "test-run-1",
		WorkflowType: WorkflowTypeVision,
		Steps: []Step{
			{Name: "detect_intent"},
			{Name: "create_discovery"},
			{Name: "approve_intent"},
		},
	}

	err := ExecuteSteps(run, nil)
	if err != nil {
		t.Fatalf("ExecuteSteps returned error: %v", err)
	}

	if run.Status != StatusCompleted {
		t.Fatalf("expected run status %q, got %q", StatusCompleted, run.Status)
	}

	if len(run.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(run.Steps))
	}

	for i, step := range run.Steps {
		if step.Status != StatusCompleted {
			t.Errorf("step %d (%s): expected status %q, got %q", i, step.Name, StatusCompleted, step.Status)
		}
		if step.CompletedAt == nil {
			t.Errorf("step %d (%s): expected CompletedAt to be set, got nil", i, step.Name)
		}
		if step.Output == "" {
			t.Errorf("step %d (%s): expected output to be non-empty", i, step.Name)
		}
	}
}

func TestExecuteStepsFailsOnStepFailure(t *testing.T) {
	origDispatch := dispatchStep
	defer func() { dispatchStep = origDispatch }()

	callCount := 0
	dispatchStep = func(stepType string, db *sql.DB) (string, error) {
		callCount++
		if callCount == 2 {
			return "", fmt.Errorf("simulated failure in step 2")
		}
		return "ok", nil
	}

	run := &WorkflowRun{
		ID:           "test-run-2",
		WorkflowType: WorkflowTypeVision,
		Steps: []Step{
			{Name: "step_one"},
			{Name: "step_two"},
			{Name: "step_three"},
		},
	}

	err := ExecuteSteps(run, nil)
	if err == nil {
		t.Fatal("expected error from ExecuteSteps, got nil")
	}

	if run.Status != StatusFailed {
		t.Fatalf("expected run status %q, got %q", StatusFailed, run.Status)
	}

	if run.Steps[0].Status != StatusCompleted {
		t.Errorf("step 1: expected %q, got %q", StatusCompleted, run.Steps[0].Status)
	}
	if run.Steps[0].CompletedAt == nil {
		t.Error("step 1: expected CompletedAt to be set")
	}

	if run.Steps[1].Status != StatusFailed {
		t.Errorf("step 2: expected %q, got %q", StatusFailed, run.Steps[1].Status)
	}
	if run.Steps[1].Error == "" {
		t.Error("step 2: expected error message")
	}

	if run.Steps[2].Status != "" {
		t.Errorf("step 3: expected status to be unchanged (empty), got %q", run.Steps[2].Status)
	}
	if run.Steps[2].CompletedAt != nil {
		t.Error("step 3: expected CompletedAt to be nil (skipped)")
	}
	if run.Steps[2].Output != "" {
		t.Errorf("step 3: expected no output, got %q", run.Steps[2].Output)
	}
}

func TestExecuteStepsStepTypes(t *testing.T) {
	allStepTypes := []string{
		"detect_intent",
		"find_reusable",
		"create_master_plan",
		"create_specific_plan",
		"approve_plans",
		"create_discovery",
		"approve_intent",
		"check_requirements",
		"validate",
		"approve_reject",
		"create_research",
		"approve_research",
		"promote_to_knowledge",
		"load_approved_context",
	}

	for _, stepType := range allStepTypes {
		t.Run(stepType, func(t *testing.T) {
			run := &WorkflowRun{
				ID:           "test-run-" + stepType,
				WorkflowType: WorkflowTypePlanning,
				Steps: []Step{
					{Name: stepType},
				},
			}

			err := ExecuteSteps(run, nil)
			if err != nil {
				t.Fatalf("ExecuteSteps for %q returned error: %v", stepType, err)
			}

			step := run.Steps[0]
			if step.Status != StatusCompleted {
				t.Errorf("expected status %q, got %q", StatusCompleted, step.Status)
			}
			if step.Output == "" {
				t.Errorf("expected non-empty output for step %q", stepType)
			}
			if step.CompletedAt == nil {
				t.Errorf("expected CompletedAt to be set for step %q", stepType)
			}
		})
	}
}

func TestDispatchStepUnknownType(t *testing.T) {
	output, err := realDispatchStep("nonexistent_step", nil)
	if err != nil {
		t.Fatalf("unknown step should not return error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output for unknown step type")
	}
}

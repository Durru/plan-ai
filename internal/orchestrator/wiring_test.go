package orchestrator_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Durru/plan-ai/internal/capabilities"
	"github.com/Durru/plan-ai/internal/modelstrategy"
	"github.com/Durru/plan-ai/internal/orchestrator"
	"github.com/Durru/plan-ai/internal/store"
)

func TestOrchestratorCreateJob_ValidatesInput(t *testing.T) {
	orch := newWiringOrchestrator(t)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "planning")
	if err != nil {
		t.Fatalf("CreateJob valid: %v", err)
	}
	if job.ID == "" || job.Status != orchestrator.JobStatusPending || job.ProjectID != "project:test" {
		t.Errorf("job = %+v, want pending with valid fields", job)
	}

	_, err = orch.CreateJob("", orchestrator.WorkflowPlanning, "planning")
	if err == nil || !strings.Contains(err.Error(), "project") {
		t.Errorf("empty projectID: want error, got %v", err)
	}

	_, err = orch.CreateJob("project:test", "", "planning")
	if err == nil || !strings.Contains(err.Error(), "workflow") {
		t.Errorf("empty workflowType: want error, got %v", err)
	}

	_, err = orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "capability") {
		t.Errorf("unknown capability: want error, got %v", err)
	}
}

func TestOrchestratorExecute_UsesCapabilityAndStrategy(t *testing.T) {
	orch := newWiringOrchestrator(t)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowResearch, "research")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.Capability != "research" {
		t.Errorf("capability = %q, want research", job.Capability)
	}

	executed, err := orch.ExecuteJob(job.ID)
	if err != nil {
		t.Fatalf("ExecuteJob: %v", err)
	}
	if executed.Status != orchestrator.JobStatusRunning {
		t.Errorf("after execute: status = %q, want running", executed.Status)
	}

	completed, err := orch.CompleteJob(job.ID)
	if err != nil {
		t.Fatalf("CompleteJob: %v", err)
	}
	if completed.Status != orchestrator.JobStatusCompleted {
		t.Errorf("after complete: status = %q, want completed", completed.Status)
	}
	if completed.FinishedAt.IsZero() {
		t.Error("FinishedAt is zero after CompleteJob")
	}

	job2, err := orch.CreateJob("project:test", orchestrator.WorkflowVision, "vision")
	if err != nil {
		t.Fatalf("CreateJob 2: %v", err)
	}
	failed, err := orch.FailJob(job2.ID, "capability timed out")
	if err != nil {
		t.Fatalf("FailJob: %v", err)
	}
	if failed.Status != orchestrator.JobStatusFailed || failed.Error != "capability timed out" {
		t.Errorf("after fail: status=%q err=%q", failed.Status, failed.Error)
	}
}

func TestOrchestratorIsMCPEntryPoint(t *testing.T) {
	orch := newWiringOrchestrator(t)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "planning")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID == "" {
		t.Fatal("CreateJob returned empty ID")
	}

	got, err := orch.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.ID != job.ID || got.Status != orchestrator.JobStatusPending {
		t.Errorf("GetJob mismatch: id=%q status=%q", got.ID, got.Status)
	}
}

func newWiringOrchestrator(t *testing.T) *orchestrator.Orchestrator {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return orchestrator.NewOrchestrator(
		capabilities.NewDefaultRegistry(db),
		modelstrategy.NewService(),
		store.NewJobRepository(db),
		store.NewJobRunRepository(db),
	)
}

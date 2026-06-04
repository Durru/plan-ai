package orchestrator_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/plan-ai/plan-ai/internal/capabilities"
	"github.com/plan-ai/plan-ai/internal/modelstrategy"
	"github.com/plan-ai/plan-ai/internal/orchestrator"
	"github.com/plan-ai/plan-ai/internal/store"
)

func TestOrchestratorCreateJob(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "planning")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID == "" {
		t.Fatal("job has empty ID")
	}
	if job.Status != orchestrator.JobStatusPending {
		t.Errorf("status = %q, want pending", job.Status)
	}
	if job.ProjectID != "project:test" {
		t.Errorf("project_id = %q", job.ProjectID)
	}
}

func TestOrchestratorCreateJobValidatesCapability(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	// Unknown capability should fail
	_, err := orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown capability")
	}
	if !strings.Contains(err.Error(), "capability") {
		t.Errorf("error = %v", err)
	}
}

func TestOrchestratorCreateJobRequiresProjectID(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	_, err := orch.CreateJob("", orchestrator.WorkflowPlanning, "planning")
	if err == nil {
		t.Fatal("expected error for empty project id")
	}
}

func TestOrchestratorCreateJobRequiresWorkflowType(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	_, err := orch.CreateJob("project:test", "", "planning")
	if err == nil {
		t.Fatal("expected error for empty workflow type")
	}
}

func TestOrchestratorExecuteCompleteFail(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowResearch, "research")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// Execute
	executed, err := orch.ExecuteJob(job.ID)
	if err != nil {
		t.Fatalf("ExecuteJob: %v", err)
	}
	if executed.Status != orchestrator.JobStatusRunning {
		t.Errorf("status = %q, want running", executed.Status)
	}

	// Complete
	completed, err := orch.CompleteJob(job.ID)
	if err != nil {
		t.Fatalf("CompleteJob: %v", err)
	}
	if completed.Status != orchestrator.JobStatusCompleted {
		t.Errorf("status = %q, want completed", completed.Status)
	}
	if completed.FinishedAt.IsZero() {
		t.Error("FinishedAt is zero")
	}

	// Fail
	job2, err := orch.CreateJob("project:test", orchestrator.WorkflowVision, "vision")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	failed, err := orch.FailJob(job2.ID, "something went wrong")
	if err != nil {
		t.Fatalf("FailJob: %v", err)
	}
	if failed.Status != orchestrator.JobStatusFailed {
		t.Errorf("status = %q, want failed", failed.Status)
	}
	if failed.Error != "something went wrong" {
		t.Errorf("error = %q", failed.Error)
	}
}

func TestOrchestratorDispatchWorkflow(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	job, err := orch.DispatchWorkflow("project:test", orchestrator.WorkflowPlanning)
	if err != nil {
		t.Fatalf("DispatchWorkflow: %v", err)
	}
	if job.Status != orchestrator.JobStatusRunning {
		t.Errorf("status = %q, want running", job.Status)
	}
	if job.Capability != "planning" {
		t.Errorf("capability = %q", job.Capability)
	}
}

func TestOrchestratorSelectCapability(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	cap, err := orch.SelectCapability(orchestrator.WorkflowPlanning)
	if err != nil {
		t.Fatalf("SelectCapability: %v", err)
	}
	if cap.Type != capabilities.CapPlanning {
		t.Errorf("type = %q", cap.Type)
	}
}

func TestOrchestratorSelectModelStrategy(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	strat, err := orch.SelectModelStrategy(modelstrategy.TaskExtraction)
	if err != nil {
		t.Fatalf("SelectModelStrategy: %v", err)
	}
	if strat.Tier != modelstrategy.TierSmall {
		t.Errorf("tier = %q", strat.Tier)
	}
}

func TestOrchestratorPersistResults(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	job, err := orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "planning")
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	run, err := orch.PersistResults(job.ID, "step1", "output data")
	if err != nil {
		t.Fatalf("PersistResults: %v", err)
	}
	if run.JobID != job.ID {
		t.Errorf("job_id = %q", run.JobID)
	}
	if run.Output != "output data" {
		t.Errorf("output = %q", run.Output)
	}
}

func TestOrchestratorLoadContext(t *testing.T) {
	db := openTestDB(t)
	orch := newTestOrchestrator(db)

	result := orch.LoadContext("project:test", "task1")
	if result != "context loading delegated to context engine" {
		t.Errorf("result = %q", result)
	}
}

func TestOrchestratorListJobs(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRepository(db)

	// No jobs yet
	jobs, err := jr.ListJobs("project:test")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("jobs = %d, want 0", len(jobs))
	}

	// Create some jobs via orchestrator
	orch := newTestOrchestrator(db)
	orch.CreateJob("project:test", orchestrator.WorkflowPlanning, "planning")
	orch.CreateJob("project:test", orchestrator.WorkflowResearch, "research")

	jobs, err = jr.ListJobs("project:test")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("jobs = %d, want 2", len(jobs))
	}
}

func TestOrchestratorJobRuns(t *testing.T) {
	db := openTestDB(t)
	rr := store.NewJobRunRepository(db)

	now := time.Now().UTC()
	run, err := rr.CreateJobRun(orchestrator.JobRun{
		ID:        "run:1",
		JobID:     "job:1",
		Step:      "analyze",
		Status:    orchestrator.JobStatusRunning,
		StartedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateJobRun: %v", err)
	}
	if run.ID != "run:1" {
		t.Errorf("id = %q", run.ID)
	}

	run.Status = orchestrator.JobStatusCompleted
	run.Output = "done"
	if _, err := rr.UpdateJobRun(run); err != nil {
		t.Fatalf("UpdateJobRun: %v", err)
	}

	runs, err := rr.ListJobRuns("job:1")
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) != 1 || runs[0].Output != "done" {
		t.Errorf("runs = %#v", runs)
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func newTestOrchestrator(db *sql.DB) *orchestrator.Orchestrator {
	capReg := capabilities.NewDefaultRegistry()
	ms := modelstrategy.NewService()
	jr := store.NewJobRepository(db)
	rr := store.NewJobRunRepository(db)
	return orchestrator.NewOrchestrator(capReg, ms, jr, rr)
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

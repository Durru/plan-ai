package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/modelstrategy"
	"github.com/Durru/plan-ai/internal/orchestrator"
)

func TestModelProfileRepositoryCreateAndGet(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewModelProfileRepository(db)

	profile, err := r.Create(modelstrategy.ModelProfile{
		ID:       "profile:1",
		Name:     "test-profile",
		Provider: "anthropic",
		Model:    "claude-sonnet-4-20250514",
		Config:   `{"max_tokens": 8192}`,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if profile.ID != "profile:1" {
		t.Errorf("id = %q", profile.ID)
	}

	got, err := r.Get("profile:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "test-profile" {
		t.Errorf("name = %q", got.Name)
	}
	if got.Provider != "anthropic" {
		t.Errorf("provider = %q", got.Provider)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestModelProfileRepositoryListAndDelete(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewModelProfileRepository(db)

	r.Create(modelstrategy.ModelProfile{ID: "profile:a", Name: "A", Provider: "openai", Model: "gpt-4", Config: "{}"})
	r.Create(modelstrategy.ModelProfile{ID: "profile:b", Name: "B", Provider: "anthropic", Model: "claude-3", Config: "{}"})

	profiles, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len = %d, want 2", len(profiles))
	}

	if err := r.Delete("profile:a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	profiles, _ = r.List()
	if len(profiles) != 1 {
		t.Errorf("after delete, len = %d", len(profiles))
	}
}

func TestPromptContractRepository(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewPromptContractRepository(db)

	pc, err := r.Create(modelstrategy.PromptContract{
		ID:           "pc:1",
		ContractType: "vision",
		Content:      `{"project_context": "test"}`,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pc.ID != "pc:1" {
		t.Errorf("id = %q", pc.ID)
	}

	contracts, err := r.ListByType("vision")
	if err != nil {
		t.Fatalf("ListByType: %v", err)
	}
	if len(contracts) != 1 {
		t.Fatalf("len = %d, want 1", len(contracts))
	}
	if contracts[0].Content != `{"project_context": "test"}` {
		t.Errorf("content = %q", contracts[0].Content)
	}

	// Empty type returns nothing
	empty, err := r.ListByType("research")
	if err != nil {
		t.Fatalf("ListByType research: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("len = %d, want 0", len(empty))
	}
}

func TestOutputSchemaRepository(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewOutputSchemaRepository(db)

	schema, err := r.Create(modelstrategy.OutputSchema{
		ID:         "schema:1",
		SchemaType: "vision",
		Fields:     `{"title": "string", "summary": "string"}`,
		Required:   `["title", "summary"]`,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if schema.ID != "schema:1" {
		t.Errorf("id = %q", schema.ID)
	}

	schemas, err := r.ListByType("vision")
	if err != nil {
		t.Fatalf("ListByType: %v", err)
	}
	if len(schemas) != 1 {
		t.Fatalf("len = %d, want 1", len(schemas))
	}
	if schemas[0].Fields != `{"title": "string", "summary": "string"}` {
		t.Errorf("fields = %q", schemas[0].Fields)
	}
}

func TestJobRepositoryCreateGetUpdateList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewJobRepository(db)

	now := time.Now().UTC()
	job, err := r.CreateJob(orchestrator.Job{
		ID:           "job:1",
		ProjectID:    "project:test",
		WorkflowType: orchestrator.WorkflowPlanning,
		Capability:   "planning",
		Strategy:     "medium",
		Status:       orchestrator.JobStatusPending,
		StartedAt:    now,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID != "job:1" {
		t.Errorf("id = %q", job.ID)
	}

	// GetJob
	got, err := r.GetJob("job:1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Status != orchestrator.JobStatusPending {
		t.Errorf("status = %q", got.Status)
	}

	// UpdateJob
	got.Status = orchestrator.JobStatusCompleted
	got.FinishedAt = time.Now().UTC()
	updated, err := r.UpdateJob(got)
	if err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	if updated.Status != orchestrator.JobStatusCompleted {
		t.Errorf("status = %q", updated.Status)
	}

	// ListJobs
	jobs, err := r.ListJobs("project:test")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("len = %d, want 1", len(jobs))
	}
}

func TestJobRepositoryListAllJobs(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewJobRepository(db)

	now := time.Now().UTC()
	r.CreateJob(orchestrator.Job{ID: "job:a", ProjectID: "project:A", WorkflowType: orchestrator.WorkflowPlanning, Capability: "planning", Status: orchestrator.JobStatusPending, StartedAt: now})
	r.CreateJob(orchestrator.Job{ID: "job:b", ProjectID: "project:B", WorkflowType: orchestrator.WorkflowResearch, Capability: "research", Status: orchestrator.JobStatusPending, StartedAt: now})

	// Empty projectID returns all jobs
	jobs, err := r.ListJobs("")
	if err != nil {
		t.Fatalf("ListJobs(''): %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("len = %d, want 2", len(jobs))
	}
}

func TestJobRunRepository(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewJobRunRepository(db)

	now := time.Now().UTC()
	run, err := r.CreateJobRun(orchestrator.JobRun{
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
	run.Output = "analysis complete"
	updated, err := r.UpdateJobRun(run)
	if err != nil {
		t.Fatalf("UpdateJobRun: %v", err)
	}
	if updated.Status != orchestrator.JobStatusCompleted {
		t.Errorf("status = %q", updated.Status)
	}

	runs, err := r.ListJobRuns("job:1")
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("len = %d, want 1", len(runs))
	}
	if runs[0].Output != "analysis complete" {
		t.Errorf("output = %q", runs[0].Output)
	}
}

func TestCapabilityRepository(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewCapabilityRepository(db)

	if err := r.Upsert("vision", "Vision Drafting", "Create vision drafts"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := r.Upsert("research", "Research", "Perform research"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	types, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("len = %d, want 2", len(types))
	}

	// Upsert should update
	if err := r.Upsert("vision", "Vision Updated", "Updated description"); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	types, _ = r.List()
	if len(types) != 2 {
		t.Errorf("after upsert, len = %d, want 2", len(types))
	}
}

func TestContextViewRepository(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewContextViewRepository(db)

	view, err := r.Create(orchestratorContextView{
		ID:        "cv:1",
		ProjectID: "project:test",
		Name:      "executive-summary",
		ViewType:  "executive",
		Content:   "Project is active",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if view.ID != "cv:1" {
		t.Errorf("id = %q", view.ID)
	}

	views, err := r.ListByProject("project:test")
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len = %d, want 1", len(views))
	}
	if views[0].Name != "executive-summary" {
		t.Errorf("name = %q", views[0].Name)
	}
}

// ──────────────────────────────────────────────
// Helper
// ──────────────────────────────────────────────

func openStoreTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

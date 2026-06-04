package context_test

import (
	"database/sql"
	"testing"

	approvedcontext "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/store"
)

func TestBuilderBuildExecutiveContext(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)
	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)

	ctx, err := builder.BuildExecutiveContext("project:test")
	if err != nil {
		t.Fatalf("BuildExecutiveContext: %v", err)
	}
	if ctx.ProjectID != "project:test" {
		t.Errorf("project_id = %q", ctx.ProjectID)
	}
	if ctx.Status == "" {
		t.Error("status is empty")
	}
	if len(ctx.WhatMissing) == 0 {
		t.Error("WhatMissing is empty")
	}
	if len(ctx.WhatNext) == 0 {
		t.Error("WhatNext is empty")
	}
}

func TestBuilderBuildExecutiveContextWithDecisions(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)

	// Seed a project record
	mustExec(t, db, `INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"project:test", "test", "/tmp/test", "active", nowStr(), nowStr())

	// Seed a master plan
	mustExec(t, db, `INSERT INTO master_plans (id, project_id, title, summary, status, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"plan:1", "project:test", "Test Plan", "A plan", "active", 1, nowStr(), nowStr())

	// Seed a phase
	mustExec(t, db, `INSERT INTO phases (id, plan_id, project_id, title, summary, status, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"phase:1", "plan:1", "project:test", "Phase 1", "First phase", "active", 1, nowStr(), nowStr())

	// Seed a task
	mustExec(t, db, `INSERT INTO tasks (id, phase_id, plan_id, project_id, title, summary, status, position, context_size, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"task:1", "phase:1", "plan:1", "project:test", "Task 1", "Do stuff", "completed", 1, "short", nowStr(), nowStr())

	// Seed a draft decision
	mustExec(t, db, `INSERT INTO decisions (id, project_id, title, context, decision, status, impact, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"dec:1", "project:test", "Draft Decision", "context", "decision", "draft", "medium", nowStr(), nowStr())

	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)
	ctx, err := builder.BuildExecutiveContext("project:test")
	if err != nil {
		t.Fatalf("BuildExecutiveContext: %v", err)
	}

	if ctx.Status != "active" {
		t.Errorf("status = %q, want active", ctx.Status)
	}
	if len(ctx.Progress) < 1 {
		t.Error("Progress should have at least 1 entry")
	}
	// The draft decision should show up in WhatMissing
	foundMissing := false
	for _, m := range ctx.WhatMissing {
		if len(m) > 0 {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Error("expected at least one WhatMissing entry")
	}
}

func TestBuilderBuildPlanningContext(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)

	// Seed a project (needed for knowledge/research queries)
	mustExec(t, db, `INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"project:test", "test", "/tmp/test", "active", nowStr(), nowStr())

	// Add some approved context
	reg := approvedcontext.NewRegistry(repo)
	reg.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeRequirement, SourceID: "src:1", Content: "Must save drafts"})
	reg.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeConstraint, SourceID: "src:1", Content: "No network"})
	reg.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeDecision, SourceID: "src:1", Content: "Use SQLite"})

	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)
	ctx, err := builder.BuildPlanningContext("project:test")
	if err != nil {
		t.Fatalf("BuildPlanningContext: %v", err)
	}
	if len(ctx.Requirements) != 1 || ctx.Requirements[0] != "Must save drafts" {
		t.Errorf("requirements = %v", ctx.Requirements)
	}
	if len(ctx.Constraints) != 1 || ctx.Constraints[0] != "No network" {
		t.Errorf("constraints = %v", ctx.Constraints)
	}
	if len(ctx.Decisions) != 1 || ctx.Decisions[0] != "Use SQLite" {
		t.Errorf("decisions = %v", ctx.Decisions)
	}
}

func TestBuilderBuildPlanningContextWithVision(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)

	mustExec(t, db, `INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"project:test", "test", "/tmp/test", "active", nowStr(), nowStr())
	mustExec(t, db, `INSERT INTO visions (id, project_id, title, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"vis:1", "project:test", "Test Vision", "Build a better system", nowStr(), nowStr())

	vq := &mockVisionQuerier{db: db}
	builder := approvedcontext.NewBuilder(repo, dq, vq, nil)

	ctx, err := builder.BuildPlanningContext("project:test")
	if err != nil {
		t.Fatalf("BuildPlanningContext: %v", err)
	}
	if ctx.Vision != "Build a better system" {
		t.Errorf("vision = %q", ctx.Vision)
	}
}

func TestBuilderBuildImplementationContext(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)

	reg := approvedcontext.NewRegistry(repo)
	reg.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeConstraint, SourceID: "src:1", Content: "Use SQLite"})
	reg.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeDecision, SourceID: "src:1", Content: "No external deps"})

	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)
	ctx, err := builder.BuildImplementationContext("project:test", "task:1")
	if err != nil {
		t.Fatalf("BuildImplementationContext: %v", err)
	}
	if ctx.ProjectID != "project:test" {
		t.Errorf("project_id = %q", ctx.ProjectID)
	}
	if len(ctx.Constraints) == 0 {
		t.Error("Constraints is empty")
	}
	if len(ctx.Validations) == 0 {
		t.Error("Validations is empty")
	}
}

func TestBuilderBuildResearchContext(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)

	mustExec(t, db, `INSERT INTO projects (id, name, root_path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"project:test", "test", "/tmp/test", "active", nowStr(), nowStr())
	mustExec(t, db, `INSERT INTO research_entries (id, project_id, topic, source, summary, conclusion, confidence, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"res:1", "project:test", "SQLite Performance", "benchmark", "How fast is SQLite?", "Very fast.", 85, nowStr(), nowStr())

	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)
	ctx, err := builder.BuildResearchContext("project:test", "SQLite")
	if err != nil {
		t.Fatalf("BuildResearchContext: %v", err)
	}
	if ctx.Topic != "SQLite" {
		t.Errorf("topic = %q", ctx.Topic)
	}
	if len(ctx.PreviousFindings) == 0 {
		t.Error("PreviousFindings is empty")
	}
}

func TestBuilderPersistContextView(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewApprovedContextRepository(db)
	dq := store.NewDomainQuerier(db)
	builder := approvedcontext.NewBuilder(repo, dq, nil, nil)

	view, err := builder.PersistContextView("test-view", "executive", "project:test", "content here")
	if err != nil {
		t.Fatalf("PersistContextView: %v", err)
	}
	if view.Name != "test-view" {
		t.Errorf("name = %q", view.Name)
	}
	if view.ViewType != "executive" {
		t.Errorf("view_type = %q", view.ViewType)
	}
	if view.Content != "content here" {
		t.Errorf("content = %q", view.Content)
	}
	if view.ID == "" {
		t.Error("ID is empty")
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

type mockVisionQuerier struct {
	db *sql.DB
}

func (m *mockVisionQuerier) ListVisions(projectID string) ([]domain.Vision, error) {
	rows, err := m.db.Query(`SELECT id, project_id, title, summary, created_at, updated_at FROM visions WHERE project_id = ? ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var visions []domain.Vision
	for rows.Next() {
		var v domain.Vision
		var createdAt, updatedAt string
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Title, &v.Summary, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		visions = append(visions, v)
	}
	return visions, rows.Err()
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	_, err := db.Exec(query, args...)
	if err != nil {
		t.Fatalf("exec: %v\nquery: %s", err, query)
	}
}

func nowStr() string {
	return "2025-01-01T00:00:00Z"
}

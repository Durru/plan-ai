package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/plan-ai/plan-ai/internal/domain"
	researchpkg "github.com/plan-ai/plan-ai/internal/research"
)

func TestDomainRepositoriesCreateGetListAndUpdateStatus(t *testing.T) {
	db := openMigratedProjectDB(t)

	plans := NewPlanRepository(db)
	phases := NewPhaseRepository(db)
	tasks := NewTaskRepository(db)
	decisions := NewDecisionRepository(db)
	researchRepo := NewResearchRepository(db)
	knowledge := NewKnowledgeRepository(db)
	validations := NewValidationRepository(db)
	snapshots := NewSnapshotRepository(db)

	master := domain.MasterPlan{ID: domain.NewID("plan"), Title: "Master", Summary: "Master summary", Status: domain.StatusDraft, Version: 1}
	if err := plans.CreateMaster(master); err != nil {
		t.Fatalf("create master plan: %v", err)
	}
	specific := domain.SpecificPlan{ID: domain.NewID("plan"), Title: "Specific", Summary: "Specific summary", Status: domain.StatusDraft, Version: 1, ParentPlanID: master.ID}
	if err := plans.CreateSpecific(specific); err != nil {
		t.Fatalf("create specific plan: %v", err)
	}
	phase := domain.Phase{ID: domain.NewID("phase"), PlanID: specific.ID, Title: "Phase", Summary: "Phase summary", Status: domain.StatusDraft, Position: 1}
	if err := phases.Create(phase); err != nil {
		t.Fatalf("create phase: %v", err)
	}
	task := domain.Task{ID: domain.NewID("task"), PhaseID: phase.ID, PlanID: specific.ID, Title: "Task", Summary: "Task summary", Status: domain.StatusDraft, Position: 1, ContextSize: domain.ContextSizeShort}
	if err := tasks.Create(task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	decision := domain.Decision{ID: domain.NewID("decision"), Title: "Decision", Context: "Context", Decision: "Chosen", Status: domain.StatusDraft, Impact: "Impact"}
	if err := decisions.Create(decision); err != nil {
		t.Fatalf("create decision: %v", err)
	}
	entry := researchpkg.ResearchEntry{ID: domain.NewID("research"), Topic: "Topic", Summary: "Summary", Confidence: 80}
	if err := researchRepo.CreateEntry(entry); err != nil {
		t.Fatalf("create research entry: %v", err)
	}
	object := domain.KnowledgeObject{ID: domain.NewID("knowledge"), Topic: "Topic", Summary: "Summary", Content: "Content", Confidence: 0.9, ReuseCount: 2}
	if err := knowledge.Create(object); err != nil {
		t.Fatalf("create knowledge object: %v", err)
	}
	validation := domain.Validation{ID: domain.NewID("validation"), TargetType: domain.ValidationTargetTask, TargetID: task.ID, Status: domain.StatusDraft, Summary: "Validation summary"}
	if err := validations.Create(validation); err != nil {
		t.Fatalf("create validation: %v", err)
	}
	snapshot := domain.Snapshot{ID: domain.NewID("snapshot"), Reason: "checkpoint", Summary: "Snapshot summary"}
	if err := snapshots.Create(snapshot); err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	statusCases := []struct {
		name   string
		update func() error
	}{
		{name: "plan", update: func() error { return plans.UpdateStatus(master.ID, domain.StatusApproved) }},
		{name: "phase", update: func() error { return phases.UpdateStatus(phase.ID, domain.StatusApproved) }},
		{name: "task", update: func() error { return tasks.UpdateStatus(task.ID, domain.StatusImplemented) }},
		{name: "decision", update: func() error { return decisions.UpdateStatus(decision.ID, domain.StatusApproved) }},
		{name: "validation", update: func() error { return validations.UpdateStatus(validation.ID, domain.StatusValidated) }},
	}
	for _, tt := range statusCases {
		t.Run("update status "+tt.name, func(t *testing.T) {
			if err := tt.update(); err != nil {
				t.Fatalf("update status: %v", err)
			}
		})
	}

	listCases := []struct {
		name string
		list func() (int, error)
		want int
	}{
		{name: "plans", list: func() (int, error) { got, err := plans.List(); return len(got), err }, want: 2},
		{name: "phases", list: func() (int, error) { got, err := phases.List(); return len(got), err }, want: 1},
		{name: "tasks", list: func() (int, error) { got, err := tasks.List(); return len(got), err }, want: 1},
		{name: "decisions", list: func() (int, error) { got, err := decisions.List(); return len(got), err }, want: 1},
		{name: "research", list: func() (int, error) { got, err := researchRepo.ListEntries(); return len(got), err }, want: 1},
		{name: "knowledge", list: func() (int, error) { got, err := knowledge.List(); return len(got), err }, want: 1},
		{name: "validations", list: func() (int, error) { got, err := validations.List(); return len(got), err }, want: 1},
		{name: "snapshots", list: func() (int, error) { got, err := snapshots.List(); return len(got), err }, want: 1},
	}
	for _, tt := range listCases {
		t.Run("list "+tt.name, func(t *testing.T) {
			got, err := tt.list()
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if got != tt.want {
				t.Fatalf("count = %d, want %d", got, tt.want)
			}
		})
	}

	if got, err := plans.GetByID(master.ID); err != nil || got.Status != domain.StatusApproved {
		t.Fatalf("get master after update = %+v, err=%v", got, err)
	}
	if got, err := tasks.GetByID(task.ID); err != nil || got.Status != domain.StatusImplemented {
		t.Fatalf("get task after update = %+v, err=%v", got, err)
	}
}

func TestCountDomainEntitiesReturnsZeroesForEmptyProject(t *testing.T) {
	db := openMigratedProjectDB(t)

	counts, err := CountDomainEntities(db)
	if err != nil {
		t.Fatalf("count domain entities: %v", err)
	}

	for name, got := range map[string]int{
		"plans":             counts.Plans,
		"phases":            counts.Phases,
		"tasks":             counts.Tasks,
		"decisions":         counts.Decisions,
		"research_entries":  counts.ResearchEntries,
		"knowledge_objects": counts.KnowledgeObjects,
		"validations":       counts.Validations,
		"snapshots":         counts.Snapshots,
	} {
		if got != 0 {
			t.Fatalf("%s count = %d, want 0", name, got)
		}
	}
}

func openMigratedProjectDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "project.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("run project migrations: %v", err)
	}
	return db
}

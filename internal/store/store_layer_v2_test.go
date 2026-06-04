package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/plan-ai/plan-ai/internal/domain"
	domainrepos "github.com/plan-ai/plan-ai/internal/store/repositories"
)

func TestStoreLayerV2CreatesSandboxStores(t *testing.T) {
	home := t.TempDir()
	global, err := OpenGlobalStore(home)
	if err != nil {
		t.Fatalf("open global store: %v", err)
	}
	defer global.Close()
	assertFileExists(t, filepath.Join(home, ".plan-ai", "global.db"))

	projectRoot := t.TempDir()
	project, err := OpenProjectStore(projectRoot)
	if err != nil {
		t.Fatalf("open project store: %v", err)
	}
	defer project.Close()
	assertFileExists(t, filepath.Join(projectRoot, ".plan-ai", "project.db"))
}

func TestStoreLayerV2RepositoriesCreateGetListUpdate(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	projectID := "project_1"
	repos := domainrepos.NewProjectRepository(db)
	if err := repos.Save(domain.Project{ID: projectID, Name: "Sandbox", RootPath: "/tmp/sandbox", Status: domain.ProjectStatusDraft}); err != nil {
		t.Fatalf("save project: %v", err)
	}
	if err := repos.UpdateStatus(projectID, domain.ProjectStatusActive); err != nil {
		t.Fatalf("update project: %v", err)
	}
	project, err := repos.GetByID(projectID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if project.Status != domain.ProjectStatusActive {
		t.Fatalf("project status = %s", project.Status)
	}
	projects, err := repos.List()
	if err != nil || len(projects) != 1 {
		t.Fatalf("list projects len=%d err=%v", len(projects), err)
	}

	visionRepo := domainrepos.NewVisionRepository(db)
	if err := visionRepo.Save(domain.Vision{ID: "vision_1", ProjectID: projectID, Title: "Vision", Summary: "Summary", ExpectedOutcome: "Outcome"}); err != nil {
		t.Fatal(err)
	}
	if err := visionRepo.Approve("vision_1"); err != nil {
		t.Fatal(err)
	}
	visions, err := visionRepo.ListByProject(projectID)
	if err != nil || len(visions) != 1 || !visions[0].Approved {
		t.Fatalf("visions=%v err=%v", visions, err)
	}

	reqRepo := domainrepos.NewRequirementRepository(db)
	if err := reqRepo.Save(domain.Requirement{ID: "req_1", ProjectID: projectID, Type: domain.RequirementTypeFunctional, Statement: "Do the thing"}); err != nil {
		t.Fatal(err)
	}
	if err := reqRepo.Approve("req_1"); err != nil {
		t.Fatal(err)
	}
	reqs, err := reqRepo.ListByType(projectID, domain.RequirementTypeFunctional)
	if err != nil || len(reqs) != 1 || !reqs[0].Approved {
		t.Fatalf("requirements=%v err=%v", reqs, err)
	}

	constraintRepo := domainrepos.NewConstraintRepository(db)
	if err := constraintRepo.Save(domain.Constraint{ID: "constraint_1", ProjectID: projectID, Type: domain.ConstraintStack, Description: "Use Go"}); err != nil {
		t.Fatal(err)
	}
	constraints, err := constraintRepo.ListByProject(projectID)
	if err != nil || len(constraints) != 1 {
		t.Fatalf("constraints=%v err=%v", constraints, err)
	}

	decisionRepo := domainrepos.NewDecisionRepository(db)
	if err := decisionRepo.Save(domain.Decision{ID: "decision_1", ProjectID: projectID, Title: "ADR", Context: "ctx", Decision: "Use SQLite", Status: domain.DecisionProposed, Impact: "local"}); err != nil {
		t.Fatal(err)
	}
	if err := decisionRepo.UpdateStatus("decision_1", domain.StatusApproved); err != nil {
		t.Fatal(err)
	}
	decisions, err := decisionRepo.ListByProject(projectID)
	if err != nil || len(decisions) != 1 || decisions[0].Status != domain.StatusApproved {
		t.Fatalf("decisions=%v err=%v", decisions, err)
	}

	researchRepo := domainrepos.NewResearchRepository(db)
	if err := researchRepo.Save(domain.Research{ID: "research_1", ProjectID: projectID, Topic: "SQLite FTS", Objective: "search", Summary: "FTS works", Confidence: 0.8}); err != nil {
		t.Fatal(err)
	}
	researchResults, err := researchRepo.Search("FTS")
	if err != nil || len(researchResults) != 1 {
		t.Fatalf("research=%v err=%v", researchResults, err)
	}
	registryResearch, err := NewResearchRepository(db).ListResearchJobs(projectID)
	if err != nil || len(registryResearch) != 1 || registryResearch[0].ID != "research_1" {
		t.Fatalf("registry research mirror=%v err=%v", registryResearch, err)
	}

	knowledgeRepo := domainrepos.NewKnowledgeRepository(db)
	if err := knowledgeRepo.Save(domain.KnowledgeObject{ID: "knowledge_1", Topic: "Store", Summary: "SQLite", Content: "Definitive store", Confidence: 0.9}); err != nil {
		t.Fatal(err)
	}
	knowledge, err := knowledgeRepo.IncrementReuseCount("knowledge_1")
	if err != nil || knowledge.ReuseCount != 1 {
		t.Fatalf("knowledge=%v err=%v", knowledge, err)
	}

	planRepo := domainrepos.NewPlanRepository(db)
	if err := planRepo.SaveMaster(domain.MasterPlan{ID: "master_1", ProjectID: projectID, Title: "Master", Summary: "Top", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := planRepo.SaveSpecific(domain.SpecificPlan{ID: "specific_1", ProjectID: projectID, MasterPlanID: "master_1", Title: "Specific", Summary: "Do", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	specifics, err := planRepo.ListSpecificsByMaster("master_1")
	if err != nil || len(specifics) != 1 {
		t.Fatalf("specifics=%v err=%v", specifics, err)
	}

	phase := domain.Phase{ID: "phase_1", PlanID: "specific_1", Title: "Phase", Summary: "Build", Status: domain.PlanStatusPending, Position: 1}
	if err := NewPhaseRepository(db).Create(phase); err != nil {
		t.Fatalf("phase create: %v", err)
	}
	taskRepo := domainrepos.NewTaskRepository(db)
	if err := taskRepo.Save(domain.Task{ID: "task_1", PhaseID: "phase_1", PlanID: "specific_1", Title: "Task", Summary: "Step", Status: domain.PlanStatusPending}); err != nil {
		t.Fatal(err)
	}
	if err := taskRepo.UpdateStatus("task_1", domain.PlanStatusActive); err != nil {
		t.Fatal(err)
	}
	tasks, err := taskRepo.ListByPhase("phase_1")
	if err != nil || len(tasks) != 1 || tasks[0].Status != domain.PlanStatusActive {
		t.Fatalf("tasks=%v err=%v", tasks, err)
	}

	validationRepo := domainrepos.NewValidationRepository(db)
	if err := validationRepo.Save(domain.Validation{ID: "validation_1", TargetType: domain.ValidationTargetTask, TargetID: "task_1", Status: domain.StatusDraft, Summary: "ok"}); err != nil {
		t.Fatal(err)
	}
	if _, err := validationRepo.GetByID("validation_1"); err != nil {
		t.Fatal(err)
	}

	snapshotRepo := domainrepos.NewSnapshotRepository(db)
	if err := snapshotRepo.Save(domain.Snapshot{ID: "snapshot_1", ProjectID: projectID, Reason: "test", Summary: "state"}); err != nil {
		t.Fatal(err)
	}
	snapshots, err := snapshotRepo.ListByProject(projectID)
	if err != nil || len(snapshots) != 1 {
		t.Fatalf("snapshots=%v err=%v", snapshots, err)
	}

	changeRepo := domainrepos.NewChangeRepository(db)
	if err := changeRepo.SaveChangeRequest(domain.ChangeRequest{ID: "change_1", ProjectID: projectID, Reason: "scope", Description: "change", Status: domain.ChangeRequestSubmitted, Requester: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := changeRepo.UpdateChangeRequestStatus("change_1", domain.ChangeRequestApproved); err != nil {
		t.Fatal(err)
	}
	if err := changeRepo.SaveImpactReport(domain.ImpactReport{ID: "impact_1", ChangeRequestID: "change_1", AffectedPlans: []string{"master_1"}, Summary: "affects plan"}); err != nil {
		t.Fatal(err)
	}
	impact, err := changeRepo.GetImpactReportByChangeRequest("change_1")
	if err != nil || len(impact.AffectedPlans) != 1 {
		t.Fatalf("impact=%v err=%v", impact, err)
	}
}

func TestWithTransactionCommitAndRollback(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "tx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE items (id TEXT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if err := WithTransaction(db, func(tx *sql.Tx) error { _, err := tx.Exec(`INSERT INTO items (id) VALUES ('commit')`); return err }); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
	if err := WithTransaction(db, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO items (id) VALUES ('rollback')`)
		if err != nil {
			return err
		}
		return errors.New("rollback")
	}); err == nil {
		t.Fatal("expected rollback error")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d want 1", count)
	}
}

func TestStoreLayerV2FTSBasic(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "fts.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO knowledge_objects_fts (id, topic, summary, content) VALUES ('k1', 'SQLite Search', 'fts', 'token search')`); err != nil {
		t.Skipf("FTS5 unavailable: %v", err)
	}
	var id string
	if err := db.QueryRow(`SELECT id FROM knowledge_objects_fts WHERE knowledge_objects_fts MATCH 'token'`).Scan(&id); err != nil {
		t.Fatalf("fts query: %v", err)
	}
	if id != "k1" {
		t.Fatalf("id=%s", id)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing: %s: %v", path, err)
	}
}

package store

import (
	"testing"

	"github.com/plan-ai/plan-ai/internal/knowledge"
	"github.com/plan-ai/plan-ai/internal/planning"
	"github.com/plan-ai/plan-ai/internal/research"
	"github.com/plan-ai/plan-ai/internal/workflows"
)

func TestPhase12ResearchKnowledgePersistence(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	researchRegistry := research.NewRegistry(NewResearchRepository(db))
	job, err := researchRegistry.CreateResearch(research.CreateResearchRequest{ProjectID: "project:test", Topic: "SQLite", Summary: "Research SQLite", Findings: []research.ResearchFinding{{Title: "WAL", Content: "Use WAL"}}, Recommendations: []research.ResearchRecommendation{{Content: "Use busy timeout"}}, Sources: []research.ResearchSource{{Title: "Docs", URL: "https://sqlite.org"}}, Confidence: 0.7})
	if err != nil {
		t.Fatalf("create research: %v", err)
	}
	gotJob, err := researchRegistry.GetResearch(job.ID)
	if err != nil {
		t.Fatalf("get research: %v", err)
	}
	if len(gotJob.Findings) != 1 || len(gotJob.Recommendations) != 1 || len(gotJob.Sources) != 1 {
		t.Fatalf("unexpected research aggregate: %+v", gotJob)
	}

	knowledgeRegistry := knowledge.NewRegistry(NewKnowledgeRepository(db))
	object, err := knowledgeRegistry.CreateKnowledge(knowledge.CreateKnowledgeRequest{ProjectID: "project:test", Title: "SQLite WAL", Category: knowledge.CategoryDatabase, Summary: "Use WAL", ResearchIDs: []string{job.ID}, Confidence: 0.9})
	if err != nil {
		t.Fatalf("create knowledge: %v", err)
	}
	matches, err := knowledgeRegistry.SearchKnowledge("wal")
	if err != nil {
		t.Fatalf("search knowledge: %v", err)
	}
	if len(matches) != 1 || matches[0].ID != object.ID || len(matches[0].ResearchIDs) != 1 {
		t.Fatalf("unexpected knowledge matches: %+v", matches)
	}
}

func TestPhase12ResearchEntryCreationMirrorsToRegistryJob(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	repo := NewResearchRepository(db)
	if err := repo.CreateEntry(research.ResearchEntry{ID: "research_service", ProjectID: "project:test", Topic: "Service path", Summary: "Created by service", Confidence: 85, Status: research.ResearchStatusApproved}); err != nil {
		t.Fatalf("create entry: %v", err)
	}

	registry := research.NewRegistry(repo)
	jobs, err := registry.ListResearch("project:test")
	if err != nil {
		t.Fatalf("list research jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected service entry mirrored as one registry job, got %d", len(jobs))
	}
	if jobs[0].ID != "research_service" || jobs[0].Topic != "Service path" || jobs[0].Summary != "Created by service" {
		t.Fatalf("unexpected mirrored job: %+v", jobs[0])
	}
	if jobs[0].Confidence != 0.85 || jobs[0].Status != research.ResearchStatusApproved {
		t.Fatalf("unexpected mirrored job status/confidence: %+v", jobs[0])
	}
}

func TestPhase12RegistryResearchCreationRemainsVisibleToServiceEntries(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	repo := NewResearchRepository(db)
	registry := research.NewRegistry(repo)
	job, err := registry.CreateResearch(research.CreateResearchRequest{ProjectID: "project:test", Topic: "Registry path", Summary: "Created by registry", Confidence: 0.7})
	if err != nil {
		t.Fatalf("create registry research: %v", err)
	}

	entries, err := repo.ListEntries()
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected registry job mirrored as one service entry, got %d", len(entries))
	}
	if entries[0].ID != job.ID || entries[0].Topic != "Registry path" || entries[0].Summary != "Created by registry" {
		t.Fatalf("unexpected mirrored entry: %+v", entries[0])
	}
	if entries[0].Confidence != 70 || entries[0].Status != research.ResearchStatusDraft {
		t.Fatalf("unexpected mirrored entry status/confidence: %+v", entries[0])
	}
}

func TestPhase12KnowledgeResearchLinksMirrorAcrossSurfaces(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	researchRepo := NewResearchRepository(db)
	if err := researchRepo.CreateEntry(research.ResearchEntry{ID: "research_linked", ProjectID: "project:test", Topic: "Linked research"}); err != nil {
		t.Fatalf("create research entry: %v", err)
	}

	knowledgeRegistry := knowledge.NewRegistry(NewKnowledgeRepository(db))
	object, err := knowledgeRegistry.CreateKnowledge(knowledge.CreateKnowledgeRequest{ProjectID: "project:test", Title: "Linked knowledge", Summary: "References research", ResearchIDs: []string{"research_linked"}, Confidence: 0.9})
	if err != nil {
		t.Fatalf("create knowledge: %v", err)
	}

	legacyLinks, err := researchRepo.ListKnowledgeLinks("research_linked")
	if err != nil {
		t.Fatalf("list legacy research knowledge links: %v", err)
	}
	if len(legacyLinks) != 1 || legacyLinks[0].KnowledgeID != object.ID {
		t.Fatalf("expected registry knowledge research_ids mirrored to legacy links, got %+v", legacyLinks)
	}

	if err := researchRepo.LinkKnowledge("research_linked", object.ID); err != nil {
		t.Fatalf("link duplicate knowledge: %v", err)
	}
	var registryLinkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM knowledge_links WHERE knowledge_id = ? AND link_type = 'research' AND target_id = ?`, object.ID, "research_linked").Scan(&registryLinkCount); err != nil {
		t.Fatalf("count registry links: %v", err)
	}
	if registryLinkCount != 1 {
		t.Fatalf("expected one registry knowledge link after duplicate legacy link, got %d", registryLinkCount)
	}
}

func TestPhase13PlanningPersistence(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	svc := planning.NewService(NewPlanningRepository(db))
	master, err := svc.CreateMasterPlan(planning.PlanningInput{ProjectID: "project:test", VisionReference: "vision-1", ApprovedRequirements: []string{"Persist plans"}, ApprovedConstraints: []string{"SQLite"}, ResearchIDs: []string{"research-1"}, KnowledgeIDs: []string{"knowledge-1"}})
	if err != nil {
		t.Fatalf("create master: %v", err)
	}
	specific, err := svc.CreateSpecificPlan(master.ID, planning.SpecificPlanInput{ProjectID: "project:test", Goal: "Build planner", Requirements: []string{"Persist plans"}, KnowledgeUsed: []string{"knowledge-1"}, ResearchUsed: []string{"research-1"}})
	if err != nil {
		t.Fatalf("create specific: %v", err)
	}
	doc, err := svc.CreateImplementationDocument(specific.ID, planning.ImplementationDocumentInput{ProjectID: "project:test", Objective: "Implement", Architecture: "Layered", Validations: []string{"go test ./..."}})
	if err != nil {
		t.Fatalf("create doc: %v", err)
	}
	if doc.SpecificPlanID != specific.ID || len(specific.KnowledgeUsed) != 1 || master.VisionReference != "vision-1" {
		t.Fatalf("unexpected planning records: %+v %+v %+v", master, specific, doc)
	}
}

func TestPhase14WorkflowRunPersistence(t *testing.T) {
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatal(err)
	}

	registry := workflows.NewRegistry(NewWorkflowRunRepository(db))
	run, err := registry.ExecuteWorkflow(workflows.WorkflowTypePlanning)
	if err != nil {
		t.Fatalf("execute workflow: %v", err)
	}
	got, err := NewWorkflowRunRepository(db).GetWorkflowRun(run.ID)
	if err != nil {
		t.Fatalf("get workflow run: %v", err)
	}
	if got.Status != workflows.StatusCompleted || got.FinishedAt.IsZero() {
		t.Fatalf("unexpected run: %+v", got)
	}
}

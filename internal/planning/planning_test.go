package planning

import "testing"

type memoryRepository struct {
	masters   []MasterPlan
	specifics []SpecificPlan
	docs      []ImplementationDocument
}

func (r *memoryRepository) CreateMasterPlan(plan MasterPlan) (MasterPlan, error) {
	r.masters = append(r.masters, plan)
	return plan, nil
}
func (r *memoryRepository) CreateSpecificPlan(plan SpecificPlan) (SpecificPlan, error) {
	r.specifics = append(r.specifics, plan)
	return plan, nil
}
func (r *memoryRepository) CreateImplementationDocument(doc ImplementationDocument) (ImplementationDocument, error) {
	r.docs = append(r.docs, doc)
	return doc, nil
}
func (r *memoryRepository) GetMasterPlan(id string) (MasterPlan, error) { return r.masters[0], nil }
func (r *memoryRepository) GetSpecificPlan(id string) (SpecificPlan, error) {
	return r.specifics[0], nil
}
func (r *memoryRepository) GetImplementationDocument(id string) (ImplementationDocument, error) {
	return r.docs[0], nil
}
func (r *memoryRepository) ListMasterPlans(projectID string) ([]MasterPlan, error) {
	return r.masters, nil
}

func TestServiceCreatesPlanningArtifactsFromApprovedInputs(t *testing.T) {
	repo := &memoryRepository{}
	svc := NewService(repo)

	master, err := svc.CreateMasterPlan(PlanningInput{
		ProjectID:            "project:test",
		VisionReference:      "vision-1",
		ApprovedRequirements: []string{"Persist research"},
		ApprovedConstraints:  []string{"Use SQLite"},
		ResearchIDs:          []string{"research-1"},
		KnowledgeIDs:         []string{"knowledge-1"},
	})
	if err != nil {
		t.Fatalf("CreateMasterPlan error: %v", err)
	}
	if master.Title == "" || master.VisionReference != "vision-1" || master.Status != StatusDraft {
		t.Fatalf("unexpected master plan: %+v", master)
	}

	specific, err := svc.CreateSpecificPlan(master.ID, SpecificPlanInput{ProjectID: "project:test", Title: "Research persistence", Goal: "Persist research", Requirements: []string{"Persist research"}, Constraints: []string{"Use SQLite"}, KnowledgeUsed: []string{"knowledge-1"}, ResearchUsed: []string{"research-1"}})
	if err != nil {
		t.Fatalf("CreateSpecificPlan error: %v", err)
	}
	if specific.MasterPlanID != master.ID || len(specific.KnowledgeUsed) != 1 || len(specific.ResearchUsed) != 1 {
		t.Fatalf("unexpected specific plan: %+v", specific)
	}

	doc, err := svc.CreateImplementationDocument(specific.ID, ImplementationDocumentInput{ProjectID: "project:test", Objective: "Build storage", Architecture: "SQLite repositories", ExpectedFiles: []string{"internal/store/foo.go"}, Validations: []string{"go test ./..."}, TestingStrategy: "TDD", RollbackStrategy: "revert migration"})
	if err != nil {
		t.Fatalf("CreateImplementationDocument error: %v", err)
	}
	if doc.SpecificPlanID != specific.ID || doc.Objective != "Build storage" || len(doc.Validations) != 1 {
		t.Fatalf("unexpected implementation document: %+v", doc)
	}
}

func TestServiceRejectsPlanningWithoutVisionReference(t *testing.T) {
	_, err := NewService(&memoryRepository{}).CreateMasterPlan(PlanningInput{ProjectID: "project:test", ApprovedRequirements: []string{"x"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

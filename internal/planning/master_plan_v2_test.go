package planning

import (
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// mockMPV2Repo implements MasterPlanV2Repository for testing.
type mockMPV2Repo struct {
	versions   []interface{}
	changes    []interface{}
	approvals  []interface{}
	evolutions []PlanEvolutionEvent
}

func newMockMPV2Repo() *mockMPV2Repo {
	return &mockMPV2Repo{}
}

func (m *mockMPV2Repo) CreateVersion(v interface{}) error {
	m.versions = append(m.versions, v)
	return nil
}
func (m *mockMPV2Repo) GetLatestVersion(planID string) (interface{}, error) {
	if len(m.versions) == 0 {
		return nil, nil
	}
	return m.versions[len(m.versions)-1], nil
}
func (m *mockMPV2Repo) ListVersions(planID string) ([]interface{}, error) {
	return m.versions, nil
}
func (m *mockMPV2Repo) CreateChange(c interface{}) error {
	m.changes = append(m.changes, c)
	return nil
}
func (m *mockMPV2Repo) ListChanges(planID string) ([]interface{}, error) {
	return m.changes, nil
}
func (m *mockMPV2Repo) CreateApproval(a interface{}) error {
	m.approvals = append(m.approvals, a)
	return nil
}
func (m *mockMPV2Repo) GetLatestApproval(planID string) (interface{}, error) {
	if len(m.approvals) == 0 {
		return nil, nil
	}
	return m.approvals[len(m.approvals)-1], nil
}
func (m *mockMPV2Repo) CreateEvolutionEvent(e PlanEvolutionEvent) (PlanEvolutionEvent, error) {
	m.evolutions = append(m.evolutions, e)
	return e, nil
}
func (m *mockMPV2Repo) ListEvolutionEvents(projectID string) ([]PlanEvolutionEvent, error) {
	return m.evolutions, nil
}

// mockStoreRepo implements Repository for backward compat.
type mockStoreRepo struct {
	plans         map[string]MasterPlan
	specificPlans map[string]SpecificPlan
}

func newMockStoreRepo() *mockStoreRepo {
	return &mockStoreRepo{
		plans:         make(map[string]MasterPlan),
		specificPlans: make(map[string]SpecificPlan),
	}
}
func (m *mockStoreRepo) CreateMasterPlan(p MasterPlan) (MasterPlan, error) {
	m.plans[p.ID] = p
	return p, nil
}
func (m *mockStoreRepo) GetMasterPlan(id string) (MasterPlan, error) {
	return m.plans[id], nil
}
func (m *mockStoreRepo) CreateSpecificPlan(p SpecificPlan) (SpecificPlan, error) {
	m.specificPlans[p.ID] = p
	return p, nil
}
func (m *mockStoreRepo) CreateImplementationDocument(d ImplementationDocument) (ImplementationDocument, error) {
	return d, nil
}
func (m *mockStoreRepo) GetSpecificPlan(id string) (SpecificPlan, error) {
	return m.specificPlans[id], nil
}
func (m *mockStoreRepo) GetImplementationDocument(id string) (ImplementationDocument, error) {
	return ImplementationDocument{}, nil
}
func (m *mockStoreRepo) ListMasterPlans(projectID string) ([]MasterPlan, error) { return nil, nil }

func TestGenerateV2(t *testing.T) {
	storeRepo := newMockStoreRepo()
	v2Repo := newMockMPV2Repo()
	gen := NewMasterPlanGenerator(storeRepo, v2Repo)

	input := PlanningInput{
		ProjectID:            "proj-1",
		VisionReference:      "vision-1",
		ApprovedRequirements: []string{"Build auth system"},
		ApprovedConstraints:  []string{"Must use OAuth2"},
		ApprovedDecisions:    []string{"Use JWT"},
	}

	phases := []PhaseDef{
		{Name: "Setup", Description: "Initialize project", Order: 1, EstimatedWeeks: 2},
		{Name: "Core", Description: "Build core features", Order: 2, EstimatedWeeks: 4},
	}
	timeline := Timeline{
		TotalWeeks: 6,
		Milestones: []Milestone{{Name: "MVP", Week: 4, Criteria: "All core features done"}},
	}
	risks := []RiskEntry{
		{Description: "Scope creep", Impact: "high", Likelihood: "medium", Mitigation: "Strict prioritization", Status: "active"},
	}

	plan, err := gen.GenerateV2(input, "Test master plan", phases, timeline, risks)
	if err != nil {
		t.Fatalf("GenerateV2 failed: %v", err)
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if plan.Version != 1 {
		t.Errorf("expected version 1, got %d", plan.Version)
	}
	if len(plan.Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(plan.Phases))
	}
	if plan.Timeline.TotalWeeks != 6 {
		t.Errorf("expected 6 weeks, got %d", plan.Timeline.TotalWeeks)
	}
	if len(plan.Risks) != 1 {
		t.Errorf("expected 1 risk, got %d", len(plan.Risks))
	}

	// Check evolution event was recorded
	if len(v2Repo.evolutions) != 1 {
		t.Errorf("expected 1 evolution event, got %d", len(v2Repo.evolutions))
	}
}

func TestGenerateV2_Validation(t *testing.T) {
	gen := NewMasterPlanGenerator(newMockStoreRepo(), newMockMPV2Repo())

	_, err := gen.GenerateV2(PlanningInput{ProjectID: "", VisionReference: ""}, "", nil, Timeline{}, nil)
	if err == nil {
		t.Fatal("expected error for empty project id")
	}

	_, err = gen.GenerateV2(PlanningInput{ProjectID: "proj-1", VisionReference: ""}, "", nil, Timeline{}, nil)
	if err == nil {
		t.Fatal("expected error for empty vision reference")
	}
}

func TestRenderMasterPlanSummary(t *testing.T) {
	plan := &MasterPlanV2{
		Title:       "Test Plan",
		Version:     2,
		Status:      StatusDraft,
		Description: "A test plan",
		Phases:      []PhaseDef{{Name: "Phase 1", Description: "First", Order: 1, EstimatedWeeks: 2}},
		Timeline:    Timeline{TotalWeeks: 2, Milestones: []Milestone{{Name: "Done", Week: 2, Criteria: "All done"}}},
		Risks:       []RiskEntry{{Description: "Risk 1", Impact: "high", Likelihood: "medium"}},
	}
	summary := RenderMasterPlanSummary(plan)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestCreateNewVersion(t *testing.T) {
	storeRepo := newMockStoreRepo()
	v2Repo := newMockMPV2Repo()
	gen := NewMasterPlanGenerator(storeRepo, v2Repo)

	now := time.Now()
	storeRepo.CreateMasterPlan(MasterPlan{ID: "mp-1", ProjectID: "proj-1", Title: "Test", Version: 1, CreatedAt: now, UpdatedAt: now})

	plan, err := gen.CreateNewVersion("mp-1", []string{"Updated scope", "Added new phase"})
	if err != nil {
		t.Fatalf("CreateNewVersion failed: %v", err)
	}
	if plan.Version != 2 {
		t.Errorf("expected version 2, got %d", plan.Version)
	}
	if plan.Changelog != "Updated scope; Added new phase" {
		t.Errorf("unexpected changelog: %s", plan.Changelog)
	}

	if len(v2Repo.evolutions) != 1 {
		t.Errorf("expected 1 evolution event, got %d", len(v2Repo.evolutions))
	}
}

func TestSubmitAndApprovePlan(t *testing.T) {
	v2Repo := newMockMPV2Repo()
	gen := NewMasterPlanGenerator(newMockStoreRepo(), v2Repo)

	if err := gen.SubmitForApproval("mp-1", "proj-1"); err != nil {
		t.Fatalf("SubmitForApproval failed: %v", err)
	}
	if err := gen.ApprovePlan("mp-1", "proj-1", "reviewer", "approved"); err != nil {
		t.Fatalf("ApprovePlan failed: %v", err)
	}

	if len(v2Repo.evolutions) != 2 {
		t.Errorf("expected 2 evolution events, got %d", len(v2Repo.evolutions))
	}
}

func TestGetEvolutionHistory(t *testing.T) {
	v2Repo := newMockMPV2Repo()
	gen := NewMasterPlanGenerator(newMockStoreRepo(), v2Repo)

	gen.SubmitForApproval("mp-1", "proj-1")
	gen.ApprovePlan("mp-1", "proj-1", "reviewer", "ok")

	events, err := gen.GetEvolutionHistory("proj-1")
	if err != nil {
		t.Fatalf("GetEvolutionHistory failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestDomainNewID(t *testing.T) {
	id := domain.NewID("test")
	if id == "" {
		t.Fatal("expected non-empty ID")
	}
}

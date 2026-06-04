package planning

import (
	"testing"
)

// mockSPV2Repo implements SpecificPlanV2Repository for testing.
type mockSPV2Repo struct {
	versions      []interface{}
	researchLinks []interface{}
	regenerations []RegenerationRecord
}

func newMockSPV2Repo() *mockSPV2Repo {
	return &mockSPV2Repo{}
}

func (m *mockSPV2Repo) CreateVersion(v interface{}) error {
	m.versions = append(m.versions, v)
	return nil
}
func (m *mockSPV2Repo) GetLatestVersion(planID string) (interface{}, error) {
	if len(m.versions) == 0 {
		return nil, nil
	}
	return m.versions[len(m.versions)-1], nil
}
func (m *mockSPV2Repo) ListVersions(planID string) ([]interface{}, error) {
	return m.versions, nil
}
func (m *mockSPV2Repo) CreateResearchLink(planID, researchID, section string, relevance float64) error {
	m.researchLinks = append(m.researchLinks, struct{}{})
	return nil
}
func (m *mockSPV2Repo) ListResearchLinks(planID string) ([]interface{}, error) {
	return m.researchLinks, nil
}
func (m *mockSPV2Repo) CreateRegeneration(r RegenerationRecord) (RegenerationRecord, error) {
	m.regenerations = append(m.regenerations, r)
	return r, nil
}
func (m *mockSPV2Repo) ListRegenerations(planID string) ([]RegenerationRecord, error) {
	return m.regenerations, nil
}

func TestGenerateV2_Specific(t *testing.T) {
	storeRepo := newMockStoreRepo()
	v2Repo := newMockSPV2Repo()
	gen := NewSpecificPlanGenerator(storeRepo, v2Repo)

	input := SpecificPlanInput{
		ProjectID:              "proj-1",
		Title:                  "Implement OAuth2",
		Goal:                   "Add OAuth2 authentication",
		ImplementationStrategy: "Use golang.org/x/oauth2",
		ValidationCriteria:     []string{"Pass all auth tests"},
	}

	tasks := []TaskDef{
		{Title: "Setup OAuth2 lib", Order: 1, Effort: "2d"},
		{Title: "Implement flow", Order: 2, Effort: "3d"},
	}
	deps := []DepDef{
		{Name: "go-oauth2", Type: "library", Description: "OAuth2 client", Critical: true},
	}
	risks := []RiskEntry{
		{Description: "Token refresh complexity", Impact: "medium", Likelihood: "low", Mitigation: "Use refresh tokens", Status: "identified"},
	}

	plan, err := gen.GenerateV2("mp-1", input, DomainBackend, tasks, deps, risks)
	if err != nil {
		t.Fatalf("GenerateV2 failed: %v", err)
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if plan.Domain != DomainBackend {
		t.Errorf("expected backend domain, got %s", plan.Domain)
	}
	if plan.Version != 1 {
		t.Errorf("expected version 1, got %d", plan.Version)
	}
	if len(plan.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(plan.Tasks))
	}
	if len(plan.Risks) != 1 {
		t.Errorf("expected 1 risk, got %d", len(plan.Risks))
	}
}

func TestGenerateV2_Specific_Validation(t *testing.T) {
	gen := NewSpecificPlanGenerator(newMockStoreRepo(), newMockSPV2Repo())

	_, err := gen.GenerateV2("", SpecificPlanInput{}, DomainGeneral, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty master plan id")
	}

	_, err = gen.GenerateV2("mp-1", SpecificPlanInput{ProjectID: "proj-1", Goal: ""}, DomainGeneral, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty goal")
	}
}

func TestAddResearchLink(t *testing.T) {
	v2Repo := newMockSPV2Repo()
	gen := NewSpecificPlanGenerator(newMockStoreRepo(), v2Repo)

	err := gen.AddResearchLink("sp-1", "res-1", "auth", 0.5)
	if err != nil {
		t.Fatalf("AddResearchLink failed: %v", err)
	}

	err = gen.AddResearchLink("sp-1", "res-2", "security", 1.5)
	if err == nil {
		t.Fatal("expected error for invalid relevance")
	}

	if len(v2Repo.researchLinks) != 1 {
		t.Errorf("expected 1 research link, got %d", len(v2Repo.researchLinks))
	}
}

func TestRegeneratePlan(t *testing.T) {
	storeRepo := newMockStoreRepo()
	v2Repo := newMockSPV2Repo()
	gen := NewSpecificPlanGenerator(storeRepo, v2Repo)

	// We need a base plan to exist for regeneration
	input := SpecificPlanInput{
		ProjectID: "proj-1",
		Title:     "Test",
		Goal:      "Test goal",
	}
	plan, err := gen.GenerateV2("mp-1", input, DomainGeneral, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateV2 failed: %v", err)
	}

	regenerated, err := gen.RegeneratePlan(plan.ID, "Requirements changed", RegenPartial)
	if err != nil {
		t.Fatalf("RegeneratePlan failed: %v", err)
	}
	if regenerated.Version != 2 {
		t.Errorf("expected version 2, got %d", regenerated.Version)
	}

	if len(v2Repo.regenerations) != 1 {
		t.Errorf("expected 1 regeneration record, got %d", len(v2Repo.regenerations))
	}
	reg := v2Repo.regenerations[0]
	if reg.VersionFrom != 1 {
		t.Errorf("expected version_from 1, got %d", reg.VersionFrom)
	}
	if reg.Scope != RegenPartial {
		t.Errorf("expected partial scope, got %s", reg.Scope)
	}
}

func TestGetResearchTrace(t *testing.T) {
	v2Repo := newMockSPV2Repo()
	gen := NewSpecificPlanGenerator(newMockStoreRepo(), v2Repo)

	gen.AddResearchLink("sp-1", "res-1", "auth", 0.8)
	gen.AddResearchLink("sp-1", "res-2", "security", 0.6)

	links, err := gen.GetResearchTrace("sp-1")
	if err != nil {
		t.Fatalf("GetResearchTrace failed: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestRenderSpecificPlanSummary(t *testing.T) {
	plan := &SpecificPlanV2{
		Title:        "Test",
		Version:      1,
		Domain:       DomainBackend,
		Status:       StatusDraft,
		Tasks:        []TaskDef{{Title: "Task 1", Order: 1, Effort: "1d"}},
		Dependencies: []DepDef{{Name: "lib-x", Type: "library", Critical: true}},
		ResearchUsed: []ResearchRef{{Summary: "OAuth2 research", Section: "auth", Relevance: 0.9}},
	}
	summary := RenderSpecificPlanSummary(plan)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

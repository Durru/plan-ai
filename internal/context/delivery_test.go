package context

import (
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// mockDeliveryRepo implements DeliveryRepository for testing.
type mockDeliveryRepo struct {
	sessions []DeliverySession
	usage    []DeliveryUsage
	budgets  map[string]DeliveryBudget
}

func newMockDeliveryRepo() *mockDeliveryRepo {
	return &mockDeliveryRepo{budgets: make(map[string]DeliveryBudget)}
}

func (m *mockDeliveryRepo) CreateSession(s DeliverySession) (DeliverySession, error) {
	m.sessions = append(m.sessions, s)
	return s, nil
}

func (m *mockDeliveryRepo) ListSessions(projectID string, level string, limit int) ([]DeliverySession, error) {
	return m.sessions, nil
}

func (m *mockDeliveryRepo) CreateUsage(u DeliveryUsage) (DeliveryUsage, error) {
	m.usage = append(m.usage, u)
	return u, nil
}

func (m *mockDeliveryRepo) GetTotalUsage(projectID string) (int, error) {
	total := 0
	for _, u := range m.usage {
		if u.ProjectID == projectID {
			total += u.Tokens
		}
	}
	return total, nil
}

func (m *mockDeliveryRepo) CreateOrUpdateBudget(b DeliveryBudget) (DeliveryBudget, error) {
	m.budgets[string(b.Level)] = b
	return b, nil
}

func (m *mockDeliveryRepo) GetBudget(projectID string, level DeliveryLevel) (DeliveryBudget, error) {
	b, ok := m.budgets[string(level)]
	if !ok {
		return DeliveryBudget{}, nil // zero value signals "not found"
	}
	return b, nil
}

// mockExecData implements ExecutiveData for testing.
type mockExecData struct{}

func (m *mockExecData) GetProjectBrief(projectID string) (domain.Project, error) {
	return domain.Project{Name: "Test Project", Status: "active"}, nil
}
func (m *mockExecData) ListPlanBriefs(projectID string) ([]domain.MasterPlan, error) {
	return []domain.MasterPlan{{Title: "Plan 1", Status: "approved"}}, nil
}
func (m *mockExecData) ListDecisionBriefs(projectID string) ([]domain.Decision, error) {
	return []domain.Decision{{Title: "Decision 1", Status: "draft"}}, nil
}

// mockPlanData implements PlanningData.
type mockPlanData struct{}

func (m *mockPlanData) ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error) {
	return []ApprovedItem{{Content: "Test " + string(itemType)}}, nil
}
func (m *mockPlanData) ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error) {
	return nil, nil
}
func (m *mockPlanData) ListResearchBriefs(projectID string) ([]ResearchBrief, error) {
	return nil, nil
}
func (m *mockPlanData) ListVisions(projectID string) ([]domain.Vision, error) {
	return nil, nil
}

// mockImplData implements ImplementationData.
type mockImplData struct{}

func (m *mockImplData) ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error) {
	return []ApprovedItem{{Content: "Test " + string(itemType)}}, nil
}
func (m *mockImplData) GetSpecificPlan(planID string) (domain.SpecificPlan, error) {
	return domain.SpecificPlan{}, nil
}
func (m *mockImplData) ListTasks(phaseID string) ([]domain.Task, error) {
	return nil, nil
}

// mockResearchData implements ResearchData.
type mockResearchData struct{}

func (m *mockResearchData) ListResearchBriefs(projectID string) ([]ResearchBrief, error) {
	return nil, nil
}
func (m *mockResearchData) ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error) {
	return nil, nil
}
func (m *mockResearchData) ListDecisionBriefs(projectID string) ([]domain.Decision, error) {
	return nil, nil
}

func TestNewDeliveryEngine(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestDeliverExecutiveContext(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	session, err := eng.DeliverContext("proj-1", LevelExecutive, map[string]string{"source": "cli"})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.Level != LevelExecutive {
		t.Errorf("expected L0_executive, got %s", session.Level)
	}
	if session.Status != DeliveryDelivered {
		t.Errorf("expected delivered, got %s", session.Status)
	}
	if session.Content == "" {
		t.Error("expected non-empty content")
	}
	if len(repo.sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(repo.sessions))
	}
	if len(repo.usage) != 1 {
		t.Errorf("expected 1 usage record, got %d", len(repo.usage))
	}
}

func TestDeliverPlanningContext(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	session, err := eng.DeliverContext("proj-1", LevelPlanning, map[string]string{})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}
	if session.Level != LevelPlanning {
		t.Errorf("expected L1_planning, got %s", session.Level)
	}
}

func TestDeliverImplementationContext(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	session, err := eng.DeliverContext("proj-1", LevelImplementation, map[string]string{"task_id": "task-1"})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}
	if session.Level != LevelImplementation {
		t.Errorf("expected L2_implementation, got %s", session.Level)
	}
}

func TestDeliverResearchContext(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	session, err := eng.DeliverContext("proj-1", LevelResearch, map[string]string{"topic": "auth"})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}
	if session.Level != LevelResearch {
		t.Errorf("expected L3_research, got %s", session.Level)
	}
}

func TestDeliverApprovalContext(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	session, err := eng.DeliverContext("proj-1", LevelApproval, map[string]string{})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}
	if session.Level != LevelApproval {
		t.Errorf("expected L4_approval, got %s", session.Level)
	}
}

func TestDeliverUnknownLevel(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	_, err := eng.DeliverContext("proj-1", "L5_unknown", map[string]string{})
	if err == nil {
		t.Fatal("expected error for unknown level")
	}
}

func TestDeliverBudgetExhausted(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	// Set budget to 0 to simulate exhaustion
	eng.SetBudget("proj-1", LevelExecutive, 0, BudgetFixed)

	_, err := eng.DeliverContext("proj-1", LevelExecutive, map[string]string{})
	if err == nil {
		t.Fatal("expected error for exhausted budget")
	}
}

func TestSetBudget(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	budget, err := eng.SetBudget("proj-1", LevelExecutive, 4096, BudgetDynamic)
	if err != nil {
		t.Fatalf("SetBudget failed: %v", err)
	}
	if budget.MaxTokens != 4096 {
		t.Errorf("expected 4096, got %d", budget.MaxTokens)
	}
	if budget.Strategy != BudgetDynamic {
		t.Errorf("expected dynamic, got %s", budget.Strategy)
	}
}

func TestGetUsageSummary(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	eng.DeliverContext("proj-1", LevelExecutive, map[string]string{})
	eng.DeliverContext("proj-1", LevelPlanning, map[string]string{})

	usage, err := eng.GetUsageSummary("proj-1")
	if err != nil {
		t.Fatalf("GetUsageSummary failed: %v", err)
	}
	if usage["total_tokens"] <= 0 {
		t.Errorf("expected positive token usage, got %d", usage["total_tokens"])
	}
}

func TestDeliveryLevels(t *testing.T) {
	levels := []DeliveryLevel{LevelExecutive, LevelPlanning, LevelImplementation, LevelResearch, LevelApproval}
	expected := []string{"L0_executive", "L1_planning", "L2_implementation", "L3_research", "L4_approval"}

	for i, l := range levels {
		if string(l) != expected[i] {
			t.Errorf("level %d: expected %s, got %s", i, expected[i], l)
		}
	}
}

func TestGetOrCreateBudget(t *testing.T) {
	repo := newMockDeliveryRepo()
	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, &mockResearchData{})

	budget, err := eng.getOrCreateBudget("proj-1", LevelImplementation)
	if err != nil {
		t.Fatalf("getOrCreateBudget failed: %v", err)
	}
	if budget.MaxTokens != 8192 {
		t.Errorf("expected default 8192 for implementation, got %d", budget.MaxTokens)
	}
}

func TestEstimateTokens(t *testing.T) {
	text := "Hello, this is a test of the token estimation function."
	tokens := estimateTokens(text)
	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
}

// domain stubs needed for compilation in the test file
var _ = domain.Project{}
var _ = time.Now

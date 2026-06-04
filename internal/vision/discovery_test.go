package vision

import (
	"testing"
	"time"
)

// mockDiscoveryRepo implements DiscoveryRepository for testing.
type mockDiscoveryRepo struct {
	sessions    map[string]DiscoverySession
	assumptions map[string]Assumption
	ambiguities map[string]Ambiguity
	approvals   map[string]VisionApproval
}

func newMockDiscoveryRepo() *mockDiscoveryRepo {
	return &mockDiscoveryRepo{
		sessions:    make(map[string]DiscoverySession),
		assumptions: make(map[string]Assumption),
		ambiguities: make(map[string]Ambiguity),
		approvals:   make(map[string]VisionApproval),
	}
}

func (m *mockDiscoveryRepo) CreateSession(s DiscoverySession) (DiscoverySession, error) {
	m.sessions[s.ID] = s
	return s, nil
}
func (m *mockDiscoveryRepo) GetSession(id string) (DiscoverySession, error) {
	return m.sessions[id], nil
}
func (m *mockDiscoveryRepo) ListSessions(projectID string) ([]DiscoverySession, error) {
	var out []DiscoverySession
	for _, s := range m.sessions {
		if s.ProjectID == projectID {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *mockDiscoveryRepo) UpdateSession(id string, status DiscoveryStatus, summary string) error {
	s := m.sessions[id]
	s.Status = status
	s.Summary = summary
	m.sessions[id] = s
	return nil
}
func (m *mockDiscoveryRepo) CreateAssumption(a Assumption) (Assumption, error) {
	m.assumptions[a.ID] = a
	return a, nil
}
func (m *mockDiscoveryRepo) ListAssumptions(sessionID string) ([]Assumption, error) {
	var out []Assumption
	for _, a := range m.assumptions {
		if a.SessionID == sessionID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (m *mockDiscoveryRepo) UpdateAssumptionStatus(id string, status AssumptionStatus, validatedBy string) error {
	a := m.assumptions[id]
	a.Status = status
	a.ValidatedBy = validatedBy
	t := time.Now()
	a.ValidatedAt = &t
	m.assumptions[id] = a
	return nil
}
func (m *mockDiscoveryRepo) CreateAmbiguity(a Ambiguity) (Ambiguity, error) {
	m.ambiguities[a.ID] = a
	return a, nil
}
func (m *mockDiscoveryRepo) ListAmbiguities(sessionID string) ([]Ambiguity, error) {
	var out []Ambiguity
	for _, a := range m.ambiguities {
		if a.SessionID == sessionID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (m *mockDiscoveryRepo) ResolveAmbiguity(id string, resolution string) error {
	a := m.ambiguities[id]
	a.Status = AmbiguityResolved
	a.Resolution = resolution
	t := time.Now()
	a.ResolvedAt = &t
	m.ambiguities[id] = a
	return nil
}
func (m *mockDiscoveryRepo) CreateApproval(a VisionApproval) (VisionApproval, error) {
	m.approvals[a.ID] = a
	return a, nil
}
func (m *mockDiscoveryRepo) GetApprovalByVision(visionID string) (VisionApproval, error) {
	for _, a := range m.approvals {
		if a.VisionID == visionID {
			return a, nil
		}
	}
	return VisionApproval{}, nil
}
func (m *mockDiscoveryRepo) ApproveApproval(id string, approvedBy string, feedback string) error {
	a := m.approvals[id]
	a.Status = ApprovalApproved
	a.ApprovedBy = approvedBy
	a.Feedback = feedback
	t := time.Now()
	a.ApprovedAt = &t
	m.approvals[id] = a
	return nil
}
func (m *mockDiscoveryRepo) RejectApproval(id string, feedback string) error {
	a := m.approvals[id]
	a.Status = ApprovalRejected
	a.Feedback = feedback
	m.approvals[id] = a
	return nil
}

func TestStartSession(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)

	session, err := eng.StartSession("proj-1", "raw context here")
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}
	if session.ProjectID != "proj-1" {
		t.Errorf("expected proj-1, got %s", session.ProjectID)
	}
	if session.Status != DiscoveryInProgress {
		t.Errorf("expected in_progress, got %s", session.Status)
	}
}

func TestIdentifyAssumption(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")

	_, err := eng.IdentifyAssumption(session.ID, "proj-1", "", AssumptionTechnical, 0.5)
	if err == nil {
		t.Fatal("expected error for empty description")
	}

	_, err = eng.IdentifyAssumption(session.ID, "proj-1", "test assumption", AssumptionTechnical, 1.5)
	if err == nil {
		t.Fatal("expected error for invalid confidence")
	}

	a, err := eng.IdentifyAssumption(session.ID, "proj-1", "users will find it intuitive", AssumptionUser, 0.4)
	if err != nil {
		t.Fatalf("IdentifyAssumption failed: %v", err)
	}
	if a.Category != AssumptionUser {
		t.Errorf("expected user category, got %s", a.Category)
	}
	if a.Status != AssumptionUnvalidated {
		t.Errorf("expected unvalidated, got %s", a.Status)
	}
}

func TestValidateAssumption(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")
	a, _ := eng.IdentifyAssumption(session.ID, "proj-1", "test", AssumptionTechnical, 0.5)

	if err := eng.ValidateAssumption(a.ID, true, "validator"); err != nil {
		t.Fatalf("ValidateAssumption failed: %v", err)
	}
	updated := repo.assumptions[a.ID]
	if updated.Status != AssumptionValidated {
		t.Errorf("expected validated, got %s", updated.Status)
	}
}

func TestIdentifyAmbiguity(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")

	_, err := eng.IdentifyAmbiguity(session.ID, "proj-1", "", "general")
	if err == nil {
		t.Fatal("expected error for empty description")
	}

	a, err := eng.IdentifyAmbiguity(session.ID, "proj-1", "unclear requirement", "requirements")
	if err != nil {
		t.Fatalf("IdentifyAmbiguity failed: %v", err)
	}
	if a.Status != AmbiguityOpen {
		t.Errorf("expected open, got %s", a.Status)
	}
}

func TestResolveAmbiguity(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")
	a, _ := eng.IdentifyAmbiguity(session.ID, "proj-1", "unclear", "general")

	err := eng.ResolveAmbiguity(a.ID, "")
	if err == nil {
		t.Fatal("expected error for empty resolution")
	}

	if err := eng.ResolveAmbiguity(a.ID, "clarified: use X"); err != nil {
		t.Fatalf("ResolveAmbiguity failed: %v", err)
	}
	updated := repo.ambiguities[a.ID]
	if updated.Status != AmbiguityResolved {
		t.Errorf("expected resolved, got %s", updated.Status)
	}
}

func TestCompleteSession(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")

	err := eng.CompleteSession(session.ID, "")
	if err == nil {
		t.Fatal("expected error for empty summary")
	}

	if err := eng.CompleteSession(session.ID, "All findings documented"); err != nil {
		t.Fatalf("CompleteSession failed: %v", err)
	}
	updated := repo.sessions[session.ID]
	if updated.Status != DiscoveryComplete {
		t.Errorf("expected complete, got %s", updated.Status)
	}
}

func TestSubmitAndApproveVision(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")

	approval, err := eng.SubmitForApproval("proj-1", session.ID, "vision-1")
	if err != nil {
		t.Fatalf("SubmitForApproval failed: %v", err)
	}
	if approval.Status != ApprovalPending {
		t.Errorf("expected pending, got %s", approval.Status)
	}

	// Approve
	if err := eng.ApproveVision(approval.ID, "reviewer", "looks good"); err != nil {
		t.Fatalf("ApproveVision failed: %v", err)
	}
	updated := repo.approvals[approval.ID]
	if updated.Status != ApprovalApproved {
		t.Errorf("expected approved, got %s", updated.Status)
	}
}

func TestSubmitAndRejectVision(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")
	approval, _ := eng.SubmitForApproval("proj-1", session.ID, "vision-1")

	err := eng.RejectVision(approval.ID, "")
	if err == nil {
		t.Fatal("expected error for empty rejection feedback")
	}

	if err := eng.RejectVision(approval.ID, "needs more analysis"); err != nil {
		t.Fatalf("RejectVision failed: %v", err)
	}
	updated := repo.approvals[approval.ID]
	if updated.Status != ApprovalRejected {
		t.Errorf("expected rejected, got %s", updated.Status)
	}
}

func TestHeuristicAnalyze(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")

	raw := "We assume users will understand the interface\n" +
		"The cost should probably be under budget\n" +
		"It is unclear what the API returns\n" +
		"We need to decide on the framework\n" +
		"This is a normal line\n"

	assumptions, ambiguities, err := eng.HeuristicAnalyze(session.ID, "proj-1", raw)
	if err != nil {
		t.Fatalf("HeuristicAnalyze failed: %v", err)
	}

	if len(assumptions) == 0 {
		t.Error("expected at least one assumption from heuristic analysis")
	}
	if len(ambiguities) == 0 {
		t.Error("expected at least one ambiguity from heuristic analysis")
	}
}

func TestGetDiscoverySummary(t *testing.T) {
	repo := newMockDiscoveryRepo()
	eng := NewDiscoveryEngine(repo)
	session, _ := eng.StartSession("proj-1", "ctx")
	eng.IdentifyAssumption(session.ID, "proj-1", "test assumption", AssumptionTechnical, 0.8)
	eng.IdentifyAmbiguity(session.ID, "proj-1", "unclear requirement", "general")
	eng.CompleteSession(session.ID, "Analysis complete")

	summary, err := eng.GetDiscoverySummary(session.ID)
	if err != nil {
		t.Fatalf("GetDiscoverySummary failed: %v", err)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

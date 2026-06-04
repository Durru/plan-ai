package agent

import (
	"testing"
)

func TestAgentRunRecord(t *testing.T) {
	r := AgentRunRecord{
		ID:        "run_1",
		ProjectID: "proj_1",
		Intent:    "create_master_plan",
		Status:    "processed",
	}
	if r.ID != "run_1" {
		t.Fatalf("expected run_1, got %s", r.ID)
	}
	if r.Status != "processed" {
		t.Fatalf("expected processed, got %s", r.Status)
	}
}

func TestDelegatedJob(t *testing.T) {
	j := DelegatedJob{
		ID:      "job_1",
		JobType: JobTypePlanning,
		Status:  JobStatusPending,
		Intent:  IntentCreateMasterPlan,
	}
	if j.Status != JobStatusPending {
		t.Fatalf("expected Pending, got %s", j.Status)
	}
}

func TestJobForIntent(t *testing.T) {
	job := JobForIntent("proj_1", IntentCreateMasterPlan, "planning")
	if job.ProjectID != "proj_1" {
		t.Fatalf("expected proj_1, got %s", job.ProjectID)
	}
	if job.Intent != IntentCreateMasterPlan {
		t.Fatalf("expected create_master_plan, got %s", job.Intent)
	}
}

func TestServiceProcessMessage_NoProject(t *testing.T) {
	s := NewService(
		&mockDetector{},
		&mockRouter{},
		&mockContextLoader{},
		&mockDelegator{},
		&mockResponseBuilder{},
		&mockAgentRunRepo{},
	)
	resp, err := s.ProcessMessage("", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "error" {
		t.Fatalf("expected error status, got %s", resp.Status)
	}
}

func TestServiceProcessMessage_UnknownIntent(t *testing.T) {
	s := NewService(
		&mockDetector{intent: IntentUnknown},
		&mockRouter{},
		&mockContextLoader{},
		&mockDelegator{},
		&mockResponseBuilder{},
		&mockAgentRunRepo{},
	)
	resp, err := s.ProcessMessage("proj_1", "blah blah")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success status, got %s", resp.Status)
	}
}

// ───── Mocks ─────

type mockDetector struct {
	intent IntentKind
}

func (m *mockDetector) DetectIntent(input string) IntentKind {
	if m.intent != "" {
		return m.intent
	}
	return IntentCreateMasterPlan
}

type mockRouter struct {
	decision RouterDecision
}

func (m *mockRouter) Route(intent IntentKind, ctx ContextPayload) RouterDecision {
	if m.decision.Workflow != "" {
		return m.decision
	}
	return RouterDecision{
		Workflow:   WorkflowPlanning,
		Capability: CapabilityPlanning,
	}
}

type mockContextLoader struct{}

func (m *mockContextLoader) Load(projectID string, keys []string) (ContextPayload, error) {
	return ContextPayload{ProjectID: projectID}, nil
}

type mockDelegator struct{}

func (m *mockDelegator) CreateJob(job DelegatedJob) (DelegatedJob, error)  { return job, nil }
func (m *mockDelegator) GetJob(id string) (DelegatedJob, error)            { return DelegatedJob{}, nil }
func (m *mockDelegator) ListJobs(projectID string) ([]DelegatedJob, error) { return nil, nil }
func (m *mockDelegator) UpdateJobStatus(id string, status JobStatus, summary string) error {
	return nil
}

type mockResponseBuilder struct{}

func (m *mockResponseBuilder) BuildSuccess(message string, decision RouterDecision) AgentResponse {
	return AgentResponse{Message: message, Status: "success"}
}
func (m *mockResponseBuilder) BuildApprovalRequired(message string, decision RouterDecision) AgentResponse {
	return AgentResponse{Message: message, Status: "approval_required"}
}
func (m *mockResponseBuilder) BuildError(err string) AgentResponse {
	return AgentResponse{Message: err, Status: "error"}
}

type mockAgentRunRepo struct{}

func (m *mockAgentRunRepo) CreateRun(run AgentRunRecord) (AgentRunRecord, error) { return run, nil }
func (m *mockAgentRunRepo) GetRun(id string) (AgentRunRecord, error)             { return AgentRunRecord{}, nil }
func (m *mockAgentRunRepo) UpdateRunStatus(id, status, response string) error    { return nil }
func (m *mockAgentRunRepo) ListRuns(projectID string, limit int) ([]AgentRunRecord, error) {
	return nil, nil
}
func (m *mockAgentRunRepo) CreateMessage(msg AgentMessage) (AgentMessage, error) { return msg, nil }
func (m *mockAgentRunRepo) ListMessages(runID string) ([]AgentMessage, error)    { return nil, nil }

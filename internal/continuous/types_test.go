package continuous

import (
	"testing"
)

func TestEventKind_Values(t *testing.T) {
	kinds := []EventKind{
		EventNewApprovedContext,
		EventNewResearch,
		EventNewKnowledge,
		EventDecisionChanged,
		EventPlanOutdated,
		EventImplementationFeedback,
		EventTaskCompleted,
		EventValidationFailed,
		EventChangeRequestCreated,
	}
	if len(kinds) != 9 {
		t.Fatalf("expected 9 event kinds, got %d", len(kinds))
	}
}

func TestContinuousEvent(t *testing.T) {
	e := ContinuousEvent{
		ID:        "evt_1",
		ProjectID: "proj_1",
		EventType: EventNewApprovedContext,
		Summary:   "New approved requirements",
	}
	if e.EventType != EventNewApprovedContext {
		t.Fatalf("expected new_approved_context, got %s", e.EventType)
	}
}

func TestProposalStatus_Values(t *testing.T) {
	statuses := []ProposalStatus{
		ProposalDraft,
		ProposalPendingApproval,
		ProposalApproved,
		ProposalRejected,
		ProposalApplied,
	}
	if len(statuses) != 5 {
		t.Fatalf("expected 5 proposal statuses, got %d", len(statuses))
	}
}

func TestPlanUpdateProposal(t *testing.T) {
	p := PlanUpdateProposal{
		ID:            "prop_1",
		ProjectID:     "proj_1",
		Reason:        "Plan needs updating",
		Status:        ProposalDraft,
		AffectedPlans: []string{"plan_1"},
	}
	if p.Status != ProposalDraft {
		t.Fatalf("expected draft, got %s", p.Status)
	}
	if len(p.AffectedPlans) != 1 {
		t.Fatalf("expected 1 affected plan, got %d", len(p.AffectedPlans))
	}
}

func TestContextLevel_Values(t *testing.T) {
	levels := []ContextLevel{
		ContextL0Executive,
		ContextL1Planning,
		ContextL2Plan,
		ContextL3Task,
		ContextL4Implementation,
	}
	if len(levels) != 5 {
		t.Fatalf("expected 5 context levels, got %d", len(levels))
	}
}

func TestContextDelivery(t *testing.T) {
	c := ContextDelivery{
		ID:        "del_1",
		ProjectID: "proj_1",
		Level:     ContextL3Task,
		Content:   "task detail",
	}
	if c.Level != ContextL3Task {
		t.Fatalf("expected L3_Task, got %s", c.Level)
	}
}

func TestProjectStatus(t *testing.T) {
	ps := ProjectStatus{
		ProjectID:        "proj_1",
		ActivePlan:       "plan_1",
		ActivePhase:      "phase_1",
		RecentEvents:     5,
		PendingProposals: 2,
	}
	if ps.RecentEvents != 5 {
		t.Fatalf("expected 5 recent events, got %d", ps.RecentEvents)
	}
}

func TestRepositories_Interface(t *testing.T) {
	var _ ContinuousEventRepository = (*mockEventRepo)(nil)
	var _ PlanUpdateProposalRepository = (*mockProposalRepo)(nil)
	var _ ContextDeliveryRepository = (*mockContextDeliveryRepo)(nil)
}

type mockEventRepo struct{}

func (m *mockEventRepo) CreateEvent(ev ContinuousEvent) (ContinuousEvent, error) { return ev, nil }
func (m *mockEventRepo) ListEvents(projectID string, limit int) ([]ContinuousEvent, error) {
	return nil, nil
}

type mockProposalRepo struct{}

func (m *mockProposalRepo) CreateProposal(p PlanUpdateProposal) (PlanUpdateProposal, error) {
	return p, nil
}
func (m *mockProposalRepo) GetProposal(id string) (PlanUpdateProposal, error) {
	return PlanUpdateProposal{}, nil
}
func (m *mockProposalRepo) ListProposals(projectID string) ([]PlanUpdateProposal, error) {
	return nil, nil
}
func (m *mockProposalRepo) UpdateProposalStatus(id string, status ProposalStatus) error { return nil }

type mockContextDeliveryRepo struct{}

func (m *mockContextDeliveryRepo) CreateDelivery(d ContextDelivery) (ContextDelivery, error) {
	return d, nil
}
func (m *mockContextDeliveryRepo) ListDeliveries(projectID string, level ContextLevel, limit int) ([]ContextDelivery, error) {
	return nil, nil
}

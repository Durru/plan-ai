package continuous_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	. "github.com/Durru/plan-ai/internal/continuous"
	"github.com/Durru/plan-ai/internal/store"
)

// ──────────────────────────────────────────────
// Continuous Planning E2E Integration Tests
// ──────────────────────────────────────────────

func TestContinuousE2E_DatabaseMigrationScenario(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	projectID := "project:e2e-pg-mariadb"
	eventStore := store.NewContinuousEventRepository(db)
	detector := NewDetector(db)
	planner := NewPlanner(e2eProposalRepo{db: db})
	approvalSvc := NewApprovalService(e2eProposalRepo{db: db})
	updater := NewUpdater(e2eProposalRepo{db: db})
	statusSvc := NewStatusService(db)

	// Step 1: Create event
	ev, err := eventStore.CreateEvent(store.ContinuousEventRecord{
		ID:        "ev:e2e:pg-mariadb",
		ProjectID: projectID,
		EventType: string(EventDecisionChanged),
		Summary:   "Switch database from PostgreSQL to MariaDB",
		Details:   `{"from":"postgresql","to":"mariadb","reason":"Licensing cost"}`,
		Source:    "decision",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if ev.ID == "" {
		t.Fatal("expected non-empty event ID")
	}

	// Step 2: Detect events
	events, err := detector.Detect(projectID)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("Detect returned 0 events")
	}
	found := false
	for _, e := range events {
		if e.EventType == EventDecisionChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("detected events missing EventDecisionChanged")
	}

	// Step 3: Detect outdated plans
	outdated, err := detector.DetectOutdatedPlans(projectID)
	if err != nil {
		t.Fatalf("DetectOutdatedPlans: %v", err)
	}
	if len(outdated) == 0 {
		t.Error("expected at least one outdated plan detection for decision_changed event")
	}

	// Step 4: Create proposal from event
	proposal, err := planner.CreateProposal(projectID, events[0],
		[]string{"master:db-migration"},
		[]string{"task:schema", "task:conn-string"},
		[]string{"decision:db-choice"},
	)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}
	if proposal.Status != ProposalDraft {
		t.Errorf("proposal status = %q, want %q", proposal.Status, ProposalDraft)
	}
	if proposal.RequiresApproval != true {
		t.Error("proposal should require approval")
	}

	// Step 5: Request approval
	requested, err := approvalSvc.RequestApproval(proposal.ID)
	if err != nil {
		t.Fatalf("RequestApproval: %v", err)
	}
	if requested.Status != ProposalPendingApproval {
		t.Errorf("status = %q, want %q", requested.Status, ProposalPendingApproval)
	}

	// Step 6: Approve
	approved, err := updater.Approve(proposal.ID)
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if approved.Status != ProposalApproved {
		t.Errorf("status = %q, want %q", approved.Status, ProposalApproved)
	}

	// Step 7: Apply
	applied, err := updater.Apply(proposal.ID)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.Status != ProposalApplied {
		t.Errorf("status = %q, want %q", applied.Status, ProposalApplied)
	}

	// Step 8: Verify status
	status, err := statusSvc.GetStatus(projectID)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.RecentEvents < 1 {
		t.Errorf("RecentEvents = %d, want >= 1", status.RecentEvents)
	}
	if status.PendingProposals != 0 {
		t.Errorf("PendingProposals = %d, want 0 (should be applied)", status.PendingProposals)
	}
}

func TestContinuousE2E_SingleToMultiTenantScenario(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	projectID := "project:e2e-multi-tenant"
	eventStore := store.NewContinuousEventRepository(db)
	detector := NewDetector(db)
	planner := NewPlanner(e2eProposalRepo{db: db})
	updater := NewUpdater(e2eProposalRepo{db: db})
	statusSvc := NewStatusService(db)

	// Create single→multi-tenant decision change
	_, err = eventStore.CreateEvent(store.ContinuousEventRecord{
		ID:        "ev:e2e:multi-tenant",
		ProjectID: projectID,
		EventType: string(EventDecisionChanged),
		Summary:   "Architecture change: Single Tenant to Multi Tenant",
		Details:   `{"from":"single-tenant","to":"multi-tenant","impact":"high"}`,
		Source:    "decision",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	// Detect
	events, err := detector.Detect(projectID)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("no events detected")
	}

	// Create, approve, apply
	proposal, err := planner.CreateProposal(projectID, events[0],
		[]string{"master:arch-change"},
		[]string{"task:tenant-model", "task:isolation"},
		[]string{"decision:tenant-arch"},
	)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}

	// Full lifecycle
	_, err = NewApprovalService(e2eProposalRepo{db: db}).RequestApproval(proposal.ID)
	if err != nil {
		t.Fatalf("RequestApproval: %v", err)
	}
	_, err = updater.Approve(proposal.ID)
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	_, err = updater.Apply(proposal.ID)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Status check
	status, err := statusSvc.GetStatus(projectID)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.PendingProposals != 0 {
		t.Errorf("PendingProposals = %d, want 0", status.PendingProposals)
	}
}

func TestContinuousE2E_StripeToLemonSqueezyScenario(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	projectID := "project:e2e-payment-switch"
	eventStore := store.NewContinuousEventRepository(db)
	detector := NewDetector(db)
	planner := NewPlanner(e2eProposalRepo{db: db})
	updater := NewUpdater(e2eProposalRepo{db: db})

	// Create change request
	_, err = eventStore.CreateEvent(store.ContinuousEventRecord{
		ID:        "ev:e2e:stripe-lemon",
		ProjectID: projectID,
		EventType: string(EventChangeRequestCreated),
		Summary:   "Payment provider switch: Stripe to LemonSqueezy",
		Details:   `{"from":"stripe","to":"lemonsqueezy"}`,
		Source:    "change_request",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	events, err := detector.Detect(projectID)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("no events detected")
	}

	// Create proposal - change requests do not force research (only EventNewResearch/EventNewKnowledge do)
	proposal, err := planner.CreateProposal(projectID, events[0],
		[]string{"master:billing"},
		[]string{"task:payment-api", "task:webhooks"},
		[]string{"decision:payment-provider"},
	)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}
	if proposal.RequiresResearch {
		t.Error("change request proposal should NOT require research by default")
	}

	// Full lifecycle
	_, _ = NewApprovalService(e2eProposalRepo{db: db}).RequestApproval(proposal.ID)
	_, _ = updater.Approve(proposal.ID)
	applied, err := updater.Apply(proposal.ID)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.Status != ProposalApplied {
		t.Errorf("status = %q", applied.Status)
	}

	// Verify proposal retrievable
	_, err = e2eProposalRepo{db: db}.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("GetProposal: %v", err)
	}
}

func TestContinuousE2E_AddAIFeatureScenario(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	projectID := "project:e2e-add-ai"
	eventStore := store.NewContinuousEventRepository(db)
	detector := NewDetector(db)
	planner := NewPlanner(e2eProposalRepo{db: db})
	updater := NewUpdater(e2eProposalRepo{db: db})

	_, err = eventStore.CreateEvent(store.ContinuousEventRecord{
		ID:        "ev:e2e:add-ai",
		ProjectID: projectID,
		EventType: string(EventChangeRequestCreated),
		Summary:   "New feature: Add AI-powered code review",
		Details:   `{"feature":"ai-code-review","priority":"high"}`,
		Source:    "product",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	events, err := detector.Detect(projectID)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("no events detected")
	}

	proposal, err := planner.CreateProposal(projectID, events[0],
		[]string{"master:ai-features"},
		[]string{"task:ai-model", "task:review-pipeline"},
		[]string{"decision:ai-provider"},
	)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}
	if proposal.ID == "" {
		t.Fatal("expected non-empty proposal ID")
	}

	// Reject instead of approve
	_, err = updater.Reject(proposal.ID)
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}

	// Verify it's rejected
	rejected, err := e2eProposalRepo{db: db}.GetProposal(proposal.ID)
	if err != nil {
		t.Fatalf("GetProposal after reject: %v", err)
	}
	if rejected.Status != ProposalRejected {
		t.Errorf("status = %q, want %q", rejected.Status, ProposalRejected)
	}
}

func TestContinuousE2E_ErrorCases(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := e2eProposalRepo{db: db}
	updater := NewUpdater(repo)

	t.Run("approve_nonexistent", func(t *testing.T) {
		_, err := updater.Approve("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent proposal")
		}
	})

	t.Run("reject_nonexistent", func(t *testing.T) {
		_, err := updater.Reject("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent proposal")
		}
	})

	t.Run("apply_nonexistent", func(t *testing.T) {
		_, err := updater.Apply("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent proposal")
		}
	})

	t.Run("apply_without_approval", func(t *testing.T) {
		p, err := repo.CreateProposal(PlanUpdateProposal{
			ID: "pup:e2e:no-approve", ProjectID: "project:test",
			Reason: "No approval test", Status: ProposalDraft,
		})
		if err != nil {
			t.Fatalf("CreateProposal: %v", err)
		}
		_, err = updater.Apply(p.ID)
		if err == nil {
			t.Fatal("expected error applying non-approved proposal")
		}
	})

	t.Run("approve_twice", func(t *testing.T) {
		p, err := repo.CreateProposal(PlanUpdateProposal{
			ID: "pup:e2e:approve-twice", ProjectID: "project:test",
			Reason: "Double approve test", Status: ProposalDraft,
		})
		if err != nil {
			t.Fatalf("CreateProposal: %v", err)
		}
		// First approve must go through pending approval
		_, _ = NewApprovalService(repo).RequestApproval(p.ID)
		_, err = updater.Approve(p.ID)
		if err != nil {
			t.Fatalf("First approve: %v", err)
		}
		_, err = updater.Approve(p.ID)
		if err == nil {
			t.Fatal("expected error approving already approved proposal")
		}
	})
}

// ──────────────────────────────────────────────
// Adapter: e2eProposalRepo wraps store repo to continuous interface
// ──────────────────────────────────────────────

type e2eProposalRepo struct {
	db *sql.DB
}

func (r e2eProposalRepo) inner() *store.PlanUpdateProposalRepository {
	return store.NewPlanUpdateProposalRepository(r.db)
}

func (r e2eProposalRepo) CreateProposal(p PlanUpdateProposal) (PlanUpdateProposal, error) {
	rec, err := r.inner().CreateProposal(toStoreRecord(p))
	if err != nil {
		return PlanUpdateProposal{}, err
	}
	return toContinuousProposal(rec), nil
}

func (r e2eProposalRepo) GetProposal(id string) (PlanUpdateProposal, error) {
	rec, err := r.inner().GetProposal(id)
	if err != nil {
		return PlanUpdateProposal{}, err
	}
	return toContinuousProposal(rec), nil
}

func (r e2eProposalRepo) ListProposals(projectID string) ([]PlanUpdateProposal, error) {
	records, err := r.inner().ListProposals(projectID)
	if err != nil {
		return nil, err
	}
	out := make([]PlanUpdateProposal, len(records))
	for i, rec := range records {
		out[i] = toContinuousProposal(rec)
	}
	return out, nil
}

func (r e2eProposalRepo) UpdateProposalStatus(id string, status ProposalStatus) error {
	return r.inner().UpdateProposalStatus(id, string(status))
}

func toStoreRecord(p PlanUpdateProposal) store.PlanUpdateProposalRecord {
	plans, _ := json.Marshal(p.AffectedPlans)
	tasks, _ := json.Marshal(p.AffectedTasks)
	decisions, _ := json.Marshal(p.AffectedDecisions)
	rr := 0
	if p.RequiresResearch {
		rr = 1
	}
	ra := 0
	if p.RequiresApproval {
		ra = 1
	}
	return store.PlanUpdateProposalRecord{
		ID: p.ID, ProjectID: p.ProjectID, Reason: p.Reason,
		AffectedPlans: string(plans), AffectedTasks: string(tasks),
		AffectedDecisions: string(decisions), SuggestedUpdates: p.SuggestedUpdates,
		RequiresResearch: rr, RequiresApproval: ra,
		Status: string(p.Status), CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toContinuousProposal(rec store.PlanUpdateProposalRecord) PlanUpdateProposal {
	var plans, tasks, decisions []string
	json.Unmarshal([]byte(rec.AffectedPlans), &plans)
	json.Unmarshal([]byte(rec.AffectedTasks), &tasks)
	json.Unmarshal([]byte(rec.AffectedDecisions), &decisions)
	return PlanUpdateProposal{
		ID: rec.ID, ProjectID: rec.ProjectID, Reason: rec.Reason,
		AffectedPlans: plans, AffectedTasks: tasks, AffectedDecisions: decisions,
		SuggestedUpdates: rec.SuggestedUpdates,
		RequiresResearch: rec.RequiresResearch != 0,
		RequiresApproval: rec.RequiresApproval != 0,
		Status:           ProposalStatus(rec.Status),
		CreatedAt:        rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

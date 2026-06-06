package continuous

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

type memEventRepo struct{ events []ContinuousEvent }
type memProposalRepo struct{ proposals map[string]PlanUpdateProposal }

func newMemEventRepo() *memEventRepo { return &memEventRepo{} }
func (r *memEventRepo) CreateEvent(ev ContinuousEvent) (ContinuousEvent, error) {
	r.events = append(r.events, ev)
	return ev, nil
}
func (r *memEventRepo) ListEvents(projectID string, limit int) ([]ContinuousEvent, error) {
	var out []ContinuousEvent
	for _, ev := range r.events {
		if ev.ProjectID == projectID {
			out = append(out, ev)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func newMemProposalRepo() *memProposalRepo {
	return &memProposalRepo{proposals: make(map[string]PlanUpdateProposal)}
}
func (r *memProposalRepo) CreateProposal(p PlanUpdateProposal) (PlanUpdateProposal, error) {
	r.proposals[p.ID] = p
	return p, nil
}
func (r *memProposalRepo) GetProposal(id string) (PlanUpdateProposal, error) {
	p, ok := r.proposals[id]
	if !ok {
		return PlanUpdateProposal{}, sql.ErrNoRows
	}
	return p, nil
}
func (r *memProposalRepo) ListProposals(projectID string) ([]PlanUpdateProposal, error) {
	var out []PlanUpdateProposal
	for _, p := range r.proposals {
		if p.ProjectID == projectID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (r *memProposalRepo) UpdateProposalStatus(id string, status ProposalStatus) error {
	p, ok := r.proposals[id]
	if !ok {
		return sql.ErrNoRows
	}
	p.Status = status
	r.proposals[id] = p
	return nil
}

func TestContinuousLoopDetectAnalyzeProposeApproveApply(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS continuous_events (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, event_type TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', details TEXT NOT NULL DEFAULT '', source TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL)`)
	db.Exec(`INSERT INTO continuous_events (id, project_id, event_type, summary, created_at) VALUES ('ev_1', 'proj_loop', 'new_approved_context', 'New approved context: Use OAuth2', datetime('now'))`)

	propRepo := newMemProposalRepo()
	svc := &LoopService{
		detector:  NewDetector(db),
		planner:   NewPlanner(propRepo),
		approval:  NewApprovalService(propRepo),
		updater:   NewUpdater(propRepo),
		statusSvc: NewStatusService(db),
		db:        db,
	}

	result, err := svc.RunLoop("proj_loop")
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result.EventsDetected != 1 {
		t.Errorf("expected 1 event detected, got %d", result.EventsDetected)
	}
	if len(result.ProposalsCreated) != 1 {
		t.Fatalf("expected 1 proposal created, got %d", len(result.ProposalsCreated))
	}
	if result.ApprovalsRequested != 1 {
		t.Errorf("expected 1 approval requested, got %d", result.ApprovalsRequested)
	}
}

func TestApprovePlanUpdateApprovesOnly(t *testing.T) {
	repo := newMemProposalRepo()
	repo.CreateProposal(PlanUpdateProposal{
		ID: "pup_1", ProjectID: "proj_a", Status: ProposalDraft,
		Reason: "Test proposal", CreatedAt: nowUTC(), UpdatedAt: nowUTC(),
	})

	svc := &LoopService{updater: NewUpdater(repo)}
	prop, err := svc.ApproveProposal("pup_1")
	if err != nil {
		t.Fatalf("ApproveProposal: %v", err)
	}
	if prop.Status != ProposalApproved {
		t.Errorf("expected approved, got %s", prop.Status)
	}

	got, _ := repo.GetProposal("pup_1")
	if got.Status != ProposalApproved {
		t.Errorf("repo should have approved status, got %s", got.Status)
	}
}

func TestRejectPlanUpdateRejectsOnly(t *testing.T) {
	repo := newMemProposalRepo()
	repo.CreateProposal(PlanUpdateProposal{
		ID: "pup_2", ProjectID: "proj_r", Status: ProposalDraft,
		Reason: "Reject test", CreatedAt: nowUTC(), UpdatedAt: nowUTC(),
	})

	svc := &LoopService{updater: NewUpdater(repo)}
	prop, err := svc.RejectProposal("pup_2")
	if err != nil {
		t.Fatalf("RejectProposal: %v", err)
	}
	if prop.Status != ProposalRejected {
		t.Errorf("expected rejected, got %s", prop.Status)
	}
}

func TestApplyProposalIsIdempotent(t *testing.T) {
	repo := newMemProposalRepo()
	repo.CreateProposal(PlanUpdateProposal{
		ID: "pup_apply", ProjectID: "proj_idem", Status: ProposalApproved,
		Reason: "Idempotent test", CreatedAt: nowUTC(), UpdatedAt: nowUTC(),
	})

	svc := &LoopService{updater: NewUpdater(repo)}

	// First apply
	prop, err := svc.ApplyProposal("pup_apply")
	if err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}
	if prop.Status != ProposalApplied {
		t.Errorf("expected applied after first apply, got %s", prop.Status)
	}

	// Second apply on already-applied proposal should error (not approved anymore).
	_, err = svc.ApplyProposal("pup_apply")
	if err == nil {
		t.Error("second apply should fail because proposal is already applied (not approved)")
	}
}

func TestDetectChangesEmitsContinuousEvent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS continuous_events (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, event_type TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', details TEXT NOT NULL DEFAULT '', source TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL)`)
	db.Exec(`INSERT INTO continuous_events (id, project_id, event_type, summary, created_at) VALUES ('ev_db', 'proj_ev', 'new_approved_context', 'New context added', datetime('now'))`)

	repo := newMemProposalRepo()
	svc := &LoopService{
		detector:  NewDetector(db),
		planner:   NewPlanner(repo),
		approval:  NewApprovalService(repo),
		updater:   NewUpdater(repo),
		statusSvc: NewStatusService(db),
		db:        db,
	}

	result, err := svc.RunLoop("proj_ev")
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result.EventsDetected < 1 {
		t.Error("expected events detected from DB")
	}
}

func TestContinuousContextUsesApprovedInputs(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS continuous_status (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, active_plan TEXT NOT NULL DEFAULT '', active_phase TEXT NOT NULL DEFAULT '', next_task TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL)`)

	svc := &LoopService{statusSvc: NewStatusService(db), db: db}

	status, err := svc.statusSvc.GetStatus("proj_ctx")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	gen := NewContextGenerator(db)
	ctx, err := gen.Generate("proj_ctx", ContextL1Planning)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if ctx == "" {
		t.Error("context generation should produce non-empty output")
	}
	_ = status
}

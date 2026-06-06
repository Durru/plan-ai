package guard

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/store"
)

func openGuardDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)"+
		"&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return db
}

func TestPlanningGuardBlocksWhenNoIntent(t *testing.T) {
	db := openGuardDB(t)
	g := NewPlanningGuard(db)

	ok, question := g.Check("proj_no_intent")
	if ok {
		t.Fatal("expected guard to block when no intents exist")
	}
	if question == "" {
		t.Fatal("expected a discovery question")
	}
	if !strings.Contains(question, "product intent") {
		t.Errorf("question should mention product intent, got: %s", question)
	}
}

func TestPlanningGuardReturnsNextDiscoveryQuestion(t *testing.T) {
	db := openGuardDB(t)
	g := NewPlanningGuard(db)

	ok, question := g.Check("proj_discovery")
	if ok {
		t.Fatal("expected block with no approved intent")
	}
	// Question must be actionable, not a generic error.
	if strings.Contains(question, "error") {
		t.Errorf("question should not contain 'error', got: %s", question)
	}
	if len(question) < 20 {
		t.Errorf("question too short, got: %s", question)
	}
}

func TestPlanningGuardPassesWhenApprovedIntentExists(t *testing.T) {
	db := openGuardDB(t)

	intentRepo := store.NewIntentV3Repository(db)
	pi, err := intentRepo.SaveProductIntent(intentv3.ProductIntent{
		ID:              "pi_approved",
		ProjectID:       "proj_approved",
		Description:     "A SaaS for task management",
		ExpectedOutcome: "Productive teams",
		Status:          "approved",
	})
	if err != nil {
		t.Fatalf("save approved intent: %v", err)
	}
	_ = pi

	g := NewPlanningGuard(db)
	ok, question := g.Check("proj_approved")
	if !ok {
		t.Fatalf("expected guard to pass with approved intent, got question: %s", question)
	}
	if question != "" {
		t.Errorf("expected empty question, got: %s", question)
	}
}

func TestGuardPlanningInputConvenienceFunction(t *testing.T) {
	db := openGuardDB(t)

	if err := GuardPlanningInput(db, "proj_missing"); err == nil {
		t.Fatal("expected error when no approved intent")
	}

	// After inserting an approved intent, it should pass.
	intentRepo := store.NewIntentV3Repository(db)
	_, err := intentRepo.SaveProductIntent(intentv3.ProductIntent{
		ID:          "pi_ok",
		ProjectID:   "proj_missing",
		Description: "An approved product",
		Status:      "approved",
	})
	if err != nil {
		t.Fatalf("save intent: %v", err)
	}

	if err := GuardPlanningInput(db, "proj_missing"); err != nil {
		t.Fatalf("expected pass after approved intent, got: %v", err)
	}
}

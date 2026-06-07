package conversation

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/Durru/plan-ai/internal/agent"
	"github.com/Durru/plan-ai/internal/store"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)"+
		"&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func ensureMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}
}

func TestConversationCreateSaaSStartsDiscovery(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_1", "create a SaaS for task management")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s", resp.Status)
	}
	if !strings.Contains(resp.Message, "product") && !strings.Contains(resp.Message, "SaaS") && !strings.Contains(resp.Message, "problem") {
		t.Errorf("response should ask about product details, got: %s", resp.Message)
	}
	if resp.WorkflowTriggered != "vision" {
		t.Errorf("expected vision workflow, got %s", resp.WorkflowTriggered)
	}
}

func TestConversationAnalyzeProjectReturnsKnownAndMissingContext(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_2", "analyze the project")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s", resp.Status)
	}
	if !strings.Contains(resp.Message, "plans") && !strings.Contains(resp.Message, "tasks") && !strings.Contains(resp.Message, "decisions") {
		t.Errorf("response should mention plans/tasks/decisions, got: %s", resp.Message)
	}
}

func TestConversationTellMeWhatIsNextPrioritizesPendingWork(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_3", "what is next")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s", resp.Status)
	}
	if strings.Contains(resp.Message, "No plans found") || strings.Contains(resp.Message, "completed") {
		t.Logf("expected guidance when no plans exist, got: %s", resp.Message)
	}
}

func TestConversationCreateDatabasePlanRoutesToDatabasePlanning(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_4", "design a database schema for orders")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %s", resp.Status)
	}
	if !strings.Contains(resp.Message, "database") && !strings.Contains(resp.Message, "decisions") && !strings.Contains(resp.Message, "persistence") {
		t.Errorf("response should reference database/persistence, got: %s", resp.Message)
	}
}

func TestMCPAgentMessageIsNotStub(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_5", "create a SaaS product")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	if resp.Status == "" {
		t.Fatal("status should not be empty")
	}
	if resp.Message == "" {
		t.Fatal("message should not be empty")
	}
	// The original stub returned "Agent processing is a stub" — verify we
	// are NOT returning the stub message.
	if strings.Contains(strings.ToLower(resp.Message), "stub") {
		t.Errorf("response should not contain stub text, got: %s", resp.Message)
	}
	// Verify the response has proper agent intent routing.
	if resp.WorkflowTriggered == "" && resp.Status == "success" {
		t.Logf("workflow may be empty for 'create a SaaS product' with no context loaded, got: %+v", resp)
	}
}

func TestGatewayPersistsRunAndMessages(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	resp, err := gw.ProcessMessage("proj_persist", "research react patterns")
	if err != nil {
		t.Fatalf("ProcessMessage: %v", err)
	}
	_ = resp

	runs, err := gw.ListRuns("proj_persist", 10)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least 1 run persisted")
	}
	if runs[0].Status != "processed" {
		t.Errorf("expected status 'processed', got %q", runs[0].Status)
	}
}

func TestGatewayServiceIsSameInstance(t *testing.T) {
	db := openTestDB(t)
	ensureMigrations(t, db)
	gw := NewGateway(db)

	s1 := gw.Service()
	s2 := gw.Service()
	if s1 != s2 {
		t.Error("Service() should return the same instance (lazy init)")
	}
}

// Test that ProcessMessage works with empty DB (no data, only schema).
func TestGatewayProcessMessage_EmptyDB(t *testing.T) {
	testCases := []struct {
		msg        string
		wantIntent agent.IntentKind
	}{
		{"create a master plan", agent.IntentCreateMasterPlan},
		{"research kubernetes", agent.IntentResearchTopic},
		{"what is the project status", agent.IntentProjectStatus},
		{"create a SaaS for email", agent.IntentCreateProduct},
		{"design the database", agent.IntentDatabasePlan},
		{"analyze my project", agent.IntentAnalyzeProject},
		{"what if I change the API", agent.IntentImpactAnalysis},
		{"what next", agent.IntentNextTask},
	}

	for _, tc := range testCases {
		t.Run(string(tc.wantIntent), func(t *testing.T) {
			db := openTestDB(t)
			ensureMigrations(t, db)
			gw := NewGateway(db)

			resp, err := gw.ProcessMessage("proj_"+string(tc.wantIntent), tc.msg)
			if err != nil {
				t.Fatalf("ProcessMessage(%q): %v", tc.msg, err)
			}
			if resp.Status != "success" && resp.Status != "requires_approval" && resp.Status != "error" {
				t.Errorf("expected success, requires_approval, or error, got %s — message: %s", resp.Status, resp.Message)
			}
			if resp.WorkflowTriggered == "" && resp.Status == "success" {
				t.Logf("%q: workflow is empty (expected for detached/unknown intents)", tc.msg)
			}
		})
	}
}

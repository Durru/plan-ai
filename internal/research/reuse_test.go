package research

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openReuseDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)"+
		"&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Minimal schema for reuse tests.
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS research_entries (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL DEFAULT '',
  topic TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT 'general',
  summary TEXT NOT NULL DEFAULT '',
  conclusion TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  confidence REAL NOT NULL DEFAULT 0,
  reuse_count INTEGER NOT NULL DEFAULT 0,
  reused_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT ''
);
CREATE VIRTUAL TABLE IF NOT EXISTS research_entries_fts USING fts5(id UNINDEXED, topic, objective, summary);
CREATE TABLE IF NOT EXISTS knowledge_objects (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL DEFAULT '',
  topic TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'general',
  summary TEXT NOT NULL DEFAULT '',
  confidence REAL NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS research_knowledge_links (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  knowledge_id TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT ''
);
`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func seedApprovedResearch(t *testing.T, db *sql.DB) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	for _, row := range []struct {
		id     string
		proj   string
		topic  string
		summary string
		status string
	}{
		{"r1", "proj_a", "OAuth2 best practices", "Use PKCE for SPA, rotate refresh tokens", "approved"},
		{"r2", "proj_a", "Database migration strategies", "Use expand-contract pattern", "approved"},
		{"r3", "proj_a", "Testing microservices", "Contract testing is essential", "draft"},
		{"r4", "proj_a", "OAuth2 in Go", "Use golang.org/x/oauth2 library", "approved"},
	} {
		_, err := db.Exec(`INSERT INTO research_entries (id, project_id, topic, summary, status, confidence, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			row.id, row.proj, row.topic, row.summary, row.status, 75, now, now)
		if err != nil {
			t.Fatalf("seed %s: %v", row.id, err)
		}
	}
}

func TestFindReusableResearchReturnsApprovedOnly(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	svc := NewReuseService(db, nil)
	results, err := svc.FindReusable("proj_a", "OAuth2")
	if err != nil {
		t.Fatalf("FindReusable: %v", err)
	}
	for _, r := range results {
		if r.Status != ResearchStatusApproved {
			t.Errorf("expected only approved entries, got %s with status %s", r.ID, r.Status)
		}
	}
	if len(results) < 2 {
		t.Errorf("expected >= 2 results for 'OAuth2', got %d", len(results))
	}
}

func TestResearchTopicReusesExistingApprovedResearch(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	svc := NewReuseService(db, nil)
	results, err := svc.FindReusable("proj_a", "database")
	if err != nil {
		t.Fatalf("FindReusable: %v", err)
	}
	found := false
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.Topic), "database") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find reusable database research")
	}
}

func TestApprovedResearchPromotesToKnowledge(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	knowledgeID, err := PromoteToKnowledge(db, "r1")
	if err != nil {
		t.Fatalf("PromoteToKnowledge: %v", err)
	}
	if knowledgeID == "" {
		t.Fatal("expected non-empty knowledge ID")
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM knowledge_objects WHERE id = ?`, knowledgeID).Scan(&count); err != nil {
		t.Fatalf("check knowledge: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 knowledge object created, got %d", count)
	}

	var linkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM research_knowledge_links WHERE research_id = 'r1'`).Scan(&linkCount); err != nil {
		t.Fatalf("check link: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("expected 1 research-knowledge link, got %d", linkCount)
	}
}

func TestPlanningExcludesDraftResearch(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	svc := NewReuseService(db, nil)
	results, err := svc.FindReusable("proj_a", "testing")
	if err != nil {
		t.Fatalf("FindReusable: %v", err)
	}
	for _, r := range results {
		if r.Status == ResearchStatusDraft {
			t.Errorf("draft research %s should not be reusable", r.ID)
		}
		if strings.EqualFold(r.ID, "r3") {
			t.Error("draft testing research (r3) should not appear in reusable results")
		}
	}
	_ = results
}

func TestResearchReuseIncrementsReuseCount(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	svc := NewReuseService(db, nil)

	if err := svc.IncrementReuseCount("r1"); err != nil {
		t.Fatalf("IncrementReuseCount: %v", err)
	}
	if err := svc.IncrementReuseCount("r1"); err != nil {
		t.Fatalf("second increment: %v", err)
	}

	var reuseCount int
	if err := db.QueryRow(`SELECT reuse_count FROM research_entries WHERE id = 'r1'`).Scan(&reuseCount); err != nil {
		t.Fatalf("query reuse_count: %v", err)
	}
	if reuseCount != 2 {
		t.Errorf("expected reuse_count=2, got %d", reuseCount)
	}
}

func TestReuseService_EnsureFTS(t *testing.T) {
	db := openReuseDB(t)
	seedApprovedResearch(t, db)

	svc := NewReuseService(db, nil)
	if err := svc.EnsureFTS(); err != nil {
		t.Fatalf("EnsureFTS: %v", err)
	}

	// Verify triggers exist by checking that an INSERT populates FTS.
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO research_entries (id, project_id, topic, summary, status, confidence, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"r_fts", "proj_fts", "FTS test topic", "FTS summary content", "approved", 80, now, now); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var ftsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM research_entries_fts WHERE research_entries_fts MATCH ?`, `"FTS"`).Scan(&ftsCount); err != nil {
		t.Fatalf("fts query: %v", err)
	}
	if ftsCount < 1 {
		t.Errorf("expected >= 1 FTS match, got %d", ftsCount)
	}
}

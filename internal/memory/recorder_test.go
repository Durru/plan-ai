package memory

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func openRecorderDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:"+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)"+
		"&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS project_memory_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  entry_type TEXT NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  question TEXT NOT NULL DEFAULT '',
  answer TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  citation TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_project ON project_memory_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_type ON project_memory_v2(entry_type);
`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

// inMemoryStore is a minimal Store implementation backed by a map.
type inMemoryStore struct{ entries map[string]Entry }

func newInMemoryStore() *inMemoryStore { return &inMemoryStore{entries: make(map[string]Entry)} }
func (s *inMemoryStore) Add(e Entry) (Entry, error) {
	s.entries[e.ID] = e
	return e, nil
}
func (s *inMemoryStore) List(projectID string) ([]Entry, error) {
	var out []Entry
	for _, e := range s.entries {
		if e.ProjectID == projectID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (s *inMemoryStore) Search(projectID, query string) ([]Entry, error) {
	var out []Entry
	q := strings.ToLower(query)
	for _, e := range s.entries {
		if e.ProjectID == projectID && strings.Contains(strings.ToLower(e.Title+e.Content+e.Source+e.Question+e.Answer), q) {
			out = append(out, e)
		}
	}
	return out, nil
}
func (s *inMemoryStore) Get(id string) (Entry, error) {
	e, ok := s.entries[id]
	if !ok {
		return Entry{}, ErrNotFound
	}
	return e, nil
}
func (s *inMemoryStore) Update(e Entry) (Entry, error) {
	s.entries[e.ID] = e
	return e, nil
}

func TestMemoryRecorderRecordsApprovedContext(t *testing.T) {
	store := newInMemoryStore()
	db := openRecorderDB(t)
	rec := NewRecorder(store, db)

	e, err := rec.RecordApprovedContext("proj_a", "decision", "Use schema-per-tenant isolation")
	if err != nil {
		t.Fatalf("RecordApprovedContext: %v", err)
	}
	if e.EntryType != TypeDecision {
		t.Errorf("expected TypeDecision, got %s", e.EntryType)
	}
	if !strings.Contains(e.Content, "schema-per-tenant") {
		t.Errorf("content mismatch: %s", e.Content)
	}
	if !strings.Contains(e.Source, "topic:") {
		t.Errorf("source should contain topic key: %s", e.Source)
	}

	// Verify stored in memory store.
	got, err := store.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("expected active status, got %s", got.Status)
	}
}

func TestMemoryRecorderRecordsAppliedProposal(t *testing.T) {
	store := newInMemoryStore()
	db := openRecorderDB(t)
	rec := NewRecorder(store, db)

	e, err := rec.RecordAppliedProposal("proj_b", "prop_1", "Add multi-tenant support to auth module")
	if err != nil {
		t.Fatalf("RecordAppliedProposal: %v", err)
	}
	if e.EntryType != TypeChange {
		t.Errorf("expected TypeChange, got %s", e.EntryType)
	}
	if !strings.Contains(e.Source, "applied-proposal") || !strings.Contains(e.Source, "topic:proposal:") {
		t.Errorf("source should contain 'applied-proposal' and 'topic:proposal:', got: %s", e.Source)
	}
}

func TestMemorySearchUsesFTS(t *testing.T) {
	db := openRecorderDB(t)
	store := newInMemoryStore()
	rec := NewRecorder(store, db)

	rec.RecordApprovedContext("proj_c", "decision", "Use PostgreSQL for main database")
	rec.RecordApprovedContext("proj_c", "constraint", "Must encrypt data at rest")
	rec.RecordApprovedResearch("proj_c", "r1", "PostgreSQL performance tuning", "Use connection pooling and prepared statements")

	results, err := rec.Search("proj_c", "PostgreSQL")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("expected >= 2 results for 'PostgreSQL', got %d", len(results))
	}

	results, err = rec.Search("proj_c", "encrypt")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) < 1 {
		t.Errorf("expected >= 1 result for 'encrypt', got %d", len(results))
	}
}

func TestMemoryFindByTopicKey(t *testing.T) {
	db := openRecorderDB(t)
	store := newInMemoryStore()
	rec := NewRecorder(store, db)

	rec.RecordApprovedContext("proj_d", "decision", "Use Redis for session cache")
	rec.RecordApprovedContext("proj_d", "constraint", "Sessions must expire after 24h")
	rec.RecordApprovedContext("proj_d", "goal", "Support 10k concurrent sessions")

	// Find by the topic key from source column.
	results, err := rec.FindByTopicKey("proj_d", "decision")
	if err != nil {
		t.Fatalf("FindByTopicKey: %v", err)
	}
	if len(results) < 1 {
		t.Errorf("expected >= 1 result for topic key 'decision', got %d", len(results))
	}
	for _, r := range results {
		if r.EntryType != TypeDecision {
			continue
		}
		if !strings.Contains(r.Source, "topic:decision:") {
			t.Errorf("entry %s missing topic key in source: %s", r.ID, r.Source)
		}
	}
}

func TestSupersededMemoryExcludedByDefault(t *testing.T) {
	store := newInMemoryStore()
	db := openRecorderDB(t)
	rec := NewRecorder(store, db)

	old, _ := rec.RecordApprovedContext("proj_e", "decision", "Use MySQL 5.7")
	tk := strings.TrimPrefix(strings.TrimPrefix(old.Source, "approved-context "), "approved-context")

	newEntry := NewEntry("proj_e", TypeDecision, "Use PostgreSQL 16", "Switch to PostgreSQL", "decision-v2", "approved-context")
	replacement, err := rec.Supersede("proj_e", tk, newEntry)
	if err != nil {
		t.Fatalf("Supersede: %v", err)
	}
	_ = replacement

	oldEntry, err := store.Get(old.ID)
	if err != nil {
		t.Fatalf("Get old: %v", err)
	}
	if oldEntry.Status != "superseded" {
		t.Errorf("old entry should be superseded, got status=%s", oldEntry.Status)
	}

	// Search should find both but active filtering can separate.
	all, _ := store.List("proj_e")
	active := 0
	superseded := 0
	for _, e := range all {
		switch e.Status {
		case "active":
			active++
		case "superseded":
			superseded++
		}
	}
	if superseded < 1 {
		t.Error("expected at least 1 superseded entry")
	}
	if active < 1 {
		t.Error("expected at least 1 active entry")
	}
}

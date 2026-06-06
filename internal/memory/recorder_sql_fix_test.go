package memory

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// sqliteStore implements memory.Store backed by an in-memory SQLite DB.
type sqliteStore struct {
	db *sql.DB
}

var _ Store = (*sqliteStore)(nil)

func newSQLiteStore(db *sql.DB) *sqliteStore {
	return &sqliteStore{db: db}
}

func (s *sqliteStore) Add(e Entry) (Entry, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = time.Now().UTC()
	}
	c := e.CreatedAt.Format(time.RFC3339)
	u := e.UpdatedAt.Format(time.RFC3339)
	_, err := s.db.Exec(`INSERT INTO project_memory_v2 (id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.ProjectID, string(e.EntryType), e.Title, e.Question, e.Answer, e.Content, e.Citation, e.Source, e.Status, c, u)
	_ = now
	return e, err
}

func (s *sqliteStore) List(projectID string) ([]Entry, error) {
	rows, err := s.db.Query(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE project_id=? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemRows(rows)
}

func (s *sqliteStore) Search(projectID, query string) ([]Entry, error) {
	q := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"
	rows, err := s.db.Query(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE project_id=? AND (title LIKE ? OR content LIKE ? OR source LIKE ? OR question LIKE ? OR answer LIKE ?) ORDER BY created_at`, projectID, q, q, q, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemRows(rows)
}

func (s *sqliteStore) Get(id string) (Entry, error) {
	row := s.db.QueryRow(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE id=?`, id)
	return scanMemRow(row)
}

func (s *sqliteStore) Update(e Entry) (Entry, error) {
	u := e.UpdatedAt.Format(time.RFC3339)
	_, err := s.db.Exec(`UPDATE project_memory_v2 SET title=?, question=?, answer=?, content=?, citation=?, source=?, status=?, updated_at=? WHERE id=?`,
		e.Title, e.Question, e.Answer, e.Content, e.Citation, e.Source, e.Status, u, e.ID)
	if err != nil {
		return e, err
	}
	return e, nil
}

func scanMemRows(rows *sql.Rows) ([]Entry, error) {
	var out []Entry
	for rows.Next() {
		e, err := scanMemRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanMemRow(row interface{ Scan(...any) error }) (Entry, error) {
	var e Entry
	var et, c, u string
	if err := row.Scan(&e.ID, &e.ProjectID, &et, &e.Title, &e.Question, &e.Answer, &e.Content, &e.Citation, &e.Source, &e.Status, &c, &u); err != nil {
		return Entry{}, err
	}
	e.EntryType = EntryType(et)
	e.CreatedAt, _ = time.Parse(time.RFC3339, c)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, u)
	return e, nil
}

func openRecorderFixDB(t *testing.T) *sql.DB {
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
CREATE INDEX IF NOT EXISTS idx_mem_project ON project_memory_v2(project_id);
`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

// ── Tests ──

func TestRecorderDBBacked_RecordApprovedContext(t *testing.T) {
	db := openRecorderFixDB(t)
	store := newSQLiteStore(db)
	rec := NewRecorder(store)

	e, err := rec.RecordApprovedContext("proj_a", "decision", "Use schema-per-tenant isolation")
	if err != nil {
		t.Fatalf("RecordApprovedContext: %v", err)
	}
	if e.EntryType != TypeDecision {
		t.Errorf("expected TypeDecision, got %s", e.EntryType)
	}

	// Verify persisted in DB
	got, err := store.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("expected active status, got %s", got.Status)
	}
	if !strings.Contains(got.Source, "topic:") {
		t.Errorf("source should contain topic key: %s", got.Source)
	}
}

func TestRecorderDBBacked_RecordAppliedProposal(t *testing.T) {
	db := openRecorderFixDB(t)
	store := newSQLiteStore(db)
	rec := NewRecorder(store)

	e, err := rec.RecordAppliedProposal("proj_b", "prop_1", "Add multi-tenant support to auth module")
	if err != nil {
		t.Fatalf("RecordAppliedProposal: %v", err)
	}
	if e.EntryType != TypeChange {
		t.Errorf("expected TypeChange, got %s", e.EntryType)
	}
	if !strings.Contains(e.Source, "applied-proposal") {
		t.Errorf("source should contain 'applied-proposal': %s", e.Source)
	}

	// Verify in DB
	got, err := store.Get(e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.EntryType != TypeChange {
		t.Errorf("persisted entry type: %s", got.EntryType)
	}
}

func TestRecorderDBBacked_FindByTopicKey(t *testing.T) {
	db := openRecorderFixDB(t)
	store := newSQLiteStore(db)
	rec := NewRecorder(store)

	e1, _ := rec.RecordApprovedContext("proj_d", "decision", "Use Redis for session cache")
	_ = e1
	e2, _ := rec.RecordApprovedContext("proj_d", "constraint", "Sessions must expire after 24h")
	_ = e2
	e3, _ := rec.RecordApprovedContext("proj_d", "goal", "Support 10k concurrent sessions")
	_ = e3

	results, err := rec.FindByTopicKey("proj_d", "decision")
	if err != nil {
		t.Fatalf("FindByTopicKey: %v", err)
	}
	if len(results) < 1 {
		t.Errorf("expected >= 1 result, got %d", len(results))
	}
	for _, r := range results {
		if r.EntryType == TypeDecision && !strings.Contains(r.Source, "topic:decision:") {
			t.Errorf("entry %s missing topic key in source: %s", r.ID, r.Source)
		}
	}
}

func TestRecorderDBBacked_Search(t *testing.T) {
	db := openRecorderFixDB(t)
	store := newSQLiteStore(db)
	rec := NewRecorder(store)

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
}

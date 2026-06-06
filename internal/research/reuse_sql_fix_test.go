package research

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// ftsTriggers mirrors the store.researchFTSTriggers for use in tests.
const ftsTriggers = `
INSERT OR IGNORE INTO research_entries_fts(research_entries_fts) VALUES('rebuild');

CREATE TRIGGER IF NOT EXISTS research_entries_fts_ai AFTER INSERT ON research_entries BEGIN
  INSERT INTO research_entries_fts(rowid, id, topic, objective, summary)
  VALUES (new.rowid, new.id, new.topic, new.topic, new.summary);
END;

CREATE TRIGGER IF NOT EXISTS research_entries_fts_ad AFTER DELETE ON research_entries BEGIN
  DELETE FROM research_entries_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS research_entries_fts_au AFTER UPDATE ON research_entries BEGIN
  DELETE FROM research_entries_fts WHERE rowid = old.rowid;
  INSERT INTO research_entries_fts(rowid, id, topic, objective, summary)
  VALUES (new.rowid, new.id, new.topic, new.topic, new.summary);
END;
`

// testFullRepo implements the full research.Repository (including
// IncrementReuseCount, EnsureFTS, and PromoteToKnowledge) backed by
// an in-memory SQLite database. This lets us test the refactored
// ReuseService end-to-end without touching the store package.
type testFullRepo struct {
	db *sql.DB
}

// Compile-time guard.
var _ Repository = (*testFullRepo)(nil)

func newFullRepo(db *sql.DB) *testFullRepo {
	return &testFullRepo{db: db}
}

// ── Entry CRUD ──

func (r *testFullRepo) CreateEntry(entry ResearchEntry) error {
	now := time.Now().UTC().Format(time.RFC3339)
	category := string(entry.Category)
	if category == "" {
		category = "general"
	}
	status := string(entry.Status)
	if status == "" {
		status = "draft"
	}
	_, err := r.db.Exec(`INSERT INTO research_entries (id, project_id, topic, source, category, summary, conclusion, status, confidence, created_at, updated_at) VALUES (?, ?, ?, '', ?, '', '', ?, ?, ?, ?)`,
		entry.ID, entry.ProjectID, entry.Topic, category, entry.Summary, status, float64(entry.Confidence), now, now)
	return err
}

func (r *testFullRepo) GetEntry(id string) (ResearchEntry, error) {
	row := r.db.QueryRow(`SELECT id, project_id, topic, category, summary, status, confidence, created_at, updated_at FROM research_entries WHERE id = ?`, id)
	var e ResearchEntry
	var c, u, cat, st string
	var conf float64
	if err := row.Scan(&e.ID, &e.ProjectID, &e.Topic, &cat, &e.Summary, &st, &conf, &c, &u); err != nil {
		return e, err
	}
	e.Category = ResearchCategory(cat)
	e.Status = ResearchStatus(st)
	e.Confidence = int(conf)
	e.CreatedAt, _ = time.Parse(time.RFC3339, c)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, u)
	return e, nil
}

func (r *testFullRepo) ListEntries() ([]ResearchEntry, error) {
	rows, err := r.db.Query(`SELECT id, project_id, topic, category, summary, status, confidence, created_at, updated_at FROM research_entries ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchEntry
	for rows.Next() {
		var e ResearchEntry
		var c, u, cat, st string
		var conf float64
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Topic, &cat, &e.Summary, &st, &conf, &c, &u); err != nil {
			return nil, err
		}
		e.Category = ResearchCategory(cat)
		e.Status = ResearchStatus(st)
		e.Confidence = int(conf)
		e.CreatedAt, _ = time.Parse(time.RFC3339, c)
		e.UpdatedAt, _ = time.Parse(time.RFC3339, u)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *testFullRepo) SearchEntries(query string) ([]ResearchEntry, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := r.db.Query(`SELECT id, project_id, topic, category, summary, status, confidence, created_at, updated_at FROM research_entries WHERE LOWER(topic) LIKE ? OR LOWER(summary) LIKE ? ORDER BY created_at`, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchEntry
	for rows.Next() {
		var e ResearchEntry
		var c, u, cat, st string
		var conf float64
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Topic, &cat, &e.Summary, &st, &conf, &c, &u); err != nil {
			return nil, err
		}
		e.Category = ResearchCategory(cat)
		e.Status = ResearchStatus(st)
		e.Confidence = int(conf)
		e.CreatedAt, _ = time.Parse(time.RFC3339, c)
		e.UpdatedAt, _ = time.Parse(time.RFC3339, u)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *testFullRepo) UpdateEntryStatus(id string, status ResearchStatus) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE research_entries SET status=?, updated_at=? WHERE id=?`, string(status), now, id)
	return err
}

func (r *testFullRepo) DeleteEntry(id string) error {
	tx, _ := r.db.Begin()
	defer tx.Rollback()
	tx.Exec(`DELETE FROM research_findings WHERE research_id=?`, id)
	tx.Exec(`DELETE FROM research_sources WHERE research_id=?`, id)
	tx.Exec(`DELETE FROM research_conclusions WHERE research_id=?`, id)
	tx.Exec(`DELETE FROM research_tags WHERE research_id=?`, id)
	tx.Exec(`DELETE FROM research_knowledge_links WHERE research_id=?`, id)
	tx.Exec(`DELETE FROM research_entries WHERE id=?`, id)
	return tx.Commit()
}

// ── Findings ──

func (r *testFullRepo) CreateFinding(f ResearchFinding) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO research_findings (id, research_id, title, content, importance, created_at) VALUES (?,?,?,?,?,?)`,
		f.ID, f.ResearchID, f.Title, f.Content, f.Importance, now)
	return err
}

func (r *testFullRepo) ListFindings(researchID string) ([]ResearchFinding, error) {
	rows, err := r.db.Query(`SELECT id, research_id, title, content, importance, created_at FROM research_findings WHERE research_id=?`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchFinding
	for rows.Next() {
		var f ResearchFinding
		var ca string
		rows.Scan(&f.ID, &f.ResearchID, &f.Title, &f.Content, &f.Importance, &ca)
		f.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		out = append(out, f)
	}
	return out, rows.Err()
}

// ── Sources ──

func (r *testFullRepo) CreateSource(s ResearchSource) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO research_sources (id, research_id, title, url, source_type, created_at) VALUES (?,?,?,?,?,?)`,
		s.ID, s.ResearchID, s.Title, s.URL, string(s.SourceType), now)
	return err
}

func (r *testFullRepo) ListSources(researchID string) ([]ResearchSource, error) {
	rows, err := r.db.Query(`SELECT id, research_id, title, url, source_type, created_at FROM research_sources WHERE research_id=?`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchSource
	for rows.Next() {
		var s ResearchSource
		var ca, st string
		rows.Scan(&s.ID, &s.ResearchID, &s.Title, &s.URL, &st, &ca)
		s.SourceType = ResearchSourceType(st)
		s.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		out = append(out, s)
	}
	return out, rows.Err()
}

// ── Conclusions ──

func (r *testFullRepo) CreateConclusion(c ResearchConclusion) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO research_conclusions (id, research_id, content, confidence, created_at) VALUES (?,?,?,?,?)`,
		c.ID, c.ResearchID, c.Content, c.Confidence, now)
	return err
}

func (r *testFullRepo) ListConclusions(researchID string) ([]ResearchConclusion, error) {
	rows, err := r.db.Query(`SELECT id, research_id, content, confidence, created_at FROM research_conclusions WHERE research_id=?`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchConclusion
	for rows.Next() {
		var c ResearchConclusion
		var ca string
		rows.Scan(&c.ID, &c.ResearchID, &c.Content, &c.Confidence, &ca)
		c.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		out = append(out, c)
	}
	return out, rows.Err()
}

// ── Tags ──

func (r *testFullRepo) AddTag(researchID, tag string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	id := fmt.Sprintf("tag_%s_%d", researchID, time.Now().UnixNano())
	_, err := r.db.Exec(`INSERT INTO research_tags (id, research_id, tag, created_at) VALUES (?,?,?,?)`, id, researchID, tag, now)
	return err
}

func (r *testFullRepo) ListTags(researchID string) ([]ResearchTag, error) {
	rows, err := r.db.Query(`SELECT id, research_id, tag FROM research_tags WHERE research_id=?`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchTag
	for rows.Next() {
		var t ResearchTag
		rows.Scan(&t.ID, &t.ResearchID, &t.Tag)
		out = append(out, t)
	}
	return out, rows.Err()
}

// ── Knowledge links ──

func (r *testFullRepo) LinkKnowledge(researchID, knowledgeID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	id := fmt.Sprintf("rklink_%s_%s", researchID, knowledgeID)
	_, err := r.db.Exec(`INSERT INTO research_knowledge_links (id, research_id, knowledge_id, created_at) VALUES (?,?,?,?)`, id, researchID, knowledgeID, now)
	return err
}

func (r *testFullRepo) ListKnowledgeLinks(researchID string) ([]ResearchKnowledgeLink, error) {
	rows, err := r.db.Query(`SELECT id, research_id, knowledge_id, created_at FROM research_knowledge_links WHERE research_id=?`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchKnowledgeLink
	for rows.Next() {
		var l ResearchKnowledgeLink
		var ca string
		rows.Scan(&l.ID, &l.ResearchID, &l.KnowledgeID, &ca)
		l.CreatedAt, _ = time.Parse(time.RFC3339, ca)
		out = append(out, l)
	}
	return out, rows.Err()
}

// ── Summary ──

func (r *testFullRepo) Summary() (ResearchSummary, error) {
	var s ResearchSummary
	rows, err := r.db.Query(`SELECT status, COUNT(*) FROM research_entries GROUP BY status`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int
		rows.Scan(&st, &n)
		s.Total += n
		switch ResearchStatus(st) {
		case ResearchStatusDraft:
			s.Draft = n
		case ResearchStatusInReview:
			s.InReview = n
		case ResearchStatusApproved:
			s.Approved = n
		case ResearchStatusRejected:
			s.Rejected = n
		case ResearchStatusArchived:
			s.Archived = n
		}
	}
	r.db.QueryRow(`SELECT COALESCE(COUNT(*),0) FROM research_findings`).Scan(&s.Findings)
	r.db.QueryRow(`SELECT COALESCE(COUNT(*),0) FROM research_sources`).Scan(&s.Sources)
	r.db.QueryRow(`SELECT COALESCE(COUNT(*),0) FROM research_conclusions`).Scan(&s.Conclusions)
	return s, nil
}

// ── Phase 7 / reuse methods (NEW on Repository) ──

func (r *testFullRepo) IncrementReuseCount(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE research_entries SET reuse_count = COALESCE(reuse_count, 0) + 1, reused_at = ? WHERE id = ?`, now, id)
	return err
}

func (r *testFullRepo) EnsureFTS() error {
	_, err := r.db.Exec(ftsTriggers)
	return err
}

func (r *testFullRepo) PromoteToKnowledge(researchID string) (knowledgeID string, err error) {
	entry, err := r.GetEntry(researchID)
	if err != nil {
		return "", fmt.Errorf("find research: %w", err)
	}
	if entry.Status != ResearchStatusApproved {
		return "", fmt.Errorf("research %s is not approved (status: %s)", researchID, entry.Status)
	}
	category := Classify(entry.Topic)
	knowledgeID = fmt.Sprintf("knowledge_%s", researchID)
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := r.db.Exec(`INSERT INTO knowledge_objects (id, project_id, topic, category, summary, confidence, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		knowledgeID, entry.ProjectID, entry.Topic, string(category), entry.Summary, float64(entry.Confidence), "approved", now, now); err != nil {
		return "", fmt.Errorf("create knowledge: %w", err)
	}
	// Link
	linkID := fmt.Sprintf("rklink_%s", researchID)
	if _, err := r.db.Exec(`INSERT INTO research_knowledge_links (id, research_id, knowledge_id, created_at) VALUES (?, ?, ?, ?)`,
		linkID, researchID, knowledgeID, now); err != nil {
		return knowledgeID, fmt.Errorf("link: %w", err)
	}
	return knowledgeID, nil
}

// ── Helper ──

func openReuseFixDB(t *testing.T) *sql.DB {
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
CREATE TABLE IF NOT EXISTS research_findings (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  importance INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS research_sources (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  title TEXT NOT NULL,
  url TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL DEFAULT 'manual',
  created_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS research_conclusions (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  confidence INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS research_tags (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  tag TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS research_knowledge_links (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  knowledge_id TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT ''
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
`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func seedApprovedFix(t *testing.T, db *sql.DB) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	rows := []struct {
		id, proj, topic, summary, status string
	}{
		{"r1", "proj_a", "OAuth2 best practices", "Use PKCE for SPA, rotate refresh tokens", "approved"},
		{"r2", "proj_a", "Database migration strategies", "Use expand-contract pattern", "approved"},
		{"r3", "proj_a", "Testing microservices", "Contract testing is essential", "draft"},
		{"r4", "proj_a", "OAuth2 in Go", "Use golang.org/x/oauth2 library", "approved"},
	}
	for _, row := range rows {
		if _, err := db.Exec(`INSERT INTO research_entries (id, project_id, topic, category, summary, status, confidence, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			row.id, row.proj, row.topic, "general", row.summary, row.status, 75, now, now); err != nil {
			t.Fatalf("seed %s: %v", row.id, err)
		}
	}
}

// ── Tests ──

func TestReuseWithRepo_FindReusable(t *testing.T) {
	db := openReuseFixDB(t)
	seedApprovedFix(t, db)

	repo := newFullRepo(db)
	svc := NewReuseService(repo)

	results, err := svc.FindReusable("proj_a", "OAuth2")
	if err != nil {
		t.Fatalf("FindReusable: %v", err)
	}
	for _, r := range results {
		if r.Status != ResearchStatusApproved {
			t.Errorf("expected approved only, got %s status=%s", r.ID, r.Status)
		}
	}
	if len(results) < 2 {
		t.Errorf("expected >= 2 results for 'OAuth2', got %d", len(results))
	}
}

func TestReuseWithRepo_IncrementReuseCount(t *testing.T) {
	db := openReuseFixDB(t)
	seedApprovedFix(t, db)

	repo := newFullRepo(db)
	svc := NewReuseService(repo)

	if err := svc.IncrementReuseCount("r1"); err != nil {
		t.Fatalf("first inc: %v", err)
	}
	if err := svc.IncrementReuseCount("r1"); err != nil {
		t.Fatalf("second inc: %v", err)
	}

	var reuseCount int
	if err := db.QueryRow(`SELECT reuse_count FROM research_entries WHERE id='r1'`).Scan(&reuseCount); err != nil {
		t.Fatalf("query: %v", err)
	}
	if reuseCount != 2 {
		t.Errorf("expected reuse_count=2, got %d", reuseCount)
	}
}

func TestReuseWithRepo_PromoteToKnowledge(t *testing.T) {
	db := openReuseFixDB(t)
	seedApprovedFix(t, db)

	repo := newFullRepo(db)
	svc := NewReuseService(repo)

	knowledgeID, err := svc.PromoteToKnowledge("r1")
	if err != nil {
		t.Fatalf("PromoteToKnowledge: %v", err)
	}
	if knowledgeID == "" {
		t.Fatal("expected non-empty knowledge ID")
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM knowledge_objects WHERE id=?`, knowledgeID).Scan(&count); err != nil {
		t.Fatalf("check knowledge: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 knowledge object, got %d", count)
	}

	var linkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM research_knowledge_links WHERE research_id='r1'`).Scan(&linkCount); err != nil {
		t.Fatalf("check link: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("expected 1 link, got %d", linkCount)
	}
}

func TestReuseWithRepo_EnsureFTS(t *testing.T) {
	db := openReuseFixDB(t)
	seedApprovedFix(t, db)

	repo := newFullRepo(db)
	svc := NewReuseService(repo)

	if err := svc.EnsureFTS(); err != nil {
		t.Fatalf("EnsureFTS: %v", err)
	}

	// Insert entry to trigger FTS population
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO research_entries (id, project_id, topic, category, summary, status, confidence, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"r_fts", "proj_fts", "FTS test topic", "general", "FTS summary content", "approved", 80, now, now); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var ftsCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM research_entries_fts`).Scan(&ftsCount); err != nil {
		t.Fatalf("fts query: %v", err)
	}
	if ftsCount < 1 {
		t.Errorf("expected >= 1 FTS row, got %d", ftsCount)
	}
}

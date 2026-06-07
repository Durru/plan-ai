package context

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// AuthorityService makes approved context the authority for what Plan-AI
// already knows. It provides:
//
//   - deduplicated storage (same content → same ID, no doubles)
//   - cross-type FTS search via approved_context_fts
//   - supersession (mark a fact as obsolete and link to its replacement)
//   - hooks for memory records and continuous planning events
//
// This implements Phase 6 (Approved Context Authority).
type AuthorityService struct {
	repo    Repository
	db      *sql.DB
	now     func() time.Time
	onAdd   func(item ApprovedItem) // memory + continuous event hook
}

// NewAuthorityService creates an authority service backed by the given repo
// and database. The db is used for FTS queries and migrations.
func NewAuthorityService(repo Repository, db *sql.DB) *AuthorityService {
	return &AuthorityService{repo: repo, db: db, now: time.Now}
}

// SetHook attaches a callback fired when an item is added or superseded.
func (s *AuthorityService) SetHook(fn func(item ApprovedItem)) {
	s.onAdd = fn
}

// Add stores an approved item. If a matching item already exists (same
// project, same content, same type), it returns the existing one without
// creating a duplicate.
//
// The returned bool is true if the item already existed.
func (s *AuthorityService) Add(item ApprovedItem) (approved ApprovedItem, existed bool, err error) {
	// Check if content already exists before storing.
	if s.IsKnown(item.ProjectID, item.Content) {
		existing, err2 := s.repo.FindApproved(item.ProjectID, item.Type, item.Content)
		if err2 == nil && len(existing) > 0 {
			return existing[0], true, nil
		}
	}
	approved, err = s.repo.StoreApproved(item)
	if err != nil {
		return approved, false, err
	}
	if s.onAdd != nil {
		s.onAdd(approved)
	}
	return approved, false, nil
}

// FindAll searches approved items of all types for matching content.
// Uses SQL LIKE by default; FTS is preferred when the approved_context_fts
// virtual table has been created (see EnsureFTS).
func (s *AuthorityService) FindAll(projectID string, query string) ([]ApprovedItem, error) {
	if s.db != nil && s.hasFTS() {
		return s.ftsSearch(projectID, query)
	}
	return s.likeSearch(projectID, query)
}

// EnsureFTS creates the FTS5 virtual table and triggers over all
// approved_* tables. Idempotent — safe to call multiple times.
func (s *AuthorityService) EnsureFTS() error {
	if s.db == nil {
		return fmt.Errorf("no database for FTS")
	}
	_, err := s.db.Exec(ftsSchema)
	return err
}

// Supersede marks an existing approved item (matching content) as
// "superseded" and creates a new item with the replacement content.
// Returns the new item.
func (s *AuthorityService) Supersede(projectID string, oldContent string, newItem ApprovedItem) (ApprovedItem, error) {
	// Find the old item to supersede.
	old, err := s.findByContent(projectID, newItem.Type, oldContent)
	if err != nil {
		return ApprovedItem{}, fmt.Errorf("find item to supersede: %w", err)
	}
	// Mark the old item as superseded (append supersedes_id to new content).
	if err := s.markSuperseded(old); err != nil {
		return ApprovedItem{}, fmt.Errorf("mark superseded: %w", err)
	}
	// Create replacement with link back.
	if newItem.ID == "" {
		newItem.ID = domain.NewID(string(newItem.Type) + "-v2")
	}
	newItem.SourceID = old.ID // link to superseded item
	approved, _, err := s.Add(newItem)
	if err != nil {
		return ApprovedItem{}, err
	}
	if s.onAdd != nil {
		s.onAdd(approved)
	}
	return approved, nil
}

// IsKnown checks if the exact content has already been approved for this
// project. Case-insensitive match.
func (s *AuthorityService) IsKnown(projectID string, content string) bool {
	for _, typ := range allTypes() {
		items, err := s.repo.FindApproved(projectID, typ, content)
		if err == nil && len(items) > 0 {
			return true
		}
	}
	return false
}

// ── private helpers ──

func (s *AuthorityService) findByContent(projectID string, typ ApprovedType, content string) (ApprovedItem, error) {
	items, err := s.repo.FindApproved(projectID, typ, content)
	if err != nil {
		return ApprovedItem{}, err
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Content), strings.TrimSpace(content)) {
			return item, nil
		}
	}
	return ApprovedItem{}, fmt.Errorf("not found")
}

func (s *AuthorityService) markSuperseded(item ApprovedItem) error {
	table, err := s.tableForType(item.Type)
	if err != nil {
		return err
	}
	now := s.now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`UPDATE `+table+` SET state = 'superseded', updated_at = ? WHERE id = ?`, now, item.ID)
	return err
}

func (s *AuthorityService) hasFTS() bool {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'approved_context_fts'`).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (s *AuthorityService) ftsSearch(projectID, query string) ([]ApprovedItem, error) {
	rows, err := s.db.Query(`SELECT id, project_id, approved_type, content, state, created_at FROM approved_context_fts WHERE project_id = ? AND approved_context_fts MATCH ? ORDER BY rank`, projectID, ftsQuery(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ApprovedItem
	for rows.Next() {
		var item ApprovedItem
		var c, u string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Type, &item.Content, &item.State, &c, &u); err != nil {
			continue
		}
		item.CreatedAt = parseTime(c)
		item.UpdatedAt = parseTime(u)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *AuthorityService) likeSearch(projectID, query string) ([]ApprovedItem, error) {
	var all []ApprovedItem
	for _, typ := range allTypes() {
		items, err := s.repo.FindApproved(projectID, typ, query)
		if err != nil {
			continue
		}
		all = append(all, items...)
	}
	return all, nil
}

func (s *AuthorityService) tableForType(typ ApprovedType) (string, error) {
	for t, tbl := range typeTableMap() {
		if t == typ {
			return tbl, nil
		}
	}
	return "", fmt.Errorf("unknown type: %s", typ)
}

func allTypes() []ApprovedType {
	return []ApprovedType{TypeRequirement, TypeConstraint, TypeDecision, TypePreference, TypeGoal, TypeReference}
}

func typeTableMap() map[ApprovedType]string {
	return map[ApprovedType]string{
		TypeRequirement: "approved_requirements",
		TypeConstraint:  "approved_constraints",
		TypeDecision:    "approved_decisions",
		TypePreference:  "approved_preferences",
		TypeGoal:        "approved_goals",
		TypeReference:   "approved_references",
	}
}

func ftsQuery(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return `""`
	}
	return strings.ReplaceAll(q, `"`, `""`)
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ftsSchema creates the FTS5 virtual table and triggers that keep it
// in sync with the 6 physical approved_* tables.
const ftsSchema = `
CREATE VIRTUAL TABLE IF NOT EXISTS approved_context_fts USING fts5(
  id, project_id, approved_type, content, state, created_at,
  content='approved_context_items',
  content_rowid='rowid'
);

-- triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS approved_context_fts_ai AFTER INSERT ON approved_requirements BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'requirement', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_ad AFTER DELETE ON approved_requirements BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'requirement', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_au AFTER UPDATE ON approved_requirements BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'requirement', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'requirement', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_constraints_ai AFTER INSERT ON approved_constraints BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'constraint', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_constraints_ad AFTER DELETE ON approved_constraints BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'constraint', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_constraints_au AFTER UPDATE ON approved_constraints BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'constraint', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'constraint', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_decisions_ai AFTER INSERT ON approved_decisions BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'decision', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_decisions_ad AFTER DELETE ON approved_decisions BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'decision', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_decisions_au AFTER UPDATE ON approved_decisions BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'decision', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'decision', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_prefs_ai AFTER INSERT ON approved_preferences BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'preference', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_prefs_ad AFTER DELETE ON approved_preferences BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'preference', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_prefs_au AFTER UPDATE ON approved_preferences BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'preference', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'preference', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_goals_ai AFTER INSERT ON approved_goals BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'goal', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_goals_ad AFTER DELETE ON approved_goals BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'goal', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_goals_au AFTER UPDATE ON approved_goals BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'goal', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'goal', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_refs_ai AFTER INSERT ON approved_references BEGIN
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'reference', new.content, new.state, new.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_refs_ad AFTER DELETE ON approved_references BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'reference', old.content, old.state, old.created_at);
END;

CREATE TRIGGER IF NOT EXISTS approved_context_fts_refs_au AFTER UPDATE ON approved_references BEGIN
  INSERT INTO approved_context_fts(approved_context_fts, rowid, id, project_id, approved_type, content, state, created_at)
  VALUES ('delete', old.rowid, old.id, old.project_id, 'reference', old.content, old.state, old.created_at);
  INSERT INTO approved_context_fts(rowid, id, project_id, approved_type, content, state, created_at)
  VALUES (new.rowid, new.id, new.project_id, 'reference', new.content, new.state, new.created_at);
END;
`

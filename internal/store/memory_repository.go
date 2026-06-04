package store

import (
	"database/sql"
	"strings"

	"github.com/plan-ai/plan-ai/internal/memory"
)

// MemoryRepository implements memory.Store backed by SQLite.
type MemoryRepository struct {
	db *sql.DB
}

// NewMemoryRepository creates a new memory repository.
func NewMemoryRepository(db *sql.DB) *MemoryRepository {
	return &MemoryRepository{db: db}
}

// Ensure MemoryRepository implements memory.Store.
var _ memory.Store = (*MemoryRepository)(nil)

// Add creates a new memory entry.
func (r *MemoryRepository) Add(e memory.Entry) (memory.Entry, error) {
	c, u := timestamps(e.CreatedAt, e.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO project_memory_v2 (id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.ProjectID, string(e.EntryType), e.Title, e.Question, e.Answer, e.Content, e.Citation, e.Source, e.Status, c, u)
	if err != nil {
		return memory.Entry{}, err
	}
	e.CreatedAt = parseRFC3339(c)
	e.UpdatedAt = parseRFC3339(u)
	return e, nil
}

// List returns all memory entries for a project, ordered by created_at.
func (r *MemoryRepository) List(projectID string) ([]memory.Entry, error) {
	rows, err := r.db.Query(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemoryEntries(rows)
}

// Search searches memory entries by keyword in title, content, and question.
func (r *MemoryRepository) Search(projectID, query string) ([]memory.Entry, error) {
	q := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"
	rows, err := r.db.Query(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE project_id = ? AND (title LIKE ? ESCAPE '\' OR content LIKE ? ESCAPE '\' OR question LIKE ?) ORDER BY created_at, id`, projectID, q, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemoryEntries(rows)
}

// Get returns a single memory entry by ID.
func (r *MemoryRepository) Get(id string) (memory.Entry, error) {
	row := r.db.QueryRow(`SELECT id, project_id, entry_type, title, question, answer, content, citation, source, status, created_at, updated_at FROM project_memory_v2 WHERE id = ?`, id)
	return scanSingleMemoryEntry(row)
}

// Update updates a memory entry.
func (r *MemoryRepository) Update(e memory.Entry) (memory.Entry, error) {
	_, u := timestamps(e.CreatedAt, e.UpdatedAt)
	_, err := r.db.Exec(`UPDATE project_memory_v2 SET title=?, question=?, answer=?, content=?, citation=?, source=?, status=?, updated_at=? WHERE id=?`,
		e.Title, e.Question, e.Answer, e.Content, e.Citation, e.Source, e.Status, u, e.ID)
	if err != nil {
		return memory.Entry{}, err
	}
	e.UpdatedAt = parseRFC3339(u)
	return e, nil
}

// ──────────────────────────────────────────────
// Scanning helpers
// ──────────────────────────────────────────────

func scanMemoryEntries(rows *sql.Rows) ([]memory.Entry, error) {
	var out []memory.Entry
	for rows.Next() {
		e, err := scanRowMemoryEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type memoryScanner interface {
	Scan(dest ...any) error
}

func scanRowMemoryEntry(s memoryScanner) (memory.Entry, error) {
	var e memory.Entry
	var et, c, u string
	if err := s.Scan(&e.ID, &e.ProjectID, &et, &e.Title, &e.Question, &e.Answer, &e.Content, &e.Citation, &e.Source, &e.Status, &c, &u); err != nil {
		return memory.Entry{}, err
	}
	e.EntryType = memory.EntryType(et)
	e.CreatedAt = parseRFC3339(c)
	e.UpdatedAt = parseRFC3339(u)
	return e, nil
}

func scanSingleMemoryEntry(row *sql.Row) (memory.Entry, error) {
	return scanRowMemoryEntry(row)
}

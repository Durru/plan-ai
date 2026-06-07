package repositories

import (
	"database/sql"
	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/research"
	"strings"
)

type ResearchRepository struct{ db *sql.DB }

func NewResearchRepository(db *sql.DB) ResearchRepository { return ResearchRepository{db: db} }

var _ domain.ResearchRepository = ResearchRepository{}

func (r ResearchRepository) Save(x domain.Research) error {
	x.ID = ensureID(x.ID, "research")
	if x.Status == "" {
		x.Status = string(research.ResearchStatusDraft)
	}
	if x.Category == "" {
		x.Category = domain.KnowledgeCategoryGeneral
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	d := x.Date.UTC().Format(timeFormat)
	if x.Date.IsZero() {
		d = c
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO research_entries (id, project_id, topic, objective, source, category, summary, conclusion, status, confidence, date, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,topic=excluded.topic,objective=excluded.objective,category=excluded.category,summary=excluded.summary,status=excluded.status,confidence=excluded.confidence,date=excluded.date,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Topic, x.Objective, "", x.Category, x.Summary, "", x.Status, x.Confidence, d, c, u); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO research_jobs (id, project_id, topic, summary, confidence, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, topic=excluded.topic, summary=excluded.summary, confidence=excluded.confidence, status=excluded.status`,
		x.ID, x.ProjectID, x.Topic, x.Summary, x.Confidence, string(x.Status), c); err != nil {
		return err
	}
	return tx.Commit()
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func (r ResearchRepository) GetByID(id string) (domain.Research, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.Research{}, err
	}
	if len(xs) == 0 {
		return domain.Research{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r ResearchRepository) ListByProject(projectID string) ([]domain.Research, error) {
	return r.list(`WHERE project_id=? ORDER BY created_at, id`, projectID)
}
func (r ResearchRepository) Search(query string) ([]domain.Research, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	return r.list(`WHERE LOWER(topic) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(objective) LIKE ? ORDER BY created_at, id`, like, like, like)
}
func (r ResearchRepository) list(where string, args ...any) ([]domain.Research, error) {
	rows, err := r.db.Query(`SELECT id, project_id, topic, objective, summary, confidence, date, category, status, created_at, updated_at FROM research_entries `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Research
	for rows.Next() {
		var x domain.Research
		var d, c, u, cat, st string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.Topic, &x.Objective, &x.Summary, &x.Confidence, &d, &cat, &st, &c, &u); err != nil {
			return nil, err
		}
		x.Date = parse(d)
		x.Category = domain.KnowledgeCategory(cat)
		x.Status = st
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r ResearchRepository) Delete(id string) error { return deleteByID(r.db, "research_entries", id) }

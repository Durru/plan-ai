package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type VisionRepository struct{ db *sql.DB }

func NewVisionRepository(db *sql.DB) VisionRepository { return VisionRepository{db: db} }

var _ domain.VisionRepository = VisionRepository{}

func (r VisionRepository) Save(v domain.Vision) error {
	v.ID = ensureID(v.ID, "vision")
	c, u := times(v.CreatedAt, v.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO visions (id, project_id, title, summary, expected_outcome, approved, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,title=excluded.title,summary=excluded.summary,expected_outcome=excluded.expected_outcome,approved=excluded.approved,updated_at=excluded.updated_at`, v.ID, v.ProjectID, v.Title, v.Summary, v.ExpectedOutcome, boolInt(v.Approved), c, u)
	return err
}
func (r VisionRepository) GetByID(id string) (domain.Vision, error) {
	rows, err := r.list(`WHERE id = ?`, id)
	if err != nil {
		return domain.Vision{}, err
	}
	if len(rows) == 0 {
		return domain.Vision{}, sql.ErrNoRows
	}
	return rows[0], nil
}
func (r VisionRepository) ListByProject(projectID string) ([]domain.Vision, error) {
	return r.list(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (r VisionRepository) list(where string, args ...any) ([]domain.Vision, error) {
	q := `SELECT id, project_id, title, summary, expected_outcome, approved, created_at, updated_at FROM visions ` + where
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Vision
	for rows.Next() {
		var v domain.Vision
		var approved int
		var c, u string
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Title, &v.Summary, &v.ExpectedOutcome, &approved, &c, &u); err != nil {
			return nil, err
		}
		v.Approved = intBool(approved)
		v.CreatedAt = parse(c)
		v.UpdatedAt = parse(u)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r VisionRepository) Approve(id string) error {
	_, err := r.db.Exec(`UPDATE visions SET approved = 1, updated_at = ? WHERE id = ?`, now(), id)
	return err
}
func (r VisionRepository) Delete(id string) error { return deleteByID(r.db, "visions", id) }

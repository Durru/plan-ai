package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type ConstraintRepository struct{ db *sql.DB }

func NewConstraintRepository(db *sql.DB) ConstraintRepository { return ConstraintRepository{db: db} }
func (r ConstraintRepository) Save(x domain.Constraint) error {
	x.ID = ensureID(x.ID, "constraint")
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO constraints (id, project_id, type, description, approved, created_at, updated_at) VALUES (?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,type=excluded.type,description=excluded.description,approved=excluded.approved,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Type, x.Description, boolInt(x.Approved), c, u)
	return err
}
func (r ConstraintRepository) GetByID(id string) (domain.Constraint, error) {
	xs, err := r.ListByProject("")
	if err != nil {
		return domain.Constraint{}, err
	}
	for _, x := range xs {
		if x.ID == id {
			return x, nil
		}
	}
	return domain.Constraint{}, sql.ErrNoRows
}
func (r ConstraintRepository) ListByProject(projectID string) ([]domain.Constraint, error) {
	q := `SELECT id, project_id, type, description, approved, created_at, updated_at FROM constraints`
	var args []any
	if projectID != "" {
		q += ` WHERE project_id = ?`
		args = []any{projectID}
	}
	q += ` ORDER BY created_at, id`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Constraint
	for rows.Next() {
		var x domain.Constraint
		var typ, c, u string
		var approved int
		if err := rows.Scan(&x.ID, &x.ProjectID, &typ, &x.Description, &approved, &c, &u); err != nil {
			return nil, err
		}
		x.Type = domain.ConstraintType(typ)
		x.Approved = intBool(approved)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r ConstraintRepository) Approve(id string) error {
	_, err := r.db.Exec(`UPDATE constraints SET approved=1, updated_at=? WHERE id=?`, now(), id)
	return err
}
func (r ConstraintRepository) Delete(id string) error { return deleteByID(r.db, "constraints", id) }

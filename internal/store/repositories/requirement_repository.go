package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type RequirementRepository struct{ db *sql.DB }

func NewRequirementRepository(db *sql.DB) RequirementRepository { return RequirementRepository{db: db} }

var _ domain.RequirementRepository = RequirementRepository{}

func (r RequirementRepository) Save(x domain.Requirement) error {
	x.ID = ensureID(x.ID, "requirement")
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO requirements (id, project_id, type, statement, approved, created_at, updated_at) VALUES (?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,type=excluded.type,statement=excluded.statement,approved=excluded.approved,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Type, x.Statement, boolInt(x.Approved), c, u)
	return err
}
func (r RequirementRepository) GetByID(id string) (domain.Requirement, error) {
	xs, err := r.list(`WHERE id = ?`, id)
	if err != nil {
		return domain.Requirement{}, err
	}
	if len(xs) == 0 {
		return domain.Requirement{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r RequirementRepository) ListByProject(projectID string) ([]domain.Requirement, error) {
	return r.list(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (r RequirementRepository) ListByType(projectID string, t domain.RequirementType) ([]domain.Requirement, error) {
	return r.list(`WHERE project_id = ? AND type = ? ORDER BY created_at, id`, projectID, t)
}
func (r RequirementRepository) list(where string, args ...any) ([]domain.Requirement, error) {
	rows, err := r.db.Query(`SELECT id, project_id, type, statement, approved, created_at, updated_at FROM requirements `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Requirement
	for rows.Next() {
		var x domain.Requirement
		var typ, c, u string
		var approved int
		if err := rows.Scan(&x.ID, &x.ProjectID, &typ, &x.Statement, &approved, &c, &u); err != nil {
			return nil, err
		}
		x.Type = domain.RequirementType(typ)
		x.Approved = intBool(approved)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r RequirementRepository) Approve(id string) error {
	_, err := r.db.Exec(`UPDATE requirements SET approved=1, updated_at=? WHERE id=?`, now(), id)
	return err
}
func (r RequirementRepository) Delete(id string) error { return deleteByID(r.db, "requirements", id) }

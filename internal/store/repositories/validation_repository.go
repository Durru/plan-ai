package repositories

import (
	"database/sql"
	"github.com/Durru/plan-ai/internal/domain"
)

type ValidationRepository struct{ db *sql.DB }

func NewValidationRepository(db *sql.DB) ValidationRepository { return ValidationRepository{db: db} }
func (r ValidationRepository) Save(x domain.Validation) error {
	x.ID = ensureID(x.ID, "validation")
	if x.Type == "" {
		x.Type = domain.ValidationTypeManual
	}
	if x.Status == "" {
		x.Status = domain.StatusDraft
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO validations (id, target_type, target_id, type, status, summary, details, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET target_type=excluded.target_type,target_id=excluded.target_id,type=excluded.type,status=excluded.status,summary=excluded.summary,details=excluded.details,updated_at=excluded.updated_at`, x.ID, x.TargetType, x.TargetID, x.Type, x.Status, x.Summary, x.Details, c, u)
	return err
}
func (r ValidationRepository) GetByID(id string) (domain.Validation, error) {
	rows, err := r.db.Query(`SELECT id,target_type,target_id,type,status,summary,details,created_at,updated_at FROM validations WHERE id=?`, id)
	if err != nil {
		return domain.Validation{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Validation{}, sql.ErrNoRows
	}
	var x domain.Validation
	var tt, typ, s, c, u string
	if err := rows.Scan(&x.ID, &tt, &x.TargetID, &typ, &s, &x.Summary, &x.Details, &c, &u); err != nil {
		return x, err
	}
	x.TargetType = domain.ValidationTargetType(tt)
	x.Type = domain.ValidationType(typ)
	x.Status = domain.Status(s)
	x.CreatedAt = parse(c)
	x.UpdatedAt = parse(u)
	return x, rows.Err()
}
func (r ValidationRepository) ListByProject(projectID string) ([]domain.Validation, error) {
	rows, err := r.db.Query(`SELECT id,target_type,target_id,type,status,summary,details,created_at,updated_at FROM validations WHERE project_id=? ORDER BY created_at,id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Validation
	for rows.Next() {
		var x domain.Validation
		var tt, typ, s, c, u string
		if err := rows.Scan(&x.ID, &tt, &x.TargetID, &typ, &s, &x.Summary, &x.Details, &c, &u); err != nil {
			return nil, err
		}
		x.TargetType = domain.ValidationTargetType(tt)
		x.Type = domain.ValidationType(typ)
		x.Status = domain.Status(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}

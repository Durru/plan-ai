package repositories

import (
	"database/sql"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type PhaseRepository struct{ db *sql.DB }

func NewPhaseRepository(db *sql.DB) PhaseRepository { return PhaseRepository{db: db} }

func (r PhaseRepository) Save(x domain.Phase) error {
	x.ID = ensureID(x.ID, "phase")
	if x.Status == "" {
		x.Status = domain.PlanStatusPending
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO phases (id, plan_id, title, summary, status, position, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET plan_id=excluded.plan_id, title=excluded.title, summary=excluded.summary, status=excluded.status, position=excluded.position, updated_at=excluded.updated_at`,
		x.ID, x.PlanID, x.Title, x.Summary, x.Status, x.Position, c, u)
	return err
}

func (r PhaseRepository) GetByID(id string) (domain.Phase, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.Phase{}, err
	}
	if len(xs) == 0 {
		return domain.Phase{}, sql.ErrNoRows
	}
	return xs[0], nil
}

func (r PhaseRepository) ListByPlan(planID string) ([]domain.Phase, error) {
	return r.list(`WHERE plan_id=? ORDER BY position, created_at, id`, planID)
}

func (r PhaseRepository) List() ([]domain.Phase, error) {
	return r.list(`ORDER BY position, created_at, id`)
}

func (r PhaseRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "phases", id, status)
}

func (r PhaseRepository) Delete(id string) error { return deleteByID(r.db, "phases", id) }

func (r PhaseRepository) list(where string, args ...any) ([]domain.Phase, error) {
	rows, err := r.db.Query(`SELECT id, plan_id, title, summary, status, position, created_at, updated_at FROM phases `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Phase
	for rows.Next() {
		var x domain.Phase
		var s, c, u string
		if err := rows.Scan(&x.ID, &x.PlanID, &x.Title, &x.Summary, &s, &x.Position, &c, &u); err != nil {
			return nil, err
		}
		x.Status = domain.Status(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}

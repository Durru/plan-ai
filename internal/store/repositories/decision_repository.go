package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type DecisionRepository struct{ db *sql.DB }

func NewDecisionRepository(db *sql.DB) DecisionRepository { return DecisionRepository{db: db} }

var _ domain.DecisionRepository = DecisionRepository{}

func (r DecisionRepository) Save(x domain.Decision) error {
	x.ID = ensureID(x.ID, "decision")
	if x.Status == "" {
		x.Status = domain.DecisionProposed
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO decisions (id, project_id, title, context, decision, rationale, alternatives, status, impact, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,title=excluded.title,context=excluded.context,decision=excluded.decision,rationale=excluded.rationale,alternatives=excluded.alternatives,status=excluded.status,impact=excluded.impact,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Title, x.Context, x.Decision, x.Rationale, x.Alternatives, x.Status, x.Impact, c, u)
	return err
}
func (r DecisionRepository) GetByID(id string) (domain.Decision, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.Decision{}, err
	}
	if len(xs) == 0 {
		return domain.Decision{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r DecisionRepository) ListByProject(projectID string) ([]domain.Decision, error) {
	return r.list(`WHERE project_id=? ORDER BY created_at, id`, projectID)
}
func (r DecisionRepository) list(where string, args ...any) ([]domain.Decision, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, context, decision, rationale, alternatives, status, impact, created_at, updated_at FROM decisions `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Decision
	for rows.Next() {
		var x domain.Decision
		var s, c, u string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.Title, &x.Context, &x.Decision, &x.Rationale, &x.Alternatives, &s, &x.Impact, &c, &u); err != nil {
			return nil, err
		}
		x.Status = domain.Status(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r DecisionRepository) UpdateStatus(id string, status domain.Status) error {
	_, err := r.db.Exec(`INSERT INTO decision_history (id, decision_id, to_status, created_at) VALUES (?, ?, ?, ?)`, ensureID("", "dh"), id, status, now())
	if err != nil {
		return err
	}
	return updateStatus(r.db, "decisions", id, status)
}
func (r DecisionRepository) Delete(id string) error { return deleteByID(r.db, "decisions", id) }

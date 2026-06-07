package repositories

import (
	"database/sql"
	"github.com/Durru/plan-ai/internal/domain"
)

type TaskRepository struct{ db *sql.DB }

func NewTaskRepository(db *sql.DB) TaskRepository { return TaskRepository{db: db} }

var _ domain.TaskRepository = TaskRepository{}

func (r TaskRepository) Save(x domain.Task) error {
	x.ID = ensureID(x.ID, "task")
	if x.Status == "" {
		x.Status = domain.PlanStatusPending
	}
	if x.ContextSize == "" {
		x.ContextSize = domain.ContextSizeShort
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO tasks (id, phase_id, plan_id, title, summary, status, position, context_size, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET phase_id=excluded.phase_id,plan_id=excluded.plan_id,title=excluded.title,summary=excluded.summary,status=excluded.status,position=excluded.position,context_size=excluded.context_size,updated_at=excluded.updated_at`, x.ID, x.PhaseID, x.PlanID, x.Title, x.Summary, x.Status, x.Position, x.ContextSize, c, u)
	return err
}
func (r TaskRepository) GetByID(id string) (domain.Task, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.Task{}, err
	}
	if len(xs) == 0 {
		return domain.Task{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r TaskRepository) ListByPhase(id string) ([]domain.Task, error) {
	return r.list(`WHERE phase_id=? ORDER BY position, created_at, id`, id)
}
func (r TaskRepository) List() ([]domain.Task, error) {
	return r.list(`ORDER BY position, created_at, id`)
}
func (r TaskRepository) ListByPlanID(planID string) ([]domain.Task, error) {
	return r.list(`WHERE plan_id=? ORDER BY position, created_at, id`, planID)
}
func (r TaskRepository) ListByStatus(status string) ([]domain.Task, error) {
	return r.list(`WHERE status=? ORDER BY position, created_at, id`, status)
}
func (r TaskRepository) list(where string, args ...any) ([]domain.Task, error) {
	rows, err := r.db.Query(`SELECT id, phase_id, plan_id, title, summary, status, position, context_size, created_at, updated_at FROM tasks `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Task
	for rows.Next() {
		var x domain.Task
		var s, cs, c, u string
		if err := rows.Scan(&x.ID, &x.PhaseID, &x.PlanID, &x.Title, &x.Summary, &s, &x.Position, &cs, &c, &u); err != nil {
			return nil, err
		}
		x.Status = domain.Status(s)
		x.ContextSize = domain.ContextSize(cs)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r TaskRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "tasks", id, status)
}
func (r TaskRepository) Delete(id string) error { return deleteByID(r.db, "tasks", id) }

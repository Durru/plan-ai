package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type SnapshotRepository struct{ db *sql.DB }

func NewSnapshotRepository(db *sql.DB) SnapshotRepository { return SnapshotRepository{db: db} }

var _ domain.SnapshotRepository = SnapshotRepository{}

func (r SnapshotRepository) Save(x domain.Snapshot) error {
	x.ID = ensureID(x.ID, "snapshot")
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO snapshots (id, project_id, reason, summary, created_at, updated_at) VALUES (?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,reason=excluded.reason,summary=excluded.summary,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Reason, x.Summary, c, u)
	return err
}
func (r SnapshotRepository) GetByID(id string) (domain.Snapshot, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.Snapshot{}, err
	}
	if len(xs) == 0 {
		return domain.Snapshot{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r SnapshotRepository) ListByProject(pid string) ([]domain.Snapshot, error) {
	return r.list(`WHERE project_id=? ORDER BY created_at,id`, pid)
}
func (r SnapshotRepository) list(where string, args ...any) ([]domain.Snapshot, error) {
	rows, err := r.db.Query(`SELECT id, project_id, reason, summary, created_at, updated_at FROM snapshots `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Snapshot
	for rows.Next() {
		var x domain.Snapshot
		var c, u string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.Reason, &x.Summary, &c, &u); err != nil {
			return nil, err
		}
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r SnapshotRepository) Delete(id string) error { return deleteByID(r.db, "snapshots", id) }

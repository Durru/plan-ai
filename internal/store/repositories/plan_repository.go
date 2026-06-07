package repositories

import (
	"database/sql"
	"github.com/Durru/plan-ai/internal/domain"
)

type PlanRepository struct{ db *sql.DB }

func NewPlanRepository(db *sql.DB) PlanRepository { return PlanRepository{db: db} }

var _ domain.PlanRepository = PlanRepository{}

func (r PlanRepository) SaveMaster(x domain.MasterPlan) error {
	x.ID = ensureID(x.ID, "plan")
	if x.Status == "" {
		x.Status = domain.StatusDraft
	}
	if x.Version == 0 {
		x.Version = 1
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO master_plans (id, project_id, title, summary, status, version, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,title=excluded.title,summary=excluded.summary,status=excluded.status,version=excluded.version,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Title, x.Summary, x.Status, x.Version, c, u)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO plans (id,type,title,summary,status,version,parent_plan_id,created_at,updated_at) VALUES (?,'master',?,?,?,?,NULL,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,summary=excluded.summary,status=excluded.status,version=excluded.version,updated_at=excluded.updated_at`, x.ID, x.Title, x.Summary, x.Status, x.Version, c, u)
	return err
}
func (r PlanRepository) SaveSpecific(x domain.SpecificPlan) error {
	x.ID = ensureID(x.ID, "plan")
	if x.MasterPlanID == "" {
		x.MasterPlanID = x.ParentPlanID
	}
	if x.Status == "" {
		x.Status = domain.StatusDraft
	}
	if x.Version == 0 {
		x.Version = 1
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO specific_plans (id, project_id, master_plan_id, title, summary, status, version, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,master_plan_id=excluded.master_plan_id,title=excluded.title,summary=excluded.summary,status=excluded.status,version=excluded.version,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.MasterPlanID, x.Title, x.Summary, x.Status, x.Version, c, u)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`INSERT INTO plans (id,type,title,summary,status,version,parent_plan_id,created_at,updated_at) VALUES (?,'specific',?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,summary=excluded.summary,status=excluded.status,version=excluded.version,parent_plan_id=excluded.parent_plan_id,updated_at=excluded.updated_at`, x.ID, x.Title, x.Summary, x.Status, x.Version, x.MasterPlanID, c, u)
	return err
}
func (r PlanRepository) GetMasterByID(id string) (domain.MasterPlan, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, summary, status, version, created_at, updated_at FROM master_plans WHERE id=?`, id)
	if err != nil {
		return domain.MasterPlan{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.MasterPlan{}, sql.ErrNoRows
	}
	var x domain.MasterPlan
	var s, c, u string
	if err := rows.Scan(&x.ID, &x.ProjectID, &x.Title, &x.Summary, &s, &x.Version, &c, &u); err != nil {
		return x, err
	}
	x.Status = domain.Status(s)
	x.CreatedAt = parse(c)
	x.UpdatedAt = parse(u)
	return x, rows.Err()
}
func (r PlanRepository) GetSpecificByID(id string) (domain.SpecificPlan, error) {
	xs, err := r.ListSpecificsByMaster("")
	if err != nil {
		return domain.SpecificPlan{}, err
	}
	for _, x := range xs {
		if x.ID == id {
			return x, nil
		}
	}
	return domain.SpecificPlan{}, sql.ErrNoRows
}
func (r PlanRepository) ListMastersByProject(pid string) ([]domain.MasterPlan, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, summary, status, version, created_at, updated_at FROM master_plans WHERE project_id=? ORDER BY created_at,id`, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MasterPlan
	for rows.Next() {
		var x domain.MasterPlan
		var s, c, u string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.Title, &x.Summary, &s, &x.Version, &c, &u); err != nil {
			return nil, err
		}
		x.Status = domain.Status(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r PlanRepository) ListSpecificsByMaster(mid string) ([]domain.SpecificPlan, error) {
	q := `SELECT id, project_id, master_plan_id, title, summary, status, version, created_at, updated_at FROM specific_plans`
	var args []any
	if mid != "" {
		q += ` WHERE master_plan_id=?`
		args = []any{mid}
	}
	q += ` ORDER BY created_at,id`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SpecificPlan
	for rows.Next() {
		var x domain.SpecificPlan
		var s, c, u string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.MasterPlanID, &x.Title, &x.Summary, &s, &x.Version, &c, &u); err != nil {
			return nil, err
		}
		x.ParentPlanID = x.MasterPlanID
		x.Status = domain.Status(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r PlanRepository) UpdatePlanStatus(id string, status domain.Status) error {
	if err := updateStatus(r.db, "plans", id, status); err != nil {
		return err
	}
	_ = updateStatus(r.db, "master_plans", id, status)
	_ = updateStatus(r.db, "specific_plans", id, status)
	return nil
}
func (r PlanRepository) Delete(id string) error {
	_ = deleteByID(r.db, "plans", id)
	_ = deleteByID(r.db, "master_plans", id)
	return deleteByID(r.db, "specific_plans", id)
}

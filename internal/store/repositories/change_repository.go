package repositories

import (
	"database/sql"
	"github.com/plan-ai/plan-ai/internal/domain"
)

type ChangeRepository struct{ db *sql.DB }

func NewChangeRepository(db *sql.DB) ChangeRepository { return ChangeRepository{db: db} }

var _ domain.ChangeRepository = ChangeRepository{}

func (r ChangeRepository) SaveChangeRequest(x domain.ChangeRequest) error {
	x.ID = ensureID(x.ID, "change")
	if x.Status == "" {
		x.Status = domain.ChangeRequestDraft
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO change_requests (id, project_id, reason, description, status, requester, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id,reason=excluded.reason,description=excluded.description,status=excluded.status,requester=excluded.requester,updated_at=excluded.updated_at`, x.ID, x.ProjectID, x.Reason, x.Description, x.Status, x.Requester, c, u)
	return err
}
func (r ChangeRepository) GetChangeRequest(id string) (domain.ChangeRequest, error) {
	xs, err := r.ListChangeRequests("")
	if err != nil {
		return domain.ChangeRequest{}, err
	}
	for _, x := range xs {
		if x.ID == id {
			return x, nil
		}
	}
	return domain.ChangeRequest{}, sql.ErrNoRows
}
func (r ChangeRepository) ListChangeRequests(pid string) ([]domain.ChangeRequest, error) {
	q := `SELECT id,project_id,reason,description,status,requester,created_at,updated_at FROM change_requests`
	var args []any
	if pid != "" {
		q += ` WHERE project_id=?`
		args = []any{pid}
	}
	q += ` ORDER BY created_at,id`
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ChangeRequest
	for rows.Next() {
		var x domain.ChangeRequest
		var s, c, u string
		if err := rows.Scan(&x.ID, &x.ProjectID, &x.Reason, &x.Description, &s, &x.Requester, &c, &u); err != nil {
			return nil, err
		}
		x.Status = domain.ChangeRequestStatus(s)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r ChangeRepository) UpdateChangeRequestStatus(id string, status domain.ChangeRequestStatus) error {
	return updateStatus(r.db, "change_requests", id, status)
}
func (r ChangeRepository) SaveImpactReport(x domain.ImpactReport) error {
	x.ID = ensureID(x.ID, "impact")
	c, _ := times(x.CreatedAt, x.CreatedAt)
	_, err := r.db.Exec(`INSERT INTO impact_reports (id,change_request_id,affected_plans,affected_phases,affected_tasks,affected_decisions,affected_entities,summary,created_at) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET affected_plans=excluded.affected_plans,affected_phases=excluded.affected_phases,affected_tasks=excluded.affected_tasks,affected_decisions=excluded.affected_decisions,affected_entities=excluded.affected_entities,summary=excluded.summary`, x.ID, x.ChangeRequestID, jsonList(x.AffectedPlans), jsonList(x.AffectedPhases), jsonList(x.AffectedTasks), jsonList(x.AffectedDecisions), jsonList(x.AffectedEntities), x.Summary, c)
	return err
}
func (r ChangeRepository) GetImpactReportByChangeRequest(id string) (domain.ImpactReport, error) {
	var x domain.ImpactReport
	var plans, phases, tasks, decs, ents, c string
	err := r.db.QueryRow(`SELECT id,change_request_id,affected_plans,affected_phases,affected_tasks,affected_decisions,affected_entities,summary,created_at FROM impact_reports WHERE change_request_id=? ORDER BY created_at DESC LIMIT 1`, id).Scan(&x.ID, &x.ChangeRequestID, &plans, &phases, &tasks, &decs, &ents, &x.Summary, &c)
	x.AffectedPlans = scanJSONList(plans)
	x.AffectedPhases = scanJSONList(phases)
	x.AffectedTasks = scanJSONList(tasks)
	x.AffectedDecisions = scanJSONList(decs)
	x.AffectedEntities = scanJSONList(ents)
	x.CreatedAt = parse(c)
	return x, err
}
func (r ChangeRepository) DeleteChangeRequest(id string) error {
	return deleteByID(r.db, "change_requests", id)
}

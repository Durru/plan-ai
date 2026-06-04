package store

import (
	"database/sql"

	"github.com/plan-ai/plan-ai/internal/change"
	"github.com/plan-ai/plan-ai/internal/continuous"
)

type DeepImpactRepository struct{ db *sql.DB }
type TargetedRegenerationRepository struct{ db *sql.DB }

func NewDeepImpactRepository(db *sql.DB) DeepImpactRepository { return DeepImpactRepository{db: db} }
func NewTargetedRegenerationRepository(db *sql.DB) TargetedRegenerationRepository {
	return TargetedRegenerationRepository{db: db}
}

var _ change.DeepImpactRepository = DeepImpactRepository{}
var _ continuous.TargetedRegenerationRepository = TargetedRegenerationRepository{}

func (r DeepImpactRepository) SaveDeepImpact(report change.DeepImpactReport) (change.DeepImpactReport, error) {
	c, u := timestamps(report.CreatedAt, report.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO change_impact_reports_v2 (id, project_id, change_type, summary, architecture_concerns, backend_concerns, migration_concerns, docs_concerns, api_concerns, plan_concerns, validation_commands, rollback_strategy, affected_plans, affected_tasks, severity, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET summary=excluded.summary, architecture_concerns=excluded.architecture_concerns, backend_concerns=excluded.backend_concerns, migration_concerns=excluded.migration_concerns, docs_concerns=excluded.docs_concerns, api_concerns=excluded.api_concerns, plan_concerns=excluded.plan_concerns, validation_commands=excluded.validation_commands, rollback_strategy=excluded.rollback_strategy, affected_plans=excluded.affected_plans, affected_tasks=excluded.affected_tasks, severity=excluded.severity, status=excluded.status, updated_at=excluded.updated_at`, report.ID, report.ProjectID, report.ChangeType, report.Summary, mustJSON(report.ArchitectureConcerns), mustJSON(report.BackendConcerns), mustJSON(report.MigrationConcerns), mustJSON(report.DocsConcerns), mustJSON(report.APIConcerns), mustJSON(report.PlanConcerns), mustJSON(report.ValidationCommands), mustJSON(report.RollbackStrategy), mustJSON(report.AffectedPlans), mustJSON(report.AffectedTasks), report.Severity, report.Status, c, u)
	if err != nil {
		return change.DeepImpactReport{}, err
	}
	report.CreatedAt, report.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return report, nil
}

func (r DeepImpactRepository) ListDeepImpacts(projectID string) ([]change.DeepImpactReport, error) {
	rows, err := r.db.Query(`SELECT id, project_id, change_type, summary, architecture_concerns, backend_concerns, migration_concerns, docs_concerns, api_concerns, plan_concerns, validation_commands, rollback_strategy, affected_plans, affected_tasks, severity, status, created_at, updated_at FROM change_impact_reports_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []change.DeepImpactReport
	for rows.Next() {
		var report change.DeepImpactReport
		var arch, backend, migration, docs, api, plan, validations, rollback, plans, tasks, c, u string
		if err := rows.Scan(&report.ID, &report.ProjectID, &report.ChangeType, &report.Summary, &arch, &backend, &migration, &docs, &api, &plan, &validations, &rollback, &plans, &tasks, &report.Severity, &report.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(arch, &report.ArchitectureConcerns)
		decodeJSON(backend, &report.BackendConcerns)
		decodeJSON(migration, &report.MigrationConcerns)
		decodeJSON(docs, &report.DocsConcerns)
		decodeJSON(api, &report.APIConcerns)
		decodeJSON(plan, &report.PlanConcerns)
		decodeJSON(validations, &report.ValidationCommands)
		decodeJSON(rollback, &report.RollbackStrategy)
		decodeJSON(plans, &report.AffectedPlans)
		decodeJSON(tasks, &report.AffectedTasks)
		report.CreatedAt, report.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, report)
	}
	return out, rows.Err()
}

func (r TargetedRegenerationRepository) SaveRegeneration(regen continuous.TargetedRegeneration) (continuous.TargetedRegeneration, error) {
	c, u := timestamps(regen.CreatedAt, regen.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO continuous_regenerations_v2 (id, project_id, reason, scope, affected_sections, preserved_sections, snapshot_required, approval_required, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET reason=excluded.reason, scope=excluded.scope, affected_sections=excluded.affected_sections, preserved_sections=excluded.preserved_sections, snapshot_required=excluded.snapshot_required, approval_required=excluded.approval_required, status=excluded.status, updated_at=excluded.updated_at`, regen.ID, regen.ProjectID, regen.Reason, regen.Scope, mustJSON(regen.AffectedSections), mustJSON(regen.PreservedSections), boolInt(regen.SnapshotRequired), boolInt(regen.ApprovalRequired), regen.Status, c, u)
	if err != nil {
		return continuous.TargetedRegeneration{}, err
	}
	regen.CreatedAt, regen.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return regen, nil
}

func (r TargetedRegenerationRepository) ListRegenerations(projectID string) ([]continuous.TargetedRegeneration, error) {
	rows, err := r.db.Query(`SELECT id, project_id, reason, scope, affected_sections, preserved_sections, snapshot_required, approval_required, status, created_at, updated_at FROM continuous_regenerations_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []continuous.TargetedRegeneration
	for rows.Next() {
		var regen continuous.TargetedRegeneration
		var affected, preserved, c, u string
		var snap, approval int
		if err := rows.Scan(&regen.ID, &regen.ProjectID, &regen.Reason, &regen.Scope, &affected, &preserved, &snap, &approval, &regen.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(affected, &regen.AffectedSections)
		decodeJSON(preserved, &regen.PreservedSections)
		regen.SnapshotRequired = snap == 1
		regen.ApprovalRequired = approval == 1
		regen.CreatedAt, regen.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, regen)
	}
	return out, rows.Err()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

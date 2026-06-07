package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type DomainCounts struct {
	Plans            int
	Phases           int
	Tasks            int
	Decisions        int
	ResearchEntries  int
	KnowledgeObjects int
	Validations      int
	Snapshots        int
}

type PlanRepository struct{ db *sql.DB }
type PhaseRepository struct{ db *sql.DB }
type TaskRepository struct{ db *sql.DB }
type DecisionRepository struct{ db *sql.DB }
type ValidationRepository struct{ db *sql.DB }
type SnapshotRepository struct{ db *sql.DB }

func NewPlanRepository(db *sql.DB) PlanRepository             { return PlanRepository{db: db} }
func NewPhaseRepository(db *sql.DB) PhaseRepository           { return PhaseRepository{db: db} }
func NewTaskRepository(db *sql.DB) TaskRepository             { return TaskRepository{db: db} }
func NewDecisionRepository(db *sql.DB) DecisionRepository     { return DecisionRepository{db: db} }
func NewValidationRepository(db *sql.DB) ValidationRepository { return ValidationRepository{db: db} }
func NewSnapshotRepository(db *sql.DB) SnapshotRepository     { return SnapshotRepository{db: db} }

func (r PlanRepository) CreateMaster(plan domain.MasterPlan) error {
	return r.create(domain.Plan{
		ID: plan.ID, Type: domain.PlanTypeMaster, Title: plan.Title, Summary: plan.Summary,
		Status: plan.Status, Version: plan.Version, CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt,
	})
}

func (r PlanRepository) CreateSpecific(plan domain.SpecificPlan) error {
	return r.create(domain.Plan{
		ID: plan.ID, Type: domain.PlanTypeSpecific, Title: plan.Title, Summary: plan.Summary,
		Status: plan.Status, Version: plan.Version, ParentPlanID: plan.ParentPlanID,
		CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt,
	})
}

func (r PlanRepository) create(plan domain.Plan) error {
	plan.ID = ensureID(plan.ID, "plan")
	plan.Status = ensureStatus(plan.Status)
	if plan.Version == 0 {
		plan.Version = 1
	}
	createdAt, updatedAt := ensureTimestamps(plan.CreatedAt, plan.UpdatedAt)
	var parent any
	if plan.ParentPlanID != "" {
		parent = plan.ParentPlanID
	}
	_, err := r.db.Exec(`INSERT INTO plans (id, type, title, summary, status, version, parent_plan_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, plan.ID, plan.Type, plan.Title, plan.Summary, plan.Status, plan.Version, parent, createdAt, updatedAt)
	return err
}

func (r PlanRepository) GetByID(id string) (domain.Plan, error) {
	var plan domain.Plan
	var parent sql.NullString
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, type, title, summary, status, version, parent_plan_id, created_at, updated_at FROM plans WHERE id = ?`, id).
		Scan(&plan.ID, &plan.Type, &plan.Title, &plan.Summary, &plan.Status, &plan.Version, &parent, &createdAt, &updatedAt)
	if err != nil {
		return plan, err
	}
	if parent.Valid {
		plan.ParentPlanID = parent.String
	}
	plan.CreatedAt = parseTime(createdAt)
	plan.UpdatedAt = parseTime(updatedAt)
	return plan, nil
}

func (r PlanRepository) List() ([]domain.Plan, error) {
	rows, err := r.db.Query(`SELECT id, type, title, summary, status, version, parent_plan_id, created_at, updated_at FROM plans ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []domain.Plan
	for rows.Next() {
		var plan domain.Plan
		var parent sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&plan.ID, &plan.Type, &plan.Title, &plan.Summary, &plan.Status, &plan.Version, &parent, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if parent.Valid {
			plan.ParentPlanID = parent.String
		}
		plan.CreatedAt = parseTime(createdAt)
		plan.UpdatedAt = parseTime(updatedAt)
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r PlanRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "plans", id, status)
}

func (r PhaseRepository) Create(phase domain.Phase) error {
	phase.ID = ensureID(phase.ID, "phase")
	phase.Status = ensureStatus(phase.Status)
	createdAt, updatedAt := ensureTimestamps(phase.CreatedAt, phase.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO phases (id, plan_id, title, summary, status, position, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, phase.ID, phase.PlanID, phase.Title, phase.Summary, phase.Status, phase.Position, createdAt, updatedAt)
	return err
}

func (r PhaseRepository) GetByID(id string) (domain.Phase, error) {
	var phase domain.Phase
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, plan_id, title, summary, status, position, created_at, updated_at FROM phases WHERE id = ?`, id).
		Scan(&phase.ID, &phase.PlanID, &phase.Title, &phase.Summary, &phase.Status, &phase.Position, &createdAt, &updatedAt)
	phase.CreatedAt = parseTime(createdAt)
	phase.UpdatedAt = parseTime(updatedAt)
	return phase, err
}

func (r PhaseRepository) List() ([]domain.Phase, error) {
	rows, err := r.db.Query(`SELECT id, plan_id, title, summary, status, position, created_at, updated_at FROM phases ORDER BY position, created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var phases []domain.Phase
	for rows.Next() {
		var phase domain.Phase
		var createdAt, updatedAt string
		if err := rows.Scan(&phase.ID, &phase.PlanID, &phase.Title, &phase.Summary, &phase.Status, &phase.Position, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		phase.CreatedAt = parseTime(createdAt)
		phase.UpdatedAt = parseTime(updatedAt)
		phases = append(phases, phase)
	}
	return phases, rows.Err()
}

func (r PhaseRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "phases", id, status)
}

func (r TaskRepository) Create(task domain.Task) error {
	task.ID = ensureID(task.ID, "task")
	task.Status = ensureStatus(task.Status)
	if task.ContextSize == "" {
		task.ContextSize = domain.ContextSizeShort
	}
	createdAt, updatedAt := ensureTimestamps(task.CreatedAt, task.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO tasks (id, phase_id, plan_id, title, summary, status, position, context_size, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, task.ID, task.PhaseID, task.PlanID, task.Title, task.Summary, task.Status, task.Position, task.ContextSize, createdAt, updatedAt)
	return err
}

func (r TaskRepository) GetByID(id string) (domain.Task, error) {
	var task domain.Task
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, phase_id, plan_id, title, summary, status, position, context_size, created_at, updated_at FROM tasks WHERE id = ?`, id).
		Scan(&task.ID, &task.PhaseID, &task.PlanID, &task.Title, &task.Summary, &task.Status, &task.Position, &task.ContextSize, &createdAt, &updatedAt)
	task.CreatedAt = parseTime(createdAt)
	task.UpdatedAt = parseTime(updatedAt)
	return task, err
}

func (r TaskRepository) List() ([]domain.Task, error) {
	rows, err := r.db.Query(`SELECT id, phase_id, plan_id, title, summary, status, position, context_size, created_at, updated_at FROM tasks ORDER BY position, created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var createdAt, updatedAt string
		if err := rows.Scan(&task.ID, &task.PhaseID, &task.PlanID, &task.Title, &task.Summary, &task.Status, &task.Position, &task.ContextSize, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		task.CreatedAt = parseTime(createdAt)
		task.UpdatedAt = parseTime(updatedAt)
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r TaskRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "tasks", id, status)
}

func (r DecisionRepository) Create(decision domain.Decision) error {
	decision.ID = ensureID(decision.ID, "decision")
	decision.Status = ensureStatus(decision.Status)
	createdAt, updatedAt := ensureTimestamps(decision.CreatedAt, decision.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO decisions (id, title, context, decision, status, impact, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, decision.ID, decision.Title, decision.Context, decision.Decision, decision.Status, decision.Impact, createdAt, updatedAt)
	return err
}

func (r DecisionRepository) GetByID(id string) (domain.Decision, error) {
	var decision domain.Decision
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, title, context, decision, status, impact, created_at, updated_at FROM decisions WHERE id = ?`, id).
		Scan(&decision.ID, &decision.Title, &decision.Context, &decision.Decision, &decision.Status, &decision.Impact, &createdAt, &updatedAt)
	decision.CreatedAt = parseTime(createdAt)
	decision.UpdatedAt = parseTime(updatedAt)
	return decision, err
}

func (r DecisionRepository) List() ([]domain.Decision, error) {
	rows, err := r.db.Query(`SELECT id, title, context, decision, status, impact, created_at, updated_at FROM decisions ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var decisions []domain.Decision
	for rows.Next() {
		var decision domain.Decision
		var createdAt, updatedAt string
		if err := rows.Scan(&decision.ID, &decision.Title, &decision.Context, &decision.Decision, &decision.Status, &decision.Impact, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		decision.CreatedAt = parseTime(createdAt)
		decision.UpdatedAt = parseTime(updatedAt)
		decisions = append(decisions, decision)
	}
	return decisions, rows.Err()
}

func (r DecisionRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "decisions", id, status)
}

func (r ValidationRepository) Create(validation domain.Validation) error {
	validation.ID = ensureID(validation.ID, "validation")
	validation.Status = ensureStatus(validation.Status)
	createdAt, updatedAt := ensureTimestamps(validation.CreatedAt, validation.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO validations (id, target_type, target_id, status, summary, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`, validation.ID, validation.TargetType, validation.TargetID, validation.Status, validation.Summary, createdAt, updatedAt)
	return err
}

func (r ValidationRepository) GetByID(id string) (domain.Validation, error) {
	var validation domain.Validation
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, target_type, target_id, status, summary, created_at, updated_at FROM validations WHERE id = ?`, id).
		Scan(&validation.ID, &validation.TargetType, &validation.TargetID, &validation.Status, &validation.Summary, &createdAt, &updatedAt)
	validation.CreatedAt = parseTime(createdAt)
	validation.UpdatedAt = parseTime(updatedAt)
	return validation, err
}

func (r ValidationRepository) List() ([]domain.Validation, error) {
	rows, err := r.db.Query(`SELECT id, target_type, target_id, status, summary, created_at, updated_at FROM validations ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var validations []domain.Validation
	for rows.Next() {
		var validation domain.Validation
		var createdAt, updatedAt string
		if err := rows.Scan(&validation.ID, &validation.TargetType, &validation.TargetID, &validation.Status, &validation.Summary, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		validation.CreatedAt = parseTime(createdAt)
		validation.UpdatedAt = parseTime(updatedAt)
		validations = append(validations, validation)
	}
	return validations, rows.Err()
}

func (r ValidationRepository) UpdateStatus(id string, status domain.Status) error {
	return updateStatus(r.db, "validations", id, status)
}

func (r SnapshotRepository) Create(snapshot domain.Snapshot) error {
	snapshot.ID = ensureID(snapshot.ID, "snapshot")
	createdAt, _ := ensureTimestamps(snapshot.CreatedAt, snapshot.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO snapshots (id, reason, summary, created_at)
VALUES (?, ?, ?, ?)`, snapshot.ID, snapshot.Reason, snapshot.Summary, createdAt)
	return err
}

func (r SnapshotRepository) GetByID(id string) (domain.Snapshot, error) {
	var snapshot domain.Snapshot
	var createdAt string
	err := r.db.QueryRow(`SELECT id, reason, summary, created_at FROM snapshots WHERE id = ?`, id).
		Scan(&snapshot.ID, &snapshot.Reason, &snapshot.Summary, &createdAt)
	snapshot.CreatedAt = parseTime(createdAt)
	snapshot.UpdatedAt = snapshot.CreatedAt
	return snapshot, err
}

func (r SnapshotRepository) List() ([]domain.Snapshot, error) {
	rows, err := r.db.Query(`SELECT id, reason, summary, created_at FROM snapshots ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []domain.Snapshot
	for rows.Next() {
		var snapshot domain.Snapshot
		var createdAt string
		if err := rows.Scan(&snapshot.ID, &snapshot.Reason, &snapshot.Summary, &createdAt); err != nil {
			return nil, err
		}
		snapshot.CreatedAt = parseTime(createdAt)
		snapshot.UpdatedAt = snapshot.CreatedAt
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, rows.Err()
}

func CountDomainEntities(db *sql.DB) (DomainCounts, error) {
	var counts DomainCounts
	tables := []struct {
		name string
		dest *int
	}{
		{name: "plans", dest: &counts.Plans},
		{name: "phases", dest: &counts.Phases},
		{name: "tasks", dest: &counts.Tasks},
		{name: "decisions", dest: &counts.Decisions},
		{name: "research_entries", dest: &counts.ResearchEntries},
		{name: "knowledge_objects", dest: &counts.KnowledgeObjects},
		{name: "validations", dest: &counts.Validations},
		{name: "snapshots", dest: &counts.Snapshots},
	}
	for _, table := range tables {
		if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table.name)).Scan(table.dest); err != nil {
			return counts, err
		}
	}
	return counts, nil
}

func updateStatus(db *sql.DB, table, id string, status domain.Status) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET status = ?, updated_at = ? WHERE id = ?", table), status, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func ensureID(id, prefix string) string {
	if id != "" {
		return id
	}
	return domain.NewID(prefix)
}

func ensureStatus(status domain.Status) domain.Status {
	if status == "" {
		return domain.StatusDraft
	}
	return status
}

func ensureTimestamps(createdAt, updatedAt time.Time) (string, string) {
	now := time.Now().UTC()
	if createdAt.IsZero() {
		createdAt = now
	}
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	return createdAt.UTC().Format(time.RFC3339), updatedAt.UTC().Format(time.RFC3339)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

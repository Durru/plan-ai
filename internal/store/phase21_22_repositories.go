package store

import (
	"database/sql"
	"time"
)

// ──────────────────────────────────────────────
// Phase 21: Agent System Repositories
// ──────────────────────────────────────────────

// AgentRunV2Record represents a persisted agent run (Phase 21).
type AgentRunV2Record struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Intent    string `json:"intent"`
	Status    string `json:"status"`
	Response  string `json:"response"` // JSON
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AgentMessageRecord represents a message in an agent conversation.
type AgentMessageRecord struct {
	ID        string `json:"id"`
	RunID     string `json:"run_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// AgentDelegatedJobRecord represents a delegated sub-agent job.
type AgentDelegatedJobRecord struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id"`
	Intent        string `json:"intent"`
	Capability    string `json:"capability"`
	WorkflowType  string `json:"workflow_type"`
	JobType       string `json:"job_type"`
	Status        string `json:"status"`
	ResultSummary string `json:"result_summary"`
	CreatedAt     string `json:"created_at"`
	CompletedAt   string `json:"completed_at"`
}

// AgentRunV2Repository persists agent run records.
type AgentRunV2Repository struct{ db *sql.DB }

// NewAgentRunV2Repository creates a new AgentRunV2Repository.
func NewAgentRunV2Repository(db *sql.DB) *AgentRunV2Repository {
	return &AgentRunV2Repository{db: db}
}

func (r *AgentRunV2Repository) CreateRun(record AgentRunV2Record) (AgentRunV2Record, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return record, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO agent_runs_v2 (id, project_id, intent, status, response, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.ProjectID, record.Intent, record.Status, record.Response, now, now); err != nil {
		return record, err
	}
	if _, err := tx.Exec(`INSERT INTO agent_runs (id, project_id, agent_name, status, metadata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.ProjectID, record.Intent, record.Status, record.Response, now, now); err != nil {
		return record, err
	}
	err = tx.Commit()
	if err != nil {
		return record, err
	}
	return r.GetRun(record.ID)
}

func (r *AgentRunV2Repository) GetRun(id string) (AgentRunV2Record, error) {
	var rec AgentRunV2Record
	err := r.db.QueryRow(`SELECT id, project_id, intent, status, response, created_at, updated_at FROM agent_runs_v2 WHERE id = ?`, id).
		Scan(&rec.ID, &rec.ProjectID, &rec.Intent, &rec.Status, &rec.Response, &rec.CreatedAt, &rec.UpdatedAt)
	return rec, err
}

func (r *AgentRunV2Repository) UpdateRunStatus(id, status, response string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE agent_runs_v2 SET status = ?, response = ?, updated_at = ? WHERE id = ?`, status, response, now, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE agent_runs SET status = ?, metadata = ?, updated_at = ? WHERE id = ?`, status, response, now, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AgentRunV2Repository) ListRuns(projectID string, limit int) ([]AgentRunV2Record, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, intent, status, response, created_at, updated_at FROM agent_runs_v2 WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []AgentRunV2Record
	for rows.Next() {
		var rec AgentRunV2Record
		if err := rows.Scan(&rec.ID, &rec.ProjectID, &rec.Intent, &rec.Status, &rec.Response, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *AgentRunV2Repository) CreateMessage(msg AgentMessageRecord) (AgentMessageRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO agent_messages (id, run_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.RunID, msg.Role, msg.Content, now)
	if err != nil {
		return msg, err
	}
	msg.CreatedAt = now
	return msg, nil
}

func (r *AgentRunV2Repository) ListMessages(runID string) ([]AgentMessageRecord, error) {
	rows, err := r.db.Query(`SELECT id, run_id, role, content, created_at FROM agent_messages WHERE run_id = ? ORDER BY created_at`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []AgentMessageRecord
	for rows.Next() {
		var m AgentMessageRecord
		if err := rows.Scan(&m.ID, &m.RunID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// DelegatedJobRepository persists delegated jobs.
type DelegatedJobRepository struct{ db *sql.DB }

// NewDelegatedJobRepository creates a new DelegatedJobRepository.
func NewDelegatedJobRepository(db *sql.DB) *DelegatedJobRepository {
	return &DelegatedJobRepository{db: db}
}

func (r *DelegatedJobRepository) CreateJob(job AgentDelegatedJobRecord) (AgentDelegatedJobRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO agent_delegated_jobs (id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.ProjectID, job.Intent, job.Capability, job.WorkflowType, job.JobType, job.Status, job.ResultSummary, now, job.CompletedAt)
	if err != nil {
		return job, err
	}
	job.CreatedAt = now
	return job, nil
}

func (r *DelegatedJobRepository) GetJob(id string) (AgentDelegatedJobRecord, error) {
	var rec AgentDelegatedJobRecord
	err := r.db.QueryRow(`SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at FROM agent_delegated_jobs WHERE id = ?`, id).
		Scan(&rec.ID, &rec.ProjectID, &rec.Intent, &rec.Capability, &rec.WorkflowType, &rec.JobType, &rec.Status, &rec.ResultSummary, &rec.CreatedAt, &rec.CompletedAt)
	return rec, err
}

func (r *DelegatedJobRepository) ListJobs(projectID string) ([]AgentDelegatedJobRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at FROM agent_delegated_jobs WHERE project_id = ? ORDER BY created_at DESC LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []AgentDelegatedJobRecord
	for rows.Next() {
		var rec AgentDelegatedJobRecord
		if err := rows.Scan(&rec.ID, &rec.ProjectID, &rec.Intent, &rec.Capability, &rec.WorkflowType, &rec.JobType, &rec.Status, &rec.ResultSummary, &rec.CreatedAt, &rec.CompletedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, rec)
	}
	return jobs, rows.Err()
}

func (r *DelegatedJobRepository) UpdateJob(id, status, summary string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE agent_delegated_jobs SET status = ?, result_summary = ?, completed_at = ? WHERE id = ?`, status, summary, now, id)
	return err
}

// ──────────────────────────────────────────────
// Phase 22: Continuous Planning Repositories
// ──────────────────────────────────────────────

// ContinuousEventRecord represents a detected continuous event.
type ContinuousEventRecord struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	EventType string `json:"event_type"`
	Summary   string `json:"summary"`
	Details   string `json:"details"` // JSON
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
}

// PlanUpdateProposalRecord represents a plan update proposal.
type PlanUpdateProposalRecord struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	Reason            string `json:"reason"`
	AffectedPlans     string `json:"affected_plans"`     // JSON array
	AffectedTasks     string `json:"affected_tasks"`     // JSON array
	AffectedDecisions string `json:"affected_decisions"` // JSON array
	SuggestedUpdates  string `json:"suggested_updates"`
	RequiresResearch  int    `json:"requires_research"`
	RequiresApproval  int    `json:"requires_approval"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ContextDeliveryRecord represents a delivered context response.
type ContextDeliveryRecord struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Level     string `json:"level"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ContinuousEventRepository persists continuous events.
type ContinuousEventRepository struct{ db *sql.DB }

// NewContinuousEventRepository creates a new ContinuousEventRepository.
func NewContinuousEventRepository(db *sql.DB) *ContinuousEventRepository {
	return &ContinuousEventRepository{db: db}
}

func (r *ContinuousEventRepository) CreateEvent(ev ContinuousEventRecord) (ContinuousEventRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO continuous_events (id, project_id, event_type, summary, details, source, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ev.ID, ev.ProjectID, ev.EventType, ev.Summary, ev.Details, ev.Source, now)
	if err != nil {
		return ev, err
	}
	ev.CreatedAt = now
	return ev, nil
}

func (r *ContinuousEventRepository) ListEvents(projectID string, limit int) ([]ContinuousEventRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, event_type, summary, details, source, created_at FROM continuous_events WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []ContinuousEventRecord
	for rows.Next() {
		var ev ContinuousEventRecord
		if err := rows.Scan(&ev.ID, &ev.ProjectID, &ev.EventType, &ev.Summary, &ev.Details, &ev.Source, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// PlanUpdateProposalRepository persists plan update proposals.
type PlanUpdateProposalRepository struct{ db *sql.DB }

// NewPlanUpdateProposalRepository creates a new PlanUpdateProposalRepository.
func NewPlanUpdateProposalRepository(db *sql.DB) *PlanUpdateProposalRepository {
	return &PlanUpdateProposalRepository{db: db}
}

func (r *PlanUpdateProposalRepository) CreateProposal(p PlanUpdateProposalRecord) (PlanUpdateProposalRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO plan_update_proposals (id, project_id, reason, affected_plans, affected_tasks, affected_decisions, suggested_updates, requires_research, requires_approval, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.ProjectID, p.Reason, p.AffectedPlans, p.AffectedTasks, p.AffectedDecisions, p.SuggestedUpdates, p.RequiresResearch, p.RequiresApproval, p.Status, now, now)
	if err != nil {
		return p, err
	}
	return r.GetProposal(p.ID)
}

func (r *PlanUpdateProposalRepository) GetProposal(id string) (PlanUpdateProposalRecord, error) {
	var p PlanUpdateProposalRecord
	err := r.db.QueryRow(`SELECT id, project_id, reason, affected_plans, affected_tasks, affected_decisions, suggested_updates, requires_research, requires_approval, status, created_at, updated_at FROM plan_update_proposals WHERE id = ?`, id).
		Scan(&p.ID, &p.ProjectID, &p.Reason, &p.AffectedPlans, &p.AffectedTasks, &p.AffectedDecisions, &p.SuggestedUpdates, &p.RequiresResearch, &p.RequiresApproval, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *PlanUpdateProposalRepository) ListProposals(projectID string) ([]PlanUpdateProposalRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, reason, affected_plans, affected_tasks, affected_decisions, suggested_updates, requires_research, requires_approval, status, created_at, updated_at FROM plan_update_proposals WHERE project_id = ? ORDER BY created_at DESC LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var proposals []PlanUpdateProposalRecord
	for rows.Next() {
		var p PlanUpdateProposalRecord
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Reason, &p.AffectedPlans, &p.AffectedTasks, &p.AffectedDecisions, &p.SuggestedUpdates, &p.RequiresResearch, &p.RequiresApproval, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		proposals = append(proposals, p)
	}
	return proposals, rows.Err()
}

func (r *PlanUpdateProposalRepository) UpdateProposalStatus(id string, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE plan_update_proposals SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	return err
}

// ContextDeliveryRepository persists context deliveries.
type ContextDeliveryRepository struct{ db *sql.DB }

// NewContextDeliveryRepository creates a new ContextDeliveryRepository.
func NewContextDeliveryRepository(db *sql.DB) *ContextDeliveryRepository {
	return &ContextDeliveryRepository{db: db}
}

func (r *ContextDeliveryRepository) CreateDelivery(d ContextDeliveryRecord) (ContextDeliveryRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO context_deliveries (id, project_id, level, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		d.ID, d.ProjectID, d.Level, d.Content, now)
	if err != nil {
		return d, err
	}
	d.CreatedAt = now
	return d, nil
}

func (r *ContextDeliveryRepository) ListDeliveries(projectID string, level string, limit int) ([]ContextDeliveryRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, level, content, created_at FROM context_deliveries WHERE project_id = ? AND level = ? ORDER BY created_at DESC LIMIT ?`, projectID, level, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deliveries []ContextDeliveryRecord
	for rows.Next() {
		var d ContextDeliveryRecord
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Level, &d.Content, &d.CreatedAt); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// ContinuousStatusRepository provides access to continuous_status snapshots.
type ContinuousStatusRepository struct{ db *sql.DB }

// NewContinuousStatusRepository creates a new repository.
func NewContinuousStatusRepository(db *sql.DB) *ContinuousStatusRepository {
	return &ContinuousStatusRepository{db: db}
}

// ContinuousStatusRecord mirrors the continuous_status table row.
type ContinuousStatusRecord struct {
	ID             string `json:"id"`
	ProjectID      string `json:"project_id"`
	ActivePlan     string `json:"active_plan"`
	ActivePhase    string `json:"active_phase"`
	NextTask       string `json:"next_task"`
	BlockedItems   string `json:"blocked_items"`
	ApprovalsNeeded string `json:"approvals_needed"`
	OutdatedPlans  string `json:"outdated_plans"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// GetLatest returns the most recent status snapshot for a project.
func (r *ContinuousStatusRepository) GetLatest(projectID string) (ContinuousStatusRecord, error) {
	var rec ContinuousStatusRecord
	err := r.db.QueryRow(
		`SELECT id, project_id, active_plan, active_phase, next_task, COALESCE(blocked_items, '[]'), COALESCE(approvals_needed, '[]'), COALESCE(outdated_plans, '[]'), created_at, updated_at FROM continuous_status WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`,
		projectID,
	).Scan(&rec.ID, &rec.ProjectID, &rec.ActivePlan, &rec.ActivePhase, &rec.NextTask, &rec.BlockedItems, &rec.ApprovalsNeeded, &rec.OutdatedPlans, &rec.CreatedAt, &rec.UpdatedAt)
	return rec, err
}

// Save creates or updates a status snapshot.
func (r *ContinuousStatusRepository) Save(rec ContinuousStatusRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO continuous_status (id, project_id, active_plan, active_phase, next_task, blocked_items, approvals_needed, outdated_plans, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		rec.ID, rec.ProjectID, rec.ActivePlan, rec.ActivePhase, rec.NextTask, rec.BlockedItems, rec.ApprovalsNeeded, rec.OutdatedPlans, now, now)
	return err
}

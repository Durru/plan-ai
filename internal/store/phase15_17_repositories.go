package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/modelstrategy"
	"github.com/Durru/plan-ai/internal/orchestrator"
)

// ──────────────────────────────────────────────
// Model Strategy Repositories
// ──────────────────────────────────────────────

type ModelProfileRepository struct{ db *sql.DB }

func NewModelProfileRepository(db *sql.DB) *ModelProfileRepository {
	return &ModelProfileRepository{db: db}
}

func (r *ModelProfileRepository) Create(profile modelstrategy.ModelProfile) (modelstrategy.ModelProfile, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO model_profiles (id, name, provider, model, config, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		profile.ID, profile.Name, profile.Provider, profile.Model, profile.Config, now, now)
	if err != nil {
		return profile, err
	}
	return r.Get(profile.ID)
}

func (r *ModelProfileRepository) Get(id string) (modelstrategy.ModelProfile, error) {
	var p modelstrategy.ModelProfile
	var createdAt, updatedAt string
	err := r.db.QueryRow(`SELECT id, name, provider, model, config, created_at, updated_at FROM model_profiles WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Provider, &p.Model, &p.Config, &createdAt, &updatedAt)
	if err != nil {
		return p, err
	}
	p.CreatedAt = parseTime(createdAt)
	p.UpdatedAt = parseTime(updatedAt)
	return p, nil
}

func (r *ModelProfileRepository) List() ([]modelstrategy.ModelProfile, error) {
	rows, err := r.db.Query(`SELECT id, name, provider, model, config, created_at, updated_at FROM model_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []modelstrategy.ModelProfile
	for rows.Next() {
		var p modelstrategy.ModelProfile
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.Name, &p.Provider, &p.Model, &p.Config, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = parseTime(createdAt)
		p.UpdatedAt = parseTime(updatedAt)
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (r *ModelProfileRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM model_profiles WHERE id = ?`, id)
	return err
}

type PromptContractRepository struct{ db *sql.DB }

func NewPromptContractRepository(db *sql.DB) *PromptContractRepository {
	return &PromptContractRepository{db: db}
}

func (r *PromptContractRepository) Create(pc modelstrategy.PromptContract) (modelstrategy.PromptContract, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO prompt_contracts (id, contract_type, content, created_at) VALUES (?, ?, ?, ?)`,
		pc.ID, pc.ContractType, pc.Content, now)
	if err != nil {
		return pc, err
	}
	return pc, nil
}

func (r *PromptContractRepository) ListByType(contractType string) ([]modelstrategy.PromptContract, error) {
	rows, err := r.db.Query(`SELECT id, contract_type, content, created_at FROM prompt_contracts WHERE contract_type = ? ORDER BY created_at`, contractType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []modelstrategy.PromptContract
	for rows.Next() {
		var pc modelstrategy.PromptContract
		var createdAt string
		if err := rows.Scan(&pc.ID, &pc.ContractType, &pc.Content, &createdAt); err != nil {
			return nil, err
		}
		pc.CreatedAt = createdAt
		contracts = append(contracts, pc)
	}
	return contracts, rows.Err()
}

type OutputSchemaRepository struct{ db *sql.DB }

func NewOutputSchemaRepository(db *sql.DB) *OutputSchemaRepository {
	return &OutputSchemaRepository{db: db}
}

func (r *OutputSchemaRepository) Create(schema modelstrategy.OutputSchema) (modelstrategy.OutputSchema, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO output_schemas (id, schema_type, fields, required, created_at) VALUES (?, ?, ?, ?, ?)`,
		schema.ID, schema.SchemaType, schema.Fields, schema.Required, now)
	if err != nil {
		return schema, err
	}
	return schema, nil
}

func (r *OutputSchemaRepository) ListByType(schemaType string) ([]modelstrategy.OutputSchema, error) {
	rows, err := r.db.Query(`SELECT id, schema_type, fields, required, created_at FROM output_schemas WHERE schema_type = ? ORDER BY created_at`, schemaType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []modelstrategy.OutputSchema
	for rows.Next() {
		var s modelstrategy.OutputSchema
		var createdAt string
		if err := rows.Scan(&s.ID, &s.SchemaType, &s.Fields, &s.Required, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt = createdAt
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

// ──────────────────────────────────────────────
// Orchestrator Repositories
// ──────────────────────────────────────────────

var _ orchestrator.JobRepository = (*JobRepository)(nil)
var _ orchestrator.JobRunRepository = (*JobRunRepository)(nil)

type JobRepository struct{ db *sql.DB }

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) CreateJob(job orchestrator.Job) (orchestrator.Job, error) {
	started := job.StartedAt.UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO jobs (id, project_id, workflow_type, capability, strategy, status, error, started_at, finished_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.ProjectID, string(job.WorkflowType), job.Capability, job.Strategy, string(job.Status), job.Error, started, "")
	if err != nil {
		return job, err
	}
	return r.GetJob(job.ID)
}

func (r *JobRepository) GetJob(id string) (orchestrator.Job, error) {
	var j orchestrator.Job
	var started, finished string
	err := r.db.QueryRow(`SELECT id, project_id, workflow_type, capability, strategy, status, error, started_at, finished_at FROM jobs WHERE id = ?`, id).
		Scan(&j.ID, &j.ProjectID, &j.WorkflowType, &j.Capability, &j.Strategy, &j.Status, &j.Error, &started, &finished)
	if err != nil {
		return j, err
	}
	j.StartedAt = parseTime(started)
	if finished != "" {
		j.FinishedAt = parseTime(finished)
	}
	return j, nil
}

func (r *JobRepository) UpdateJob(job orchestrator.Job) (orchestrator.Job, error) {
	finished := ""
	if !job.FinishedAt.IsZero() {
		finished = job.FinishedAt.UTC().Format(time.RFC3339)
	}
	_, err := r.db.Exec(`UPDATE jobs SET status = ?, error = ?, strategy = ?, finished_at = ? WHERE id = ?`,
		string(job.Status), job.Error, job.Strategy, finished, job.ID)
	if err != nil {
		return job, err
	}
	return r.GetJob(job.ID)
}

func (r *JobRepository) ListJobs(projectID string) ([]orchestrator.Job, error) {
	var rows *sql.Rows
	var err error
	if strings.TrimSpace(projectID) == "" {
		rows, err = r.db.Query(`SELECT id, project_id, workflow_type, capability, strategy, status, error, started_at, finished_at FROM jobs ORDER BY started_at DESC`)
	} else {
		rows, err = r.db.Query(`SELECT id, project_id, workflow_type, capability, strategy, status, error, started_at, finished_at FROM jobs WHERE project_id = ? ORDER BY started_at DESC`, projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []orchestrator.Job
	for rows.Next() {
		var j orchestrator.Job
		var started, finished string
		if err := rows.Scan(&j.ID, &j.ProjectID, &j.WorkflowType, &j.Capability, &j.Strategy, &j.Status, &j.Error, &started, &finished); err != nil {
			return nil, err
		}
		j.StartedAt = parseTime(started)
		if finished != "" {
			j.FinishedAt = parseTime(finished)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

type JobRunRepository struct{ db *sql.DB }

func NewJobRunRepository(db *sql.DB) *JobRunRepository {
	return &JobRunRepository{db: db}
}

func (r *JobRunRepository) CreateJobRun(run orchestrator.JobRun) (orchestrator.JobRun, error) {
	started := run.StartedAt.UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO job_runs (id, job_id, step, status, output, error, started_at, finished_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.JobID, run.Step, string(run.Status), run.Output, run.Error, started, "")
	if err != nil {
		return run, err
	}
	return run, nil
}

func (r *JobRunRepository) UpdateJobRun(run orchestrator.JobRun) (orchestrator.JobRun, error) {
	finished := ""
	if !run.FinishedAt.IsZero() {
		finished = run.FinishedAt.UTC().Format(time.RFC3339)
	}
	_, err := r.db.Exec(`UPDATE job_runs SET status = ?, output = ?, error = ?, finished_at = ? WHERE id = ?`,
		string(run.Status), run.Output, run.Error, finished, run.ID)
	if err != nil {
		return run, err
	}
	return run, nil
}

func (r *JobRunRepository) ListJobRuns(jobID string) ([]orchestrator.JobRun, error) {
	rows, err := r.db.Query(`SELECT id, job_id, step, status, output, error, started_at, finished_at FROM job_runs WHERE job_id = ? ORDER BY started_at`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []orchestrator.JobRun
	for rows.Next() {
		var run orchestrator.JobRun
		var started, finished string
		if err := rows.Scan(&run.ID, &run.JobID, &run.Step, &run.Status, &run.Output, &run.Error, &started, &finished); err != nil {
			return nil, err
		}
		run.StartedAt = parseTime(started)
		if finished != "" {
			run.FinishedAt = parseTime(finished)
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// CapabilityRepository persists capability records.
type CapabilityRepository struct{ db *sql.DB }

func NewCapabilityRepository(db *sql.DB) *CapabilityRepository {
	return &CapabilityRepository{db: db}
}

func (r *CapabilityRepository) Upsert(ctype, name, description string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	id := fmt.Sprintf("cap:%s", ctype)
	_, err := r.db.Exec(`INSERT INTO capabilities (id, type, name, description, created_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(type) DO UPDATE SET name = excluded.name, description = excluded.description`,
		id, ctype, name, description, now)
	return err
}

func (r *CapabilityRepository) List() ([]string, error) {
	rows, err := r.db.Query(`SELECT type FROM capabilities ORDER BY type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

// ──────────────────────────────────────────────
// Context View Repository
// ──────────────────────────────────────────────

type ContextViewRepository struct{ db *sql.DB }

func NewContextViewRepository(db *sql.DB) *ContextViewRepository {
	return &ContextViewRepository{db: db}
}

func (r *ContextViewRepository) Create(view orchestratorContextView) (orchestratorContextView, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO context_views_v2 (id, project_id, name, view_type, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		view.ID, view.ProjectID, view.Name, view.ViewType, view.Content, now, now)
	if err != nil {
		return view, err
	}
	return view, nil
}

func (r *ContextViewRepository) ListByProject(projectID string) ([]orchestratorContextView, error) {
	rows, err := r.db.Query(`SELECT id, project_id, name, view_type, content, created_at, updated_at FROM context_views_v2 WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []orchestratorContextView
	for rows.Next() {
		var v orchestratorContextView
		var createdAt, updatedAt string
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Name, &v.ViewType, &v.Content, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		v.CreatedAt = parseTime(createdAt)
		v.UpdatedAt = parseTime(updatedAt)
		views = append(views, v)
	}
	return views, rows.Err()
}

type orchestratorContextView struct {
	ID        string
	ProjectID string
	Name      string
	ViewType  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Ensure compile-time assertions
var (
	_ domain.Project = domain.Project{}
)

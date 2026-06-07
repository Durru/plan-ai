package store

import (
	"database/sql"

	"github.com/Durru/plan-ai/internal/opencode"
)

type SubagentTaskRepository struct{ db *sql.DB }
type OpenCodeWorkflowRepository struct{ db *sql.DB }

type SubagentTaskRecord struct {
	ID               string
	ProjectID        string
	AgentType        string
	Objective        string
	Capability       string
	Status           string
	Provenance       string
	ValidationStatus string
	Isolated         bool
	Temporary        bool
	MemoryPolicy     string
	ResultSummary    string
	CreatedAt        string
	UpdatedAt        string
}

func NewSubagentTaskRepository(db *sql.DB) SubagentTaskRepository {
	return SubagentTaskRepository{db: db}
}
func NewOpenCodeWorkflowRepository(db *sql.DB) OpenCodeWorkflowRepository {
	return OpenCodeWorkflowRepository{db: db}
}

var _ opencode.WorkflowRepository = OpenCodeWorkflowRepository{}

func (r SubagentTaskRepository) SaveSubagentTaskRecord(task SubagentTaskRecord) (SubagentTaskRecord, error) {
	c, u := timestamps(parseRFC3339(task.CreatedAt), parseRFC3339(task.UpdatedAt))
	_, err := r.db.Exec(`INSERT INTO subagent_tasks_v2 (id, project_id, agent_type, objective, capability, status, provenance, validation_status, isolated, temporary, memory_policy, result_summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET objective=excluded.objective, capability=excluded.capability, status=excluded.status, provenance=excluded.provenance, validation_status=excluded.validation_status, isolated=excluded.isolated, temporary=excluded.temporary, memory_policy=excluded.memory_policy, result_summary=excluded.result_summary, updated_at=excluded.updated_at`, task.ID, task.ProjectID, task.AgentType, task.Objective, task.Capability, task.Status, task.Provenance, task.ValidationStatus, boolInt(task.Isolated), boolInt(task.Temporary), task.MemoryPolicy, task.ResultSummary, c, u)
	if err != nil {
		return SubagentTaskRecord{}, err
	}
	task.CreatedAt, task.UpdatedAt = c, u
	return task, nil
}

func (r SubagentTaskRepository) ListSubagentTaskRecords(projectID string) ([]SubagentTaskRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, agent_type, objective, capability, status, provenance, validation_status, isolated, temporary, memory_policy, result_summary, created_at, updated_at FROM subagent_tasks_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SubagentTaskRecord
	for rows.Next() {
		var task SubagentTaskRecord
		var isolated, temporary int
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.AgentType, &task.Objective, &task.Capability, &task.Status, &task.Provenance, &task.ValidationStatus, &isolated, &temporary, &task.MemoryPolicy, &task.ResultSummary, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		task.Isolated = isolated == 1
		task.Temporary = temporary == 1
		out = append(out, task)
	}
	return out, rows.Err()
}

func (r OpenCodeWorkflowRepository) SaveWorkflowRegistration(reg opencode.WorkflowRegistration) (opencode.WorkflowRegistration, error) {
	c, u := timestamps(reg.CreatedAt, reg.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO opencode_workflows_v2 (id, project_id, commands, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET commands=excluded.commands, status=excluded.status, updated_at=excluded.updated_at`, reg.ID, reg.ProjectID, mustJSON(reg.Commands), reg.Status, c, u)
	if err != nil {
		return opencode.WorkflowRegistration{}, err
	}
	reg.CreatedAt, reg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return reg, nil
}

func (r OpenCodeWorkflowRepository) ListWorkflowRegistrations(projectID string) ([]opencode.WorkflowRegistration, error) {
	rows, err := r.db.Query(`SELECT id, project_id, commands, status, created_at, updated_at FROM opencode_workflows_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []opencode.WorkflowRegistration
	for rows.Next() {
		var reg opencode.WorkflowRegistration
		var commands, c, u string
		if err := rows.Scan(&reg.ID, &reg.ProjectID, &commands, &reg.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(commands, &reg.Commands)
		reg.CreatedAt, reg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, reg)
	}
	return out, rows.Err()
}

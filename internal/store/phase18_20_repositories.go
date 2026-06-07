package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// This repository mirrors data from change_requests (the canonical source).
// All writes should go through change_repository.go first.
//
// ──────────────────────────────────────────────
// Phase 18: Change Engine Repositories
// ──────────────────────────────────────────────

// ChangeEventRecord is the DB representation of a change event.
type ChangeEventRecord struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	ChangeType  string `json:"change_type"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ChangeReportRecord is the DB representation of an impact report.
type ChangeReportRecord struct {
	ID               string `json:"id"`
	ChangeEventID    string `json:"change_event_id"`
	ProjectID        string `json:"project_id"`
	Analysis         string `json:"analysis"`          // JSON
	AffectedEntities string `json:"affected_entities"` // JSON array
	ReviewRequired   int    `json:"review_required"`
	Summary          string `json:"summary"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// SnapshotV2Record is the DB representation of a snapshot.
type SnapshotV2Record struct {
	ID             string `json:"id"`
	ProjectID      string `json:"project_id"`
	Reason         string `json:"reason"`
	EntitySnapshot string `json:"entity_snapshot"` // JSON
	CreatedAt      string `json:"created_at"`
}

// EntityStateRecord is the DB representation of an entity invalidation state.
type EntityStateRecord struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Status       string `json:"status"`
	LastChangeID string `json:"last_change_id"`
	Reason       string `json:"reason"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ChangeEventRepository persists change events.
type ChangeEventRepository struct{ db *sql.DB }

// NewChangeEventRepository creates a new repository.
func NewChangeEventRepository(db *sql.DB) *ChangeEventRepository {
	return &ChangeEventRepository{db: db}
}

func (r *ChangeEventRepository) Create(ev ChangeEventRecord) (ChangeEventRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO change_events (id, project_id, change_type, summary, description, severity, status, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ev.ID, ev.ProjectID, ev.ChangeType, ev.Summary, ev.Description, ev.Severity, ev.Status, ev.Source, now, now)
	if err != nil {
		return ev, err
	}

	crID := domain.NewID("change")
	r.db.Exec(`INSERT INTO change_requests (id, project_id, reason, description, status, requester, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET reason=excluded.reason, description=excluded.description, status=excluded.status, updated_at=excluded.updated_at`,
		crID, ev.ProjectID, ev.Summary, ev.Description, ev.Status, "", now, now)

	return r.Get(ev.ID)
}

func (r *ChangeEventRepository) Get(id string) (ChangeEventRecord, error) {
	var ev ChangeEventRecord
	err := r.db.QueryRow(`SELECT id, project_id, change_type, summary, description, severity, status, source, created_at, updated_at FROM change_events WHERE id = ?`, id).
		Scan(&ev.ID, &ev.ProjectID, &ev.ChangeType, &ev.Summary, &ev.Description, &ev.Severity, &ev.Status, &ev.Source, &ev.CreatedAt, &ev.UpdatedAt)
	return ev, err
}

func (r *ChangeEventRepository) ListByProject(projectID string, limit int) ([]ChangeEventRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, change_type, summary, description, severity, status, source, created_at, updated_at FROM change_events WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []ChangeEventRecord
	for rows.Next() {
		var ev ChangeEventRecord
		if err := rows.Scan(&ev.ID, &ev.ProjectID, &ev.ChangeType, &ev.Summary, &ev.Description, &ev.Severity, &ev.Status, &ev.Source, &ev.CreatedAt, &ev.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, ev)
	}
	return records, rows.Err()
}

func (r *ChangeEventRepository) UpdateStatus(id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE change_events SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	return err
}

// ChangeReportRepository persists impact reports.
type ChangeReportRepository struct{ db *sql.DB }

func NewChangeReportRepository(db *sql.DB) *ChangeReportRepository {
	return &ChangeReportRepository{db: db}
}

func (r *ChangeReportRepository) Create(report ChangeReportRecord) (ChangeReportRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO change_reports (id, change_event_id, project_id, analysis, affected_entities, review_required, summary, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		report.ID, report.ChangeEventID, report.ProjectID, report.Analysis, report.AffectedEntities, report.ReviewRequired, report.Summary, now, now)
	if err != nil {
		return report, err
	}
	return report, nil
}

func (r *ChangeReportRepository) GetByEvent(changeEventID string) (ChangeReportRecord, error) {
	var report ChangeReportRecord
	err := r.db.QueryRow(`SELECT id, change_event_id, project_id, analysis, affected_entities, review_required, summary, created_at, updated_at FROM change_reports WHERE change_event_id = ?`, changeEventID).
		Scan(&report.ID, &report.ChangeEventID, &report.ProjectID, &report.Analysis, &report.AffectedEntities, &report.ReviewRequired, &report.Summary, &report.CreatedAt, &report.UpdatedAt)
	return report, err
}

// SnapshotV2Repository persists v2 snapshots.
type SnapshotV2Repository struct{ db *sql.DB }

func NewSnapshotV2Repository(db *sql.DB) *SnapshotV2Repository {
	return &SnapshotV2Repository{db: db}
}

func (r *SnapshotV2Repository) Create(snap SnapshotV2Record) (SnapshotV2Record, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO snapshots_v2 (id, project_id, reason, entity_snapshot, created_at) VALUES (?, ?, ?, ?, ?)`,
		snap.ID, snap.ProjectID, snap.Reason, snap.EntitySnapshot, now)
	if err != nil {
		return snap, err
	}
	return snap, nil
}

func (r *SnapshotV2Repository) GetByID(id string) (SnapshotV2Record, error) {
	var s SnapshotV2Record
	err := r.db.QueryRow(`SELECT id, project_id, reason, entity_snapshot, created_at FROM snapshots_v2 WHERE id = ?`, id).
		Scan(&s.ID, &s.ProjectID, &s.Reason, &s.EntitySnapshot, &s.CreatedAt)
	return s, err
}

func (r *SnapshotV2Repository) ListByProject(projectID string, limit int) ([]SnapshotV2Record, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, reason, entity_snapshot, created_at FROM snapshots_v2 WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snaps []SnapshotV2Record
	for rows.Next() {
		var s SnapshotV2Record
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Reason, &s.EntitySnapshot, &s.CreatedAt); err != nil {
			return nil, err
		}
		snaps = append(snaps, s)
	}
	return snaps, rows.Err()
}

// EntityStateRepository persists entity invalidation states.
type EntityStateRepository struct{ db *sql.DB }

func NewEntityStateRepository(db *sql.DB) *EntityStateRepository {
	return &EntityStateRepository{db: db}
}

func (r *EntityStateRepository) Upsert(state EntityStateRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO entity_states (id, project_id, entity_type, entity_id, status, last_change_id, reason, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, entity_type, entity_id) DO UPDATE SET status = excluded.status, last_change_id = excluded.last_change_id, reason = excluded.reason, updated_at = excluded.updated_at`,
		state.ID, state.ProjectID, state.EntityType, state.EntityID, state.Status, state.LastChangeID, state.Reason, now, now)
	return err
}

func (r *EntityStateRepository) Get(entityType, entityID, projectID string) (EntityStateRecord, error) {
	var state EntityStateRecord
	err := r.db.QueryRow(`SELECT id, project_id, entity_type, entity_id, status, last_change_id, reason, created_at, updated_at FROM entity_states WHERE entity_type = ? AND entity_id = ? AND project_id = ?`, entityType, entityID, projectID).
		Scan(&state.ID, &state.ProjectID, &state.EntityType, &state.EntityID, &state.Status, &state.LastChangeID, &state.Reason, &state.CreatedAt, &state.UpdatedAt)
	return state, err
}

func (r *EntityStateRepository) ListByProject(projectID string) ([]EntityStateRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, entity_type, entity_id, status, last_change_id, reason, created_at, updated_at FROM entity_states WHERE project_id = ? ORDER BY entity_type, entity_id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var states []EntityStateRecord
	for rows.Next() {
		var s EntityStateRecord
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.EntityType, &s.EntityID, &s.Status, &s.LastChangeID, &s.Reason, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

// ──────────────────────────────────────────────
// Phase 19: MCP Server Repositories
// ──────────────────────────────────────────────

// MCPToolRecord represents a registered MCP tool.
type MCPToolRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaDef   string `json:"schema_def"` // JSON
	Enabled     int    `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// MCPRunRecord represents a tool execution record.
type MCPRunRecord struct {
	ID           string `json:"id"`
	ToolName     string `json:"tool_name"`
	Arguments    string `json:"arguments"` // JSON
	Success      int    `json:"success"`
	ErrorMessage string `json:"error_message"`
	CreatedAt    string `json:"created_at"`
}

// MCPToolRepository persists MCP tool registrations.
type MCPToolRepository struct{ db *sql.DB }

func NewMCPToolRepository(db *sql.DB) *MCPToolRepository {
	return &MCPToolRepository{db: db}
}

func (r *MCPToolRepository) Upsert(tool MCPToolRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO mcp_tools (id, name, description, schema_def, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET description = excluded.description, schema_def = excluded.schema_def, enabled = excluded.enabled, updated_at = excluded.updated_at`,
		tool.ID, tool.Name, tool.Description, tool.SchemaDef, tool.Enabled, now, now)
	return err
}

func (r *MCPToolRepository) List() ([]MCPToolRecord, error) {
	rows, err := r.db.Query(`SELECT id, name, description, schema_def, enabled, created_at, updated_at FROM mcp_tools WHERE enabled = 1 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tools []MCPToolRecord
	for rows.Next() {
		var t MCPToolRecord
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.SchemaDef, &t.Enabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

// MCPRunRepository persists tool execution records.
type MCPRunRepository struct{ db *sql.DB }

func NewMCPRunRepository(db *sql.DB) *MCPRunRepository {
	return &MCPRunRepository{db: db}
}

func (r *MCPRunRepository) Create(run MCPRunRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO mcp_runs (id, tool_name, arguments, success, error_message, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		run.ID, run.ToolName, run.Arguments, run.Success, run.ErrorMessage, now)
	return err
}

func (r *MCPRunRepository) ListByTool(toolName string, limit int) ([]MCPRunRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, tool_name, arguments, success, error_message, created_at FROM mcp_runs WHERE tool_name = ? ORDER BY created_at DESC LIMIT ?`, toolName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []MCPRunRecord
	for rows.Next() {
		var run MCPRunRecord
		if err := rows.Scan(&run.ID, &run.ToolName, &run.Arguments, &run.Success, &run.ErrorMessage, &run.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// ──────────────────────────────────────────────
// Phase 20: OpenCode Integration Repositories
// ──────────────────────────────────────────────

// OpenCodeDetectionRecord represents a detection result.
type OpenCodeDetectionRecord struct {
	ID            string `json:"id"`
	ProjectRoot   string `json:"project_root"`
	Found         int    `json:"found"`
	ConfigPath    string `json:"config_path"`
	IsInitialized int    `json:"is_initialized"`
	AgentName     string `json:"agent_name"`
	SkillCount    int    `json:"skill_count"`
	CreatedAt     string `json:"created_at"`
}

// OpenCodeIntegrationStateRecord represents the integration state.
type OpenCodeIntegrationStateRecord struct {
	ID             string `json:"id"`
	ProjectRoot    string `json:"project_root"`
	Mode           string `json:"mode"`
	Enabled        int    `json:"enabled"`
	ReadOnly       int    `json:"read_only"`
	LastDetectedAt string `json:"last_detected_at"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// OpenCodeDoctorCheckRecord represents a doctor check result.
type OpenCodeDoctorCheckRecord struct {
	ID          string `json:"id"`
	ProjectRoot string `json:"project_root"`
	CheckName   string `json:"check_name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
}

// OpenCodeDetectionRepository persists detection results.
type OpenCodeDetectionRepository struct{ db *sql.DB }

func NewOpenCodeDetectionRepository(db *sql.DB) *OpenCodeDetectionRepository {
	return &OpenCodeDetectionRepository{db: db}
}

func (r *OpenCodeDetectionRepository) Create(det OpenCodeDetectionRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO opencode_detections (id, project_root, found, config_path, is_initialized, agent_name, skill_count, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		det.ID, det.ProjectRoot, det.Found, det.ConfigPath, det.IsInitialized, det.AgentName, det.SkillCount, now)
	return err
}

func (r *OpenCodeDetectionRepository) Latest(projectRoot string) (OpenCodeDetectionRecord, error) {
	var det OpenCodeDetectionRecord
	err := r.db.QueryRow(`SELECT id, project_root, found, config_path, is_initialized, agent_name, skill_count, created_at FROM opencode_detections WHERE project_root = ? ORDER BY created_at DESC LIMIT 1`, projectRoot).
		Scan(&det.ID, &det.ProjectRoot, &det.Found, &det.ConfigPath, &det.IsInitialized, &det.AgentName, &det.SkillCount, &det.CreatedAt)
	return det, err
}

// OpenCodeIntegrationRepository persists integration state.
type OpenCodeIntegrationRepository struct{ db *sql.DB }

func NewOpenCodeIntegrationRepository(db *sql.DB) *OpenCodeIntegrationRepository {
	return &OpenCodeIntegrationRepository{db: db}
}

func (r *OpenCodeIntegrationRepository) Upsert(state OpenCodeIntegrationStateRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO opencode_integration_state (id, project_root, mode, enabled, read_only, last_detected_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_root) DO UPDATE SET mode = excluded.mode, enabled = excluded.enabled, read_only = excluded.read_only, last_detected_at = excluded.last_detected_at, updated_at = excluded.updated_at`,
		state.ID, state.ProjectRoot, state.Mode, state.Enabled, state.ReadOnly, state.LastDetectedAt, now, now)
	return err
}

func (r *OpenCodeIntegrationRepository) Get(projectRoot string) (OpenCodeIntegrationStateRecord, error) {
	var state OpenCodeIntegrationStateRecord
	err := r.db.QueryRow(`SELECT id, project_root, mode, enabled, read_only, last_detected_at, created_at, updated_at FROM opencode_integration_state WHERE project_root = ?`, projectRoot).
		Scan(&state.ID, &state.ProjectRoot, &state.Mode, &state.Enabled, &state.ReadOnly, &state.LastDetectedAt, &state.CreatedAt, &state.UpdatedAt)
	return state, err
}

// OpenCodeDoctorRepository persists doctor check results.
type OpenCodeDoctorRepository struct{ db *sql.DB }

func NewOpenCodeDoctorRepository(db *sql.DB) *OpenCodeDoctorRepository {
	return &OpenCodeDoctorRepository{db: db}
}

func (r *OpenCodeDoctorRepository) Create(check OpenCodeDoctorCheckRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO opencode_doctor_checks (id, project_root, check_name, status, message, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		check.ID, check.ProjectRoot, check.CheckName, check.Status, check.Message, now)
	return err
}

func (r *OpenCodeDoctorRepository) LatestByProject(projectRoot string) ([]OpenCodeDoctorCheckRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_root, check_name, status, message, created_at FROM opencode_doctor_checks WHERE project_root = ? ORDER BY created_at DESC`, projectRoot)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var checks []OpenCodeDoctorCheckRecord
	for rows.Next() {
		var c OpenCodeDoctorCheckRecord
		if err := rows.Scan(&c.ID, &c.ProjectRoot, &c.CheckName, &c.Status, &c.Message, &c.CreatedAt); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

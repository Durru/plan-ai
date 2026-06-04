package store

import (
	"database/sql"
	"time"
)

// ──────────────────────────────────────────────
// Phase 24: Vision Discovery Engine Repositories
// ──────────────────────────────────────────────

// VisionDiscoverySessionRecord represents a discovery session.
type VisionDiscoverySessionRecord struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`
	Summary    string `json:"summary"`
	RawContext string `json:"raw_context"`
	Findings   string `json:"findings"` // JSON array
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// VisionAssumptionRecord represents an identified assumption.
type VisionAssumptionRecord struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	SessionID   string  `json:"session_id"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Confidence  float64 `json:"confidence"`
	Status      string  `json:"status"`
	ValidatedBy string  `json:"validated_by"`
	ValidatedAt string  `json:"validated_at"`
	CreatedAt   string  `json:"created_at"`
}

// VisionAmbiguityRecord represents an identified ambiguity.
type VisionAmbiguityRecord struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	SessionID   string `json:"session_id"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Resolution  string `json:"resolution"`
	Status      string `json:"status"`
	ResolvedAt  string `json:"resolved_at"`
	CreatedAt   string `json:"created_at"`
}

// VisionApprovalRecord represents a vision approval.
type VisionApprovalRecord struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	SessionID  string `json:"session_id"`
	VisionID   string `json:"vision_id"`
	Status     string `json:"status"`
	ApprovedBy string `json:"approved_by"`
	ApprovedAt string `json:"approved_at"`
	Feedback   string `json:"feedback"`
	CreatedAt  string `json:"created_at"`
}

// VisionDiscoverySessionRepository persists vision discovery sessions.
type VisionDiscoverySessionRepository struct{ db *sql.DB }

// NewVisionDiscoverySessionRepository creates a new VisionDiscoverySessionRepository.
func NewVisionDiscoverySessionRepository(db *sql.DB) *VisionDiscoverySessionRepository {
	return &VisionDiscoverySessionRepository{db: db}
}

func (r *VisionDiscoverySessionRepository) Create(s VisionDiscoverySessionRecord) (VisionDiscoverySessionRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO vision_discovery_sessions (id, project_id, status, summary, raw_context, findings, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.ProjectID, s.Status, s.Summary, s.RawContext, s.Findings, now, now)
	if err != nil {
		return s, err
	}
	s.CreatedAt = now
	s.UpdatedAt = now
	return s, nil
}

func (r *VisionDiscoverySessionRepository) Get(id string) (VisionDiscoverySessionRecord, error) {
	var rec VisionDiscoverySessionRecord
	err := r.db.QueryRow(`SELECT id, project_id, status, summary, raw_context, findings, created_at, updated_at FROM vision_discovery_sessions WHERE id = ?`, id).
		Scan(&rec.ID, &rec.ProjectID, &rec.Status, &rec.Summary, &rec.RawContext, &rec.Findings, &rec.CreatedAt, &rec.UpdatedAt)
	return rec, err
}

func (r *VisionDiscoverySessionRepository) ListByProject(projectID string, limit int) ([]VisionDiscoverySessionRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Query(`SELECT id, project_id, status, summary, raw_context, findings, created_at, updated_at FROM vision_discovery_sessions WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []VisionDiscoverySessionRecord
	for rows.Next() {
		var rec VisionDiscoverySessionRecord
		if err := rows.Scan(&rec.ID, &rec.ProjectID, &rec.Status, &rec.Summary, &rec.RawContext, &rec.Findings, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *VisionDiscoverySessionRepository) Update(id, status, summary string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE vision_discovery_sessions SET status = ?, summary = ?, updated_at = ? WHERE id = ?`, status, summary, now, id)
	return err
}

// VisionAssumptionRepository persists vision assumptions.
type VisionAssumptionRepository struct{ db *sql.DB }

// NewVisionAssumptionRepository creates a new VisionAssumptionRepository.
func NewVisionAssumptionRepository(db *sql.DB) *VisionAssumptionRepository {
	return &VisionAssumptionRepository{db: db}
}

func (r *VisionAssumptionRepository) Create(a VisionAssumptionRecord) (VisionAssumptionRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO vision_assumptions (id, project_id, session_id, description, category, confidence, status, validated_by, validated_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ProjectID, a.SessionID, a.Description, a.Category, a.Confidence, a.Status, a.ValidatedBy, a.ValidatedAt, now)
	if err != nil {
		return a, err
	}
	a.CreatedAt = now
	return a, nil
}

func (r *VisionAssumptionRepository) ListBySession(sessionID string) ([]VisionAssumptionRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, session_id, description, category, confidence, status, validated_by, validated_at, created_at FROM vision_assumptions WHERE session_id = ? ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []VisionAssumptionRecord
	for rows.Next() {
		var a VisionAssumptionRecord
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.Description, &a.Category, &a.Confidence, &a.Status, &a.ValidatedBy, &a.ValidatedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, rows.Err()
}

func (r *VisionAssumptionRepository) UpdateStatus(id, status, validatedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if status == "validated" || status == "invalidated" {
		_, err := r.db.Exec(`UPDATE vision_assumptions SET status = ?, validated_by = ?, validated_at = ? WHERE id = ?`, status, validatedBy, now, id)
		return err
	}
	_, err := r.db.Exec(`UPDATE vision_assumptions SET status = ? WHERE id = ?`, status, id)
	return err
}

// VisionAmbiguityRepository persists vision ambiguities.
type VisionAmbiguityRepository struct{ db *sql.DB }

// NewVisionAmbiguityRepository creates a new VisionAmbiguityRepository.
func NewVisionAmbiguityRepository(db *sql.DB) *VisionAmbiguityRepository {
	return &VisionAmbiguityRepository{db: db}
}

func (r *VisionAmbiguityRepository) Create(a VisionAmbiguityRecord) (VisionAmbiguityRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO vision_ambiguities (id, project_id, session_id, description, category, resolution, status, resolved_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ProjectID, a.SessionID, a.Description, a.Category, a.Resolution, a.Status, a.ResolvedAt, now)
	if err != nil {
		return a, err
	}
	a.CreatedAt = now
	return a, nil
}

func (r *VisionAmbiguityRepository) ListBySession(sessionID string) ([]VisionAmbiguityRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, session_id, description, category, resolution, status, resolved_at, created_at FROM vision_ambiguities WHERE session_id = ? ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []VisionAmbiguityRecord
	for rows.Next() {
		var a VisionAmbiguityRecord
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.Description, &a.Category, &a.Resolution, &a.Status, &a.ResolvedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, rows.Err()
}

func (r *VisionAmbiguityRepository) Resolve(id, resolution string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE vision_ambiguities SET status = 'resolved', resolution = ?, resolved_at = ? WHERE id = ?`, resolution, now, id)
	return err
}

// VisionApprovalRepository persists vision approvals.
type VisionApprovalRepository struct{ db *sql.DB }

// NewVisionApprovalRepository creates a new VisionApprovalRepository.
func NewVisionApprovalRepository(db *sql.DB) *VisionApprovalRepository {
	return &VisionApprovalRepository{db: db}
}

func (r *VisionApprovalRepository) Create(a VisionApprovalRecord) (VisionApprovalRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO vision_approvals (id, project_id, session_id, vision_id, status, approved_by, approved_at, feedback, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ProjectID, a.SessionID, a.VisionID, a.Status, a.ApprovedBy, a.ApprovedAt, a.Feedback, now)
	if err != nil {
		return a, err
	}
	a.CreatedAt = now
	return a, nil
}

func (r *VisionApprovalRepository) GetByVision(visionID string) (VisionApprovalRecord, error) {
	var a VisionApprovalRecord
	err := r.db.QueryRow(`SELECT id, project_id, session_id, vision_id, status, approved_by, approved_at, feedback, created_at FROM vision_approvals WHERE vision_id = ? ORDER BY created_at DESC LIMIT 1`, visionID).
		Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.VisionID, &a.Status, &a.ApprovedBy, &a.ApprovedAt, &a.Feedback, &a.CreatedAt)
	return a, err
}

func (r *VisionApprovalRepository) Approve(id, approvedBy, feedback string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE vision_approvals SET status = 'approved', approved_by = ?, approved_at = ?, feedback = ? WHERE id = ?`, approvedBy, now, feedback, id)
	return err
}

func (r *VisionApprovalRepository) Reject(id, feedback string) error {
	_, err := r.db.Exec(`UPDATE vision_approvals SET status = 'rejected', feedback = ? WHERE id = ?`, feedback, id)
	return err
}

// ──────────────────────────────────────────────
// Phase 25: Master Plan V2 Repositories
// ──────────────────────────────────────────────

// MasterPlanVersionRecord represents a versioned master plan.
type MasterPlanVersionRecord struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Version      int    `json:"version"`
	PlanID       string `json:"plan_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Phases       string `json:"phases"`       // JSON array
	Timeline     string `json:"timeline"`     // JSON object
	Risks        string `json:"risks"`        // JSON array
	Dependencies string `json:"dependencies"` // JSON array
	Status       string `json:"status"`
	Changelog    string `json:"changelog"`
	CreatedAt    string `json:"created_at"`
}

// MasterPlanChangeRecord represents a change between versions.
type MasterPlanChangeRecord struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	PlanID      string `json:"plan_id"`
	VersionFrom int    `json:"version_from"`
	VersionTo   int    `json:"version_to"`
	ChangeType  string `json:"change_type"`
	Description string `json:"description"`
	Author      string `json:"author"`
	CreatedAt   string `json:"created_at"`
}

// MasterPlanApprovalRecord represents a plan approval.
type MasterPlanApprovalRecord struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	PlanID     string `json:"plan_id"`
	Version    int    `json:"version"`
	Status     string `json:"status"`
	ApprovedBy string `json:"approved_by"`
	ApprovedAt string `json:"approved_at"`
	Feedback   string `json:"feedback"`
	CreatedAt  string `json:"created_at"`
}

// PlanEvolutionEventRecord represents an evolution event.
type PlanEvolutionEventRecord struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	EventType   string `json:"event_type"`
	Description string `json:"description"`
	Details     string `json:"details"` // JSON
	CreatedAt   string `json:"created_at"`
}

// MasterPlanVersionRepository persists master plan versions.
type MasterPlanVersionRepository struct{ db *sql.DB }

// NewMasterPlanVersionRepository creates a new MasterPlanVersionRepository.
func NewMasterPlanVersionRepository(db *sql.DB) *MasterPlanVersionRepository {
	return &MasterPlanVersionRepository{db: db}
}

func (r *MasterPlanVersionRepository) Create(v MasterPlanVersionRecord) (MasterPlanVersionRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO master_plan_versions (id, project_id, version, plan_id, title, description, phases, timeline, risks, dependencies, status, changelog, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.ProjectID, v.Version, v.PlanID, v.Title, v.Description, v.Phases, v.Timeline, v.Risks, v.Dependencies, v.Status, v.Changelog, now)
	if err != nil {
		return v, err
	}
	v.CreatedAt = now
	return v, nil
}

func (r *MasterPlanVersionRepository) GetLatest(planID string) (MasterPlanVersionRecord, error) {
	var v MasterPlanVersionRecord
	err := r.db.QueryRow(`SELECT id, project_id, version, plan_id, title, description, phases, timeline, risks, dependencies, status, changelog, created_at FROM master_plan_versions WHERE plan_id = ? ORDER BY version DESC LIMIT 1`, planID).
		Scan(&v.ID, &v.ProjectID, &v.Version, &v.PlanID, &v.Title, &v.Description, &v.Phases, &v.Timeline, &v.Risks, &v.Dependencies, &v.Status, &v.Changelog, &v.CreatedAt)
	return v, err
}

func (r *MasterPlanVersionRepository) ListByPlan(planID string) ([]MasterPlanVersionRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, version, plan_id, title, description, phases, timeline, risks, dependencies, status, changelog, created_at FROM master_plan_versions WHERE plan_id = ? ORDER BY version DESC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []MasterPlanVersionRecord
	for rows.Next() {
		var v MasterPlanVersionRecord
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Version, &v.PlanID, &v.Title, &v.Description, &v.Phases, &v.Timeline, &v.Risks, &v.Dependencies, &v.Status, &v.Changelog, &v.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, v)
	}
	return records, rows.Err()
}

// MasterPlanChangeRepository persists plan changes.
type MasterPlanChangeRepository struct{ db *sql.DB }

// NewMasterPlanChangeRepository creates a new MasterPlanChangeRepository.
func NewMasterPlanChangeRepository(db *sql.DB) *MasterPlanChangeRepository {
	return &MasterPlanChangeRepository{db: db}
}

func (r *MasterPlanChangeRepository) Create(c MasterPlanChangeRecord) (MasterPlanChangeRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO master_plan_changes (id, project_id, plan_id, version_from, version_to, change_type, description, author, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.ProjectID, c.PlanID, c.VersionFrom, c.VersionTo, c.ChangeType, c.Description, c.Author, now)
	if err != nil {
		return c, err
	}
	c.CreatedAt = now
	return c, nil
}

func (r *MasterPlanChangeRepository) ListByPlan(planID string) ([]MasterPlanChangeRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, plan_id, version_from, version_to, change_type, description, author, created_at FROM master_plan_changes WHERE plan_id = ? ORDER BY created_at DESC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []MasterPlanChangeRecord
	for rows.Next() {
		var c MasterPlanChangeRecord
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.PlanID, &c.VersionFrom, &c.VersionTo, &c.ChangeType, &c.Description, &c.Author, &c.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, c)
	}
	return records, rows.Err()
}

// MasterPlanApprovalRepository persists plan approvals.
type MasterPlanApprovalRepository struct{ db *sql.DB }

// NewMasterPlanApprovalRepository creates a new MasterPlanApprovalRepository.
func NewMasterPlanApprovalRepository(db *sql.DB) *MasterPlanApprovalRepository {
	return &MasterPlanApprovalRepository{db: db}
}

func (r *MasterPlanApprovalRepository) Create(a MasterPlanApprovalRecord) (MasterPlanApprovalRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO master_plan_approvals (id, project_id, plan_id, version, status, approved_by, approved_at, feedback, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ProjectID, a.PlanID, a.Version, a.Status, a.ApprovedBy, a.ApprovedAt, a.Feedback, now)
	if err != nil {
		return a, err
	}
	a.CreatedAt = now
	return a, nil
}

func (r *MasterPlanApprovalRepository) GetLatest(planID string) (MasterPlanApprovalRecord, error) {
	var a MasterPlanApprovalRecord
	err := r.db.QueryRow(`SELECT id, project_id, plan_id, version, status, approved_by, approved_at, feedback, created_at FROM master_plan_approvals WHERE plan_id = ? ORDER BY created_at DESC LIMIT 1`, planID).
		Scan(&a.ID, &a.ProjectID, &a.PlanID, &a.Version, &a.Status, &a.ApprovedBy, &a.ApprovedAt, &a.Feedback, &a.CreatedAt)
	return a, err
}

// PlanEvolutionEventRepository persists evolution events.
type PlanEvolutionEventRepository struct{ db *sql.DB }

// NewPlanEvolutionEventRepository creates a new PlanEvolutionEventRepository.
func NewPlanEvolutionEventRepository(db *sql.DB) *PlanEvolutionEventRepository {
	return &PlanEvolutionEventRepository{db: db}
}

func (r *PlanEvolutionEventRepository) Create(e PlanEvolutionEventRecord) (PlanEvolutionEventRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO plan_evolution_events (id, project_id, entity_type, entity_id, event_type, description, details, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.ProjectID, e.EntityType, e.EntityID, e.EventType, e.Description, e.Details, now)
	if err != nil {
		return e, err
	}
	e.CreatedAt = now
	return e, nil
}

func (r *PlanEvolutionEventRepository) ListByProject(projectID string, limit int) ([]PlanEvolutionEventRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, project_id, entity_type, entity_id, event_type, description, details, created_at FROM plan_evolution_events WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []PlanEvolutionEventRecord
	for rows.Next() {
		var e PlanEvolutionEventRecord
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.EntityType, &e.EntityID, &e.EventType, &e.Description, &e.Details, &e.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, e)
	}
	return records, rows.Err()
}

// ──────────────────────────────────────────────
// Phase 26: Specific Plan V2 Repositories
// ──────────────────────────────────────────────

// SpecificPlanVersionRecord represents a versioned specific plan.
type SpecificPlanVersionRecord struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Version      int    `json:"version"`
	PlanID       string `json:"plan_id"`
	Domain       string `json:"domain"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Tasks        string `json:"tasks"`        // JSON array
	Dependencies string `json:"dependencies"` // JSON array
	Risks        string `json:"risks"`        // JSON array
	Status       string `json:"status"`
	Changelog    string `json:"changelog"`
	CreatedAt    string `json:"created_at"`
}

// SpecificPlanResearchLinkRecord represents a link between a plan and research.
type SpecificPlanResearchLinkRecord struct {
	ID         string  `json:"id"`
	ProjectID  string  `json:"project_id"`
	PlanID     string  `json:"plan_id"`
	ResearchID string  `json:"research_id"`
	Section    string  `json:"section"`
	Relevance  float64 `json:"relevance"`
	CreatedAt  string  `json:"created_at"`
}

// SpecificPlanRegenerationRecord represents a regeneration event.
type SpecificPlanRegenerationRecord struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	PlanID      string `json:"plan_id"`
	VersionFrom int    `json:"version_from"`
	VersionTo   int    `json:"version_to"`
	Reason      string `json:"reason"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// SpecificPlanVersionRepository persists specific plan versions.
type SpecificPlanVersionRepository struct{ db *sql.DB }

// NewSpecificPlanVersionRepository creates a new SpecificPlanVersionRepository.
func NewSpecificPlanVersionRepository(db *sql.DB) *SpecificPlanVersionRepository {
	return &SpecificPlanVersionRepository{db: db}
}

func (r *SpecificPlanVersionRepository) Create(v SpecificPlanVersionRecord) (SpecificPlanVersionRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO specific_plan_versions (id, project_id, version, plan_id, domain, title, description, tasks, dependencies, risks, status, changelog, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.ProjectID, v.Version, v.PlanID, v.Domain, v.Title, v.Description, v.Tasks, v.Dependencies, v.Risks, v.Status, v.Changelog, now)
	if err != nil {
		return v, err
	}
	v.CreatedAt = now
	return v, nil
}

func (r *SpecificPlanVersionRepository) GetLatest(planID string) (SpecificPlanVersionRecord, error) {
	var v SpecificPlanVersionRecord
	err := r.db.QueryRow(`SELECT id, project_id, version, plan_id, domain, title, description, tasks, dependencies, risks, status, changelog, created_at FROM specific_plan_versions WHERE plan_id = ? ORDER BY version DESC LIMIT 1`, planID).
		Scan(&v.ID, &v.ProjectID, &v.Version, &v.PlanID, &v.Domain, &v.Title, &v.Description, &v.Tasks, &v.Dependencies, &v.Risks, &v.Status, &v.Changelog, &v.CreatedAt)
	return v, err
}

func (r *SpecificPlanVersionRepository) ListByPlan(planID string) ([]SpecificPlanVersionRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, version, plan_id, domain, title, description, tasks, dependencies, risks, status, changelog, created_at FROM specific_plan_versions WHERE plan_id = ? ORDER BY version DESC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SpecificPlanVersionRecord
	for rows.Next() {
		var v SpecificPlanVersionRecord
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Version, &v.PlanID, &v.Domain, &v.Title, &v.Description, &v.Tasks, &v.Dependencies, &v.Risks, &v.Status, &v.Changelog, &v.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, v)
	}
	return records, rows.Err()
}

// SpecificPlanResearchLinkRepository persists research links.
type SpecificPlanResearchLinkRepository struct{ db *sql.DB }

// NewSpecificPlanResearchLinkRepository creates a new SpecificPlanResearchLinkRepository.
func NewSpecificPlanResearchLinkRepository(db *sql.DB) *SpecificPlanResearchLinkRepository {
	return &SpecificPlanResearchLinkRepository{db: db}
}

func (r *SpecificPlanResearchLinkRepository) Create(l SpecificPlanResearchLinkRecord) (SpecificPlanResearchLinkRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO specific_plan_research_links (id, project_id, plan_id, research_id, section, relevance, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.ProjectID, l.PlanID, l.ResearchID, l.Section, l.Relevance, now)
	if err != nil {
		return l, err
	}
	l.CreatedAt = now
	return l, nil
}

func (r *SpecificPlanResearchLinkRepository) ListByPlan(planID string) ([]SpecificPlanResearchLinkRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, plan_id, research_id, section, relevance, created_at FROM specific_plan_research_links WHERE plan_id = ? ORDER BY created_at`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SpecificPlanResearchLinkRecord
	for rows.Next() {
		var l SpecificPlanResearchLinkRecord
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.PlanID, &l.ResearchID, &l.Section, &l.Relevance, &l.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, l)
	}
	return records, rows.Err()
}

// SpecificPlanRegenerationRepository persists regenerations.
type SpecificPlanRegenerationRepository struct{ db *sql.DB }

// NewSpecificPlanRegenerationRepository creates a new SpecificPlanRegenerationRepository.
func NewSpecificPlanRegenerationRepository(db *sql.DB) *SpecificPlanRegenerationRepository {
	return &SpecificPlanRegenerationRepository{db: db}
}

func (r *SpecificPlanRegenerationRepository) Create(reg SpecificPlanRegenerationRecord) (SpecificPlanRegenerationRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO specific_plan_regenerations (id, project_id, plan_id, version_from, version_to, reason, scope, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reg.ID, reg.ProjectID, reg.PlanID, reg.VersionFrom, reg.VersionTo, reg.Reason, reg.Scope, reg.Status, now)
	if err != nil {
		return reg, err
	}
	reg.CreatedAt = now
	return reg, nil
}

func (r *SpecificPlanRegenerationRepository) ListByPlan(planID string) ([]SpecificPlanRegenerationRecord, error) {
	rows, err := r.db.Query(`SELECT id, project_id, plan_id, version_from, version_to, reason, scope, status, created_at FROM specific_plan_regenerations WHERE plan_id = ? ORDER BY created_at DESC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SpecificPlanRegenerationRecord
	for rows.Next() {
		var reg SpecificPlanRegenerationRecord
		if err := rows.Scan(&reg.ID, &reg.ProjectID, &reg.PlanID, &reg.VersionFrom, &reg.VersionTo, &reg.Reason, &reg.Scope, &reg.Status, &reg.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, reg)
	}
	return records, rows.Err()
}

// ──────────────────────────────────────────────
// Phase 27: Context Delivery Engine Repositories
// ──────────────────────────────────────────────

// ContextDeliverySessionRecord represents a delivery session.
type ContextDeliverySessionRecord struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Level        string `json:"level"`
	BudgetTokens int    `json:"budget_tokens"`
	TokensUsed   int    `json:"tokens_used"`
	Content      string `json:"content"`
	Metadata     string `json:"metadata"` // JSON
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// ContextDeliveryUsageRecord represents a usage metric entry.
type ContextDeliveryUsageRecord struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	SessionID string `json:"session_id"`
	Level     string `json:"level"`
	Tokens    int    `json:"tokens"`
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
}

// ContextDeliveryBudgetRecord represents a budget for a level.
type ContextDeliveryBudgetRecord struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	Level        string `json:"level"`
	MaxTokens    int    `json:"max_tokens"`
	CurrentUsage int    `json:"current_usage"`
	Strategy     string `json:"strategy"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ContextDeliverySessionRepository persists delivery sessions.
type ContextDeliverySessionRepository struct{ db *sql.DB }

// NewContextDeliverySessionRepository creates a new ContextDeliverySessionRepository.
func NewContextDeliverySessionRepository(db *sql.DB) *ContextDeliverySessionRepository {
	return &ContextDeliverySessionRepository{db: db}
}

func (r *ContextDeliverySessionRepository) Create(s ContextDeliverySessionRecord) (ContextDeliverySessionRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO context_delivery_sessions (id, project_id, level, budget_tokens, tokens_used, content, metadata, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.ProjectID, s.Level, s.BudgetTokens, s.TokensUsed, s.Content, s.Metadata, s.Status, now)
	if err != nil {
		return s, err
	}
	s.CreatedAt = now
	return s, nil
}

func (r *ContextDeliverySessionRepository) ListByProject(projectID string, level string, limit int) ([]ContextDeliverySessionRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows *sql.Rows
	var err error
	if level != "" {
		rows, err = r.db.Query(`SELECT id, project_id, level, budget_tokens, tokens_used, content, metadata, status, created_at FROM context_delivery_sessions WHERE project_id = ? AND level = ? ORDER BY created_at DESC LIMIT ?`, projectID, level, limit)
	} else {
		rows, err = r.db.Query(`SELECT id, project_id, level, budget_tokens, tokens_used, content, metadata, status, created_at FROM context_delivery_sessions WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []ContextDeliverySessionRecord
	for rows.Next() {
		var s ContextDeliverySessionRecord
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Level, &s.BudgetTokens, &s.TokensUsed, &s.Content, &s.Metadata, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, s)
	}
	return records, rows.Err()
}

// ContextDeliveryUsageRepository persists usage metrics.
type ContextDeliveryUsageRepository struct{ db *sql.DB }

// NewContextDeliveryUsageRepository creates a new ContextDeliveryUsageRepository.
func NewContextDeliveryUsageRepository(db *sql.DB) *ContextDeliveryUsageRepository {
	return &ContextDeliveryUsageRepository{db: db}
}

func (r *ContextDeliveryUsageRepository) Create(u ContextDeliveryUsageRecord) (ContextDeliveryUsageRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO context_delivery_usage (id, project_id, session_id, level, tokens, source, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.ProjectID, u.SessionID, u.Level, u.Tokens, u.Source, now)
	if err != nil {
		return u, err
	}
	u.CreatedAt = now
	return u, nil
}

func (r *ContextDeliveryUsageRepository) GetTotalByProject(projectID string) (int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COALESCE(SUM(tokens), 0) FROM context_delivery_usage WHERE project_id = ?`, projectID).Scan(&total)
	return total, err
}

// ContextDeliveryBudgetRepository persists budgets.
type ContextDeliveryBudgetRepository struct{ db *sql.DB }

// NewContextDeliveryBudgetRepository creates a new ContextDeliveryBudgetRepository.
func NewContextDeliveryBudgetRepository(db *sql.DB) *ContextDeliveryBudgetRepository {
	return &ContextDeliveryBudgetRepository{db: db}
}

func (r *ContextDeliveryBudgetRepository) CreateOrUpdate(b ContextDeliveryBudgetRecord) (ContextDeliveryBudgetRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO context_delivery_budgets (id, project_id, level, max_tokens, current_usage, strategy, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET max_tokens = ?, current_usage = ?, strategy = ?, updated_at = ?`,
		b.ID, b.ProjectID, b.Level, b.MaxTokens, b.CurrentUsage, b.Strategy, now, now,
		b.MaxTokens, b.CurrentUsage, b.Strategy, now)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (r *ContextDeliveryBudgetRepository) GetByLevel(projectID, level string) (ContextDeliveryBudgetRecord, error) {
	var b ContextDeliveryBudgetRecord
	err := r.db.QueryRow(`SELECT id, project_id, level, max_tokens, current_usage, strategy, created_at, updated_at FROM context_delivery_budgets WHERE project_id = ? AND level = ?`, projectID, level).
		Scan(&b.ID, &b.ProjectID, &b.Level, &b.MaxTokens, &b.CurrentUsage, &b.Strategy, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

package planning

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Master Plan V2 Types
// ──────────────────────────────────────────────

// MasterPlanV2 extends MasterPlan with richer fields.
type MasterPlanV2 struct {
	ID           string      `json:"id"`
	ProjectID    string      `json:"project_id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Phases       []PhaseDef  `json:"phases"`
	Timeline     Timeline    `json:"timeline"`
	Risks        []RiskEntry `json:"risks"`
	Dependencies []string    `json:"dependencies"`
	Status       Status      `json:"status"`
	Version      int         `json:"version"`
	Changelog    string      `json:"changelog"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// PhaseDef defines a phase in a master plan.
type PhaseDef struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Objectives     []string `json:"objectives"`
	Order          int      `json:"order"`
	EstimatedWeeks int      `json:"estimated_weeks"`
}

// Timeline represents a plan timeline.
type Timeline struct {
	StartDate  *time.Time  `json:"start_date,omitempty"`
	TargetDate *time.Time  `json:"target_date,omitempty"`
	TotalWeeks int         `json:"total_weeks"`
	Milestones []Milestone `json:"milestones"`
}

// Milestone represents a key milestone.
type Milestone struct {
	Name     string `json:"name"`
	Week     int    `json:"week"`
	Criteria string `json:"criteria"`
}

// RiskEntry represents a documented risk.
type RiskEntry struct {
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Likelihood  string `json:"likelihood"`
	Mitigation  string `json:"mitigation"`
	Status      string `json:"status"`
}

// PlanChangeType defines the type of plan change.
type PlanChangeType string

const (
	ChangeCreated      PlanChangeType = "created"
	ChangeUpdated      PlanChangeType = "updated"
	ChangeApproved     PlanChangeType = "approved"
	ChangeArchived     PlanChangeType = "archived"
	ChangeRestructured PlanChangeType = "restructured"
	ChangeAmended      PlanChangeType = "amended"
)

// PlanEvolutionEvent represents a tracked evolution event.
type PlanEvolutionEvent struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Details     string    `json:"details"`
	CreatedAt   time.Time `json:"created_at"`
}

// ──────────────────────────────────────────────
// Master Plan V2 Repository Interface
// ──────────────────────────────────────────────

// MasterPlanV2Repository provides persistence for master plan v2 entities.
type MasterPlanV2Repository interface {
	CreateVersion(record interface{}) error
	GetLatestVersion(planID string) (interface{}, error)
	ListVersions(planID string) ([]interface{}, error)

	CreateChange(record interface{}) error
	ListChanges(planID string) ([]interface{}, error)

	CreateApproval(record interface{}) error
	GetLatestApproval(planID string) (interface{}, error)

	CreateEvolutionEvent(event PlanEvolutionEvent) (PlanEvolutionEvent, error)
	ListEvolutionEvents(projectID string) ([]PlanEvolutionEvent, error)
}

// ──────────────────────────────────────────────
// Master Plan Generator
// ──────────────────────────────────────────────

// MasterPlanGenerator creates and manages master plans with versioning.
type MasterPlanGenerator struct {
	storeRepo Repository
	v2Repo    MasterPlanV2Repository
	now       func() time.Time
}

// NewMasterPlanGenerator creates a new MasterPlanGenerator.
func NewMasterPlanGenerator(storeRepo Repository, v2Repo MasterPlanV2Repository) *MasterPlanGenerator {
	return &MasterPlanGenerator{storeRepo: storeRepo, v2Repo: v2Repo, now: time.Now().UTC}
}

// GenerateV2 creates a new versioned master plan from input.
func (g *MasterPlanGenerator) GenerateV2(input PlanningInput, description string, phases []PhaseDef, timeline Timeline, risks []RiskEntry) (*MasterPlanV2, error) {
	if strings.TrimSpace(input.ProjectID) == "" {
		return nil, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(input.VisionReference) == "" {
		return nil, fmt.Errorf("vision reference is required")
	}

	now := g.now()
	title := "Master Plan"
	if len(input.ApprovedRequirements) > 0 {
		title = input.ApprovedRequirements[0]
	}
	if description == "" {
		description = "Auto-generated master plan"
	}

	plan := &MasterPlanV2{
		ID:           domain.NewID("mpv2"),
		ProjectID:    input.ProjectID,
		Title:        title,
		Description:  description,
		Phases:       phases,
		Timeline:     timeline,
		Risks:        risks,
		Dependencies: []string{},
		Status:       StatusDraft,
		Version:      1,
		Changelog:    "Initial creation",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Persist base master plan (for backward compatibility)
	base := MasterPlan{
		ID:        plan.ID,
		ProjectID: plan.ProjectID,
		Title:     plan.Title,
		Status:    plan.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := g.storeRepo.CreateMasterPlan(base); err != nil {
		return nil, fmt.Errorf("create base master plan: %w", err)
	}

	// Record evolution event
	evt := PlanEvolutionEvent{
		ID:          domain.NewID("evt"),
		ProjectID:   plan.ProjectID,
		EntityType:  "master_plan",
		EntityID:    plan.ID,
		EventType:   string(ChangeCreated),
		Description: fmt.Sprintf("Created master plan v%d: %s", plan.Version, plan.Title),
		Details:     fmt.Sprintf(`{"version":%d,"title":%q}`, plan.Version, plan.Title),
		CreatedAt:   now,
	}
	if _, err := g.v2Repo.CreateEvolutionEvent(evt); err != nil {
		return nil, fmt.Errorf("record evolution: %w", err)
	}

	return plan, nil
}

// CreateNewVersion creates a new version of an existing master plan.
func (g *MasterPlanGenerator) CreateNewVersion(planID string, changes []string) (*MasterPlanV2, error) {
	// Get latest version
	latestRaw, err := g.v2Repo.GetLatestVersion(planID)
	if err != nil {
		return nil, fmt.Errorf("get latest version: %w", err)
	}
	_ = latestRaw

	// For this implementation, we create a new version by fetching existing plan
	existing, err := g.storeRepo.GetMasterPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("get master plan: %w", err)
	}

	now := g.now()
	version := existing.Version + 1
	changelog := strings.Join(changes, "; ")
	if changelog == "" {
		changelog = fmt.Sprintf("Version %d update", version)
	}

	v2 := &MasterPlanV2{
		ID:        existing.ID,
		ProjectID: existing.ProjectID,
		Title:     existing.Title,
		Status:    existing.Status,
		Version:   version,
		Changelog: changelog,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: now,
	}

	// Record evolution
	evt := PlanEvolutionEvent{
		ID:          domain.NewID("evt"),
		ProjectID:   v2.ProjectID,
		EntityType:  "master_plan",
		EntityID:    v2.ID,
		EventType:   string(ChangeUpdated),
		Description: fmt.Sprintf("Updated to v%d: %s", version, changelog),
		Details:     fmt.Sprintf(`{"version":%d,"changelog":%q}`, version, changelog),
		CreatedAt:   now,
	}
	if _, err := g.v2Repo.CreateEvolutionEvent(evt); err != nil {
		return nil, fmt.Errorf("record evolution: %w", err)
	}

	return v2, nil
}

// SubmitForApproval submits a master plan for approval.
func (g *MasterPlanGenerator) SubmitForApproval(planID, projectID string) error {
	now := g.now()
	evt := PlanEvolutionEvent{
		ID:          domain.NewID("evt"),
		ProjectID:   projectID,
		EntityType:  "master_plan",
		EntityID:    planID,
		EventType:   string(ChangeUpdated),
		Description: "Submitted for approval",
		Details:     `{"action":"submitted_for_approval"}`,
		CreatedAt:   now,
	}
	_, err := g.v2Repo.CreateEvolutionEvent(evt)
	return err
}

// ApprovePlan marks a master plan as approved.
func (g *MasterPlanGenerator) ApprovePlan(planID, projectID, approvedBy, feedback string) error {
	now := g.now()

	evt := PlanEvolutionEvent{
		ID:          domain.NewID("evt"),
		ProjectID:   projectID,
		EntityType:  "master_plan",
		EntityID:    planID,
		EventType:   string(ChangeApproved),
		Description: fmt.Sprintf("Approved by %s", approvedBy),
		Details:     fmt.Sprintf(`{"approved_by":%q,"feedback":%q}`, approvedBy, feedback),
		CreatedAt:   now,
	}
	_, err := g.v2Repo.CreateEvolutionEvent(evt)
	return err
}

// GetEvolutionHistory returns the evolution history for a project.
func (g *MasterPlanGenerator) GetEvolutionHistory(projectID string) ([]PlanEvolutionEvent, error) {
	return g.v2Repo.ListEvolutionEvents(projectID)
}

// RenderSummary produces a concise markdown summary of a master plan.
func RenderMasterPlanSummary(plan *MasterPlanV2) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s (v%d)\n", plan.Title, plan.Version))
	b.WriteString(fmt.Sprintf("Status: %s | Created: %s\n", plan.Status, plan.CreatedAt.Format("2006-01-02")))
	if plan.Description != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", plan.Description))
	}

	if len(plan.Phases) > 0 {
		b.WriteString("\n### Phases\n")
		for _, p := range plan.Phases {
			b.WriteString(fmt.Sprintf("- **%s**: %s (%d weeks)\n", p.Name, p.Description, p.EstimatedWeeks))
		}
	}

	if plan.Timeline.TotalWeeks > 0 {
		b.WriteString(fmt.Sprintf("\n### Timeline\nTotal: %d weeks\n", plan.Timeline.TotalWeeks))
		for _, m := range plan.Timeline.Milestones {
			b.WriteString(fmt.Sprintf("- Week %d: %s\n", m.Week, m.Name))
		}
	}

	if len(plan.Risks) > 0 {
		b.WriteString(fmt.Sprintf("\n### Risks (%d)\n", len(plan.Risks)))
		for _, r := range plan.Risks {
			b.WriteString(fmt.Sprintf("- %s [%s/%s]\n", r.Description, r.Impact, r.Likelihood))
		}
	}

	return b.String()
}

// ──────────────────────────────────────────────
// V2StoreAdapter adapts store-level records to the MasterPlanV2Repository
// ──────────────────────────────────────────────

// V2StoreAdapter bridges store records to the MasterPlanV2Repository interface.
// In a full implementation, this would use the concrete store repositories.
// For now it provides a lightweight adapter pattern.
type V2StoreAdapter struct {
	createVersion     func(interface{}) error
	getLatestVersion  func(string) (interface{}, error)
	listVersions      func(string) ([]interface{}, error)
	createChange      func(interface{}) error
	listChanges       func(string) ([]interface{}, error)
	createApproval    func(interface{}) error
	getLatestApproval func(string) (interface{}, error)
	createEvent       func(PlanEvolutionEvent) (PlanEvolutionEvent, error)
	listEvents        func(string) ([]PlanEvolutionEvent, error)
}

// NewV2StoreAdapter creates a new adapter with the given functions.
func NewV2StoreAdapter(
	createVersion func(interface{}) error,
	getLatestVersion func(string) (interface{}, error),
	listVersions func(string) ([]interface{}, error),
	createChange func(interface{}) error,
	listChanges func(string) ([]interface{}, error),
	createApproval func(interface{}) error,
	getLatestApproval func(string) (interface{}, error),
	createEvent func(PlanEvolutionEvent) (PlanEvolutionEvent, error),
	listEvents func(string) ([]PlanEvolutionEvent, error),
) *V2StoreAdapter {
	return &V2StoreAdapter{
		createVersion:     createVersion,
		getLatestVersion:  getLatestVersion,
		listVersions:      listVersions,
		createChange:      createChange,
		listChanges:       listChanges,
		createApproval:    createApproval,
		getLatestApproval: getLatestApproval,
		createEvent:       createEvent,
		listEvents:        listEvents,
	}
}

func (a *V2StoreAdapter) CreateVersion(v interface{}) error { return a.createVersion(v) }
func (a *V2StoreAdapter) GetLatestVersion(id string) (interface{}, error) {
	return a.getLatestVersion(id)
}
func (a *V2StoreAdapter) ListVersions(id string) ([]interface{}, error) { return a.listVersions(id) }
func (a *V2StoreAdapter) CreateChange(c interface{}) error              { return a.createChange(c) }
func (a *V2StoreAdapter) ListChanges(id string) ([]interface{}, error)  { return a.listChanges(id) }
func (a *V2StoreAdapter) CreateApproval(a2 interface{}) error           { return a.createApproval(a2) }
func (a *V2StoreAdapter) GetLatestApproval(id string) (interface{}, error) {
	return a.getLatestApproval(id)
}
func (a *V2StoreAdapter) CreateEvolutionEvent(e PlanEvolutionEvent) (PlanEvolutionEvent, error) {
	return a.createEvent(e)
}
func (a *V2StoreAdapter) ListEvolutionEvents(id string) ([]PlanEvolutionEvent, error) {
	return a.listEvents(id)
}

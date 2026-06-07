package planning

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Specific Plan V2 Types
// ──────────────────────────────────────────────

// DomainKind categorizes specific plans by domain.
type DomainKind string

const (
	DomainBackend  DomainKind = "backend"
	DomainFrontend DomainKind = "frontend"
	DomainData     DomainKind = "data"
	DomainInfra    DomainKind = "infrastructure"
	DomainDevOps   DomainKind = "devops"
	DomainTesting  DomainKind = "testing"
	DomainSecurity DomainKind = "security"
	DomainResearch DomainKind = "research"
	DomainGeneral  DomainKind = "general"
)

// RegenerationScope defines the scope of a plan regeneration.
type RegenerationScope string

const (
	RegenFull    RegenerationScope = "full"
	RegenPartial RegenerationScope = "partial"
	RegenTasks   RegenerationScope = "tasks_only"
	RegenRisks   RegenerationScope = "risks_only"
	RegenDeps    RegenerationScope = "dependencies_only"
)

// SpecificPlanV2 extends SpecificPlan with richer fields.
type SpecificPlanV2 struct {
	ID                     string        `json:"id"`
	ProjectID              string        `json:"project_id"`
	MasterPlanID           string        `json:"master_plan_id"`
	Domain                 DomainKind    `json:"domain"`
	Title                  string        `json:"title"`
	Description            string        `json:"description"`
	Tasks                  []TaskDef     `json:"tasks"`
	Dependencies           []DepDef      `json:"dependencies"`
	Risks                  []RiskEntry   `json:"risks"`
	ResearchUsed           []ResearchRef `json:"research_used"`
	KnowledgeUsed          []string      `json:"knowledge_used"`
	ImplementationStrategy string        `json:"implementation_strategy"`
	ValidationCriteria     []string      `json:"validation_criteria"`
	Status                 Status        `json:"status"`
	Version                int           `json:"version"`
	Changelog              string        `json:"changelog"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
}

// TaskDef defines a task within a specific plan.
type TaskDef struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Order              int      `json:"order"`
	Effort             string   `json:"effort"`
	Dependencies       []string `json:"dependencies"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}

// DepDef defines a dependency for a specific plan.
type DepDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Critical    bool   `json:"critical"`
}

// ResearchRef links research to a specific plan.
type ResearchRef struct {
	ResearchID string  `json:"research_id"`
	Section    string  `json:"section"`
	Relevance  float64 `json:"relevance"`
	Summary    string  `json:"summary"`
}

// RegenerationRecord tracks a regeneration event.
type RegenerationRecord struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	PlanID      string            `json:"plan_id"`
	VersionFrom int               `json:"version_from"`
	VersionTo   int               `json:"version_to"`
	Reason      string            `json:"reason"`
	Scope       RegenerationScope `json:"scope"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

// ──────────────────────────────────────────────
// Specific Plan V2 Repository Interface
// ──────────────────────────────────────────────

// SpecificPlanV2Repository provides persistence for specific plan v2 entities.
type SpecificPlanV2Repository interface {
	CreateVersion(record interface{}) error
	GetLatestVersion(planID string) (interface{}, error)
	ListVersions(planID string) ([]interface{}, error)

	CreateResearchLink(planID, researchID, section string, relevance float64) error
	ListResearchLinks(planID string) ([]interface{}, error)

	CreateRegeneration(record RegenerationRecord) (RegenerationRecord, error)
	ListRegenerations(planID string) ([]RegenerationRecord, error)
}

// ──────────────────────────────────────────────
// Specific Plan Generator
// ──────────────────────────────────────────────

// SpecificPlanGenerator creates and manages specific plans with domain awareness.
type SpecificPlanGenerator struct {
	storeRepo Repository
	v2Repo    SpecificPlanV2Repository
	now       func() time.Time
}

// NewSpecificPlanGenerator creates a new SpecificPlanGenerator.
func NewSpecificPlanGenerator(storeRepo Repository, v2Repo SpecificPlanV2Repository) *SpecificPlanGenerator {
	return &SpecificPlanGenerator{storeRepo: storeRepo, v2Repo: v2Repo, now: time.Now().UTC}
}

// GenerateV2 creates a new versioned specific plan from input.
func (g *SpecificPlanGenerator) GenerateV2(masterPlanID string, input SpecificPlanInput, kind DomainKind, tasks []TaskDef, deps []DepDef, risks []RiskEntry) (*SpecificPlanV2, error) {
	if strings.TrimSpace(masterPlanID) == "" {
		return nil, fmt.Errorf("master plan id is required")
	}
	if strings.TrimSpace(input.ProjectID) == "" {
		return nil, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(input.Goal) == "" {
		return nil, fmt.Errorf("goal is required")
	}

	now := g.now()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = input.Goal
	}

	strategy := strings.TrimSpace(input.ImplementationStrategy)
	if strategy == "" {
		strategy = "Implement using approved requirements and research."
	}

	plan := &SpecificPlanV2{
		ID:                     domain.NewID("spv2"),
		ProjectID:              input.ProjectID,
		MasterPlanID:           masterPlanID,
		Domain:                 kind,
		Title:                  title,
		Description:            input.Goal,
		Tasks:                  tasks,
		Dependencies:           deps,
		Risks:                  risks,
		ResearchUsed:           []ResearchRef{},
		KnowledgeUsed:          input.KnowledgeUsed,
		ImplementationStrategy: strategy,
		ValidationCriteria:     input.ValidationCriteria,
		Status:                 StatusDraft,
		Version:                1,
		Changelog:              "Initial creation",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// Persist base specific plan (backward compatibility)
	base := SpecificPlan{
		ID:           plan.ID,
		ProjectID:    plan.ProjectID,
		MasterPlanID: plan.MasterPlanID,
		Title:        plan.Title,
		Goal:         plan.Description,
		Status:       plan.Status,
		Version:      plan.Version,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := g.storeRepo.CreateSpecificPlan(base); err != nil {
		return nil, fmt.Errorf("create base specific plan: %w", err)
	}

	return plan, nil
}

// AddResearchLink links research to a specific plan for traceability.
func (g *SpecificPlanGenerator) AddResearchLink(planID, researchID, section string, relevance float64) error {
	if relevance < 0 || relevance > 1.0 {
		return fmt.Errorf("relevance must be between 0 and 1")
	}
	return g.v2Repo.CreateResearchLink(planID, researchID, section, relevance)
}

// RegeneratePlan creates a new version of a specific plan.
func (g *SpecificPlanGenerator) RegeneratePlan(planID, reason string, scope RegenerationScope) (*SpecificPlanV2, error) {
	// Get latest version info
	latestRaw, err := g.v2Repo.GetLatestVersion(planID)
	if err != nil {
		return nil, fmt.Errorf("get latest version: %w", err)
	}
	_ = latestRaw

	existing, err := g.storeRepo.GetSpecificPlan(planID)
	if err != nil {
		return nil, fmt.Errorf("get specific plan: %w", err)
	}

	now := g.now()
	versionFrom := existing.Version
	versionTo := existing.Version + 1

	plan := &SpecificPlanV2{
		ID:        existing.ID,
		ProjectID: existing.ProjectID,
		Title:     existing.Title,
		Status:    StatusDraft,
		Version:   versionTo,
		Changelog: fmt.Sprintf("Regenerated: %s (%s)", reason, scope),
		CreatedAt: existing.CreatedAt,
		UpdatedAt: now,
	}

	// Record regeneration
	reg := RegenerationRecord{
		ID:          domain.NewID("regen"),
		ProjectID:   plan.ProjectID,
		PlanID:      planID,
		VersionFrom: versionFrom,
		VersionTo:   versionTo,
		Reason:      reason,
		Scope:       scope,
		Status:      "completed",
		CreatedAt:   now,
	}
	if _, err := g.v2Repo.CreateRegeneration(reg); err != nil {
		return nil, fmt.Errorf("record regeneration: %w", err)
	}

	return plan, nil
}

// GetResearchTrace returns the research lineage for a specific plan.
func (g *SpecificPlanGenerator) GetResearchTrace(planID string) ([]interface{}, error) {
	return g.v2Repo.ListResearchLinks(planID)
}

// RenderSpecificPlanSummary produces a concise summary of a specific plan.
func RenderSpecificPlanSummary(plan *SpecificPlanV2) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s (v%d) [%s]\n", plan.Title, plan.Version, plan.Domain))
	b.WriteString(fmt.Sprintf("Status: %s | Created: %s\n", plan.Status, plan.CreatedAt.Format("2006-01-02")))

	if len(plan.Tasks) > 0 {
		b.WriteString(fmt.Sprintf("\n### Tasks (%d)\n", len(plan.Tasks)))
		for _, t := range plan.Tasks {
			b.WriteString(fmt.Sprintf("- [ ] %s (%s)\n", t.Title, t.Effort))
		}
	}

	if len(plan.Dependencies) > 0 {
		b.WriteString(fmt.Sprintf("\n### Dependencies (%d)\n", len(plan.Dependencies)))
		for _, d := range plan.Dependencies {
			critical := ""
			if d.Critical {
				critical = " [CRITICAL]"
			}
			b.WriteString(fmt.Sprintf("- %s (%s)%s\n", d.Name, d.Type, critical))
		}
	}

	if len(plan.ResearchUsed) > 0 {
		b.WriteString(fmt.Sprintf("\n### Research Used (%d)\n", len(plan.ResearchUsed)))
		for _, r := range plan.ResearchUsed {
			b.WriteString(fmt.Sprintf("- %s (section: %s, relevance: %.0f%%)\n", r.Summary, r.Section, r.Relevance*100))
		}
	}

	return b.String()
}

// ──────────────────────────────────────────────
// Specific V2StoreAdapter
// ──────────────────────────────────────────────

// SpecificV2StoreAdapter bridges store records to the SpecificPlanV2Repository.
type SpecificV2StoreAdapter struct {
	createVersion      func(interface{}) error
	getLatestVersion   func(string) (interface{}, error)
	listVersions       func(string) ([]interface{}, error)
	createResearchLink func(string, string, string, float64) error
	listResearchLinks  func(string) ([]interface{}, error)
	createRegeneration func(RegenerationRecord) (RegenerationRecord, error)
	listRegenerations  func(string) ([]RegenerationRecord, error)
}

// NewSpecificV2StoreAdapter creates a new SpecificV2StoreAdapter.
func NewSpecificV2StoreAdapter(
	createVersion func(interface{}) error,
	getLatestVersion func(string) (interface{}, error),
	listVersions func(string) ([]interface{}, error),
	createResearchLink func(string, string, string, float64) error,
	listResearchLinks func(string) ([]interface{}, error),
	createRegeneration func(RegenerationRecord) (RegenerationRecord, error),
	listRegenerations func(string) ([]RegenerationRecord, error),
) *SpecificV2StoreAdapter {
	return &SpecificV2StoreAdapter{
		createVersion:      createVersion,
		getLatestVersion:   getLatestVersion,
		listVersions:       listVersions,
		createResearchLink: createResearchLink,
		listResearchLinks:  listResearchLinks,
		createRegeneration: createRegeneration,
		listRegenerations:  listRegenerations,
	}
}

func (a *SpecificV2StoreAdapter) CreateVersion(v interface{}) error { return a.createVersion(v) }
func (a *SpecificV2StoreAdapter) GetLatestVersion(id string) (interface{}, error) {
	return a.getLatestVersion(id)
}
func (a *SpecificV2StoreAdapter) ListVersions(id string) ([]interface{}, error) {
	return a.listVersions(id)
}
func (a *SpecificV2StoreAdapter) CreateResearchLink(planID, researchID, section string, relevance float64) error {
	return a.createResearchLink(planID, researchID, section, relevance)
}
func (a *SpecificV2StoreAdapter) ListResearchLinks(id string) ([]interface{}, error) {
	return a.listResearchLinks(id)
}
func (a *SpecificV2StoreAdapter) CreateRegeneration(r RegenerationRecord) (RegenerationRecord, error) {
	return a.createRegeneration(r)
}
func (a *SpecificV2StoreAdapter) ListRegenerations(id string) ([]RegenerationRecord, error) {
	return a.listRegenerations(id)
}

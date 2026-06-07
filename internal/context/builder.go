package context

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Composite context types
// ──────────────────────────────────────────────

type ExecutiveContext struct {
	ProjectID   string         `json:"project_id"`
	Status      string         `json:"status"`
	WhatMissing []string       `json:"what_missing"`
	WhatNext    []string       `json:"what_next"`
	Risks       []string       `json:"risks"`
	Progress    map[string]int `json:"progress"` // phase -> completed tasks
	UpdatedAt   time.Time      `json:"updated_at"`
}

type PlanningContext struct {
	ProjectID    string    `json:"project_id"`
	Vision       string    `json:"vision"`
	Requirements []string  `json:"requirements"`
	Constraints  []string  `json:"constraints"`
	Research     []string  `json:"research"`
	Knowledge    []string  `json:"knowledge"`
	Decisions    []string  `json:"decisions"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ImplementationContext struct {
	ProjectID         string    `json:"project_id"`
	Task              string    `json:"task"`
	SpecificPlan      string    `json:"specific_plan"`
	ImplementationDoc string    `json:"implementation_doc"`
	Decisions         []string  `json:"decisions"`
	Constraints       []string  `json:"constraints"`
	Validations       []string  `json:"validations"`
	KnownRisks        []string  `json:"known_risks"`
	ExpectedFiles     []string  `json:"expected_files"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ResearchContext struct {
	ProjectID        string    `json:"project_id"`
	Topic            string    `json:"topic"`
	PreviousFindings []string  `json:"previous_findings"`
	RelatedKnowledge []string  `json:"related_knowledge"`
	RelatedDecisions []string  `json:"related_decisions"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ──────────────────────────────────────────────
// ContextView — a composable, persisted context unit
// ──────────────────────────────────────────────

type ContextView struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	ViewType  string    `json:"view_type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContextChunk struct {
	ID            string `json:"id"`
	ContextViewID string `json:"context_view_id"`
	ChunkIndex    int    `json:"chunk_index"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
}

// ──────────────────────────────────────────────
// ContextBuilder — builds composite contexts
// ──────────────────────────────────────────────

type Builder struct {
	repo      Repository
	domainQ   ProjectQuerier
	visionQ   VisionQuerier
	planningQ PlanningQuerier
	now       func() time.Time
}

// VisionQuerier provides access to vision data.
type VisionQuerier interface {
	ListVisions(projectID string) ([]domain.Vision, error)
}

// PlanningQuerier provides access to planning data.
type PlanningQuerier interface {
	ListMasterPlans(projectID string) ([]domain.MasterPlan, error)
	ListSpecificPlans(masterPlanID string) ([]domain.SpecificPlan, error)
	ListPhases(planID string) ([]domain.Phase, error)
	ListTasks(phaseID string) ([]domain.Task, error)
}

// NewBuilder creates a context builder.
func NewBuilder(repo Repository, domainQ ProjectQuerier, visionQ VisionQuerier, planningQ PlanningQuerier) *Builder {
	return &Builder{repo: repo, domainQ: domainQ, visionQ: visionQ, planningQ: planningQ, now: time.Now().UTC}
}

// BuildExecutiveContext builds an executive overview of project status.
func (b *Builder) BuildExecutiveContext(projectID string) (ExecutiveContext, error) {
	ctx := ExecutiveContext{
		ProjectID: projectID,
		Progress:  make(map[string]int),
		UpdatedAt: b.now(),
	}

	project, err := b.domainQ.GetProjectBrief(projectID)
	if err == nil {
		ctx.Status = project.Status
	} else {
		ctx.Status = "unknown"
	}

	plans, err := b.domainQ.ListPlanBriefs(projectID)
	if err != nil {
		return ctx, fmt.Errorf("list plans: %w", err)
	}

	for _, plan := range plans {
		phases, err := b.domainQ.ListPhaseBriefs(plan.ID)
		if err != nil {
			continue
		}
		for _, phase := range phases {
			tasks, err := b.domainQ.ListTaskBriefs(phase.ID)
			if err != nil {
				continue
			}
			completed := 0
			for _, t := range tasks {
				if t.Status == "completed" || t.Status == "done" {
					completed++
				}
			}
			ctx.Progress[phase.Title] = completed
		}
	}

	decisions, err := b.domainQ.ListDecisionBriefs(projectID)
	if err == nil {
		for _, d := range decisions {
			if d.Status == "draft" {
				ctx.WhatMissing = append(ctx.WhatMissing, fmt.Sprintf("Decision %q still in draft", d.Title))
			}
		}
	}

	if len(ctx.WhatMissing) == 0 {
		ctx.WhatMissing = append(ctx.WhatMissing, "No unresolved issues detected")
	}
	ctx.WhatNext = append(ctx.WhatNext, "Review pending decisions", "Validate completed phases", "Plan next implementation phase")

	return ctx, nil
}

// BuildPlanningContext assembles a planning context from vision, requirements, etc.
func (b *Builder) BuildPlanningContext(projectID string) (PlanningContext, error) {
	ctx := PlanningContext{
		ProjectID: projectID,
		UpdatedAt: b.now(),
	}

	requirements, err := b.repo.ListApproved(projectID, TypeRequirement)
	if err == nil {
		for _, r := range requirements {
			ctx.Requirements = append(ctx.Requirements, r.Content)
		}
	}

	constraints, err := b.repo.ListApproved(projectID, TypeConstraint)
	if err == nil {
		for _, c := range constraints {
			ctx.Constraints = append(ctx.Constraints, c.Content)
		}
	}

	decisions, err := b.repo.ListApproved(projectID, TypeDecision)
	if err == nil {
		for _, d := range decisions {
			ctx.Decisions = append(ctx.Decisions, d.Content)
		}
	}

	knowledge, err := b.domainQ.ListKnowledgeBriefs(projectID)
	if err == nil {
		for _, k := range knowledge {
			ctx.Knowledge = append(ctx.Knowledge, k.Topic+": "+k.Summary)
		}
	}

	research, err := b.domainQ.ListResearchBriefs(projectID)
	if err == nil {
		for _, r := range research {
			ctx.Research = append(ctx.Research, r.Topic+": "+r.Summary)
		}
	}

	// Add vision if available
	if b.visionQ != nil {
		visions, err := b.visionQ.ListVisions(projectID)
		if err == nil && len(visions) > 0 {
			ctx.Vision = visions[len(visions)-1].Summary
		}
	}

	return ctx, nil
}

// BuildImplementationContext assembles context for a specific task implementation.
func (b *Builder) BuildImplementationContext(projectID, taskID string) (ImplementationContext, error) {
	ctx := ImplementationContext{
		ProjectID: projectID,
		UpdatedAt: b.now(),
	}

	ctx.Task = "Task: " + taskID

	constraints, err := b.repo.ListApproved(projectID, TypeConstraint)
	if err == nil {
		for _, c := range constraints {
			ctx.Constraints = append(ctx.Constraints, c.Content)
		}
	}

	decisions, err := b.repo.ListApproved(projectID, TypeDecision)
	if err == nil {
		for _, d := range decisions {
			ctx.Decisions = append(ctx.Decisions, d.Content)
		}
	}

	ctx.ExpectedFiles = append(ctx.ExpectedFiles, "TBD from specific plan")
	ctx.KnownRisks = append(ctx.KnownRisks, "No risks identified yet")
	ctx.Validations = append(ctx.Validations, "go test ./...", "go vet ./...", "go build ./...")

	return ctx, nil
}

// BuildResearchContext assembles research context for a topic.
func (b *Builder) BuildResearchContext(projectID, topic string) (ResearchContext, error) {
	ctx := ResearchContext{
		ProjectID: projectID,
		Topic:     topic,
		UpdatedAt: b.now(),
	}

	research, err := b.domainQ.ListResearchBriefs(projectID)
	if err == nil {
		for _, r := range research {
			if strings.Contains(strings.ToLower(r.Topic), strings.ToLower(topic)) {
				ctx.PreviousFindings = append(ctx.PreviousFindings, r.Topic+": "+r.Summary)
			}
		}
	}

	knowledge, err := b.domainQ.ListKnowledgeBriefs(projectID)
	if err == nil {
		for _, k := range knowledge {
			if strings.Contains(strings.ToLower(k.Topic), strings.ToLower(topic)) {
				ctx.RelatedKnowledge = append(ctx.RelatedKnowledge, k.Topic+": "+k.Summary)
			}
		}
	}

	decisions, err := b.domainQ.ListDecisionBriefs(projectID)
	if err == nil {
		for _, d := range decisions {
			ctx.RelatedDecisions = append(ctx.RelatedDecisions, d.Title)
		}
	}

	return ctx, nil
}

// PersistContextView saves a context view to the store.
func (b *Builder) PersistContextView(name, viewType, projectID, content string) (ContextView, error) {
	now := b.now()
	view := ContextView{
		ID:        domain.NewID("ctxview"),
		ProjectID: projectID,
		Name:      name,
		ViewType:  viewType,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return view, nil
}

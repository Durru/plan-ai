package context

import (
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Context Delivery Engine Types
// ──────────────────────────────────────────────

// DeliveryLevel defines the granularity level of context delivery.
type DeliveryLevel string

const (
	LevelExecutive      DeliveryLevel = "L0_executive"
	LevelPlanning       DeliveryLevel = "L1_planning"
	LevelImplementation DeliveryLevel = "L2_implementation"
	LevelResearch       DeliveryLevel = "L3_research"
	LevelApproval       DeliveryLevel = "L4_approval"
)

// DeliveryStatus defines the status of a delivery session.
type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "pending"
	DeliveryInProgress DeliveryStatus = "in_progress"
	DeliveryDelivered  DeliveryStatus = "delivered"
	DeliveryFailed     DeliveryStatus = "failed"
)

// BudgetStrategy defines how context budgets are managed.
type BudgetStrategy string

const (
	BudgetFixed      BudgetStrategy = "fixed"
	BudgetDynamic    BudgetStrategy = "dynamic"
	BudgetPercentage BudgetStrategy = "percentage"
)

// DeliverySession represents a context delivery session.
type DeliverySession struct {
	ID           string                 `json:"id"`
	ProjectID    string                 `json:"project_id"`
	Level        DeliveryLevel          `json:"level"`
	BudgetTokens int                    `json:"budget_tokens"`
	TokensUsed   int                    `json:"tokens_used"`
	Content      string                 `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	Status       DeliveryStatus         `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

// DeliveryBudget represents token budget configuration for a level.
type DeliveryBudget struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	Level        DeliveryLevel  `json:"level"`
	MaxTokens    int            `json:"max_tokens"`
	CurrentUsage int            `json:"current_usage"`
	Strategy     BudgetStrategy `json:"strategy"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// DeliveryUsage represents a token usage record.
type DeliveryUsage struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	SessionID string    `json:"session_id"`
	Level     string    `json:"level"`
	Tokens    int       `json:"tokens"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// ──────────────────────────────────────────────
// Delivery Repository Interface
// ──────────────────────────────────────────────

// DeliveryRepository provides persistence for context delivery entities.
type DeliveryRepository interface {
	CreateSession(session DeliverySession) (DeliverySession, error)
	ListSessions(projectID string, level string, limit int) ([]DeliverySession, error)

	CreateUsage(usage DeliveryUsage) (DeliveryUsage, error)
	GetTotalUsage(projectID string) (int, error)

	CreateOrUpdateBudget(budget DeliveryBudget) (DeliveryBudget, error)
	GetBudget(projectID string, level DeliveryLevel) (DeliveryBudget, error)
}

// ──────────────────────────────────────────────
// Querier interface for the Delivery Engine
// ──────────────────────────────────────────────

// ExecutiveData provides data for L0 executive context.
type ExecutiveData interface {
	GetProjectBrief(projectID string) (domain.Project, error)
	ListPlanBriefs(projectID string) ([]domain.MasterPlan, error)
	ListDecisionBriefs(projectID string) ([]domain.Decision, error)
}

// PlanningData provides data for L1 planning context.
type PlanningData interface {
	ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error)
	ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error)
	ListResearchBriefs(projectID string) ([]ResearchBrief, error)
	ListVisions(projectID string) ([]domain.Vision, error)
}

// ImplementationData provides data for L2 implementation context.
type ImplementationData interface {
	ListApproved(projectID string, itemType ApprovedType) ([]ApprovedItem, error)
	GetSpecificPlan(planID string) (domain.SpecificPlan, error)
	ListTasks(phaseID string) ([]domain.Task, error)
}

// ResearchData provides data for L3 research context.
type ResearchData interface {
	ListResearchBriefs(projectID string) ([]ResearchBrief, error)
	ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error)
	ListDecisionBriefs(projectID string) ([]domain.Decision, error)
}

// ──────────────────────────────────────────────
// Context Delivery Engine
// ──────────────────────────────────────────────

// DeliveryEngine delivers context at the right granularity with budget awareness.
type DeliveryEngine struct {
	repo           DeliveryRepository
	execData       ExecutiveData
	planData       PlanningData
	implData       ImplementationData
	researchData   ResearchData
	defaultBudgets map[DeliveryLevel]int
	now            func() time.Time
}

// NewDeliveryEngine creates a new ContextDeliveryEngine.
func NewDeliveryEngine(
	repo DeliveryRepository,
	execData ExecutiveData,
	planData PlanningData,
	implData ImplementationData,
	researchData ResearchData,
) *DeliveryEngine {
	return &DeliveryEngine{
		repo:         repo,
		execData:     execData,
		planData:     planData,
		implData:     implData,
		researchData: researchData,
		defaultBudgets: map[DeliveryLevel]int{
			LevelExecutive:      2048,
			LevelPlanning:       4096,
			LevelImplementation: 8192,
			LevelResearch:       4096,
			LevelApproval:       2048,
		},
		now: time.Now().UTC,
	}
}

// DeliverContext builds and delivers context at the requested level.
func (e *DeliveryEngine) DeliverContext(projectID string, level DeliveryLevel, params map[string]string) (*DeliverySession, error) {
	budget, err := e.getOrCreateBudget(projectID, level)
	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	if budget.CurrentUsage >= budget.MaxTokens {
		return nil, fmt.Errorf("budget exhausted for level %s (%d/%d tokens)", level, budget.CurrentUsage, budget.MaxTokens)
	}

	var content string
	var tokens int

	switch level {
	case LevelExecutive:
		content, tokens, err = e.buildExecutiveContext(projectID, budget.MaxTokens-budget.CurrentUsage)
	case LevelPlanning:
		content, tokens, err = e.buildPlanningContext(projectID, budget.MaxTokens-budget.CurrentUsage)
	case LevelImplementation:
		taskID := params["task_id"]
		content, tokens, err = e.buildImplementationContext(projectID, taskID, budget.MaxTokens-budget.CurrentUsage)
	case LevelResearch:
		topic := params["topic"]
		content, tokens, err = e.buildResearchContext(projectID, topic, budget.MaxTokens-budget.CurrentUsage)
	case LevelApproval:
		content, tokens, err = e.buildApprovalContext(projectID, budget.MaxTokens-budget.CurrentUsage)
	default:
		return nil, fmt.Errorf("unknown delivery level: %s", level)
	}
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	session := DeliverySession{
		ID:           domain.NewID("cds"),
		ProjectID:    projectID,
		Level:        level,
		BudgetTokens: budget.MaxTokens,
		TokensUsed:   tokens,
		Content:      content,
		Metadata:     map[string]interface{}{"params": params},
		Status:       DeliveryDelivered,
		CreatedAt:    e.now(),
	}
	if _, err := e.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	// Record usage
	usage := DeliveryUsage{
		ID:        domain.NewID("cdu"),
		ProjectID: projectID,
		SessionID: session.ID,
		Level:     string(level),
		Tokens:    tokens,
		Source:    params["source"],
		CreatedAt: e.now(),
	}
	if _, err := e.repo.CreateUsage(usage); err != nil {
		return nil, fmt.Errorf("record usage: %w", err)
	}

	return &session, nil
}

// getOrCreateBudget retrieves or creates a default budget for a level.
func (e *DeliveryEngine) getOrCreateBudget(projectID string, level DeliveryLevel) (DeliveryBudget, error) {
	budget, err := e.repo.GetBudget(projectID, level)
	if err != nil || budget.ID == "" {
		// Create default budget
		maxTokens, ok := e.defaultBudgets[level]
		if !ok {
			maxTokens = 4096
		}
		budget = DeliveryBudget{
			ID:        domain.NewID("cdb"),
			ProjectID: projectID,
			Level:     level,
			MaxTokens: maxTokens,
			Strategy:  BudgetFixed,
			CreatedAt: e.now(),
			UpdatedAt: e.now(),
		}
		return e.repo.CreateOrUpdateBudget(budget)
	}
	return budget, nil
}

// ──────────────────────────────────────────────
// Level Builders
// ──────────────────────────────────────────────

func (e *DeliveryEngine) buildExecutiveContext(projectID string, budget int) (string, int, error) {
	var b strings.Builder

	project, err := e.execData.GetProjectBrief(projectID)
	if err == nil {
		b.WriteString(fmt.Sprintf("# %s\n\n", project.Name))
		b.WriteString(fmt.Sprintf("Status: %s\n", project.Status))
	}

	plans, err := e.execData.ListPlanBriefs(projectID)
	if err == nil {
		b.WriteString(fmt.Sprintf("\n## Plans (%d)\n", len(plans)))
		for _, plan := range plans {
			b.WriteString(fmt.Sprintf("- %s [%s]\n", plan.Title, plan.Status))
		}
	}

	decisions, err := e.execData.ListDecisionBriefs(projectID)
	if err == nil {
		pending := 0
		for _, d := range decisions {
			if d.Status == "draft" {
				pending++
			}
		}
		b.WriteString(fmt.Sprintf("\n## Pending Decisions: %d\n", pending))
	}

	result := b.String()
	tokens := estimateTokens(result)
	return result, tokens, nil
}

func (e *DeliveryEngine) buildPlanningContext(projectID string, budget int) (string, int, error) {
	var b strings.Builder

	b.WriteString("# Planning Context\n\n")

	requirements, err := e.planData.ListApproved(projectID, TypeRequirement)
	if err == nil && len(requirements) > 0 {
		b.WriteString(fmt.Sprintf("## Requirements (%d)\n", len(requirements)))
		for _, r := range requirements {
			b.WriteString(fmt.Sprintf("- %s\n", r.Content))
		}
	}

	constraints, err := e.planData.ListApproved(projectID, TypeConstraint)
	if err == nil && len(constraints) > 0 {
		b.WriteString(fmt.Sprintf("\n## Constraints (%d)\n", len(constraints)))
		for _, c := range constraints {
			b.WriteString(fmt.Sprintf("- %s\n", c.Content))
		}
	}

	decisions, err := e.planData.ListApproved(projectID, TypeDecision)
	if err == nil && len(decisions) > 0 {
		b.WriteString(fmt.Sprintf("\n## Decisions (%d)\n", len(decisions)))
		for _, d := range decisions {
			b.WriteString(fmt.Sprintf("- %s\n", d.Content))
		}
	}

	knowledge, err := e.planData.ListKnowledgeBriefs(projectID)
	if err == nil && len(knowledge) > 0 {
		b.WriteString(fmt.Sprintf("\n## Knowledge (%d)\n", len(knowledge)))
		for _, k := range knowledge {
			b.WriteString(fmt.Sprintf("- %s: %s\n", k.Topic, k.Summary))
		}
	}

	// Avoid duplicate research in planning context
	research, err := e.planData.ListResearchBriefs(projectID)
	if err == nil && len(research) > 0 {
		b.WriteString(fmt.Sprintf("\n## Available Research (%d)\n", len(research)))
		for _, r := range research {
			b.WriteString(fmt.Sprintf("- %s: %s\n", r.Topic, r.Summary))
		}
	}

	result := b.String()
	tokens := estimateTokens(result)
	return result, tokens, nil
}

func (e *DeliveryEngine) buildImplementationContext(projectID, taskID string, budget int) (string, int, error) {
	var b strings.Builder

	b.WriteString("# Implementation Context\n\n")

	if taskID != "" {
		b.WriteString(fmt.Sprintf("Task: %s\n\n", taskID))
	}

	constraints, err := e.implData.ListApproved(projectID, TypeConstraint)
	if err == nil && len(constraints) > 0 {
		b.WriteString("## Constraints\n")
		for _, c := range constraints {
			b.WriteString(fmt.Sprintf("- %s\n", c.Content))
		}
	}

	decisions, err := e.implData.ListApproved(projectID, TypeDecision)
	if err == nil && len(decisions) > 0 {
		b.WriteString("\n## Decisions\n")
		for _, d := range decisions {
			b.WriteString(fmt.Sprintf("- %s\n", d.Content))
		}
	}

	result := b.String()
	tokens := estimateTokens(result)
	return result, tokens, nil
}

func (e *DeliveryEngine) buildResearchContext(projectID, topic string, budget int) (string, int, error) {
	var b strings.Builder

	b.WriteString("# Research Context\n\n")

	if topic != "" {
		b.WriteString(fmt.Sprintf("Topic: %s\n\n", topic))
	}

	research, err := e.researchData.ListResearchBriefs(projectID)
	if err == nil {
		b.WriteString(fmt.Sprintf("## Previous Research (%d)\n", len(research)))
		for _, r := range research {
			if topic == "" || strings.Contains(strings.ToLower(r.Topic), strings.ToLower(topic)) {
				b.WriteString(fmt.Sprintf("- %s: %s\n", r.Topic, r.Summary))
			}
		}
	}

	knowledge, err := e.researchData.ListKnowledgeBriefs(projectID)
	if err == nil {
		b.WriteString("\n## Related Knowledge\n")
		for _, k := range knowledge {
			if topic == "" || strings.Contains(strings.ToLower(k.Topic), strings.ToLower(topic)) {
				b.WriteString(fmt.Sprintf("- %s: %s\n", k.Topic, k.Summary))
			}
		}
	}

	result := b.String()
	tokens := estimateTokens(result)
	return result, tokens, nil
}

func (e *DeliveryEngine) buildApprovalContext(projectID string, budget int) (string, int, error) {
	var b strings.Builder

	b.WriteString("# Approval Context\n\n")

	decisions, err := e.execData.ListDecisionBriefs(projectID)
	if err == nil {
		pending := 0
		for _, d := range decisions {
			if d.Status == "draft" {
				b.WriteString(fmt.Sprintf("- [PENDING] Decision: %s\n", d.Title))
				pending++
			}
		}
		if pending == 0 {
			b.WriteString("No pending decisions.\n")
		}
	}

	result := b.String()
	tokens := estimateTokens(result)
	return result, tokens, nil
}

// ──────────────────────────────────────────────
// Budget Management
// ──────────────────────────────────────────────

// SetBudget sets or updates a budget for a level.
func (e *DeliveryEngine) SetBudget(projectID string, level DeliveryLevel, maxTokens int, strategy BudgetStrategy) (DeliveryBudget, error) {
	budget := DeliveryBudget{
		ID:        domain.NewID("cdb"),
		ProjectID: projectID,
		Level:     level,
		MaxTokens: maxTokens,
		Strategy:  strategy,
		CreatedAt: e.now(),
		UpdatedAt: e.now(),
	}
	return e.repo.CreateOrUpdateBudget(budget)
}

// GetUsageSummary returns token usage summary for a project.
func (e *DeliveryEngine) GetUsageSummary(projectID string) (map[string]int, error) {
	total, err := e.repo.GetTotalUsage(projectID)
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"total_tokens": total,
	}, nil
}

// estimateTokens estimates token count from text (rough: ~4 chars per token).
func estimateTokens(text string) int {
	return len([]rune(text)) / 4
}

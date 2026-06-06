package agent

import (
	"fmt"
)

// Service is the main agent orchestrator.
type Service struct {
	detector        IntentDetector
	router          Router
	contextLoader   ContextLoader
	delegator       Delegator
	responseBuilder ResponseBuilder
	runRepo         AgentRunRepository
	planningGuard   PlanningGuard
}

// PlanningGuard checks whether planning can proceed for a project.
// A nil guard allows planning without checks (for backward compat).
type PlanningGuard interface {
	IsPlanningAllowed(projectID string) (ok bool, reason string)
}

// NewService creates a new agent service.
func NewService(
	detector IntentDetector,
	router Router,
	contextLoader ContextLoader,
	delegator Delegator,
	responseBuilder ResponseBuilder,
	runRepo AgentRunRepository,
) *Service {
	return &Service{
		detector:        detector,
		router:          router,
		contextLoader:   contextLoader,
		delegator:       delegator,
		responseBuilder: responseBuilder,
		runRepo:         runRepo,
	}
}

// SetPlanningGuard attaches an optional planning guard. When set, the service
// blocks planning intents (create_master_plan, create_specific_plan) until
// the guard is satisfied.
func (s *Service) SetPlanningGuard(g PlanningGuard) {
	s.planningGuard = g
}

// ProcessMessage handles a user message and returns an agent response.
func (s *Service) ProcessMessage(projectID, userInput string) (AgentResponse, error) {
	if projectID == "" {
		return s.responseBuilder.BuildError("no project context available"), nil
	}

	// 1. Detect intent
	intent := s.detector.DetectIntent(userInput)
	if intent == IntentUnknown {
		return s.responseBuilder.BuildSuccess(
			"I'm not sure what you'd like to do. You can ask me to create a plan, research a topic, check project status, or request changes.",
			RouterDecision{Workflow: WorkflowStatus, Capability: CapabilityContext, ContextKeys: []string{}},
		), nil
	}

	// Guard planning intents
	if s.planningGuard != nil && (intent == IntentCreateMasterPlan || intent == IntentCreateSpecificPlan) {
		if ok, reason := s.planningGuard.IsPlanningAllowed(projectID); !ok {
			return s.responseBuilder.BuildError(reason), nil
		}
	}

	// 2. Load minimal context
	decision := s.router.Route(intent, ContextPayload{ProjectID: projectID})
	ctx, err := s.contextLoader.Load(projectID, decision.ContextKeys)
	if err != nil {
		return s.responseBuilder.BuildError(fmt.Sprintf("loading context: %v", err)), nil
	}

	// 3. Re-route with full context
	decision = s.router.Route(intent, ctx)

	// 4. Create delegation if needed
	if decision.Workflow != WorkflowStatus && decision.Workflow != WorkflowApproval {
		job := JobForIntent(projectID, intent, decision.Workflow)
		_, err := s.delegator.CreateJob(job)
		if err != nil {
			return s.responseBuilder.BuildError(fmt.Sprintf("creating job: %v", err)), nil
		}
	}

	// 5. Record the agent run
	run := AgentRunRecord{
		ID:        fmt.Sprintf("run_%d", len(ctx.Plans)+len(ctx.Tasks)),
		ProjectID: projectID,
		Intent:    string(intent),
		Status:    "processed",
	}
	if _, err := s.runRepo.CreateRun(run); err != nil {
		return s.responseBuilder.BuildError(fmt.Sprintf("recording run: %v", err)), nil
	}

	// 6. Build response based on decision
	message := s.buildMessage(intent, decision, ctx)
	if decision.RequiresApproval {
		return s.responseBuilder.BuildApprovalRequired(message, decision), nil
	}
	return s.responseBuilder.BuildSuccess(message, decision), nil
}

func (s *Service) buildMessage(intent IntentKind, decision RouterDecision, ctx ContextPayload) string {
	switch intent {
	case IntentCreateMasterPlan:
		reqCount := len(ctx.Approved.Requirements)
		decCount := len(ctx.Approved.Decisions)
		if reqCount == 0 && decCount == 0 {
			return "I'd like to create a master plan, but I don't see any approved requirements or decisions yet. Please add some approved context first (use 'plan-ai approved add')."
		}
		return fmt.Sprintf("I can create a master plan based on %d approved requirements and %d approved decisions. Shall I proceed?", reqCount, decCount)

	case IntentCreateSpecificPlan:
		planCount := len(ctx.Plans)
		if planCount == 0 {
			return "I'd like to create a specific plan, but I don't see any master plans yet. Please create a master plan first."
		}
		return fmt.Sprintf("I can create a specific implementation plan under one of %d existing master plans. Which plan should I work under?", planCount)

	case IntentResearchTopic:
		return "I'll research this topic and add the findings to your project knowledge."

	case IntentProjectStatus:
		return fmt.Sprintf("Here's your project status: %d plans, %d phases, %d tasks, %d decisions.",
			len(ctx.Plans), len(ctx.Phases), len(ctx.Tasks), len(ctx.Decisions))

	case IntentNextTask:
		return s.buildNextTaskMessage(ctx)

	case IntentChangeRequest:
		return "I'll analyze the impact of this change request and prepare a report."

	case IntentApprove:
		return "Got it. I'll proceed with the approval."

	case IntentReject:
		return "Understood. I'll note the rejection. Can you share why this was rejected?"

	case IntentValidate:
		return "I'll run the validations and report back."

	case IntentUpdatePlan:
		return "I'll prepare the plan update for your review."

	case IntentAnalyzeProject:
		return fmt.Sprintf("Project analysis: %d plans, %d phases, %d tasks, %d decisions, %d knowledge entries. %d approved requirements.",
			len(ctx.Plans), len(ctx.Phases), len(ctx.Tasks), len(ctx.Decisions), len(ctx.Knowledge), len(ctx.Approved.Requirements))

	case IntentCreateProduct:
		if len(ctx.Visions) > 0 && len(ctx.Approved.Requirements) > 0 {
			return fmt.Sprintf("You have %d existing visions and %d approved requirements. I can create a new product intent for your SaaS. What should the product do?", len(ctx.Visions), len(ctx.Approved.Requirements))
		}
		return "I'd like to understand the product you want to create. Tell me about the SaaS or app — what problem does it solve, who is it for, and what are the key features?"

	case IntentDatabasePlan:
		if len(ctx.Decisions) == 0 {
			return "I need some decisions or plans first to design a database. What entity types are involved? Use existing plans as a reference."
		}
		return fmt.Sprintf("I can design a database schema based on %d decisions and %d plans. What persistence strategy are you considering — SQLite, PostgreSQL, or something else?", len(ctx.Decisions), len(ctx.Plans))

	case IntentImpactAnalysis:
		return fmt.Sprintf("Impact analysis: I'll examine %d plans, %d tasks, and %d decisions to assess consequences and risks.", len(ctx.Plans), len(ctx.Tasks), len(ctx.Decisions))

	default:
		return "I understand what you're asking. Let me work on that."
	}
}

func (s *Service) buildNextTaskMessage(ctx ContextPayload) string {
	for _, t := range ctx.Tasks {
		status, _ := t["status"].(string)
		if status == "pending" || status == "draft" {
			title, _ := t["title"].(string)
			id, _ := t["id"].(string)
			return fmt.Sprintf("Your next task is: %s (ID: %s). Would you like me to prepare implementation context for it?", title, id)
		}
	}
	if len(ctx.Plans) == 0 {
		return "No plans found yet. Start by creating a master plan."
	}
	return "All tasks appear to be completed. Great work! Consider creating a new plan or reviewing the project status."
}

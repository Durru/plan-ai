package agent

// DefaultRouter routes intents to workflows, capabilities, and strategies.
type DefaultRouter struct {
	workflowSelector   WorkflowSelector
	capabilitySelector CapabilitySelector
}

// NewRouter creates a new DefaultRouter.
func NewRouter(ws WorkflowSelector, cs CapabilitySelector) *DefaultRouter {
	return &DefaultRouter{
		workflowSelector:   ws,
		capabilitySelector: cs,
	}
}

// Route produces a RouterDecision for the given intent and context.
func (r *DefaultRouter) Route(intent IntentKind, ctx ContextPayload) RouterDecision {
	workflow := r.workflowSelector.Select(intent)
	capability := r.capabilitySelector.Select(intent)

	decision := RouterDecision{
		Workflow:      workflow,
		Capability:    capability,
		ModelStrategy: ModelStrategyDefault,
		ContextKeys:   r.contextKeys(intent),
	}

	// Require approval for destructive or impactful actions
	switch intent {
	case IntentCreateMasterPlan, IntentCreateSpecificPlan:
		decision.RequiresApproval = true
	case IntentChangeRequest:
		decision.RequiresApproval = true
	case IntentApprove:
		decision.RequiresApproval = false
	case IntentReject:
		decision.RequiresApproval = true // double-check rejection
	case IntentUpdatePlan:
		decision.RequiresApproval = true
	default:
		decision.RequiresApproval = false
	}

	return decision
}

func (r *DefaultRouter) contextKeys(intent IntentKind) []string {
	switch intent {
	case IntentCreateMasterPlan:
		return []string{"approved", "visions", "decisions"}
	case IntentCreateSpecificPlan:
		return []string{"master_plans", "decisions", "knowledge"}
	case IntentResearchTopic:
		return []string{"research", "knowledge"}
	case IntentProjectStatus, IntentNextTask:
		return []string{"plans", "phases", "tasks", "decisions"}
	case IntentAnalyzeProject:
		return []string{"plans", "phases", "tasks", "decisions", "visions", "knowledge", "approved"}
	case IntentCreateProduct:
		return []string{"visions", "approved", "knowledge"}
	case IntentDatabasePlan:
		return []string{"knowledge", "decisions", "plans"}
	case IntentImpactAnalysis:
		return []string{"plans", "tasks", "decisions", "validations"}
	case IntentChangeRequest:
		return []string{"plans", "decisions", "tasks"}
	case IntentImplementationHelp:
		return []string{"implementation", "tasks", "decisions"}
	case IntentUpdatePlan:
		return []string{"plans", "phases", "tasks"}
	case IntentApprove:
		return []string{"plans"}
	case IntentReject:
		return []string{"plans"}
	case IntentValidate:
		return []string{"validations", "tasks"}
	default:
		return []string{"status"}
	}
}

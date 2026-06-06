package agent

// DefaultWorkflowSelector maps intents to workflow types.
type DefaultWorkflowSelector struct{}

// NewWorkflowSelector creates a new DefaultWorkflowSelector.
func NewWorkflowSelector() *DefaultWorkflowSelector {
	return &DefaultWorkflowSelector{}
}

// Select returns the workflow type for the given intent.
func (s *DefaultWorkflowSelector) Select(intent IntentKind) string {
	switch intent {
	case IntentCreateMasterPlan, IntentCreateSpecificPlan:
		return WorkflowPlanning
	case IntentResearchTopic:
		return WorkflowResearch
	case IntentApprove, IntentReject:
		return WorkflowApproval
	case IntentChangeRequest:
		return WorkflowImpact
	case IntentProjectStatus, IntentNextTask:
		return WorkflowStatus
	case IntentImplementationHelp, IntentUpdatePlan:
		return WorkflowPlanning
	case IntentValidate:
		return WorkflowApproval
	case IntentAnalyzeProject:
		return WorkflowContext
	case IntentCreateProduct:
		return WorkflowVision
	case IntentDatabasePlan:
		return WorkflowPlanning
	case IntentImpactAnalysis:
		return WorkflowImpact
	default:
		return WorkflowStatus
	}
}

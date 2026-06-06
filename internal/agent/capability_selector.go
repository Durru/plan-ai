package agent

// DefaultCapabilitySelector maps intents to capabilities.
type DefaultCapabilitySelector struct{}

// NewCapabilitySelector creates a new DefaultCapabilitySelector.
func NewCapabilitySelector() *DefaultCapabilitySelector {
	return &DefaultCapabilitySelector{}
}

// Select returns the capability for the given intent.
func (s *DefaultCapabilitySelector) Select(intent IntentKind) string {
	switch intent {
	case IntentCreateMasterPlan, IntentCreateSpecificPlan:
		return CapabilityPlanning
	case IntentResearchTopic:
		return CapabilityResearch
	case IntentApprove, IntentReject, IntentValidate:
		return CapabilityPlanning
	case IntentChangeRequest:
		return CapabilityChange
	case IntentProjectStatus, IntentNextTask:
		return CapabilityContext
	case IntentImplementationHelp, IntentUpdatePlan:
		return CapabilityPlanning
	case IntentAnalyzeProject:
		return CapabilityContext
	case IntentCreateProduct:
		return CapabilityVision
	case IntentDatabasePlan:
		return CapabilityPlanning
	case IntentImpactAnalysis:
		return CapabilityChange
	default:
		return CapabilityContext
	}
}

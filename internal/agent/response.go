package agent

import "fmt"

// DefaultResponseBuilder builds AgentResponse values.
type DefaultResponseBuilder struct{}

// NewResponseBuilder creates a DefaultResponseBuilder.
func NewResponseBuilder() *DefaultResponseBuilder {
	return &DefaultResponseBuilder{}
}

// BuildSuccess creates a success response.
func (b *DefaultResponseBuilder) BuildSuccess(message string, decision RouterDecision) AgentResponse {
	return AgentResponse{
		Message:             message,
		Status:              "success",
		RequiresApproval:    false,
		SuggestedNextAction: suggestNextAction(decision),
		ContextUsed:         decision.ContextKeys,
		WorkflowTriggered:   decision.Workflow,
		CreatedEntities:     map[string]string{},
	}
}

// BuildApprovalRequired creates a response that requires user approval.
func (b *DefaultResponseBuilder) BuildApprovalRequired(message string, decision RouterDecision) AgentResponse {
	return AgentResponse{
		Message:             message,
		Status:              "requires_approval",
		RequiresApproval:    true,
		SuggestedNextAction: "Approve or reject the proposed action",
		ContextUsed:         decision.ContextKeys,
		WorkflowTriggered:   decision.Workflow,
		CreatedEntities:     map[string]string{},
	}
}

// BuildError creates an error response.
func (b *DefaultResponseBuilder) BuildError(err string) AgentResponse {
	return AgentResponse{
		Message:             fmt.Sprintf("I couldn't process that: %s", err),
		Status:              "error",
		RequiresApproval:    false,
		SuggestedNextAction: "Try rephrasing your request or ask for help",
		ContextUsed:         []string{},
		WorkflowTriggered:   "",
		CreatedEntities:     map[string]string{},
	}
}

func suggestNextAction(decision RouterDecision) string {
	switch decision.Workflow {
	case WorkflowPlanning:
		return "I'll prepare the planning artifacts. You'll need to review and approve them."
	case WorkflowResearch:
		return "I'll research the topic and summarize findings."
	case WorkflowApproval:
		return "Please review the details and confirm."
	case WorkflowImpact:
		return "I'll analyze the impact of this change."
	case WorkflowStatus:
		return "Here's the current project status."
	default:
		return "How can I help you next?"
	}
}

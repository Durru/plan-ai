package continuous

import (
	"fmt"
	"strings"
)

// Planner creates plan update proposals from detected events.
type Planner struct {
	repo PlanUpdateProposalRepository
}

// NewPlanner creates a new Planner.
func NewPlanner(repo PlanUpdateProposalRepository) *Planner {
	return &Planner{repo: repo}
}

// CreateProposal creates a plan update proposal from a detected event.
func (p *Planner) CreateProposal(projectID string, event ContinuousEvent, affectedPlans, affectedTasks, affectedDecisions []string) (PlanUpdateProposal, error) {
	requiresApproval := true // proposals always require approval

	proposal := PlanUpdateProposal{
		ID:                fmt.Sprintf("pup_%d", len(affectedPlans)+len(affectedTasks)),
		ProjectID:         projectID,
		Reason:            fmt.Sprintf("Event: %s — %s", event.EventType, event.Summary),
		AffectedPlans:     affectedPlans,
		AffectedTasks:     affectedTasks,
		AffectedDecisions: affectedDecisions,
		SuggestedUpdates:  p.buildSuggestedUpdates(event, affectedPlans, affectedTasks),
		RequiresResearch:  event.EventType == EventNewResearch || event.EventType == EventNewKnowledge,
		RequiresApproval:  requiresApproval,
		Status:            ProposalDraft,
		CreatedAt:         nowUTC(),
		UpdatedAt:         nowUTC(),
	}

	return p.repo.CreateProposal(proposal)
}

func (p *Planner) buildSuggestedUpdates(event ContinuousEvent, affectedPlans, affectedTasks []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Following %s, the following updates are suggested:\n", event.EventType))

	if len(affectedPlans) > 0 {
		b.WriteString(fmt.Sprintf("- Review and update affected plans: %s\n", strings.Join(affectedPlans, ", ")))
	}
	if len(affectedTasks) > 0 {
		b.WriteString(fmt.Sprintf("- Re-evaluate affected tasks: %s\n", strings.Join(affectedTasks, ", ")))
	}
	b.WriteString("- Verify plan coherence after updates")

	return b.String()
}

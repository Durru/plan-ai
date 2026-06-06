// Package guard provides the Phase 5 discovery-first guardrail: no default
// planning path may create a plan without an approved Product Intent.
package guard

import (
	"database/sql"
	"fmt"

	"github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/store"
)

// PlanningGuard checks whether planning can proceed for a project.
type PlanningGuard struct {
	db *sql.DB
}

// NewPlanningGuard creates a guard backed by the given database.
func NewPlanningGuard(db *sql.DB) *PlanningGuard {
	return &PlanningGuard{db: db}
}

// Check returns (true, "") if an approved product intent exists. If none is
// found, it returns (false, usefulDiscoveryQuestion) — a question the user
// should answer before planning can proceed.
func (g *PlanningGuard) Check(projectID string) (ok bool, nextQuestion string) {
	intentRepo := store.NewIntentV3Repository(g.db)
	svc := intentv3.NewService(intentRepo, store.NewIntentV3DiscoveryResultRepository(g.db))

	if svc.IsApprovedProductIntent(projectID) {
		return true, ""
	}

	return false, g.nextDiscoveryQuestion(projectID)
}

func (g *PlanningGuard) nextDiscoveryQuestion(projectID string) string {
	intentRepo := store.NewIntentV3Repository(g.db)
	intents, err := intentRepo.ListProductIntents(projectID)

	if err != nil || len(intents) == 0 {
		return "Before I can create a plan, I need to understand your product intent. What problem does your product solve? Who is it for? What are the key features? You can start with: `plan-ai intent create`"
	}

	counts := map[intentv3.ProductIntentStatus]int{}
	for _, pi := range intents {
		counts[pi.Status]++
	}

	if counts[intentv3.StatusPendingApproval] > 0 {
		return fmt.Sprintf("You have %d product intent(s) pending approval. Please approve one first with: `plan-ai intent approve <id>`", counts[intentv3.StatusPendingApproval])
	}

	if counts[intentv3.StatusDraft] > 0 {
		return fmt.Sprintf("You have %d draft product intent(s). Please submit one for approval with: `plan-ai intent submit <id>`", counts[intentv3.StatusDraft])
	}

	return "No approved product intent found. Create and approve one first with: `plan-ai intent create`"
}

// GuardPlanningInput is the canonical entry point — CLI and MCP handlers
// call this before creating any plan.
func GuardPlanningInput(db *sql.DB, projectID string) error {
	ok, question := NewPlanningGuard(db).Check(projectID)
	if !ok {
		return fmt.Errorf("%s", question)
	}
	return nil
}

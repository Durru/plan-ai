package continuous

import (
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type TargetedRegeneration struct {
	ID                string
	ProjectID         string
	Reason            string
	Scope             string
	AffectedSections  []string
	PreservedSections []string
	SnapshotRequired  bool
	ApprovalRequired  bool
	Status            ProposalStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TargetedRegenerationRepository interface {
	SaveRegeneration(TargetedRegeneration) (TargetedRegeneration, error)
	ListRegenerations(projectID string) ([]TargetedRegeneration, error)
}

type PlanningV2Service struct {
	repo TargetedRegenerationRepository
}

func NewPlanningV2Service(repo TargetedRegenerationRepository) PlanningV2Service {
	return PlanningV2Service{repo: repo}
}

func (s PlanningV2Service) Regenerate(projectID, reason, scope string) (TargetedRegeneration, error) {
	return s.repo.SaveRegeneration(BuildTargetedRegeneration(projectID, reason, scope))
}

func (s PlanningV2Service) List(projectID string) ([]TargetedRegeneration, error) {
	return s.repo.ListRegenerations(projectID)
}

func BuildTargetedRegeneration(projectID, reason, scope string) TargetedRegeneration {
	if strings.TrimSpace(reason) == "" {
		reason = "Targeted continuous planning update"
	}
	if strings.TrimSpace(scope) == "" {
		scope = "affected-sections"
	}
	sections := []string{"dependencies", "risks", "validations"}
	lower := strings.ToLower(reason)
	if strings.Contains(lower, "database") || strings.Contains(lower, "postgresql") || strings.Contains(lower, "mariadb") {
		sections = append(sections, "stack", "migrations", "rollback")
	}
	return TargetedRegeneration{
		ID:                domain.NewID("regenv2"),
		ProjectID:         projectID,
		Reason:            reason,
		Scope:             scope,
		AffectedSections:  sections,
		PreservedSections: []string{"approved vision", "unaffected requirements", "unrelated tasks"},
		SnapshotRequired:  true,
		ApprovalRequired:  true,
		Status:            ProposalDraft,
	}
}

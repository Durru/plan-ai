package planning

import (
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type PlanEvolutionBlueprint struct {
	ID             string
	ProjectID      string
	Objective      string
	Scope          []string
	Exclusions     []string
	Dependencies   []string
	Stack          []string
	Versions       []string
	Libraries      []string
	Folders        []string
	Files          []string
	Validations    []string
	Tests          []string
	Risks          []string
	Rollback       []string
	ApprovedInputs []string
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PlanEvolutionRepository interface {
	SaveBlueprint(PlanEvolutionBlueprint) (PlanEvolutionBlueprint, error)
	ListBlueprints(projectID string) ([]PlanEvolutionBlueprint, error)
}

type PlanEvolutionEngine struct{ repo PlanEvolutionRepository }

func NewPlanEvolutionEngine(repo PlanEvolutionRepository) PlanEvolutionEngine {
	return PlanEvolutionEngine{repo: repo}
}

func (e PlanEvolutionEngine) Generate(projectID, objective string, approvedInputs []string) (PlanEvolutionBlueprint, error) {
	return e.repo.SaveBlueprint(BuildPlanEvolutionBlueprint(projectID, objective, approvedInputs))
}

func (e PlanEvolutionEngine) List(projectID string) ([]PlanEvolutionBlueprint, error) {
	return e.repo.ListBlueprints(projectID)
}

func BuildPlanEvolutionBlueprint(projectID, objective string, approvedInputs []string) PlanEvolutionBlueprint {
	cleanObjective := strings.TrimSpace(objective)
	if cleanObjective == "" {
		cleanObjective = "Generate implementation-ready plan from approved project facts"
	}
	if len(approvedInputs) == 0 {
		approvedInputs = []string{"approved intent", "approved vision", "approved requirements", "approved research", "approved constraints"}
	}
	return PlanEvolutionBlueprint{
		ID:             domain.NewID("planv3"),
		ProjectID:      projectID,
		Objective:      cleanObjective,
		Scope:          approvedInputs,
		Exclusions:     []string{"unapproved candidate requirements", "real OpenCode config mutation"},
		Dependencies:   []string{"approved vision document", "approved requirements", "validated research evidence"},
		Stack:          []string{"Go", "SQLite", "Cobra CLI"},
		Versions:       []string{"Plan-AI V2 Stage C"},
		Libraries:      []string{"modernc.org/sqlite", "spf13/cobra"},
		Folders:        []string{"cmd/plan-ai", "internal/planning", "internal/context", "internal/store"},
		Files:          []string{"cmd/plan-ai/main.go", "internal/store/store.go", "scripts/test-sandbox.sh"},
		Validations:    []string{"go test ./...", "go vet ./...", "go build ./...", "bash scripts/test-sandbox.sh"},
		Tests:          []string{"unit tests for generated plan sections", "sandbox CLI smoke test"},
		Risks:          []string{"planning from unapproved facts", "oversized implementation package"},
		Rollback:       []string{"revert additive Stage C migrations and CLI wiring"},
		ApprovedInputs: approvedInputs,
		Status:         StatusDraft,
	}
}

package store

import (
	"database/sql"

	domainrepos "github.com/Durru/plan-ai/internal/store/repositories"
)

type Repositories struct {
	Project     domainrepos.ProjectRepository
	Vision      domainrepos.VisionRepository
	Requirement domainrepos.RequirementRepository
	Constraint  domainrepos.ConstraintRepository
	Decision    domainrepos.DecisionRepository
	Research    domainrepos.ResearchRepository
	Knowledge   domainrepos.KnowledgeRepository
	Plan        domainrepos.PlanRepository
	Phase       domainrepos.PhaseRepository
	Task        domainrepos.TaskRepository
	Validation  domainrepos.ValidationRepository
	Snapshot    domainrepos.SnapshotRepository
	Change      domainrepos.ChangeRepository
}

func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		Project:     domainrepos.NewProjectRepository(db),
		Vision:      domainrepos.NewVisionRepository(db),
		Requirement: domainrepos.NewRequirementRepository(db),
		Constraint:  domainrepos.NewConstraintRepository(db),
		Decision:    domainrepos.NewDecisionRepository(db),
		Research:    domainrepos.NewResearchRepository(db),
		Knowledge:   domainrepos.NewKnowledgeRepository(db),
		Plan:        domainrepos.NewPlanRepository(db),
		Phase:       domainrepos.NewPhaseRepository(db),
		Task:        domainrepos.NewTaskRepository(db),
		Validation:  domainrepos.NewValidationRepository(db),
		Snapshot:    domainrepos.NewSnapshotRepository(db),
		Change:      domainrepos.NewChangeRepository(db),
	}
}

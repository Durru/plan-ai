package planning

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service { return &Service{repo: repo, now: time.Now().UTC} }

func (s *Service) CreateMasterPlan(input PlanningInput) (MasterPlan, error) {
	if strings.TrimSpace(input.ProjectID) == "" {
		return MasterPlan{}, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(input.VisionReference) == "" {
		return MasterPlan{}, fmt.Errorf("vision reference is required")
	}
	if len(input.ApprovedRequirements) == 0 && len(input.ApprovedConstraints) == 0 && len(input.ApprovedDecisions) == 0 {
		return MasterPlan{}, fmt.Errorf("approved context is required")
	}
	now := s.now()
	title := "Master Plan"
	if len(input.ApprovedRequirements) > 0 {
		title = input.ApprovedRequirements[0]
	}
	plan := MasterPlan{ID: domain.NewID("master"), ProjectID: input.ProjectID, Title: title, VisionReference: input.VisionReference, Objectives: input.ApprovedRequirements, Scope: input.ApprovedRequirements, OutOfScope: []string{}, RecommendedSpecificPlans: input.ApprovedRequirements, Risks: []string{}, Assumptions: input.ApprovedConstraints, Status: StatusDraft, CreatedAt: now, UpdatedAt: now}
	return s.repo.CreateMasterPlan(plan)
}

func (s *Service) CreateSpecificPlan(masterPlanID string, input SpecificPlanInput) (SpecificPlan, error) {
	if strings.TrimSpace(masterPlanID) == "" {
		return SpecificPlan{}, fmt.Errorf("master plan id is required")
	}
	if strings.TrimSpace(input.ProjectID) == "" {
		return SpecificPlan{}, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(input.Goal) == "" {
		return SpecificPlan{}, fmt.Errorf("goal is required")
	}
	now := s.now()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = input.Goal
	}
	strategy := strings.TrimSpace(input.ImplementationStrategy)
	if strategy == "" {
		strategy = "Implement the approved requirements using referenced research and knowledge."
	}
	plan := SpecificPlan{ID: domain.NewID("specific"), ProjectID: input.ProjectID, MasterPlanID: masterPlanID, Title: title, Goal: input.Goal, Requirements: input.Requirements, Constraints: input.Constraints, Decisions: input.Decisions, KnowledgeUsed: input.KnowledgeUsed, ResearchUsed: input.ResearchUsed, ImplementationStrategy: strategy, Risks: input.Risks, ValidationCriteria: input.ValidationCriteria, Status: StatusDraft, CreatedAt: now, UpdatedAt: now}
	return s.repo.CreateSpecificPlan(plan)
}

func (s *Service) CreateImplementationDocument(specificPlanID string, input ImplementationDocumentInput) (ImplementationDocument, error) {
	if strings.TrimSpace(specificPlanID) == "" {
		return ImplementationDocument{}, fmt.Errorf("specific plan id is required")
	}
	if strings.TrimSpace(input.ProjectID) == "" {
		return ImplementationDocument{}, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(input.Objective) == "" {
		return ImplementationDocument{}, fmt.Errorf("objective is required")
	}
	now := s.now()
	doc := ImplementationDocument{ID: domain.NewID("impl"), ProjectID: input.ProjectID, SpecificPlanID: specificPlanID, Objective: input.Objective, Architecture: input.Architecture, ExpectedFiles: input.ExpectedFiles, ExpectedDirectories: input.ExpectedDirectories, Validations: input.Validations, KnownRisks: input.KnownRisks, TestingStrategy: input.TestingStrategy, RollbackStrategy: input.RollbackStrategy, CreatedAt: now, UpdatedAt: now}
	return s.repo.CreateImplementationDocument(doc)
}

package context

import (
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type ImplementationPackage struct {
	ID              string
	ProjectID       string
	PlanID          string
	ModelTarget     string
	WhatToDo        string
	HowToDoIt       string
	FilesToTouch    []string
	FilesNotToTouch []string
	Examples        []string
	Commands        []string
	Validations     []string
	RollbackNotes   []string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ImplementationPackageRepository interface {
	SaveImplementationPackage(ImplementationPackage) (ImplementationPackage, error)
	ListImplementationPackages(projectID string) ([]ImplementationPackage, error)
}

type ImplementationPackageService struct {
	repo ImplementationPackageRepository
}

func NewImplementationPackageService(repo ImplementationPackageRepository) ImplementationPackageService {
	return ImplementationPackageService{repo: repo}
}

func (s ImplementationPackageService) Create(projectID, planID, modelTarget, objective string) (ImplementationPackage, error) {
	return s.repo.SaveImplementationPackage(BuildImplementationPackage(projectID, planID, modelTarget, objective))
}

func (s ImplementationPackageService) List(projectID string) ([]ImplementationPackage, error) {
	return s.repo.ListImplementationPackages(projectID)
}

func BuildImplementationPackage(projectID, planID, modelTarget, objective string) ImplementationPackage {
	if strings.TrimSpace(modelTarget) == "" {
		modelTarget = "opencode"
	}
	if strings.TrimSpace(objective) == "" {
		objective = "Implement the approved plan without touching unapproved runtime state."
	}
	return ImplementationPackage{
		ID:              domain.NewID("implpkg"),
		ProjectID:       projectID,
		PlanID:          planID,
		ModelTarget:     modelTarget,
		WhatToDo:        objective,
		HowToDoIt:       "Use approved plan sections, preserve MVP compatibility, and keep changes additive.",
		FilesToTouch:    []string{"cmd/plan-ai/main.go", "internal", "scripts/test-sandbox.sh"},
		FilesNotToTouch: []string{"/root/.config/opencode", "/root/.plan-ai", ".plan-ai runtime state"},
		Examples:        []string{"Prefer CREATE TABLE IF NOT EXISTS for additive migrations."},
		Commands:        []string{"gofmt -w cmd internal", "go test ./...", "go vet ./...", "go build ./...", "bash scripts/test-sandbox.sh"},
		Validations:     []string{"sandbox cleans runtime artifacts", "existing MVP commands still pass"},
		RollbackNotes:   []string{"Revert Stage C source and migration additions."},
		Status:          "draft",
	}
}

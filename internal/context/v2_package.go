package context

import (
	"strings"
	"time"
)

type PackageType string

const (
	PackageVision         PackageType = "vision"
	PackageResearch       PackageType = "research"
	PackagePlanning       PackageType = "planning"
	PackageImplementation PackageType = "implementation"
	PackageChange         PackageType = "change"
)

type SmartPackage struct {
	ID          string
	ProjectID   string
	Type        PackageType
	ModelTarget string
	Summary     string
	Content     string
	Priority    int
	TokenBudget int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SmartPackageRepository interface {
	SavePackage(SmartPackage) (SmartPackage, error)
	ListPackages(projectID string) ([]SmartPackage, error)
	GetPackage(id string) (SmartPackage, error)
}

type SmartPackageService struct{ repo SmartPackageRepository }

func NewSmartPackageService(repo SmartPackageRepository) SmartPackageService {
	return SmartPackageService{repo: repo}
}

func (s SmartPackageService) Create(projectID string, typ PackageType, model, content string, budget int) (SmartPackage, error) {
	return s.repo.SavePackage(BuildSmartPackage(projectID, typ, model, content, budget))
}

func (s SmartPackageService) List(projectID string) ([]SmartPackage, error) {
	return s.repo.ListPackages(projectID)
}

func (s SmartPackageService) Get(id string) (SmartPackage, error) { return s.repo.GetPackage(id) }

func BuildSmartPackage(projectID string, typ PackageType, model, content string, budget int) SmartPackage {
	if model == "" {
		model = "generic"
	}
	if budget <= 0 {
		budget = 4096
	}
	summary := strings.TrimSpace(content)
	if len(summary) > 96 {
		summary = summary[:96]
	}
	if summary == "" {
		summary = "Context package awaiting approved project facts."
	}
	return SmartPackage{ProjectID: projectID, Type: typ, ModelTarget: model, Summary: summary, Content: content, Priority: priorityForType(typ), TokenBudget: budget}
}

func priorityForType(typ PackageType) int {
	switch typ {
	case PackageImplementation:
		return 1
	case PackageChange:
		return 2
	case PackagePlanning:
		return 3
	case PackageVision:
		return 4
	default:
		return 5
	}
}

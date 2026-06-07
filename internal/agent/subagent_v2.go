package agent

import (
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type SubagentType string
type ValidationStatus string

const (
	SubagentResearch      SubagentType     = "research"
	SubagentArchitecture  SubagentType     = "architecture"
	SubagentUI            SubagentType     = "ui"
	SubagentUX            SubagentType     = "ux"
	SubagentSecurity      SubagentType     = "security"
	SubagentBackend       SubagentType     = "backend"
	SubagentDatabase      SubagentType     = "database"
	SubagentValidation    SubagentType     = "validation"
	ValidationPending     ValidationStatus = "pending"
	ValidationPassed      ValidationStatus = "passed"
	ValidationNeedsReview ValidationStatus = "needs_review"
)

type SubagentTask struct {
	ID               string
	ProjectID        string
	AgentType        SubagentType
	Objective        string
	Capability       string
	Status           JobStatus
	Provenance       string
	ValidationStatus ValidationStatus
	Isolated         bool
	Temporary        bool
	MemoryPolicy     string
	ResultSummary    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SubagentTaskRepository interface {
	SaveSubagentTask(SubagentTask) (SubagentTask, error)
	ListSubagentTasks(projectID string) ([]SubagentTask, error)
}

type SubagentOrchestrator struct{ repo SubagentTaskRepository }

func NewSubagentOrchestrator(repo SubagentTaskRepository) SubagentOrchestrator {
	return SubagentOrchestrator{repo: repo}
}

func (o SubagentOrchestrator) Create(projectID string, agentType SubagentType, objective string) (SubagentTask, error) {
	return o.repo.SaveSubagentTask(BuildSubagentTask(projectID, agentType, objective))
}

func (o SubagentOrchestrator) List(projectID string) ([]SubagentTask, error) {
	return o.repo.ListSubagentTasks(projectID)
}

func BuildSubagentTask(projectID string, agentType SubagentType, objective string) SubagentTask {
	if agentType == "" {
		agentType = SubagentValidation
	}
	if strings.TrimSpace(objective) == "" {
		objective = "Validate the next Plan-AI task and return a bounded result."
	}
	return SubagentTask{
		ID:               domain.NewID("subagentv2"),
		ProjectID:        projectID,
		AgentType:        agentType,
		Objective:        objective,
		Capability:       capabilityForSubagent(agentType),
		Status:           JobStatusPending,
		Provenance:       "created by Plan-AI V2 subagent orchestrator",
		ValidationStatus: ValidationPending,
		Isolated:         true,
		Temporary:        true,
		MemoryPolicy:     "no-independent-persistent-memory",
	}
}

func capabilityForSubagent(agentType SubagentType) string {
	switch agentType {
	case SubagentResearch:
		return CapabilityResearch
	case SubagentArchitecture, SubagentBackend, SubagentDatabase:
		return CapabilityPlanning
	case SubagentValidation, SubagentSecurity:
		return "validation"
	default:
		return CapabilityContext
	}
}

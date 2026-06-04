package research

import (
	"strings"
	"time"
)

type AgentType string

const (
	AgentMarket         AgentType = "market"
	AgentTechnical      AgentType = "technical"
	AgentArchitecture   AgentType = "architecture"
	AgentUI             AgentType = "ui"
	AgentUX             AgentType = "ux"
	AgentSecurity       AgentType = "security"
	AgentInfrastructure AgentType = "infrastructure"
)

type OrchestrationRun struct {
	ID         string
	ProjectID  string
	Agent      AgentType
	Topic      string
	Summary    string
	Evidence   []string
	Confidence int
	Status     ResearchStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrchestrationRepository interface {
	SaveRun(OrchestrationRun) (OrchestrationRun, error)
	ListRuns(projectID string) ([]OrchestrationRun, error)
	GetRun(id string) (OrchestrationRun, error)
}

type Orchestrator struct{ repo OrchestrationRepository }

func NewOrchestrator(repo OrchestrationRepository) Orchestrator { return Orchestrator{repo: repo} }

func (o Orchestrator) Run(projectID string, agent AgentType, topic string) (OrchestrationRun, error) {
	return o.repo.SaveRun(BuildRun(projectID, agent, topic))
}

func (o Orchestrator) List(projectID string) ([]OrchestrationRun, error) {
	return o.repo.ListRuns(projectID)
}

func BuildRun(projectID string, agent AgentType, topic string) OrchestrationRun {
	if agent == "" {
		agent = AgentTechnical
	}
	cleanTopic := strings.TrimSpace(topic)
	if cleanTopic == "" {
		cleanTopic = "General project research"
	}
	return OrchestrationRun{
		ProjectID:  projectID,
		Agent:      agent,
		Topic:      cleanTopic,
		Summary:    string(agent) + " research summary for " + cleanTopic,
		Evidence:   []string{"deterministic V2 research orchestration evidence", "persisted for knowledge synthesis"},
		Confidence: 75,
		Status:     ResearchStatusCompleted,
	}
}

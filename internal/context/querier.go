package context

// ProjectBrief is a lightweight project representation used by context builders.
type ProjectBrief struct {
	ID     string
	Name   string
	Status string
}

// PlanBrief is a lightweight plan representation.
type PlanBrief struct {
	ID     string
	Title  string
	Status string
}

// PhaseBrief is a lightweight phase representation.
type PhaseBrief struct {
	ID     string
	Title  string
	Status string
}

// TaskBrief is a lightweight task representation.
type TaskBrief struct {
	ID     string
	Title  string
	Status string
}

// DecisionBrief is a lightweight decision representation.
type DecisionBrief struct {
	ID     string
	Title  string
	Status string
}

// ResearchBrief is a lightweight research entry representation.
type ResearchBrief struct {
	ID      string
	Topic   string
	Summary string
}

// KnowledgeBrief is a lightweight knowledge object representation.
type KnowledgeBrief struct {
	ID      string
	Topic   string
	Summary string
}

// ProjectQuerier provides read-only access to project data for context building.
// This interface lives in the context package to avoid import cycles with store.
type ProjectQuerier interface {
	GetProjectBrief(id string) (ProjectBrief, error)
	ListPlanBriefs(projectID string) ([]PlanBrief, error)
	ListPhaseBriefs(planID string) ([]PhaseBrief, error)
	ListTaskBriefs(phaseID string) ([]TaskBrief, error)
	ListDecisionBriefs(projectID string) ([]DecisionBrief, error)
	ListResearchBriefs(projectID string) ([]ResearchBrief, error)
	ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error)
}

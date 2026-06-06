package impact

type NodeType string

const (
	NodeDecision  NodeType = "decision"
	NodeResearch  NodeType = "research"
	NodeKnowledge NodeType = "knowledge"
	NodePlan      NodeType = "plan"
	NodePhase     NodeType = "phase"
	NodeTask      NodeType = "task"
	NodeFile      NodeType = "file"
)

type EdgeType string

const (
	EdgeDependsOn   EdgeType = "depends_on"
	EdgeDerivedFrom EdgeType = "derived_from"
	EdgeAffects     EdgeType = "affects"
	EdgeImplements  EdgeType = "implements"
	EdgeSupersedes  EdgeType = "supersedes"
	EdgeReferences  EdgeType = "references"
)

type Node struct {
	ID         string         `json:"id"`
	Type       NodeType       `json:"type"`
	ProjectID  string         `json:"project_id"`
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties"`
}

type Edge struct {
	ID         string   `json:"id"`
	SourceID   string   `json:"source_id"`
	SourceType NodeType `json:"source_type"`
	TargetID   string   `json:"target_id"`
	TargetType NodeType `json:"target_type"`
	EdgeType   EdgeType `json:"edge_type"`
	ProjectID  string   `json:"project_id"`
	Weight     int      `json:"weight"`
}

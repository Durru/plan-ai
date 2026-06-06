package impact

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Graph struct {
	nodes map[string]Node
	edges []Edge
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]Node),
		edges: make([]Edge, 0),
	}
}

func (g *Graph) Nodes() []Node {
	out := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	return out
}

func (g *Graph) Edges() []Edge {
	out := make([]Edge, len(g.edges))
	copy(out, g.edges)
	return out
}

func (g *Graph) AddNode(n Node) {
	g.nodes[n.ID] = n
}

func (g *Graph) AddEdge(e Edge) {
	g.edges = append(g.edges, e)
}

func (g *Graph) Build(projectID string, edges []Edge) {
	for _, e := range edges {
		if e.ProjectID == "" {
			e.ProjectID = projectID
		}
		g.AddEdge(e)
	}
}

func (g *Graph) BuildFromEntityLinks(db *sql.DB, projectID string) error {
	rows, err := db.Query(
		`SELECT id, project_id, source_type, source_id, target_type, target_id, link_type, created_at
		 FROM entity_links WHERE project_id = ?`, projectID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, pid, sourceType, sourceID, targetType, targetID, linkType, createdAt string
		if err := rows.Scan(&id, &pid, &sourceType, &sourceID, &targetType, &targetID, &linkType, &createdAt); err != nil {
			return err
		}

		edgeType := mapLinkTypeToEdgeType(linkType)
		edge := Edge{
			ID:         id,
			SourceID:   sourceID,
			SourceType: NodeType(sourceType),
			TargetID:   targetID,
			TargetType: NodeType(targetType),
			EdgeType:   edgeType,
			ProjectID:  pid,
			Weight:     1,
		}
		g.AddEdge(edge)

		sourceNode := Node{
			ID:        sourceID,
			Type:      NodeType(sourceType),
			ProjectID: pid,
			Label:     sourceType + ":" + sourceID,
			Properties: map[string]any{
				"created_at": createdAt,
			},
		}
		g.AddNode(sourceNode)

		targetNode := Node{
			ID:        targetID,
			Type:      NodeType(targetType),
			ProjectID: pid,
			Label:     targetType + ":" + targetID,
			Properties: map[string]any{
				"created_at": createdAt,
			},
		}
		g.AddNode(targetNode)
	}
	return rows.Err()
}

func mapLinkTypeToEdgeType(linkType string) EdgeType {
	switch linkType {
	case "depends_on":
		return EdgeDependsOn
	case "derived_from":
		return EdgeDerivedFrom
	case "affects":
		return EdgeAffects
	case "implements":
		return EdgeImplements
	case "supersedes":
		return EdgeSupersedes
	case "references":
		return EdgeReferences
	default:
		return EdgeReferences
	}
}

func newEdgeID() string {
	return uuid.New().String()
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

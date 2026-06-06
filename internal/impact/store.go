package impact

import "database/sql"

type EdgeRepository struct {
	db *sql.DB
}

func NewEdgeRepository(db *sql.DB) *EdgeRepository {
	return &EdgeRepository{db: db}
}

func (r *EdgeRepository) SaveEdge(e Edge) error {
	if e.ID == "" {
		e.ID = newEdgeID()
	}
	if e.Weight == 0 {
		e.Weight = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO impact_edges (id, project_id, source_type, source_id, target_type, target_id, edge_type, weight, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(source_type, source_id, target_type, target_id, edge_type) DO UPDATE SET
		   weight = excluded.weight`,
		e.ID, e.ProjectID, string(e.SourceType), e.SourceID, string(e.TargetType), e.TargetID, string(e.EdgeType), e.Weight, nowUTC(),
	)
	return err
}

func (r *EdgeRepository) ListEdges(projectID string) ([]Edge, error) {
	rows, err := r.db.Query(
		`SELECT id, project_id, source_type, source_id, target_type, target_id, edge_type, weight
		 FROM impact_edges WHERE project_id = ?`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		var srcType, tgtType, eType string
		if err := rows.Scan(&e.ID, &e.ProjectID, &srcType, &e.SourceID, &tgtType, &e.TargetID, &eType, &e.Weight); err != nil {
			return nil, err
		}
		e.SourceType = NodeType(srcType)
		e.TargetType = NodeType(tgtType)
		e.EdgeType = EdgeType(eType)
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

func (r *EdgeRepository) ListEdgesBySource(sourceType, sourceID string) ([]Edge, error) {
	rows, err := r.db.Query(
		`SELECT id, project_id, source_type, source_id, target_type, target_id, edge_type, weight
		 FROM impact_edges WHERE source_type = ? AND source_id = ?`, sourceType, sourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		var sType, tType, eType string
		if err := rows.Scan(&e.ID, &e.ProjectID, &sType, &e.SourceID, &tType, &e.TargetID, &eType, &e.Weight); err != nil {
			return nil, err
		}
		e.SourceType = NodeType(sType)
		e.TargetType = NodeType(tType)
		e.EdgeType = EdgeType(eType)
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

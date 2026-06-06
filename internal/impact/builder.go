package impact

import (
	"database/sql"

	"github.com/google/uuid"
)

// BuildFromEntityLinks reads the entity_links table, converts each row
// into an impact edge, and persists it via the EdgeRepository.
func BuildFromEntityLinks(db *sql.DB, projectID string) error {
	repo := NewEdgeRepository(db)

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

		_ = createdAt

		edge := Edge{
			ID:         uuid.New().String(),
			SourceID:   sourceID,
			SourceType: NodeType(sourceType),
			TargetID:   targetID,
			TargetType: NodeType(targetType),
			EdgeType:   mapLinkTypeToEdgeType(linkType),
			ProjectID:  pid,
			Weight:     1,
		}
		if err := repo.SaveEdge(edge); err != nil {
			return err
		}
	}
	return rows.Err()
}

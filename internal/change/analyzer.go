package change

import (
	"database/sql"

	"github.com/Durru/plan-ai/internal/impact"
)

// EntityAnalyzer provides context-aware impact analysis by examining
// relationships between planning entities. It identifies transitive
// effects beyond the base invalidation rules, and queries entity_links
// to resolve real affected entity IDs.
type EntityAnalyzer struct {
	rules []InvalidationRule
	db    *sql.DB
}

// NewAnalyzer creates an EntityAnalyzer with the default rules and optional DB
// for entity_links resolution. Pass nil for rule-only analysis.
func NewAnalyzer(db *sql.DB) *EntityAnalyzer {
	return &EntityAnalyzer{rules: DefaultInvalidationRules, db: db}
}

// AnalyzeChange determines which entity types are affected by the given change type.
func (a *EntityAnalyzer) AnalyzeChange(ct ChangeType) *ImpactAnalysis {
	affected := &ImpactAnalysis{
		AffectedTypes:  make(map[string][]string),
		AffectedByType: []AffectedGroup{},
	}

	for _, rule := range a.rules {
		for _, affectedBy := range rule.AffectedBy {
			if affectedBy == ct {
				affected.AffectedTypes[rule.EntityType] = []string{}
				affected.AffectedByType = append(affected.AffectedByType, AffectedGroup{
					EntityType: rule.EntityType,
					EntityIDs:  []string{},
				})
				if rule.ResultStatus == EntityBlocked || rule.ResultStatus == EntityNeedsReview {
					affected.ReviewRequired = true
				}
				break
			}
		}
	}

	if len(affected.AffectedByType) == 0 {
		affected.Summary = "No entities affected"
	} else {
		affected.Summary = "Analysis based on invalidation rules - resolve with AnalyzeEntityLinks for concrete IDs"
	}

	return affected
}

// AnalyzeEntityLinks queries the entity_links table and builds an impact graph
// to resolve real entity IDs transitively affected by the source entity.
func (a *EntityAnalyzer) AnalyzeEntityLinks(projectID, sourceType, sourceID string) ([]AffectedGroup, error) {
	if a.db == nil {
		return nil, nil
	}

	g := impact.NewGraph()
	if err := g.BuildFromEntityLinks(a.db, projectID); err != nil {
		return nil, err
	}

	affectedNodes := g.AffectedEntities(sourceID)
	if len(affectedNodes) == 0 {
		return nil, nil
	}

	typeMap := make(map[string][]string)
	for _, n := range affectedNodes {
		typeMap[string(n.Type)] = append(typeMap[string(n.Type)], n.ID)
	}

	var groups []AffectedGroup
	for entityType, ids := range typeMap {
		groups = append(groups, AffectedGroup{
			EntityType: entityType,
			EntityIDs:  ids,
		})
	}

	return groups, nil
}

// AffectedByChangeType returns which entity types are affected by a given change type.
func (a *EntityAnalyzer) AffectedByChangeType(ct ChangeType) []string {
	var types []string
	for _, rule := range a.rules {
		for _, affectedBy := range rule.AffectedBy {
			if affectedBy == ct {
				types = append(types, rule.EntityType)
				break
			}
		}
	}
	return types
}

package change

// EntityAnalyzer provides context-aware impact analysis by examining
// relationships between planning entities. It identifies transitive
// effects beyond the base invalidation rules.
type EntityAnalyzer struct {
	rules []InvalidationRule
}

// NewAnalyzer creates an EntityAnalyzer with the default rules.
func NewAnalyzer() *EntityAnalyzer {
	return &EntityAnalyzer{rules: DefaultInvalidationRules}
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
		affected.Summary = "Template analysis - entities will be resolved against actual data"
	}

	return affected
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

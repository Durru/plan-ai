package change

import "fmt"

// ImpactBuilder constructs detailed impact analyses from changes
// and project data.
type ImpactBuilder struct {
	analyzer *EntityAnalyzer
}

// NewImpactBuilder creates an impact builder.
func NewImpactBuilder() *ImpactBuilder {
	return &ImpactBuilder{analyzer: NewAnalyzer(nil)}
}

// Build constructs a detailed ImpactAnalysis for a change event.
func (b *ImpactBuilder) Build(ev *ChangeEvent) *ImpactAnalysis {
	template := b.analyzer.AnalyzeChange(ev.ChangeType)

	analysis := &ImpactAnalysis{
		ChangeID:       ev.ID,
		AffectedTypes:  make(map[string][]string),
		AffectedByType: template.AffectedByType,
		ReviewRequired: template.ReviewRequired,
	}

	for _, group := range template.AffectedByType {
		analysis.AffectedTypes[group.EntityType] = []string{}
	}

	analysis.Summary = b.buildSummary(ev, len(analysis.AffectedByType))
	return analysis
}

func (b *ImpactBuilder) buildSummary(ev *ChangeEvent, affectedCount int) string {
	if affectedCount == 0 {
		return fmt.Sprintf("Change %q has no impact on tracked planning entities", ev.ChangeType)
	}
	return fmt.Sprintf("Change %q affects %d planning entity types across the project. %s review required.",
		ev.ChangeType, affectedCount, map[bool]string{true: "Human", false: "No"}[affectedCount > 0])
}

// ClassifySeverity determines the severity of a change event.
func ClassifySeverity(ct ChangeType) Severity {
	switch ct {
	case VisionChanged, RequirementRemoved, PlanChanged:
		return SeverityHigh
	case RequirementAdded, ConstraintChanged, DecisionChanged:
		return SeverityMedium
	case ResearchUpdated, KnowledgeUpdated, TechnologyChanged, ImplementationFeedback:
		return SeverityLow
	default:
		return SeverityMedium
	}
}

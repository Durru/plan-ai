package knowledge

import (
	"fmt"
	"strings"

	"github.com/Durru/plan-ai/internal/domain"
)

// IsValidRelationType returns true when value is one of the official
// knowledge relation types declared in internal/domain.
func IsValidRelationType(value domain.KnowledgeRelationType) bool {
	switch value {
	case domain.KnowledgeRelationRelated,
		domain.KnowledgeRelationDependsOn,
		domain.KnowledgeRelationAlternativeTo,
		domain.KnowledgeRelationExtends:
		return true
	default:
		return false
	}
}

// IsValidReferenceType returns true when value is one of the official
// knowledge reference types declared in internal/domain.
func IsValidReferenceType(value domain.KnowledgeReferenceType) bool {
	switch value {
	case domain.KnowledgeReferencePlan,
		domain.KnowledgeReferenceDecision,
		domain.KnowledgeReferenceResearch,
		domain.KnowledgeReferenceTechnology:
		return true
	default:
		return false
	}
}

// NormalizeTag trims, lowercases, and validates a user-provided tag.
// Empty tags are rejected. Tags with internal whitespace or separators
// other than dashes / underscores are accepted when single words; multi
// word tags are accepted with a single space.
func NormalizeTag(raw string) (string, error) {
	tag := strings.ToLower(strings.TrimSpace(raw))
	if tag == "" {
		return "", fmt.Errorf("tag is required")
	}
	if len(tag) > 64 {
		return "", fmt.Errorf("tag must be 64 characters or less")
	}
	for _, r := range tag {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("tag contains control characters")
		}
	}
	return tag, nil
}

// NormalizeRelationType trims and validates a user-provided relation
// type, returning the canonical lower-snake form.
func NormalizeRelationType(raw string) (domain.KnowledgeRelationType, error) {
	value := domain.KnowledgeRelationType(strings.ToLower(strings.TrimSpace(raw)))
	if !IsValidRelationType(value) {
		return "", fmt.Errorf("unknown relation type %q (expected related, depends_on, alternative_to, extends)", raw)
	}
	return value, nil
}

// NormalizeReferenceType trims and validates a user-provided reference
// type, returning the canonical lower-snake form.
func NormalizeReferenceType(raw string) (domain.KnowledgeReferenceType, error) {
	value := domain.KnowledgeReferenceType(strings.ToLower(strings.TrimSpace(raw)))
	if !IsValidReferenceType(value) {
		return "", fmt.Errorf("unknown reference type %q (expected plan, decision, research, technology)", raw)
	}
	return value, nil
}

// NormalizeStatus trims and validates a user-provided knowledge status,
// returning the canonical lower-snake form. Empty input yields
// StatusDraft.
func NormalizeStatus(raw string) (domain.KnowledgeStatus, error) {
	value := domain.KnowledgeStatus(strings.ToLower(strings.TrimSpace(raw)))
	if value == "" {
		return domain.KnowledgeStatusDraft, nil
	}
	switch value {
	case domain.KnowledgeStatusDraft,
		domain.KnowledgeStatusReviewed,
		domain.KnowledgeStatusApproved,
		domain.KnowledgeStatusArchived:
		return value, nil
	default:
		return "", fmt.Errorf("unknown status %q (expected draft, reviewed, approved, archived)", raw)
	}
}

// NormalizeSourceType trims and validates a user-provided source type.
// Empty input yields SourceManual.
func NormalizeSourceType(raw string) (domain.KnowledgeSourceType, error) {
	value := domain.KnowledgeSourceType(strings.ToLower(strings.TrimSpace(raw)))
	if value == "" {
		return domain.KnowledgeSourceManual, nil
	}
	switch value {
	case domain.KnowledgeSourceManual,
		domain.KnowledgeSourceResearch,
		domain.KnowledgeSourceImported,
		domain.KnowledgeSourceGenerated:
		return value, nil
	default:
		return "", fmt.Errorf("unknown source type %q (expected manual, research, imported, generated)", raw)
	}
}

// NormalizeCategory trims, lowercases, and validates a user-provided
// category. Empty input yields Classify(topic).
func NormalizeCategory(raw, topic string) domain.KnowledgeCategory {
	value := domain.KnowledgeCategory(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case domain.KnowledgeCategoryDatabase,
		domain.KnowledgeCategoryAuthentication,
		domain.KnowledgeCategoryBilling,
		domain.KnowledgeCategoryFrontend,
		domain.KnowledgeCategoryBackend,
		domain.KnowledgeCategorySecurity,
		domain.KnowledgeCategoryDeployment,
		domain.KnowledgeCategoryArchitecture,
		domain.KnowledgeCategoryTesting,
		domain.KnowledgeCategoryMCP,
		domain.KnowledgeCategoryAgents,
		domain.KnowledgeCategoryAI,
		domain.KnowledgeCategoryDevops,
		domain.KnowledgeCategoryIntegration,
		domain.KnowledgeCategoryGeneral:
		return value
	default:
		return Classify(topic)
	}
}

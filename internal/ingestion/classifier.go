package ingestion

import "strings"

func Classify(text string) Classification {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "vision:") || strings.Contains(lower, "objective:") || strings.Contains(lower, "goal:"):
		return ClassificationVision
	case strings.Contains(lower, "must ") || strings.Contains(lower, "shall ") || strings.Contains(lower, "requirement"):
		return ClassificationRequirement
	case strings.Contains(lower, "constraint") || strings.Contains(lower, "only") || strings.Contains(lower, "do not") || strings.Contains(lower, "must not"):
		return ClassificationConstraint
	case strings.Contains(lower, "prefer") || strings.Contains(lower, "preference"):
		return ClassificationPreference
	case strings.Contains(lower, "decision:") || strings.Contains(lower, "decided"):
		return ClassificationDecision
	case strings.Contains(lower, "reference") || strings.Contains(lower, "http://") || strings.Contains(lower, "https://") || strings.Contains(lower, ".png") || strings.Contains(lower, ".jpg"):
		return ClassificationReference
	default:
		return ClassificationUnknown
	}
}

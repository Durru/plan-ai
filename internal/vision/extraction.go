package vision

import (
	"strings"

	"github.com/Durru/plan-ai/internal/ingestion"
)

func Extract(projectID string, sources []ingestion.IngestedSource) Draft {
	text := joinSources(sources)
	sentences := splitSentences(text)
	draft := Draft{ProjectID: projectID, Summary: firstObjective(sentences), Title: titleFromObjective(firstObjective(sentences))}
	for _, sentence := range sentences {
		lower := strings.ToLower(sentence)
		switch {
		case strings.Contains(lower, "for ") && (strings.Contains(lower, "users") || strings.Contains(lower, "teams") || strings.Contains(lower, "admins")):
			draft.TargetUsers = appendUnique(draft.TargetUsers, sentence)
		case strings.Contains(lower, "must") || strings.Contains(lower, "constraint") || strings.Contains(lower, "only") || strings.Contains(lower, "do not"):
			draft.Constraints = appendUnique(draft.Constraints, sentence)
		case strings.Contains(lower, "prefer"):
			draft.UXGoals = appendUnique(draft.UXGoals, sentence)
		case strings.Contains(lower, "feature") || strings.Contains(lower, "allow") || strings.Contains(lower, "support") || strings.Contains(lower, "save"):
			draft.FunctionalGoals = appendUnique(draft.FunctionalGoals, sentence)
		case strings.Contains(lower, "success") || strings.Contains(lower, "outcome"):
			draft.SuccessCriteria = appendUnique(draft.SuccessCriteria, sentence)
			if draft.ExpectedOutcome == "" {
				draft.ExpectedOutcome = sentence
			}
		}
	}
	for _, source := range sources {
		draft.VisualReferences = appendUnique(draft.VisualReferences, ingestion.ExtractReferences(source.NormalizedContent)...)
	}
	draft.MissingInformation = MissingInformation(draft)
	return draft
}

func joinSources(sources []ingestion.IngestedSource) string {
	parts := make([]string, 0, len(sources))
	for _, source := range sources {
		parts = append(parts, source.NormalizedContent)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func splitSentences(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool { return r == '.' || r == '\n' || r == ';' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func firstObjective(sentences []string) string {
	for _, sentence := range sentences {
		lower := strings.ToLower(sentence)
		if strings.Contains(lower, "build") || strings.Contains(lower, "create") || strings.Contains(lower, "deliver") || strings.Contains(lower, "objective") || strings.Contains(lower, "vision") {
			return sentence
		}
	}
	if len(sentences) > 0 {
		return sentences[0]
	}
	return ""
}

func titleFromObjective(objective string) string {
	objective = strings.TrimSpace(objective)
	if objective == "" {
		return "Untitled vision"
	}
	words := strings.Fields(objective)
	if len(words) > 8 {
		words = words[:8]
	}
	return strings.Join(words, " ")
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

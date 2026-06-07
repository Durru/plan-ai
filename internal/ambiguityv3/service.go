package ambiguityv3

import (
	"fmt"
	"strings"

	"github.com/Durru/plan-ai/internal/discoveryv3"
	"github.com/Durru/plan-ai/internal/intentv3"
)

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) AnalyzeProductIntent(pi intentv3.ProductIntent, questions []discoveryv3.Question, answers []discoveryv3.Answer) AmbiguityReport {
	report := AmbiguityReport{IntentID: pi.ID}

	addKnown(&report, "description", pi.Description)
	addKnown(&report, "expected_outcome", pi.ExpectedOutcome)
	addKnown(&report, "desired_experience", pi.DesiredExperience)
	addKnown(&report, "desired_result", pi.DesiredResult)
	addKnown(&report, "success_definition", pi.SuccessDefinition)
	addKnown(&report, "failure_definition", pi.FailureDefinition)
	if len(pi.UserExpectations) > 0 {
		report.KnownAreas = append(report.KnownAreas, "user_expectations")
	}
	if len(pi.NonExpectations) > 0 {
		report.KnownAreas = append(report.KnownAreas, "non_expectations")
	}

	addMissing(&report, "expected_outcome", pi.ExpectedOutcome, "what the product must achieve is not explicit")
	addMissing(&report, "desired_experience", pi.DesiredExperience, "the intended user experience is not explicit")
	addMissing(&report, "desired_result", pi.DesiredResult, "the concrete desired result is not explicit")
	addMissing(&report, "success_definition", pi.SuccessDefinition, "success criteria are not explicit")
	addMissing(&report, "failure_definition", pi.FailureDefinition, "failure criteria are not explicit")
	if len(pi.UserExpectations) == 0 {
		report.MissingInformation = append(report.MissingInformation, MissingInformation{Field: "user_expectations", Reason: "user expectations are not listed"})
	}
	if len(pi.NonExpectations) == 0 {
		report.Assumptions = append(report.Assumptions, Assumption{ID: "assume-no-non-expectations", Reason: "no explicit non-expectations were provided"})
	}

	answered := map[discoveryv3.QuestionID]bool{}
	for _, a := range answers {
		answered[a.QuestionID] = true
	}
	for _, q := range questions {
		if !answered[q.ID] {
			report.UnknownAreas = append(report.UnknownAreas, UnknownArea{Level: q.Level, QuestionID: q.ID, Question: q.Question, Required: q.Required})
			if q.Required {
				report.NeedsToKnow = append(report.NeedsToKnow, fmt.Sprintf("%s: %s", q.Level, q.Question))
			}
		}
	}

	report.Conflicts = detectConflicts(strings.Join([]string{
		pi.Description,
		pi.ExpectedOutcome,
		pi.DesiredExperience,
		pi.DesiredResult,
		strings.Join(pi.UserExpectations, " "),
		strings.Join(pi.NonExpectations, " "),
		pi.SuccessDefinition,
		pi.FailureDefinition,
	}, " "))

	report.Score = score(report)
	return report
}

func (s Service) AnalyzeText(input string) AmbiguityReport {
	report := AmbiguityReport{}
	text := strings.ToLower(strings.TrimSpace(input))
	if text == "" {
		report.MissingInformation = append(report.MissingInformation, MissingInformation{Field: "input", Reason: "no input provided"})
		report.NeedsToKnow = append(report.NeedsToKnow, "project intent description")
		report.Score = score(report)
		return report
	}
	report.KnownAreas = append(report.KnownAreas, "raw_input")
	vagueTerms := []string{"maybe", "probably", "something", "stuff", "etc", "nice", "simple", "fast", "soon", "later", "algún", "algo", "rápido", "simple"}
	for _, term := range vagueTerms {
		if strings.Contains(text, term) {
			report.Assumptions = append(report.Assumptions, Assumption{ID: "vague-" + sanitizeID(term), Reason: fmt.Sprintf("input uses vague term %q", term)})
		}
	}
	for _, keyword := range []string{"success", "outcome", "resultado", "éxito"} {
		if strings.Contains(text, keyword) {
			report.KnownAreas = append(report.KnownAreas, "outcome_signal")
			break
		}
	}
	if !containsAny(text, []string{"success", "outcome", "resultado", "éxito"}) {
		report.MissingInformation = append(report.MissingInformation, MissingInformation{Field: "success_definition", Reason: "input does not explain how success is recognized"})
		report.NeedsToKnow = append(report.NeedsToKnow, "success definition")
	}
	report.Conflicts = detectConflicts(text)
	report.Score = score(report)
	return report
}

func addKnown(report *AmbiguityReport, field, value string) {
	if strings.TrimSpace(value) != "" {
		report.KnownAreas = append(report.KnownAreas, field)
	}
}

func addMissing(report *AmbiguityReport, field, value, reason string) {
	if strings.TrimSpace(value) == "" {
		report.MissingInformation = append(report.MissingInformation, MissingInformation{Field: field, Reason: reason})
		report.NeedsToKnow = append(report.NeedsToKnow, field)
	}
}

func detectConflicts(input string) []Conflict {
	text := strings.ToLower(input)
	pairs := []struct{ a, b, id string }{
		{"simple", "complex", "simple-vs-complex"},
		{"minimal", "feature-rich", "minimal-vs-feature-rich"},
		{"cheap", "premium", "cheap-vs-premium"},
		{"fast", "thorough", "fast-vs-thorough"},
		{"enterprise", "consumer", "enterprise-vs-consumer"},
	}
	var out []Conflict
	for _, p := range pairs {
		if strings.Contains(text, p.a) && strings.Contains(text, p.b) {
			out = append(out, Conflict{ID: p.id, Evidence: fmt.Sprintf("contains both %q and %q", p.a, p.b)})
		}
	}
	return out
}

func score(report AmbiguityReport) AmbiguityScore {
	s := 0
	s += len(report.MissingInformation) * 10
	s += len(report.Assumptions) * 5
	s += len(report.Conflicts) * 15
	for _, unknown := range report.UnknownAreas {
		if unknown.Required {
			s += 3
		} else {
			s++
		}
	}
	if s > 100 {
		s = 100
	}
	return AmbiguityScore(s)
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func sanitizeID(s string) string {
	replacer := strings.NewReplacer(" ", "-", "á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u")
	return replacer.Replace(strings.ToLower(s))
}

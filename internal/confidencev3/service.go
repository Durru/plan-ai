package confidencev3

import (
	"strings"

	"github.com/Durru/plan-ai/internal/ambiguityv3"
	"github.com/Durru/plan-ai/internal/discoveryv3"
	"github.com/Durru/plan-ai/internal/intentv3"
)

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) Evaluate(pi intentv3.ProductIntent, discovery *intentv3.DiscoveryResult, questions []discoveryv3.Question, answers []discoveryv3.Answer, ambiguity ambiguityv3.AmbiguityReport) ConfidenceReport {
	report := ConfidenceReport{IntentID: pi.ID}

	report.IntentScore = boundedScore(fieldCoverage([]string{
		pi.Description,
		pi.ExpectedOutcome,
		pi.DesiredResult,
		pi.SuccessDefinition,
		pi.FailureDefinition,
	}, 20))
	report.VisionScore = boundedScore(fieldCoverage([]string{pi.ExpectedOutcome, pi.DesiredResult, pi.SuccessDefinition}, 33))
	report.UXScore = boundedScore(fieldCoverage([]string{pi.DesiredExperience}, 70) + listCoverage(pi.UserExpectations, 15) + listCoverage(pi.NonExpectations, 15))
	report.BusinessScore = scoreBusiness(pi, discovery)
	report.RequirementsScore = scoreRequirements(discovery, questions, answers)
	report.ConstraintsScore = scoreConstraints(pi, discovery)

	avg := (int(report.IntentScore) + int(report.VisionScore) + int(report.UXScore) + int(report.BusinessScore) + int(report.RequirementsScore) + int(report.ConstraintsScore)) / 6
	penalty := int(ambiguity.Score) / 4
	report.IntentConfidence = boundedScore(avg - penalty)
	report.Strengths = strengths(report)
	report.Weaknesses = weaknesses(report)
	report.Recommendations = recommendations(report)
	return report
}

func fieldCoverage(values []string, points int) int {
	score := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			score += points
		}
	}
	return score
}

func listCoverage(values []string, points int) int {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return points
	}
	return points * 2
}

func scoreBusiness(pi intentv3.ProductIntent, discovery *intentv3.DiscoveryResult) Score {
	score := 0
	if strings.TrimSpace(pi.ExpectedOutcome) != "" {
		score += 35
	}
	if strings.TrimSpace(pi.SuccessDefinition) != "" {
		score += 25
	}
	if strings.TrimSpace(pi.FailureDefinition) != "" {
		score += 20
	}
	if discovery != nil && len(discovery.Objectives) > 0 {
		score += 20
	}
	return boundedScore(score)
}

func scoreRequirements(discovery *intentv3.DiscoveryResult, questions []discoveryv3.Question, answers []discoveryv3.Answer) Score {
	score := 0
	if discovery != nil {
		score += min(len(discovery.Objectives)*12, 30)
		score += min(len(discovery.Expectations)*10, 20)
		if len(discovery.Gaps) == 0 {
			score += 15
		}
	}
	if len(questions) == 0 {
		return boundedScore(score)
	}
	answeredRequired, totalRequired := 0, 0
	answered := map[discoveryv3.QuestionID]bool{}
	for _, a := range answers {
		answered[a.QuestionID] = true
	}
	for _, q := range questions {
		if q.Required {
			totalRequired++
			if answered[q.ID] {
				answeredRequired++
			}
		}
	}
	if totalRequired > 0 {
		score += (answeredRequired * 35) / totalRequired
	}
	return boundedScore(score)
}

func scoreConstraints(pi intentv3.ProductIntent, discovery *intentv3.DiscoveryResult) Score {
	score := 0
	if len(pi.NonExpectations) > 0 {
		score += 35
	}
	if discovery != nil && len(discovery.Restrictions) > 0 {
		score += 45
	}
	if strings.TrimSpace(pi.FailureDefinition) != "" {
		score += 20
	}
	return boundedScore(score)
}

func strengths(r ConfidenceReport) []string {
	var out []string
	if r.IntentScore >= 70 {
		out = append(out, "intent is well defined")
	}
	if r.UXScore >= 70 {
		out = append(out, "desired experience is clear")
	}
	if r.BusinessScore >= 70 {
		out = append(out, "business outcome is clear")
	}
	if r.RequirementsScore >= 70 {
		out = append(out, "requirements context is strong")
	}
	if r.ConstraintsScore >= 70 {
		out = append(out, "constraints are explicit")
	}
	return out
}

func weaknesses(r ConfidenceReport) []string {
	var out []string
	if r.IntentScore < 70 {
		out = append(out, "intent definition is incomplete")
	}
	if r.VisionScore < 70 {
		out = append(out, "vision/outcome clarity is low")
	}
	if r.UXScore < 70 {
		out = append(out, "UX expectations need clarification")
	}
	if r.BusinessScore < 70 {
		out = append(out, "business success criteria need clarification")
	}
	if r.RequirementsScore < 70 {
		out = append(out, "requirements discovery needs more answers")
	}
	if r.ConstraintsScore < 70 {
		out = append(out, "constraints and non-expectations are weak")
	}
	return out
}

func recommendations(r ConfidenceReport) []string {
	var out []string
	if r.IntentConfidence < 80 {
		out = append(out, "continue progressive discovery before approving downstream planning")
	}
	if r.RequirementsScore < 70 {
		out = append(out, "answer required discovery questions")
	}
	if r.ConstraintsScore < 70 {
		out = append(out, "capture restrictions and non-expectations")
	}
	if len(out) == 0 {
		out = append(out, "intent confidence is sufficient for the next alignment phase")
	}
	return out
}

func boundedScore(score int) Score {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return Score(score)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

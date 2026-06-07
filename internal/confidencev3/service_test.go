package confidencev3_test

import (
	"testing"

	"github.com/Durru/plan-ai/internal/ambiguityv3"
	"github.com/Durru/plan-ai/internal/confidencev3"
	"github.com/Durru/plan-ai/internal/discoveryv3"
	"github.com/Durru/plan-ai/internal/intentv3"
)

func TestEvaluateHighConfidenceIntent(t *testing.T) {
	svc := confidencev3.NewService()
	pi := intentv3.ProductIntent{
		ID:                "pintent_1",
		Description:       "Build a CRM",
		ExpectedOutcome:   "Sales teams track deals",
		DesiredExperience: "Fast and focused",
		DesiredResult:     "Pipeline visibility",
		UserExpectations:  []string{"simple", "reliable"},
		NonExpectations:   []string{"not a social network"},
		SuccessDefinition: "Teams create and close deals",
		FailureDefinition: "Users abandon setup",
	}
	discovery := &intentv3.DiscoveryResult{Objectives: []string{"track deals", "report pipeline"}, Expectations: []string{"multi-user"}, Restrictions: []string{"budget"}}
	questions := []discoveryv3.Question{{ID: "q1", Required: true}, {ID: "q2", Required: true}}
	answers := []discoveryv3.Answer{{QuestionID: "q1"}, {QuestionID: "q2"}}

	report := svc.Evaluate(pi, discovery, questions, answers, ambiguityv3.AmbiguityReport{Score: 0})

	if report.IntentConfidence < 80 {
		t.Fatalf("expected high confidence, got %d", report.IntentConfidence)
	}
	if len(report.Weaknesses) != 0 {
		t.Fatalf("expected no weaknesses, got %#v", report.Weaknesses)
	}
}

func TestEvaluateLowConfidenceIntent(t *testing.T) {
	svc := confidencev3.NewService()
	report := svc.Evaluate(intentv3.ProductIntent{ID: "pintent_1", Description: "Build something"}, nil, nil, nil, ambiguityv3.AmbiguityReport{Score: 60})

	if report.IntentConfidence >= 50 {
		t.Fatalf("expected low confidence, got %d", report.IntentConfidence)
	}
	if len(report.Weaknesses) == 0 {
		t.Fatal("expected weaknesses")
	}
	if len(report.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}

func TestEvaluateRequirementsScoreUsesRequiredAnswers(t *testing.T) {
	svc := confidencev3.NewService()
	pi := intentv3.ProductIntent{ID: "pintent_1", Description: "Build CRM", ExpectedOutcome: "Track deals", DesiredResult: "Pipeline", SuccessDefinition: "Deals tracked"}
	questions := []discoveryv3.Question{{ID: "q1", Required: true}, {ID: "q2", Required: true}, {ID: "q3", Required: false}}
	answers := []discoveryv3.Answer{{QuestionID: "q1"}}

	report := svc.Evaluate(pi, nil, questions, answers, ambiguityv3.AmbiguityReport{})

	if report.RequirementsScore == 0 {
		t.Fatal("expected requirements score from answered required question")
	}
	if report.RequirementsScore >= 70 {
		t.Fatalf("expected partial requirements score below 70, got %d", report.RequirementsScore)
	}
}

func TestAmbiguityPenalizesConfidence(t *testing.T) {
	svc := confidencev3.NewService()
	pi := intentv3.ProductIntent{ID: "pintent_1", Description: "Build CRM", ExpectedOutcome: "Track deals", DesiredExperience: "Fast", DesiredResult: "Pipeline", SuccessDefinition: "Deals tracked", FailureDefinition: "No adoption", UserExpectations: []string{"simple"}, NonExpectations: []string{"complex"}}

	lowAmbiguity := svc.Evaluate(pi, nil, nil, nil, ambiguityv3.AmbiguityReport{Score: 0})
	highAmbiguity := svc.Evaluate(pi, nil, nil, nil, ambiguityv3.AmbiguityReport{Score: 80})

	if highAmbiguity.IntentConfidence >= lowAmbiguity.IntentConfidence {
		t.Fatalf("expected ambiguity penalty; low=%d high=%d", lowAmbiguity.IntentConfidence, highAmbiguity.IntentConfidence)
	}
}

package ambiguityv3_test

import (
	"testing"

	"github.com/Durru/plan-ai/internal/ambiguityv3"
	"github.com/Durru/plan-ai/internal/discoveryv3"
	"github.com/Durru/plan-ai/internal/intentv3"
)

func TestAnalyzeProductIntentReportsMissingInformation(t *testing.T) {
	svc := ambiguityv3.NewService()
	report := svc.AnalyzeProductIntent(intentv3.ProductIntent{ID: "pintent_1", Description: "Build a CRM"}, nil, nil)

	if report.Score == 0 {
		t.Fatal("expected non-zero ambiguity score")
	}
	if len(report.MissingInformation) == 0 {
		t.Fatal("expected missing information")
	}
	if !containsNeed(report.NeedsToKnow, "expected_outcome") {
		t.Fatalf("expected needs-to-know to include expected_outcome, got %#v", report.NeedsToKnow)
	}
}

func TestAnalyzeProductIntentTracksUnknownRequiredQuestions(t *testing.T) {
	svc := ambiguityv3.NewService()
	questions := []discoveryv3.Question{
		{ID: "q1", Level: discoveryv3.LevelProject, Question: "Who is the user?", Required: true},
		{ID: "q2", Level: discoveryv3.LevelProject, Question: "What is optional?", Required: false},
	}
	answers := []discoveryv3.Answer{{QuestionID: "q2", IntentID: "pintent_1", Answer: "optional answer"}}
	report := svc.AnalyzeProductIntent(intentv3.ProductIntent{ID: "pintent_1", Description: "Build a CRM", ExpectedOutcome: "Track deals", DesiredExperience: "Fast", DesiredResult: "Pipeline", SuccessDefinition: "Deals tracked", FailureDefinition: "No adoption", UserExpectations: []string{"simple"}}, questions, answers)

	if len(report.UnknownAreas) != 1 {
		t.Fatalf("expected 1 unknown area, got %d", len(report.UnknownAreas))
	}
	if report.UnknownAreas[0].QuestionID != "q1" {
		t.Fatalf("expected q1 unknown, got %s", report.UnknownAreas[0].QuestionID)
	}
}

func TestAnalyzeProductIntentDetectsConflicts(t *testing.T) {
	svc := ambiguityv3.NewService()
	report := svc.AnalyzeProductIntent(intentv3.ProductIntent{ID: "pintent_1", Description: "Build a simple but complex enterprise and consumer app"}, nil, nil)

	if len(report.Conflicts) < 2 {
		t.Fatalf("expected conflicts, got %#v", report.Conflicts)
	}
}

func TestAnalyzeTextDetectsVagueLanguage(t *testing.T) {
	svc := ambiguityv3.NewService()
	report := svc.AnalyzeText("Maybe build something simple later")

	if report.Score == 0 {
		t.Fatal("expected ambiguity score")
	}
	if len(report.Assumptions) == 0 {
		t.Fatal("expected vague-language assumptions")
	}
	if !containsNeed(report.NeedsToKnow, "success definition") {
		t.Fatalf("expected success definition need, got %#v", report.NeedsToKnow)
	}
}

func containsNeed(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

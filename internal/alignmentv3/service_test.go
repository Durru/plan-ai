package alignmentv3_test

import (
	"testing"

	"github.com/plan-ai/plan-ai/internal/alignmentv3"
	"github.com/plan-ai/plan-ai/internal/intentv3"
)

func approvedIntent() intentv3.ProductIntent {
	return intentv3.ProductIntent{
		ID:                "pintent_1",
		Description:       "Build a CRM for sales teams",
		ExpectedOutcome:   "Sales teams track deals",
		DesiredExperience: "Fast focused workflow",
		DesiredResult:     "Pipeline visibility",
		UserExpectations:  []string{"simple", "reliable"},
		NonExpectations:   []string{"not a social network"},
		SuccessDefinition: "Deals are created and closed",
		FailureDefinition: "Users abandon setup",
		Status:            intentv3.StatusApproved,
	}
}

func TestRegistryRequiresApprovedIntent(t *testing.T) {
	svc := alignmentv3.NewService()
	pi := approvedIntent()
	reg := svc.Registry(pi)
	if !reg.Approved {
		t.Fatal("expected approved registry")
	}
	pi.Status = intentv3.StatusDraft
	if svc.Registry(pi).Approved {
		t.Fatal("draft intent should not be approved")
	}
}

func TestOutcomeValidationDetectsGap(t *testing.T) {
	svc := alignmentv3.NewService()
	report := svc.Outcome(approvedIntent(), "A docs website")
	if report.Valid {
		t.Fatal("expected outcome gap")
	}
	if len(report.GapAnalysis) == 0 {
		t.Fatal("expected gap analysis")
	}
}

func TestPlanAndTaskAlignment(t *testing.T) {
	svc := alignmentv3.NewService()
	pi := approvedIntent()
	plan := svc.Plan(pi, "Build CRM pipeline tracking so sales teams track deals")
	if !plan.Relevant {
		t.Fatalf("expected relevant plan, got %#v", plan)
	}
	task := svc.Task(pi, "Add unrelated weather widget")
	if task.Relevant {
		t.Fatalf("expected unrelated task to be irrelevant")
	}
}

func TestFrameworkRequiresApprovedIntent(t *testing.T) {
	svc := alignmentv3.NewService()
	pi := approvedIntent()
	if !svc.Framework(pi).Ready {
		t.Fatal("approved intent should be framework-ready")
	}
	pi.Status = intentv3.StatusDraft
	if svc.Framework(pi).Ready {
		t.Fatal("draft intent should not be framework-ready")
	}
}

func TestReviewProducesRecommendation(t *testing.T) {
	svc := alignmentv3.NewService()
	report := svc.Review(approvedIntent(), "Docs only", "Unrelated docs plan", "Weather widget")
	if len(report.Risks) == 0 {
		t.Fatal("expected risks")
	}
	if len(report.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}

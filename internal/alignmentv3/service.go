package alignmentv3

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/plan-ai/plan-ai/internal/intentv3"
)

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) Registry(pi intentv3.ProductIntent) RegistryReport {
	return RegistryReport{
		IntentID:             pi.ID,
		Approved:             pi.Status == intentv3.StatusApproved,
		ApprovedExpectations: pi.UserExpectations,
		ApprovedPreferences:  nonEmptyList(pi.DesiredExperience),
		ApprovedUX:           pi.DesiredExperience,
		ApprovedOutcomes:     nonEmptyList(pi.ExpectedOutcome, pi.DesiredResult),
		ApprovedReferences:   nil,
	}
}

func (s Service) KnowledgeGraph(pi intentv3.ProductIntent) KnowledgeGraphReport {
	links := []TraceLink{
		{FromType: "intent", FromID: pi.ID, ToType: "outcome", ToID: stableID("outcome", pi.ExpectedOutcome), Reason: "expected outcome derives from approved intent"},
		{FromType: "intent", FromID: pi.ID, ToType: "experience", ToID: stableID("ux", pi.DesiredExperience), Reason: "desired experience derives from approved intent"},
		{FromType: "intent", FromID: pi.ID, ToType: "success", ToID: stableID("success", pi.SuccessDefinition), Reason: "success criteria preserve intent"},
	}
	return KnowledgeGraphReport{IntentID: pi.ID, Links: links}
}

func (s Service) Consistency(pi intentv3.ProductIntent) ConsistencyReport {
	text := strings.ToLower(strings.Join([]string{pi.Description, pi.ExpectedOutcome, pi.DesiredExperience, pi.DesiredResult, strings.Join(pi.UserExpectations, " "), strings.Join(pi.NonExpectations, " ")}, " "))
	conflicts := detectConflicts(text)
	return ConsistencyReport{IntentID: pi.ID, Consistent: len(conflicts) == 0, Conflicts: conflicts, Drift: nil}
}

func (s Service) Outcome(pi intentv3.ProductIntent, currentOutcome string) OutcomeReport {
	gaps := []string{}
	if strings.TrimSpace(pi.ExpectedOutcome) == "" {
		gaps = append(gaps, "expected outcome is missing")
	}
	if strings.TrimSpace(currentOutcome) == "" {
		gaps = append(gaps, "current outcome is missing")
	} else if pi.ExpectedOutcome != "" && !overlaps(pi.ExpectedOutcome, currentOutcome) {
		gaps = append(gaps, "current outcome does not mention expected outcome terms")
	}
	return OutcomeReport{IntentID: pi.ID, ExpectedOutcome: pi.ExpectedOutcome, CurrentOutcome: currentOutcome, GapAnalysis: gaps, Valid: len(gaps) == 0}
}

func (s Service) UX(pi intentv3.ProductIntent) UXReport {
	goals := nonEmptyList(pi.DesiredExperience)
	rules := append([]string{}, pi.UserExpectations...)
	for _, avoid := range pi.NonExpectations {
		rules = append(rules, "avoid: "+avoid)
	}
	score := Score(40)
	if pi.DesiredExperience != "" {
		score += 30
	}
	if len(pi.UserExpectations) > 0 {
		score += 15
	}
	if len(pi.NonExpectations) > 0 {
		score += 15
	}
	return UXReport{IntentID: pi.ID, Goals: goals, Rules: rules, References: nil, Consistency: bound(score)}
}

func (s Service) Feature(pi intentv3.ProductIntent, feature string) FeatureMapping {
	return FeatureMapping{Feature: feature, IntentID: pi.ID, Outcome: pi.ExpectedOutcome, Vision: pi.DesiredResult, SuccessCriteria: pi.SuccessDefinition, Purpose: purpose(feature, pi)}
}

func (s Service) Plan(pi intentv3.ProductIntent, plan string) PlanAlignmentReport {
	score := alignmentScore(pi, plan)
	findings := []string{}
	if score < 50 {
		findings = append(findings, "plan does not reference enough intent terms")
	}
	if pi.Status != intentv3.StatusApproved {
		findings = append(findings, "intent is not approved")
	}
	return PlanAlignmentReport{IntentID: pi.ID, Plan: plan, Score: score, Relevant: score >= 50 && pi.Status == intentv3.StatusApproved, Findings: findings}
}

func (s Service) Task(pi intentv3.ProductIntent, task string) TaskAlignmentReport {
	relevant := alignmentScore(pi, task) >= 40
	reason := "task maps to intent terms"
	if !relevant {
		reason = "task lacks clear purpose against intent"
	}
	return TaskAlignmentReport{IntentID: pi.ID, Task: task, Relevant: relevant, Reason: reason}
}

func (s Service) Continuous(pi intentv3.ProductIntent, currentOutcome, plan, task string) ContinuousAlignmentReport {
	report := ContinuousAlignmentReport{Health: 100}
	if !s.Outcome(pi, currentOutcome).Valid {
		report.OutcomeDrift = append(report.OutcomeDrift, "current outcome diverges from expected outcome")
	}
	if !s.Plan(pi, plan).Relevant {
		report.PlanningDrift = append(report.PlanningDrift, "plan is weakly aligned")
	}
	if !s.Task(pi, task).Relevant {
		report.ExecutionDrift = append(report.ExecutionDrift, "task is weakly aligned")
	}
	drift := len(report.IntentDrift) + len(report.VisionDrift) + len(report.OutcomeDrift) + len(report.PlanningDrift) + len(report.ExecutionDrift)
	report.Health = bound(Score(100 - drift*20))
	return report
}

func (s Service) References() []ReferenceProduct {
	return []ReferenceProduct{
		{Name: "Linear", Screens: []string{"issues", "cycles"}, UX: []string{"fast", "keyboard-first"}, Workflows: []string{"triage", "planning"}, Components: []string{"sidebar", "command menu"}},
		{Name: "Notion", Screens: []string{"docs", "databases"}, UX: []string{"flexible", "document-first"}, Workflows: []string{"knowledge capture"}, Components: []string{"blocks", "tables"}},
		{Name: "Stripe", Screens: []string{"dashboard", "checkout"}, UX: []string{"trustworthy", "precise"}, Workflows: []string{"billing", "payments"}, Components: []string{"forms", "status badges"}},
		{Name: "GitHub", Screens: []string{"issues", "pull requests"}, UX: []string{"collaborative", "auditable"}, Workflows: []string{"review", "merge"}, Components: []string{"diff", "timeline"}},
		{Name: "Slack", Screens: []string{"channels", "threads"}, UX: []string{"conversational"}, Workflows: []string{"team communication"}, Components: []string{"messages", "notifications"}},
		{Name: "Monday", Screens: []string{"boards", "automations"}, UX: []string{"visual", "workflow-oriented"}, Workflows: []string{"tracking", "automation"}, Components: []string{"board", "status columns"}},
	}
}

func (s Service) DNA(pi intentv3.ProductIntent) ProductDNA {
	return ProductDNA{ProductID: pi.ID, ProductDNA: nonEmptyList(pi.Description, pi.ExpectedOutcome), DesignDNA: nonEmptyList(pi.DesiredExperience), BusinessDNA: nonEmptyList(pi.DesiredResult, pi.SuccessDefinition), TechnicalDNA: []string{"additive", "deterministic", "sandbox-safe"}}
}

func (s Service) Impact(pi intentv3.ProductIntent, change string) IntentImpactReport {
	score := alignmentScore(pi, change)
	impact := IntentImpactReport{IntentID: pi.ID, IntentImpactScore: score}
	impact.TechnicalImpact = []string{"implementation scope must preserve approved intent"}
	impact.FunctionalImpact = []string{fmt.Sprintf("change alignment score: %d%%", score)}
	impact.UXImpact = nonEmptyList(pi.DesiredExperience)
	impact.BusinessImpact = nonEmptyList(pi.ExpectedOutcome)
	impact.VisionImpact = nonEmptyList(pi.DesiredResult)
	return impact
}

func (s Service) Context(pi intentv3.ProductIntent) AlignmentContext {
	return AlignmentContext{IntentID: pi.ID, WhatToDo: pi.Description, WhyItExists: pi.ExpectedOutcome, DesiredOutcome: pi.DesiredResult, Avoid: pi.NonExpectations, ContextSummary: fmt.Sprintf("Build %q to achieve %q while avoiding %s", pi.Description, pi.ExpectedOutcome, strings.Join(pi.NonExpectations, ", "))}
}

func (s Service) Review(pi intentv3.ProductIntent, currentOutcome, plan, task string) ProductReviewReport {
	outcome := s.Outcome(pi, currentOutcome)
	planReport := s.Plan(pi, plan)
	taskReport := s.Task(pi, task)
	review := ProductReviewReport{IntentID: pi.ID}
	review.IntentReview = alignmentScore(pi, pi.Description+" "+pi.ExpectedOutcome+" "+pi.DesiredResult)
	review.VisionReview = alignmentScore(pi, pi.DesiredResult)
	if outcome.Valid {
		review.OutcomeReview = 90
	} else {
		review.OutcomeReview = 45
		review.Risks = append(review.Risks, outcome.GapAnalysis...)
	}
	review.ProjectReview = bound((review.IntentReview + review.VisionReview + review.OutcomeReview + planReport.Score) / 4)
	if taskReport.Relevant {
		review.AlignmentReview = bound((review.ProjectReview + 80) / 2)
	} else {
		review.AlignmentReview = bound((review.ProjectReview + 30) / 2)
		review.Risks = append(review.Risks, taskReport.Reason)
	}
	if review.AlignmentReview < 75 {
		review.Recommendations = append(review.Recommendations, "realign plan/task with Product Intent before implementation")
	} else {
		review.Recommendations = append(review.Recommendations, "alignment is sufficient for implementation review")
	}
	return review
}

func (s Service) Framework(pi intentv3.ProductIntent) FrameworkReport {
	stages := []string{"Intent", "Discovery", "Approval", "Vision", "Outcome", "Knowledge", "Planning", "Execution", "Validation", "Alignment"}
	ready := pi.Status == intentv3.StatusApproved && pi.Description != "" && pi.ExpectedOutcome != ""
	summary := "Intent-to-implementation framework requires approved Product Intent"
	if ready {
		summary = "Intent-to-implementation framework is ready"
	}
	return FrameworkReport{IntentID: pi.ID, Stages: stages, Ready: ready, Summary: summary}
}

func alignmentScore(pi intentv3.ProductIntent, text string) Score {
	text = strings.ToLower(text)
	terms := strings.Fields(strings.ToLower(strings.Join([]string{pi.Description, pi.ExpectedOutcome, pi.DesiredExperience, pi.DesiredResult, pi.SuccessDefinition}, " ")))
	if len(terms) == 0 {
		return 0
	}
	hits := 0
	seen := map[string]bool{}
	for _, term := range terms {
		term = strings.Trim(term, ",.;:!?()[]{}")
		if len(term) < 4 || seen[term] {
			continue
		}
		seen[term] = true
		if strings.Contains(text, term) {
			hits++
		}
	}
	if len(seen) == 0 {
		return 0
	}
	return bound(Score((hits * 100) / len(seen)))
}

func purpose(feature string, pi intentv3.ProductIntent) string {
	if strings.TrimSpace(feature) == "" {
		return "feature must be defined before purpose can be assessed"
	}
	return fmt.Sprintf("%s exists to support %s", feature, fallback(pi.ExpectedOutcome, pi.DesiredResult, pi.Description))
}

func detectConflicts(text string) []string {
	pairs := []struct{ a, b string }{{"simple", "complex"}, {"minimal", "feature-rich"}, {"enterprise", "consumer"}, {"cheap", "premium"}}
	var out []string
	for _, p := range pairs {
		if strings.Contains(text, p.a) && strings.Contains(text, p.b) {
			out = append(out, p.a+" conflicts with "+p.b)
		}
	}
	return out
}

func overlaps(a, b string) bool { return alignmentScore(intentv3.ProductIntent{Description: a}, b) > 0 }

func nonEmptyList(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}

func stableID(prefix, value string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(value))))
	return fmt.Sprintf("%s_%08x", prefix, h.Sum32())
}
func fallback(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "approved intent"
}
func bound(s Score) Score {
	if s < 0 {
		return 0
	}
	if s > 100 {
		return 100
	}
	return s
}

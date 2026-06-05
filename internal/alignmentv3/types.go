package alignmentv3

import "github.com/plan-ai/plan-ai/internal/intentv3"

type Score int

type RegistryReport struct {
	IntentID             string
	Approved             bool
	ApprovedExpectations []string
	ApprovedPreferences  []string
	ApprovedUX           string
	ApprovedOutcomes     []string
	ApprovedReferences   []string
}

type TraceLink struct {
	FromType string
	FromID   string
	ToType   string
	ToID     string
	Reason   string
}

type KnowledgeGraphReport struct {
	IntentID string
	Links    []TraceLink
}

type ConsistencyReport struct {
	IntentID   string
	Consistent bool
	Conflicts  []string
	Drift      []string
}

type OutcomeReport struct {
	IntentID        string
	ExpectedOutcome string
	CurrentOutcome  string
	GapAnalysis     []string
	Valid           bool
}

type UXReport struct {
	IntentID    string
	Goals       []string
	Rules       []string
	References  []string
	Consistency Score
}

type FeatureMapping struct {
	Feature         string
	IntentID        string
	Outcome         string
	Vision          string
	SuccessCriteria string
	Purpose         string
}

type PlanAlignmentReport struct {
	IntentID string
	Plan     string
	Score    Score
	Relevant bool
	Findings []string
}

type TaskAlignmentReport struct {
	IntentID string
	Task     string
	Relevant bool
	Reason   string
}

type ContinuousAlignmentReport struct {
	IntentDrift    []string
	VisionDrift    []string
	OutcomeDrift   []string
	PlanningDrift  []string
	ExecutionDrift []string
	Health         Score
}

type ReferenceProduct struct {
	Name       string
	Screens    []string
	UX         []string
	Workflows  []string
	Components []string
}

type ProductDNA struct {
	ProductID    string
	ProductDNA   []string
	DesignDNA    []string
	BusinessDNA  []string
	TechnicalDNA []string
}

type IntentImpactReport struct {
	IntentID          string
	TechnicalImpact   []string
	FunctionalImpact  []string
	UXImpact          []string
	BusinessImpact    []string
	VisionImpact      []string
	IntentImpactScore Score
}

type AlignmentContext struct {
	IntentID       string
	WhatToDo       string
	WhyItExists    string
	DesiredOutcome string
	Avoid          []string
	ContextSummary string
}

type ProductReviewReport struct {
	IntentID        string
	ProjectReview   Score
	IntentReview    Score
	VisionReview    Score
	OutcomeReview   Score
	AlignmentReview Score
	Risks           []string
	Recommendations []string
}

type FrameworkReport struct {
	IntentID string
	Stages   []string
	Ready    bool
	Summary  string
}

type Input struct {
	Intent         intentv3.ProductIntent
	CurrentOutcome string
	Plan           string
	Task           string
	Feature        string
}

package agent

import "testing"

func TestDetectIntent_AnalyzeProject(t *testing.T) {
	d := NewIntentDetector()
	cases := []string{
		"analyze my project",
		"summarize the project state",
		"what do I have so far",
		"show me what we have",
		"where are we",
		"audit the project",
	}
	for _, input := range cases {
		got := d.DetectIntent(input)
		if got != IntentAnalyzeProject {
			t.Errorf("DetectIntent(%q) = %q, want IntentAnalyzeProject", input, got)
		}
	}
}

func TestDetectIntent_CreateProduct(t *testing.T) {
	d := NewIntentDetector()
	cases := []string{
		"create a SaaS for task management",
		"build an app for team collaboration",
		"make a product for email marketing",
		"start a platform for online courses",
		"create a new application",
	}
	for _, input := range cases {
		got := d.DetectIntent(input)
		if got != IntentCreateProduct {
			t.Errorf("DetectIntent(%q) = %q, want IntentCreateProduct", input, got)
		}
	}
}

func TestDetectIntent_DatabasePlan(t *testing.T) {
	d := NewIntentDetector()
	cases := []string{
		"design a database schema for orders",
		"plan the postgres model",
		"create a db plan",
		"design the sqlite database",
	}
	for _, input := range cases {
		got := d.DetectIntent(input)
		if got != IntentDatabasePlan {
			t.Errorf("DetectIntent(%q) = %q, want IntentDatabasePlan", input, got)
		}
	}
}

func TestDetectIntent_ImpactAnalysis(t *testing.T) {
	d := NewIntentDetector()
	cases := []string{
		"what if I change the auth system",
		"what would happen if I remove the cache",
		"analyze impact of modifying the API",
		"what are the consequences of this change",
		"assess risks of the new architecture",
	}
	for _, input := range cases {
		got := d.DetectIntent(input)
		if got != IntentImpactAnalysis {
			t.Errorf("DetectIntent(%q) = %q, want IntentImpactAnalysis", input, got)
		}
	}
}

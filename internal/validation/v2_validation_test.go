package validation

import (
	"strings"
	"testing"
)

func TestV2Cases_NotEmpty(t *testing.T) {
	cases := V2Cases()
	if len(cases) == 0 {
		t.Fatal("V2Cases() must return at least one case")
	}
}

func TestV2Stages_NotEmpty(t *testing.T) {
	stages := V2Stages()
	if len(stages) == 0 {
		t.Fatal("V2Stages() must return at least one stage")
	}
}

func TestV2Stages_Count(t *testing.T) {
	if got := len(V2Stages()); got != 9 {
		t.Errorf("expected 9 V2 stages, got %d", got)
	}
}

func TestV2Cases_Count(t *testing.T) {
	if got := len(V2Cases()); got != 7 {
		t.Errorf("expected 7 V2 cases, got %d", got)
	}
}

func TestValidateV2Cases_TotalChecks(t *testing.T) {
	summary := ValidateV2Cases()
	expected := len(V2Cases()) * len(V2Stages()) // 7 * 9 = 63
	if summary.Total != expected {
		t.Errorf("expected %d total checks, got %d", expected, summary.Total)
	}
}

func TestValidateV2Cases_AllPass(t *testing.T) {
	summary := ValidateV2Cases()
	if summary.Passed != summary.Total {
		t.Errorf("expected all %d checks to pass, but %d failed",
			summary.Total, summary.Failed)
		for _, r := range summary.Results {
			if !r.Passed {
				t.Logf("  FAIL: case=%q stage=%q detail=%s", r.CaseName, r.StageName, r.Detail)
			}
		}
	}
}

func TestValidateV2Cases_EveryCaseHasAllStages(t *testing.T) {
	summary := ValidateV2Cases()
	stages := V2Stages()

	for _, c := range V2Cases() {
		for _, s := range stages {
			found := false
			for _, r := range summary.Results {
				if r.CaseName == c.Name && r.StageName == s.Name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("missing result for case=%q stage=%q", c.Name, s.Name)
			}
		}
	}
}

func TestDetectIntents(t *testing.T) {
	tests := []struct {
		idea     string
		expected []string
		want     int
	}{
		{"Build a subscription billing system", []string{"subscription", "billing", "auth"}, 2},
		{"Build a landing page", []string{"ecommerce", "analytics"}, 0},
		{"CRM with contacts and pipeline", []string{"CRM", "contacts", "pipeline"}, 3},
		{"", []string{"any"}, 0},
	}
	for _, tt := range tests {
		got := detectIntents(tt.idea, tt.expected)
		if len(got) != tt.want {
			t.Errorf("detectIntents(%q, %v) = %v (len=%d), want %d detected", tt.idea, tt.expected, got, len(got), tt.want)
		}
	}
}

func TestV2Cases_EachHasIdea(t *testing.T) {
	for _, c := range V2Cases() {
		if strings.TrimSpace(c.Idea) == "" {
			t.Errorf("case %q has empty idea", c.Name)
		}
	}
}

func TestV2Cases_EachHasIntents(t *testing.T) {
	for _, c := range V2Cases() {
		if len(c.ExpectedIntents) == 0 {
			t.Errorf("case %q has no expected intents", c.Name)
		}
	}
}

func TestV2Cases_EachHasVisionDimensions(t *testing.T) {
	for _, c := range V2Cases() {
		if len(c.ExpectedVisionDimensions) == 0 {
			t.Errorf("case %q has no expected vision dimensions", c.Name)
		}
	}
}

func TestV2Stages_EachHasName(t *testing.T) {
	for _, s := range V2Stages() {
		if strings.TrimSpace(s.Name) == "" {
			t.Error("found a V2Stage with empty name")
		}
	}
}

func BenchmarkValidateV2Cases(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateV2Cases()
	}
}

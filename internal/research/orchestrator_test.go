package research

import "testing"

func TestBuildRunCreatesEvidence(t *testing.T) {
	run := BuildRun("project", AgentSecurity, "checkout security")
	if run.Agent != AgentSecurity || len(run.Evidence) == 0 || run.Status != ResearchStatusCompleted {
		t.Fatalf("unexpected run: %#v", run)
	}
}

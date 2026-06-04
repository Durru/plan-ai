package intent

import "testing"

func TestDetectorCreatesSaaSCRMIntentProfile(t *testing.T) {
	profile := NewDetector().Detect("project", "quiero un SaaS CRM")

	if profile.PrimaryIntent.Name != "CRM" {
		t.Fatalf("primary intent = %q, want CRM", profile.PrimaryIntent.Name)
	}
	for _, expected := range []string{"SaaS", "multi-user", "admin panel", "subscriptions", "reports", "automations"} {
		if !hasGoal(profile.SecondaryGoals, expected) {
			t.Fatalf("missing candidate goal %q in %#v", expected, profile.SecondaryGoals)
		}
	}
	if profile.Approved {
		t.Fatal("detected intent must not be approved automatically")
	}
}

func hasGoal(goals []Goal, name string) bool {
	for _, goal := range goals {
		if goal.Name == name && goal.State == SignalCandidate {
			return true
		}
	}
	return false
}

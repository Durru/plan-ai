package agent

import "testing"

func TestBuildSubagentTaskIsTemporaryAndIsolated(t *testing.T) {
	task := BuildSubagentTask("project", SubagentSecurity, "Review auth")
	if !task.Temporary || !task.Isolated || task.MemoryPolicy != "no-independent-persistent-memory" {
		t.Fatalf("unexpected isolation policy: %#v", task)
	}
	if task.Capability != "validation" || task.Status != JobStatusPending {
		t.Fatalf("unexpected task routing: %#v", task)
	}
}

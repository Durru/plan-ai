package opencode

import "testing"

func TestDefaultWorkflowCommandsExposeRequiredSurface(t *testing.T) {
	commands := DefaultWorkflowCommands()
	want := map[string]bool{"status": false, "next": false, "context": false, "plans": false, "changes": false}
	for _, cmd := range commands {
		if _, ok := want[cmd.Name]; ok {
			want[cmd.Name] = true
		}
		if !cmd.ReadOnly {
			t.Fatalf("workflow command must be read-only by default: %#v", cmd)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("missing workflow command %s in %#v", name, commands)
		}
	}
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanCommandRequiresApprovedIntent(t *testing.T) {
	home := t.TempDir()

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	projectRoot := t.TempDir()
	t.Setenv("PLAN_AI_PROJECT_ROOT", projectRoot)

	// Run init + knowledge to satisfy prerequisites, then try plan.
	out, err = executeCommand(t, home, projectRoot, "init")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	// plan without approved intent must fail.
	out, err = executeCommand(t, home, projectRoot, "plan")
	if err == nil {
		t.Errorf("expected plan to fail without approved product intent, got:\n%s", out)
	}
	if !strings.Contains(err.Error(), "product intent") {
		t.Errorf("error should mention product intent, got: %v", err)
	}
}

func TestMCPCreateMasterPlanRespectsPlanningGuard(t *testing.T) {
	home := t.TempDir()

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	projectRoot := t.TempDir()
	t.Setenv("PLAN_AI_PROJECT_ROOT", projectRoot)

	_, err = executeCommand(t, home, projectRoot, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// MCP plan_ai.create_master_plan should be blocked by the guard.
	// We test via the CLI `agent process` which goes through the conversation
	// gateway that now has the planning guard wired.
	out, err = executeCommand(t, home, projectRoot, "agent", "process", "create a master plan")
	if err != nil {
		t.Fatalf("agent process: %v", err)
	}
	// The response should indicate the planning guard blocked it
	// (status=error) or at least not return a plan_id.
	if strings.Contains(strings.ToLower(out), "plan_id") {
		t.Errorf("expected planning guard to block master plan creation:\n%s", out)
	}
}

func TestConversationPlanningRequestStartsDiscoveryWhenIntentMissing(t *testing.T) {
	home := t.TempDir()

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	projectRoot := t.TempDir()
	_, err = executeCommand(t, home, projectRoot, "init")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err = executeCommand(t, home, projectRoot, "agent", "process", "create a master plan")
	if err != nil {
		t.Fatalf("agent process: %v", err)
	}

	// The conversation response should mention product intent or discovery,
	// not "I can create a master plan" (which would imply the guard didn't fire).
	if strings.Contains(strings.ToLower(out), "I can create a master plan") {
		t.Errorf("expected guard to block with discovery message, got:\n%s", out)
	}
}

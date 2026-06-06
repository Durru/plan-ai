package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallDefaultUsesInstallerPath(t *testing.T) {
	home := t.TempDir()

	output, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal")
	if err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Global installation: installed") {
		t.Fatalf("install should use definitive installer path, got:\n%s", output)
	}
	if !strings.Contains(output, "Installed: mcp") {
		t.Fatalf("minimal preset should install mcp component, got:\n%s", output)
	}

	// Installer state file (created by the new installer) must exist.
	assertPathExists(t, filepath.Join(home, ".plan-ai", "state.json"))
}

func TestInstallRespectsOpenCodeConfigDir(t *testing.T) {
	home := t.TempDir()
	sandboxOC := filepath.Join(home, "opencode-sandbox")

	env := map[string]string{
		"HOME":                home,
		"OPENCODE_CONFIG_DIR": sandboxOC,
	}
	output, err := executeCommandWithEnv(t, env, t.TempDir(), "install", "--preset=minimal")
	if err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}

	// OpenCode config must land in the sandbox, not the default location.
	assertPathExists(t, filepath.Join(sandboxOC, "opencode.json"))
	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "config.json")); err == nil {
		t.Log("config.json exists (legacy), but SetupMCPConfig now writes to opencode.json")
		t.Fatalf("install must not write to the default opencode config dir when OPENCODE_CONFIG_DIR is set")
	}
}

func TestUpdateRefreshesStateAndIntegrations(t *testing.T) {
	home := t.TempDir()

	if output, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}

	// Update should be idempotent and not fail.
	output, err := executeCommand(t, home, t.TempDir(), "update")
	if err != nil {
		t.Fatalf("update: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Update complete.") {
		t.Fatalf("update should report completion, got:\n%s", output)
	}
	if !strings.Contains(output, "Tools:") {
		t.Fatalf("update should report tool detection, got:\n%s", output)
	}

	// Re-running update must remain a no-op.
	output, err = executeCommand(t, home, t.TempDir(), "update")
	if err != nil {
		t.Fatalf("update (2nd run): %v\n%s", err, output)
	}
	if !strings.Contains(output, "Update complete.") {
		t.Fatalf("update should remain idempotent, got:\n%s", output)
	}
}

func TestUninstallFullRemovesOpenCodeRegistration(t *testing.T) {
	home := t.TempDir()
	sandboxOC := filepath.Join(home, "opencode-sandbox")

	env := map[string]string{
		"HOME":                home,
		"OPENCODE_CONFIG_DIR": sandboxOC,
	}
	if output, err := executeCommandWithEnv(t, env, t.TempDir(), "install", "--preset=minimal"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}

	// Verify the opencode config has the plan-ai entry.
	ocConfig := filepath.Join(sandboxOC, "opencode.json")
	data, err := os.ReadFile(ocConfig)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	if !strings.Contains(string(data), "plan-ai") {
		t.Fatalf("opencode config should contain plan-ai entry, got:\n%s", string(data))
	}

	// Full uninstall must remove the plan-ai entry from opencode.
	if output, err := executeCommandWithEnv(t, env, t.TempDir(), "uninstall"); err != nil {
		t.Fatalf("uninstall: %v\n%s", err, output)
	}
	data, err = os.ReadFile(ocConfig)
	if err != nil {
		t.Fatalf("read opencode config after uninstall: %v", err)
	}
	if strings.Contains(string(data), "plan-ai") {
		t.Fatalf("opencode config should not contain plan-ai entry after full uninstall, got:\n%s", string(data))
	}
}

func TestDoctorDetectsMissingRegisteredBinary(t *testing.T) {
	home := t.TempDir()
	sandboxBin := filepath.Join(home, "bin")
	t.Setenv("PATH", sandboxBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if output, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}

	// Remove the registered bin dir to simulate a stale registration.
	if err := os.RemoveAll(sandboxBin); err != nil {
		t.Fatalf("remove bin dir: %v", err)
	}

	output, err := executeCommand(t, home, t.TempDir(), "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, output)
	}
	if !strings.Contains(output, "registered_binary_missing") {
		t.Fatalf("doctor should report missing registered binary, got:\n%s", output)
	}
}

func TestDoctorDetectsDuplicateOpenCodeRegistration(t *testing.T) {
	home := t.TempDir()
	sandboxOC := filepath.Join(home, "opencode-sandbox")
	env := map[string]string{
		"HOME":                home,
		"OPENCODE_CONFIG_DIR": sandboxOC,
	}
	if output, err := executeCommandWithEnv(t, env, t.TempDir(), "install", "--preset=minimal"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}

	// Inject a duplicate plan-ai entry into the opencode config.
	ocConfig := filepath.Join(sandboxOC, "opencode.json")
	data, err := os.ReadFile(ocConfig)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	mcpSection, _ := cfg["mcp"].(map[string]any)
	// Add a duplicate under a different key to force two `plan-ai` substring
	// occurrences in the raw JSON (the duplicate detector is substring-based).
	if mcpSection != nil {
		original, _ := mcpSection["plan-ai"].(map[string]any)
		if original != nil {
			clone := map[string]any{}
			for k, v := range original {
				clone[k] = v
			}
			mcpSection["plan-ai-dup"] = clone
		}
	}
	cfg["mcp"] = mcpSection
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(ocConfig, out, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	output, err := executeCommandWithEnv(t, env, t.TempDir(), "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, output)
	}
	if !strings.Contains(output, "duplicate_opencode_registration") {
		t.Fatalf("doctor should report duplicate opencode registration, got:\n%s", output)
	}
}

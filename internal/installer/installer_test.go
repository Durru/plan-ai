package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ─────────────────────────────────────────────

func testInstaller(t *testing.T) (*Installer, string) {
	t.Helper()
	home := t.TempDir()
	inst := NewInstaller(home)
	return inst, home
}

// ── state.json ──────────────────────────────────────────

func TestInstaller_StateIsCreatedOnInstall(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	statePath := filepath.Join(home, ".plan-ai", "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatalf("state.json was not created at %s", statePath)
	}

	if err := inst.LoadState(); err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if inst.State.Version != "1" {
		t.Fatalf("state version = %q, want %q", inst.State.Version, "1")
	}
	if inst.State.InstalledAt == "" {
		t.Fatal("InstalledAt should be set")
	}
	if inst.State.Preset != "minimal" {
		t.Fatalf("preset = %q, want %q", inst.State.Preset, "minimal")
	}
}

// ── components ──────────────────────────────────────────

func TestInstaller_FullPlanAI_InstallsAllComponents(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "full-plan-ai",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	_ = inst.LoadState()
	allComponents := []string{"intent", "planning", "mcp", "opencode-agent", "docs", "context", "alignment"}
	for _, c := range allComponents {
		if !inst.State.Components[c].Installed {
			t.Errorf("component %q should be installed with full-plan-ai preset", c)
		}
	}
}

func TestInstaller_EcosystemOnly_InstallsOnlyEcosystem(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "ecosystem-only",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	_ = inst.LoadState()
	ecosystem := []string{"mcp", "opencode-agent", "docs"}
	nonEcosystem := []string{"intent", "planning", "context", "alignment"}

	for _, c := range ecosystem {
		if !inst.State.Components[c].Installed {
			t.Errorf("ecosystem component %q should be installed", c)
		}
	}
	for _, c := range nonEcosystem {
		if inst.State.Components[c].Installed {
			t.Errorf("non-ecosystem component %q should NOT be installed", c)
		}
	}
}

func TestInstaller_Minimal_InstallsOnlyMCP(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	_ = inst.LoadState()
	if !inst.State.Components["mcp"].Installed {
		t.Error("mcp component should be installed with minimal preset")
	}
	for name, cs := range inst.State.Components {
		if name != "mcp" && cs.Installed {
			t.Errorf("component %q should NOT be installed with minimal preset", name)
		}
	}
}

func TestInstaller_CustomPreset_InstallsSelectedOnly(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset:     "custom",
		Components: []string{"mcp", "docs"},
		BinDir:     filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	_ = inst.LoadState()
	if !inst.State.Components["mcp"].Installed {
		t.Error("mcp should be installed")
	}
	if !inst.State.Components["docs"].Installed {
		t.Error("docs should be installed")
	}
	if inst.State.Components["planning"].Installed {
		t.Error("planning should NOT be installed with custom mcp+docs")
	}
}

// ── dry-run ─────────────────────────────────────────────

func TestInstaller_DryRunDoesNotCreateState(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		DryRun:  true,
		Preset:  "full-plan-ai",
		BinDir:  filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install dry-run: %v", err)
	}

	statePath := filepath.Join(home, ".plan-ai", "state.json")
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("state.json was created despite --dry-run at %s", statePath)
	}
}

func TestInstaller_DryRunDoesNotTouchFiles(t *testing.T) {
	inst, home := testInstaller(t)

	// Create a fake opencode config to verify it's not touched
	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(ocDir, "opencode.json")
	origContent := `{"agent_name":"my-agent"}`
	if err := os.WriteFile(cfgPath, []byte(origContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set OPENCODE_CONFIG_DIR so the installer finds it
	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	err := inst.Install(InstallOptions{
		DryRun:    true,
		Preset:    "full-plan-ai",
		BinDir:    filepath.Join(home, "bin"),
		AllowReal: true,
	})
	if err != nil {
		t.Fatalf("Install dry-run: %v", err)
	}

	// Verify opencode config was NOT modified
	data, _ := os.ReadFile(cfgPath)
	if string(data) != origContent {
		t.Fatalf("opencode config was modified despite --dry-run:\n  want: %q\n  got:  %q", origContent, string(data))
	}
}

// ── tools detection ─────────────────────────────────────

func TestInstaller_DetectTools(t *testing.T) {
	inst, home := testInstaller(t)

	// Explicitly set OPENCODE_CONFIG_DIR to an empty dir so we don't
	// accidentally detect the real opencode installation.
	t.Setenv("OPENCODE_CONFIG_DIR", filepath.Join(home, ".config", "opencode"))

	// Create a fake bin dir with some tools
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a fake go binary (just a marker file)
	if err := os.WriteFile(filepath.Join(binDir, "go"), []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create a fake git binary
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}
	// Put binDir on PATH
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	tools := inst.DetectTools()
	if !tools.Go {
		t.Error("go should be detected")
	}
	if !tools.Git {
		t.Error("git should be detected")
	}
	// opencode detection depends on whether the binary is on PATH in CI;
	// we only verify mcp-server is NOT detected (it shouldn't exist anywhere)
	if tools.MCPBinary {
		t.Error("plan-ai-mcp-server should NOT be detected (not on PATH)")
	}
}

func TestInstaller_DetectToolsOpenCode(t *testing.T) {
	inst, home := testInstaller(t)

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(ocDir, "opencode.json")
	if err := os.WriteFile(cfgPath, []byte(`{"agent_name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	tools := inst.DetectTools()
	if !tools.OpenCode {
		t.Error("opencode should be detected when OPENCODE_CONFIG_DIR has config")
	}
}

// ── sync idempotent ─────────────────────────────────────

func TestInstaller_SyncIsIdempotent(t *testing.T) {
	inst, home := testInstaller(t)

	opts := InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	}

	// First install
	if err := inst.Install(opts); err != nil {
		t.Fatalf("first Install: %v", err)
	}

	// Sync should succeed without error
	if err := inst.Sync(opts); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	// Sync again — still idempotent
	if err := inst.Sync(opts); err != nil {
		t.Fatalf("second Sync: %v", err)
	}

	_ = inst.LoadState()
	if inst.State.Preset != "minimal" {
		t.Fatalf("preset changed to %q after sync", inst.State.Preset)
	}
}

// ── uninstall ───────────────────────────────────────────

func TestInstaller_UninstallRemovesComponents(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "full-plan-ai",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Uninstall specific components
	err = inst.Uninstall([]string{"docs", "context", "alignment"})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}

	_ = inst.LoadState()
	if inst.State.Components["docs"].Installed {
		t.Error("docs should be uninstalled")
	}
	if inst.State.Components["context"].Installed {
		t.Error("context should be uninstalled")
	}
	if inst.State.Components["alignment"].Installed {
		t.Error("alignment should be uninstalled")
	}
	if !inst.State.Components["mcp"].Installed {
		t.Error("mcp should still be installed")
	}
	if !inst.State.Components["intent"].Installed {
		t.Error("intent should still be installed")
	}
}

func TestInstaller_UninstallAll(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "full-plan-ai",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Delete state file to simulate full uninstall
	err = inst.Uninstall(nil) // nil means remove everything
	if err != nil {
		t.Fatalf("Uninstall all: %v", err)
	}

	statePath := filepath.Join(home, ".plan-ai", "state.json")
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatal("state.json should be removed after full uninstall")
	}
}

// ── backup before opencode ──────────────────────────────

func TestInstaller_BackupsOpenCodeConfigBeforeModifying(t *testing.T) {
	inst, home := testInstaller(t)

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(ocDir, "opencode.json")
	origContent := `{"agent_name":"my-app"}`
	if err := os.WriteFile(cfgPath, []byte(origContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	err := inst.Install(InstallOptions{
		Preset:    "minimal",
		BinDir:    filepath.Join(home, "bin"),
		AllowReal: true,
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Check backup exists
	backupsDir := filepath.Join(inst.DataDir, "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("read backups dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "opencode.json") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no backup of opencode config found in backups dir")
	}
}

func TestInstaller_NoBackupOnDryRun(t *testing.T) {
	inst, home := testInstaller(t)

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(ocDir, "opencode.json")
	if err := os.WriteFile(cfgPath, []byte(`{"agent_name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	err := inst.Install(InstallOptions{
		DryRun:    true,
		Preset:    "minimal",
		BinDir:    filepath.Join(home, "bin"),
		AllowReal: true,
	})
	if err != nil {
		t.Fatalf("Install dry-run: %v", err)
	}

	backupsDir := filepath.Join(inst.DataDir, "backups")
	if _, err := os.Stat(backupsDir); !os.IsNotExist(err) {
		t.Fatal("backups dir should not exist after dry-run")
	}
}

// ── doctor report ───────────────────────────────────────

func TestInstaller_DoctorAfterInstall(t *testing.T) {
	inst, home := testInstaller(t)

	// Fake git on PATH
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "go"), []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	// Install minimal
	err := inst.Install(InstallOptions{
		Preset: "minimal",
		BinDir: binDir,
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	report := inst.Doctor()
	if report.StateExists != true {
		t.Error("StateExists should be true after install")
	}
	if report.Tools.Git != true {
		t.Error("git should be detected")
	}
	if report.Tools.Go != true {
		t.Error("go should be detected")
	}
	if report.ComponentsInstalled == 0 {
		t.Error("should have at least one component installed")
	}
}

func TestInstaller_DoctorNoInstall(t *testing.T) {
	inst, _ := testInstaller(t)

	report := inst.Doctor()
	if report.StateExists {
		t.Error("StateExists should be false without install")
	}
}

// ── validate that install doesn't touch unselected configs ──

func TestInstaller_DoesNotTouchOtherMCPConfigs(t *testing.T) {
	inst, home := testInstaller(t)

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(ocDir, "opencode.json")
	origContent := `{"agent_name":"my-app","mcp":{"my-server":{"type":"local","command":["my-srv"]}}}`
	if err := os.WriteFile(cfgPath, []byte(origContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	err := inst.Install(InstallOptions{
		Preset:    "minimal",
		BinDir:    filepath.Join(home, "bin"),
		AllowReal: true,
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Read back and verify other MCP entries survived
	data, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(data), `my-server`) {
		t.Fatal("existing MCP entry 'my-server' was removed from config")
	}
	if !strings.Contains(string(data), `my-app`) {
		t.Fatal("existing agent_name 'my-app' was overwritten")
	}
}

// ── global install vs project init ──────────────────────

func TestInstaller_GlobalInstallCreatesState(t *testing.T) {
	inst, home := testInstaller(t)

	err := inst.Install(InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Global state should exist
	statePath := filepath.Join(home, ".plan-ai", "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("global state.json should exist after install")
	}
}

func TestInstaller_ProjectInitWithoutGlobalInstallFails(t *testing.T) {
	inst, home := testInstaller(t)

	// Try to init a project without global install
	err := inst.InitProject("/some/project", InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	})
	if err == nil {
		t.Fatal("expected error when initing project without global install")
	}
}

func TestInstaller_ProjectInitAfterGlobalInstallSucceeds(t *testing.T) {
	inst, home := testInstaller(t)

	// First install globally
	if err := inst.Install(InstallOptions{
		Preset: "full-plan-ai",
		BinDir: filepath.Join(home, "bin"),
	}); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Then init a project
	projectDir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := inst.InitProject(projectDir, InstallOptions{
		Preset: "minimal",
		BinDir: filepath.Join(home, "bin"),
	})
	if err != nil {
		t.Fatalf("InitProject: %v", err)
	}

	// Project should have .plan-ai/config.json
	projCfg := filepath.Join(projectDir, ".plan-ai", "config.json")
	if _, err := os.Stat(projCfg); os.IsNotExist(err) {
		t.Fatal("project config.json was not created")
	}
}

// ── validate OpenCode config ────────────────────────────

func TestInstaller_ValidatesOpenCodeConfigAfterInstall(t *testing.T) {
	inst, home := testInstaller(t)

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	err := inst.Install(InstallOptions{
		Preset:    "minimal",
		BinDir:    filepath.Join(home, "bin"),
		AllowReal: true,
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Validate that generated config is valid JSON and has required fields
	cfgPath := filepath.Join(ocDir, "opencode.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}

	if !strings.Contains(string(data), `"mcp"`) {
		t.Fatal("opencode config should have mcp section")
	}

	report := inst.Doctor()
	if !report.OpenCodeValid {
		t.Error("OpenCode config should be valid after install")
	}
}

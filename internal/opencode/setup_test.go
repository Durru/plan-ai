package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupService_GeneratesAllArtifacts(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	// All artifact paths must be non-empty.
	assertNotZero(t, result.OpenCodeConfigPath, "OpenCodeConfigPath")
	assertNotZero(t, result.MCPRegistryPath, "MCPRegistryPath")
	assertNotZero(t, result.AgentPath, "AgentPath")
	assertNotZero(t, result.ProfilesPath, "ProfilesPath")
	assertNotZero(t, result.PromptsPath, "PromptsPath")
	assertNotZero(t, result.WorkflowsPath, "WorkflowsPath")
	assertNotZero(t, result.SyncMarkerPath, "SyncMarkerPath")

	// Only the sync marker lives under projectRoot/.plan-ai
	if !strings.Contains(result.SyncMarkerPath, filepath.Join(projectRoot, ".plan-ai")) {
		t.Fatalf("SyncMarkerPath %q should be under projectRoot .plan-ai", result.SyncMarkerPath)
	}

	// All other artifacts live under opencodeDir
	for name, path := range map[string]string{
		"OpenCodeConfigPath": result.OpenCodeConfigPath,
		"MCPRegistryPath":    result.MCPRegistryPath,
		"AgentPath":          result.AgentPath,
		"ProfilesPath":       result.ProfilesPath,
		"PromptsPath":        result.PromptsPath,
		"WorkflowsPath":      result.WorkflowsPath,
	} {
		if !strings.HasPrefix(path, opencodeDir) {
			t.Fatalf("%s %q should be under opencodeDir", name, path)
		}
	}

	// All files must actually exist on disk
	for name, path := range map[string]string{
		"OpenCodeConfigPath": result.OpenCodeConfigPath,
		"MCPRegistryPath":    result.MCPRegistryPath,
		"AgentPath":          result.AgentPath,
		"ProfilesPath":       result.ProfilesPath,
		"PromptsPath":        result.PromptsPath,
		"WorkflowsPath":      result.WorkflowsPath,
		"SyncMarkerPath":     result.SyncMarkerPath,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s file %q does not exist: %v", name, path, err)
		}
	}
}

func TestSetupService_DetectsExistingConfig(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	// Pre-create an opencode.json
	existingPath := filepath.Join(opencodeDir, "opencode.json")
	if err := os.WriteFile(existingPath, []byte(`{"agent_name":"custom-agent","agent_role":"custom-role"}`), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	// Should have detected existing, not overwritten
	if result.OpenCodeConfigPath != existingPath {
		t.Fatalf("expected config path %q, got %q", existingPath, result.OpenCodeConfigPath)
	}
}

func TestSetupService_DetectsExistingJSONC(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	existingPath := filepath.Join(opencodeDir, "opencode.jsonc")
	if err := os.WriteFile(existingPath, []byte(`{"agent_name":"jsonc-agent"}`), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	if result.OpenCodeConfigPath != existingPath {
		t.Fatalf("expected config path %q, got %q", existingPath, result.OpenCodeConfigPath)
	}
}

func TestSetupService_DetectsDotOpenCode(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	dotDir := filepath.Join(opencodeDir, ".opencode")
	if err := os.MkdirAll(dotDir, 0755); err != nil {
		t.Fatal(err)
	}
	existingPath := filepath.Join(dotDir, "opencode.json")
	if err := os.WriteFile(existingPath, []byte(`{"agent_name":"dot-agent"}`), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	if result.OpenCodeConfigPath != existingPath {
		t.Fatalf("expected config path %q, got %q", existingPath, result.OpenCodeConfigPath)
	}
}

func TestSetupService_GeneratedConfigIsValidJSON(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	data, err := os.ReadFile(result.OpenCodeConfigPath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}

	// Verify it's valid JSON with expected fields
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse opencode config: %v", err)
	}
	if cfg["agent_name"] != "plan-ai" {
		t.Fatalf("expected agent_name='plan-ai', got %v", cfg["agent_name"])
	}
	if _, ok := cfg["skills"]; !ok {
		t.Fatal("expected skills in config")
	}
}

func TestSetupService_NoRealPathsTouched(t *testing.T) {
	// Ensure that with sandbox env overrides nothing leaks to real paths.
	realHome := t.TempDir()
	realConfig := filepath.Join(realHome, ".config", "opencode")

	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	_, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	// Real opencode config must NOT exist
	if _, err := os.Stat(realConfig); !os.IsNotExist(err) {
		t.Fatalf("real config path %q should not exist", realConfig)
	}

	// No artifacts should appear under real home
	entries, err := os.ReadDir(realHome)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Fatalf("real home should be empty, got %d entries", len(entries))
	}
}

func TestSetupService_SyncMarkerHasCorrectContent(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	data, err := os.ReadFile(result.SyncMarkerPath)
	if err != nil {
		t.Fatalf("read sync marker: %v", err)
	}

	var marker syncMarker
	if err := json.Unmarshal(data, &marker); err != nil {
		t.Fatalf("parse sync marker: %v", err)
	}

	if marker.Status != "synced" {
		t.Fatalf("expected status 'synced', got %q", marker.Status)
	}
	if marker.OpenCodeConfigDir != opencodeDir {
		t.Fatalf("expected opencodeConfigDir %q, got %q", opencodeDir, marker.OpenCodeConfigDir)
	}
	if marker.ProjectRoot != projectRoot {
		t.Fatalf("expected projectRoot %q, got %q", projectRoot, marker.ProjectRoot)
	}
	if marker.SyncedAt == "" {
		t.Fatal("expected non-empty synced_at timestamp")
	}

	// Verify artifact paths are recorded
	for key, expected := range map[string]string{
		"opencode_config": result.OpenCodeConfigPath,
		"mcp_registry":    result.MCPRegistryPath,
		"agent":           result.AgentPath,
		"profiles":        result.ProfilesPath,
		"prompts":         result.PromptsPath,
	} {
		got, ok := marker.Artifacts[key]
		if !ok {
			t.Fatalf("sync marker missing artifact key %q", key)
		}
		if got != expected {
			t.Fatalf("sync marker artifact %q: expected %q, got %q", key, expected, got)
		}
	}
}

func TestSetupService_GeneratedMCPRegistryIsValid(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	data, err := os.ReadFile(result.MCPRegistryPath)
	if err != nil {
		t.Fatalf("read mcp registry: %v", err)
	}

	var items []setupMCPItem
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("parse mcp registry: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one MCP item")
	}

	found := false
	for _, item := range items {
		if item.Name == "plan_ai.status" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'plan_ai.status' in MCP registry")
	}
}

func TestSetupService_GeneratedAgentIsValid(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	data, err := os.ReadFile(result.AgentPath)
	if err != nil {
		t.Fatalf("read agent: %v", err)
	}

	var agent setupAgent
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("parse agent: %v", err)
	}
	if agent.Name != "plan-ai" {
		t.Fatalf("expected agent name 'plan-ai', got %q", agent.Name)
	}
	if len(agent.Skills) == 0 {
		t.Fatal("expected non-empty skills list")
	}
}

func TestSetupService_EmptyPathsReturnError(t *testing.T) {
	svc := NewSetupService()

	_, err := svc.Run("", "/tmp/project")
	if err == nil {
		t.Fatal("expected error for empty opencodeDir")
	}

	_, err = svc.Run("/tmp/opencode", "")
	if err == nil {
		t.Fatal("expected error for empty projectRoot")
	}
}

func TestDetectOpenCodeConfig_NotFound(t *testing.T) {
	tmp := t.TempDir()
	result := DetectOpenCodeConfig(tmp)
	if result.Found {
		t.Fatal("expected Found=false for empty dir")
	}
}

func TestDetectOpenCodeConfig_Found(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "opencode.json")
	if err := os.WriteFile(cfg, []byte(`{"agent_name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	result := DetectOpenCodeConfig(tmp)
	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.ConfigPath != cfg {
		t.Fatalf("expected ConfigPath %q, got %q", cfg, result.ConfigPath)
	}
}

func TestSetupService_AllArtifactsUnderSandboxPaths(t *testing.T) {
	opencodeDir := t.TempDir()
	projectRoot := t.TempDir()

	svc := NewSetupService()
	result, err := svc.Run(opencodeDir, projectRoot)
	if err != nil {
		t.Fatalf("SetupService.Run: %v", err)
	}

	// The opencode config dir should be under opencodeDir
	if !strings.HasPrefix(result.OpenCodeConfigPath, opencodeDir) {
		t.Fatalf("opencode config %q not under opencodeDir %q", result.OpenCodeConfigPath, opencodeDir)
	}

	// The sync marker should be under projectRoot/.plan-ai
	expectedSyncPrefix := filepath.Join(projectRoot, ".plan-ai")
	if !strings.HasPrefix(result.SyncMarkerPath, expectedSyncPrefix) {
		t.Fatalf("sync marker %q not under %q", result.SyncMarkerPath, expectedSyncPrefix)
	}

	// Verify generated JSON files are valid by reading them back
	for name, path := range map[string]string{
		"opencode_config": result.OpenCodeConfigPath,
		"mcp_registry":    result.MCPRegistryPath,
		"agent":           result.AgentPath,
		"profiles":        result.ProfilesPath,
		"prompts":         result.PromptsPath,
		"sync_marker":     result.SyncMarkerPath,
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("%s file is empty", name)
		}
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatalf("%s is not valid JSON: %v", name, err)
		}
	}
}

func assertNotZero(t *testing.T, val, name string) {
	t.Helper()
	if val == "" {
		t.Fatalf("%s must not be empty", name)
	}
}

package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupMCPConfig_BackupWhenConfigExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPENCODE_CONFIG_DIR", dir)

	original := map[string]any{
		"theme": "dark",
		"mcp": map[string]any{
			"other-tool": map[string]any{"command": "other", "type": "local"},
		},
	}
	originalData, _ := json.MarshalIndent(original, "", "  ")
	configPath := filepath.Join(dir, "opencode.json")
	if err := os.WriteFile(configPath, originalData, 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	backupPath, err := SetupMCPConfig(dir, "", false) // false: test uses sandbox OPENCODE_CONFIG_DIR
	if err != nil {
		t.Fatalf("SetupMCPConfig: %v", err)
	}

	if backupPath == "" {
		t.Fatal("expected non-empty backup path when prior config exists")
	}
	if !strings.Contains(backupPath, "pre-mcp-write") {
		t.Errorf("backup path %q missing reason tag", backupPath)
	}
	if !strings.HasSuffix(backupPath, ".bak") {
		t.Errorf("backup path %q missing .bak suffix", backupPath)
	}

	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !strings.Contains(string(backup), "other-tool") {
		t.Errorf("backup missing user data: %s", backup)
	}
	if !strings.Contains(string(backup), `"theme": "dark"`) {
		t.Errorf("backup missing theme: %s", backup)
	}

	target, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !strings.Contains(string(target), "plan-ai") {
		t.Errorf("target missing plan-ai: %s", target)
	}
}

func TestSetupMCPConfig_NoBackupWhenConfigAbsent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPENCODE_CONFIG_DIR", dir)

	backupPath, err := SetupMCPConfig(dir, "", false) // false: test uses sandbox OPENCODE_CONFIG_DIR
	if err != nil {
		t.Fatalf("SetupMCPConfig: %v", err)
	}
	if backupPath != "" {
		t.Errorf("expected empty backup path when no prior config, got %q", backupPath)
	}

	configPath := filepath.Join(dir, "opencode.json")
	target, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !strings.Contains(string(target), "plan-ai") {
		t.Errorf("target missing plan-ai: %s", target)
	}
}

func TestSetupMCPConfig_AtomicNoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPENCODE_CONFIG_DIR", dir)

	if _, err := SetupMCPConfig(dir, "", false); err != nil {
		t.Fatalf("SetupMCPConfig: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorFixRepairsStaleState(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	ocConfigPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(ocConfigPath); err == nil {
		if err := os.Remove(ocConfigPath); err != nil {
			t.Fatalf("remove oc config: %v", err)
		}
	}

	out, err = executeCommand(t, home, t.TempDir(), "doctor", "--fix")
	if err != nil {
		t.Fatalf("doctor --fix: %v\n%s", err, out)
	}
	if !strings.Contains(out, "--- Fix ---") {
		t.Errorf("doctor --fix output missing '--- Fix ---' section:\n%s", out)
	}

	if _, err := os.Stat(ocConfigPath); err != nil {
		t.Fatalf("opencode config not restored: %v", err)
	}
	data, err := os.ReadFile(ocConfigPath)
	if err != nil {
		t.Fatalf("read oc config: %v", err)
	}
	if !strings.Contains(string(data), "plan-ai") {
		t.Errorf("opencode config missing plan-ai after --fix:\n%s", data)
	}
}

func TestDoctorFixRepairsRegisteredBinaryMissing(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	// Doctor with --fix should find the registered_binary_missing issue
	// and re-run sync (the sync itself may succeed even without the binary
	// — it's the doctor that reports the issue, not sync).
	out, err = executeCommand(t, home, t.TempDir(), "doctor", "--fix")
	if err != nil {
		t.Fatalf("doctor --fix: %v\n%s", err, out)
	}

	if !strings.Contains(out, "--- Fix ---") {
		t.Errorf("doctor --fix output missing '--- Fix ---' section:\n%s", out)
	}
}

func TestSetupMCPConfigIsAtomic(t *testing.T) {
	home := t.TempDir()

	ocDir := filepath.Join(home, ".config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatalf("mkdir ocDir: %v", err)
	}
	t.Setenv("OPENCODE_CONFIG_DIR", ocDir)

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir binDir: %v", err)
	}

	out, err := executeCommand(t, home, t.TempDir(), "install", "--preset=minimal", "--bin-dir="+binDir)
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	configPath := filepath.Join(ocDir, "opencode.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("opencode.json not written: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Errorf("config.json not valid JSON: %v\n%s", err, data)
	}
	if _, ok := cfg["mcp"]; !ok {
		t.Errorf("opencode.json missing mcp: %s", data)
	}

	entries, err := os.ReadDir(ocDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

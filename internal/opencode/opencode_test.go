package opencode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetector_NoConfig(t *testing.T) {
	d := NewDetector()
	// Use a temp dir with no opencode config
	tmp := t.TempDir()
	result := d.Detect(tmp)
	if result.Found {
		t.Fatal("expected no opencode config in empty temp dir")
	}
}

func TestDetector_FindJSON(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "opencode.json")
	os.WriteFile(cfg, []byte(`{"agent_name": "test-agent", "agent_role": "tester"}`), 0644)

	d := NewDetector()
	result := d.Detect(tmp)
	if !result.Found {
		t.Fatal("expected to find opencode.json")
	}
	if result.ConfigPath != cfg {
		t.Fatalf("expected config path %q, got %q", cfg, result.ConfigPath)
	}
	if result.AgentName != "test-agent" {
		t.Fatalf("expected agent name 'test-agent', got %q", result.AgentName)
	}
	if result.AgentRole != "tester" {
		t.Fatalf("expected agent role 'tester', got %q", result.AgentRole)
	}
}

func TestDetector_FindJSONC(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "opencode.jsonc")
	os.WriteFile(cfg, []byte(`{"agent_name": "jsonc-agent"}`), 0644)

	d := NewDetector()
	result := d.Detect(tmp)
	if !result.Found {
		t.Fatal("expected to find opencode.jsonc")
	}
	if result.AgentName != "jsonc-agent" {
		t.Fatalf("expected 'jsonc-agent', got %q", result.AgentName)
	}
}

func TestDetector_FindDotOpenCode(t *testing.T) {
	tmp := t.TempDir()
	dotDir := filepath.Join(tmp, ".opencode")
	os.MkdirAll(dotDir, 0755)
	cfg := filepath.Join(dotDir, "opencode.json")
	os.WriteFile(cfg, []byte(`{"agent_name": "dot-agent", "agent_role": "dot-role", "skills": ["go", "ts"]}`), 0644)

	d := NewDetector()
	result := d.Detect(tmp)
	if !result.Found {
		t.Fatal("expected to find .opencode/opencode.json")
	}
	if result.AgentName != "dot-agent" {
		t.Fatalf("expected 'dot-agent', got %q", result.AgentName)
	}
	if !result.HasSkills {
		t.Fatal("expected HasSkills to be true")
	}
	if result.SkillCount != 2 {
		t.Fatalf("expected 2 skills, got %d", result.SkillCount)
	}
}

func TestDetector_WithSelfInit(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "opencode.json")
	os.WriteFile(cfg, []byte(`{"agent_name": "self-init-agent", "agent": {"mode": "full"}}`), 0644)

	d := NewDetector()
	result := d.Detect(tmp)
	if !result.HasSelfInit {
		t.Fatal("expected HasSelfInit to be true")
	}
}

func TestDetector_ParentDir(t *testing.T) {
	// Config in parent dir should be detected from subdir
	parent := t.TempDir()
	cfg := filepath.Join(parent, "opencode.json")
	os.WriteFile(cfg, []byte(`{"agent_name": "parent-agent"}`), 0644)

	child := filepath.Join(parent, "subdir")
	os.MkdirAll(child, 0755)

	d := NewDetector()
	result := d.Detect(child)
	if !result.Found {
		t.Fatal("expected to detect parent dir config")
	}
	if result.AgentName != "parent-agent" {
		t.Fatalf("expected 'parent-agent', got %q", result.AgentName)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Enabled {
		t.Fatal("expected integration to be enabled by default")
	}
	if cfg.Mode != ModeTool {
		t.Fatalf("expected ModeTool, got %q", cfg.Mode)
	}
	if !cfg.ReadOnly {
		t.Fatal("expected read-only by default")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	cfg := DefaultConfig()
	cfg.Mode = ModeHybrid
	cfg.AutoDetect = false

	err := SaveConfig(tmp, cfg)
	if err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.Mode != ModeHybrid {
		t.Fatalf("expected ModeHybrid, got %q", loaded.Mode)
	}
	if loaded.AutoDetect != false {
		t.Fatal("expected AutoDetect to be false")
	}
}

func TestConfig_LoadNonExistent(t *testing.T) {
	tmp := t.TempDir()
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error loading non-existent config: %v", err)
	}
	if !cfg.Enabled {
		t.Fatal("expected defaults to have integration enabled")
	}
}

func TestToolRegistry(t *testing.T) {
	tmp := t.TempDir()

	r1 := NewToolRegistry(tmp)
	r1.Add(RegistryItem{Name: "tool_a", Description: "Tool A"})
	r1.Add(RegistryItem{Name: "tool_b", Description: "Tool B", PlansBuilt: 5})
	err := r1.Save()
	if err != nil {
		t.Fatalf("save registry: %v", err)
	}

	r2 := NewToolRegistry(tmp)
	err = r2.Load()
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}

	items := r2.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "tool_a" {
		t.Fatalf("expected 'tool_a', got %q", items[0].Name)
	}
}

func TestToolRegistry_Increment(t *testing.T) {
	tmp := t.TempDir()

	r := NewToolRegistry(tmp)
	r.Add(RegistryItem{Name: "counter", PlansBuilt: 1})
	r.IncrementPlanCount("counter")

	// Check via save/load
	r.Save()
	r2 := NewToolRegistry(tmp)
	r2.Load()
	items := r2.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PlansBuilt != 2 {
		t.Fatalf("expected PlansBuilt=2, got %d", items[0].PlansBuilt)
	}
}

func TestNewDoctor(t *testing.T) {
	d := NewDoctor()
	if d == nil {
		t.Fatal("expected non-nil doctor")
	}
}

func TestDoctor_VersionCheck(t *testing.T) {
	d := NewDoctor()
	detected := &DetectionResult{Found: true}
	cfg := DefaultConfig()

	results := d.RunChecks(detected, cfg)
	if len(results) == 0 {
		t.Fatal("expected at least one check result")
	}

	// At least one should pass
	hasPass := false
	for _, r := range results {
		if r.Status == "pass" {
			hasPass = true
			break
		}
	}
	if !hasPass {
		t.Fatal("expected at least one passing check")
	}
}

func TestDoctor_NoOpenCode(t *testing.T) {
	d := NewDoctor()
	detected := &DetectionResult{Found: false}
	cfg := DefaultConfig()

	results := d.RunChecks(detected, cfg)
	hasWarn := false
	for _, r := range results {
		if r.Status == "warn" {
			hasWarn = true
			break
		}
	}
	if !hasWarn {
		t.Fatal("expected at least one warning when no opencode config")
	}
}

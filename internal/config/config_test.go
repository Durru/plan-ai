package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGlobalDirectoryHelpers(t *testing.T) {
	base := filepath.Join("/home", "tester")
	wantRoot := filepath.Join(base, ".plan-ai")

	if got := GlobalDir(base); got != wantRoot {
		t.Fatalf("GlobalDir() = %q, want %q", got, wantRoot)
	}

	cases := map[string]string{
		"config":   GlobalConfigPath(base),
		"database": GlobalDBPath(base),
		"cache":    GlobalCacheDir(base),
		"logs":     GlobalLogsDir(base),
		"skills":   GlobalSkillsDir(base),
		"data":     GlobalDataDir(base),
		"backups":  GlobalBackupsDir(base),
	}

	expected := map[string]string{
		"config":   filepath.Join(wantRoot, "config.json"),
		"database": filepath.Join(wantRoot, "global.db"),
		"cache":    filepath.Join(wantRoot, "cache"),
		"logs":     filepath.Join(wantRoot, "logs"),
		"skills":   filepath.Join(wantRoot, "skills"),
		"data":     filepath.Join(wantRoot, "data"),
		"backups":  filepath.Join(wantRoot, "backups"),
	}

	for name, got := range cases {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestProjectDirectoryHelpers(t *testing.T) {
	project := filepath.Join("/workspace", "demo")
	wantRoot := filepath.Join(project, ".plan-ai")

	if got := ProjectDir(project); got != wantRoot {
		t.Fatalf("ProjectDir() = %q, want %q", got, wantRoot)
	}

	cases := map[string]string{
		"config":    ProjectConfigPath(project),
		"database":  ProjectDBPath(project),
		"cache":     ProjectCacheDir(project),
		"snapshots": ProjectSnapshotsDir(project),
		"exports":   ProjectExportsDir(project),
		"docs":      ProjectDocsDir(project),
		"locks":     ProjectLocksDir(project),
		"backups":   ProjectBackupsDir(project),
	}

	expected := map[string]string{
		"config":    filepath.Join(wantRoot, "config.json"),
		"database":  filepath.Join(wantRoot, "project.db"),
		"cache":     filepath.Join(wantRoot, "cache"),
		"snapshots": filepath.Join(wantRoot, "snapshots"),
		"exports":   filepath.Join(wantRoot, "exports"),
		"docs":      filepath.Join(wantRoot, "docs"),
		"locks":     filepath.Join(wantRoot, "locks"),
		"backups":   filepath.Join(wantRoot, "backups"),
	}

	for name, got := range cases {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestSaveAndLoadConfigs(t *testing.T) {
	root := t.TempDir()

	global := GlobalConfig{
		Version:      "0.1.0-dev",
		InstalledAt:  "2026-06-02T00:00:00Z",
		GlobalDir:    filepath.Join(root, ".plan-ai"),
		GlobalDB:     filepath.Join(root, ".plan-ai", "global.db"),
		Integrations: map[string]any{"test": true},
	}
	if err := SaveGlobalConfig(filepath.Join(root, "global.json"), global); err != nil {
		t.Fatalf("save global config: %v", err)
	}
	loadedGlobal, err := LoadGlobalConfig(filepath.Join(root, "global.json"))
	if err != nil {
		t.Fatalf("load global config: %v", err)
	}
	if !reflect.DeepEqual(loadedGlobal, global) {
		t.Fatalf("loaded global config = %#v, want %#v", loadedGlobal, global)
	}

	project := ProjectConfig{
		Version:      "0.1.0-dev",
		ProjectName:  "demo",
		ProjectRoot:  "/workspace/demo",
		ProjectDB:    "/workspace/demo/.plan-ai/project.db",
		CreatedAt:    "2026-06-02T00:00:00Z",
		Integrations: map[string]any{"test": "value"},
	}
	if err := SaveProjectConfig(filepath.Join(root, "project", "config.json"), project); err != nil {
		t.Fatalf("save project config: %v", err)
	}
	loadedProject, err := LoadProjectConfig(filepath.Join(root, "project", "config.json"))
	if err != nil {
		t.Fatalf("load project config: %v", err)
	}
	if !reflect.DeepEqual(loadedProject, project) {
		t.Fatalf("loaded project config = %#v, want %#v", loadedProject, project)
	}
}

func TestSaveConfigsWritesReadableJSON(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	cfg := GlobalConfig{Version: "0.1.0-dev", Integrations: map[string]any{}}

	if err := SaveGlobalConfig(path, cfg); err != nil {
		t.Fatalf("save global config: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("config is not valid JSON: %s", data)
	}
	if got, want := string(data[len(data)-1]), "\n"; got != want {
		t.Fatalf("config should end with newline, got %q", got)
	}
	if !strings.Contains(string(data), "\n  \"version\"") {
		t.Fatalf("config should be indented for readability: %s", data)
	}
}

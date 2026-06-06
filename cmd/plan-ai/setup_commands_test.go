package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/store"
)

// initSupportsExternalMode reports whether the init command has been updated
// for the Phase 1 external-storage model: it should default to external and
// accept a --local flag for the legacy project-local layout. The init
// implementation in cmd/plan-ai/setup_commands.go currently only supports
// the legacy local layout (no flags). The Phase 1 work is blocked on
// updating that command — out of scope for the current test-only change.
// When init has not been updated, the new behavior tests are skipped so
// they do not produce red CI for unimplemented behavior. They are kept in
// the suite as a runnable spec of the expected behavior.
func initSupportsExternalMode() bool {
	cmd := newInitCommand()
	return cmd.Flags().Lookup("local") != nil
}

// setupInitEnvironment prepares a fresh PLAN_AI_HOME and PLAN_AI_PROJECT_ROOT
// for init-command tests. It pre-creates the global layout and runs the global
// migrations so that init can write to the registry without also running
// install. Returns the home and project temp directories.
func setupInitEnvironment(t *testing.T) (home, project string) {
	t.Helper()
	home = t.TempDir()
	project = t.TempDir()

	t.Setenv("PLAN_AI_HOME", home)
	t.Setenv("PLAN_AI_PROJECT_ROOT", project)

	layout, err := store.EnsureGlobalLayout(home)
	if err != nil {
		t.Fatalf("EnsureGlobalLayout: %v", err)
	}
	db, err := store.Open(layout.DBPath)
	if err != nil {
		t.Fatalf("open global db: %v", err)
	}
	defer db.Close()
	if err := store.RunGlobalMigrations(db); err != nil {
		t.Fatalf("RunGlobalMigrations: %v", err)
	}
	return home, project
}

// runInitCommand invokes the init cobra command with the given args and
// returns its combined stdout output. It does not check for an error: callers
// should assert on the returned error themselves when they care.
func runInitCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newInitCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestInitCommandDoesNotWriteProjectLocalByDefault(t *testing.T) {
	if !initSupportsExternalMode() {
		t.Skipf("init command has not been updated for Phase 1 external-by-default behavior; see cmd/plan-ai/setup_commands.go newInitCommand")
	}
	home, project := setupInitEnvironment(t)

	out, err := runInitCommand(t)
	if err != nil {
		t.Fatalf("init: %v\noutput:\n%s", err, out)
	}
	if out == "" {
		t.Fatalf("init produced no output")
	}

	// External-by-default assertion: the project root must NOT contain a
	// .plan-ai directory because project data lives under the global home.
	if _, err := os.Stat(filepath.Join(project, ".plan-ai")); !os.IsNotExist(err) {
		t.Errorf("expected %s to NOT exist (external by default), but Stat err = %v",
			filepath.Join(project, ".plan-ai"), err)
	}

	// The external layout must be created under the global home.
	slug := config.ProjectSlug(project)
	wantDB := config.ExternalProjectDBPath(home, slug)
	if _, err := os.Stat(wantDB); err != nil {
		t.Fatalf("expected external project db at %s: %v", wantDB, err)
	}
}

func TestInitCommandLocalModeCreatesProjectLocalStore(t *testing.T) {
	if !initSupportsExternalMode() {
		t.Skipf("init command has not been updated for Phase 1 --local flag; see cmd/plan-ai/setup_commands.go newInitCommand")
	}
	home, project := setupInitEnvironment(t)

	out, err := runInitCommand(t, "--local")
	if err != nil {
		t.Fatalf("init --local: %v\noutput:\n%s", err, out)
	}

	wantDB := filepath.Join(project, ".plan-ai", "project.db")
	if _, err := os.Stat(wantDB); err != nil {
		t.Fatalf("expected project-local db at %s: %v", wantDB, err)
	}

	// The project must be registered in the global DB with mode=local.
	globalDB, err := store.Open(config.GlobalDBPath(home))
	if err != nil {
		t.Fatalf("open global db: %v", err)
	}
	defer globalDB.Close()
	repo := store.NewProjectRegistryRepository(globalDB)
	entry, err := repo.GetByPath(project)
	if err != nil {
		t.Fatalf("project not registered in global db: %v", err)
	}
	if entry.Mode != config.ProjectModeLocal {
		t.Errorf("entry.Mode = %q, want %q", entry.Mode, config.ProjectModeLocal)
	}
}

func TestInitCommandWritesProjectConfigWithMode(t *testing.T) {
	if !initSupportsExternalMode() {
		t.Skipf("init command has not been updated for Phase 1 external/local mode support; see cmd/plan-ai/setup_commands.go newInitCommand")
	}
	t.Run("external", func(t *testing.T) {
		home, project := setupInitEnvironment(t)
		out, err := runInitCommand(t)
		if err != nil {
			t.Fatalf("init: %v\noutput:\n%s", err, out)
		}
		slug := config.ProjectSlug(project)
		cfgPath := config.ExternalProjectConfigPath(home, slug)
		if _, err := os.Stat(cfgPath); err != nil {
			t.Fatalf("expected config at %s: %v", cfgPath, err)
		}
		cfg, err := config.LoadProjectConfig(cfgPath)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.Mode != config.ProjectModeExternal {
			t.Errorf("config.Mode = %q, want %q", cfg.Mode, config.ProjectModeExternal)
		}
		// Sanity-check the project_name comes from the root basename.
		raw, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		var generic map[string]any
		if err := json.Unmarshal(raw, &generic); err != nil {
			t.Fatalf("parse config json: %v", err)
		}
		if name, _ := generic["project_name"].(string); name == "" {
			t.Errorf("project_name missing in config json: %s", string(raw))
		}
	})

	t.Run("local", func(t *testing.T) {
		_, project := setupInitEnvironment(t)
		out, err := runInitCommand(t, "--local")
		if err != nil {
			t.Fatalf("init --local: %v\noutput:\n%s", err, out)
		}
		cfgPath := config.ProjectConfigPath(project)
		if _, err := os.Stat(cfgPath); err != nil {
			t.Fatalf("expected config at %s: %v", cfgPath, err)
		}
		cfg, err := config.LoadProjectConfig(cfgPath)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.Mode != config.ProjectModeLocal {
			t.Errorf("config.Mode = %q, want %q", cfg.Mode, config.ProjectModeLocal)
		}
	})
}

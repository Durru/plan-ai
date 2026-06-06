package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/installer"
	"github.com/plan-ai/plan-ai/internal/opencode"
	"github.com/plan-ai/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

func newBootstrapCommand() *cobra.Command {
	var allowReal bool
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Install global storage, initialize the current project, and wire OpenCode/MCP artifacts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}

			globalLayout, err := store.EnsureGlobalLayout(home)
			if err != nil {
				return err
			}
			if _, err := os.Stat(globalLayout.ConfigPath); os.IsNotExist(err) {
				cfg := config.GlobalConfig{Version: configVersion, InstalledAt: nowUTC(), GlobalDir: globalLayout.Dir, GlobalDB: globalLayout.DBPath, Integrations: map[string]any{}}
				if err := config.SaveGlobalConfig(globalLayout.ConfigPath, cfg); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			globalDB, err := store.Open(globalLayout.DBPath)
			if err != nil {
				return err
			}
			defer globalDB.Close()
			if err := store.RunGlobalMigrations(globalDB); err != nil {
				return err
			}
			fmt.Fprintln(out, "Global installation: installed")
			fmt.Fprintf(out, "Global dir: %s\n", globalLayout.Dir)
			fmt.Fprintf(out, "Global db: %s\n", globalLayout.DBPath)

			projectLayout, err := store.EnsureProjectLayout(projectRoot)
			if err != nil {
				return err
			}
			projectName := filepath.Base(projectRoot)
			if _, err := os.Stat(projectLayout.ConfigPath); os.IsNotExist(err) {
				cfg := config.ProjectConfig{Version: configVersion, ProjectName: projectName, ProjectRoot: projectRoot, ProjectDB: projectLayout.DBPath, CreatedAt: nowUTC(), Integrations: map[string]any{}}
				if err := config.SaveProjectConfig(projectLayout.ConfigPath, cfg); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			projectDB, err := store.Open(projectLayout.DBPath)
			if err != nil {
				return err
			}
			defer projectDB.Close()
			if err := store.RunProjectMigrations(projectDB); err != nil {
				return err
			}
			if err := store.UpsertProjectState(projectDB, store.ProjectID(projectRoot), projectName, projectRoot, "initialized"); err != nil {
				return err
			}
			if err := store.UpsertKnownProject(globalDB, store.ProjectID(projectRoot), projectName, projectRoot); err != nil {
				return err
			}
			fmt.Fprintln(out, "Project initialization: initialized")
			fmt.Fprintf(out, "Project dir: %s\n", projectLayout.Dir)
			fmt.Fprintf(out, "Project db: %s\n", projectLayout.DBPath)

			opencodeDir, err := resolveOpenCodeConfigDirForWrite(allowReal)
			if err != nil {
				fmt.Fprintf(out, "OpenCode integration: skipped (%v)\n", err)
				fmt.Fprintln(out, "Run `plan-ai setup opencode --allow-real-opencode` to write real OpenCode config.")
				return nil
			}
			result, err := opencode.NewSetupService().Run(opencodeDir, projectRoot)
			if err != nil {
				return fmt.Errorf("setup opencode: %w", err)
			}
			fmt.Fprintln(out, "OpenCode integration artifacts generated.")
			fmt.Fprintf(out, "  opencode config: %s\n", result.OpenCodeConfigPath)
			fmt.Fprintf(out, "  mcp registry:    %s\n", result.MCPRegistryPath)
			fmt.Fprintf(out, "  agent:           %s\n", result.AgentPath)

			// Write unified MCP config
			backupPath, err := opencode.SetupMCPConfig(home, "", allowReal)
			if err != nil {
				return fmt.Errorf("setup opencode mcp config: %w", err)
			}
			if backupPath != "" {
				fmt.Fprintf(out, "  previous config backed up to: %s\n", backupPath)
			}
			fmt.Fprintln(out, "Unified OpenCode MCP config written.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "allow writing to real ~/.config/opencode when OPENCODE_CONFIG_DIR is not set")
	return cmd
}

func newInstallCommand() *cobra.Command {
	var (
		dryRun     bool
		preset     string
		components []string
		allowReal  bool
		binDir     string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Plan-AI globally.",
		Long: `Install Plan-AI globally with component selection and presets.

The installer is the single, canonical install path. It detects tools,
records state under $HOME/.plan-ai/state.json, runs global store
migrations, and configures OpenCode via the centralized authority
(SetupMCPConfig), honoring $OPENCODE_CONFIG_DIR when set.

Presets:
  full-plan-ai     All components (default)
  ecosystem-only   MCP server, OpenCode agent, and documentation
  minimal          MCP server only
  custom           Select individual components with --component

Components:
  intent           Product Intent, discovery, and ambiguity analysis
  planning         Master plans, specific plans, tasks, and phases
  mcp              MCP server binary, protocol, tools, and registry
  opencode-agent   OpenCode agent registration, profiles, prompts
  docs             Installation docs, quickstart, integration guides
  context          L0-L4 context generation and approved context
  alignment        Alignment checks, validation, and confidence scoring`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}

			if binDir == "" {
				binDir = filepath.Join(home, ".local", "bin")
			}

			inst := installer.NewInstaller(home)

			// Validate preset/components
			if preset != "" && preset != "custom" {
				if _, ok := installer.Presets[preset]; !ok {
					return fmt.Errorf("unknown preset %q — valid: full-plan-ai, ecosystem-only, minimal, custom", preset)
				}
			}
			if preset == "custom" && len(components) == 0 {
				return fmt.Errorf("preset 'custom' requires at least one --component flag")
			}
			for _, c := range components {
				valid := false
				for _, ac := range installer.AllComponents {
					if c == ac {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("unknown component %q — valid: %s", c, strings.Join(installer.AllComponents, ", "))
				}
			}
			if preset == "" {
				preset = "full-plan-ai"
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "Dry-run mode — no changes will be made")
				fmt.Fprintf(cmd.OutOrStdout(), "Preset: %s\n", preset)
				if preset == "custom" {
					fmt.Fprintf(cmd.OutOrStdout(), "Components: %s\n", strings.Join(components, ", "))
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Bin dir: %s\n", binDir)
				fmt.Fprintf(cmd.OutOrStdout(), "State dir: %s\n", filepath.Join(home, ".plan-ai"))

				comps := installer.Presets[preset]
				if preset == "custom" {
					comps = components
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Components to install:")
				for _, c := range comps {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s: %s\n", c, installer.ComponentDescriptions[c])
				}

				tools := inst.DetectTools()
				fmt.Fprintln(cmd.OutOrStdout(), "\nTools detected:")
				fmt.Fprintf(cmd.OutOrStdout(), "  opencode: %v\n", tools.OpenCode)
				fmt.Fprintf(cmd.OutOrStdout(), "  git:      %v\n", tools.Git)
				fmt.Fprintf(cmd.OutOrStdout(), "  go:       %v\n", tools.Go)
				fmt.Fprintf(cmd.OutOrStdout(), "  mcp-srv:  %v\n", tools.MCPBinary)
				return nil
			}

			if err := inst.Install(installer.InstallOptions{
				Preset:     preset,
				Components: components,
				BinDir:     binDir,
				AllowReal:  allowReal,
			}); err != nil {
				return fmt.Errorf("install: %w", err)
			}

			// Run global store migrations so the registry / known_projects
			// schema (added in migration 0008) exists immediately after
			// install. Idempotent — RunGlobalMigrations is a no-op on a
			// fully-migrated DB.
			layout, err := store.EnsureGlobalLayout(home)
			if err != nil {
				return err
			}
			if db, err := store.Open(layout.DBPath); err == nil {
				if mErr := store.RunGlobalMigrations(db); mErr != nil {
					db.Close()
					return fmt.Errorf("run global migrations: %w", mErr)
				}
				db.Close()
			}
			// Write the global config file. The install command is the
			// canonical place where the global layout is provisioned, so
			// it owns writing the initial config.json.
			if err := writeGlobalConfigIfMissing(layout.ConfigPath, preset, binDir); err != nil {
				return fmt.Errorf("write global config: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Global installation: installed")
			fmt.Fprintf(cmd.OutOrStdout(), "Global dir: %s\n", layout.Dir)
			fmt.Fprintf(cmd.OutOrStdout(), "Global db: %s\n", layout.DBPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Preset: %s\n", preset)
			fmt.Fprintf(cmd.OutOrStdout(), "Bin dir: %s\n", binDir)

			_ = inst.LoadState()
			installed := make([]string, 0)
			for name, cs := range inst.State.Components {
				if cs.Installed {
					installed = append(installed, name)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\n", strings.Join(installed, ", "))

			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without making changes")
	cmd.Flags().StringVar(&preset, "preset", "", "Install preset: full-plan-ai, ecosystem-only, minimal, custom")
	cmd.Flags().StringSliceVar(&components, "component", nil, "Component(s) to install (repeatable, requires --preset=custom)")
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "Allow writing to real ~/.config/opencode")
	cmd.Flags().StringVar(&binDir, "bin-dir", "", "Binary install directory (default: $HOME/.local/bin)")
	return cmd
}


func newInitCommand() *cobra.Command {
	var (
		opencodeArtifacts bool
		localMode         bool
		allowReal         bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Plan-AI persistence for the current project.",
		Long: `Initialize Plan-AI persistence for the current project.

By default, project data is stored externally under the global Plan-AI
home (~/.plan-ai/projects/<slug>/). Pass --local to keep the legacy
project-local store at <root>/.plan-ai/ instead.

OpenCode integration is safe by default: Plan-AI never writes to real
~/.config/opencode/ unless you pass --allow-real-opencode. For sandbox
use, set OPENCODE_CONFIG_DIR instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}

			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}

			mode := config.ProjectModeExternal
			if localMode {
				mode = config.ProjectModeLocal
			}

			projectName := filepath.Base(projectRoot)
			projectID := store.ProjectID(projectRoot)
			slug := config.ProjectSlug(projectRoot)

			var (
				dbPath     string
				configPath string
				dir        string
			)

			if mode == config.ProjectModeLocal {
				layout, err := store.EnsureProjectLayout(projectRoot)
				if err != nil {
					return err
				}
				dbPath = layout.DBPath
				configPath = layout.ConfigPath
				dir = layout.Dir
			} else {
				extLayout, err := store.EnsureExternalProjectLayout(home, slug)
				if err != nil {
					return err
				}
				dbPath = extLayout.DBPath
				configPath = extLayout.ConfigPath
				dir = extLayout.Dir
			}

			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				cfg := config.ProjectConfig{
					Version:      configVersion,
					ProjectName:  projectName,
					ProjectRoot:  projectRoot,
					ProjectDB:    dbPath,
					Mode:         mode,
					CreatedAt:    nowUTC(),
					Integrations: map[string]any{},
				}
				if err := config.SaveProjectConfig(configPath, cfg); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := store.RunProjectMigrations(db); err != nil {
				return err
			}
			if err := store.UpsertProjectState(db, projectID, projectName, projectRoot, "initialized"); err != nil {
				return err
			}

			globalStatus := "missing"
			if globalDB, err := openInstalledGlobalStore(); err == nil {
				defer globalDB.Close()
				entry := store.ProjectRegistryEntry{
					ID:       projectID,
					Name:     projectName,
					RootPath: projectRoot,
					Slug:     slug,
					Mode:     mode,
				}
				if _, err := store.NewProjectRegistryRepository(globalDB).Register(entry); err != nil {
					return err
				}
				globalStatus = "installed"
			}

			// Setup OpenCode MCP config — only if explicitly allowed or
			// OPENCODE_CONFIG_DIR is set (sandbox). Without --allow-real-opencode
			// we skip this step to avoid modifying the real OpenCode config.
			if _, err := resolveOpenCodeConfigDirForWrite(allowReal); err == nil {
				backupPath, err := opencode.SetupMCPConfig(home, "", allowReal)
				if err != nil {
					return fmt.Errorf("setup opencode mcp: %w", err)
				}
				_ = backupPath
			}

			// Generate foundation artifacts when --opencode is set
			if opencodeArtifacts {
				if _, err := resolveOpenCodeConfigDirForWrite(allowReal); err != nil {
					return fmt.Errorf("cannot generate opencode artifacts: %w (pass --allow-real-opencode to write to real config)", err)
				}
				if err := opencode.GenerateProjectArtifacts(home); err != nil {
					return fmt.Errorf("generate opencode artifacts: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "OpenCode foundation artifacts generated.")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Project initialization: initialized")
			fmt.Fprintf(cmd.OutOrStdout(), "Mode: %s\n", mode)
			fmt.Fprintf(cmd.OutOrStdout(), "Project dir: %s\n", dir)
			fmt.Fprintf(cmd.OutOrStdout(), "Project db: %s\n", dbPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Git detected: %t\n", detectGit(projectRoot))
			fmt.Fprintf(cmd.OutOrStdout(), "Global installation: %s\n", globalStatus)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opencodeArtifacts, "opencode", false, "also generate OpenCode foundation artifact directories (profiles, prompts)")
	cmd.Flags().BoolVar(&localMode, "local", false, "use project-local storage at <root>/.plan-ai/ instead of the global default")
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "allow writing to real ~/.config/opencode when OPENCODE_CONFIG_DIR is not set")
	return cmd
}

func newSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Plan-AI integrations (opencode, MCP, etc.).",
	}
	cmd.AddCommand(newSetupOpenCodeCommand())
	cmd.AddCommand(newSetupOpenCodeWorkflowsCommand())
	cmd.AddCommand(newSetupUpdateVpsCommand())
	return cmd
}

func newSetupOpenCodeCommand() *cobra.Command {
	var allowReal bool
	cmd := &cobra.Command{
		Use:   "opencode",
		Short: "Generate OpenCode integration artifacts (config, agent, MCP, profiles, prompts, sync marker).",
		Long: `Generates the following artifacts:

  - opencode.json       — Agent configuration for Plan-AI
  - mcp-registry.json   — MCP server registrations (plan_ai.status, plan_ai.plan, plan_ai.knowledge)
  - agents/plan-ai.json — Registered agent with skills
  - profiles/           — Agent profiles (architect, researcher, reviewer)
  - prompts/            — System prompts for each profile
  - .plan-ai/opencode-sync.json — Sync marker for Plan-AI detection

All paths respect sandbox env vars:
  OPENCODE_CONFIG_DIR   — Where OpenCode config lives (required unless --allow-real-opencode is passed)
  PLAN_AI_PROJECT_ROOT  — Project root directory (default: cwd)
  PLAN_AI_HOME          — Plan-AI home directory (default: ~/.plan-ai)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opencodeDir, err := resolveOpenCodeConfigDirForWrite(allowReal)
			if err != nil {
				return fmt.Errorf("resolve opencode config dir: %w", err)
			}
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			svc := opencode.NewSetupService()
			result, err := svc.Run(opencodeDir, projectRoot)
			if err != nil {
				return fmt.Errorf("setup opencode: %w", err)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode integration artifacts generated.")
			fmt.Fprintf(out, "  opencode config: %s\n", result.OpenCodeConfigPath)
			fmt.Fprintf(out, "  mcp registry:    %s\n", result.MCPRegistryPath)
			fmt.Fprintf(out, "  agent:           %s\n", result.AgentPath)
			fmt.Fprintf(out, "  profiles:        %s\n", result.ProfilesPath)
			fmt.Fprintf(out, "  prompts:         %s\n", result.PromptsPath)
			fmt.Fprintf(out, "  workflows:       %s\n", result.WorkflowsPath)
			fmt.Fprintf(out, "  sync marker:     %s\n", result.SyncMarkerPath)

			// Also write unified MCP config
			home, err := resolveHomeRoot()
			if err != nil {
				return fmt.Errorf("resolve home: %w", err)
			}
			backupPath, err := opencode.SetupMCPConfig(home, "", allowReal)
			if err != nil {
				return fmt.Errorf("setup mcp config: %w", err)
			}
			if backupPath != "" {
				fmt.Fprintf(out, "  previous config backed up to: %s\n", backupPath)
			}
			fmt.Fprintln(out, "Unified MCP config written.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "allow writing to real ~/.config/opencode when OPENCODE_CONFIG_DIR is not set")
	return cmd
}

func newSetupOpenCodeWorkflowsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "opencode-workflows",
		Short: "Register OpenCode workflow command surface in the Plan-AI store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			reg, err := opencode.NewWorkflowService(store.NewOpenCodeWorkflowRepository(db)).Register(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode workflows registered.")
			fmt.Fprintf(out, "  id: %s\n", reg.ID)
			for _, workflow := range reg.Commands {
			fmt.Fprintf(out, "  %s: %s\n", workflow.Name, workflow.Command)
		}
		return nil
		},
	}
}

// writeGlobalConfigIfMissing writes a minimal global config.json if one
// does not already exist. It is intentionally non-destructive: a user-
// edited config.json is preserved across re-installs.
func writeGlobalConfigIfMissing(path, preset, binDir string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	cfg := map[string]any{
		"version":     1,
		"preset":      preset,
		"bin_dir":     binDir,
		"installedAt": timeNowUTC(),
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func timeNowUTC() string { return time.Now().UTC().Format(time.RFC3339) }

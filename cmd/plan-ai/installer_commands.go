package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Durru/plan-ai/internal/installer"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

// newSyncCommand returns the sync command using the new installer.
func newSyncCommand() *cobra.Command {
	var (
		dryRun    bool
		allowReal bool
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Re-apply current installation state (idempotent).",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("home dir: %w", err)
			}

			inst := installer.NewInstaller(home)
			if err := inst.LoadState(); err != nil {
				return fmt.Errorf("install first: run 'plan-ai install' — %w", err)
			}

			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "Dry-run: sync would re-apply installation state")
				fmt.Fprintf(cmd.OutOrStdout(), "  Preset: %s\n", inst.State.Preset)
				fmt.Fprintf(cmd.OutOrStdout(), "  State:  %s\n", filepath.Join(home, ".plan-ai", "state.json"))
				return nil
			}

			if err := inst.Sync(installer.InstallOptions{
				AllowReal: allowReal,
			}); err != nil {
				return fmt.Errorf("sync: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Sync complete.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without making changes")
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "Allow writing to real ~/.config/opencode")
	return cmd
}

// newUpdateCommand returns the top-level update command. It re-detects tools,
// refreshes the state, re-runs the OpenCode integration, and ensures the
// global store schema is current. It is idempotent: running it twice is the
// same as running it once.
func newUpdateCommand() *cobra.Command {
	var (
		dryRun    bool
		allowReal bool
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh global installation state and re-detect tools.",
		Long: `Re-detect installed tools, refresh $HOME/.plan-ai/state.json,
re-apply the OpenCode MCP registration (honoring $OPENCODE_CONFIG_DIR),
and ensure the global store schema is current.

This is idempotent — running it twice is the same as running it once.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}

			inst := installer.NewInstaller(home)
			if err := inst.LoadState(); err != nil {
				return fmt.Errorf("update requires an existing install; run 'plan-ai install' first: %w", err)
			}

			if dryRun {
				tools := inst.DetectTools()
				fmt.Fprintln(cmd.OutOrStdout(), "Dry-run: update would re-detect and refresh")
				fmt.Fprintf(cmd.OutOrStdout(), "  State:    %s\n", filepath.Join(home, ".plan-ai", "state.json"))
				fmt.Fprintf(cmd.OutOrStdout(), "  Preset:   %s\n", inst.State.Preset)
				fmt.Fprintf(cmd.OutOrStdout(), "  opencode: %v\n", tools.OpenCode)
				fmt.Fprintf(cmd.OutOrStdout(), "  git:      %v\n", tools.Git)
				fmt.Fprintf(cmd.OutOrStdout(), "  go:       %v\n", tools.Go)
				fmt.Fprintf(cmd.OutOrStdout(), "  mcp-srv:  %v\n", tools.MCPBinary)
				return nil
			}

			if err := inst.Sync(installer.InstallOptions{
				AllowReal: allowReal,
				BinDir:    inst.State.BinDir,
			}); err != nil {
				return fmt.Errorf("update: %w", err)
			}

			// Ensure global DB schema is current.
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

			tools := inst.State.Tools
			fmt.Fprintln(cmd.OutOrStdout(), "Update complete.")
			fmt.Fprintf(cmd.OutOrStdout(), "  State: %s\n", filepath.Join(home, ".plan-ai", "state.json"))
			fmt.Fprintf(cmd.OutOrStdout(), "  Tools: opencode=%v git=%v go=%v mcp-srv=%v\n",
				tools.OpenCode, tools.Git, tools.Go, tools.MCPBinary)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be updated without making changes")
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "Allow writing to real ~/.config/opencode")
	return cmd
}

func newUninstallCommand() *cobra.Command {
	var (
		dryRun     bool
		components []string
		allowReal  bool
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Plan-AI components or everything.",
		Long: `Uninstall Plan-AI by component or remove everything.

Without --component, removes all Plan-AI data and state.
With --component, removes only the specified components.

OpenCode config modification is safe by default: Plan-AI never writes to
real ~/.config/opencode/ unless you pass --allow-real-opencode. Without
it, the OpenCode config cleanup step is skipped.

Examples:
  plan-ai uninstall                   # Remove everything
  plan-ai uninstall --component docs  # Remove docs component only
  plan-ai uninstall --dry-run         # Show what would be removed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return fmt.Errorf("home dir: %w", err)
			}

			inst := installer.NewInstaller(home)
			if err := inst.LoadState(); err != nil {
				return fmt.Errorf("nothing to uninstall — %w", err)
			}

			if dryRun {
				if len(components) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Dry-run: full uninstall would remove:")
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", filepath.Join(home, ".plan-ai"))
					if !allowReal {
						fmt.Fprintln(cmd.OutOrStdout(), "  - OpenCode config cleanup: skipped (pass --allow-real-opencode to modify real config)")
					}
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "Dry-run: would uninstall components:")
					for _, c := range components {
						fmt.Fprintf(cmd.OutOrStdout(), "  - %s: %s\n", c, installer.ComponentDescriptions[c])
					}
				}
				return nil
			}

			if err := inst.Uninstall(components); err != nil {
				return fmt.Errorf("uninstall: %w", err)
			}

			// Only clean OpenCode config when explicitly allowed or sandboxed.
			if _, ocErr := resolveOpenCodeConfigDirForWrite(allowReal); ocErr == nil {
				if err := inst.RemovePlanAIFromOpenConfig(); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not clean OpenCode config: %v\n", err)
				}
			}

			if len(components) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Plan-AI fully uninstalled.")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled components: %s\n", strings.Join(components, ", "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be uninstalled without making changes")
	cmd.Flags().StringSliceVar(&components, "component", nil, "Component(s) to uninstall (repeatable, omit for full uninstall)")
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "allow modifying real ~/.config/opencode when OPENCODE_CONFIG_DIR is not set")
	return cmd
}

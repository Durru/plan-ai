package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plan-ai/plan-ai/internal/installer"
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

// newUninstallCommand returns the uninstall command using the new installer.
func newUninstallCommand() *cobra.Command {
	var (
		dryRun     bool
		components []string
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Plan-AI components or everything.",
		Long: `Uninstall Plan-AI by component or remove everything.

Without --component, removes all Plan-AI data and state.
With --component, removes only the specified components.

Examples:
  plan-ai uninstall                   # Remove everything
  plan-ai uninstall --component docs  # Remove docs component only
  plan-ai uninstall --dry-run         # Show what would be removed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
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
	return cmd
}

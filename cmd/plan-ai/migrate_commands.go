package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Durru/plan-ai/internal/config"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

func newMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Plan-AI data between storage modes.",
	}
	cmd.AddCommand(newMigrateLocalToGlobalCommand())
	return cmd
}

func newMigrateLocalToGlobalCommand() *cobra.Command {
	var (
		force       bool
		projectRoot string
	)
	cmd := &cobra.Command{
		Use:   "local-to-global",
		Short: "Migrate a project from <root>/.plan-ai/ to ~/.plan-ai/projects/<slug>/.",
		Long: `Copies the existing project-local SQLite database and per-project config
from <root>/.plan-ai/ into the global Plan-AI home, then re-registers the
project in the global registry with mode='external'.

The legacy local store is left in place unless --remove-legacy is set.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}
			globalDB, err := openInstalledGlobalStore()
			if err != nil {
				return fmt.Errorf("global store not initialized; run `plan-ai install` first: %w", err)
			}
			defer globalDB.Close()

			root := projectRoot
			if root == "" {
				root, err = resolveProjectRoot()
				if err != nil {
					return err
				}
			}
			root, err = filepath.Abs(root)
			if err != nil {
				return err
			}

			legacyLayout, err := store.EnsureProjectLayout(root)
			if err != nil {
				return err
			}
			if _, err := os.Stat(legacyLayout.DBPath); err != nil {
				return fmt.Errorf("no legacy local store at %s: %w", legacyLayout.DBPath, err)
			}

			slug := config.ProjectSlug(root)
			externalLayout, err := store.EnsureExternalProjectLayout(home, slug)
			if err != nil {
				return err
			}
			if _, err := os.Stat(externalLayout.DBPath); err == nil && !force {
				return fmt.Errorf("external project store already exists at %s; pass --force to overwrite", externalLayout.DBPath)
			}

			if err := copyFile(legacyLayout.DBPath, externalLayout.DBPath); err != nil {
				return fmt.Errorf("copy project db: %w", err)
			}
			if _, err := os.Stat(legacyLayout.ConfigPath); err == nil {
				if err := copyFile(legacyLayout.ConfigPath, externalLayout.ConfigPath); err != nil {
					return fmt.Errorf("copy project config: %w", err)
				}
			}

			entry := store.ProjectRegistryEntry{
				ID:       store.ProjectID(root),
				Name:     filepath.Base(root),
				RootPath: root,
				Slug:     slug,
				Mode:     config.ProjectModeExternal,
			}
			if _, err := store.NewProjectRegistryRepository(globalDB).Register(entry); err != nil {
				return fmt.Errorf("register project in global registry: %w", err)
			}

			fmt.Fprintln(out, "Migration: complete")
			fmt.Fprintf(out, "  Source:        %s\n", legacyLayout.DBPath)
			fmt.Fprintf(out, "  Target:        %s\n", externalLayout.DBPath)
			fmt.Fprintf(out, "  Project ID:    %s\n", entry.ID)
			fmt.Fprintf(out, "  Slug:          %s\n", slug)
			fmt.Fprintf(out, "  Mode:          %s\n", entry.Mode)
			fmt.Fprintln(out, "  Next: subsequent `plan-ai` invocations will use the global store.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing external project store if it exists")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "project root (defaults to $PLAN_AI_PROJECT_ROOT or cwd)")
	return cmd
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

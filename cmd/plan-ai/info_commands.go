package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/capabilities"
	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/continuous"
	"github.com/plan-ai/plan-ai/internal/core"
	"github.com/plan-ai/plan-ai/internal/installer"
	"github.com/plan-ai/plan-ai/internal/knowledge"
	"github.com/plan-ai/plan-ai/internal/opencode"
	"github.com/plan-ai/plan-ai/internal/research"
	"github.com/plan-ai/plan-ai/internal/scanner"
	"github.com/plan-ai/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

func newVersionCommand(app core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Plan-AI version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Plan-AI %s\n", app.Version)
		},
	}
}

func newDoctorCommand() *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check Plan-AI health: store, migrations, OpenCode, and component installation.",
		Long: `Check Plan-AI health.

By default doctor is read-only and only reports issues. Pass --fix to attempt
repairs for stale state issues (stale_state_opencode_missing,
registered_binary_missing) by re-running the local installer sync. Doctor
--fix never writes to the user's real ~/.config/opencode; the installer's
refuse-to-write-real-OpenCode check still applies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			globalDB := config.GlobalDBPath(home)
			projectDB := config.ProjectDBPath(projectRoot)
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Plan-AI doctor")
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "--- Storage ---")
			fmt.Fprintf(out, "Global db: %s (%s)\n", globalDB, installedLabel(config.GlobalDir(home), globalDB))
			fmt.Fprintf(out, "Project db: %s (%s)\n", projectDB, initializedLabel(config.ProjectDir(projectRoot), projectDB))
			if pathExists(globalDB) {
				db, err := store.Open(globalDB)
				if err != nil {
					return err
				}
				defer db.Close()
				if err := store.RunGlobalMigrations(db); err != nil {
					return err
				}
				fmt.Fprintln(out, "Global migrations: ok")
			}
			if pathExists(projectDB) {
				db, err := store.Open(projectDB)
				if err != nil {
					return err
				}
				defer db.Close()
				if err := store.RunProjectMigrations(db); err != nil {
					return err
				}
				fmt.Fprintln(out, "Project migrations: ok")
			}

			// OpenCode integration check
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "--- OpenCode Integration ---")
			detector := opencode.NewDetector()
			detected := detector.Detect(projectRoot)
			if detected.Found {
				fmt.Fprintf(out, "Config found: %s\n", detected.ConfigPath)
				if detected.AgentName != "" {
					fmt.Fprintf(out, "Agent: %s\n", detected.AgentName)
				}
				fmt.Fprintf(out, "Skills: %d\n", detected.SkillCount)
				fmt.Fprintf(out, "Initialized: %v\n", detected.IsInitialized)
			} else {
				fmt.Fprintln(out, "No OpenCode config detected (standalone mode)")
			}

			ocCfg, err := opencode.LoadConfig(home)
			if err == nil {
				fmt.Fprintf(out, "Integration mode: %s (enabled: %v, read-only: %v)\n", ocCfg.Mode, ocCfg.Enabled, ocCfg.ReadOnly)
			}

			doc := opencode.NewDoctor()
			checks := doc.RunChecks(detected, ocCfg)
			for _, c := range checks {
				fmt.Fprintf(out, "  [%s] %s\n", c.Status, c.Message)
			}

			// Installer-based health report
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "--- Component Installation (Gentle-AI) ---")
			inst := installer.NewInstaller(home)
			report := inst.Doctor()
			fmt.Fprintf(out, "State exists:  %v\n", report.StateExists)
			if report.StateValid {
				fmt.Fprintf(out, "Preset:        %s\n", report.Preset)
				fmt.Fprintf(out, "Components:    %d/%d installed\n", report.ComponentsInstalled, report.ComponentsTotal)
				fmt.Fprintf(out, "Data dir:      %s\n", report.DataDir)
			}
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Tools:")
			fmt.Fprintf(out, "  opencode:    %v\n", report.Tools.OpenCode)
			fmt.Fprintf(out, "  git:         %v\n", report.Tools.Git)
			fmt.Fprintf(out, "  go:          %v\n", report.Tools.Go)
			fmt.Fprintf(out, "  plan-ai-mcp: %v\n", report.Tools.MCPBinary)

			if report.StateValid {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "OpenCode sync:")
				fmt.Fprintf(out, "  Config path: %s\n", report.OpenCodeConfigPath)
				fmt.Fprintf(out, "  Valid:       %v\n", report.OpenCodeValid)
			}

			if len(report.Issues) > 0 {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "--- Issues ---")
				for _, issue := range report.Issues {
					fmt.Fprintf(out, "  [%s] %s: %s\n", issue.Severity, issue.Code, issue.Message)
				}
			}

			if fix && len(report.Issues) > 0 {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "--- Fix ---")
				repairable := false
				for _, issue := range report.Issues {
					if issue.Code == "stale_state_opencode_missing" || issue.Code == "registered_binary_missing" {
						repairable = true
						break
					}
				}
				if !repairable {
					fmt.Fprintln(out, "No auto-repairable issues found.")
					return nil
				}
				fmt.Fprintln(out, "Re-running installer sync to repair local installation...")
				inst := installer.NewInstaller(home)
				// AllowReal=false is intentional: doctor --fix only repairs the
				// local sandbox installation. The refuse-to-write-real-OpenCode
				// guard still applies and will error if the user has no
				// OPENCODE_CONFIG_DIR set and home == real home.
				if err := inst.Sync(installer.InstallOptions{Preset: report.Preset, BinDir: report.BinDir}); err != nil {
					fmt.Fprintf(out, "  fix failed: %v\n", err)
					fmt.Fprintln(out, "  (--fix repairs the local installation only; it does not write to real OpenCode config)")
					return err
				}
				fmt.Fprintln(out, "  fix complete. Run `plan-ai doctor` to verify.")
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt to repair stale state issues (local install only; never touches real OpenCode config)")
	return cmd
}

func newStatusCommand(app core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Plan-AI persistence status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}

			globalDir := config.GlobalDir(home)
			globalDB := config.GlobalDBPath(home)
			projectDir := config.ProjectDir(projectRoot)
			projectDB := config.ProjectDBPath(projectRoot)

			// Phase 1: when the project is registered in the global registry,
			// the canonical project location is the external one; otherwise we
			// fall back to the legacy project-local layout.
			mode := ""
			if _, err := os.Stat(globalDB); err == nil {
				if gDB, err := sql.Open("sqlite", globalDB); err == nil {
					if repo, repoErr := store.NewProjectRegistryRepository(gDB).GetByID(store.ProjectID(projectRoot)); repoErr == nil {
						mode = repo.Mode
						if repo.Mode == config.ProjectModeExternal {
							extDir := config.ExternalProjectDir(home, repo.Slug)
							projectDir = extDir
							projectDB = config.ExternalProjectDBPath(home, repo.Slug)
						}
					}
					gDB.Close()
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Plan-AI %s\n", app.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Global installation: %s\n", installedLabel(globalDir, globalDB))
			fmt.Fprintf(cmd.OutOrStdout(), "Global dir: %s\n", globalDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Global db: %s\n", globalDB)
			if mode != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Project mode: %s\n", mode)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Project initialization: %s\n", initializedLabel(projectDir, projectDB))
			fmt.Fprintf(cmd.OutOrStdout(), "Project dir: %s\n", projectDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Project db: %s\n", projectDB)
			counts := store.DomainCounts{}
			var scanSummary *store.ScanSummary
			var knowledgeSummary knowledge.Summary
			var researchSummary research.ResearchSummary
			hasKnowledge := false
			hasResearch := false
			if pathExists(projectDB) {
				db, err := openInitializedProjectStore()
				if err != nil {
					return err
				}
				defer db.Close()
				counts, err = store.CountDomainEntities(db)
				if err != nil {
					return err
				}
				summary, err := store.NewScanRepository(db).GetScanSummary()
				if err != nil && err != sql.ErrNoRows {
					return err
				}
				if err == nil {
					scanSummary = &summary
				}
				knowledgeSummary, err = knowledge.NewService(store.NewKnowledgeRepository(db)).GetSummary()
				if err != nil {
					return err
				}
				hasKnowledge = true
				researchSummary, err = research.NewService(store.NewResearchRepository(db)).GetSummary()
				if err != nil {
					return err
				}
				hasResearch = true
			}
			printScanSummary(cmd, scanSummary)
			printDomainCounts(cmd, counts)
			if hasKnowledge {
				printKnowledgeSummary(cmd, knowledgeSummary)
			}
			if hasResearch {
				printResearchSummary(cmd, researchSummary)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", app.Version)
			return nil
		},
	}
}

func newScanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the current project deterministically.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectDir := config.ProjectDir(projectRoot)
			projectDB := config.ProjectDBPath(projectRoot)
			if initializedLabel(projectDir, projectDB) != "initialized" {
				return fmt.Errorf("project is not initialized; run plan-ai init first")
			}

			result, err := scanner.Default().Scan(projectRoot)
			if err != nil {
				return err
			}

			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			if _, err := store.NewScanRepository(db).CreateScan(scanRecordFromResult(result)); err != nil {
				return err
			}

			printScanResult(cmd, result)
			return nil
		},
	}
}

func newCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List registered capabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectDB := config.ProjectDBPath(projectRoot)
			if !pathExists(projectDB) {
				r := capabilities.NewDefaultRegistry(nil)
				caps := r.ListCapabilities()
				fmt.Fprintln(out, "Registered capabilities (in-memory, project not initialized):")
				for _, c := range caps {
					fmt.Fprintf(out, "  %s: %s\n", c.Type, c.Name)
				}
				return nil
			}
			db, err := store.Open(projectDB)
			if err != nil {
				return fmt.Errorf("open project db: %w", err)
			}
			defer db.Close()
			if err := store.RunProjectMigrations(db); err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}
			r := capabilities.NewDefaultRegistry(db)
			caps := r.ListCapabilities()
			fmt.Fprintln(out, "Registered capabilities:")
			for _, c := range caps {
				fmt.Fprintf(out, "  %s: %s\n", c.Type, c.Name)
			}
			return nil
		},
	}
}

func newNextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Get the next pending task or action item.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectID := store.ProjectID(projectRoot)

			// First try the continuous planning status for the next task.
			db, err := openInitializedProjectStore()
			if err != nil {
				return fmt.Errorf("open store: %w", err)
			}
			defer db.Close()

			statusSvc := continuous.NewStatusService(db)
			status, err := statusSvc.GetStatus(projectID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if status.NextTask != "" {
				fmt.Fprintf(out, "Next task: %s\n", status.NextTask)
			} else {
				fmt.Fprintln(out, "No pending tasks.")
			}
			if len(status.BlockedItems) > 0 {
				fmt.Fprintf(out, "Blocked: %s\n", strings.Join(status.BlockedItems, ", "))
			}
			if len(status.ApprovalsNeeded) > 0 {
				fmt.Fprintf(out, "Approvals needed: %s\n", strings.Join(status.ApprovalsNeeded, "; "))
			}
			return nil
		},
	}
}

// ── Print helpers ──

func printScanResult(cmd *cobra.Command, result *scanner.Result) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Project scan completed.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Git:")
	fmt.Fprintf(out, "  detected: %t\n", result.GitDetected)
	fmt.Fprintf(out, "  branch: %s\n", dash(result.GitBranch))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Languages:")
	if len(result.Languages) == 0 {
		fmt.Fprintln(out, "  -")
	} else {
		for _, language := range result.Languages {
			fmt.Fprintf(out, "  %s: %d files\n", language.Language, language.Files)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Package managers:")
	if len(result.PackageManagers) == 0 {
		fmt.Fprintln(out, "  -")
	} else {
		for _, manager := range result.PackageManagers {
			fmt.Fprintf(out, "  %s\n", manager.Name)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Frameworks:")
	if len(result.Frameworks) == 0 {
		fmt.Fprintln(out, "  -")
	} else {
		for _, framework := range result.Frameworks {
			fmt.Fprintf(out, "  %s\n", framework.Name)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Dependencies:")
	if len(result.Dependencies) == 0 {
		fmt.Fprintln(out, "  -")
	} else {
		for _, dependency := range result.Dependencies {
			fmt.Fprintf(out, "  %s (%s)\n", dependency.Name, dependency.Source)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Files indexed:")
	fmt.Fprintf(out, "  %d\n", len(result.Files))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Fingerprint:")
	fmt.Fprintf(out, "  %s\n", result.Fingerprint)
}

func printScanSummary(cmd *cobra.Command, summary *store.ScanSummary) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Scan:")
	if summary == nil {
		fmt.Fprintln(out, "  not scanned yet")
		return
	}
	fmt.Fprintf(out, "  latest: %s\n", summary.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(out, "  git: %t\n", summary.GitDetected)
	fmt.Fprintf(out, "  branch: %s\n", dash(summary.GitBranch))
	fmt.Fprintf(out, "  languages: %s\n", joinOrDash(summary.LanguageNames))
	fmt.Fprintf(out, "  frameworks: %s\n", joinOrDash(summary.FrameworkNames))
	fmt.Fprintf(out, "  files indexed: %d\n", summary.FileCount)
}

func printDomainCounts(cmd *cobra.Command, counts store.DomainCounts) {
	fmt.Fprintf(cmd.OutOrStdout(), "Plans: %d\n", counts.Plans)
	fmt.Fprintf(cmd.OutOrStdout(), "Phases: %d\n", counts.Phases)
	fmt.Fprintf(cmd.OutOrStdout(), "Tasks: %d\n", counts.Tasks)
	fmt.Fprintf(cmd.OutOrStdout(), "Decisions: %d\n", counts.Decisions)
	fmt.Fprintf(cmd.OutOrStdout(), "Research: %d\n", counts.ResearchEntries)
	fmt.Fprintf(cmd.OutOrStdout(), "Knowledge: %d\n", counts.KnowledgeObjects)
	fmt.Fprintf(cmd.OutOrStdout(), "Validations: %d\n", counts.Validations)
	fmt.Fprintf(cmd.OutOrStdout(), "Snapshots: %d\n", counts.Snapshots)
}



func scanRecordFromResult(result *scanner.Result) store.Scan {
	record := store.Scan{
		ProjectRoot: result.ProjectRoot,
		GitDetected: result.GitDetected,
		GitBranch:   result.GitBranch,
		Fingerprint: result.Fingerprint,
		Summary:     result.Summary,
		CreatedAt:   result.CreatedAt,
	}
	for _, language := range result.Languages {
		record.Languages = append(record.Languages, store.ScanLanguage{Language: language.Language, FilesCount: language.Files})
	}
	for _, framework := range result.Frameworks {
		record.Frameworks = append(record.Frameworks, store.ScanFramework{Framework: framework.Name, Evidence: framework.Evidence})
	}
	for _, manager := range result.PackageManagers {
		record.PackageManagers = append(record.PackageManagers, store.ScanPackageManager{Manager: manager.Name, Evidence: manager.Evidence})
	}
	for _, dependency := range result.Dependencies {
		record.Dependencies = append(record.Dependencies, store.ScanDependency{Name: dependency.Name, Version: dependency.Version, Source: dependency.Source})
	}
	for _, file := range result.Files {
		record.Files = append(record.Files, store.ScanFile{Path: file.Path, Kind: string(file.Kind), SizeBytes: file.Size})
	}
	return record
}

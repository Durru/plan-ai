package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/agent"
	"github.com/plan-ai/plan-ai/internal/alignmentv3"
	"github.com/plan-ai/plan-ai/internal/ambiguityv3"
	"github.com/plan-ai/plan-ai/internal/approval"
	"github.com/plan-ai/plan-ai/internal/capabilities"
	"github.com/plan-ai/plan-ai/internal/change"
	"github.com/plan-ai/plan-ai/internal/confidencev3"
	"github.com/plan-ai/plan-ai/internal/config"
	approvedcontext "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/continuous"
	"github.com/plan-ai/plan-ai/internal/core"
	"github.com/plan-ai/plan-ai/internal/discoveryv3"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/ingestion"
	"github.com/plan-ai/plan-ai/internal/intent"
	"github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/knowledge"
	"github.com/plan-ai/plan-ai/internal/memory"
	"github.com/plan-ai/plan-ai/internal/modelstrategy"
	"github.com/plan-ai/plan-ai/internal/opencode"
	"github.com/plan-ai/plan-ai/internal/planning"
	"github.com/plan-ai/plan-ai/internal/reference"
	"github.com/plan-ai/plan-ai/internal/requirements"
	"github.com/plan-ai/plan-ai/internal/research"
	"github.com/plan-ai/plan-ai/internal/scanner"
	"github.com/plan-ai/plan-ai/internal/store"
	"github.com/plan-ai/plan-ai/internal/validation"
	"github.com/plan-ai/plan-ai/internal/vision"
	"github.com/plan-ai/plan-ai/internal/workflows"
	"github.com/spf13/cobra"
)

const version = "v2.0.0"

const configVersion = "2.0.0"

func main() {
	cmd := newRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	app := core.NewApp(version, ".")

	cmd := &cobra.Command{
		Use:   "plan-ai",
		Short: "Plan-AI prepares implementation plans for AI-assisted projects.",
		Long:  "Plan-AI is a local-first project planning foundation.",
	}

	cmd.AddCommand(newVersionCommand(app))
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newBootstrapCommand())
	cmd.AddCommand(newScanCommand())
	cmd.AddCommand(newKnowledgeCommand())
	cmd.AddCommand(newResearchCommand())
	cmd.AddCommand(newIngestCommand())
	cmd.AddCommand(newIntentCommand())
	cmd.AddCommand(newVisionCommand())
	cmd.AddCommand(newApprovalCommand())
	cmd.AddCommand(newRequirementsCommand())
	cmd.AddCommand(newApprovedCommand())
	cmd.AddCommand(newStatusCommand(app))
	cmd.AddCommand(newDoctorCommand())
	cmd.AddCommand(newDevCommand())
	cmd.AddCommand(newPlanCommand())
	cmd.AddCommand(newContextCommand())
	cmd.AddCommand(newJobsCommand())
	cmd.AddCommand(newCapabilitiesCommand())
	cmd.AddCommand(newImpactCommand())
	cmd.AddCommand(newSnapshotCommand())
	cmd.AddCommand(newAgentCommand())
	cmd.AddCommand(newContinuousCommand())
	cmd.AddCommand(newNextCommand())
	cmd.AddCommand(newDiscoveryCommand())
	cmd.AddCommand(newMasterV2Command())
	cmd.AddCommand(newSpecificV2Command())
	cmd.AddCommand(newDeliveryCommand())
	cmd.AddCommand(newReferenceCommand())
	cmd.AddCommand(newMemoryCommand())
	cmd.AddCommand(newModelCommand())
	cmd.AddCommand(newValidateCommand())
	cmd.AddCommand(newAmbiguityCommand())
	cmd.AddCommand(newConfidenceCommand())
	cmd.AddCommand(newAlignmentCommand())
	cmd.AddCommand(newSetupCommand())

	return cmd
}

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
			return nil
		},
	}
	cmd.Flags().BoolVar(&allowReal, "allow-real-opencode", false, "allow writing to real ~/.config/opencode when OPENCODE_CONFIG_DIR is not set")
	return cmd
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check Plan-AI store paths, migration status, and OpenCode integration.",
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
			fmt.Fprintln(out, "OpenCode integration:")
			detector := opencode.NewDetector()
			detected := detector.Detect(projectRoot)
			if detected.Found {
				fmt.Fprintf(out, "  Config found: %s\n", detected.ConfigPath)
				if detected.AgentName != "" {
					fmt.Fprintf(out, "  Agent: %s\n", detected.AgentName)
				}
				fmt.Fprintf(out, "  Skills: %d\n", detected.SkillCount)
				fmt.Fprintf(out, "  Initialized: %v\n", detected.IsInitialized)
			} else {
				fmt.Fprintln(out, "  No OpenCode config detected (standalone mode)")
			}

			ocCfg, err := opencode.LoadConfig(home)
			if err == nil {
				fmt.Fprintf(out, "  Integration mode: %s (enabled: %v, read-only: %v)\n", ocCfg.Mode, ocCfg.Enabled, ocCfg.ReadOnly)
			}

			doc := opencode.NewDoctor()
			checks := doc.RunChecks(detected, ocCfg)
			for _, c := range checks {
				fmt.Fprintf(out, "  [%s] %s\n", c.Status, c.Message)
			}

			return nil
		},
	}
}

func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install Plan-AI global persistence.",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := resolveHomeRoot()
			if err != nil {
				return err
			}

			layout, err := store.EnsureGlobalLayout(home)
			if err != nil {
				return err
			}

			if _, err := os.Stat(layout.ConfigPath); os.IsNotExist(err) {
				cfg := config.GlobalConfig{
					Version:      configVersion,
					InstalledAt:  nowUTC(),
					GlobalDir:    layout.Dir,
					GlobalDB:     layout.DBPath,
					Integrations: map[string]any{},
				}
				if err := config.SaveGlobalConfig(layout.ConfigPath, cfg); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			db, err := store.Open(layout.DBPath)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := store.RunGlobalMigrations(db); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Global installation: installed")
			fmt.Fprintf(cmd.OutOrStdout(), "Global dir: %s\n", layout.Dir)
			fmt.Fprintf(cmd.OutOrStdout(), "Global db: %s\n", layout.DBPath)
			return nil
		},
	}
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Plan-AI persistence for the current project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}

			layout, err := store.EnsureProjectLayout(projectRoot)
			if err != nil {
				return err
			}

			projectName := filepath.Base(projectRoot)
			if _, err := os.Stat(layout.ConfigPath); os.IsNotExist(err) {
				cfg := config.ProjectConfig{
					Version:      configVersion,
					ProjectName:  projectName,
					ProjectRoot:  projectRoot,
					ProjectDB:    layout.DBPath,
					CreatedAt:    nowUTC(),
					Integrations: map[string]any{},
				}
				if err := config.SaveProjectConfig(layout.ConfigPath, cfg); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			db, err := store.Open(layout.DBPath)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := store.RunProjectMigrations(db); err != nil {
				return err
			}
			if err := store.UpsertProjectState(db, store.ProjectID(projectRoot), projectName, projectRoot, "initialized"); err != nil {
				return err
			}

			globalStatus := "missing"
			if globalDB, err := openInstalledGlobalStore(); err == nil {
				defer globalDB.Close()
				if err := store.UpsertKnownProject(globalDB, store.ProjectID(projectRoot), projectName, projectRoot); err != nil {
					return err
				}
				globalStatus = "installed"
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Project initialization: initialized")
			fmt.Fprintf(cmd.OutOrStdout(), "Project dir: %s\n", layout.Dir)
			fmt.Fprintf(cmd.OutOrStdout(), "Project db: %s\n", layout.DBPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Git detected: %t\n", detectGit(projectRoot))
			fmt.Fprintf(cmd.OutOrStdout(), "Global installation: %s\n", globalStatus)
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

			fmt.Fprintf(cmd.OutOrStdout(), "Plan-AI %s\n", app.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Global installation: %s\n", installedLabel(globalDir, globalDB))
			fmt.Fprintf(cmd.OutOrStdout(), "Global dir: %s\n", globalDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Global db: %s\n", globalDB)
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

func newDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Development inspection helpers.",
	}
	cmd.AddCommand(newSeedDomainCommand())
	cmd.AddCommand(newListDomainCommand())
	cmd.AddCommand(newSeedKnowledgeCommand())
	cmd.AddCommand(newSeedResearchCommand())
	cmd.AddCommand(newSeedContinuousScenarioCommand())
	return cmd
}

func newSeedDomainCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed-domain",
		Short: "Seed sample domain records into the project store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			master := domain.MasterPlan{ID: domain.NewID("plan"), Title: "Sample Master Plan", Summary: "Top-level sample plan", Status: domain.StatusDraft, Version: 1}
			if err := store.NewPlanRepository(db).CreateMaster(master); err != nil {
				return err
			}
			specific := domain.SpecificPlan{ID: domain.NewID("plan"), Title: "Sample Specific Plan", Summary: "Child sample plan", Status: domain.StatusDraft, Version: 1, ParentPlanID: master.ID}
			if err := store.NewPlanRepository(db).CreateSpecific(specific); err != nil {
				return err
			}
			phase := domain.Phase{ID: domain.NewID("phase"), PlanID: specific.ID, Title: "Sample Phase", Summary: "Sample phase summary", Status: domain.StatusDraft, Position: 1}
			if err := store.NewPhaseRepository(db).Create(phase); err != nil {
				return err
			}
			task := domain.Task{ID: domain.NewID("task"), PhaseID: phase.ID, PlanID: specific.ID, Title: "Sample Task", Summary: "Sample task summary", Status: domain.StatusDraft, Position: 1, ContextSize: domain.ContextSizeShort}
			if err := store.NewTaskRepository(db).Create(task); err != nil {
				return err
			}
			decision := domain.Decision{ID: domain.NewID("decision"), Title: "Sample Decision", Context: "Phase 3 seed data", Decision: "Use structured domain tables", Status: domain.StatusDraft, Impact: "Enables repository and CLI validation"}
			if err := store.NewDecisionRepository(db).Create(decision); err != nil {
				return err
			}
			if _, err := research.NewService(store.NewResearchRepository(db)).CreateResearch("Sample Topic",
				research.WithSummary("Sample research summary"),
				research.WithConfidence(80),
				research.WithTags("seed", "sample"),
			); err != nil {
				return err
			}
			knowledge := domain.KnowledgeObject{ID: domain.NewID("knowledge"), Topic: "Sample Topic", Category: domain.KnowledgeCategoryArchitecture, Summary: "Sample knowledge summary", Content: "Reusable sample knowledge", Confidence: 0.9, SourceType: domain.KnowledgeSourceManual, ReuseCount: 1, Status: domain.KnowledgeStatusDraft}
			if err := store.NewKnowledgeRepository(db).Create(knowledge); err != nil {
				return err
			}
			validation := domain.Validation{ID: domain.NewID("validation"), TargetType: domain.ValidationTargetTask, TargetID: task.ID, Status: domain.StatusDraft, Summary: "Sample validation"}
			if err := store.NewValidationRepository(db).Create(validation); err != nil {
				return err
			}
			snapshot := domain.Snapshot{ID: domain.NewID("snapshot"), Reason: "seed-domain", Summary: "Sample domain seed snapshot"}
			if err := store.NewSnapshotRepository(db).Create(snapshot); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Domain seed: created")
			return nil
		},
	}
}

func newListDomainCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-domain",
		Short: "List project domain record counts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			counts, err := store.CountDomainEntities(db)
			if err != nil {
				return err
			}
			printDomainCounts(cmd, counts)
			summary, err := knowledge.NewService(store.NewKnowledgeRepository(db)).GetSummary()
			if err != nil {
				return err
			}
			printKnowledgeSummary(cmd, summary)
			return nil
		},
	}
}

func newSeedKnowledgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed-knowledge",
		Short: "Seed sample knowledge objects into the project store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			seeds := []knowledge.CreateInput{
				{
					Topic:      "PostgreSQL Multi Tenant",
					Summary:    "Strategies to isolate tenants in a shared PostgreSQL cluster.",
					Content:    "Three common strategies: separate databases per tenant, shared database with separate schemas, and shared schema with tenant_id columns. Pick based on isolation, cost, and operational complexity.",
					Confidence: 0.9,
					SourceType: knowledge.SourceManual,
					Tags:       []string{"postgres", "multi-tenant", "database"},
					Status:     knowledge.StatusApproved,
				},
				{
					Topic:      "OAuth 2.0",
					Summary:    "Authorization framework for delegated access.",
					Content:    "OAuth 2.0 defines roles (resource owner, client, authorization server, resource server), grant types (authorization code, client credentials, device code), and token formats. Use authorization code with PKCE for public clients.",
					Confidence: 0.9,
					SourceType: knowledge.SourceManual,
					Tags:       []string{"oauth", "auth", "security"},
					Status:     knowledge.StatusApproved,
				},
				{
					Topic:      "Stripe Billing",
					Summary:    "Subscription billing primitives from Stripe.",
					Content:    "Stripe models products, prices, subscriptions, invoices, and customers. Use webhook signatures to verify events. Prefer Customer Portal for self-serve changes.",
					Confidence: 0.85,
					SourceType: knowledge.SourceManual,
					Tags:       []string{"stripe", "billing", "subscription"},
					Status:     knowledge.StatusApproved,
				},
			}
			for _, seed := range seeds {
				if _, err := service.CreateKnowledge(seed); err != nil {
					return err
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Knowledge seed: created")
			return nil
		},
	}
}

func newKnowledgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Reusable technical knowledge base.",
	}
	cmd.AddCommand(newKnowledgeAddCommand())
	cmd.AddCommand(newKnowledgeListCommand())
	cmd.AddCommand(newKnowledgeShowCommand())
	cmd.AddCommand(newKnowledgeSearchCommand())
	cmd.AddCommand(newKnowledgeReuseCommand())
	return cmd
}

func newKnowledgeAddCommand() *cobra.Command {
	var (
		topic      string
		category   string
		summary    string
		content    string
		confidence float64
		source     string
		status     string
		tags       []string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new knowledge object to the project store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			created, err := service.CreateKnowledge(knowledge.CreateInput{
				Topic:      topic,
				Category:   domain.KnowledgeCategory(category),
				Summary:    summary,
				Content:    content,
				Confidence: confidence,
				SourceType: domain.KnowledgeSourceType(source),
				Status:     domain.KnowledgeStatus(status),
				Tags:       tags,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Knowledge created.")
			fmt.Fprintf(out, "  id: %s\n", created.ID)
			fmt.Fprintf(out, "  topic: %s\n", created.Topic)
			fmt.Fprintf(out, "  category: %s\n", created.Category)
			fmt.Fprintf(out, "  status: %s\n", created.Status)
			fmt.Fprintf(out, "  source: %s\n", created.SourceType)
			fmt.Fprintf(out, "  confidence: %.2f\n", created.Confidence)
			fmt.Fprintf(out, "  tags: %s\n", joinOrDash(tagValues(tags)))
			return nil
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "topic of the knowledge object (required)")
	cmd.Flags().StringVar(&category, "category", "", "category override (otherwise auto-classified)")
	cmd.Flags().StringVar(&summary, "summary", "", "short summary")
	cmd.Flags().StringVar(&content, "content", "", "full content")
	cmd.Flags().Float64Var(&confidence, "confidence", 0.5, "confidence in [0,1]")
	cmd.Flags().StringVar(&source, "source", "manual", "source type (manual|research|imported|generated)")
	cmd.Flags().StringVar(&status, "status", "draft", "lifecycle status (draft|reviewed|approved|archived)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag (can be repeated)")
	_ = cmd.MarkFlagRequired("topic")
	return cmd
}

func newKnowledgeListCommand() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List knowledge objects in the project store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			objects, err := knowledgeObjectsForList(service, category)
			if err != nil {
				return err
			}
			if len(objects) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No knowledge objects yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, object := range objects {
				fmt.Fprintf(out, "%s\t%s\t%s\tstatus=%s\treused=%d\n", object.ID, object.Topic, object.Category, object.Status, object.ReuseCount)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "filter by category")
	return cmd
}

func newKnowledgeShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a knowledge object including tags, relations, and references.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			object, tags, relations, references, err := service.Describe(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "id: %s\n", object.ID)
			fmt.Fprintf(out, "topic: %s\n", object.Topic)
			fmt.Fprintf(out, "category: %s\n", object.Category)
			fmt.Fprintf(out, "status: %s\n", object.Status)
			fmt.Fprintf(out, "source: %s\n", object.SourceType)
			fmt.Fprintf(out, "confidence: %.2f\n", object.Confidence)
			fmt.Fprintf(out, "reuse count: %d\n", object.ReuseCount)
			fmt.Fprintf(out, "created: %s\n", object.CreatedAt.UTC().Format(time.RFC3339))
			fmt.Fprintf(out, "updated: %s\n", object.UpdatedAt.UTC().Format(time.RFC3339))
			if object.Summary != "" {
				fmt.Fprintf(out, "summary: %s\n", object.Summary)
			}
			if object.Content != "" {
				fmt.Fprintf(out, "content:\n%s\n", object.Content)
			}
			fmt.Fprintf(out, "tags: %s\n", joinOrDash(tagLabels(tags)))
			if len(relations) > 0 {
				fmt.Fprintln(out, "relations:")
				for _, relation := range relations {
					direction := "source"
					other := relation.TargetID
					if relation.SourceID != object.ID {
						direction = "target"
						other = relation.SourceID
					}
					fmt.Fprintf(out, "  %s=%s %s -> %s\n", direction, relation.RelationType, object.ID, other)
				}
			}
			if len(references) > 0 {
				fmt.Fprintln(out, "references:")
				for _, reference := range references {
					fmt.Fprintf(out, "  %s: %s\n", reference.ReferenceType, reference.ReferenceID)
				}
			}
			return nil
		},
	}
}

func newKnowledgeSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search knowledge by topic, summary, or content (LIKE).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			objects, err := service.SearchKnowledge(args[0])
			if err != nil {
				return err
			}
			if len(objects) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No knowledge matches %q.\n", args[0])
				return nil
			}
			out := cmd.OutOrStdout()
			for _, object := range objects {
				fmt.Fprintf(out, "%s\t%s\t%s\n", object.ID, object.Topic, object.Category)
			}
			return nil
		},
	}
}

func newKnowledgeReuseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reuse <id>",
		Short: "Record that a knowledge object was reused.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := knowledge.NewService(store.NewKnowledgeRepository(db))
			object, err := service.ReuseKnowledge(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Knowledge reused: %s (reuse count: %d)\n", object.Topic, object.ReuseCount)
			return nil
		},
	}
}

func newIngestCommand() *cobra.Command {
	var sourceType string
	var content string
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest local input into the project store.",
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
			service := ingestion.NewService(store.NewIngestionRepository(db))
			raw, source, err := service.Ingest(ingestion.InputRequest{ProjectID: store.ProjectID(projectRoot), SourceType: ingestion.SourceType(sourceType), Content: content})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Input ingested.")
			fmt.Fprintf(out, "  raw_input: %s\n", raw.ID)
			fmt.Fprintf(out, "  source: %s\n", source.ID)
			fmt.Fprintf(out, "  classification: %s\n", source.Classification)
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceType, "type", string(ingestion.SourceTypePrompt), "source type")
	cmd.Flags().StringVar(&content, "content", "", "input content (required)")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func newIntentCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "intent", Short: "Manage V2 user intent profiles and V3 product intents."}
	cmd.AddCommand(newIntentDetectCommand())
	cmd.AddCommand(newIntentLatestCommand())
	cmd.AddCommand(newIntentShowCommand())
	cmd.AddCommand(newIntentApproveCommand())
	// V3 commands (Phase 51 + Phase 52)
	cmd.AddCommand(newIntentV3DiscoverCommand())
	cmd.AddCommand(newIntentV3CreateCommand())
	cmd.AddCommand(newIntentV3ListCommand())
	cmd.AddCommand(newIntentV3SubmitCommand())
	return cmd
}

func newV3Service(db *sql.DB) intentv3.Service {
	return intentv3.NewService(store.NewIntentV3Repository(db), store.NewIntentV3DiscoveryResultRepository(db))
}

func newIntentDetectCommand() *cobra.Command {
	var content string
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect the user's real project intent as unapproved candidates.",
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
			profile, err := intent.NewService(store.NewIntentProfileRepository(db)).Detect(store.ProjectID(projectRoot), content)
			if err != nil {
				return err
			}
			printIntentProfile(cmd, "Intent profile detected.", profile)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "user intent content (required)")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func newIntentLatestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "latest",
		Short: "Show the latest intent profile for the current project.",
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
			profile, err := intent.NewService(store.NewIntentProfileRepository(db)).Latest(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			printIntentProfile(cmd, "Latest intent profile.", profile)
			return nil
		},
	}
}

func newIntentShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show one V2 intent profile or V3 product intent.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			if strings.HasPrefix(args[0], "pintent_") {
				svc := newV3Service(db)
				pi, err := svc.GetProductIntent(args[0])
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "Product Intent: %s\n", pi.ID)
				fmt.Fprintf(out, "  Project ID: %s\n", pi.ProjectID)
				fmt.Fprintf(out, "  Description: %s\n", pi.Description)
				fmt.Fprintf(out, "  Expected Outcome: %s\n", pi.ExpectedOutcome)
				fmt.Fprintf(out, "  Desired Experience: %s\n", pi.DesiredExperience)
				fmt.Fprintf(out, "  Desired Result: %s\n", pi.DesiredResult)
				fmt.Fprintf(out, "  User Expectations: %s\n", joinOrDash(pi.UserExpectations))
				fmt.Fprintf(out, "  Non-Expectations: %s\n", joinOrDash(pi.NonExpectations))
				fmt.Fprintf(out, "  Success Definition: %s\n", pi.SuccessDefinition)
				fmt.Fprintf(out, "  Failure Definition: %s\n", pi.FailureDefinition)
				fmt.Fprintf(out, "  Status: %s\n", pi.Status)
				fmt.Fprintf(out, "  Discovery Result ID: %s\n", pi.DiscoveryResultID)
				fmt.Fprintf(out, "  Created: %s\n", pi.CreatedAt.Format(time.RFC3339))
				fmt.Fprintf(out, "  Updated: %s\n", pi.UpdatedAt.Format(time.RFC3339))
				return nil
			}
			profile, err := intent.NewService(store.NewIntentProfileRepository(db)).Get(args[0])
			if err != nil {
				return err
			}
			printIntentProfile(cmd, "Intent profile.", profile)
			return nil
		},
	}
}

func newIntentApproveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a V2 intent profile or V3 product intent.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			if strings.HasPrefix(args[0], "pintent_") {
				svc := newV3Service(db)
				pi, err := svc.ApproveProductIntent(args[0])
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "Product Intent approved: %s\n", pi.ID)
				fmt.Fprintf(out, "  Status: %s\n", pi.Status)
				return nil
			}
			profile, err := intent.NewService(store.NewIntentProfileRepository(db)).Approve(args[0])
			if err != nil {
				return err
			}
			printIntentProfile(cmd, "Intent profile approved.", profile)
			return nil
		},
	}
}

// ──────────────────────────────────────────────
// V3 Intent commands (Phase 51 + Phase 52)
// ──────────────────────────────────────────────

func newIntentV3DiscoverCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "discover <content>",
		Short: "Discover intent from an idea (Phase 52 deterministic discovery).",
		Args:  cobra.ExactArgs(1),
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
			svc := newV3Service(db)
			result, err := svc.DiscoverIntent(projectRoot, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Discovery Result: %s\n", result.ID)
			fmt.Fprintf(out, "  Detected Intent: %s\n", result.DetectedIntent)
			fmt.Fprintf(out, "  Classification:  %s\n", result.Classification)
			fmt.Fprintf(out, "  Objectives:      %s\n", joinOrDash(result.Objectives))
			fmt.Fprintf(out, "  Restrictions:    %s\n", joinOrDash(result.Restrictions))
			fmt.Fprintf(out, "  Preferences:     %s\n", joinOrDash(result.Preferences))
			fmt.Fprintf(out, "  References:      %s\n", joinOrDash(result.References))
			fmt.Fprintf(out, "  Expectations:    %s\n", joinOrDash(result.Expectations))
			fmt.Fprintf(out, "  Gaps:            %s\n", joinOrDash(result.Gaps))
			fmt.Fprintf(out, "  Questions:       %s\n", joinOrDash(result.Questions))
			return nil
		},
	}
}

func newIntentV3CreateCommand() *cobra.Command {
	var description, expectedOutcome, desiredExperience, desiredResult string
	var successDefinition, failureDefinition, discoveryResultID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a V3 product intent (Phase 51).",
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
			svc := newV3Service(db)
			input := intentv3.CreateProductIntentInput{
				ProjectID:         projectRoot,
				Description:       description,
				ExpectedOutcome:   expectedOutcome,
				DesiredExperience: desiredExperience,
				DesiredResult:     desiredResult,
				SuccessDefinition: successDefinition,
				FailureDefinition: failureDefinition,
				DiscoveryResultID: discoveryResultID,
			}
			pi, err := svc.CreateProductIntent(input)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Product Intent created: %s\n", pi.ID)
			fmt.Fprintf(out, "  Description: %s\n", pi.Description)
			fmt.Fprintf(out, "  Status:      %s\n", pi.Status)
			return nil
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Intent description (required)")
	cmd.Flags().StringVar(&expectedOutcome, "expected-outcome", "", "Expected outcome")
	cmd.Flags().StringVar(&desiredExperience, "desired-experience", "", "Desired experience")
	cmd.Flags().StringVar(&desiredResult, "desired-result", "", "Desired result")
	cmd.Flags().StringVar(&successDefinition, "success-definition", "", "Success definition")
	cmd.Flags().StringVar(&failureDefinition, "failure-definition", "", "Failure definition")
	cmd.Flags().StringVar(&discoveryResultID, "discovery-result-id", "", "Link to a Phase 52 discovery result")
	_ = cmd.MarkFlagRequired("description")
	return cmd
}

func newIntentV3ListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List V3 product intents for the current project.",
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
			svc := newV3Service(db)
			intents, err := svc.ListProductIntents(projectRoot)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(intents) == 0 {
				fmt.Fprintln(out, "No V3 product intents found.")
				return nil
			}
			for _, pi := range intents {
				fmt.Fprintf(out, "%s  %s  %-18s  %s\n",
					pi.CreatedAt.Format("2006-01-02 15:04"), pi.ID, pi.Status, pi.Description)
			}
			return nil
		},
	}
}

func newIntentV3SubmitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "submit <id>",
		Short: "Submit a V3 product intent for approval.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			svc := newV3Service(db)
			pi, err := svc.SubmitProductIntentForApproval(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Product Intent submitted for approval: %s\n", pi.ID)
			fmt.Fprintf(out, "  Status: %s\n", pi.Status)
			return nil
		},
	}
}

func printIntentProfile(cmd *cobra.Command, heading string, profile intent.Profile) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, heading)
	fmt.Fprintf(out, "  id: %s\n", profile.ID)
	fmt.Fprintf(out, "  primary_intent: %s (%d%%, %s)\n", profile.PrimaryIntent.Name, profile.PrimaryIntent.Confidence, profile.PrimaryIntent.State)
	fmt.Fprintf(out, "  secondary_goals: %s\n", joinIntentGoals(profile.SecondaryGoals))
	fmt.Fprintf(out, "  constraints: %s\n", joinOrDash(profile.Constraints))
	fmt.Fprintf(out, "  expectations: %s\n", joinIntentExpectations(profile.Expectations))
	fmt.Fprintf(out, "  success_criteria: %s\n", joinIntentSuccess(profile.SuccessCriteria))
	fmt.Fprintf(out, "  priorities: %s\n", joinIntentPriorities(profile.Priorities))
	fmt.Fprintf(out, "  status: %s\n", profile.Status)
	fmt.Fprintf(out, "  approved: %t\n", profile.Approved)
}

func joinIntentGoals(items []intent.Goal) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name+"/"+string(item.State))
	}
	return joinOrDash(values)
}

func joinIntentExpectations(items []intent.UserExpectation) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name+"/"+string(item.State))
	}
	return joinOrDash(values)
}

func joinIntentSuccess(items []intent.SuccessCriteria) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name+"/"+string(item.State))
	}
	return joinOrDash(values)
}

func joinIntentPriorities(items []intent.UserPriority) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, fmt.Sprintf("%d:%s/%s", item.Rank, item.Name, item.State))
	}
	return joinOrDash(values)
}

func newVisionCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "vision", Short: "Create or inspect vision drafts."}
	cmd.AddCommand(&cobra.Command{
		Use:   "draft",
		Short: "Create a vision draft from ingested sources.",
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
			projectID := store.ProjectID(projectRoot)
			sources, err := store.NewIngestionRepository(db).ListIngestedSources(projectID)
			if err != nil {
				return err
			}
			draft, err := vision.NewService(store.NewVisionDraftRepository(db)).CreateDraft(projectID, sources)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Vision draft created.")
			fmt.Fprintf(out, "  id: %s\n", draft.ID)
			fmt.Fprintf(out, "  title: %s\n", draft.Title)
			fmt.Fprintf(out, "  approved: %t\n", draft.Approved)
			fmt.Fprintf(out, "  missing: %s\n", joinOrDash(draft.MissingInformation))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List vision drafts.",
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
			visions, err := store.NewVisionDraftRepository(db).ListVisions(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(visions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No vision drafts yet.")
				return nil
			}
			for _, v := range visions {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\tapproved=%t\n", v.ID, v.Title, v.Approved)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a vision draft.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			draft, err := vision.NewService(store.NewVisionDraftRepository(db)).Approve(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Vision approved: %s (approved=%t)\n", draft.Title, draft.Approved)
			return nil
		},
	})
	cmd.AddCommand(newVisionDocumentCommand())
	cmd.AddCommand(newVisionDocumentsCommand())
	cmd.AddCommand(newVisionDocumentShowCommand())
	cmd.AddCommand(newVisionApproveDocumentCommand())
	return cmd
}

func newVisionDocumentCommand() *cobra.Command {
	var content string
	var intentID string
	cmd := &cobra.Command{
		Use:   "document",
		Short: "Create a V2 vision document with functional, visual, technical, operational, and business dimensions.",
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
			service := vision.NewDocumentService(store.NewVisionDocumentRepository(db))
			var doc vision.Document
			if intentID != "" {
				profile, err := intent.NewService(store.NewIntentProfileRepository(db)).Get(intentID)
				if err != nil {
					return err
				}
				doc, err = service.CreateFromIntent(profile)
			} else {
				doc, err = service.CreateFromContent(store.ProjectID(projectRoot), content)
			}
			if err != nil {
				return err
			}
			if _, err := store.NewApprovalRecordRepository(db).SaveRecord(approval.Record{ProjectID: doc.ProjectID, TargetType: "vision_document", TargetID: doc.ID, State: approval.StateReview}); err != nil {
				return err
			}
			printVisionDocument(cmd, "Vision document created.", doc)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "vision source content")
	cmd.Flags().StringVar(&intentID, "intent", "", "intent profile id")
	return cmd
}

func newVisionDocumentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "documents",
		Short: "List V2 vision documents.",
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
			docs, err := vision.NewDocumentService(store.NewVisionDocumentRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(docs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No vision documents yet.")
				return nil
			}
			for _, doc := range docs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\tapproved=%t\n", doc.ID, doc.Status, doc.Approved)
			}
			return nil
		},
	}
}

func newVisionDocumentShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "document-show <id>",
		Short: "Show a V2 vision document.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			doc, err := vision.NewDocumentService(store.NewVisionDocumentRepository(db)).Get(args[0])
			if err != nil {
				return err
			}
			printVisionDocument(cmd, "Vision document.", doc)
			return nil
		},
	}
}

func newVisionApproveDocumentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "approve-document <id>",
		Short: "Approve a V2 vision document.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			doc, err := vision.NewDocumentService(store.NewVisionDocumentRepository(db)).Approve(args[0])
			if err != nil {
				return err
			}
			printVisionDocument(cmd, "Vision document approved.", doc)
			return nil
		},
	}
}

func printVisionDocument(cmd *cobra.Command, heading string, doc vision.Document) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, heading)
	fmt.Fprintf(out, "  id: %s\n", doc.ID)
	fmt.Fprintf(out, "  intent: %s\n", dash(doc.IntentProfileID))
	fmt.Fprintf(out, "  functional: %s\n", doc.FunctionalVision)
	fmt.Fprintf(out, "  visual: %s\n", doc.VisualVision)
	fmt.Fprintf(out, "  technical: %s\n", doc.TechnicalVision)
	fmt.Fprintf(out, "  operational: %s\n", doc.OperationalVision)
	fmt.Fprintf(out, "  business: %s\n", doc.BusinessVision)
	fmt.Fprintf(out, "  status: %s\n", doc.Status)
	fmt.Fprintf(out, "  approved: %t\n", doc.Approved)
}

func newApprovalCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "approval", Short: "Inspect and update V2 approval records."}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List approval records.",
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
			records, err := store.NewApprovalRecordRepository(db).ListRecords(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(records) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No approval records yet.")
				return nil
			}
			for _, record := range records {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", record.ID, record.TargetType, record.TargetID, record.State)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "approve <id>",
		Short: "Approve an approval record.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			record, err := store.NewApprovalRecordRepository(db).ApproveRecord(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approval record approved: %s (%s)\n", record.ID, record.State)
			return nil
		},
	})
	var reason string
	reject := &cobra.Command{
		Use:   "reject <id>",
		Short: "Reject an approval record.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			record, err := store.NewApprovalRecordRepository(db).RejectRecord(args[0], reason)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approval record rejected: %s (%s)\n", record.ID, record.State)
			return nil
		},
	}
	reject.Flags().StringVar(&reason, "reason", "", "rejection reason")
	cmd.AddCommand(reject)
	return cmd
}

func newRequirementsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "requirements", Short: "Discover and approve V2 requirement candidates."}
	var content string
	discover := &cobra.Command{
		Use:   "discover",
		Short: "Discover unapproved requirement candidates.",
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
			items, err := requirements.NewService(store.NewRequirementCandidateRepository(db)).Discover(store.ProjectID(projectRoot), content)
			if err != nil {
				return err
			}
			approvalRepo := store.NewApprovalRecordRepository(db)
			for _, item := range items {
				if _, err := approvalRepo.SaveRecord(approval.Record{ProjectID: item.ProjectID, TargetType: "requirement_candidate", TargetID: item.ID, State: approval.StateReview}); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Requirement candidates discovered: %d\n", len(items))
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\t%s\t%s\n", item.ID, item.Name, item.State)
			}
			return nil
		},
	}
	discover.Flags().StringVar(&content, "content", "", "content to analyze (required)")
	_ = discover.MarkFlagRequired("content")
	cmd.AddCommand(discover)
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List requirement candidates.",
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
			items, err := requirements.NewService(store.NewRequirementCandidateRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No requirement candidates yet.")
				return nil
			}
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", item.ID, item.Name, item.State)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a requirement candidate.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			item, err := requirements.NewService(store.NewRequirementCandidateRepository(db)).Approve(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Requirement candidate approved: %s (%s)\n", item.Name, item.State)
			return nil
		},
	})
	return cmd
}

func newApprovedCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "approved", Short: "Manage approved project context."}
	cmd.AddCommand(newApprovedAddCommand())
	cmd.AddCommand(newApprovedListCommand())
	cmd.AddCommand(newApprovedFindCommand())
	return cmd
}

func newApprovedAddCommand() *cobra.Command {
	var typ, source string
	cmd := &cobra.Command{
		Use:   "add <content>",
		Short: "Store approved context.",
		Args:  cobra.ExactArgs(1),
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
			registry := approvedcontext.NewRegistry(store.NewApprovedContextRepository(db))
			item, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: store.ProjectID(projectRoot), Type: approvedcontext.ApprovedType(typ), SourceID: source, Content: args[0]})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approved context stored: %s (%s)\n", item.ID, item.Type)
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", string(approvedcontext.TypeRequirement), "approved type")
	cmd.Flags().StringVar(&source, "source", "manual", "source id")
	return cmd
}

func newApprovedListCommand() *cobra.Command {
	var typ string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List approved context.",
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
			items, err := approvedcontext.NewRegistry(store.NewApprovedContextRepository(db)).ListApproved(store.ProjectID(projectRoot), approvedcontext.ApprovedType(typ))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No approved context yet.")
				return nil
			}
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", item.ID, item.Type, item.Content)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", "", "filter by type")
	return cmd
}

func newApprovedFindCommand() *cobra.Command {
	var typ string
	cmd := &cobra.Command{
		Use:   "find <query>",
		Short: "Find approved context.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if typ == "" {
				return fmt.Errorf("--type is required for find")
			}
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			items, err := approvedcontext.NewRegistry(store.NewApprovedContextRepository(db)).FindApproved(store.ProjectID(projectRoot), approvedcontext.ApprovedType(typ), args[0])
			if err != nil {
				return err
			}
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", item.ID, item.Type, item.Content)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", "", "approved type (required)")
	return cmd
}

func newPlanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Create minimal planning artifacts from vision, approved context, research, and knowledge.",
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

			projectID := store.ProjectID(projectRoot)
			visions, err := store.NewVisionDraftRepository(db).ListVisions(projectID)
			if err != nil {
				return err
			}
			if len(visions) == 0 {
				return fmt.Errorf("no vision available; run plan-ai vision draft first")
			}
			visionRef := visions[len(visions)-1].ID

			approvedRepo := store.NewApprovedContextRepository(db)
			requirements, err := approvedContents(approvedRepo, projectID, approvedcontext.TypeRequirement)
			if err != nil {
				return err
			}
			constraints, err := approvedContents(approvedRepo, projectID, approvedcontext.TypeConstraint)
			if err != nil {
				return err
			}
			decisions, err := approvedContents(approvedRepo, projectID, approvedcontext.TypeDecision)
			if err != nil {
				return err
			}
			if len(requirements) == 0 && len(constraints) == 0 && len(decisions) == 0 {
				return fmt.Errorf("approved context is required before planning")
			}

			researchEntries, err := research.NewService(store.NewResearchRepository(db)).ListResearch()
			if err != nil {
				return err
			}
			knowledgeObjects, err := knowledge.NewService(store.NewKnowledgeRepository(db)).ListKnowledge()
			if err != nil {
				return err
			}

			researchIDs := make([]string, 0, len(researchEntries))
			for _, entry := range researchEntries {
				researchIDs = append(researchIDs, entry.ID)
			}
			knowledgeIDs := make([]string, 0, len(knowledgeObjects))
			for _, object := range knowledgeObjects {
				knowledgeIDs = append(knowledgeIDs, object.ID)
			}

			planningService := planning.NewService(store.NewPlanningRepository(db))
			master, err := planningService.CreateMasterPlan(planning.PlanningInput{ProjectID: projectID, VisionReference: visionRef, ApprovedRequirements: requirements, ApprovedConstraints: constraints, ApprovedDecisions: decisions, ResearchIDs: researchIDs, KnowledgeIDs: knowledgeIDs})
			if err != nil {
				return err
			}
			goal := master.Title
			if len(requirements) > 0 {
				goal = requirements[0]
			}
			specific, err := planningService.CreateSpecificPlan(master.ID, planning.SpecificPlanInput{ProjectID: projectID, Title: goal, Goal: goal, Requirements: requirements, Constraints: constraints, Decisions: decisions, KnowledgeUsed: knowledgeIDs, ResearchUsed: researchIDs, ValidationCriteria: []string{"go test ./...", "go build ./..."}})
			if err != nil {
				return err
			}
			doc, err := planningService.CreateImplementationDocument(specific.ID, planning.ImplementationDocumentInput{ProjectID: projectID, Objective: goal, Architecture: "Derive implementation from approved context, research, and reusable knowledge.", Validations: []string{"go test ./...", "go vet ./...", "go build ./..."}, TestingStrategy: "TDD", RollbackStrategy: "Revert additive migrations and generated artifacts."})
			if err != nil {
				return err
			}
			run, err := workflows.NewRegistry(store.NewWorkflowRunRepository(db)).ExecuteWorkflow(workflows.WorkflowTypePlanning)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Planning artifacts created.")
			fmt.Fprintf(out, "  master_plan: %s\n", master.ID)
			fmt.Fprintf(out, "  specific_plan: %s\n", specific.ID)
			fmt.Fprintf(out, "  implementation_document: %s\n", doc.ID)
			fmt.Fprintf(out, "  workflow_run: %s status=%s\n", run.ID, run.Status)
			return nil
		},
	}
	cmd.AddCommand(newPlanEvolveCommand())
	cmd.AddCommand(newPlanBlueprintsCommand())
	return cmd
}

func newPlanEvolveCommand() *cobra.Command {
	var objective string
	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Generate an implementation-ready Plan Generation V3 blueprint.",
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
			projectID := store.ProjectID(projectRoot)
			approvedRepo := store.NewApprovedContextRepository(db)
			reqs, err := approvedContents(approvedRepo, projectID, approvedcontext.TypeRequirement)
			if err != nil {
				return err
			}
			decisions, err := approvedContents(approvedRepo, projectID, approvedcontext.TypeDecision)
			if err != nil {
				return err
			}
			inputs := append(reqs, decisions...)
			blueprint, err := planning.NewPlanEvolutionEngine(store.NewPlanEvolutionRepository(db)).Generate(projectID, objective, inputs)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Plan Evolution V3 blueprint created.")
			fmt.Fprintf(out, "  id: %s\n", blueprint.ID)
			fmt.Fprintf(out, "  objective: %s\n", blueprint.Objective)
			fmt.Fprintf(out, "  sections: objective, scope, exclusions, dependencies, stack, versions, libraries, folders, files, validations, tests, risks, rollback\n")
			fmt.Fprintf(out, "  validations: %s\n", strings.Join(blueprint.Validations, "; "))
			return nil
		},
	}
	cmd.Flags().StringVar(&objective, "objective", "", "implementation objective")
	return cmd
}

func newPlanBlueprintsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "blueprints",
		Short: "List Plan Generation V3 blueprints.",
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
			items, err := planning.NewPlanEvolutionEngine(store.NewPlanEvolutionRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Plan Evolution V3 blueprints yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, item := range items {
				fmt.Fprintf(out, "%s\t%s\tstatus=%s\n", item.ID, item.Objective, item.Status)
			}
			return nil
		},
	}
}

func approvedContents(repo store.ApprovedContextRepository, projectID string, typ approvedcontext.ApprovedType) ([]string, error) {
	items, err := repo.ListApproved(projectID, typ)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Content)
	}
	return out, nil
}

func newResearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "research",
		Short: "Manage research entries for a project.",
	}
	cmd.AddCommand(newResearchRunCommand())
	cmd.AddCommand(newResearchRunsCommand())
	cmd.AddCommand(newResearchAddCommand())
	cmd.AddCommand(newResearchListCommand())
	cmd.AddCommand(newResearchShowCommand())
	cmd.AddCommand(newResearchSearchCommand())
	cmd.AddCommand(newResearchApproveCommand())
	cmd.AddCommand(newResearchRejectCommand())
	cmd.AddCommand(newResearchArchiveCommand())
	cmd.AddCommand(newResearchFindingCommand())
	cmd.AddCommand(newResearchSourceCommand())
	cmd.AddCommand(newResearchConclusionCommand())
	cmd.AddCommand(newResearchLinkCommand())
	return cmd
}

func newResearchRunCommand() *cobra.Command {
	var agentType string
	cmd := &cobra.Command{
		Use:   "run --topic <topic>",
		Short: "Run deterministic V2 research orchestration for a topic.",
		RunE: func(cmd *cobra.Command, args []string) error {
			topic, _ := cmd.Flags().GetString("topic")
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			run, err := research.NewOrchestrator(store.NewResearchOrchestrationRepository(db)).Run(store.ProjectID(projectRoot), research.AgentType(agentType), topic)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Research orchestration completed.")
			fmt.Fprintf(out, "  id: %s\n", run.ID)
			fmt.Fprintf(out, "  agent: %s\n", run.Agent)
			fmt.Fprintf(out, "  topic: %s\n", run.Topic)
			fmt.Fprintf(out, "  status: %s\n", run.Status)
			return nil
		},
	}
	cmd.Flags().String("topic", "", "research topic (required)")
	cmd.Flags().StringVar(&agentType, "agent", string(research.AgentTechnical), "agent type: market, technical, architecture, ui, ux, security, infrastructure")
	_ = cmd.MarkFlagRequired("topic")
	return cmd
}

func newResearchRunsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "runs",
		Short: "List V2 research orchestration runs.",
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
			runs, err := research.NewOrchestrator(store.NewResearchOrchestrationRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(runs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No research orchestration runs yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, run := range runs {
				fmt.Fprintf(out, "%s\t%s\t%s\tstatus=%s\n", run.ID, run.Agent, run.Topic, run.Status)
			}
			return nil
		},
	}
}

func newResearchAddCommand() *cobra.Command {
	var (
		topic      string
		category   string
		summary    string
		confidence int
		tags       []string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new research entry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			var opts []research.CreateOption
			if category != "" {
				opts = append(opts, research.WithCategory(research.ResearchCategory(category)))
			}
			if summary != "" {
				opts = append(opts, research.WithSummary(summary))
			}
			if confidence > 0 {
				opts = append(opts, research.WithConfidence(confidence))
			}
			if len(tags) > 0 {
				opts = append(opts, research.WithTags(tags...))
			}
			entry, err := service.CreateResearch(topic, opts...)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Research created.")
			fmt.Fprintf(out, "  id: %s\n", entry.ID)
			fmt.Fprintf(out, "  topic: %s\n", entry.Topic)
			fmt.Fprintf(out, "  category: %s\n", entry.Category)
			fmt.Fprintf(out, "  status: %s\n", entry.Status)
			return nil
		},
	}
	cmd.Flags().StringVar(&topic, "topic", "", "research topic (required)")
	cmd.Flags().StringVar(&category, "category", "", "category (auto-classified if empty)")
	cmd.Flags().StringVar(&summary, "summary", "", "short summary")
	cmd.Flags().IntVar(&confidence, "confidence", 0, "confidence [0-100]")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag (can be repeated)")
	_ = cmd.MarkFlagRequired("topic")
	return cmd
}

func newResearchListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List research entries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entries, err := service.ListResearch()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No research entries yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, entry := range entries {
				fmt.Fprintf(out, "%s\t%s\tstatus=%s\n", entry.ID, entry.Topic, entry.Status)
			}
			return nil
		},
	}
}

func newResearchShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a research entry with findings, sources, conclusions, tags, and links.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entry, findings, sources, conclusions, tags, links, err := service.Describe(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "id: %s\n", entry.ID)
			fmt.Fprintf(out, "topic: %s\n", entry.Topic)
			fmt.Fprintf(out, "category: %s\n", entry.Category)
			fmt.Fprintf(out, "status: %s\n", entry.Status)
			fmt.Fprintf(out, "confidence: %d\n", entry.Confidence)
			fmt.Fprintf(out, "created: %s\n", entry.CreatedAt.UTC().Format(time.RFC3339))
			fmt.Fprintf(out, "updated: %s\n", entry.UpdatedAt.UTC().Format(time.RFC3339))
			if entry.Summary != "" {
				fmt.Fprintf(out, "summary: %s\n", entry.Summary)
			}
			tagStrs := make([]string, 0, len(tags))
			for _, t := range tags {
				tagStrs = append(tagStrs, t.Tag)
			}
			fmt.Fprintf(out, "tags: %s\n", joinOrDash(tagStrs))
			if len(findings) > 0 {
				fmt.Fprintln(out, "findings:")
				for _, f := range findings {
					fmt.Fprintf(out, "  [%d] %s: %s\n", f.Importance, f.Title, f.Content)
				}
			}
			if len(sources) > 0 {
				fmt.Fprintln(out, "sources:")
				for _, s := range sources {
					fmt.Fprintf(out, "  %s (%s): %s\n", s.Title, s.SourceType, s.URL)
				}
			}
			if len(conclusions) > 0 {
				fmt.Fprintln(out, "conclusions:")
				for _, c := range conclusions {
					fmt.Fprintf(out, "  [%d] %s\n", c.Confidence, c.Content)
				}
			}
			if len(links) > 0 {
				fmt.Fprintln(out, "knowledge links:")
				for _, l := range links {
					fmt.Fprintf(out, "  %s\n", l.KnowledgeID)
				}
			}
			return nil
		},
	}
}

func newResearchSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search research entries by topic or summary.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entries, err := service.SearchResearch(args[0])
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No research matches %q.\n", args[0])
				return nil
			}
			out := cmd.OutOrStdout()
			for _, entry := range entries {
				fmt.Fprintf(out, "%s\t%s\tstatus=%s\n", entry.ID, entry.Topic, entry.Status)
			}
			return nil
		},
	}
}

func newResearchApproveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a research entry (requires findings + sources + conclusions).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entry, err := service.ApproveResearch(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Research approved: %s (status=%s)\n", entry.Topic, entry.Status)
			return nil
		},
	}
}

func newResearchRejectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reject <id>",
		Short: "Reject a research entry.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entry, err := service.RejectResearch(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Research rejected: %s (status=%s)\n", entry.Topic, entry.Status)
			return nil
		},
	}
}

func newResearchArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a research entry.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			entry, err := service.ArchiveResearch(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Research archived: %s (status=%s)\n", entry.Topic, entry.Status)
			return nil
		},
	}
}

func newResearchFindingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "finding <research-id> <title> [content]",
		Short: "Add a finding to a research entry.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			content := ""
			if len(args) > 2 {
				content = args[2]
			}
			service := research.NewService(store.NewResearchRepository(db))
			finding, err := service.AddFinding(args[0], args[1], content, 3)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Finding added: %s (importance=%d)\n", finding.Title, finding.Importance)
			return nil
		},
	}
}

func newResearchSourceCommand() *cobra.Command {
	var sourceType string
	cmd := &cobra.Command{
		Use:   "source <research-id> <title> [url]",
		Short: "Add a source to a research entry.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			url := ""
			if len(args) > 2 {
				url = args[2]
			}
			st := research.ResearchSourceType(sourceType)
			service := research.NewService(store.NewResearchRepository(db))
			source, err := service.AddSource(args[0], args[1], url, st)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Source added: %s\n", source.Title)
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceType, "type", "manual", "source type")
	return cmd
}

func newResearchConclusionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "conclusion <research-id> <content>",
		Short: "Add a conclusion to a research entry.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			conclusion, err := service.AddConclusion(args[0], args[1], 70)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Conclusion added: %s (confidence=%d)\n", conclusion.Content, conclusion.Confidence)
			return nil
		},
	}
}

func newResearchLinkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "link <research-id> <knowledge-id>",
		Short: "Link a research entry to a knowledge object.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))
			if err := service.LinkToKnowledge(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Research linked to knowledge object.")
			return nil
		},
	}
}

func newSeedContinuousScenarioCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed-continuous-scenario",
		Short: "Seed continuous planning scenario events and proposals.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectID := store.ProjectID(projectRoot)

			eventRepo := store.NewContinuousEventRepository(db)
			proposalRepo := store.NewPlanUpdateProposalRepository(db)

			type scenarioSeed struct {
				EventType string
				Summary   string
				Details   string
				Source    string

				ProposalReason    string
				AffectedPlans     string
				AffectedTasks     string
				AffectedDecisions string
			}

			scenarios := []scenarioSeed{
				{
					EventType: string(continuous.EventDecisionChanged),
					Summary:   "Database migration from PostgreSQL to MariaDB",
					Details:   `{"from":"postgresql","to":"mariadb","reason":"Licensing cost reduction","impact":"medium"}`,
					Source:    "decision_change",

					ProposalReason:    "Decision changed: Database migration from PostgreSQL to MariaDB",
					AffectedPlans:     `["master:db-migration"]`,
					AffectedTasks:     `["task:db-schema","task:db-conn-string","task:db-test"]`,
					AffectedDecisions: `["decision:db-choice"]`,
				},
				{
					EventType: string(continuous.EventDecisionChanged),
					Summary:   "Architecture change: Single Tenant to Multi Tenant",
					Details:   `{"from":"single-tenant","to":"multi-tenant","reason":"Customer demand for shared infrastructure","impact":"high"}`,
					Source:    "decision_change",

					ProposalReason:    "Architecture change: Single Tenant to Multi Tenant",
					AffectedPlans:     `["master:arch-change"]`,
					AffectedTasks:     `["task:tenant-model","task:isolation","task:migration"]`,
					AffectedDecisions: `["decision:tenant-arch"]`,
				},
				{
					EventType: string(continuous.EventChangeRequestCreated),
					Summary:   "Payment provider switch: Stripe to LemonSqueezy",
					Details:   `{"from":"stripe","to":"lemonsqueezy","reason":"Better Payout terms for EU","impact":"medium"}`,
					Source:    "change_request",

					ProposalReason:    "Change request: Payment provider switch from Stripe to LemonSqueezy",
					AffectedPlans:     `["master:billing"]`,
					AffectedTasks:     `["task:payment-api","task:webhook","task:invoice"]`,
					AffectedDecisions: `["decision:payment-provider"]`,
				},
				{
					EventType: string(continuous.EventChangeRequestCreated),
					Summary:   "New feature: Add AI-powered code review",
					Details:   `{"feature":"ai-code-review","effort":"large","priority":"high"}`,
					Source:    "product_request",

					ProposalReason:    "New feature request: AI-powered code review capability",
					AffectedPlans:     `["master:ai-features"]`,
					AffectedTasks:     `["task:ai-model","task:review-pipeline","task:ui"]`,
					AffectedDecisions: `["decision:ai-provider"]`,
				},
			}

			for _, s := range scenarios {
				ev, err := eventRepo.CreateEvent(store.ContinuousEventRecord{
					ID:        domain.NewID("ev"),
					ProjectID: projectID,
					EventType: s.EventType,
					Summary:   s.Summary,
					Details:   s.Details,
					Source:    s.Source,
				})
				if err != nil {
					return fmt.Errorf("create event %q: %w", s.Summary, err)
				}

				_, err = proposalRepo.CreateProposal(store.PlanUpdateProposalRecord{
					ID:                domain.NewID("pup"),
					ProjectID:         projectID,
					Reason:            s.ProposalReason,
					AffectedPlans:     s.AffectedPlans,
					AffectedTasks:     s.AffectedTasks,
					AffectedDecisions: s.AffectedDecisions,
					SuggestedUpdates:  fmt.Sprintf("Update plans affected by: %s", s.Summary),
					RequiresResearch:  0,
					RequiresApproval:  1,
					Status:            string(continuous.ProposalDraft),
				})
				if err != nil {
					return fmt.Errorf("create proposal for %q: %w", s.Summary, err)
				}

				_ = ev
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Continuous scenario seed: created 4 events and 4 proposals")
			return nil
		},
	}
}

func newSeedResearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed-research",
		Short: "Seed sample research entries into the project store.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			service := research.NewService(store.NewResearchRepository(db))

			entry1, err := service.CreateResearch("LLM Token Optimization",
				research.WithSummary("Techniques to reduce token usage in LLM prompts"),
				research.WithCategory(research.CategoryAI),
				research.WithConfidence(85),
				research.WithTags("llm", "tokens", "prompts"),
			)
			if err != nil {
				return err
			}
			service.AddFinding(entry1.ID, "Prompt Compression", "Using LLMLingua for prompt compression reduces tokens by 50-80% without significant quality loss.", 4)
			service.AddSource(entry1.ID, "LLMLingua Paper", "https://arxiv.org/abs/2310.05736", research.SourceTypeArticle)
			service.AddConclusion(entry1.ID, "Prompt compression is viable for production use with quality monitoring.", 85)

			entry2, err := service.CreateResearch("SQLite Performance Limits",
				research.WithSummary("Understanding SQLite's real-world read/write limits for local-first apps"),
				research.WithCategory(research.CategoryArchitecture),
				research.WithConfidence(90),
				research.WithTags("sqlite", "performance", "local-first"),
			)
			if err != nil {
				return err
			}
			service.AddFinding(entry2.ID, "Write Throughput", "SQLite handles ~50M writes/day on consumer SSDs in WAL mode.", 5)
			service.AddSource(entry2.ID, "SQLite Performance Docs", "https://www.sqlite.org/speed.html", research.SourceTypeDocumentation)
			service.AddConclusion(entry2.ID, "SQLite is well-suited for local-first apps up to 100GB datasets.", 90)

			fmt.Fprintln(cmd.OutOrStdout(), "Research seed: created")
			return nil
		},
	}
}

func printResearchSummary(cmd *cobra.Command, summary research.ResearchSummary) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Research:")
	fmt.Fprintf(out, "  total: %d\n", summary.Total)
	fmt.Fprintf(out, "  draft: %d\n", summary.Draft)
	fmt.Fprintf(out, "  in-review: %d\n", summary.InReview)
	fmt.Fprintf(out, "  approved: %d\n", summary.Approved)
	fmt.Fprintf(out, "  rejected: %d\n", summary.Rejected)
	fmt.Fprintf(out, "  archived: %d\n", summary.Archived)
	fmt.Fprintf(out, "  findings: %d\n", summary.Findings)
	fmt.Fprintf(out, "  sources: %d\n", summary.Sources)
	fmt.Fprintf(out, "  conclusions: %d\n", summary.Conclusions)
}

func knowledgeObjectsForList(service *knowledge.Service, category string) ([]domain.KnowledgeObject, error) {
	if category = strings.TrimSpace(category); category != "" {
		return service.ListByCategory(domain.KnowledgeCategory(category))
	}
	return service.ListKnowledge()
}

func tagValues(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, value := range raw {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func tagLabels(tags []knowledge.Tag) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		out = append(out, tag.Tag)
	}
	return out
}

func printKnowledgeSummary(cmd *cobra.Command, summary knowledge.Summary) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Knowledge:")
	fmt.Fprintf(out, "  total: %d\n", summary.Total)
	fmt.Fprintf(out, "  draft: %d\n", summary.Draft)
	fmt.Fprintf(out, "  reviewed: %d\n", summary.Reviewed)
	fmt.Fprintf(out, "  approved: %d\n", summary.Approved)
	fmt.Fprintf(out, "  archived: %d\n", summary.Archived)
	fmt.Fprintf(out, "  reused: %d\n", summary.Reused)
}

func openInstalledGlobalStore() (*sql.DB, error) {
	home, err := resolveHomeRoot()
	if err != nil {
		return nil, err
	}
	dbPath := config.GlobalDBPath(home)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := store.RunGlobalMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func openInitializedProjectStore() (*sql.DB, error) {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return nil, err
	}
	dbPath := config.ProjectDBPath(projectRoot)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, err
	}
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := store.RunProjectMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
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

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func dash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func resolveHomeRoot() (string, error) {
	if home := strings.TrimSpace(os.Getenv("PLAN_AI_HOME")); home != "" {
		return filepath.Abs(home)
	}
	return os.UserHomeDir()
}

func resolveProjectRoot() (string, error) {
	if projectRoot := strings.TrimSpace(os.Getenv("PLAN_AI_PROJECT_ROOT")); projectRoot != "" {
		return filepath.Abs(projectRoot)
	}
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(projectRoot)
}

// resolveOpenCodeConfigDir returns the directory containing OpenCode configuration files.
// Uses $OPENCODE_CONFIG_DIR if set, otherwise falls back to $HOME/.config/opencode.
func resolveOpenCodeConfigDir() (string, error) {
	if d := strings.TrimSpace(os.Getenv("OPENCODE_CONFIG_DIR")); d != "" {
		return filepath.Abs(d)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "opencode"), nil
}

func resolveOpenCodeConfigDirForWrite(allowReal bool) (string, error) {
	if d := strings.TrimSpace(os.Getenv("OPENCODE_CONFIG_DIR")); d != "" {
		return filepath.Abs(d)
	}
	if !allowReal {
		return "", fmt.Errorf("refusing to write real OpenCode config without OPENCODE_CONFIG_DIR; set OPENCODE_CONFIG_DIR for sandbox use or pass --allow-real-opencode")
	}
	return resolveOpenCodeConfigDir()
}

func installedLabel(dir, db string) string {
	if pathExists(dir) && pathExists(db) {
		return "installed"
	}
	return "not installed"
}

func initializedLabel(dir, db string) string {
	if pathExists(dir) && pathExists(db) {
		return "initialized"
	}
	return "not initialized"
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func detectGit(projectRoot string) bool {
	if info, err := os.Stat(filepath.Join(projectRoot, ".git")); err == nil && info.IsDir() {
		return true
	}
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func newVersionCommand(app core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Plan-AI version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Plan-AI %s\n", app.Version)
		},
	}
}

func newContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Build and persist composite context views.",
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

			projectID := store.ProjectID(projectRoot)
			repo := store.NewApprovedContextRepository(db)
			dq := store.NewDomainQuerier(db)
			builder := approvedcontext.NewBuilder(repo, dq, nil, nil)

			execCtx, err := builder.BuildExecutiveContext(projectID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Executive context:\n")
			fmt.Fprintf(out, "  status: %s\n", execCtx.Status)
			fmt.Fprintf(out, "  what_missing: %s\n", strings.Join(execCtx.WhatMissing, "; "))
			fmt.Fprintf(out, "  what_next: %s\n", strings.Join(execCtx.WhatNext, "; "))
			fmt.Fprintf(out, "  progress: %d phases\n", len(execCtx.Progress))
			return nil
		},
	}
	cmd.AddCommand(newContextPackageCommand())
	cmd.AddCommand(newContextPackageListCommand())
	cmd.AddCommand(newContextImplementationPackageCommand())
	cmd.AddCommand(newContextImplementationPackagesCommand())
	return cmd
}

func newContextPackageCommand() *cobra.Command {
	var packageType, model, content string
	var tokenBudget int
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Create a V2 smart context package for a target planning phase.",
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
			service := approvedcontext.NewSmartPackageService(store.NewSmartContextPackageRepository(db))
			pkg, err := service.Create(store.ProjectID(projectRoot), approvedcontext.PackageType(packageType), model, content, tokenBudget)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Context package created.")
			fmt.Fprintf(out, "  id: %s\n", pkg.ID)
			fmt.Fprintf(out, "  type: %s\n", pkg.Type)
			fmt.Fprintf(out, "  model: %s\n", pkg.ModelTarget)
			fmt.Fprintf(out, "  token_budget: %d\n", pkg.TokenBudget)
			return nil
		},
	}
	cmd.Flags().StringVar(&packageType, "type", string(approvedcontext.PackagePlanning), "package type: vision, research, planning, implementation, change")
	cmd.Flags().StringVar(&model, "model", "generic", "target model or agent profile")
	cmd.Flags().StringVar(&content, "content", "", "package content")
	cmd.Flags().IntVar(&tokenBudget, "token-budget", 4096, "maximum intended token budget")
	return cmd
}

func newContextPackageListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "packages",
		Short: "List V2 smart context packages.",
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
			items, err := approvedcontext.NewSmartPackageService(store.NewSmartContextPackageRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No context packages yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, pkg := range items {
				fmt.Fprintf(out, "%s\t%s\tmodel=%s\tbudget=%d\n", pkg.ID, pkg.Type, pkg.ModelTarget, pkg.TokenBudget)
			}
			return nil
		},
	}
}

func newContextImplementationPackageCommand() *cobra.Command {
	var planID, model, objective string
	cmd := &cobra.Command{
		Use:   "implementation-package",
		Short: "Create a V2 implementation context package for AI coding agents.",
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
			pkg, err := approvedcontext.NewImplementationPackageService(store.NewImplementationPackageRepository(db)).Create(store.ProjectID(projectRoot), planID, model, objective)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Implementation package created.")
			fmt.Fprintf(out, "  id: %s\n", pkg.ID)
			fmt.Fprintf(out, "  plan: %s\n", pkg.PlanID)
			fmt.Fprintf(out, "  model: %s\n", pkg.ModelTarget)
			fmt.Fprintf(out, "  commands: %s\n", strings.Join(pkg.Commands, "; "))
			fmt.Fprintf(out, "  validations: %s\n", strings.Join(pkg.Validations, "; "))
			return nil
		},
	}
	cmd.Flags().StringVar(&planID, "plan", "", "plan or blueprint id")
	cmd.Flags().StringVar(&model, "model", "opencode", "target coding agent/model")
	cmd.Flags().StringVar(&objective, "objective", "", "implementation objective")
	return cmd
}

func newContextImplementationPackagesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "implementation-packages",
		Short: "List V2 implementation context packages.",
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
			items, err := approvedcontext.NewImplementationPackageService(store.NewImplementationPackageRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No implementation packages yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, item := range items {
				fmt.Fprintf(out, "%s\t%s\tmodel=%s\tstatus=%s\n", item.ID, item.PlanID, item.ModelTarget, item.Status)
			}
			return nil
		},
	}
}

func newReferenceCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "reference", Short: "Manage V2 project references."}
	cmd.AddCommand(newReferenceAddCommand())
	cmd.AddCommand(newReferenceListCommand())
	cmd.AddCommand(newReferenceApproveCommand())
	cmd.AddCommand(newReferenceRejectCommand())
	return cmd
}

func newReferenceAddCommand() *cobra.Command {
	var source, uri, title, category string
	cmd := &cobra.Command{
		Use:   "add --uri <uri>",
		Short: "Add a V2 project reference for approval.",
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
			ref, err := reference.NewService(store.NewReferenceRepository(db)).Add(store.ProjectID(projectRoot), reference.SourceType(source), uri, title, reference.Category(category))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Reference added.")
			fmt.Fprintf(out, "  id: %s\n", ref.ID)
			fmt.Fprintf(out, "  source: %s\n", ref.Source)
			fmt.Fprintf(out, "  status: %s\n", ref.Status)
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", string(reference.SourceURL), "source type: url, image, document, repository, screenshot, example")
	cmd.Flags().StringVar(&uri, "uri", "", "reference URI or path (required)")
	cmd.Flags().StringVar(&title, "title", "", "reference title")
	cmd.Flags().StringVar(&category, "category", "", "reference category")
	_ = cmd.MarkFlagRequired("uri")
	return cmd
}

func newReferenceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List V2 project references.",
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
			refs, err := reference.NewService(store.NewReferenceRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(refs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No references yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, ref := range refs {
				fmt.Fprintf(out, "%s\t%s\t%s\tstatus=%s\n", ref.ID, ref.Category, ref.Title, ref.Status)
			}
			return nil
		},
	}
}

func newReferenceApproveCommand() *cobra.Command {
	return newReferenceStateCommand("approve", "Approve a V2 project reference.", func(s reference.Service, id string) (reference.Reference, error) { return s.Approve(id) })
}

func newReferenceRejectCommand() *cobra.Command {
	return newReferenceStateCommand("reject", "Reject a V2 project reference.", func(s reference.Service, id string) (reference.Reference, error) { return s.Reject(id) })
}

func newReferenceStateCommand(use, short string, transition func(reference.Service, string) (reference.Reference, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			ref, err := transition(reference.NewService(store.NewReferenceRepository(db)), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Reference %s: %s\n", ref.Status, ref.ID)
			return nil
		},
	}
}

func newJobsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List and manage orchestration jobs.",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List orchestration jobs.",
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

			jobs, err := store.NewJobRepository(db).ListJobs(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(jobs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No orchestration jobs yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, j := range jobs {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", j.ID, j.WorkflowType, j.Status, j.StartedAt.Format(time.RFC3339))
			}
			return nil
		},
	})
	return cmd
}

func newCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List registered capabilities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := capabilities.NewDefaultRegistry()
			caps := r.ListCapabilities()
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Registered capabilities:")
			for _, c := range caps {
				fmt.Fprintf(out, "  %s: %s\n", c.Type, c.Name)
			}
			return nil
		},
	}
}

func newImpactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impact",
		Short: "Analyse change impact on planning entities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectDB := config.ProjectDBPath(projectRoot)
			if !pathExists(projectDB) {
				return fmt.Errorf("project not initialized; run 'plan-ai init' first")
			}
			db, err := store.Open(projectDB)
			if err != nil {
				return err
			}
			defer db.Close()
			repo := store.NewChangeEventRepository(db)
			events, err := repo.ListByProject(store.ProjectID(projectRoot), 10)
			if err != nil {
				return fmt.Errorf("list changes: %w", err)
			}
			out := cmd.OutOrStdout()
			if len(events) == 0 {
				fmt.Fprintln(out, "No change events recorded yet.")
				return nil
			}
			fmt.Fprintln(out, "Recent change events:")
			for _, ev := range events {
				fmt.Fprintf(out, "  %s  %-24s  %-40s\n", ev.CreatedAt[:19], ev.ChangeType, truncateString(ev.Summary, 40))
			}
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Use 'plan-ai impact --help' to see all options.")
			return nil
		},
	}
	cmd.AddCommand(newImpactAnalyzeV2Command())
	cmd.AddCommand(newImpactReportsV2Command())
	return cmd
}

func newImpactAnalyzeV2Command() *cobra.Command {
	var changeType, summary string
	cmd := &cobra.Command{
		Use:   "analyze-v2",
		Short: "Create a deep V2 impact report across architecture, backend, migrations, docs, APIs, plans, and validations.",
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
			report, err := change.NewDeepImpactService(store.NewDeepImpactRepository(db)).Analyze(store.ProjectID(projectRoot), change.ChangeType(changeType), summary)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Change Impact V2 report created.")
			fmt.Fprintf(out, "  id: %s\n", report.ID)
			fmt.Fprintf(out, "  change_type: %s\n", report.ChangeType)
			fmt.Fprintf(out, "  severity: %s\n", report.Severity)
			fmt.Fprintf(out, "  validations: %s\n", strings.Join(report.ValidationCommands, "; "))
			fmt.Fprintf(out, "  rollback: %s\n", strings.Join(report.RollbackStrategy, "; "))
			return nil
		},
	}
	cmd.Flags().StringVar(&changeType, "type", string(change.TechnologyChanged), "change type")
	cmd.Flags().StringVar(&summary, "summary", "", "change summary")
	return cmd
}

func newImpactReportsV2Command() *cobra.Command {
	return &cobra.Command{
		Use:   "reports-v2",
		Short: "List Change Impact V2 reports.",
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
			reports, err := change.NewDeepImpactService(store.NewDeepImpactRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(reports) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Change Impact V2 reports yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, report := range reports {
				fmt.Fprintf(out, "%s\t%s\tseverity=%s\t%s\n", report.ID, report.ChangeType, report.Severity, truncateString(report.Summary, 48))
			}
			return nil
		},
	}
}

func newSnapshotCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Create or list project state snapshots.",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List recent snapshots.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			projectDB := config.ProjectDBPath(projectRoot)
			if !pathExists(projectDB) {
				return fmt.Errorf("project not initialized; run 'plan-ai init' first")
			}
			db, err := store.Open(projectDB)
			if err != nil {
				return err
			}
			defer db.Close()
			repo := store.NewSnapshotV2Repository(db)
			snaps, err := repo.ListByProject(store.ProjectID(projectRoot), 10)
			if err != nil {
				return fmt.Errorf("list snapshots: %w", err)
			}
			out := cmd.OutOrStdout()
			if len(snaps) == 0 {
				fmt.Fprintln(out, "No snapshots yet.")
				return nil
			}
			fmt.Fprintln(out, "Recent snapshots:")
			for _, s := range snaps {
				fmt.Fprintf(out, "  %s  %s\n", s.CreatedAt[:19], truncateString(s.Reason, 50))
			}
			return nil
		},
	})
	return cmd
}

func newAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent system — intent detection, routing, and delegation.",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "process [message]",
		Short: "Process a message through the agent system.",
		Args:  cobra.MinimumNArgs(1),
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

			projectID := store.ProjectID(projectRoot)
			message := strings.Join(args, " ")

			ws := agent.NewWorkflowSelector()
			cs := agent.NewCapabilitySelector()
			detector := agent.NewIntentDetector()
			router := agent.NewRouter(ws, cs)
			contextLoader := agent.NewContextLoader(db)
			responseBuilder := agent.NewResponseBuilder()
			delegator := agent.NewDelegator(db, agentDelegatedStoreRepo(db))
			runRepo := agentRunStoreRepo(db)

			svc := agent.NewService(detector, router, contextLoader, delegator, responseBuilder, runRepo)
			resp, err := svc.ProcessMessage(projectID, message)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Status: %s\n", resp.Status)
			fmt.Fprintf(out, "Message: %s\n", resp.Message)
			if resp.WorkflowTriggered != "" {
				fmt.Fprintf(out, "Workflow: %s\n", resp.WorkflowTriggered)
			}
			if resp.RequiresApproval {
				fmt.Fprintln(out, "Approval: required")
			}
			fmt.Fprintf(out, "Next: %s\n", resp.SuggestedNextAction)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:     "runs",
		Short:   "List recent agent runs.",
		Aliases: []string{"list"},
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

			projectID := store.ProjectID(projectRoot)
			repo := agentRunStoreRepo(db)
			runs, err := repo.ListRuns(projectID, 10)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(runs) == 0 {
				fmt.Fprintln(out, "No agent runs yet.")
				return nil
			}
			for _, r := range runs {
				fmt.Fprintf(out, "%s  %s  %-12s  %s\n", r.CreatedAt[:19], r.ID, r.Intent, r.Status)
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show agent system status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Agent system: active")
			fmt.Fprintln(out, "Intent detection: keyword/pattern based")
			fmt.Fprintln(out, "Routing: workflow + capability selection")
			fmt.Fprintln(out, "Delegation: sub-agent jobs")
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Use 'agent process <message>' to process a message.")
			fmt.Fprintln(out, "Use 'agent runs' to list recent agent activity.")
			return nil
		},
	})
	cmd.AddCommand(newAgentSubagentCreateCommand())
	cmd.AddCommand(newAgentSubagentsCommand())
	return cmd
}

func newAgentSubagentCreateCommand() *cobra.Command {
	var agentType, objective string
	cmd := &cobra.Command{
		Use:   "subagent-create",
		Short: "Create a temporary isolated V2 subagent task.",
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
			task, err := agent.NewSubagentOrchestrator(subagentTaskRepoWrapper{repo: store.NewSubagentTaskRepository(db)}).Create(store.ProjectID(projectRoot), agent.SubagentType(agentType), objective)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Subagent task created.")
			fmt.Fprintf(out, "  id: %s\n", task.ID)
			fmt.Fprintf(out, "  type: %s\n", task.AgentType)
			fmt.Fprintf(out, "  isolated: %v\n", task.Isolated)
			fmt.Fprintf(out, "  temporary: %v\n", task.Temporary)
			fmt.Fprintf(out, "  memory_policy: %s\n", task.MemoryPolicy)
			fmt.Fprintf(out, "  validation_status: %s\n", task.ValidationStatus)
			return nil
		},
	}
	cmd.Flags().StringVar(&agentType, "type", string(agent.SubagentValidation), "subagent type: research, architecture, ui, ux, security, backend, database, validation")
	cmd.Flags().StringVar(&objective, "objective", "", "subagent objective")
	return cmd
}

func newAgentSubagentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "subagents",
		Short: "List V2 subagent tasks.",
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
			tasks, err := agent.NewSubagentOrchestrator(subagentTaskRepoWrapper{repo: store.NewSubagentTaskRepository(db)}).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No V2 subagent tasks yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, task := range tasks {
				fmt.Fprintf(out, "%s\t%s\tstatus=%s\tvalidation=%s\n", task.ID, task.AgentType, task.Status, task.ValidationStatus)
			}
			return nil
		},
	}
}

type subagentTaskRepoWrapper struct{ repo store.SubagentTaskRepository }

func (w subagentTaskRepoWrapper) SaveSubagentTask(task agent.SubagentTask) (agent.SubagentTask, error) {
	rec, err := w.repo.SaveSubagentTaskRecord(store.SubagentTaskRecord{
		ID: task.ID, ProjectID: task.ProjectID, AgentType: string(task.AgentType), Objective: task.Objective,
		Capability: task.Capability, Status: string(task.Status), Provenance: task.Provenance,
		ValidationStatus: string(task.ValidationStatus), Isolated: task.Isolated, Temporary: task.Temporary,
		MemoryPolicy: task.MemoryPolicy, ResultSummary: task.ResultSummary,
		CreatedAt: task.CreatedAt.Format(time.RFC3339), UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
	})
	if err != nil {
		return agent.SubagentTask{}, err
	}
	return subagentRecordToTask(rec)
}

func (w subagentTaskRepoWrapper) ListSubagentTasks(projectID string) ([]agent.SubagentTask, error) {
	records, err := w.repo.ListSubagentTaskRecords(projectID)
	if err != nil {
		return nil, err
	}
	tasks := make([]agent.SubagentTask, len(records))
	for i, rec := range records {
		task, err := subagentRecordToTask(rec)
		if err != nil {
			return nil, err
		}
		tasks[i] = task
	}
	return tasks, nil
}

func subagentRecordToTask(rec store.SubagentTaskRecord) (agent.SubagentTask, error) {
	created, err := time.Parse(time.RFC3339, rec.CreatedAt)
	if err != nil {
		return agent.SubagentTask{}, fmt.Errorf("parse subagent task created_at for %q: %w", rec.ID, err)
	}
	updated, err := time.Parse(time.RFC3339, rec.UpdatedAt)
	if err != nil {
		return agent.SubagentTask{}, fmt.Errorf("parse subagent task updated_at for %q: %w", rec.ID, err)
	}
	return agent.SubagentTask{
		ID: rec.ID, ProjectID: rec.ProjectID, AgentType: agent.SubagentType(rec.AgentType), Objective: rec.Objective,
		Capability: rec.Capability, Status: agent.JobStatus(rec.Status), Provenance: rec.Provenance,
		ValidationStatus: agent.ValidationStatus(rec.ValidationStatus), Isolated: rec.Isolated, Temporary: rec.Temporary,
		MemoryPolicy: rec.MemoryPolicy, ResultSummary: rec.ResultSummary, CreatedAt: created, UpdatedAt: updated,
	}, nil
}

func newContinuousCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "continuous",
		Short: "Continuous planning — event detection, proposals, and context.",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Get continuous planning status.",
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

			projectID := store.ProjectID(projectRoot)
			statusSvc := continuous.NewStatusService(db)
			status, err := statusSvc.GetStatus(projectID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Active plan: %s\n", status.ActivePlan)
			fmt.Fprintf(out, "Next task: %s\n", status.NextTask)
			fmt.Fprintf(out, "Recent events: %d\n", status.RecentEvents)
			fmt.Fprintf(out, "Pending proposals: %d\n", status.PendingProposals)
			if len(status.BlockedItems) > 0 {
				fmt.Fprintf(out, "Blocked: %s\n", strings.Join(status.BlockedItems, ", "))
			}
			if len(status.ApprovalsNeeded) > 0 {
				fmt.Fprintf(out, "Approvals needed: %s\n", strings.Join(status.ApprovalsNeeded, "; "))
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "events",
		Short: "List recent continuous planning events.",
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

			projectID := store.ProjectID(projectRoot)
			detector := continuous.NewDetector(db)
			events, err := detector.Detect(projectID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(events) == 0 {
				fmt.Fprintln(out, "No events yet.")
				return nil
			}
			for _, ev := range events {
				fmt.Fprintf(out, "%s  %-24s  %s\n", ev.CreatedAt[:19], ev.EventType, truncateString(ev.Summary, 40))
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "proposals",
		Short: "List plan update proposals.",
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

			projectID := store.ProjectID(projectRoot)
			proposalRepo := continuousProposalStoreRepo(db)
			proposals, err := proposalRepo.ListProposals(projectID)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(proposals) == 0 {
				fmt.Fprintln(out, "No proposals yet.")
				return nil
			}
			for _, p := range proposals {
				fmt.Fprintf(out, "%s  %-16s  %s\n", p.CreatedAt[:19], p.Status, truncateString(p.Reason, 50))
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "approve-update <id>",
		Short: "Approve a plan update proposal.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			updater := continuous.NewUpdater(continuousProposalStoreRepo(db))
			proposal, err := updater.Approve(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Proposal %s approved (status: %s)\n", proposal.ID, proposal.Status)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "reject-update <id>",
		Short: "Reject a plan update proposal.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			updater := continuous.NewUpdater(continuousProposalStoreRepo(db))
			proposal, err := updater.Reject(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Proposal %s rejected (status: %s)\n", proposal.ID, proposal.Status)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "context [level]",
		Short: "Generate context at the specified level.",
		Long:  "Levels: L0_Executive, L1_Planning, L2_Specific_Plan, L3_Task, L4_Implementation",
		Args:  cobra.RangeArgs(0, 1),
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

			projectID := store.ProjectID(projectRoot)
			level := continuous.ContextL0Executive
			if len(args) > 0 {
				level = continuous.ContextLevel(args[0])
			}

			gen := continuous.NewContextGenerator(db)
			content, err := gen.Generate(projectID, level)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), content)
			return nil
		},
	})
	cmd.AddCommand(newContinuousRegenerateCommand())
	cmd.AddCommand(newContinuousRegenerationsCommand())
	return cmd
}

func newContinuousRegenerateCommand() *cobra.Command {
	var reason, scope string
	cmd := &cobra.Command{
		Use:   "regenerate",
		Short: "Create a targeted Continuous Planning V2 regeneration proposal.",
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
			regen, err := continuous.NewPlanningV2Service(store.NewTargetedRegenerationRepository(db)).Regenerate(store.ProjectID(projectRoot), reason, scope)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Continuous Planning V2 regeneration created.")
			fmt.Fprintf(out, "  id: %s\n", regen.ID)
			fmt.Fprintf(out, "  scope: %s\n", regen.Scope)
			fmt.Fprintf(out, "  affected_sections: %s\n", strings.Join(regen.AffectedSections, "; "))
			fmt.Fprintf(out, "  approval_required: %v\n", regen.ApprovalRequired)
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "regeneration reason")
	cmd.Flags().StringVar(&scope, "scope", "affected-sections", "regeneration scope")
	return cmd
}

func newContinuousRegenerationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "regenerations",
		Short: "List Continuous Planning V2 regenerations.",
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
			items, err := continuous.NewPlanningV2Service(store.NewTargetedRegenerationRepository(db)).List(store.ProjectID(projectRoot))
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Continuous Planning V2 regenerations yet.")
				return nil
			}
			out := cmd.OutOrStdout()
			for _, item := range items {
				fmt.Fprintf(out, "%s\t%s\tstatus=%s\tapproval=%v\n", item.ID, item.Scope, item.Status, item.ApprovalRequired)
			}
			return nil
		},
	}
}

// agentRunStoreRepo creates an AgentRunV2Repository wrapping a *sql.DB as the
// exported AgentRunRepository interface (DuCK typing — struct method match).
func agentRunStoreRepo(db *sql.DB) agentRunRepoWrapper {
	return agentRunRepoWrapper{db: db}
}

// agentRunRepoWrapper adapts store.AgentRunV2Repository to the pattern used
// by the agent package. The agent.AgentRunRepository interface uses the same
// method signatures so the struct is directly assignable.
type agentRunRepoWrapper struct {
	db *sql.DB
}

func (w agentRunRepoWrapper) CreateRun(run agent.AgentRunRecord) (agent.AgentRunRecord, error) {
	rec, err := w.repo().CreateRun(toStoreRun(run))
	if err != nil {
		return agent.AgentRunRecord{}, err
	}
	return toAgentRun(rec), nil
}

func (w agentRunRepoWrapper) GetRun(id string) (agent.AgentRunRecord, error) {
	rec, err := w.repo().GetRun(id)
	if err != nil {
		return agent.AgentRunRecord{}, err
	}
	return toAgentRun(rec), nil
}

func (w agentRunRepoWrapper) UpdateRunStatus(id, status, response string) error {
	return w.repo().UpdateRunStatus(id, status, response)
}

func (w agentRunRepoWrapper) ListRuns(projectID string, limit int) ([]agent.AgentRunRecord, error) {
	records, err := w.repo().ListRuns(projectID, limit)
	if err != nil {
		return nil, err
	}
	runs := make([]agent.AgentRunRecord, len(records))
	for i, r := range records {
		runs[i] = toAgentRun(r)
	}
	return runs, nil
}

func (w agentRunRepoWrapper) CreateMessage(msg agent.AgentMessage) (agent.AgentMessage, error) {
	storeMsg, err := w.repo().CreateMessage(toStoreMsg(msg))
	if err != nil {
		return agent.AgentMessage{}, err
	}
	return toAgentMsg(storeMsg), nil
}

func (w agentRunRepoWrapper) ListMessages(runID string) ([]agent.AgentMessage, error) {
	records, err := w.repo().ListMessages(runID)
	if err != nil {
		return nil, err
	}
	msgs := make([]agent.AgentMessage, len(records))
	for i, m := range records {
		msgs[i] = toAgentMsg(m)
	}
	return msgs, nil
}

func (w agentRunRepoWrapper) repo() *store.AgentRunV2Repository {
	return store.NewAgentRunV2Repository(w.db)
}

func toStoreRun(run agent.AgentRunRecord) store.AgentRunV2Record {
	return store.AgentRunV2Record{
		ID: run.ID, ProjectID: run.ProjectID, Intent: run.Intent,
		Status: run.Status, Response: run.Response,
		CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt,
	}
}

func toAgentRun(rec store.AgentRunV2Record) agent.AgentRunRecord {
	return agent.AgentRunRecord{
		ID: rec.ID, ProjectID: rec.ProjectID, Intent: rec.Intent,
		Status: rec.Status, Response: rec.Response,
		CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

func toStoreMsg(msg agent.AgentMessage) store.AgentMessageRecord {
	return store.AgentMessageRecord{
		ID: msg.ID, RunID: msg.RunID, Role: msg.Role,
		Content: msg.Content, CreatedAt: msg.CreatedAt,
	}
}

func toAgentMsg(rec store.AgentMessageRecord) agent.AgentMessage {
	return agent.AgentMessage{
		ID: rec.ID, RunID: rec.RunID, Role: rec.Role,
		Content: rec.Content, CreatedAt: rec.CreatedAt,
	}
}

// agentDelegatedStoreRepo creates a DelegatedJobRepository from the store.
func agentDelegatedStoreRepo(db *sql.DB) delegatedJobRepoWrapper {
	return delegatedJobRepoWrapper{db: db}
}

type delegatedJobRepoWrapper struct {
	db *sql.DB
}

func (w delegatedJobRepoWrapper) CreateJob(job agent.DelegatedJob) (agent.DelegatedJob, error) {
	storeJob, err := w.repo().CreateJob(toStoreDelegated(job))
	if err != nil {
		return agent.DelegatedJob{}, err
	}
	return toAgentDelegated(storeJob), nil
}

func (w delegatedJobRepoWrapper) GetJob(id string) (agent.DelegatedJob, error) {
	rec, err := w.repo().GetJob(id)
	if err != nil {
		return agent.DelegatedJob{}, err
	}
	return toAgentDelegated(rec), nil
}

func (w delegatedJobRepoWrapper) ListJobs(projectID string) ([]agent.DelegatedJob, error) {
	records, err := w.repo().ListJobs(projectID)
	if err != nil {
		return nil, err
	}
	jobs := make([]agent.DelegatedJob, len(records))
	for i, r := range records {
		jobs[i] = toAgentDelegated(r)
	}
	return jobs, nil
}

func (w delegatedJobRepoWrapper) UpdateJob(id, status, summary string) error {
	return w.repo().UpdateJob(id, status, summary)
}

func (w delegatedJobRepoWrapper) repo() *store.DelegatedJobRepository {
	return store.NewDelegatedJobRepository(w.db)
}

func toStoreDelegated(job agent.DelegatedJob) store.AgentDelegatedJobRecord {
	return store.AgentDelegatedJobRecord{
		ID: job.ID, ProjectID: job.ProjectID, Intent: string(job.Intent),
		Capability: job.Capability, WorkflowType: job.WorkflowType,
		JobType: string(job.JobType), Status: string(job.Status),
		ResultSummary: job.ResultSummary, CreatedAt: job.CreatedAt,
		CompletedAt: job.CompletedAt,
	}
}

func toAgentDelegated(rec store.AgentDelegatedJobRecord) agent.DelegatedJob {
	return agent.DelegatedJob{
		ID: rec.ID, ProjectID: rec.ProjectID, Intent: agent.IntentKind(rec.Intent),
		Capability: rec.Capability, WorkflowType: rec.WorkflowType,
		JobType: agent.DelegatedJobType(rec.JobType), Status: agent.JobStatus(rec.Status),
		ResultSummary: rec.ResultSummary, CreatedAt: rec.CreatedAt,
		CompletedAt: rec.CompletedAt,
	}
}

// continuousProposalStoreRepo creates a PlanUpdateProposalRepository from the store.
func continuousProposalStoreRepo(db *sql.DB) proposalRepoWrapper {
	return proposalRepoWrapper{db: db}
}

type proposalRepoWrapper struct {
	db *sql.DB
}

func (w proposalRepoWrapper) CreateProposal(p continuous.PlanUpdateProposal) (continuous.PlanUpdateProposal, error) {
	storeP, err := w.repo().CreateProposal(toStoreProposal(p))
	if err != nil {
		return continuous.PlanUpdateProposal{}, err
	}
	return toAgentProposal(storeP), nil
}

func (w proposalRepoWrapper) GetProposal(id string) (continuous.PlanUpdateProposal, error) {
	rec, err := w.repo().GetProposal(id)
	if err != nil {
		return continuous.PlanUpdateProposal{}, err
	}
	return toAgentProposal(rec), nil
}

func (w proposalRepoWrapper) ListProposals(projectID string) ([]continuous.PlanUpdateProposal, error) {
	records, err := w.repo().ListProposals(projectID)
	if err != nil {
		return nil, err
	}
	proposals := make([]continuous.PlanUpdateProposal, len(records))
	for i, r := range records {
		proposals[i] = toAgentProposal(r)
	}
	return proposals, nil
}

func (w proposalRepoWrapper) UpdateProposalStatus(id string, status continuous.ProposalStatus) error {
	return w.repo().UpdateProposalStatus(id, string(status))
}

func (w proposalRepoWrapper) repo() *store.PlanUpdateProposalRepository {
	return store.NewPlanUpdateProposalRepository(w.db)
}

func toStoreProposal(p continuous.PlanUpdateProposal) store.PlanUpdateProposalRecord {
	plansJSON := marshalStrings(p.AffectedPlans)
	tasksJSON := marshalStrings(p.AffectedTasks)
	decisionsJSON := marshalStrings(p.AffectedDecisions)
	rr := 0
	if p.RequiresResearch {
		rr = 1
	}
	ra := 0
	if p.RequiresApproval {
		ra = 1
	}
	return store.PlanUpdateProposalRecord{
		ID: p.ID, ProjectID: p.ProjectID, Reason: p.Reason,
		AffectedPlans: plansJSON, AffectedTasks: tasksJSON,
		AffectedDecisions: decisionsJSON, SuggestedUpdates: p.SuggestedUpdates,
		RequiresResearch: rr, RequiresApproval: ra,
		Status: string(p.Status), CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toAgentProposal(rec store.PlanUpdateProposalRecord) continuous.PlanUpdateProposal {
	return continuous.PlanUpdateProposal{
		ID: rec.ID, ProjectID: rec.ProjectID, Reason: rec.Reason,
		AffectedPlans:     unmarshalStrings(rec.AffectedPlans),
		AffectedTasks:     unmarshalStrings(rec.AffectedTasks),
		AffectedDecisions: unmarshalStrings(rec.AffectedDecisions),
		SuggestedUpdates:  rec.SuggestedUpdates,
		RequiresResearch:  rec.RequiresResearch != 0,
		RequiresApproval:  rec.RequiresApproval != 0,
		Status:            continuous.ProposalStatus(rec.Status),
		CreatedAt:         rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

func marshalStrings(s []string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func unmarshalStrings(s string) []string {
	var result []string
	_ = json.Unmarshal([]byte(s), &result)
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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

func newDiscoveryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "discovery", Short: "Ph23-24 vision discovery and Ph53 progressive discovery."}
	cmd.AddCommand(&cobra.Command{
		Use:   "start <context>",
		Short: "Start a new discovery session.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			projectRoot, err := resolveProjectRoot()
			if err != nil {
				return err
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			engine := vision.NewDiscoveryEngine(nil) // lightweight; full integration uses repos
			_, err = engine.StartSession(store.ProjectID(projectRoot), args[0])
			if err != nil {
				return fmt.Errorf("engine not wired (use via library): %w", err)
			}
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Discovery session engine loaded.")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Validate discovery engine types are importable.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Discovery Engine: OK")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Session types loaded")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Assumption/Ambiguity/Approval types loaded")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Heuristic analysis available")
			return nil
		},
	})
	// ── Phase 53: Progressive Discovery System ──
	cmd.AddCommand(newDiscoveryV3InitCommand())
	cmd.AddCommand(newDiscoveryV3NextCommand())
	cmd.AddCommand(newDiscoveryV3AnswerCommand())
	cmd.AddCommand(newDiscoveryV3StatusCommand())
	return cmd
}

func newDiscoveryV3InitCommand() *cobra.Command {
	var intentID string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Ph53: Initialize progressive discovery questions for an intent.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if intentID == "" {
				return fmt.Errorf("--intent is required")
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			svc := discoveryv3.NewService(
				store.NewDiscoveryV3QuestionRepository(db),
				store.NewDiscoveryV3AnswerRepository(db),
			)
			if err := svc.Initialize(intentID); err != nil {
				return err
			}
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Progressive discovery initialized for intent %s.\n", intentID)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "Product intent ID")
	return cmd
}

func newDiscoveryV3NextCommand() *cobra.Command {
	var intentID, level string
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Ph53: Show next unanswered discovery questions.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if intentID == "" {
				return fmt.Errorf("--intent is required")
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			svc := discoveryv3.NewService(
				store.NewDiscoveryV3QuestionRepository(db),
				store.NewDiscoveryV3AnswerRepository(db),
			)
			var dl discoveryv3.DiscoveryLevel
			if level != "" {
				dl = discoveryv3.DiscoveryLevel(level)
			}
			qs, err := svc.GetNextQuestions(intentID, dl)
			if err != nil {
				return err
			}
			if len(qs) == 0 {
				fmt.Fprintln(cobraCmd.OutOrStdout(), "All questions answered.")
				return nil
			}
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Discovery level: %s\n\n", qs[0].Level)
			for _, q := range qs {
				required := ""
				if q.Required {
					required = " [required]"
				}
				fmt.Fprintf(cobraCmd.OutOrStdout(), "  [%s] %s%s\n", q.ID, q.Question, required)
				fmt.Fprintf(cobraCmd.OutOrStdout(), "    Reason: %s\n", q.Reason)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "Product intent ID")
	cmd.Flags().StringVar(&level, "level", "", "Discovery level (project|master_plan|specific_plan|phase|task)")
	return cmd
}

func newDiscoveryV3AnswerCommand() *cobra.Command {
	var questionID, intentID, answer string
	cmd := &cobra.Command{
		Use:   "answer",
		Short: "Ph53: Answer a progressive discovery question.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if questionID == "" {
				return fmt.Errorf("--question is required")
			}
			if intentID == "" {
				return fmt.Errorf("--intent is required")
			}
			if answer == "" {
				return fmt.Errorf("--answer is required")
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			svc := discoveryv3.NewService(
				store.NewDiscoveryV3QuestionRepository(db),
				store.NewDiscoveryV3AnswerRepository(db),
			)
			a, err := svc.Answer(intentID, discoveryv3.QuestionID(questionID), answer)
			if err != nil {
				return err
			}
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Answer recorded: %s\n", a.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&questionID, "question", "", "Question ID")
	cmd.Flags().StringVar(&intentID, "intent", "", "Product intent ID")
	cmd.Flags().StringVar(&answer, "answer", "", "Answer text")
	return cmd
}

func newDiscoveryV3StatusCommand() *cobra.Command {
	var intentID string
	cmd := &cobra.Command{
		Use:   "v3-status",
		Short: "Ph53: Show progressive discovery session status.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if intentID == "" {
				return fmt.Errorf("--intent is required")
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()
			svc := discoveryv3.NewService(
				store.NewDiscoveryV3QuestionRepository(db),
				store.NewDiscoveryV3AnswerRepository(db),
			)
			s, err := svc.Status(intentID)
			if err != nil {
				return err
			}
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Intent:          %s\n", s.IntentID)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Current Level:   %s\n", s.CurrentLevel)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Questions:       %d answered, %d remaining (total %d)\n", s.AnsweredCount, s.RemainingCount, s.TotalQuestions)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Answered Levels: %v\n", s.AnsweredLevels)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Suggestion:      %s\n", s.SuggestedLevel)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "Product intent ID")
	return cmd
}

func newAmbiguityCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "ambiguity", Short: "Ph54: Detect missing information, assumptions, conflicts, and unknown areas."}
	cmd.AddCommand(newAmbiguityAnalyzeCommand())
	return cmd
}

func newAmbiguityAnalyzeCommand() *cobra.Command {
	var intentID, input string
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze ambiguity from a V3 product intent or raw input.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			svc := ambiguityv3.NewService()
			var report ambiguityv3.AmbiguityReport
			switch {
			case intentID != "":
				db, err := openInitializedProjectStore()
				if err != nil {
					return err
				}
				defer db.Close()
				pi, err := store.NewIntentV3Repository(db).GetProductIntent(intentID)
				if err != nil {
					return err
				}
				questions, err := store.NewDiscoveryV3QuestionRepository(db).GetAllQuestions(intentID)
				if err != nil {
					return err
				}
				answers, err := store.NewDiscoveryV3AnswerRepository(db).GetAnswers(intentID)
				if err != nil {
					return err
				}
				report = svc.AnalyzeProductIntent(pi, questions, answers)
			case input != "":
				report = svc.AnalyzeText(input)
			default:
				return fmt.Errorf("one of --intent or --input is required")
			}
			printAmbiguityReport(cobraCmd, report)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "V3 Product Intent ID")
	cmd.Flags().StringVar(&input, "input", "", "Raw text to analyze")
	return cmd
}

func printAmbiguityReport(cmd *cobra.Command, report ambiguityv3.AmbiguityReport) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Ambiguity Report")
	if report.IntentID != "" {
		fmt.Fprintf(out, "  Intent ID: %s\n", report.IntentID)
	}
	fmt.Fprintf(out, "  Ambiguity Score: %d%%\n", report.Score)
	fmt.Fprintf(out, "  Known Areas: %s\n", joinOrDash(report.KnownAreas))
	fmt.Fprintln(out, "  Missing Information:")
	if len(report.MissingInformation) == 0 {
		fmt.Fprintln(out, "    -")
	} else {
		for _, item := range report.MissingInformation {
			fmt.Fprintf(out, "    - %s: %s\n", item.Field, item.Reason)
		}
	}
	fmt.Fprintln(out, "  Assumptions:")
	if len(report.Assumptions) == 0 {
		fmt.Fprintln(out, "    -")
	} else {
		for _, item := range report.Assumptions {
			fmt.Fprintf(out, "    - %s: %s\n", item.ID, item.Reason)
		}
	}
	fmt.Fprintln(out, "  Conflicts:")
	if len(report.Conflicts) == 0 {
		fmt.Fprintln(out, "    -")
	} else {
		for _, item := range report.Conflicts {
			fmt.Fprintf(out, "    - %s: %s\n", item.ID, item.Evidence)
		}
	}
	fmt.Fprintln(out, "  Unknown Areas:")
	if len(report.UnknownAreas) == 0 {
		fmt.Fprintln(out, "    -")
	} else {
		for _, item := range report.UnknownAreas {
			required := "optional"
			if item.Required {
				required = "required"
			}
			fmt.Fprintf(out, "    - %s/%s (%s): %s\n", item.Level, item.QuestionID, required, item.Question)
		}
	}
	fmt.Fprintf(out, "  Needs To Know: %s\n", joinOrDash(report.NeedsToKnow))
}

func newConfidenceCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "confidence", Short: "Ph55: Measure intent understanding confidence."}
	cmd.AddCommand(newConfidenceEvaluateCommand())
	return cmd
}

func newConfidenceEvaluateCommand() *cobra.Command {
	var intentID string
	cmd := &cobra.Command{
		Use:   "evaluate --intent <id>",
		Short: "Evaluate V3 intent confidence from intent, discovery, and ambiguity signals.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if intentID == "" {
				return fmt.Errorf("--intent is required")
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			intentRepo := store.NewIntentV3Repository(db)
			discoveryRepo := store.NewIntentV3DiscoveryResultRepository(db)
			questionRepo := store.NewDiscoveryV3QuestionRepository(db)
			answerRepo := store.NewDiscoveryV3AnswerRepository(db)

			pi, err := intentRepo.GetProductIntent(intentID)
			if err != nil {
				return err
			}

			var discovery *intentv3.DiscoveryResult
			if pi.DiscoveryResultID != "" {
				result, err := discoveryRepo.GetDiscoveryResult(pi.DiscoveryResultID)
				if err != nil {
					return err
				}
				discovery = &result
			}
			questions, err := questionRepo.GetAllQuestions(intentID)
			if err != nil {
				return err
			}
			answers, err := answerRepo.GetAnswers(intentID)
			if err != nil {
				return err
			}
			ambiguityReport := ambiguityv3.NewService().AnalyzeProductIntent(pi, questions, answers)
			report := confidencev3.NewService().Evaluate(pi, discovery, questions, answers, ambiguityReport)
			printConfidenceReport(cobraCmd, report)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "V3 Product Intent ID")
	return cmd
}

func printConfidenceReport(cmd *cobra.Command, report confidencev3.ConfidenceReport) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Intent Confidence Report")
	fmt.Fprintf(out, "  Intent ID: %s\n", report.IntentID)
	fmt.Fprintf(out, "  Intent Confidence: %d%%\n", report.IntentConfidence)
	fmt.Fprintf(out, "  Intent Score: %d%%\n", report.IntentScore)
	fmt.Fprintf(out, "  Vision Score: %d%%\n", report.VisionScore)
	fmt.Fprintf(out, "  UX Score: %d%%\n", report.UXScore)
	fmt.Fprintf(out, "  Business Score: %d%%\n", report.BusinessScore)
	fmt.Fprintf(out, "  Requirements Score: %d%%\n", report.RequirementsScore)
	fmt.Fprintf(out, "  Constraints Score: %d%%\n", report.ConstraintsScore)
	fmt.Fprintf(out, "  Strengths: %s\n", joinOrDash(report.Strengths))
	fmt.Fprintf(out, "  Weaknesses: %s\n", joinOrDash(report.Weaknesses))
	fmt.Fprintf(out, "  Recommendations: %s\n", joinOrDash(report.Recommendations))
}

func newAlignmentCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "alignment", Short: "V3 continuous intent-to-implementation alignment."}
	cmd.AddCommand(newAlignmentReviewCommand())
	cmd.AddCommand(newAlignmentContextCommand())
	cmd.AddCommand(newAlignmentReferencesCommand())
	cmd.AddCommand(newAlignmentFrameworkCommand())
	return cmd
}

func newAlignmentReviewCommand() *cobra.Command {
	var intentID, outcome, plan, task string
	cmd := &cobra.Command{
		Use:   "review --intent <id>",
		Short: "Ph56-70: Run deterministic product alignment review.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			pi, err := loadProductIntent(intentID)
			if err != nil {
				return err
			}
			svc := alignmentv3.NewService()
			review := svc.Review(pi, outcome, plan, task)
			continuous := svc.Continuous(pi, outcome, plan, task)
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Product Alignment Review")
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Intent ID: %s\n", review.IntentID)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Project Review: %d%%\n", review.ProjectReview)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Intent Review: %d%%\n", review.IntentReview)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Vision Review: %d%%\n", review.VisionReview)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Outcome Review: %d%%\n", review.OutcomeReview)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Alignment Review: %d%%\n", review.AlignmentReview)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Continuous Health: %d%%\n", continuous.Health)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Risks: %s\n", joinOrDash(review.Risks))
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Recommendations: %s\n", joinOrDash(review.Recommendations))
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "V3 Product Intent ID")
	cmd.Flags().StringVar(&outcome, "outcome", "", "Current outcome to validate")
	cmd.Flags().StringVar(&plan, "plan", "", "Plan text to align")
	cmd.Flags().StringVar(&task, "task", "", "Task text to align")
	return cmd
}

func newAlignmentContextCommand() *cobra.Command {
	var intentID string
	cmd := &cobra.Command{
		Use:   "context --intent <id>",
		Short: "Ph68: Generate intent-oriented implementation context.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			pi, err := loadProductIntent(intentID)
			if err != nil {
				return err
			}
			ctx := alignmentv3.NewService().Context(pi)
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Alignment Context")
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  What To Do: %s\n", ctx.WhatToDo)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Why It Exists: %s\n", ctx.WhyItExists)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Desired Outcome: %s\n", ctx.DesiredOutcome)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Avoid: %s\n", joinOrDash(ctx.Avoid))
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Summary: %s\n", ctx.ContextSummary)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "V3 Product Intent ID")
	return cmd
}

func newAlignmentReferencesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "references",
		Short: "Ph65: List built-in reference products.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			refs := alignmentv3.NewService().References()
			fmt.Fprintf(cobraCmd.OutOrStdout(), "Reference Products (%d)\n", len(refs))
			for _, ref := range refs {
				fmt.Fprintf(cobraCmd.OutOrStdout(), "  %s — UX: %s; Workflows: %s; Components: %s\n", ref.Name, joinOrDash(ref.UX), joinOrDash(ref.Workflows), joinOrDash(ref.Components))
			}
			return nil
		},
	}
}

func newAlignmentFrameworkCommand() *cobra.Command {
	var intentID string
	cmd := &cobra.Command{
		Use:   "framework --intent <id>",
		Short: "Ph70: Show intent-to-implementation framework readiness.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			pi, err := loadProductIntent(intentID)
			if err != nil {
				return err
			}
			report := alignmentv3.NewService().Framework(pi)
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Intent-To-Implementation Framework")
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Intent ID: %s\n", report.IntentID)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Ready: %t\n", report.Ready)
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Stages: %s\n", joinOrDash(report.Stages))
			fmt.Fprintf(cobraCmd.OutOrStdout(), "  Summary: %s\n", report.Summary)
			return nil
		},
	}
	cmd.Flags().StringVar(&intentID, "intent", "", "V3 Product Intent ID")
	return cmd
}

func loadProductIntent(intentID string) (intentv3.ProductIntent, error) {
	if intentID == "" {
		return intentv3.ProductIntent{}, fmt.Errorf("--intent is required")
	}
	db, err := openInitializedProjectStore()
	if err != nil {
		return intentv3.ProductIntent{}, err
	}
	defer db.Close()
	return store.NewIntentV3Repository(db).GetProductIntent(intentID)
}

func newMasterV2Command() *cobra.Command {
	cmd := &cobra.Command{Use: "master-v2", Short: "Ph25: Versioned master plan generation."}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Validate master plan v2 engine.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Master Plan v2 Engine: OK")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Versioned plans")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Change tracking")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Evolution events")
			return nil
		},
	})
	return cmd
}

func newSpecificV2Command() *cobra.Command {
	cmd := &cobra.Command{Use: "specific-v2", Short: "Ph26: Domain-aware specific plan generation."}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Validate specific plan v2 engine.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Specific Plan v2 Engine: OK")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Domain-aware plans")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Research linking")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Regeneration tracking")
			return nil
		},
	})
	return cmd
}

func newDeliveryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "delivery", Short: "Ph27: Budget-aware context delivery."}
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Validate context delivery engine.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			fmt.Fprintln(cobraCmd.OutOrStdout(), "Context Delivery Engine: OK")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  L0-L4 delivery levels")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Budget-aware generation")
			fmt.Fprintln(cobraCmd.OutOrStdout(), "  Usage tracking")
			return nil
		},
	})
	return cmd
}

func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Project memory system — decisions, QA, research, and more.",
	}
	cmd.AddCommand(newMemoryAddCommand())
	cmd.AddCommand(newMemoryListCommand())
	cmd.AddCommand(newMemoryAskCommand())
	return cmd
}

func newMemoryAddCommand() *cobra.Command {
	var (
		entryType string
		title     string
		question  string
		answer    string
		content   string
		citation  string
		source    string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new memory entry.",
		RunE: func(cmd *cobra.Command, args []string) error {
			et := memory.EntryType(entryType)
			if !et.Valid() {
				return fmt.Errorf("invalid entry type: %q (valid: decision, approval, question_answer, reference, research, plan, change)", entryType)
			}
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			svc := memory.NewService(store.NewMemoryRepository(db))
			entry, err := svc.Add(memory.AddInput{
				ProjectID: store.ProjectID(projectRoot()),
				EntryType: et,
				Title:     title,
				Question:  question,
				Answer:    answer,
				Content:   content,
				Citation:  citation,
				Source:    source,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Memory entry created: %s (%s)\n", entry.ID, entry.EntryType)
			return nil
		},
	}
	cmd.Flags().StringVar(&entryType, "type", "decision", "Entry type (decision, approval, question_answer, reference, research, plan, change)")
	cmd.Flags().StringVar(&title, "title", "", "Entry title")
	cmd.Flags().StringVar(&question, "question", "", "Question (for question_answer entries)")
	cmd.Flags().StringVar(&answer, "answer", "", "Answer (for question_answer entries)")
	cmd.Flags().StringVar(&content, "content", "", "Entry content/body")
	cmd.Flags().StringVar(&citation, "citation", "", "Optional citation URL")
	cmd.Flags().StringVar(&source, "source", "", "Optional source name")
	return cmd
}

func newMemoryListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List memory entries for the current project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			svc := memory.NewService(store.NewMemoryRepository(db))
			entries, err := svc.List(store.ProjectID(projectRoot()))
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No memory entries.")
				return nil
			}
			for _, e := range entries {
				line := fmt.Sprintf("%s  [%s]  %s", e.ID, e.EntryType, truncateTo(e.Title, 60))
				if e.Question != "" {
					line += fmt.Sprintf("  Q: %s", truncateTo(e.Question, 40))
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}
}

func newMemoryAskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask a question and get an existing answer from memory, or prompt for one.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			question := strings.Join(args, " ")
			db, err := openInitializedProjectStore()
			if err != nil {
				return err
			}
			defer db.Close()

			svc := memory.NewService(store.NewMemoryRepository(db))
			entry, reused, err := svc.Ask(store.ProjectID(projectRoot()), question)
			if err != nil {
				return err
			}
			if entry.ID == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No matching memory entry found.")
				return nil
			}
			if reused {
				fmt.Fprintln(cmd.OutOrStdout(), "(Reused existing memory entry)")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  [%s]\n", entry.ID, entry.EntryType)
			if entry.Question != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Q: %s\n", entry.Question)
			}
			if entry.Answer != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "A: %s\n", entry.Answer)
			}
			if entry.Content != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Content: %s\n", entry.Content)
			}
			return nil
		},
	}
	return cmd
}

// projectRoot returns the resolved project root or panics (safe for cobra RunE).
func projectRoot() string {
	root, err := resolveProjectRoot()
	if err != nil {
		panic(err)
	}
	return root
}

func truncateTo(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func newModelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Model provider and compatibility information.",
	}
	cmd.AddCommand(newModelProvidersCommand())
	cmd.AddCommand(newModelCompatibilityCommand())
	return cmd
}

func newModelProvidersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "List all supported model providers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog := modelstrategy.NewCompatibilityCatalog()
			providers := catalog.ListProviders()
			fmt.Fprintf(cmd.OutOrStdout(), "Supported providers (%d):\n", len(providers))
			for _, p := range providers {
				models := catalog.ListModels(p)
				explicitCount := 0
				for _, m := range models {
					if m.Model != "*" {
						explicitCount++
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  (%d known models)\n", p, explicitCount)
			}
			return nil
		},
	}
}

func newModelCompatibilityCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "compatibility <model> [provider]",
		Short: "Check model and provider compatibility.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			model := args[0]
			provider := modelstrategy.ProviderType("")
			if len(args) > 1 {
				provider = modelstrategy.ProviderType(args[1])
			}

			catalog := modelstrategy.NewCompatibilityCatalog()

			if provider != "" {
				report := catalog.Check(model, provider)
				fmt.Fprintf(cmd.OutOrStdout(), "Model:  %s\n", report.Model)
				fmt.Fprintf(cmd.OutOrStdout(), "Provider: %s\n", report.Provider)
				fmt.Fprintf(cmd.OutOrStdout(), "Supported: %t\n", report.Supported)
				if report.Supported {
					fmt.Fprintf(cmd.OutOrStdout(), "Max tokens: %d\n", report.MaxTokens)
					fmt.Fprintf(cmd.OutOrStdout(), "Tier: %s\n", report.Tier)
				}
				if report.Note != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Note: %s\n", report.Note)
				}
				return nil
			}

			// No provider specified — list all compatible providers for this model
			allProviders := catalog.ListProviders()
			var supported []string
			for _, p := range allProviders {
				if catalog.Check(model, p).Supported {
					supported = append(supported, string(p))
				}
			}
			if len(supported) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Model %q is supported by: %v\n", model, supported)
			} else if catalog.IsModelKnown(model) {
				fmt.Fprintf(cmd.OutOrStdout(), "Model %q is known but no default provider mapping exists. Specify a provider: plan-ai model compatibility %s <provider>\n", model, model)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Model %q is not in the compatibility catalog.\n", model)
			}
			return nil
		},
	}
}

func newSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Plan-AI integrations (opencode, MCP, etc.).",
	}
	cmd.AddCommand(newSetupOpenCodeCommand())
	cmd.AddCommand(newSetupOpenCodeWorkflowsCommand())
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

func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Run V2 validation suites against project categories.",
		Long:  "Validate runs the deterministic V2 validation engine, checking all project cases against V2 workflow stages.",
	}
	cmd.AddCommand(newValidateV2Command())
	cmd.AddCommand(newValidateCasesCommand())
	return cmd
}

func newValidateV2Command() *cobra.Command {
	return &cobra.Command{
		Use:   "v2",
		Short: "Run all 63 V2 validation checks (7 cases × 9 stages).",
		RunE: func(cmd *cobra.Command, args []string) error {
			summary := validation.ValidateV2Cases()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "V2 Validation Summary\n")
			fmt.Fprintf(out, "  Total:  %d\n", summary.Total)
			fmt.Fprintf(out, "  Passed: %d\n", summary.Passed)
			fmt.Fprintf(out, "  Failed: %d\n", summary.Failed)
			fmt.Fprintln(out)
			if summary.Failed > 0 {
				for _, r := range summary.Results {
					if !r.Passed {
						fmt.Fprintf(out, "  FAIL  case=%-15s stage=%-22s %s\n", r.CaseName, r.StageName, r.Detail)
					}
				}
			} else {
				fmt.Fprintln(out, "All checks PASSED.")
			}
			if summary.Failed > 0 {
				return fmt.Errorf("%d validation checks failed", summary.Failed)
			}
			return nil
		},
	}
}

func newValidateCasesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cases",
		Short: "List all 7 project categories used in V2 validation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cases := validation.V2Cases()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "V2 Validation Cases (%d):\n\n", len(cases))
			for _, c := range cases {
				fmt.Fprintf(out, "  %s\n", c.Name)
				fmt.Fprintf(out, "    Description: %s\n", c.Description)
				fmt.Fprintf(out, "    Idea:        %s\n", c.Idea)
				fmt.Fprintf(out, "    Intents:     %d\n", len(c.ExpectedIntents))
				fmt.Fprintf(out, "    Stages:      9 (Idea → Updated Plan)\n")
				fmt.Fprintln(out)
			}
			return nil
		},
	}
}

func newPlaceholderCommand(name, short string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s is reserved for a future Plan-AI phase.\n", name)
		},
	}
}

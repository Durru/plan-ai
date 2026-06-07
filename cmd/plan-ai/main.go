package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/agent"
	"github.com/Durru/plan-ai/internal/conversation"
	"github.com/Durru/plan-ai/internal/alignmentv3"
	"github.com/Durru/plan-ai/internal/ambiguityv3"
	"github.com/Durru/plan-ai/internal/confidencev3"
	"github.com/Durru/plan-ai/internal/config"
	"github.com/Durru/plan-ai/internal/continuous"
	"github.com/Durru/plan-ai/internal/core"
	"github.com/Durru/plan-ai/internal/version"
	"github.com/Durru/plan-ai/internal/discoveryv3"
	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/intentv3"
	"github.com/Durru/plan-ai/internal/knowledge"
	"github.com/Durru/plan-ai/internal/research"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/Durru/plan-ai/internal/vision"
	"github.com/spf13/cobra"
)

var ver = version.Version

const configVersion = "2.0.0"

func main() {
	cmd := newRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	app := core.NewApp(ver, ".")

	cmd := &cobra.Command{
		Use:   "plan-ai",
		Short: "Plan-AI prepares implementation plans for AI-assisted projects.",
		Long:  "Plan-AI is a local-first project planning foundation.",
	}

	cmd.AddCommand(newVersionCommand(app))
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newBootstrapCommand())
	cmd.AddCommand(newMigrateCommand())
	cmd.AddCommand(newSyncCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newUninstallCommand())
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
	cmd.AddCommand(newMCPCommand())
	cmd.AddCommand(newUpdateVpsCommand())

	return cmd
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

	home, err := resolveHomeRoot()
	if err != nil {
		return nil, err
	}

	// The global store is mandatory in Phase 1. If it isn't installed, do NOT
	// silently fall back to a project-local store.
	if _, err := os.Stat(config.GlobalDBPath(home)); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("global store not initialized; run `plan-ai install` first")
		}
		return nil, err
	}
	globalDB, err := openInstalledGlobalStore()
	if err != nil {
		return nil, err
	}
	defer globalDB.Close()

	resolver := store.NewProjectResolver(home, globalDB)
	loc, err := resolver.Resolve(projectRoot)
	if err != nil {
		if errors.Is(err, store.ErrLegacyLocalStoreFound) {
			return nil, fmt.Errorf("legacy project-local store found at %s; use 'plan-ai init --local' to keep it or 'plan-ai migrate local-to-global' to migrate", config.ProjectDBPath(projectRoot))
		}
		return nil, err
	}

	db, err := store.Open(loc.DBPath)
	if err != nil {
		return nil, err
	}
	if err := store.RunProjectMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
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

			gw := conversation.NewGateway(db)
			resp, err := gw.ProcessMessage(projectID, message)
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



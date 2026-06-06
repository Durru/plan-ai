package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/approval"
	"github.com/plan-ai/plan-ai/internal/change"
	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/ingestion"
	"github.com/plan-ai/plan-ai/internal/intent"
	intentv3 "github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/knowledge"
	"github.com/plan-ai/plan-ai/internal/memory"
	"github.com/plan-ai/plan-ai/internal/modelstrategy"
	"github.com/plan-ai/plan-ai/internal/reference"
	"github.com/plan-ai/plan-ai/internal/research"
	"github.com/plan-ai/plan-ai/internal/store"
	"github.com/plan-ai/plan-ai/internal/vision"
	"github.com/spf13/cobra"
)

// ──────────────────────────────────────────────
// Knowledge commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Ingest commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Intent commands (V2 + V3)
// ──────────────────────────────────────────────

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

// ── V3 Intent commands (Phase 51 + Phase 52) ──

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

// ── Intent helpers ──

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

// ──────────────────────────────────────────────
// Vision commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Research commands
// ──────────────────────────────────────────────

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

// ── Research + knowledge helpers ──

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
	for _, raw := range raw {
		out = append(out, raw)
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

// ──────────────────────────────────────────────
// Reference commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Jobs commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Impact commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Snapshot commands
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Memory commands
// ──────────────────────────────────────────────

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
			pRoot, pErr := projectRoot()
			if pErr != nil { return pErr }
			entry, err := svc.Add(memory.AddInput{
				ProjectID: store.ProjectID(pRoot),
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
			pRoot, pErr := projectRoot()
			if pErr != nil { return pErr }
			entries, err := svc.List(store.ProjectID(pRoot))
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
			pRoot, pErr := projectRoot()
			if pErr != nil { return pErr }
			entry, reused, err := svc.Ask(store.ProjectID(pRoot), question)
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

// projectRoot returns the resolved project root and any error.
// Prefer resolveProjectRoot() directly in new code. This is kept for
// backward compatibility with callers that already use it.
func projectRoot() (string, error) {
	return resolveProjectRoot()
}

func truncateTo(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ──────────────────────────────────────────────
// Model commands
// ──────────────────────────────────────────────

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

package main

import (
	"fmt"
	"strings"

	"github.com/Durru/plan-ai/internal/approval"
	approvedcontext "github.com/Durru/plan-ai/internal/context"
	"github.com/Durru/plan-ai/internal/requirements"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

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
		Short: "Store approved context (deduplicated, with FTS search).",
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
			authority := approvedcontext.NewAuthorityService(store.NewApprovedContextRepository(db), db)
			item, existed, err := authority.Add(approvedcontext.ApprovedItem{
				ProjectID: store.ProjectID(projectRoot),
				Type:      approvedcontext.ApprovedType(typ),
				SourceID:  source,
				Content:   args[0],
			})
			if err != nil {
				return err
			}
			if existed {
				fmt.Fprintf(cmd.OutOrStdout(), "Approved context already exists: %s (%s)\n", item.ID, item.Type)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Approved context stored: %s (%s)\n", item.ID, item.Type)
			}
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

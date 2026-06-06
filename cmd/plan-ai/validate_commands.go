package main

import (
	"fmt"

	"github.com/plan-ai/plan-ai/internal/validation"
	"github.com/spf13/cobra"
)

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

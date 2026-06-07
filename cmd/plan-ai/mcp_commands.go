package main

import (
	"encoding/json"
	"fmt"
	"os"

	mcpserver "github.com/Durru/plan-ai/internal/mcp"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/spf13/cobra"
)

func newMCPCommand() *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run Plan-AI MCP integrations.",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve Plan-AI MCP over stdio.",
		RunE: func(cmd *cobra.Command, args []string) error {
			deps := mcpserver.DefaultToolDependencies()
			projectRoot, _ := store.ResolveProjectRoot()
			ctx := mcpserver.ToolContext{ProjectRoot: projectRoot}
			return mcpserver.ServeSDKStdio(ctx, deps, os.Getenv("PLAN_AI_MCP_MINIMAL") == "true")
		},
	}

	listToolsCmd := &cobra.Command{
		Use:   "list-tools",
		Short: "List all registered MCP tools with their schemas.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := mcpserver.ToolContext{}
			s := mcpserver.NewServer(ctx)
			deps := mcpserver.DefaultToolDependencies()
			if err := mcpserver.RegisterDefaultTools(s, &deps); err != nil {
				return fmt.Errorf("register tools: %w", err)
			}
			if os.Getenv("PLAN_AI_MCP_MINIMAL") == "true" {
				s.SetMinimalMode(true)
			}
			tools := s.ListTools()
			data, _ := json.MarshalIndent(tools, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	validateToolsCmd := &cobra.Command{
		Use:   "validate-tools",
		Short: "Validate registered MCP tools for OpenCode compatibility.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := mcpserver.ToolContext{}
			s := mcpserver.NewServer(ctx)
			deps := mcpserver.DefaultToolDependencies()
			if err := mcpserver.RegisterDefaultTools(s, &deps); err != nil {
				return fmt.Errorf("register tools: %w", err)
			}
			tools := s.ListTools()
			allValid, results := mcpserver.ValidateAllTools(tools)
			for _, r := range results {
				if !r.Valid {
					fmt.Fprintf(cmd.ErrOrStderr(), "FAIL  %s\n", r.Name)
					for _, issue := range r.Issues {
						fmt.Fprintf(cmd.ErrOrStderr(), "       - %s\n", issue)
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "OK    %s\n", r.Name)
					for _, w := range r.Warnings {
						fmt.Fprintf(cmd.OutOrStdout(), "       warn: %s\n", w)
					}
				}
			}
			if allValid {
				fmt.Fprintln(cmd.OutOrStdout(), "VALIDATE_TOOLS_OK")
			} else {
				return fmt.Errorf("VALIDATE_TOOLS_FAIL")
			}
			return nil
		},
	}

	callToolCmd := &cobra.Command{
		Use:   "call-tool <name> [json-args]",
		Short: "Call an MCP tool directly and print its result.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := mcpserver.ToolContext{}
			s := mcpserver.NewServer(ctx)
			deps := mcpserver.DefaultToolDependencies()
			if err := mcpserver.RegisterDefaultTools(s, &deps); err != nil {
				return fmt.Errorf("register tools: %w", err)
			}
			if os.Getenv("PLAN_AI_MCP_MINIMAL") == "true" {
				s.SetMinimalMode(true)
			}
			toolName := args[0]
			var mcpArgs map[string]any
			if len(args) > 1 {
				if err := json.Unmarshal([]byte(args[1]), &mcpArgs); err != nil {
					return fmt.Errorf("invalid json args: %w", err)
				}
			}
			result := s.ExecuteTool(toolName, mcpArgs)
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}

	mcpCmd.AddCommand(serveCmd, listToolsCmd, validateToolsCmd, callToolCmd)
	return mcpCmd
}

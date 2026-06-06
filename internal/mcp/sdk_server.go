package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// sdkHandlerAdapter returns a function that converts a ToolHandlerFunc into an
// SDK-compatible handler.
func sdkHandlerAdapter() func(ToolHandlerFunc) func(context.Context, mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	return func(fn ToolHandlerFunc) func(context.Context, mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			result, err := fn(req.GetArguments())
			if err != nil {
				return mcpsdk.NewToolResultError(err.Error()), nil
			}
			if result == nil {
				result = &ToolResult{Success: true}
			}
			text, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return mcpsdk.NewToolResultError(err.Error()), nil
			}
			if !result.Success {
				return mcpsdk.NewToolResultError(string(text)), nil
			}
			return mcpsdk.NewToolResultText(string(text)), nil
		}
	}
}

// NewSDKServer creates the production MCP server backed by mark3labs/mcp-go.
func NewSDKServer(ctx ToolContext, deps ToolDependencies, minimal bool) (*mcpserver.MCPServer, error) {
	srv := mcpserver.NewMCPServer(
		"plan-ai",
		"2.0.0",
		mcpserver.WithToolCapabilities(true),
	)
	if err := RegisterSDKDefaultTools(srv, &deps, minimal); err != nil {
		return nil, fmt.Errorf("register tools: %w", err)
	}
	return srv, nil
}

// ServeSDKStdio serves the mark3labs MCP server over stdio.
// If ctx.ProjectRoot is non-empty, a shared project store is opened and
// reused across all tool calls for the lifetime of the process.
func ServeSDKStdio(ctx ToolContext, deps ToolDependencies, minimal bool) error {
	if ctx.ProjectRoot != "" {
		if err := SetSharedProjectStore(ctx.ProjectRoot); err != nil {
			return fmt.Errorf("shared store: %w", err)
		}
		defer CloseSharedProjectStore()
	}

	srv, err := NewSDKServer(ctx, deps, minimal)
	if err != nil {
		return err
	}
	return mcpserver.ServeStdio(srv)
}

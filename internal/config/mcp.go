package config

import "path/filepath"

// BinNameMCP is the CLI binary that serves MCP through `plan-ai mcp serve`.
const BinNameMCP = "plan-ai"

// MCPCommand returns the canonical MCP server command entry for OpenCode configs.
// If binDir is non-empty, it uses the absolute path; otherwise just the basename.
func MCPCommand(binDir string) []string {
	cmd := BinNameMCP
	if binDir != "" {
		cmd = filepath.Join(binDir, BinNameMCP)
	}
	return []string{cmd, "mcp", "serve"}
}

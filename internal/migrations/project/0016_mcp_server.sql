-- Phase 19: MCP Server
-- Tracks MCP tool registrations and execution history.

CREATE TABLE IF NOT EXISTS mcp_tools (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  schema_def TEXT NOT NULL DEFAULT '{}',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS mcp_runs (
  id TEXT PRIMARY KEY,
  tool_name TEXT NOT NULL,
  arguments TEXT NOT NULL DEFAULT '{}',
  success INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mcp_tools_name ON mcp_tools(name);
CREATE INDEX IF NOT EXISTS idx_mcp_tools_enabled ON mcp_tools(enabled);
CREATE INDEX IF NOT EXISTS idx_mcp_runs_tool ON mcp_runs(tool_name);
CREATE INDEX IF NOT EXISTS idx_mcp_runs_created ON mcp_runs(created_at);

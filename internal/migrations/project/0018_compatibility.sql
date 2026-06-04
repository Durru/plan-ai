-- Phase 18-20 compatibility tables and views.
-- Additive only — does not drop or rename any existing table.

-- tool_runs: compatibility view over mcp_runs (Phase 19 MCP Server)
CREATE VIEW IF NOT EXISTS tool_runs AS
SELECT id, tool_name, arguments, success, error_message, created_at FROM mcp_runs;

-- tool_audit: compatibility view joining mcp_tools and mcp_runs
CREATE VIEW IF NOT EXISTS tool_audit AS
SELECT
  r.id AS run_id,
  t.id AS tool_id,
  t.name AS tool_name,
  t.description,
  r.arguments,
  r.success,
  r.error_message,
  r.created_at AS executed_at
FROM mcp_runs r
LEFT JOIN mcp_tools t ON r.tool_name = t.name;

-- provider_registry: model provider registry (Phase 19 / MCP)
CREATE TABLE IF NOT EXISTS provider_registry (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  provider_type TEXT NOT NULL DEFAULT '',
  endpoint TEXT NOT NULL DEFAULT '',
  config TEXT NOT NULL DEFAULT '{}',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- skill_registry: skill registry (Phase 20 / OpenCode Integration)
CREATE TABLE IF NOT EXISTS skill_registry (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  version TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  checksum TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_provider_registry_name ON provider_registry(name);
CREATE INDEX IF NOT EXISTS idx_provider_registry_type ON provider_registry(provider_type);
CREATE INDEX IF NOT EXISTS idx_skill_registry_name ON skill_registry(name);
CREATE INDEX IF NOT EXISTS idx_skill_registry_source ON skill_registry(source);

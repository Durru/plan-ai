-- Phase 21/22 compatibility migration.
-- Runtime source of truth remains the inline migration runner in internal/store/store.go.
-- Adds contract-compatible names requested for Agent System and Continuous Planning.

CREATE VIEW IF NOT EXISTS delegated_jobs AS
SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at
FROM agent_delegated_jobs;

CREATE VIEW IF NOT EXISTS agent_responses AS
SELECT
  id,
  id AS run_id,
  response AS content,
  status,
  created_at,
  updated_at
FROM agent_runs_v2;

CREATE TABLE IF NOT EXISTS continuous_status (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  active_plan TEXT NOT NULL DEFAULT '',
  active_phase TEXT NOT NULL DEFAULT '',
  next_task TEXT NOT NULL DEFAULT '',
  blocked_items TEXT NOT NULL DEFAULT '[]',
  approvals_needed TEXT NOT NULL DEFAULT '[]',
  outdated_plans TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE VIEW IF NOT EXISTS context_delivery_logs AS
SELECT id, project_id, level, content, created_at
FROM context_deliveries;

CREATE INDEX IF NOT EXISTS idx_continuous_status_project ON continuous_status(project_id);

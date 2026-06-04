-- Phase 21 Agent System migration.
-- Runtime source of truth remains the inline migration runner in internal/store/store.go.
-- Migration number 0019 is used because 0018 was already allocated to Phase 18-20 compatibility.

CREATE TABLE IF NOT EXISTS agent_runs_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'processed',
  response TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_messages (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  role TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  FOREIGN KEY (run_id) REFERENCES agent_runs_v2(id)
);

CREATE TABLE IF NOT EXISTS agent_delegated_jobs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent TEXT NOT NULL DEFAULT '',
  capability TEXT NOT NULL DEFAULT '',
  workflow_type TEXT NOT NULL DEFAULT '',
  job_type TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  result_summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  completed_at TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_agent_runs_v2_project ON agent_runs_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_runs_v2_status ON agent_runs_v2(status);
CREATE INDEX IF NOT EXISTS idx_agent_messages_run ON agent_messages(run_id);
CREATE INDEX IF NOT EXISTS idx_agent_delegated_jobs_project ON agent_delegated_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_delegated_jobs_status ON agent_delegated_jobs(status);

CREATE VIEW IF NOT EXISTS agent_runs AS SELECT * FROM agent_runs_v2;
CREATE VIEW IF NOT EXISTS subagent_outputs AS
  SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at
  FROM agent_delegated_jobs;

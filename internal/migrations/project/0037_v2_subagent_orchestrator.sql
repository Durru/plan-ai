CREATE TABLE IF NOT EXISTS subagent_tasks_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  agent_type TEXT NOT NULL,
  objective TEXT NOT NULL DEFAULT '',
  capability TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  provenance TEXT NOT NULL DEFAULT '',
  validation_status TEXT NOT NULL DEFAULT 'pending',
  isolated INTEGER NOT NULL DEFAULT 1,
  temporary INTEGER NOT NULL DEFAULT 1,
  memory_policy TEXT NOT NULL DEFAULT 'no-independent-persistent-memory',
  result_summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_subagent_tasks_v2_project ON subagent_tasks_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_subagent_tasks_v2_type ON subagent_tasks_v2(agent_type);

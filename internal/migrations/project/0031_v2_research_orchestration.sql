CREATE TABLE IF NOT EXISTS research_orchestration_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  agent_type TEXT NOT NULL,
  topic TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  evidence TEXT NOT NULL DEFAULT '[]',
  confidence INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_research_orchestration_project ON research_orchestration_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_research_orchestration_agent ON research_orchestration_runs(agent_type);

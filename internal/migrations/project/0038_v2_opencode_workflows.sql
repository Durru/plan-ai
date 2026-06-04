CREATE TABLE IF NOT EXISTS opencode_workflows_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  commands TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'synced',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_opencode_workflows_v2_project ON opencode_workflows_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_opencode_workflows_v2_status ON opencode_workflows_v2(status);

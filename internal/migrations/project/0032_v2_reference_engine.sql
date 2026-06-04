CREATE TABLE IF NOT EXISTS project_references_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_type TEXT NOT NULL,
  uri TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT 'functional',
  notes TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'needs_review',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_project_references_v2_project ON project_references_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_project_references_v2_status ON project_references_v2(status);

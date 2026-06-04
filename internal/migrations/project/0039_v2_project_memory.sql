CREATE TABLE IF NOT EXISTS project_memory_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  entry_type TEXT NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  question TEXT NOT NULL DEFAULT '',
  answer TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  citation TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_project ON project_memory_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_type ON project_memory_v2(entry_type);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_status ON project_memory_v2(status);

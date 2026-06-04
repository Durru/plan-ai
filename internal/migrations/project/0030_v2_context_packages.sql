CREATE TABLE IF NOT EXISTS context_packages_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  package_type TEXT NOT NULL,
  model_target TEXT NOT NULL DEFAULT 'generic',
  summary TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  priority INTEGER NOT NULL DEFAULT 5,
  token_budget INTEGER NOT NULL DEFAULT 4096,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_context_packages_v2_project ON context_packages_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_context_packages_v2_type ON context_packages_v2(package_type);

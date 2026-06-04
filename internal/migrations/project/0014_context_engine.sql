CREATE TABLE IF NOT EXISTS context_views_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  name TEXT NOT NULL,
  view_type TEXT NOT NULL DEFAULT 'general',
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
ALTER TABLE context_views ADD COLUMN view_type TEXT NOT NULL DEFAULT 'general';
CREATE INDEX IF NOT EXISTS idx_context_views_v2_project ON context_views_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_context_views_v2_type ON context_views_v2(view_type);
CREATE INDEX IF NOT EXISTS idx_context_chunks_view ON context_chunks(context_view_id);

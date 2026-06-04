CREATE TABLE IF NOT EXISTS vision_documents (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent_profile_id TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  functional_vision TEXT NOT NULL DEFAULT '',
  visual_vision TEXT NOT NULL DEFAULT '',
  technical_vision TEXT NOT NULL DEFAULT '',
  operational_vision TEXT NOT NULL DEFAULT '',
  business_vision TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  approved INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_vision_documents_project ON vision_documents(project_id);
CREATE INDEX IF NOT EXISTS idx_vision_documents_status ON vision_documents(status);

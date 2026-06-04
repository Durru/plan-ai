CREATE TABLE IF NOT EXISTS approval_records (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  state TEXT NOT NULL DEFAULT 'draft',
  reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_approval_records_project ON approval_records(project_id);
CREATE INDEX IF NOT EXISTS idx_approval_records_target ON approval_records(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_approval_records_state ON approval_records(state);

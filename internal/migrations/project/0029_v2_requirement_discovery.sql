CREATE TABLE IF NOT EXISTS requirement_candidates (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  dependencies TEXT NOT NULL DEFAULT '[]',
  ambiguities TEXT NOT NULL DEFAULT '[]',
  state TEXT NOT NULL DEFAULT 'candidate',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_requirement_candidates_project ON requirement_candidates(project_id);
CREATE INDEX IF NOT EXISTS idx_requirement_candidates_state ON requirement_candidates(state);

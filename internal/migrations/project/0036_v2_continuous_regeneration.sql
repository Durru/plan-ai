CREATE TABLE IF NOT EXISTS continuous_regenerations_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT 'affected-sections',
  affected_sections TEXT NOT NULL DEFAULT '[]',
  preserved_sections TEXT NOT NULL DEFAULT '[]',
  snapshot_required INTEGER NOT NULL DEFAULT 1,
  approval_required INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_continuous_regenerations_v2_project ON continuous_regenerations_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_continuous_regenerations_v2_status ON continuous_regenerations_v2(status);

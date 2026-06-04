CREATE TABLE IF NOT EXISTS implementation_packages_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_id TEXT NOT NULL DEFAULT '',
  model_target TEXT NOT NULL DEFAULT 'opencode',
  what_to_do TEXT NOT NULL DEFAULT '',
  how_to_do_it TEXT NOT NULL DEFAULT '',
  files_to_touch TEXT NOT NULL DEFAULT '[]',
  files_not_to_touch TEXT NOT NULL DEFAULT '[]',
  examples TEXT NOT NULL DEFAULT '[]',
  commands TEXT NOT NULL DEFAULT '[]',
  validations TEXT NOT NULL DEFAULT '[]',
  rollback_notes TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_implementation_packages_v2_project ON implementation_packages_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_implementation_packages_v2_plan ON implementation_packages_v2(plan_id);

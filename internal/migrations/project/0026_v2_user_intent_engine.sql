CREATE TABLE IF NOT EXISTS intent_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  primary_intent TEXT NOT NULL DEFAULT '{}',
  secondary_goals TEXT NOT NULL DEFAULT '[]',
  constraints_json TEXT NOT NULL DEFAULT '[]',
  expectations TEXT NOT NULL DEFAULT '[]',
  success_criteria TEXT NOT NULL DEFAULT '[]',
  priorities TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  approved INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_intent_profiles_project ON intent_profiles(project_id);
CREATE INDEX IF NOT EXISTS idx_intent_profiles_status ON intent_profiles(status);

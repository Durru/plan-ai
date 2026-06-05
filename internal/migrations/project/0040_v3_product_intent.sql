CREATE TABLE IF NOT EXISTS intent_v3_product_intents (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  description TEXT NOT NULL,
  expected_outcome TEXT NOT NULL DEFAULT '',
  desired_experience TEXT NOT NULL DEFAULT '',
  desired_result TEXT NOT NULL DEFAULT '',
  user_expectations TEXT NOT NULL DEFAULT '[]',
  non_expectations TEXT NOT NULL DEFAULT '[]',
  success_definition TEXT NOT NULL DEFAULT '',
  failure_definition TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  discovery_result_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intent_v3_product_intents_project ON intent_v3_product_intents(project_id);
CREATE INDEX IF NOT EXISTS idx_intent_v3_product_intents_status ON intent_v3_product_intents(status);

CREATE TABLE IF NOT EXISTS intent_v3_discovery_results (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  raw_input TEXT NOT NULL,
  detected_intent TEXT NOT NULL DEFAULT '',
  objectives TEXT NOT NULL DEFAULT '[]',
  restrictions TEXT NOT NULL DEFAULT '[]',
  preferences TEXT NOT NULL DEFAULT '[]',
  refs_list TEXT NOT NULL DEFAULT '[]',
  expectations TEXT NOT NULL DEFAULT '[]',
  classification TEXT NOT NULL DEFAULT '',
  gaps TEXT NOT NULL DEFAULT '[]',
  questions TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intent_v3_discovery_results_project ON intent_v3_discovery_results(project_id);

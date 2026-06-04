CREATE TABLE IF NOT EXISTS model_profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  provider TEXT NOT NULL,
  model TEXT NOT NULL,
  config TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS prompt_contracts (
  id TEXT PRIMARY KEY,
  contract_type TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS output_schemas (
  id TEXT PRIMARY KEY,
  schema_type TEXT NOT NULL,
  fields TEXT NOT NULL DEFAULT '{}',
  required TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_prompt_contracts_type ON prompt_contracts(contract_type);
CREATE INDEX IF NOT EXISTS idx_output_schemas_type ON output_schemas(schema_type);

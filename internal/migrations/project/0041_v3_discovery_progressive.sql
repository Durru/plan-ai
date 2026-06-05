CREATE TABLE IF NOT EXISTS discovery_v3_questions (
  id TEXT PRIMARY KEY,
  intent_id TEXT NOT NULL,
  level TEXT NOT NULL,
  question TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  required INTEGER NOT NULL DEFAULT 0,
  related_fields TEXT NOT NULL DEFAULT '[]',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_questions_intent ON discovery_v3_questions(intent_id);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_questions_level ON discovery_v3_questions(level);

CREATE TABLE IF NOT EXISTS discovery_v3_answers (
  id TEXT PRIMARY KEY,
  question_id TEXT NOT NULL,
  intent_id TEXT NOT NULL,
  level TEXT NOT NULL,
  answer TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_answers_intent ON discovery_v3_answers(intent_id);
CREATE INDEX IF NOT EXISTS idx_discovery_v3_answers_question ON discovery_v3_answers(question_id);

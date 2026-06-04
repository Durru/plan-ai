-- Phase 8 Store Layer v2 global migration.
-- Runtime source of truth currently remains the inline migration runner in internal/store/store.go.
-- This file mirrors the schema for review/documentation and future extraction.

CREATE TABLE IF NOT EXISTS global_config (key TEXT PRIMARY KEY, value TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_tools (id TEXT PRIMARY KEY, name TEXT NOT NULL, version TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_integrations (id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'inactive', config TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_skills (id TEXT PRIMARY KEY, name TEXT NOT NULL, source TEXT NOT NULL DEFAULT '', checksum TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_skill_cache (id TEXT PRIMARY KEY, skill_id TEXT NOT NULL, cache_key TEXT NOT NULL, value TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(skill_id, cache_key));
CREATE TABLE IF NOT EXISTS global_knowledge (id TEXT PRIMARY KEY, topic TEXT NOT NULL, category TEXT NOT NULL DEFAULT 'general', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_research (id TEXT PRIMARY KEY, topic TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_templates (id TEXT PRIMARY KEY, name TEXT NOT NULL, template_type TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_model_profiles (id TEXT PRIMARY KEY, name TEXT NOT NULL, provider TEXT NOT NULL DEFAULT '', model TEXT NOT NULL DEFAULT '', config TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_logs (id TEXT PRIMARY KEY, level TEXT NOT NULL, message TEXT NOT NULL, metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL);

CREATE INDEX IF NOT EXISTS idx_global_tools_name ON global_tools(name);
CREATE INDEX IF NOT EXISTS idx_global_skills_name ON global_skills(name);
CREATE INDEX IF NOT EXISTS idx_global_knowledge_topic ON global_knowledge(topic);
CREATE INDEX IF NOT EXISTS idx_global_research_topic ON global_research(topic);

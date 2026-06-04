# Domain Model Migration Strategy

## Principle

Existing tables from migrations 0001–0005 are NOT destroyed, altered, or
renamed. The domain model v2 is additive — new entities get new tables,
existing entities keep their current schema.

## Existing Table Coverage

| Legacy entity | Table | Migration | Status |
|---|---|---|---|
| Project | `project_state` | 0001 | Compatible (uses raw status, new model adds `ProjectStatus`) |
| Plan (Master/Specific) | `plans` | 0002 | Compatible |
| Phase | `phases` | 0002 | Compatible |
| Task | `tasks` | 0002 | Compatible |
| Decision | `decisions` | 0002 | Compatible (new fields are additive) |
| ResearchEntry | `research_entries` | 0002/0005 | Compatible (Research struct has backward compat fields) |
| KnowledgeObject | `knowledge_objects` | 0002/0004 | Compatible (KnowledgeType field is new) |
| Validation | `validations` | 0002 | Compatible (new fields are additive) |
| Snapshot | `snapshots` | 0002 | Compatible (ProjectID field is new) |

## New Entity Tables Needed (Phase 8)

| Entity | Proposed table | Fields |
|---|---|---|
| Vision | `visions` | id, project_id, title, summary, expected_outcome, approved, created_at, updated_at |
| Requirement | `requirements` | id, project_id, type, statement, approved, created_at, updated_at |
| Constraint | `constraints` | id, project_id, type, description, approved, created_at, updated_at |
| ChangeRequest | `change_requests` | id, project_id, reason, description, status, requester, created_at, updated_at |
| ImpactReport | `impact_reports` | id, change_request_id, affected_plans, affected_phases, affected_tasks, affected_decisions, summary, created_at |
| ImplementationDocument | `implementation_documents` | id, specific_plan_id, title, content, version, created_at, updated_at |

## Existing Table Extensions (Phase 8, additive only)

| Table | New columns | Default |
|---|---|---|
| `project_state` | `description TEXT NOT NULL DEFAULT ''` | empty string |
| `decisions` | `project_id TEXT NOT NULL DEFAULT ''`, `rationale TEXT NOT NULL DEFAULT ''`, `alternatives TEXT NOT NULL DEFAULT ''` | empties |
| `validations` | `type TEXT NOT NULL DEFAULT 'manual'`, `details TEXT NOT NULL DEFAULT ''` | manual, empty |
| `snapshots` | `project_id TEXT NOT NULL DEFAULT ''` | empty |
| `knowledge_objects` | `type TEXT NOT NULL DEFAULT 'reference'` | reference |

## Migration Order

1. Phase 8 migration 0006: Create new entity tables (CREATE TABLE IF NOT EXISTS)
2. Phase 8 migration 0007: Add new columns to existing tables (ALTER TABLE ADD COLUMN)
3. No data backfill is needed — new columns have safe defaults

## Rollback

All migrations use `CREATE TABLE IF NOT EXISTS` / `ALTER TABLE ADD COLUMN`.
They are idempotent. Rollback would involve dropping the new tables and columns,
but this is not recommended once data exists.

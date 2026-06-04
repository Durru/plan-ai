# Storage Layer

Plan-AI uses two local SQLite stores:

- Global store: `$PLAN_AI_HOME/.plan-ai/global.db` or `~/.plan-ai/global.db`.
- Project store: `$PLAN_AI_PROJECT_ROOT/.plan-ai/project.db` or `<cwd>/.plan-ai/project.db`.

`PLAN_AI_HOME` is intentionally treated as the home root. The code appends `.plan-ai`; tests must use sandbox roots such as `.tmp/home`.

## Responsibilities

The Phase 8 store layer connects the Phase 7 domain model to persistence without implementing engines. It prepares tables for future ingestion, vision, planning, research, knowledge, validation, change, context, MCP, and agent outputs, but only minimal repository behavior is implemented.

## Migrations

Migrations are additive and recorded in `schema_migrations`. Existing legacy tables remain readable. Migration `0007_store_layer_v2` adds definitive project tables and indexes. Global migration `0007_store_layer_v2_global` adds global configuration, tool, skill, template, model profile, knowledge, research, and log tables.

No migration drops or rewrites legacy data.

## Transactions

`store.WithTransaction` wraps multiple SQL operations with rollback-on-error and commit-on-success.

## FTS5

FTS5 virtual tables are prepared for:

- `knowledge_objects_fts`
- `research_entries_fts`
- `implementation_documents_fts`
- `raw_inputs_fts`

If a SQLite build lacks FTS5, tests skip the FTS assertion with an explicit message. The normal repositories still use `LIKE` search fallback behavior.

## CLI safety

`plan-ai install` prepares the global store. `plan-ai init` prepares the project store. `plan-ai status` reports store status. `plan-ai doctor` safely checks paths and migration status; it does not touch real OpenCode configuration.

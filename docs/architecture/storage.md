# Storage Architecture

Plan-AI uses two local SQLite databases plus readable JSON configuration files.

## Global Store

The global store lives at `~/.plan-ai/global.db` with supporting directories:

```text
~/.plan-ai/
  global.db
  config.json
  cache/
  skills/
  logs/
  data/
  backups/
```

For sandboxed or automated validation, set `PLAN_AI_HOME`. When set, the same layout is created below that home root, for example `$PLAN_AI_HOME/.plan-ai/global.db`.

The global store is for installation-level data. Phase 2 stores only the migration ledger and basic global tables:

- `schema_migrations`
- `global_metadata`
- `global_settings`
- `known_projects`

Later phases may use this store for preferences, detected tools, reusable research, and skill metadata. Those future tables are intentionally not created in Phase 2.

## Project Store

Each project gets a local store at `<project>/.plan-ai/project.db` with supporting directories:

```text
<project>/.plan-ai/
  project.db
  config.json
  cache/
  snapshots/
  exports/
  docs/
  locks/
  backups/
```

For sandboxed or automated validation, set `PLAN_AI_PROJECT_ROOT`. When set, `init` and project status use that directory as the project root, creating `$PLAN_AI_PROJECT_ROOT/.plan-ai/project.db`.

The project store is for project-scoped state. Phase 2 stores only the migration ledger and basic project tables:

- `schema_migrations`
- `project_metadata`
- `project_settings`
- `project_state`

Later phases may use this store for project context, plans, phases, tasks, decisions, snapshots, and validations. Those future planning tables are intentionally not created in Phase 2.

## Why SQLite

SQLite keeps Plan-AI local-first, portable, and easy to back up. It requires no server process, works well for CLI workflows, and gives stronger structure than ad hoc JSON files for state that will grow over time.

## Why `modernc.org/sqlite`

Plan-AI uses `modernc.org/sqlite` because it is a pure Go SQLite driver. That avoids CGO requirements and makes the CLI easier to build and run on Ubuntu 22.04 and other common environments.

## Why Two Databases

Global data and project data have different lifecycles:

- Global data follows the user installation across projects.
- Project data belongs to the repository and can be backed up, copied, or inspected independently.

Keeping them separate avoids leaking project-specific state into global configuration while still allowing the global store to remember known projects.

## Migrations

Migrations are managed by a small internal runner in `internal/store`. Each database has a `schema_migrations` table with:

- `id`
- `name`
- `applied_at`

Migrations run once and are safe to re-run. The `install` and `init` commands run the relevant migrations every time, making both commands idempotent.

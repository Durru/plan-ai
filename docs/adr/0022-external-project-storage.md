# ADR 0022: External Project Storage

**Status:** Accepted
**Date:** 2026-06-06
**Phase:** 1

## Context

ADR 0002 (Storage Layer) established project-local SQLite databases at `<project>/.plan-ai/project.db`. This worked for single-project workflows but created problems:

- Multiple projects shared no global state.
- Install/uninstall/doctor had no unified registry of known projects.
- The `~/.plan-ai/` directory was underutilized despite being the natural home for user-wide configuration.

Phase 1 introduced external storage as the default: `~/.plan-ai/projects/<slug>/project.db`.

## Decision

1. **Default external**. `plan-ai init` stores project data under `~/.plan-ai/projects/<slug>/project.db` where `slug = config.ProjectSlug(rootPath)`.
2. **Explicit local**. `plan-ai init --local` creates the legacy `<root>/.plan-ai/project.db`. This is opt-in and visible.
3. **Global registry**. `known_projects` table in `~/.plan-ai/global.db` tracks every project's path, slug, and storage mode.
4. **Lazy resolution**. `project_resolver.Resolve(rootPath)` returns the external location for unregistered projects, and `store.OpenStore` provisions the DB on first access.
5. **Migration path**. `plan-ai migrate local-to-global [--force] [--project-root]` copies existing local data into external storage.

## Rationale

External storage by default solves the multi-project, install-once problem. Users install Plan-AI once with `plan-ai install`, then `plan-ai init` in any project. The global registry tracks all projects for `doctor`, `update`, and `uninstall`.

The `--local` escape hatch preserves backward compatibility for existing projects that rely on `<root>/.plan-ai/` paths.

## Consequences

- `plan-ai init` without `--local` creates `~/.plan-ai/projects/<slug>/project.db`.
- `known_projects` table has a UNIQUE constraint on `path`.
- `plan-ai migrate local-to-global` copies data and updates the registry.
- `plan-ai status` reports `Project mode: external` or `Project mode: local`.
- All Phase 2+ commands use the resolver; no command silently falls back to local.

## Supersedes

This ADR supersedes ADR 0002's recommendation of project-local DB as the only option. ADR 0002 remains accurate for the local path but is amended to note that external storage is now the default.

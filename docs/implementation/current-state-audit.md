# Current State Audit — Pre-Phase 7

Plan-AI currently contains a working Go/Cobra/SQLite foundation through the old Phase 6. During this audit, an incomplete old Phase 7 Skill Intelligence spike was found in active code and archived because the roadmap now restarts with the Definitive Domain Model.

## Repository state

| Area | Current state | Decision |
|---|---|---|
| Go module | `github.com/plan-ai/plan-ai`, Cobra CLI, SQLite via `modernc.org/sqlite` | Keep |
| CLI | `install`, `init`, `scan`, `status`, `knowledge`, `research`, `dev`, placeholder `plan` | Keep base; future phases should split the monolithic file |
| Store | Inline SQLite migrations `0001`–`0005`, global/project layouts, repositories | Keep as legacy baseline before Store Layer redesign |
| Scanner | Deterministic project scan, stack/dependency detection, fingerprint | Keep; will become Ingestion Layer input |
| Knowledge | Deterministic Knowledge Base with categories, tags, relations, references | Keep as old Phase 5 reference; likely refactor in Phase 12 |
| Research | Deterministic Research Engine with findings, sources, conclusions | Keep as old Phase 6 reference; likely refactor in Phase 12 |
| MCP | Placeholder only | Keep placeholder; implement in Phase 19 |
| Integrations | Placeholder only | Keep placeholder; implement after MCP/context foundations |
| Skills | Old Phase 7 spike was active | Archived; not part of active roadmap |
| Docs | Phase-specific docs up to old Phase 6 plus methodology/templates | Keep useful docs; add new roadmap/orientation docs |
| Sandbox | `scripts/test-sandbox.sh` validates install/init/scan/knowledge/research using env vars | Keep and align to repo-local `.tmp` convention over time |

## Old phases that appear implemented

- Phase 0: repository structure, CLI shell, config helpers.
- Phase 1: methodology, workflows, templates, prompts, `agent.md`.
- Phase 2: SQLite persistence, `install`, `init`, `status`.
- Phase 3: domain entities and repositories.
- Phase 4: scanner and project fingerprinting.
- Phase 5: Knowledge Base.
- Phase 6: Research Engine.

## Active commands found

- `plan-ai version`
- `plan-ai install`
- `plan-ai init`
- `plan-ai scan`
- `plan-ai status`
- `plan-ai knowledge ...`
- `plan-ai research ...`
- `plan-ai dev seed-domain`
- `plan-ai dev list-domain`
- `plan-ai dev seed-knowledge`
- `plan-ai dev seed-research`
- `plan-ai plan` placeholder

The old `skills` command group and `dev seed-skills` command were removed from active code and archived.

## Tests found

Package tests exist for:

- `cmd/plan-ai`
- `internal/config`
- `internal/core`
- `internal/domain`
- `internal/knowledge`
- `internal/research`
- `internal/scanner`
- `internal/store`

The current suite passes after archiving the old Phase 7 spike.

## What stays

- Go/Cobra CLI foundation.
- SQLite layout/migration runner as a baseline.
- Global/project config and sandbox environment override pattern.
- Scanner as deterministic ingestion precursor.
- Knowledge and Research packages as reusable references.
- Existing tests and sandbox script.
- Methodology/templates/prompts as planning references.

## What needs refactor later

- `cmd/plan-ai/main.go` is too large and mixes command wiring, output rendering, seeding, and service orchestration.
- Current domain entities are old-roadmap models; they need a definitive domain pass in the new Phase 7.
- Store migrations mix early domain, research, knowledge, validation, and snapshots in legacy tables; Phase 8 must define the definitive schema.
- Knowledge and Research should move into the new Research/Knowledge Engine after Approved Context exists.
- MCP, integrations, context, planner, validation, change, and skills are placeholders or legacy concepts.

## What was archived or removed

- Archived old Phase 7 Skill Intelligence spike under `docs/archive/old-phases/phase-7-skill-intelligence/`.
- Removed active migration `0006_skill_intelligence` from the migration runner.
- Removed active `skills` CLI command group and `dev seed-skills` from the CLI.
- Removed ignored build artifact `plan-ai` from the working tree.

## Risks and debt

- The repo has no commits yet, so all files appear untracked in `git status`.
- The migration runner uses inline SQL and `ALTER TABLE` statements; future schema changes need careful migration discipline.
- The old Research classifier intentionally uses substring matching, which can produce false positives.
- The sandbox script currently uses `/tmp/plan-ai-sandbox`; the new convention should prefer repo-local `.tmp/plan-ai-sandbox`.
- Active code still reflects old phases; the new roadmap must redesign the domain model before adding new engines.

## Actions done in this audit

- Verified baseline failure caused by old active Skill Intelligence migration count mismatch.
- Archived old Skill Intelligence implementation files as text reference.
- Removed active old Skill Intelligence CLI/store integration.
- Restored tests to green with the old Phase 6 boundary as the active baseline.
- Added transition docs and new roadmap orientation.

## Pending before new Phase 7

- Start new Phase 7 with a domain-model-first design, not with skills/capabilities.
- Decide whether old domain tables are migrated in place or superseded by new tables.
- Define canonical domain object names, statuses, relationships, and approval boundaries.
- Decide CLI surface for the new domain model before Store Layer redesign.

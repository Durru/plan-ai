# Phase 7 Readiness — Definitive Domain Model

The repository is now prepared to start the **new Phase 7 — Definitive Domain Model**. The old Phase 7 Skill Intelligence spike is archived and inactive.

## Why we do not continue the old Phase 7

The previous Phase 7 targeted Skill Intelligence before the project had a definitive domain model and definitive store layer. The roadmap changed: Plan-AI must first define the truth model of the project before adding capability discovery, OpenCode integration, MCP tools, or planner workflows.

Continuing the old Phase 7 would hard-code concepts too early and couple the CLI/store to a capability system before the core domain is stable.

## Accepted architecture direction

Plan-AI is independent first, integrable later.

Core layers now target this order:

1. Store Layer
2. Ingestion Layer
3. Vision Layer
4. Approved Context Layer
5. Research / Knowledge Layer
6. Planning Layer
7. Workflow Layer
8. Model Strategy Layer
9. Orchestrator / Capability Registry
10. Context Layer
11. Change Layer
12. MCP Layer
13. Integration Layer
14. OpenCode Integration Layer
15. Optional Agent Layer

## What is prepared

- Active code compiles and tests pass without old Skill Intelligence.
- Old Skill Intelligence code is preserved under `docs/archive/old-phases/phase-7-skill-intelligence/`.
- Current state audit documents reusable modules and technical debt.
- New roadmap is documented in `docs/roadmap.md`.
- Architecture overview docs now point to the new direction.
- `agent.md` has been shortened into a practical repo instruction file.

## What is intentionally not prepared yet

- No definitive domain model implementation.
- No new migration beyond old Phase 6.
- No MCP server implementation.
- No OpenCode setup or config mutation.
- No skill execution or skill registry.
- No LLM calls.
- No planner implementation.

## Checklist to start new Phase 7

- [ ] Define canonical domain objects and relationships.
- [ ] Define lifecycle statuses and approval boundaries.
- [ ] Decide which old tables remain compatibility references.
- [ ] Write tests for the new domain model before implementation.
- [ ] Keep CLI/Core/MCP/Integration boundaries separate.
- [ ] Keep all validation in sandbox env vars.

## Verification commands

Use sandboxed paths for runtime checks:

```bash
HOME="$PWD/.tmp/home" \
PLAN_AI_HOME="$PWD/.tmp/home" \
PLAN_AI_PROJECT_ROOT="$PWD/.tmp/project" \
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" \
go run ./cmd/plan-ai status
```

Code verification:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
```

## Acceptance criteria for starting Phase 7

- Current codebase compiles.
- Current tests pass or documented failures are accepted.
- No real `~/.plan-ai` is touched during validation.
- No real OpenCode configuration is touched.
- Roadmap points to **Phase 7 — Definitive Domain Model**.
- Old Skill Intelligence remains archived only.

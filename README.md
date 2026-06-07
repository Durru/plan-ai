# Plan-AI

**Local-first continuous implementation planning for AI-assisted projects.**

[![Go Build](https://github.com/Durru/plan-ai/actions/workflows/go.yml/badge.svg)](https://github.com/Durru/plan-ai/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![OpenCode](https://img.shields.io/badge/OpenCode-1.16+-6C47FF)](https://opencode.ai)

---

## What Plan-AI Does

AI coding fails when the plan lives only in chat. Plan-AI gives the project its own **planning memory** — durable, searchable, local SQLite. Before any agent writes code, Plan-AI ensures:

1. **Intent is understood** — Discovery questions, ambiguity analysis, confidence scoring
2. **Research is done once** — Reuse approved research, track freshness, promote to knowledge
3. **Decisions are recorded** — Supersedable, searchable, linked to evidence
4. **Plans are structured** — Master plans → specific plans → phases → tasks
5. **Changes are tracked** — 10-state lifecycle, impact analysis, continuous proposals
6. **Context is derived** — L0-L4 delivery from approved entities only

## The Six Principles (v4)

Every feature enforces these invariants:

| # | Principle | Enforcement |
|---|-----------|-------------|
| 1 | Nada aprobado se vuelve a preguntar | `context.AuthorityService.IsKnown()` dedup |
| 2 | Nada investigado se vuelve a investigar | `research.ReuseService.FindReusable()` |
| 3 | Nada decidido se vuelve a decidir | Decision supersession + `supersedes_id` |
| 4 | Nada planificado se vuelve a planificar | `PlanningGuard` requires approved intent |
| 5 | Solo se actualiza lo afectado | Impact Graph BFS traversal from entity_links |
| 6 | El contexto se deriva de entidades aprobadas | `context.DeliveryEngine` L0-L4 |

## Architecture

```
User → plan-ai CLI ───────────────┐
                                  ├── conversation.Gateway ── agent.Service
OpenCode → plan_ai.agent_message ─┘        │
                                            │
        ┌──── Intent Detection ──── Router ──── Context Loading ────┤
        │                                                           │
        ▼                                                           ▼
  PlanningGuard ──►   planning.Service  ──►  store.Repositories
  (approved intent)   research.Service      (SQLite persistence)
                      context.Authority
                      continuous.Loop
                      memory.Recorder
```

## Install

```bash
# Quick install (recommended)
curl -fsSL https://raw.githubusercontent.com/Durru/plan-ai/main/scripts/install.sh | bash

# Or via go install
go install github.com/Durru/plan-ai/cmd/plan-ai@latest

# Or build from source
git clone https://github.com/Durru/plan-ai.git
cd plan-ai && go build -o plan-ai ./cmd/plan-ai/
sudo mv plan-ai /usr/local/bin/
```

**Safe by default.** `plan-ai install` never touches your real OpenCode config unless you pass `--allow-real-opencode`. Use sandbox mode:

```bash
OPENCODE_CONFIG_DIR=/tmp/sandbox-oc plan-ai install
```

After install, set up globally and verify:

```bash
plan-ai install --allow-real-opencode
plan-ai doctor
```

## Quickstart

```bash
# 1. Install globally
plan-ai install

# 2. Init a project (external storage by default)
cd my-project
plan-ai init

# 3. Discover intent
plan-ai intent create --description "SaaS for task management"

# 4. Approve context
plan-ai approved add --type requirement "Multi-tenant isolation via separate DBs"
plan-ai approved add --type decision "Use schema-per-tenant"

# 5. Plan with guardrails (blocked until intent approved)
plan-ai intent approve <id>
plan-ai plan

# 6. Continuous planning
plan-ai continuous events
plan-ai continuous proposals
```

## Core Commands

```
# Setup & Health
plan-ai install        Install globally with tool detection
plan-ai init           Initialize project (--local for legacy)
plan-ai update         Refresh state and integrations
plan-ai uninstall      Remove components (--allow-real-opencode for OC cleanup)
plan-ai doctor         Check stores, migrations, OpenCode health
plan-ai doctor --fix   Repair stale state (local install only)
plan-ai status         Project status overview

# Discovery & Intent
plan-ai intent create  Create Product Intent (Phase 51)
plan-ai intent approve Approve intent (unblocks planning)
plan-ai discovery start Start progressive discovery (Phase 53)
plan-ai ambiguity analyze  Analyze missing info (Phase 54)
plan-ai confidence evaluate Score understanding (Phase 55)
plan-ai alignment review    Align intent to implementation (Phase 56)

# Context & Research
plan-ai approved add   Store approved context (deduplicated, FTS-backed)
plan-ai approved find  Search approved context
plan-ai research add   Create research entry
plan-ai research reuse Check for reusable approved research
plan-ai knowledge add  Store knowledge from research
plan-ai memory add     Record durable memory (FTS-backed)

# Planning
plan-ai plan           Create master plan, specific plan, implementation doc
plan-ai agent process  Natural-language planning via conversation gateway

# Continuous
plan-ai continuous events      List detected events
plan-ai continuous proposals   List plan update proposals
plan-ai continuous status      Show continuous planning health
plan-ai continuous context L1  Generate planning context

# Validation
plan-ai validate v2    Run 63+ deterministic validation cases

# MCP Tools (38 available via OpenCode)
plan_ai.init_project    plan_ai.create_master_plan    plan_ai.agent_message
plan_ai.approve_plan    plan_ai.continuous_status     plan_ai.run_workflow
plan_ai.detect_changes  plan_ai.rollback_snapshot     plan_ai.create_product_intent
```

## Storage Model

| Store | Path | Purpose |
|-------|------|---------|
| Global | `~/.plan-ai/global.db` | Install state, known projects, global config |
| Project (external) | `~/.plan-ai/projects/<slug>/project.db` | Project data (default) |
| Project (local) | `<project>/.plan-ai/project.db` | Legacy mode (`--local`) |

All persistent — SQLite WAL mode, FTS5 search, idempotent migrations. Runtime data is gitignored.

## OpenCode Integration

**Never corrupted — safety by design:**

| Guard | Where | Mechanism |
|-------|-------|-----------|
| AllowReal flag | CLI (install/init/uninstall) | `--allow-real-opencode` required |
| Built-in guard | `opencode.SetupMCPConfig` | `getpwuid_r` real-home check |
| Sandbox | `$OPENCODE_CONFIG_DIR` | Redirects all writes |
| Merge | Not overwrite | Preserves existing `mcp` entries |
| Backup | Before every write | `.pre-mcp-write.<timestamp>.bak` |
| Atomic | temp+rename | Crash-safe `atomicfile.WriteFile` |

Zero tests touch real OpenCode config. All sandboxed via `t.TempDir()`.

## Package Map (31 packages)

```
internal/
├── agent/         Intent detection, routing, delegation, response building
├── guard/         Planning guardrail — blocks planning without approved intent
├── conversation/  CLI+MCP gateway — single entry point for natural language
├── planning/      Master plans v2, specific plans, implementation documents
├── research/      Research engine — classification, CRUD, reuse, promotion
├── knowledge/     Knowledge base — relations, references, tags
├── vision/        Vision drafts, discovery, documents
├── context/       Approved context authority, FTS search, delivery engine L0-L4
├── change/        Change engine — events, impact analysis, snapshots, versioning
├── continuous/    Continuous planning — detector, loop, proposals, context gen
├── memory/        Durable memory — FTS-backed recorder, Q&A reuse
├── intentv3/      Product Intent engine (V3 lifecycle — draft→approved)
├── discoveryv3/   Progressive discovery — 5-level question engine
├── alertv3/       Product alignment review (Phase 56-70)
├── ambiguityv3/   Missing info, assumptions, conflicts analysis
├── confidencev3/  Intent confidence scoring
├── domain/        Canonical domain types (deprecated for planning.*, research.*)
├── store/         SQLite schema, 85+ tables, migrations, 14 typed repos
├── config/        Paths, layout, config I/O
├── installer/     Component-based install/sync/uninstall, tool detection
├── opencode/      OpenCode integration, setup, doctor, detection, workflows
├── mcp/           MCP server — 39 tools, handlers, SDK integration
├── atomicfile/    Crash-safe file writes (WriteFile + WriteFileWithBackup)
├── impact/        Impact Graph — nodes, edges, BFS traversal, builder
├── capabilities/  Skill registry — SQL-backed, auto-seeded
├── workflows/     Workflow engine — realDispatchStep with 15 step types
├── orchestrator/  Job orchestrator — CreateJob, Execute, capability selection
├── scanner/       Deterministic project scanner (git, languages, deps, files)
├── ingestion/     Input ingestion — classify, extract, normalize
├── modelstrategy/ Model selection, provider classification, retry engine
├── validation/    V2 validation — 63+ test cases
```

## Development

```bash
# Build
go build ./...

# Test (all sandboxed via t.TempDir, never touches real ~/.config/opencode)
go test ./...

# Lint
go vet ./...

# Full check
go test -count=1 ./... && go vet ./...
```

**Codebase metrics:** 31 packages, 85+ DB tables, 39 MCP tools, 120+ CLI commands, 0 raw SQL in service layer, 0 import cycles.

## Audit Status (2026-06-06)

| Category | Status |
|----------|--------|
| CRITICAL issues | 0 |
| OpenCode write paths | 11 auditados, todos con guard |
| Raw SQL in handlers | 0 |
| Data races | 0 (sync.Mutex on shared state) |
| Memory safety (DoS) | Content-Length capped at 10 MB |
| Error discards | 53 → 10 benign (continuous counters) |
| Orphan packages | 0 (4 deleted) |
| Stub packages | 0 (7 documented) |

Full report: [docs/audit/complete-repo-audit.md](docs/audit/complete-repo-audit.md)

## Documentation

- [Installation](docs/install.md)
- [Quickstart](docs/quickstart.md)
- [CLI Reference](docs/cli-reference.md)
- [OpenCode Integration](docs/opencode-integration.md)
- [MCP Reference](docs/mcp-reference.md)
- [Architecture](docs/architecture.md)
- [Project Structure](docs/project-structure.md)
- [Manual Validation](docs/manual-validation.md)
- [Repository Audit](docs/audit/complete-repo-audit.md)
- [ADRs](docs/adr/) — 26 architectural decision records

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), and [SECURITY.md](SECURITY.md).

The [AGENTS.md](AGENTS.md) file contains the inviolable rule: this VPS is for building only, never for testing Plan-AI. All tests run in sandbox. A separate VPS exists for integration testing.

## License

MIT — see [LICENSE](LICENSE).

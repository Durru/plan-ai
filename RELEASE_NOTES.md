# Release Notes

## v1.0.0 — Planning-AI MVP

### Overview

First stable release of Plan-AI: a local-first continuous implementation planning engine for AI-assisted software projects.

### What's included

**Core engines (13):**
- Vision engine — ingestion, discovery sessions, draft/approve/finalize
- Context engine — approved context management, L0-L4 delivery
- Research engine — entries, findings, sources, conclusions
- Knowledge engine — reusable technical knowledge base
- Planning engine — master plans, specific plans, implementation docs
- Change engine — change detection, impact analysis, project snapshots
- Ingestion engine — input classification and storage
- Scanner engine — deterministic project scanning
- Continuous planning — event detection, update proposals
- Agent system — intent detection, routing, delegation
- Workflow engine — execution registry
- Orchestrator — job queue and orchestration
- Model strategy — LLM provider registry, budget tracking

**Interfaces:**
- CLI (20+ commands via Cobra)
- MCP server (30 tools via stdio JSON-RPC)
- OpenCode integration (6 generated artifacts)

**Storage:**
- Two-tier SQLite (global + project stores)
- 22 schema migrations
- 10+ repository implementations

**Quality:**
- Zero release-risk TODO/FIXME/HACK/TEMP/STUB/MOCK markers in active source/scripts
- All tests pass (`go test ./...`)
- All vet checks pass (`go vet ./...`)
- All builds pass (`go build ./...`)
- Sandbox validation passes (`bash scripts/test-sandbox.sh`)

### Breaking changes

None (first release).

### Upgrade notes

No upgrade path from previous versions (first release).

### Known limitations

- CLI output is text-only (no JSON output flag)
- MCP server runs in stdio mode only (no TCP/HTTP transport)
- Continuous planning requires manual trigger (no background daemon)
- Agent system is rule-based (no ML model)

## v2.0.0 — Plan-AI V2: User-Intent-Driven Planning Lifecycle

### Overview

Plan-AI V2 converts the MVP from a generic planning system into a user-intent-driven planning lifecycle. The system now converts ambiguous project ideas into approved vision, specialized research, compressed context, master plans, implementation packages, change impact analysis, and continuous evolution — while preserving complete MVP behavior.

V2 is purely additive over MVP. No existing modules, tables, commands, docs, or MCP tools were modified.

### What's added

**17 new phases (34–50):**

**User Truth Layer (Stage A):**
- Phase 34 — User Intent Engine: `plan-ai intent detect/show/approve`
- Phase 35 — Vision Engine: 5-dimension vision documents, discovery sessions
- Phase 36 — Approval Workflow: first-class approval lifecycle with audit trail
- Phase 40 — Requirement Discovery Engine: candidate requirement detection

**Evidence and Context Layer (Stage B):**
- Phase 38 — Research Orchestration: 7 specialized research agent roles
- Phase 39 — Reference Engine: URL/document/repository reference states
- Phase 37 — Context Delivery Engine: typed context packages with token budgets

**Planning and Implementation Layer (Stage C):**
- Phase 41 — Plan Generation V3 / Plan Evolution Engine: 13-section blueprints
- Phase 42 — Implementation Context Engine: model-targeted packages
- Phase 43 — Change Impact Engine V2: deep impact analysis with severity
- Phase 44 — Continuous Planning V2: targeted plan regeneration

**Agent and Integration Layer (Stage D):**
- Phase 45 — Subagent Orchestrator: isolated, temporary, auditable subagents
- Phase 46 — OpenCode Deep Integration: workflow commands surfaced in OpenCode
- Phase 47 — Project Memory System: durable decision memory with reuse
- Phase 48 — Model Compatibility Layer: 7 provider contracts

**Validation and Release (Stage E):**
- Phase 49 — Real Project Validation: 7 cases × 9 stages = 63 deterministic checks
- Phase 50 — V2 Release: complete documentation and audit

**New CLI commands (2):**
- `plan-ai validate v2` — Run all 63 V2 validation checks
- `plan-ai validate cases` — List all 7 project validation categories

### New modules
- `internal/validation/` — Deterministic V2 validation engine
- `internal/reference/` — External reference management
- `internal/memory/` — Project memory system
- `internal/modelstrategy/` — Provider compatibility contracts

### Feature matrix

| Category | Features |
|----------|----------|
| V2 Validation | 4 features (engine, cases, sandbox, tests) |
| **Total** | **82 features** (78 MVP + 4 V2) |

### Quality

- `go test ./internal/validation/` — 12 tests, 63 V2 checks, all pass
- `plan-ai validate v2` — 63/63 checks passed
- `scripts/test-sandbox.sh` — includes Phase 49-50 validation verification
- All existing MVP quality gates continue to pass

### Breaking changes

None. V2 is strictly additive over MVP Phases 0–33.

### Upgrade notes

If you have an existing Plan-AI MVP installation:
- Run `plan-ai install` and `plan-ai init` (idempotent) to ensure latest migrations
- New V2 commands are available immediately — no data migration required

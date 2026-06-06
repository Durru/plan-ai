# Plan-AI Problems To Fix

This file is the remediation backlog for the problems already found in Plan-AI. Use it to locate, prioritize, and fix the issues without rediscovering the same context.

## Comparative Diagnosis — Why Engram And Gentle-AI Work

The current remediation plan is grounded in a code-by-code comparison with two working references:

- Engram: `/root/engram`
- Gentle-AI: `/root/Gentle/gentle-ai`

The core lesson is direct: Plan-AI does not mainly fail because it lacks features. It fails because multiple parts of the system are competing to be the authority for MCP, OpenCode setup, generated artifacts, and runtime lifecycle. Engram and Gentle-AI work because each critical concern has one owner and one path.

### Engram Reference Pattern

Engram works because MCP is a first-class command on one binary, backed by an MCP SDK and one long-lived store lifecycle.

| Concern | Engram evidence | Plan-AI current state |
|---|---|---|
| One binary command | `/root/engram/cmd/engram/main.go` uses `engram mcp` | ✅ `plan-ai mcp serve` is the only MCP entry point; `cmd/mcp-server/` deleted |
| MCP SDK | `/root/engram/internal/mcp/mcp.go` uses `github.com/mark3labs/mcp-go` | ✅ `internal/mcp/sdk_server.go` uses `mark3labs/mcp-go`; custom `protocol.go` deprecated |
| Tool schema authority | Engram registers tools with `mcp.NewTool`, `mcp.WithString`, `mcp.Required` | ✅ `RegisterSDKDefaultTools` registers via SDK's `NewToolWithRawSchema` — SDK holds canonical schema; production path skips custom `JSONSchema` |
| Store lifecycle | Engram opens one Store for the MCP process and closes it on exit | ✅ `ServeSDKStdio` opens shared store; `openStore` returns cleanup func; handlers use `defer cleanup()` |
| OpenCode setup | `/root/engram/internal/setup/setup.go` installs plugin and injects `command: ["engram", "mcp"]` | ✅ Writes `["plan-ai", "mcp", "serve"]` from `config.MCPCommand()` single authority |

**Why this works:** Engram has a deep MCP module. Callers only need to know `engram mcp`; the implementation hides protocol, schema, tool dispatch, and store lifecycle behind a small interface.

### Gentle-AI Reference Pattern

Gentle-AI works because installation is a staged pipeline, not a set of unrelated file writes.

| Concern | Gentle-AI evidence | Plan-AI contrast |
|---|---|---|
| Install orchestration | `/root/Gentle/gentle-ai/internal/cli/run.go` resolves selection, builds stage plan, executes, verifies | Plan-AI installer manages state and config writes but lacks full pipeline semantics |
| Prepare/apply/rollback | `/root/Gentle/gentle-ai/internal/pipeline/orchestrator.go` runs prepare, apply, rollback | Plan-AI backs up some OpenCode files but has no general rollback pipeline |
| Backup snapshots | `/root/Gentle/gentle-ai/internal/backup/snapshot.go` and `restore.go` snapshot/restore target files | Plan-AI has narrower OpenCode config backups |
| Atomic writes | `/root/Gentle/gentle-ai/internal/components/filemerge/writer.go` writes through temp files and no-ops unchanged content | Plan-AI writes config files directly in several paths |
| Generated artifacts | `/root/Gentle/gentle-ai/internal/components/golden_test.go` checks stable golden outputs | Plan-AI tests existence/shape but not stable generated artifact contracts |

**Why this works:** Gentle-AI has a deep installer module. A caller asks for install; the implementation handles planning, backup, mutation, rollback, verification, and idempotency locally.

### Plan-AI Root Causes Found In Code

These are the concrete root causes that must guide fixes:

1. `internal/mcp/protocol.go` owns JSON-RPC framing, handshake, tool listing, and tool-call result formatting manually.
2. `cmd/mcp-server/main.go` creates a separate deployable instead of `plan-ai mcp serve`.
3. `internal/opencode/setup.go` and `internal/installer/sync.go` both write OpenCode MCP config with different behavior.
4. `internal/mcp/handlers.go` is a 1600-line transport module that also opens stores, creates domain entities, persists data, and formats responses.
5. `cmd/plan-ai/main.go` is a 5000-line Cobra monolith importing most of the application.
6. Runtime MCP tools, OpenCode registry entries, and docs are manually duplicated instead of generated from one source.
7. Tests currently protect the custom implementation instead of protecting the desired Engram/Gentle-style contracts.

### Required Architecture Direction

Use these reference patterns as the target shape:

1. MCP must become `plan-ai mcp serve` on the main binary.
2. MCP must use `mark3labs/mcp-go` instead of custom protocol code.
3. Runtime tool definitions must be the source for MCP exposure, OpenCode registry, and docs.
4. OpenCode setup must have one authority called by install, bootstrap, sync, and setup commands.
5. Installer behavior must move toward a staged pipeline: detect, plan, backup, apply, rollback, verify.
6. Generated OpenCode/agent/skill artifacts need golden tests.
7. Handlers and commands must become adapters over service modules, not places where business logic lives.

## Priority 0 — Product Must Work

### 1. MCP rearchitecture

**Problem:** Plan-AI still has a custom JSON-RPC/MCP implementation. Engram works reliably by using `mark3labs/mcp-go`, which handles protocol details, schema generation, capability negotiation, and stdio behavior.

**Why it matters:** MCP clients are strict. Custom protocol code creates drift and can break OpenCode or other clients through small schema/framing mismatches.

**Fix direction:** Replace the custom MCP protocol layer with `mark3labs/mcp-go` and keep handlers as thin wrappers over service code.

**Files to inspect:**
- `internal/mcp/protocol.go`
- `internal/mcp/tools.go`
- `internal/mcp/handlers.go`
- `cmd/mcp-server/main.go`

### 2. Single binary ✅ RESOLVED

`cmd/mcp-server/` was removed. MCP serving and debug commands
(`list-tools`, `call-tool`, `validate-tools`) are now subcommands
of `plan-ai mcp`. `config.MCPCommand()` is the single authority
used by both installer and opencode packages.

**Resolution:** PR [#unified-contract] — deleted `cmd/mcp-server/main.go`,
moved debug subcommands into `cmd/plan-ai/main.go`, updated all scripts
and docs.

**Files affected:**
- `cmd/plan-ai/main.go` — added list-tools, call-tool, validate-tools
- `internal/config/mcp.go` — shared MCPCommand(binDir)
- `internal/installer/sync.go` — delegates to config.MCPCommand
- `internal/opencode/setup.go` — delegates to config.MCPCommand
- `scripts/uninstall.sh`, `scripts/test-sandbox.sh`, `scripts/release-check.sh`

### 3. OpenCode setup duplication

**Problem:** OpenCode config generation/sync logic exists in more than one place.

**Why it matters:** Duplicate config writers create drift. One path can generate valid config while another writes stale or conflicting entries.

**Fix direction:** Choose one OpenCode config authority and make all install/bootstrap/sync flows call it.

**Files to inspect:**
- `internal/opencode/setup.go`
- `internal/installer/sync.go`
- `cmd/plan-ai/main.go`

### 4. Real MCP tool exposure

**Problem:** Minimal MCP exposure has previously limited Plan-AI to a small subset of tools.

**Why it matters:** OpenCode should see the actual Plan-AI capability surface, not an artificially constrained demo registry.

**Fix direction:** Keep minimal mode only as an explicit diagnostic/safe mode. Default behavior should expose validated production tools.

**Files to inspect:**
- `internal/mcp/tools.go`
- `internal/opencode/setup.go`
- `docs/mcp-reference.md`
- `docs/opencode-integration.md`

## Priority 1 — OpenCode Native Experience

### 5. OpenCode Foundation

**Problem:** Plan-AI is not yet packaged as a first-class OpenCode agent/skill experience.

**Why it matters:** The user interacts through OpenCode. If Plan-AI is only a CLI/MCP backend, the experience is incomplete.

**Fix direction:** Add OpenCode-native artifacts and generation commands.

**Needed outputs:**
- `.opencode/agents/plan-ai.md`
- `.opencode/skills/plan-ai-*/SKILL.md`
- `plan-ai init --opencode`
- Generated `AGENTS.md` for planned projects
- Ecosystem positioning vs `conductor`, `micode`, and `Agentic`

**Files to inspect:**
- `internal/opencode/setup.go`
- `internal/opencode/types.go`
- `docs/opencode-integration.md`
- `docs/opencode-integration-guide.md`
- `docs/plan-ai-v3-master-plan.md`

### 6. Multi-AI skill format decision ✅ DECIDED: Option B

**Decision:** Plan-AI will generate only Claude Code + OpenCode artifacts.
No canonical YAML format is added unless a concrete need for a 3rd
CLI arises.

**Rationale:**
- The existing installer already handles OpenCode integration.
- Claude Code is covered by AGENTS.md + patterns.
- `internal/skills/` remains empty/archived — no new abstraction.
- A canonical YAML format would be speculative work with unclear ROI.

**Next step:** Remove the `internal/skills/` placeholder entirely once
the capability registry boundaries are defined (domain model phase).

**Files to inspect:**
- `internal/skills/doc.go` — archived, `package skills` kept inactive

### 7. OpenCode Plugin

**Problem:** Plan-AI has no OpenCode plugin plan yet.

**Why it matters:** A plugin can add deeper workflow hooks, but should not come before stable CLI/MCP/skill integration.

**Fix direction:** Add this as a post-core phase, not a prerequisite.

**Files to inspect:**
- `docs/plan-ai-v3-master-plan.md`

## Priority 2 — Internal Intelligence

### 8. Internal Intelligence Engine

**Problem:** Plan-AI has planning, memory, knowledge, and V3 alignment artifacts, but no unified retrieval/intelligence layer.

**Why it matters:** The system cannot reliably answer: "what context matters right now?" without combining lexical search, graph relations, and optional semantic search.

**Fix direction:** Add a local hybrid engine:
- SQLite FTS5 for deterministic search
- Existing knowledge relations/references as graph layer
- Optional `chromem-go` vector/RAG layer
- Qdrant only as a future scale-out adapter

**Files to inspect:**
- `internal/store/store.go`
- `internal/store/memory_repository.go`
- `internal/store/knowledge_repositories.go`
- `internal/alignmentv3/service.go`
- `docs/plan-ai-v3-master-plan.md`

### 9. Replace LIKE search

**Problem:** Current memory and knowledge search relies on `LIKE`.

**Why it matters:** `LIKE` is not enough for project-scale retrieval, ranking, or agent context selection.

**Fix direction:** Add FTS5-backed indexes and repository methods for ranked search.

**Files to inspect:**
- `internal/store/memory_repository.go`
- `internal/store/knowledge_repositories.go`
- `internal/store/store.go`

## Priority 3 — Architecture Cleanup

### 10. Split Cobra monolith

**Problem:** `cmd/plan-ai/main.go` is too large and mixes command wiring, output rendering, seeding, and service orchestration.

**Why it matters:** Every new feature becomes harder to review and easier to break.

**Fix direction:** Split command groups into focused files/packages while preserving CLI behavior.

**Files to inspect:**
- `cmd/plan-ai/main.go`
- `cmd/plan-ai/installer_commands.go`

### 11. Thin MCP handlers

**Problem:** MCP handlers have historically mixed transport concerns with business logic and direct persistence.

**Why it matters:** The system becomes hard to test without MCP and hard to reuse from CLI/OpenCode flows.

**Fix direction:** Move behavior into service/core packages. Keep MCP handlers as adapters.

**Files to inspect:**
- `internal/mcp/handlers.go`
- `internal/mcp/tools.go`
- `internal/*/service.go`

### 12. Shared DB lifecycle

**Problem:** SQLite connection lifecycle has been identified as a risk in tool handling.

**Why it matters:** Opening/closing DB connections per tool call is inefficient and can cause lock/lifecycle bugs.

**Fix direction:** Use an explicit server/app lifecycle with shared DB handles where appropriate.

**Files to inspect:**
- `cmd/mcp-server/main.go`
- `internal/mcp/handlers.go`
- `internal/store/db.go`
- `internal/store/store.go`

### 13. Schema consolidation

**Problem:** Compatibility tables, views, and mirrored writes accumulated across phases.

**Why it matters:** Schema drift makes future changes risky and confuses the canonical data model.

**Fix direction:** Create a schema consolidation milestone before removing compatibility layers.

**Files to inspect:**
- `internal/store/store.go`
- `internal/store/migrations.go`
- `internal/migrations/project/`
- `docs/audit/technical-debt.md`

## Priority 4 — Deferred Product Gaps

### 14. Real provider adapters

**Problem:** Plan-AI has contracts for model/provider usage, but real provider execution is not connected.

**Why it matters:** Planning intelligence remains scaffolded until real provider adapters exist.

**Fix direction:** Implement provider adapters as a separate integration track after MCP/OpenCode foundations are stable.

**Files to inspect:**
- `docs/audit/gaps.md`
- `docs/audit/roadmap-adjustments.md`
- `internal/model*` or future provider packages

### 15. Engram deep integration

**Problem:** Engram is useful externally, but Plan-AI does not yet have a deep sync/query integration.

**Why it matters:** Important cross-session discoveries should become first-class planning context.

**Fix direction:** Define explicit sync/query boundaries instead of ad-hoc memory calls.

**Files to inspect:**
- `docs/audit/gaps.md`
- future `internal/intelligence/` or `internal/engram/`

### 16. Advanced Skill Intelligence

**Problem:** The old Skill Intelligence spike was archived.

**Why it matters:** Reintroducing it too early risks repeating the same architectural mistake.

**Fix direction:** Reintroduce only after core OpenCode/skills format and internal intelligence are stable.

**Files to inspect:**
- `docs/archive/old-phases/phase-7-skill-intelligence/`
- `docs/implementation/current-state-audit.md`

### 17. TUI

**Problem:** No TUI exists.

**Why it matters:** Useful later for human review, but not urgent for core correctness.

**Fix direction:** Keep deferred until CLI/MCP/OpenCode paths are reliable.

## Historical Corrections

### Phase 0 is not the functionalization block

**Phase 0:** clean foundation only.

**Files:**
- `docs/adr/0001-project-foundation.md`
- `docs/methodology/vision.md`
- `docs/implementation/current-state-audit.md`

**Functionalization block:** phases 23-27 converted the technical MVP into a functional planning product with user vision, real plan generation, reusable research, and intelligent context.

## Guiding Philosophy

Every fix must move Plan-AI closer to:

> "Lo que me imagine es lo que obtuve"

If a change does not improve intent-to-result alignment, reduce friction, or make planning more reliable, it is not urgent.

## Immediate Fix Order

1. MCP rearchitecture with `mark3labs/mcp-go`.
2. Single binary via `plan-ai mcp serve`.
3. Unify OpenCode setup/sync authority.
4. Add OpenCode Foundation artifacts and generation.
5. Add Internal Intelligence Engine with FTS5 first.
6. Split `cmd/plan-ai/main.go` by command group.
7. Consolidate schema/migration strategy.
8. Add real provider adapters.

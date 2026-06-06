# Plan-AI Complete Repository Audit

**Date:** 2026-06-06  
**Repository:** `github.com/Durru/plan-ai`  
**Commit:** `8706e21`

---

## 1. Executive Summary

| Metric | Count |
|--------|-------|
| Go packages | 33 |
| Non-test .go files | 210 |
| Test .go files | 83 |
| Test functions | 559 |
| CLI commands | ~120 (across 40+ subcommands) |
| MCP tools | 38 |
| Database tables | 80+ |
| ADR documents | 26 |
| Raw SQL in handlers | 0 |
| Raw SQL outside store | 6 (reuse/recorder) |
| Stub/near-empty files | 12 |
| Import cycles | 0 |

**Overall health:** `go build ./...` ✅, `go vet ./...` ✅, `go test ./...` 33/33 ✅

---

## 2. Critical Architectural Issues

### #1 — MCP Handlers Bypass Canonical Services (CRITICAL)
`internal/mcp/handlers.go` calls `repos.Plan`, `repos.Research`, `repos.Knowledge`, etc. directly — **never** calls `planning.Service`, `research.Service`, `knowledge.Service`. This violates all 6 v4 principles.

**Lines with direct repo access:** 38+ locations throughout handlers.go

### #2 — Two Parallel Type Systems (CRITICAL)
- **Canonical:** `planning.MasterPlan`, `research.ResearchEntry`, `vision.Draft`
- **Store-direct:** `domain.MasterPlan`, `domain.Research`, `domain.Vision`
- Both used concurrently. MCP writes domain types directly.

### #3 — Three Parallel Change Systems (HIGH)
- `change_requests` (repositories/change_repository.go)
- `change_events` (phase18_20_repositories.go)
- `change_impact_reports_v2` (v2_stage_d_repositories.go)

### #4 — Raw SQL in Service Layer (HIGH)
- `research/reuse.go` — 6 raw `db.Query/Exec` calls bypassing typed repos
- `memory/recorder.go` — raw `db.Query` for project_memory_v2

### #5 — Workflow Engine Does Nothing (HIGH)
- `dispatchStep` only does `log.Printf` — no real side effects
- 4 workflow files are step-name lists with zero logic

### #6 — Orchestrator Never Invoked (HIGH)
- `orchestrator/orchestrator.go` exists (165 LOC) but MCP and CLI never call it

### #7 — 6 approved_* Tables Not Unified (HIGH)
- `approved_requirements`, `approved_constraints`, `approved_decisions`, `approved_preferences`, `approved_references`, `approved_goals`

---

## 3. Phase Status Summary

| Phase | Description | % Complete | Key Missing |
|-------|-------------|------------|-------------|
| 1 | External Project Store | 100% | — |
| 2 | Install Once Lifecycle | 100% | — |
| 3 | Safe OpenCode ADR | 100% | — |
| 4 | Conversation Gateway | 100% | — |
| 5 | Discovery-First Guardrail | 100% | — |
| 6 | Approved Context Authority | 100% | — |
| 7 | Research Reuse | 85% | Freshness tracking |
| 8 | Permanent Memory Recorder | 100% | — |
| 9 | Continuous Planning Loop | 100% | — |
| 10 | Service-Backed MCP Handlers | 80% | remaining context handlers |
| 11 | Documentation | 60% | 6 ADRs missing |
| 12 | E2E Validation | 90% | Some seed-based bypasses |
| 13 | Research Intelligence Platform | 40% | Freshness, mandatory layer, decision proposals |
| 14 | Universal Level Model | 25% | Outcomes, 10-state CR, unified approved_context, change_audit |
| 15 | Architectural Authority | 10% | 38+ repo bypasses, intentv3 active, no v4 middleware |
| 16 | Impact/Workflows/Capabilities | 30% | Real workflow steps, orchestrator wiring, impact graph integration |

---

## 4. OpenCode Safety Audit

| Check | Result |
|-------|--------|
| SetupMCPConfig writes to correct file | ✅ opencode.json (merge), NOT config.json |
| invalidOpenCodeKeys | ✅ Empty — no provider/agent stripping |
| init --allow-real-opencode guard | ✅ Present (setup_commands.go:414) |
| Installer refuse check | ✅ `user.Current().HomeDir` (getpwuid_r) |
| Backup before write | ✅ WriteFileWithBackup on critical paths |
| Atomic writes | ✅ temp+rename on SetupMCPConfig |
| Test sandboxing | ✅ All tests use t.TempDir() or executeCommand |
| No test touches real config | ✅ Zero violations |

**Risk:** `GenerateProjectArtifacts` ignores `$OPENCODE_CONFIG_DIR`. Mitigated by CLI guard.

---

## 5. Package Inventory

### Core Services
| Package | Responsibility | Status |
|---------|---------------|--------|
| `agent/` | Intent detection, routing, delegation | ✅ Complete |
| `guard/` | Planning guardrail | ✅ Complete |
| `conversation/` | CLI+MCP gateway | ✅ Complete |
| `planning/` | Master/specific plans, evolution | ✅ Complete (stubs deleted) |
| `research/` | Research engine, reuse, orchestration | ⚠️ Raw SQL in reuse.go |
| `knowledge/` | Knowledge base, relations, tags | ✅ Complete |
| `vision/` | Vision drafts, documents | ✅ Complete |
| `context/` | Approved context, delivery engine, authority | ⚠️ Missing buildImplementationContext sections |
| `change/` | Change events, impact, snapshots, versioning | ❌ analyzer.go placeholder |
| `continuous/` | Event detection, proposals, loop, status | ✅ Complete |
| `memory/` | Durable project memory, recorder | ⚠️ Raw SQL in recorder.go |
| `intentv3/` | V3 Product Intent engine | ⚠️ Should be deleted (Phase 15) |
| `discoveryv3/` | Progressive discovery engine | ✅ Complete |
| `alertv3/`, `ambiguityv3/`, `confidencev3/` | Alignment/ambiguity/confidence | ✅ Complete |

### Infrastructure
| Package | Responsibility | Status |
|---------|---------------|--------|
| `store/` | SQLite schema, migrations, 14 sub-repos | ✅ Complete |
| `config/` | Paths, layout, config I/O | ✅ Complete |
| `domain/` | Canonical domain types | ⚠️ Decision missing supersession fields |
| `installer/` | Component-based install/sync/uninstall | ✅ Complete |
| `opencode/` | OpenCode integration, setup, doctor, detection | ⚠️ Artifacts bypass OPENCODE_CONFIG_DIR |
| `mcp/` | MCP server, tools, handlers, SDK | ❌ 38+ service bypasses |
| `atomicfile/` | Atomic write + backup primitives | ✅ Complete |
| `scanner/` | Project scanner (git, languages, deps) | ✅ Complete |
| `ingestion/` | Input ingestion pipeline | ⚠️ parser.go no-op |
| `impact/` | Impact graph (nodes, edges, traversal) | ✅ New (Phase 16) |
| `capabilities/` | Skill/capability registry | ⚠️ Defaults hardcoded, not DB-seeded |
| `workflows/` | Workflow engine, step execution | ❌ dispatchStep does nothing |
| `orchestrator/` | Job orchestrator | ❌ Never invoked |
| `modelstrategy/` | Model selection, provider catalog | ✅ Complete |
| `reference/` | Project references | ✅ Complete |
| `requirements/` | Requirement discovery | ✅ Complete |
| `validation/` | V2 validation framework | ✅ Complete |
| `skills/` | Placeholder (decided not to implement) | ⚠️ Should document or delete |
| `planner/` | Vestigial empty package | ❌ Should delete |
| `integrations/` | Empty package | ⚠️ Should document or delete |
| `core/` | App identity (unused) | ⚠️ Zero callers |

---

## 6. Database Table Map

### Schema Layer — Entity Tables
| Layer | Table | Column Count |
|-------|-------|-------------|
| 0 | visions | 8 |
| 2 | requirements | 7, constraints | 7 |
| 3 | research_entries | 14 (with ALTERs) |
| 4 | knowledge_objects | 13 (with ALTERs) |
| 5 | decisions | 10 (with supersedes) |
| 6 | master_plans, plans | 8 |
| 7 | specific_plans | 10 |
| 8 | implementation_documents | 9 |
| 9 | phases | 8 |
| 10 | tasks | 10 |

### Schema Layer — Supporting Tables
| Purpose | Tables |
|---------|--------|
| Approved Context | `approved_requirements`, `approved_constraints`, `approved_decisions`, `approved_preferences`, `approved_references`, `approved_goals` |
| Research | `research_jobs`, `research_findings`, `research_sources`, `research_conclusions`, `research_recommendations`, `research_tags`, `research_knowledge_links` |
| Knowledge | `knowledge_tags`, `knowledge_relations`, `knowledge_references`, `knowledge_links` |
| Change | `change_requests`, `change_events`, `change_reports`, `change_impact_reports_v2`, `change_audit` (missing) |
| Snapshots | `snapshots`, `snapshots_v2`, `entity_states` |
| Continuous | `continuous_status`, `continuous_events`, `plan_update_proposals`, `context_deliveries` |
| Agent | `agent_runs`, `agent_runs_v2`, `agent_messages`, `agent_delegated_jobs`, `subagent_tasks_v2` |
| Workflow | `workflow_runs`, `jobs`, `job_runs` |
| Context | `context_views`, `context_views_v2`, `context_chunks` |
| Impact | `impact_edges`, `entity_links` |
| Capabilities | `capabilities`, `capabilities_v2` |
| MCP | `mcp_tools`, `mcp_runs` |
| OpenCode | `opencode_detections`, `opencode_integration_state`, `opencode_doctor_checks` |

---

## 7. Fix Priority Matrix

| Priority | Issue | Affected Files | Effort |
|----------|-------|---------------|--------|
| P0 | `GenerateProjectArtifacts` bypass OPENCODE_CONFIG_DIR | artifacts.go | 5 min |
| P0 | Decision domain missing SupersedesID/SupersededByID | domain/decision.go | 10 min |
| P0 | research/reuse.go raw SQL | reuse.go | 15 min |
| P0 | memory/recorder.go raw SQL | recorder.go | 15 min |
| P1 | 6 approved_* tables → unified | store.go + migration | 30 min |
| P1 | 10-state ChangeRequest lifecycle | domain/change.go | 20 min |
| P1 | Delete vestigial planner/ package | planner/ | 5 min |
| P2 | Workflow steps real execution | workflows/execution.go | 30 min |
| P2 | buildImplementationContext complete sections | context/delivery.go | 20 min |
| P2 | Capabilities DB seed | capabilities/ | 15 min |
| P2 | Orchestrator integration with MCP | mcp/handlers.go | 30 min |
| P3 | MCP handlers → canonical services | handlers.go (38+ spots) | 2h+ |
| P3 | 3 change systems → 1 | change/ | 1h+ |
| P3 | 2 type systems → 1 | planning/ + domain/ | 3h+ |
| P3 | intentv3 delete/downgrade | intentv3/ | 30 min |

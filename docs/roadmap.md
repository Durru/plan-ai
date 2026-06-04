# Plan-AI Roadmap

This roadmap replaces the old phase sequence after Phase 6. The next implementation phase is **Phase 7 — Definitive Domain Model**.

| Phase | Objective | Implements | Expected result | Validation |
|---|---|---|---|---|
| Pre-Phase 7 — Audit and realignment | Stabilize repo after old Phase 6 and archive misaligned work | Audit docs, cleanup, new roadmap, sandbox verification | Clean baseline for the new architecture | `go test ./...`, `go build ./...`, sandbox script |
| Phase 7 — Definitive Domain Model | Define Plan-AI's canonical truth model | Core entities, relationships, statuses, approval concepts | Domain model usable by store/planner/context layers | Domain unit tests and docs |
| Phase 8 — Definitive Store Layer | Persist the canonical domain safely | SQLite schema, migrations, repositories, FTS where needed | Durable global/project stores | Migration/repository tests |
| Phase 9 — Ingestion Layer | Import project facts deterministically | Project scan inputs, docs ingestion, file metadata | Raw project evidence enters Plan-AI | Sandbox scan fixtures |
| Phase 10 — Vision Engine | Capture project/product direction | Vision records, goals, constraints, success criteria | Planning has a stable north star | Vision object tests and CLI checks |
| Phase 11 — Approved Context Engine | Separate approved truth from drafts | Approval records, context snapshots, provenance | Consumers can trust approved context | Approval boundary tests |
| Phase 12 — Research + Knowledge Engine | Rebuild research/knowledge on approved context | Research entries, knowledge objects, sources, confidence | Reusable evidence-backed project knowledge | Research/knowledge tests |
| Phase 13 — Planning Framework | Generate structured implementation preparation | Plans, phases, tasks, dependencies, validation paths | Human/agent-ready planning artifacts | Plan/task lifecycle tests |
| Phase 14 — Workflow Engine | Track work readiness and transitions | Next/done flows, blockers, review gates | Continuous planning workflow works locally | CLI workflow tests |
| Phase 15 — Model Strategy Layer | Manage LLM usage as temporary intelligence | Model profiles, prompts, budgets, provenance | LLM usage is configurable and auditable | Config/unit tests; no hard-coded provider |
| Phase 16 — Orchestrator + Capability Registry | Know what Plan-AI can do and when | Capabilities, optional skill intelligence, routing metadata | Extensible local capability model | Registry tests |
| Phase 17 — Context Engine | Serve right-sized context | Context levels, selection rules, summaries | Agents/tools get minimal useful context | Context contract tests |
| Phase 18 — Change Engine | Detect and classify project changes | Diffs, impact analysis inputs, replan triggers | Plan-AI knows when plans become stale | Change fixture tests |
| Phase 19 — Definitive MCP | Expose Plan-AI over stdio MCP | MCP server, tools, resources, context endpoints | External tools can query Plan-AI | MCP protocol tests |
| Phase 20 — OpenCode Integration | Integrate with OpenCode safely | Setup command, optional config generation, detection | OpenCode can discover Plan-AI without manual hacks | Sandbox OpenCode config tests |
| Phase 21 — Plan-AI Agent | Provide optional agent behavior | Agent prompt/config, optional subagents | Agent can use Plan-AI as source of truth | Agent artifact tests |
| Phase 22 — Continuous Planning | Close the loop from change to replan | Monitor/update cycles, exports, snapshots | Plan-AI supports ongoing project planning | End-to-end sandbox flow |

## Current rule

Do not implement phases out of order unless the roadmap is explicitly changed. Capability/skill work belongs after the definitive domain and store layers, not before them.

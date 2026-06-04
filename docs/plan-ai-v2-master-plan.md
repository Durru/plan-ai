# Plan-AI V2: Post-MVP Master Plan

**Status:** ✅ Completed — all 17 phases implemented, validated, and released  
**Scope:** Phases 34–50  
**Baseline:** Plan-AI MVP Phases 0–33 are complete and immutable as compatibility foundation.

---

## 1. Executive Summary

Plan-AI V2 turns the MVP from a generic planning system into a user-intent-driven planning lifecycle.

The V2 system must convert ambiguous project ideas into approved vision, specialized research, compressed context, master plans, implementation packages, change impact analysis, and continuous evolution while preserving the complete MVP behavior.

V2 is not a rewrite. It is an additive compatibility layer over the MVP.

### Non-Negotiable Rules

1. Do not replace MVP modules.
2. Do not delete existing tables, commands, docs, or MCP tools.
3. Do not break sandbox validation or existing release-candidate checks.
4. Add new migrations only from `0026_*` onward.
5. Any schema break requires a `_v2` table plus compatibility views.
6. Every generated decision, requirement, architecture choice, plan, reference, and change must be traceable and approvable.

---

## 2. Compatibility Foundation

### 2.1 MVP Systems V2 Must Extend

| V2 Area | MVP Foundation | Rule |
|---|---|---|
| Intent | `internal/agent/`, `internal/capabilities/` | Extend intent classification; do not replace existing agent status/routing. |
| Vision | `internal/vision/`, `vision_discovery_sessions`, `vision_approvals` | Add richer vision dimensions over current draft/discovery flow. |
| Approval | `internal/workflows/`, `internal/continuous/approval.go` | Generalize approval states without breaking proposal approvals. |
| Context | `internal/context/`, `context_delivery_*` | Reconcile L0–L4 names first, then add typed context packages. |
| Research | `internal/research/`, `internal/workflows/research_workflow.go` | Add orchestration and agent roles; results still end in Knowledge. |
| References | Knowledge and decision repositories | New module; references must link into knowledge, vision, and decisions. |
| Planning | `internal/planning/`, `master_plan_versions`, `specific_plan_versions` | Avoid naming collision with existing Master/Specific Plan V2. |
| Change | `internal/change/`, `internal/continuous/` | Add deeper impact reports, not a second change engine. |
| Subagents | `internal/agent/`, `internal/orchestrator/` | Make delegation real but temporary, isolated, and auditable. |
| OpenCode | `internal/opencode/`, MCP tools | Extend artifacts and commands; never mutate real user OpenCode config without explicit sandbox/env path. |
| Memory | snapshots, knowledge, decisions, plans | Add memory graph/indexing without replacing current stores. |
| Models | `internal/modelstrategy/` | Add provider contracts and output schemas; model reasons, Plan-AI controls context. |

### 2.2 Storage and Migration Rules

- Continue project migrations from `0026_*`.
- Use `CREATE TABLE IF NOT EXISTS` for new tables.
- Use `ALTER TABLE ADD COLUMN ... DEFAULT ...` for compatible extensions.
- Use `_v2` only when the shape cannot remain backward-compatible.
- Add compatibility views when old and new names must coexist.
- Keep inline migrations in `internal/store/store.go` as the runtime source of truth.
- Mirror SQL files under `internal/migrations/project/` for reviewability.
- Keep global store and project store separation:
  - Global: `~/.plan-ai/global.db`
  - Project: `<project>/.plan-ai/project.db`

### 2.3 CLI, MCP, and OpenCode Extension Rules

#### CLI

Prefer extending existing command groups:

- `vision`
- `approved`
- `research`
- `knowledge`
- `plan`
- `agent`
- `continuous`
- `setup`

Add new top-level groups only when the concept is genuinely cross-cutting, such as:

- `intent`
- `reference`
- `memory`

#### MCP

- Register tools in `internal/mcp/tools.go`.
- Implement handlers in `internal/mcp/handlers.go`.
- Preserve `plan_ai.<verb>_<noun>` naming.
- All V2 tools must return structured errors instead of fake success.

#### OpenCode

Extend `internal/opencode/setup.go` artifacts:

- `opencode.json`
- `mcp-registry.json`
- `agents/plan-ai.json`
- `profiles.json`
- `prompts.json`
- `.plan-ai/opencode-sync.json`

OpenCode deep integration must remain sandbox-safe by default and respect:

- `OPENCODE_CONFIG_DIR`
- `PLAN_AI_PROJECT_ROOT`
- `PLAN_AI_HOME`

---

## 3. Known V2 Risks and Required Mitigations

| Risk | Problem | Mitigation |
|---|---|---|
| Plan Generation V2 name collision | MVP already has Master Plan V2 and Specific Plan V2. | Internally call Phase 41 **Plan Generation V3 / Plan Evolution Engine** while keeping user-facing V2 release language. |
| Context level drift | L0–L4 terminology appears in multiple places. | Phase 37 starts with canonical context taxonomy and compatibility aliases. |
| Thin MVP agent processing | Agent process currently does not represent full delegated runtime. | Phase 45 builds real isolated job lifecycle beside existing agent status commands. |
| Migration numbering mismatch | ADR and migration numbering diverged. | Continue migrations from `0026`; document numbering in each ADR. |
| Duplicate research/knowledge flows | Research orchestration could bypass Knowledge. | Phase 38 must persist every validated result into Knowledge. |
| OpenCode mutation risk | Deep integration can accidentally touch real config. | All tests must run under `OPENCODE_CONFIG_DIR`; real config writes require explicit user path. |

---

## 4. V2 Phase Specifications

### Phase 34 — User Intent Engine

**Goal:** Understand what the user really wants to build before planning.

**Extends:** `internal/agent/`, `internal/capabilities/`, ingestion outputs.

**New concepts:**

- `Intent`
- `Goal`
- `UserExpectation`
- `SuccessCriteria`
- `UserPriority`

**Output:** `Intent Profile`

**CLI:** `plan-ai intent detect`, `plan-ai intent show`, `plan-ai intent approve`

**MCP:** `plan_ai.detect_intent`, `plan_ai.get_intent_profile`

**Acceptance:** Given “quiero un SaaS CRM”, Plan-AI identifies likely domains such as SaaS, CRM, multi-user, admin panel, subscriptions, reports, and automations as candidate intent signals without treating them as approved requirements.

---

### Phase 35 — Vision Engine

**Goal:** Define what the user wants to see, experience, operate, and sell before plans are generated.

**Extends:** `internal/vision/`, `vision_discovery_sessions`, `vision_assumptions`, `vision_ambiguities`, `vision_approvals`.

**Vision dimensions:**

- Functional vision
- Visual vision
- Technical vision
- Operational vision
- Business vision

**Output:** Approved `Vision Document`

**Rule:** Plan generation remains blocked until the required vision sections are approved or explicitly deferred.

---

### Phase 36 — Approval Workflow

**Goal:** Make approval a first-class lifecycle for all important project facts.

**Extends:** `internal/workflows/`, continuous proposal approval, vision approvals.

**States:**

- Draft
- Review
- Clarification
- Approved
- Rejected

**Applies to:**

- Vision
- Requirements
- Constraints
- Architecture
- Functionalities
- References
- Plan changes

**Acceptance:** Every decision stores status, approver metadata, timestamp, rationale, and audit trail.

---

### Phase 37 — Context Delivery Engine

**Goal:** Serve the smallest useful context package for a model or implementation task.

**Extends:** `internal/context/`, `context_delivery_sessions`, `context_delivery_usage`, `context_delivery_budgets`.

**Context types:**

- Vision Context
- Research Context
- Planning Context
- Implementation Context
- Change Context

**Model targets:** GPT, Claude, Gemini, Qwen, DeepSeek, OpenRouter, OpenAI-compatible APIs.

**Acceptance:** Context output is prioritized, compressed, bounded, and traceable to stored facts.

---

### Phase 38 — Research Orchestration

**Goal:** Coordinate specialized research agents and persist evidence into Knowledge.

**Extends:** `internal/research/`, `internal/workflows/research_workflow.go`, Knowledge repositories.

**Subagents:**

- Market Research
- Technical Research
- Architecture Research
- UI Research
- UX Research
- Security Research
- Infrastructure Research

**Acceptance:** Each research run stores topic, evidence, confidence, sources, conclusions, and resulting Knowledge entries.

---

### Phase 39 — Reference Engine

**Goal:** Work with real external or user-provided references.

**New module:** `internal/reference/`

**Inputs:**

- URLs
- Images
- Documents
- Repositories
- Examples
- Screenshots

**Reference states:** approved, rejected, needs-review.

**Reference categories:** visual, UX, functional, technical, business.

**Acceptance:** “quiero algo parecido a Stripe” stores explicit reference records without copying or assuming implementation details.

---

### Phase 40 — Requirement Discovery Engine

**Goal:** Detect missing functionality, ambiguity, and dependencies before planning.

**Extends:** `internal/vision/discovery.go`, `internal/vision/extraction.go`, `internal/ingestion/`.

**Acceptance:** For “quiero ecommerce”, Plan-AI proposes candidate requirements such as cart, checkout, coupons, SEO, blog, analytics, inventory, payments, tax, and fulfillment, but requires approval before adding them to scope.

---

### Phase 41 — Plan Generation V3 / Plan Evolution Engine

**Goal:** Generate implementation-ready plans without requiring extra research at implementation time.

**Extends:** `internal/planning/`, existing Master Plan V2 and Specific Plan V2 tables.

**Plan sections:**

- Objective
- Scope
- Exclusions
- Dependencies
- Stack
- Versions
- Libraries
- Folder structure
- Files
- Validations
- Tests
- Risks
- Rollback

**Acceptance:** Plans reference approved intent, vision, requirements, research, and constraints.

---

### Phase 42 — Implementation Context Engine

**Goal:** Generate an implementation package for AI coding agents.

**Extends:** `internal/context/`, implementation documents.

**Implementation Package contains:**

- What to do
- How to do it
- Files to touch
- Files not to touch
- Examples
- Commands
- Validations
- Rollback notes

**Targets:** GPT, Claude, Gemini, Qwen, DeepSeek, OpenCode.

---

### Phase 43 — Change Impact Engine V2

**Goal:** Analyze changes deeply across architecture, backend, migrations, docs, APIs, plans, and validations.

**Extends:** `internal/change/`, snapshots, impact reports.

**Example:** PostgreSQL → MariaDB produces affected plans, schema changes, API concerns, migration tasks, rollback strategy, and validation commands.

---

### Phase 44 — Continuous Planning V2

**Goal:** Keep plans alive without regenerating the whole project.

**Extends:** `internal/continuous/`.

**Capabilities:**

- Detect changes
- Regenerate affected plan sections
- Keep consistency
- Maintain snapshots
- Require approvals

**Potential CLI:** `plan-ai continuous daemon`

**Acceptance:** Targeted plan regeneration works for bounded changes and preserves unaffected sections.

---

### Phase 45 — Subagent Orchestrator

**Goal:** Delegate specialized work while keeping the main agent lean.

**Extends:** `internal/agent/`, `internal/orchestrator/`, capabilities registry.

**Agent types:**

- Research Agent
- Architecture Agent
- UI Agent
- UX Agent
- Security Agent
- Backend Agent
- Database Agent
- Validation Agent

**Rules:**

- Temporary
- Isolated
- No independent persistent memory
- Results return to Plan-AI store
- Every result has provenance and validation status

---

### Phase 46 — OpenCode Deep Integration

**Goal:** Make OpenCode consume Plan-AI as a real planning backend.

**Extends:** `internal/opencode/`, MCP setup, CLI command surface.

**OpenCode workflows:**

- Read status
- Read next task
- Read context
- Read plans
- Read changes
- Update progress

**Commands surfaced:**

- `plan-ai status`
- `plan-ai next`
- `plan-ai context`
- `plan-ai plans`
- `plan-ai changes`

**Acceptance:** Sandbox OpenCode config can invoke Plan-AI MCP/CLI workflows without touching real OpenCode config.

---

### Phase 47 — Project Memory System

**Goal:** Prevent repeated questions by storing durable project memory.

**New module:** `internal/memory/`

**Stores:**

- Decisions
- Approvals
- Questions
- Answers
- References
- Research
- Plans
- Change history

**Acceptance:** Previously answered questions are reused with citation to stored memory.

---

### Phase 48 — Model Compatibility Layer

**Goal:** Make Plan-AI work with multiple LLM providers while keeping Plan-AI in control.

**Extends:** `internal/modelstrategy/`.

**Providers:**

- GPT
- Claude
- Gemini
- DeepSeek
- Qwen
- OpenRouter
- OpenAI-compatible APIs

**Rule:** The model reasons; Plan-AI controls context, memory, research, planning, approvals, and persistence.

---

### Phase 49 — Real Project Validation ✅

**Goal:** Validate V2 against realistic project categories.

**Implementation:** Deterministic in-memory validation engine at `internal/validation/v2_validation.go` with 7 project cases × 9 V2 stages = 63 checks. Pure rule-based, no external calls.

**CLI:** `plan-ai validate v2`, `plan-ai validate cases`.

**Tests:** 12 Go tests covering cases, stages, pass/fail counts, intent detection, and benchmarks.

**Cases:**

- SaaS
- Ecommerce
- Landing Page
- MCP Server
- Mobile App
- API
- CRM

**Flow:** idea → intent → vision → approvals → research → plans → implementation context → change → updated plan.

**Verification:** `go test ./internal/validation/` passes, `plan-ai validate v2` reports 63/63 passed.

---

### Phase 50 — Plan-AI V2 Release ✅

**Goal:** Release V2 with complete docs, architecture, workflows, and audit.

**Docs:**

- Architecture guide
- OpenCode guide
- MCP guide
- Workflow guide
- Subagent guide
- Context Engine guide
- Planning Engine guide
- Memory guide
- Model compatibility guide

**Final audit:**

- Functional
- Architecture
- Performance
- Consistency
- Context
- Memory
- LLM compatibility

**Artifacts updated:**
- `README.md` — added validate command to CLI section
- `FEATURE_MATRIX.md` — added V2 Validation section (4 features), updated totals
- `FINAL_AUDIT_REPORT.md` — added V2 Release Candidate section
- `RELEASE_NOTES.md` — added v2.0.0 section
- `docs/cli-reference.md` — added validate command reference
- `docs/plan-ai-v2-master-plan.md` — marked phases 49-50 as complete
- `scripts/test-sandbox.sh` — added validate v2 and validate cases verification

**Verification:**
- `go build ./...` passes
- `go test ./...` passes
- `go vet ./...` passes
- `scripts/test-sandbox.sh` updated with Phase 49-50 validation

---

## 5. Dependency Graph

```text
34 Intent
  ↓
35 Vision → 36 Approval
  ↓          ↓
40 Requirement Discovery
  ↓
38 Research Orchestration → 39 Reference Engine
  ↓                         ↓
37 Context Delivery Engine ←┘
  ↓
41 Plan Generation V3
  ↓
42 Implementation Context
  ↓
43 Change Impact V2
  ↓
44 Continuous Planning V2
  ↓
45 Subagent Orchestrator
  ↓
46 OpenCode Deep Integration
  ↓
47 Project Memory
  ↓
48 Model Compatibility
  ↓
49 Real Project Validation
  ↓
50 V2 Release
```

Approval gates apply after phases 35, 36, 40, 41, 42, 43, 44, and 49.

---

## 6. Implementation Strategy

### Stage A — User Truth Layer

- Phase 34 Intent
- Phase 35 Vision
- Phase 36 Approval
- Phase 40 Requirement Discovery

**Outcome:** Plan-AI understands and records what the user actually wants before planning.

### Stage B — Evidence and Context Layer

- Phase 38 Research Orchestration
- Phase 39 Reference Engine
- Phase 37 Context Delivery Engine

**Outcome:** Plan-AI can gather evidence and serve minimal model-specific context.

### Stage C — Planning and Implementation Layer

- Phase 41 Plan Generation V3
- Phase 42 Implementation Context
- Phase 43 Change Impact V2
- Phase 44 Continuous Planning V2

**Outcome:** Plan-AI creates implementation-ready plans and updates only affected parts.

### Stage D — Agent and Integration Layer

- Phase 45 Subagent Orchestrator
- Phase 46 OpenCode Deep Integration
- Phase 47 Project Memory
- Phase 48 Model Compatibility

**Outcome:** Plan-AI becomes a durable planning backend for AI agents and multiple LLMs.

### Stage E — Validation and Release

- Phase 49 Real Project Validation
- Phase 50 V2 Release

**Outcome:** V2 is validated on real project archetypes and released with complete docs.

---

## 7. Validation Criteria

Every phase must pass:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
```

Every phase must also include:

- Additive migration or explicit “no migration needed”.
- CLI validation.
- MCP validation when a tool is added.
- Sandbox validation.
- Documentation update.
- No regression of MVP feature matrix.
- Approval workflow coverage when project truth changes.

---

## 8. Approval Checklist

Before implementing V2, approve or revise these decisions:

- [ ] V2 builds additively on MVP 0–33.
- [ ] Phase 41 is internally named **Plan Generation V3 / Plan Evolution Engine** to avoid MVP naming collision.
- [ ] Migrations continue from `0026_*`.
- [ ] Context taxonomy reconciliation is mandatory before expanding Context Delivery.
- [ ] OpenCode deep integration remains sandbox-safe by default.
- [ ] No V2 planning output becomes implementation scope until approved by the Approval Workflow.

---

## 9. Next Step After Approval

After this document is approved, create an implementation plan for Stage A:

1. Phase 34 — User Intent Engine
2. Phase 35 — Vision Engine
3. Phase 36 — Approval Workflow
4. Phase 40 — Requirement Discovery Engine

Stage A should be implemented first because every later V2 capability depends on knowing what the user actually wants and what has been approved.

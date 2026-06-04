# Plan-AI V3: Intent Alignment System

**Status:** 🔲 Planned — design phase (this document)  
**Scope:** Phases 51–70  
**Baseline:** Plan-AI MVP Phases 0–33 + Plan-AI V2 Phases 34–50 (both complete)  
**Primary Metric:** Alignment level between final product/output and original user intent

---

## 1. Executive Summary

Plan-AI V3 transforms the system from **continuous planning** into **continuous alignment** between real user intent and the built product.

V2 solved "what does the user want?" — it captured intent, built vision, orchestrated research, and generated plans. V3 solves "is what we're building still what the user wanted?" — it makes intent the durable, measurable north star that every artifact, plan, decision, and review is checked against.

The fundamental shift:

| V2 (Planning) | V3 (Alignment) |
|---|---|
| Capture intent once, then plan | Intent is a living artifact — tracked, measured, reconciled |
| Plans derived from approved context | Plans continuously checked against intent |
| Change impact on plans | Change impact on alignment level |
| Validation of features | Validation of intent coverage |
| "Did we build it right?" | "Did we build the right thing?" |

### Non-Negotiable Rules

1. V3 is additive over both MVP and V2 — zero breaking changes to existing commands, tables, docs, or MCP tools.
2. New migrations continue from `0040_*` (V2's last migration was `0039_*`).
3. Every new entity, decision, and artifact must be traceable to an **approved intent**.
4. No plan or task is generated without an approved intent reference.
5. Once intent is approved, it must never be re-asked — the system stores and reuses it.
6. The alignment metric must be quantitative, not just qualitative (e.g., percentage coverage, drift distance).
7. Avoid asking everything upfront — progressive intent discovery is preferred over one-shot specification.
8. Every generated artifact must be reviewable by the user before it becomes implementation scope.

---

## 2. Compatibility Foundation

### 2.1 V2 Systems V3 Must Extend

| V3 Area | V2 Foundation | Rule |
|---|---|---|
| Intent | Phase 34 User Intent Engine | Extend intent profile with lifecycle, versioning, and formalization — do not replace existing `intent detect/show/approve` |
| Vision | Phase 35 Vision Engine + Phase 36 Approval | Intent becomes the upstream of vision; vision continues to exist as strategic expression of intent |
| Approval | Phase 36 Approval Workflow | Add intent dimension to all approval states; approval becomes alignment-aware |
| Research | Phase 38 Research Orchestration | Add intent-tagged research; research findings include alignment relevance scores |
| References | Phase 39 Reference Engine + Phase 50 Reference tables | References include intent alignment metadata; references checked against product identity |
| Requirements | Phase 40 Requirement Discovery | Requirements explicitly linked to intent dimensions; each requirement measurable against intent |
| Planning | Phase 41 Plan Generation V3 / Plan Evolution Engine | Plans include alignment sections; plan drift detection added; plan approval requires intent check |
| Context | Phase 37 Context Delivery Engine | Context packages include intent alignment summary |
| Change | Phase 43 Change Impact V2 | Impact reports include alignment impact score |
| Memory | Phase 47 Project Memory | Memory entries tagged with intent provenance |
| Validation | Phase 49 Real Project Validation | New V3 validation engine with alignment checks |

### 2.2 Storage and Migration Rules

- Continue project migrations from `0040_*`.
- Use `CREATE TABLE IF NOT EXISTS` for new tables.
- Use `ALTER TABLE ADD COLUMN ... DEFAULT ...` for compatible extensions.
- Use `_v3` suffix only when the shape cannot remain backward-compatible.
- Keep inline migrations in `internal/store/store.go` as runtime source of truth.
- Mirror SQL files under `internal/migrations/project/` for reviewability.
- Two-tier store remains: Global `~/.plan-ai/global.db` + Project `<project>/.plan-ai/project.db`.

### 2.3 CLI, MCP, and Integration Extension Rules

**CLI:**
- Prefer extending existing command groups: `intent`, `vision`, `plan`, `validate`, `context`.
- New top-level groups only for genuinely cross-cutting V3 concepts: `alignment`, `traceability`, `coverage`.

**MCP:**
- Register tools in `internal/mcp/tools.go` with `plan_ai.<verb>_<noun>` naming.
- Implement handlers in `internal/mcp/handlers.go`.
- All new tools return structured errors.

**OpenCode:**
- Extend `internal/opencode/setup.go` artifacts.
- Deep integration remains sandbox-safe.
- Respect `OPENCODE_CONFIG_DIR`, `PLAN_AI_PROJECT_ROOT`, `PLAN_AI_HOME`.

---

## 3. V3 Phase Specifications

### Stage A: Intent Foundation (Phases 51–56)

Make intent the first-class, measurable, living truth source that every downstream artifact is measured against.

---

#### Phase 51 — Intent Capture & Formalization

**Goal:** Transform V2's basic intent detection into structured, formalized intent profiles that support lifecycle management, measurement, and traceability.

**Extends:** `internal/agent/`, Phase 34 Intent Engine.

**New concepts:**
- `IntentProfile` — structured document with dimensions, priorities, scope boundaries, success criteria
- `IntentDimension` — a single axis of intent (e.g., UX simplicity, data privacy, mobile-first, admin power)
- `IntentSignal` — raw evidence that contributed to the profile (user utterance, document, reference)
- `IntentMetadata` — source, confidence, timestamp, version

**Output:** Formal `IntentProfile` with structured dimensions, each with weight/priority and success criteria.

**CLI:** Extends `plan-ai intent` with:
- `plan-ai intent formalize` — upgrade detected intent to structured profile
- `plan-ai intent profile` — show full formalized intent profile
- `plan-ai intent dimension add|list|remove` — manage intent dimensions

**MCP:** `plan_ai.formalize_intent`, `plan_ai.get_intent_profile`, `plan_ai.add_intent_dimension`

**Acceptance:** Given detected intent "quiero un SaaS CRM", formalization produces a structured profile with dimensions (multi-user, subscription billing, reporting, permissions), each with weight, evidence, and success criteria — without re-asking the user for information already captured.

---

#### Phase 52 — Intent Lifecycle Management

**Goal:** Give intent a proper lifecycle — versions, states, approvals, deprecation — so it can evolve with the project without losing history.

**Extends:** Phase 36 Approval Workflow, Phase 47 Project Memory.

**Intent states:**
- `discovered` — raw detection, not yet formalized
- `draft` — profile exists but unapproved
- `active` — approved and actively tracked
- `superseded` — replaced by a newer version
- `archived` — no longer relevant, but kept for audit

**Output:** Versioned intent history with approvals, rationale, and deprecation chain.

**CLI:**
- `plan-ai intent approve` (extends existing)
- `plan-ai intent version` — list intent version history
- `plan-ai intent deprecate` — mark intent version as superseded

**MCP:** `plan_ai.approve_intent`, `plan_ai.list_intent_versions`

**Acceptance:** Intent can transition through all states. Deprecation preserves the full profile for audit and alignment history calculations.

---

#### Phase 53 — Alignment Metrics Engine

**Goal:** Define and compute a quantitative **alignment level** between any artifact (plan, task, implementation, validation) and the originating intent.

**Extends:** Phase 36 Approval Workflow, `internal/validation/`.

**New concepts:**
- `AlignmentScore` — 0.0–1.0 score computed per artifact against intent
- `AlignmentDimensionScore` — per-dimension breakdown of the score
- `AlignmentGap` — specific dimension where coverage is missing or weak
- `AlignmentThreshold` — minimum acceptable score per artifact type

**Core metric formula (first approximation):**
```
AlignmentScore = weighted_mean(dimension_coverage × dimension_priority)
```

Where `dimension_coverage` = how well the artifact addresses that intent dimension (0.0–1.0).

**CLI:**
- `plan-ai alignment score` — compute alignment score for current state
- `plan-ai alignment score --artifact <type> --id <id>` — score a specific artifact
- `plan-ai alignment overview` — summary of all alignment scores

**MCP:** `plan_ai.compute_alignment_score`, `plan_ai.get_alignment_overview`

**Acceptance:** Given a project with intent profile containing 3 dimensions (weights 0.5, 0.3, 0.2), running alignment score returns a 0.0–1.0 value with per-dimension breakdown that updates when artifacts change.

---

#### Phase 54 — Intent Drift Detection

**Goal:** Proactively detect when plans, tasks, decisions, or output drift away from approved intent dimensions.

**Extends:** Phase 43 Change Impact V2, Phase 44 Continuous Planning V2, `internal/continuous/`.

**New concepts:**
- `DriftEvent` — detected drift with affected dimensions, magnitude, trend
- `DriftMagnitude` — quantitative measure of how far from intent (0.0 = no drift, 1.0 = complete misalignment)
- `DriftTrend` — direction of drift over time (stable, increasing, decreasing)

**Drift sources:**
- Change requests that reduce alignment coverage
- New decisions that contradict intent dimensions
- Plan modifications that drop intent-aligned tasks
- Scope creep not covered by any intent dimension

**CLI:**
- `plan-ai alignment drift` — show current drift status
- `plan-ai alignment drift --history` — show drift over time

**MCP:** `plan_ai.detect_drift`, `plan_ai.get_drift_history`

**Acceptance:** When a plan modification drops a task tagged with a critical intent dimension, a DriftEvent is generated with magnitude > 0 and trend = "increasing". Drift events integrate into continuous planning proposals.

---

#### Phase 55 — Intent Reconciliation

**Goal:** Handle multiple, overlapping, or conflicting intents — reconcile them into a coherent alignment baseline.

**Extends:** Phase 36 Approval Workflow, Phase 50 Reference Engine.

**New concepts:**
- `IntentConflict` — detected contradiction between two intent dimensions
- `ReconciliationStrategy` — how conflicts are resolved (merge, prioritize, defer)
- `WeightedPriority` — relative importance when intents compete

**Conflict types:**
- Scope conflict (build both mobile app AND admin panel — resource constraint)
- Priority conflict (security first vs. time-to-market first)
- Constraint conflict (budget limit vs. feature scope)
- Technical conflict (React vs. Svelte based on different intents)

**CLI:**
- `plan-ai intent reconcile` — detect and guide reconciliation
- `plan-ai intent reconcile --resolve` — apply reconciliation strategy

**MCP:** `plan_ai.reconcile_intents`, `plan_ai.list_intent_conflicts`

**Acceptance:** When two approved intents have conflicting dimensions, reconciliation detects the conflict, proposes strategies, and stores the resolution as a decision in Project Memory.

---

#### Phase 56 — Intent-Based Decision Engine

**Goal:** Every architectural, technical, and design decision must be explicitly linked to at least one intent dimension — with an alignment justification.

**Extends:** Phase 36 Approval Workflow, `internal/domain/decision.go`, Phase 47 Project Memory.

**New concepts:**
- `IntentLinkedDecision` — decision with `intent_dimension_id`, `alignment_rationale`, `alternative_impact` (how alternatives would score)
- `DecisionAlignmentRequirement` — policy that certain decision types require intent linking

**CLI:**
- `plan-ai decision add --intent-dimension <id> --rationale <text>` — add intent-linked decision
- `plan-ai decision alignment` — show all decisions and their intent links

**MCP:** `plan_ai.add_intent_linked_decision`, `plan_ai.list_linked_decisions`

**Acceptance:** Every decision recorded after Phase 56 requires an intent dimension ID and rationale. Unlinked decisions are flagged with a warning. The alignment overview includes decision coverage.

---

### Stage B: Traceability & Consistency (Phases 57–61)

Trace every artifact back to intent, and ensure consistency across the entire chain.

---

#### Phase 57 — Bidirectional Traceability

**Goal:** Every artifact (plan, task, decision, research, validation) is traceable upstream to intent AND downstream to implementation output.

**Extends:** Phase 34–36 (Intent/Vision/Approval), Phase 41–42 (Plan/Impl Context), Phase 47 Project Memory.

**New concepts:**
- `TraceLink` — immutable record: source entity → target entity with trace type
- `TraceGraph` — full graph of all trace links from intent to implementation
- `TraceCoverage` — percentage of intent dimensions with complete downstream traces

**Trace types:**
- `implements` — artifact directly fulfills an intent dimension
- `informs` — artifact provides evidence or context for intent
- `constrains` — artifact limits or bounds intent
- `validates` — artifact checks intent fulfillment
- `derives` — artifact is derived from intent

**CLI:**
- `plan-ai traceability graph` — show trace graph
- `plan-ai traceability path --from <id> --to <id>` — trace path between any two entities
- `plan-ai traceability coverage` — show trace coverage percentage

**MCP:** `plan_ai.get_trace_graph`, `plan_ai.get_trace_path`

**Acceptance:** Given a complete project, running trace graph shows intent → vision → requirements → plans → tasks → validations as a connected graph. Coverage reports percents per dimension.

---

#### Phase 58 — Consistency Engine

**Goal:** Ensure plans, decisions, and scope remain internally consistent and externally consistent with approved intent.

**Extends:** Phase 44 Continuous Planning V2, Phase 43 Change Impact V2.

**New concepts:**
- `ConsistencyRule` — declarative rule that checks consistency between entities
- `ConsistencyViolation` — detected violation with affected entities and severity
- `ConsistencyReport` — aggregate report of all violations

**Rule categories:**
- **Intent-Plan consistency:** Every plan section must reference an intent dimension
- **Decision-Intent consistency:** Every decision must be compatible with all active intent dimensions
- **Plan-Plan consistency:** Plans at same level must not contradict each other
- **Scope-Intent consistency:** All scope items must trace to at least one intent dimension

**CLI:**
- `plan-ai consistency check` — run all consistency rules
- `plan-ai consistency rules` — list registered rules
- `plan-ai consistency violations` — show active violations

**MCP:** `plan_ai.check_consistency`, `plan_ai.list_consistency_violations`

**Acceptance:** After adding a decision that contradicts an intent dimension, consistency check reports a violation with severity, affected entities, and suggested remediation.

---

#### Phase 59 — Provenance Chain

**Goal:** Every project fact stores its full provenance — where it came from, what intent it serves, who approved it, and what it replaced.

**Extends:** Phase 36 Approval Workflow, Phase 47 Project Memory, `internal/domain/`.

**New concepts:**
- `ProvenanceRecord` — single provenance entry with source, timestamp, actor, intent link
- `ProvenanceChain` — ordered sequence of provenance records for an entity
- `ProvenanceGraph` — full provenance network across all entities

**Provenance metadata:**
- `source_type`: user_input, ingestion, research, agent, plan_generation, user_approval
- `source_detail`: specific command, file, URL, or agent that produced it
- `intent_link`: which intent dimension(s) this provenance serves
- `supersedes`: which previous entity this replaces (if any)

**CLI:**
- `plan-ai provenance get <entity-type> <id>` — get provenance chain for an entity
- `plan-ai provenance graph` — show full provenance graph
- `plan-ai provenance search --query <q>` — search provenance by source or intent

**MCP:** `plan_ai.get_provenance`, `plan_ai.search_provenance`

**Acceptance:** Every entity created after Phase 59 has a provenance chain. Querying provenance on a requirement shows: created from ingestion → refined in vision → approved by user → linked to intent dimension "multi-user".

---

#### Phase 60 — Gap Analyzer

**Goal:** Detect intent dimensions that have no or insufficient coverage in current plans, tasks, decisions, or validations.

**Extends:** Phase 53 Alignment Metrics, Phase 57 Traceability.

**New concepts:**
- `CoverageGap` — intent dimension with below-threshold coverage
- `GapSeverity` — critical, major, minor based on dimension weight and gap size
- `GapRemediation` — suggested action to close the gap (new plan, task, research, decision)

**CLI:**
- `plan-ai alignment gaps` — list all coverage gaps sorted by severity
- `plan-ai alignment gaps --dimension <id>` — detailed gap for one dimension
- `plan-ai alignment gaps --remediate` — generate remediation proposals

**MCP:** `plan_ai.list_coverage_gaps`, `plan_ai.suggest_gap_remediation`

**Acceptance:** For an intent dimension "mobile-first" with no tasks tagged, gap analysis reports a critical gap. After generating a task that covers it, gap severity drops or gap is removed.

---

#### Phase 61 — Coverage Reports

**Goal:** Generate structured, exportable alignment coverage reports for stakeholders, agents, and audit.

**Extends:** Phase 53–60, `internal/validation/`.

**Report types:**
- **Executive Summary:** alignment score, drift status, top gaps, recommendations
- **Per-Dimension Report:** each intent dimension with coverage, drift, trace completeness
- **Full Trace Report:** end-to-end trace from intent → implementation with scores
- **Audit Report:** provenance, approvals, changes over time for compliance

**CLI:**
- `plan-ai coverage report [--type executive|dimension|trace|audit]` — generate report
- `plan-ai coverage report --export <file> --format markdown|json` — export to file

**MCP:** `plan_ai.generate_coverage_report`, `plan_ai.export_coverage_report`

**Acceptance:** Running `plan-ai coverage report --type executive` produces a document with alignment score, drift trend, top 3 gaps, and actionable recommendations. JSON export is machine-parseable.

---

### Stage C: Planning/Task Alignment (Phases 62–64)

Make the planning engine alignment-aware — every plan and task explicitly serves intent dimensions.

---

#### Phase 62 — Intent-Aligned Task Generation

**Goal:** Generate tasks that explicitly reference which intent dimension they serve, with an estimated alignment contribution.

**Extends:** Phase 41 Plan Generation V3 / Plan Evolution Engine, Phase 42 Implementation Context Engine.

**New concepts:**
- `AlignmentTag` — task-level tag linking to intent dimension
- `AlignmentContribution` — estimated 0.0–1.0 contribution of this task to fulfilling the dimension
- `CriticalTask` — task whose contribution is essential for an intent dimension

**Rules:**
- Every generated task must carry at least one AlignmentTag.
- Tasks without AlignmentTag are flagged as `unattached` and excluded from alignment scoring.
- Critical intent dimensions must have at least one CriticalTask in each phase.

**CLI:**
- `plan-ai plan generate --align` — generate alignment-tagged tasks
- `plan-ai plan alignment-tasks` — show tasks grouped by intent dimension

**MCP:** `plan_ai.generate_aligned_tasks`, `plan_ai.get_alignment_tagged_tasks`

**Acceptance:** Plan generation produces tasks where each task includes `alignment_tags: [{dimension_id, contribution}]`. Tasks covering all intent dimensions with at least critical coverage.

---

#### Phase 63 — Implementation Progress Tracking

**Goal:** Track alignment score changes as implementation progresses — measure whether the project is converging toward or diverging from intent.

**Extends:** Phase 53 Alignment Metrics, Phase 62 Aligned Tasks.

**New concepts:**
- `AlignmentSnapshot` — point-in-time alignment score across all dimensions
- `AlignmentTrend` — direction over time: converging, stable, diverging
- `AlignmentMilestone` — target alignment score for a phase or release

**CLI:**
- `plan-ai alignment track` — show alignment trend
- `plan-ai alignment milestone set --score <0.0-1.0>` — set target milestone
- `plan-ai alignment history` — show alignment over time

**MCP:** `plan_ai.track_alignment`, `plan_ai.get_alignment_history`

**Acceptance:** After implementing 3/10 tasks tagged with dimension "multi-user", alignment score for that dimension increases proportionally. Trend shows "converging" when scores increase over snapshots.

---

#### Phase 64 — Adaptive Plan Regeneration

**Goal:** When alignment drift is detected, regenerate only the affected plan sections to restore alignment — without touching aligned sections.

**Extends:** Phase 44 Continuous Planning V2, Phase 54 Drift Detection, Phase 62 Aligned Tasks.

**New concepts:**
- `TargetedRegenerationRequest` — what to regenerate and why (which intent dimension, which drift)
- `RegenerationBoundary` — defines what is in/out of scope for regeneration
- `AlignmentPreservationRule` — sections already aligned must not be changed

**Rules:**
- Regeneration is triggered only when drift exceeds threshold for a dimension.
- Aligned sections are frozen — regeneration never touches them.
- User must approve regeneration proposals before they are applied.

**CLI:**
- `plan-ai plan regenerate --dimension <id>` — trigger targeted regeneration for a dimension
- `plan-ai plan regenerate --status` — show pending regeneration requests

**MCP:** `plan_ai.request_plan_regeneration`, `plan_ai.list_pending_regenerations`

**Acceptance:** When dimension "mobile-first" drifts below alignment threshold, regeneration request is created affecting only mobile-related plan sections. Approved regeneration updates tasks without changing non-mobile sections.

---

### Stage D: Product Identity & References (Phases 65–68)

Connect product identity and external references back to intent — ensuring the product position stays aligned.

---

#### Phase 65 — Product Identity Model

**Goal:** Define a formal product identity derived from intent dimensions — capturing what the product IS and IS NOT based on approved intent.

**Extends:** Phase 35 Vision Engine, Phase 51 Intent Formalization.

**New concepts:**
- `ProductIdentity` — structured definition: positioning statement, scope boundaries, personality traits
- `IdentityBoundary` — explicit what-we-do / what-we-dont-do based on intent
- `PersonalityTrait` — derived from intent: professional, playful, enterprise, minimalist, etc.
- `PositioningStatement` — one-sentence product identity

**CLI:**
- `plan-ai identity show` — show current product identity
- `plan-ai identity derive` — derive identity from active intent dimensions
- `plan-ai identity boundaries` — show scope boundaries

**MCP:** `plan_ai.get_product_identity`, `plan_ai.derive_identity`

**Acceptance:** Given intent with dimensions (enterprise, security-first, simple UX), product identity derivation produces positioning statement, identity boundaries (what features are excluded), and personality traits (professional, secure, minimalist).

---

#### Phase 66 — Reference-Intent Consistency

**Goal:** Every external reference (URL, document, screenshot, example repo) must be checked against product identity and intent. References that contradict intent are flagged.

**Extends:** Phase 39 Reference Engine, Phase 65 Product Identity.

**New concepts:**
- `ReferenceAlignmentCheck` — automated check: does this reference align with or contradict intent?
- `ReferenceConflictFlag` — reference marked as potentially misaligned
- `ReferenceUseJustification` — why this reference is being used (which intent dimension it serves)

**CLI:**
- `plan-ai reference check-alignment` — check all references against intent
- `plan-ai reference flag <id> --reason <text>` — flag a reference as misaligned

**MCP:** `plan_ai.check_reference_alignment`, `plan_ai.flag_reference`

**Acceptance:** Adding a reference to a "consumer-social-app" when intent is "enterprise-crm" generates an alignment warning with explanation of which intent dimensions are contradicted.

---

#### Phase 67 — Intent-Based Validation

**Goal:** Validate implementation output against original intent — not just against requirements or tests.

**Extends:** Phase 49 Validation, Phase 53 Alignment Metrics.

**New concepts:**
- `IntentValidationCheck` — check that maps implementation behavior to intent fulfillment
- `IntentValidationReport` — report: does the implementation satisfy the original intent?
- `IntentTestSuggestion` — suggested test that directly validates an intent dimension

**CLI:**
- `plan-ai validate intent` — run intent-based validation
- `plan-ai validate intent --dimension <id>` — validate a specific dimension

**MCP:** `plan_ai.validate_intent_alignment`, `plan_ai.suggest_intent_tests`

**Acceptance:** `plan-ai validate intent` produces a report mapping each intent dimension to implementation evidence. Dimension "multi-user" shows: covered by auth system (+), but missing role-based access control (-).

---

#### Phase 68 — Product-Intent Synchronization

**Goal:** Keep product identity synchronized with evolving intent — when intent changes, product identity updates automatically and flags affected plans/decisions.

**Extends:** Phase 52 Intent Lifecycle, Phase 64 Adaptive Regeneration, Phase 65 Product Identity.

**New concepts:**
- `IdentityChangeEvent` — triggered when intent evolution affects product identity
- `SyncImpact` — which plans, decisions, references, and validations are affected by the identity shift
- `MigrationPath` — suggested steps to realign product identity with new intent

**CLI:**
- `plan-ai identity sync` — sync product identity with current intent
- `plan-ai identity sync --impact` — show impact of sync
- `plan-ai identity sync --apply` — apply sync (requires approval)

**MCP:** `plan_ai.sync_product_identity`, `plan_ai.get_identity_sync_impact`

**Acceptance:** When an intent dimension "enterprise-only" is deprecated and "smb-friendly" is added, identity sync detects the change, generates IdentityChangeEvent, produces impact report (affected plans, decisions, references), and proposes migration path.

---

### Stage E: Review Framework & Release (Phases 69–70)

Systematic review and formal release of V3.

---

#### Phase 69 — Alignment Review Framework

**Goal:** Provide a structured, repeatable alignment review process that can be run at any point — producing actionable alignment improvement plans.

**Extends:** Phase 53–68, Phase 49 Validation, `internal/validation/`.

**Review dimensions:**
1. **Intent completeness** — are all user intents captured and formalized?
2. **Alignment coverage** — how well does the current state cover each dimension?
3. **Drift status** — is alignment improving or degrading over time?
4. **Trace completeness** — are all artifacts traceable to intent?
5. **Consistency** — are all artifacts internally and externally consistent?
6. **Provenance** — does every fact have a complete provenance chain?
7. **Gap severity** — what are the most critical unmet intent dimensions?
8. **Product identity fit** — does the product identity still match intent?

**Output:** `AlignmentReviewReport` with per-dimension scores, overall alignment health, top 5 risks, and prioritized remediation plan.

**CLI:**
- `plan-ai alignment review` — run full alignment review
- `plan-ai alignment review --quick` — fast check (no deep trace)
- `plan-ai alignment review --export <file>` — export report

**MCP:** `plan_ai.run_alignment_review`, `plan_ai.get_review_report`

**Acceptance:** Running `plan-ai alignment review` produces a report with 8 dimension scores (0–100), overall health score, top 5 risks, and a ranked list of remediation actions with estimated effort.

---

#### Phase 70 — Plan-AI V3 Release

**Goal:** Release V3 with complete documentation, validation, and audit.

**Validation criteria:**
1. All V3 CLI commands implemented and working.
2. All V3 MCP tools registered and functional.
3. `plan-ai validate v3` — new validation suite (cases × V3 stages).
4. Documentation updated for all new features.
5. No regression of MVP or V2 feature matrix.
6. Sandbox validation passes for V3 workflows.
7. Zero release-risk markers in active source/scripts.

**Documentation:**
- `docs/plan-ai-v3-master-plan.md` — updated (this doc, phases 69–70 marked ✅)
- `docs/cli-reference.md` — all new CLI flags/commands documented
- `docs/architecture.md` — alignment engine layer added to architecture
- `docs/alignment-system.md` — new: alignment metrics, drift detection, traceability
- `docs/coverage-reports.md` — new: report types and formats
- `docs/intent-lifecycle.md` — new: intent lifecycle management
- `README.md` — updated with V3 capabilities
- `FEATURE_MATRIX.md` — V3 section added
- `RELEASE_NOTES.md` — v3.0.0 section added
- `FINAL_AUDIT_REPORT.md` — V3 release candidate section
- `scripts/test-sandbox.sh` — V3 workflows verified

**Quality gates:**
- `go build ./...` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
- `gofmt -d .` — no diffs
- `scripts/test-sandbox.sh` — PASS with V3 extensions
- `scripts/release-check.sh` — PASS (if created)

**Feature matrix update:**
- New features added to `FEATURE_MATRIX.md` under "V3 Intent Alignment" category.
- Total feature count updated.

---

## 4. Dependency Graph

```text
51 Intent Formalization
  ↓
52 Intent Lifecycle
  ↓
53 Alignment Metrics
  ↓                    ↓
54 Drift Detection    56 Intent-Based Decisions
  ↓                    ↓
55 Intent Reconciliation
  ↓
57 Bidirectional Traceability
  ↓
58 Consistency Engine
  ↓
59 Provenance Chain
  ↓
60 Gap Analyzer
  ↓
61 Coverage Reports
  ↓
62 Aligned Task Generation
  ↓
63 Progress Tracking → 64 Adaptive Plan Regeneration
  ↓
65 Product Identity
  ↓
66 Reference-Intent Consistency
  ↓
67 Intent-Based Validation
  ↓
68 Product-Intent Synchronization
  ↓
69 Alignment Review Framework
  ↓
70 V3 Release
```

**Approval gates apply after phases:** 51, 52, 55, 56, 57, 62, 64, 65, 68, 69.

---

## 5. Implementation Strategy

### Stage A — Intent Foundation (Phases 51–56)

| Phase | Area | Effort Estimate |
|-------|------|----------------|
| 51 | Intent Capture & Formalization | Medium — new structs, CLI extension, migration 0040 |
| 52 | Intent Lifecycle Management | Medium — status machine, version tracking, migration 0041 |
| 53 | Alignment Metrics Engine | High — core metric design, scoring algorithms, migration 0042 |
| 54 | Intent Drift Detection | Medium — drift computation, continuous integration, migration 0043 |
| 55 | Intent Reconciliation | Medium — conflict detection, resolution strategies, migration 0044 |
| 56 | Intent-Based Decision Engine | Medium — decision linking, CLI, migration 0045 |

**Outcome:** Intent is first-class, measurable, and drives all downstream decisions. Alignment score exists as a live metric.

### Stage B — Traceability & Consistency (Phases 57–61)

| Phase | Area | Effort Estimate |
|-------|------|----------------|
| 57 | Bidirectional Traceability | High — trace graph, trace links storage, CLI, migration 0046 |
| 58 | Consistency Engine | Medium — rules engine, violation detection, migration 0047 |
| 59 | Provenance Chain | Medium — provenance records, chain construction, migration 0048 |
| 60 | Gap Analyzer | Medium — gap computation, remediation suggestions, migration 0049 |
| 61 | Coverage Reports | Medium — report generation, export formats, migration 0050 |

**Outcome:** Every artifact is traceable to intent. Consistency is enforced. Gaps are visible.

### Stage C — Planning/Task Alignment (Phases 62–64)

| Phase | Area | Effort Estimate |
|-------|------|----------------|
| 62 | Intent-Aligned Task Generation | Medium — alignment tags, task generation rules, migration 0051 |
| 63 | Implementation Progress Tracking | Medium — alignment snapshots, trends, milestones, migration 0052 |
| 64 | Adaptive Plan Regeneration | High — targeted regeneration, boundaries, preservation rules, migration 0053 |

**Outcome:** Plans and tasks are alignment-aware. Drift triggers targeted regeneration.

### Stage D — Product Identity & References (Phases 65–68)

| Phase | Area | Effort Estimate |
|-------|------|----------------|
| 65 | Product Identity Model | Medium — identity derivation, boundaries, CLI, migration 0054 |
| 66 | Reference-Intent Consistency | Small — alignment checks, flagging, migration 0055 |
| 67 | Intent-Based Validation | Medium — validation checks, intent tests, CLI, migration 0056 |
| 68 | Product-Intent Synchronization | Medium — identity change events, sync impact, migration 0057 |

**Outcome:** Product identity is explicitly derived from and synchronized with intent.

### Stage E — Review Framework & Release (Phases 69–70)

| Phase | Area | Effort Estimate |
|-------|------|----------------|
| 69 | Alignment Review Framework | Medium — 8-dimension review, report generation, CLI |
| 70 | V3 Release | Medium — docs, validation, feature matrix, release notes |

**Outcome:** V3 is validated, documented, and released.

---

## 6. Data Model / Migration Expectations

### New Tables (Project Store)

| Migration | Table | Purpose |
|-----------|-------|---------|
| 0040 | `intent_profiles` | Formalized intent profiles with dimensions |
| 0040 | `intent_dimensions` | Individual intent dimensions with weights |
| 0041 | `intent_versions` | Versioned intent lifecycle |
| 0042 | `alignment_scores` | Computed alignment scores snapshots |
| 0043 | `drift_events` | Detected drift records |
| 0044 | `intent_conflicts` | Detected/reconciled intent conflicts |
| 0045 | `intent_linked_decisions` | Decisions with intent dimension references |
| 0046 | `trace_links` | Bidirectional trace links between entities |
| 0047 | `consistency_violations` | Detected consistency violations |
| 0048 | `provenance_records` | Provenance chain entries |
| 0049 | `coverage_gaps` | Detected coverage gaps |
| 0050 | `coverage_reports` | Generated coverage reports |
| 0051 | `alignment_tags` | Task-level alignment tags |
| 0052 | `alignment_snapshots` | Point-in-time alignment state |
| 0053 | `regeneration_requests` | Targeted regeneration requests |
| 0054 | `product_identities` | Product identity records |
| 0055 | `reference_alignment_checks` | Reference alignment check results |
| 0056 | `intent_validations` | Intent-based validation results |
| 0057 | `identity_change_events` | Product identity sync events |

### Extended Tables (ALTER TABLE ADD COLUMN)

| Table | New Column | Migration |
|-------|-----------|-----------|
| `decisions` | `intent_dimension_id` | 0045 |
| `tasks` | `alignment_tags` (JSON) | 0051 |
| `plans` | `alignment_section` (JSON) | 0051 |
| `references` | `intent_alignment_status` | 0055 |
| `validations` | `intent_dimension_id` | 0056 |

### Global Store

Minimal changes. May add `global_alignment_thresholds` table in migration 0042 for cross-project alignment baselines.

---

## 7. CLI Expectations

### New Top-Level Command Groups

- `plan-ai alignment` — alignment metrics, drift, review, tracking
- `plan-ai traceability` — trace graph, path, coverage
- `plan-ai coverage` — coverage reports, gaps

### Extended Existing Command Groups

- `plan-ai intent` — formalize, profile, dimension, version, deprecate, reconcile
- `plan-ai decision` — add with intent link, alignment
- `plan-ai plan` — generate with alignment, regenerate, alignment-tasks
- `plan-ai reference` — check-alignment, flag
- `plan-ai validate` — intent, v3
- `plan-ai identity` — show, derive, boundaries, sync

### Existing Commands That Remain Unchanged

All MVP and V2 commands remain untouched. No `--force` flags, no renamed subcommands, no removed flags.

---

## 8. Sandbox / Test Expectations

### Test Categories

1. **Unit tests** — per new package/function
2. **Alignment metric tests** — verify score computation with known inputs
3. **Drift detection tests** — verify drift is detected when intent-task links are broken
4. **Traceability tests** — verify trace links are created and queryable
5. **Consistency tests** — verify violations are detected
6. **Provenance tests** — verify chains are correct
7. **CLI integration tests** — verify all new CLI commands
8. **MCP tool tests** — verify all new MCP tools
9. **Validation tests** — V3 validation suite

### Sandbox Scenarios

Extend `scripts/test-sandbox.sh` with:

1. **Intent lifecycle scenario:** detect → formalize → approve → version → deprecate
2. **Alignment scenario:** compute score → detect drift → generate aligned tasks → track progress
3. **Traceability scenario:** create trace links → query trace graph → compute coverage
4. **Product identity scenario:** derive identity → check references → sync identity
5. **Full V3 review scenario:** run alignment review → generate report → export

---

## 9. Acceptance Criteria Per Phase

Every phase passes these gates:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
```

Plus phase-specific criteria:

| Phase | Specific Acceptance |
|-------|---------------------|
| 51 | Structured intent profile with dimensions, weights, and evidence (no re-ask) |
| 52 | Intent lifecycle: discovered → draft → active → superseded → archived |
| 53 | AlignmentScore computed per artifact, per dimension, with formula documented |
| 54 | DriftEvent generated when plan drops intent-aligned task ( > threshold) |
| 55 | IntentConflict detected + ResolutionStrategy stored |
| 56 | Every decision has intent_dimension_id + alignment_rationale |
| 57 | TraceGraph shows complete intent→implementation paths; TraceCoverage reported |
| 58 | ConsistencyViolation generated for decision contradicting intent |
| 59 | ProvenanceChain available for every entity; queryable by source |
| 60 | CoverageGap list sorted by severity; remediation suggestions available |
| 61 | exportable CoverageReport in markdown + JSON formats |
| 62 | Every generated task has AlignmentTag; unattached tasks flagged |
| 63 | AlignmentSnapshot + Trend available; milestones settable and trackable |
| 64 | TargetedRegenerationRequest affects only drifting sections; aligned sections frozen |
| 65 | ProductIdentity derived from intent; IdentityBoundaries explicit |
| 66 | ReferenceAlignmentCheck warns on contradiction; ReferenceConflictFlag actionable |
| 67 | IntentValidationReport per dimension; IntentTestSuggestion generated |
| 68 | IdentityChangeEvent on intent change; SyncImpact report with MigrationPath |
| 69 | AlignmentReviewReport with 8 dimension scores, health, top 5 risks, remediation |
| 70 | All V3 CLI/MCP/tests pass; docs complete; zero regressions; feature matrix updated |

---

## 10. Risks and Guardrails

### Critical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Scope creep** — V3 tries to solve too much | Delays, incomplete phases | 20 phases with hard scope boundaries; each phase has specific acceptance criteria |
| **Metric oversimplification** — AlignmentScore reduces complex judgment to a single number | False precision, gaming the metric | Score is per-dimension AND aggregate; always show breakdown; document formula limitations |
| **User fatigue** — asking too many questions upfront | Users abandon the tool | Rule: never ask the same intent question twice; progressive discovery; batch questions |
| **Performance** — trace graph grows large | Slow queries on large projects | Indexed trace links; pagination; lazy graph loading |
| **Drift noise** — too many drift events for small changes | Alert fatigue | Configurable drift thresholds; minimum magnitude to trigger event; trend-based filtering |
| **Consistency over-enforcement** — rigid rules block legitimate evolution | User frustration | Consistency violations are warnings by default; require user approval to enforce as blockers |
| **Provenance storage cost** — every entity gets provenance chain | Database size growth | Provenance has configurable retention; archival strategy for old provenance |

### Non-Negotiable Guardrails

1. **No plan without approved intent.** Generating any plan, task, or implementation document without an approved intent reference is a hard error.
2. **Once approved, never re-ask.** The system must check Project Memory before asking any intent-related question. Repeated questions for the same information are a bug.
3. **Progressive disclosure.** Do not prompt the user for all 20 phases upfront. Reveal questions as the project progresses and context accumulates.
4. **V3 is additive.** Breaking MVP or V2 commands, tables, or behaviors is a release blocker.
5. **Every new artifact is reviewable.** No automatically-generated plan, task, or decision becomes implementation scope without user review.
6. **Alignment score is transparent.** The user can always see how the score is computed — formula, dimension weights, and per-dimension contributions.
7. **Drift is observable.** All drift events are logged and visible. No silent course correction.
8. **Retention limits apply.** Provenance, drift events, and alignment snapshots have configurable retention to prevent unbounded storage growth.

---

## 11. Next Step After Approval

After this document is approved, create an implementation plan for **Stage A — Intent Foundation** (Phases 51–56).

**Recommended first implementation slice:** Phase 51 (Intent Capture & Formalization) + Phase 52 (Intent Lifecycle Management).

**Rationale:** Every subsequent V3 phase depends on having formalized intent with a proper lifecycle. Without Phase 51/52, there is no alignment to measure, no drift to detect, and no traceability to build. These two phases form the foundational layer that the entire V3 architecture sits on.

**Stage A delivers the critical primitive:** a `plan-ai intent formalize` command that turns raw user expressions into structured, versioned, approvable intent profiles — the single source of truth for all alignment calculations to follow.

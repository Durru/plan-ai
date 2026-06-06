# Remaining Implementation Phases

**Status:** Implementation backlog
**Source plan:** `docs/implementation/external-storage-conversation-first-plan.md`
**Product vision:** `docs/architecture/plan-ai-product-vision.md`
**Architecture specs:**
- `docs/architecture/Plan-AI_Niveles_y_Gestion_de_Cambios.docx` (Universal Level Model)
- `docs/architecture/PlanAI_Especificacion_Arquitectonica_Oficial_v4.docx` (21-layer official spec)

## Approved Direction

Plan-AI will become a global, conversation-first planning layer.

Default project storage will be external:

```txt
~/.plan-ai/projects/<project-id>/project.db
```

Project-local storage remains available only through explicit opt-in:

```sh
plan-ai init --local
```

Legacy local stores must not be used silently.

## Phase 1: External Project Store And Global Registry

### Objective

Move the default project source of truth from `<repo>/.plan-ai` to the global Plan-AI home.

### Why This Comes First

Every later feature depends on stable project identity. Conversation, memory, research reuse, approved context, and continuous planning all need to know which project they are operating on without relying on a mutable filesystem path.

### Current Gaps

- `config.ProjectDir(projectDir)` resolves to `<project>/.plan-ai`.
- `store.EnsureProjectLayout(projectRoot)` creates project-local directories.
- `openInitializedProjectStore()` opens `<project>/.plan-ai/project.db`.
- `store.ProjectID(rootPath)` uses a path-derived identity.
- MCP handlers open project-local stores through `store.OpenProjectStore(projectRoot)`.

### Implement

- Add global project registry tables.
- Add `ProjectRegistryRepository`.
- Add `ProjectResolver`.
- Add external project layout helpers.
- Make `plan-ai init` create/open external project stores by default.
- Add `plan-ai init --local` for explicit project-local mode.
- Add detection for legacy `<repo>/.plan-ai/project.db`.
- Add `plan-ai migrate local-to-global` as the migration path.
- Make CLI and MCP use the same project resolver.

### Key Files

- `internal/config/config.go`
- `internal/store/store.go`
- `internal/store/project_store.go`
- `internal/store/global_store.go`
- `cmd/plan-ai/main.go`
- `cmd/plan-ai/setup_commands.go`
- `internal/mcp/handlers.go`

### Tests

- `TestInitCommandDoesNotWriteProjectLocalByDefault`
- `TestEnsureExternalProjectLayoutCreatesExpectedPaths`
- `TestRegisterProjectCreatesStableID`
- `TestMCPInitProjectUsesGlobalRegistry`
- `TestInitCommandLocalModeCreatesProjectLocalStore`
- `TestLegacyLocalStoreIsNotUsedSilently`
- `TestMigrateLocalToGlobalCopiesProjectStore`

### Done When

- `plan-ai init` does not create `<repo>/.plan-ai` by default.
- Project data is stored under `~/.plan-ai/projects/<project-id>/`.
- Project IDs remain stable across repeated runs from the same path.
- Local storage only happens with `--local`.
- Existing local stores produce an explicit migration/local-mode message.

## Phase 2: Install Once Lifecycle

### Objective

Make global installation, update, uninstall, doctor, and setup coherent and idempotent.

### Current Gaps

- `plan-ai install` has both legacy and newer installer paths.
- There is no top-level `plan-ai update` command.
- `SetupMCPConfig()` ignores `OPENCODE_CONFIG_DIR`.
- OpenCode config mutation has multiple authorities.
- Full uninstall does not consistently remove Plan-AI-owned integration artifacts.

### Implement

- Make `Installer.Install()` the only install path.
- Add top-level `plan-ai update`.
- Make install use a staged pipeline: detect, plan, backup, apply, verify.
- Centralize OpenCode config mutation in `internal/opencode`.
- Make installer, setup, bootstrap, sync, and update call the same OpenCode authority.
- Make uninstall remove only Plan-AI-owned artifacts.
- Make doctor verify binary path, MCP tools, OpenCode config, duplicate registrations, and stale config.

### Key Files

- `cmd/plan-ai/setup_commands.go`
- `cmd/plan-ai/installer_commands.go`
- `internal/installer/installer.go`
- `internal/installer/sync.go`
- `internal/opencode/setup.go`
- `internal/config/mcp.go`

### Tests

- `TestInstallDefaultUsesInstallerPath`
- `TestInstallRespectsOpenCodeConfigDir`
- `TestUpdateRefreshesStateAndIntegrations`
- `TestUninstallFullRemovesOpenCodeRegistration`
- `TestDoctorDetectsMissingRegisteredBinary`
- `TestDoctorDetectsDuplicateOpenCodeRegistration`

### Done When

- `plan-ai install` always uses the definitive installer.
- `plan-ai update` repairs and refreshes global installation state.
- `plan-ai uninstall` removes Plan-AI-owned integration artifacts safely.
- `plan-ai doctor` reports real integration health, not only file existence.

## Phase 3: Safe OpenCode Auto-Configuration ADR

### Objective

Update the architectural decision record so install-time OpenCode mutation is allowed when it is safe, explicit, backed up, and reversible.

### Current Gaps

- `docs/adr/0017-opencode-integration.md` says Plan-AI detects OpenCode but never writes to it.
- The product vision requires install/setup to auto-register MCP and configure integrations.

### Implement

- Add a new ADR that supersedes or amends ADR 0017.
- Define ownership markers for generated artifacts.
- Define backup and rollback requirements.
- Define explicit user consent rules for writing real OpenCode config.
- Define how `OPENCODE_CONFIG_DIR` is respected in tests and sandbox flows.

### Key Files

- `docs/adr/0017-opencode-integration.md`
- `docs/adr/0021-install-once-integration.md`
- `docs/opencode-integration.md`
- `docs/opencode-integration-guide.md`

### Tests

- No Go test required for the ADR itself.
- Existing installer/OpenCode tests must encode the ADR rules.

### Done When

- The ADR explains why safe write behavior replaces read-only detection for install/setup.
- Future implementation does not contradict existing architectural docs.

## Phase 4: Conversation Gateway

### Objective

Make conversation the primary Plan-AI UX while preserving commands as automation primitives.

### Current Gaps

- CLI `agent process` exists, but MCP `plan_ai.agent_message` routes to a stub.
- Natural messages are not reliably routed.
- The agent still behaves like a command router more than a product workflow.
- Conversation state is not first-class.

### Implement

- Add `internal/conversation`.
- Route CLI `agent process` and MCP `plan_ai.agent_message` through the same service.
- Add conversation intents for project analysis, product creation, next step, database plan, and impact analysis.
- Persist conversation runs and messages.
- Return clear next actions instead of command instructions where possible.

### Key Files

- `internal/agent/service.go`
- `internal/agent/intent.go`
- `internal/agent/router.go`
- `internal/mcp/handlers.go`
- `internal/mcp/tools.go`
- `cmd/plan-ai/main.go`
- `internal/conversation/*`

### Tests

- `TestConversationCreateSaaSStartsDiscovery`
- `TestConversationAnalyzeProjectReturnsKnownAndMissingContext`
- `TestConversationTellMeWhatIsNextPrioritizesPendingWork`
- `TestConversationCreateDatabasePlanRoutesToDatabasePlanning`
- `TestMCPAgentMessageIsNotStub`

### Done When

- `plan_ai.agent_message` executes real Plan-AI workflow logic.
- The same conversation service is used by CLI and MCP.
- Common user messages produce useful next steps.

## Phase 5: Discovery-First Guardrail

### Objective

Prevent Plan-AI from generating plans before it understands the user's intent well enough.

### Current Gaps

- `discoveryv3` exists, but planning paths do not consistently require approved intent/discovery context.
- `plan-ai plan` can run after vision and approved context without enforcing V3 Product Intent.

### Implement

- Add `PlanningGuard`.
- Use `intentv3.Service.IsApprovedProductIntent()` as a core guard.
- Return next discovery question when planning is requested too early.
- Wire the guard into CLI, MCP, and conversation planning paths.
- Keep any bypass explicit and visible for automation.

### Key Files

- `internal/discoveryv3/service.go`
- `internal/intentv3/service.go`
- `internal/planning/service.go`
- `cmd/plan-ai/plan_commands.go`
- `internal/mcp/handlers.go`

### Tests

- `TestPlanCommandRequiresApprovedIntent`
- `TestPlanningGuardReturnsNextDiscoveryQuestion`
- `TestMCPCreateMasterPlanRespectsPlanningGuard`
- `TestConversationPlanningRequestStartsDiscoveryWhenIntentMissing`

### Done When

- No default planning path creates a plan without enough approved intent/discovery context.
- Missing context produces a useful discovery prompt, not a weak plan or generic error.

## Phase 6: Approved Context Authority

### Objective

Make approved context the authority for what Plan-AI already knows and should not re-ask.

### Current Gaps

- Approved context exists but is split across multiple physical tables.
- Some planning/MCP paths bypass context services and read/write directly.
- Approved facts do not automatically emit memory or continuous planning events.

### Implement

- Add `context.AuthorityService`.
- Add unified `approved_context_items` projection.
- Add FTS over approved context.
- Make planning consume approved facts only.
- Emit memory records and continuous events when facts are approved.
- Support supersession/obsolescence so old approved facts can be replaced safely.

### Key Files

- `internal/context/*`
- `internal/store/ingestion_vision_context_repositories.go`
- `internal/store/store.go`
- `cmd/plan-ai/context_commands.go`
- `cmd/plan-ai/plan_commands.go`

### Tests

- `TestApprovedContextAuthorityDedupesFacts`
- `TestFindApprovedUsesFTS`
- `TestApprovedContextWritesMemoryRecord`
- `TestApprovedContextEmitsContinuousEvent`
- `TestPlanningUsesApprovedFactsOnly`

### Done When

- Approved facts are searchable, deduped, reusable, and authoritative.
- Plan-AI stops asking for already approved facts unless they are ambiguous or superseded.

## Phase 7: Research Reuse

### Objective

Make research reusable by default so important topics are investigated once and reused in future plans.

### Current Gaps

- Research and knowledge are structured, but new research does not first check reusable approved research.
- Research search is weaker than knowledge FTS.
- Approved research is not automatically promoted into reusable knowledge.

### Implement

- Add `research.ReuseService`.
- Add `FindReusableResearch(projectID, topic)`.
- Add FTS-backed research search.
- Exclude draft/unapproved research from planning.
- Promote approved research to knowledge where appropriate.
- Track reuse counts and last reuse timestamps.

### Key Files

- `internal/research/service.go`
- `internal/research/orchestrator.go`
- `internal/store/research_repositories.go`
- `internal/knowledge/service.go`
- `internal/store/knowledge_repositories.go`

### Tests

- `TestFindReusableResearchReturnsApprovedOnly`
- `TestResearchTopicReusesExistingApprovedResearch`
- `TestApprovedResearchPromotesToKnowledge`
- `TestPlanningExcludesDraftResearch`
- `TestResearchReuseIncrementsReuseCount`

### Done When

- New research is created only when no approved reusable research exists or existing research is stale/insufficient.
- Planning uses approved, relevant research rather than every research entry.

## Phase 8: Permanent Memory Recorder

### Objective

Make Plan-AI's memory automatic, durable, searchable, and independent of the current model session.

### Current Gaps

- Memory exists and has FTS, but important project events do not automatically write memory records.
- Memory does not yet carry enough artifact/provenance metadata.
- Question answering is too exact-match oriented.

### Implement

- Add `memory.Recorder`.
- Record memory on approved context, approved research, approved plans, applied proposals, and change events.
- Add topic keys and artifact links.
- Add supersession and scope handling.
- Use FTS-backed memory lookup for answer reuse.

### Key Files

- `internal/memory/memory.go`
- `internal/store/memory_repository.go`
- `internal/store/store.go`

### Tests

- `TestMemoryRecorderRecordsApprovedContext`
- `TestMemoryRecorderRecordsAppliedProposal`
- `TestMemorySearchUsesFTS`
- `TestMemoryFindByTopicKey`
- `TestSupersededMemoryExcludedByDefault`

### Done When

- Plan-AI records important project knowledge without relying on the model or chat session.
- Conversation, planning, and context delivery can retrieve durable memory.

## Phase 9: Continuous Planning Loop

### Objective

Wire the continuous planning pieces into a real end-to-end loop.

### Current Gaps

- Detector, planner, updater, status, context generation, and MCP tools exist.
- The full detect -> analyze -> propose -> approve -> apply loop is not consistently wired.
- Some MCP handlers use direct SQL and bypass service invariants.

### Implement

- Add `continuous.LoopService`.
- Route change detection into continuous events.
- Create proposals from events.
- Split approve/reject/apply behavior into explicit service calls.
- Make `Updater.Apply()` perform real plan update/regeneration work, not just status changes.
- Make context generation use service modules instead of duplicated SQL.

### Key Files

- `internal/continuous/*`
- `internal/change/*`
- `internal/mcp/handlers.go`
- `internal/mcp/tools.go`

### Tests

- `TestContinuousLoopDetectAnalyzeProposeApproveApply`
- `TestApprovePlanUpdateApprovesOnly`
- `TestRejectPlanUpdateRejectsOnly`
- `TestApplyProposalIsIdempotent`
- `TestDetectChangesEmitsContinuousEvent`
- `TestContinuousContextUsesApprovedInputs`

### Done When

- Continuous planning proposes changes instead of silently mutating plans.
- Applying approved proposals is idempotent and updates the relevant planning state.
- CLI and MCP use the same continuous loop service.

## Phase 10: Service-Backed MCP Handlers

### Objective

Reduce MCP handlers from business-logic owners to adapters over service modules.

### Current Gaps

- `internal/mcp/handlers.go` opens stores, runs SQL, creates domain entities, and formats responses directly.
- Some handlers bypass planning, context, research, and continuous services.

### Implement

- Introduce handler dependencies for services instead of raw functions where needed.
- Move business logic into domain services.
- Keep handlers responsible for argument parsing and response shaping only.
- Ensure CLI and MCP paths share service invariants.

### Key Files

- `internal/mcp/handlers.go`
- `internal/mcp/dependencies.go`
- `internal/mcp/tools.go`
- `internal/planning/*`
- `internal/context/*`
- `internal/research/*`
- `internal/continuous/*`

### Tests

- Existing MCP tests should remain green.
- Add contract tests proving CLI and MCP create equivalent artifacts for the same request.

### Done When

- MCP is a transport adapter, not an alternate business logic path.
- Tool behavior matches CLI/conversation behavior.

## Phase 11: Documentation And ADR Alignment

### Objective

Align docs and ADRs with the new product direction.

### Current Gaps

- `docs/adr/0002-storage-layer.md` documents project-local DB as accepted.
- `docs/adr/0017-opencode-integration.md` documents read-only OpenCode detection.
- Existing docs may still describe command-first or local-first behavior.

### Implement

- Add ADR for external project storage.
- Add ADR for safe install-time integration writes.
- Add ADR for conversation gateway.
- Add ADR for approved context authority if unified projection is implemented.
- Update storage, install, OpenCode, MCP, and quickstart docs.

### Key Files

- `docs/adr/0002-storage-layer.md`
- `docs/adr/0017-opencode-integration.md`
- `docs/architecture/storage.md`
- `docs/storage-layer.md`
- `docs/install.md`
- `docs/quickstart.md`
- `docs/opencode-integration.md`
- `docs/mcp-reference.md`

### Done When

- Docs no longer contradict runtime behavior.
- New contributors can implement against the new source-of-truth model without rereading old sessions.

## Phase 12: End-To-End Validation

### Objective

Verify the complete product experience from install to continuous replanning.

### Implement

- Add sandbox tests for install/update/uninstall/doctor.
- Add E2E flow for `plan-ai init` with external storage.
- Add E2E flow for MCP `agent_message` creating product discovery.
- Add E2E flow for approved context -> plan -> change -> proposal -> apply.
- Validate repo cleanliness after default init.

### Tests

- `go test ./...`
- `go vet ./...`
- `go build ./...`
- sandbox install/update/uninstall script
- OpenCode config sandbox test

### Done When

- Default user flow works without writing into the repository.
- Conversation-first flow works through MCP.
- Plans are guarded by discovery and approved context.
- Continuous planning loop can process a change safely.

## Phase 13: Research Intelligence Platform

### Objective

Upgrade Plan-AI research from simple research records into a persistent Research Intelligence Platform that supports reusable knowledge, decision proposals, freshness tracking, implementation-ready context, and continuous updates.

Plan-AI must research before planning when important technical decisions are not yet backed by approved evidence.

### Required Flow

```txt
Need detected
-> Research Job
-> Research Object
-> Knowledge Object
-> Decision Proposal
-> Approved Decision
-> Plan
-> Implementation Context
```

### Explore First

Before implementation, inspect the real codebase and adapt this phase to existing modules.

Review:

- repo structure
- `internal/research`
- `internal/knowledge`
- `internal/planning`
- `internal/store`
- `internal/context`
- `internal/workflows`
- `internal/mcp`
- SQLite migrations
- current entities
- current tests
- research, knowledge, planning, context, workflow, orchestrator, and MCP docs
- related ADRs
- what previous phases already implemented

Do not duplicate modules if the existing seam is correct. Extend existing modules first.

### Current Gaps

- Research exists, but it is not yet the mandatory evidence layer before important plans.
- Research Jobs exist in parts of the schema, but they are not the central entry point for research needs.
- Research results are not yet modeled as rich persistent Research Objects.
- Sources and findings exist, but the platform needs stronger structured classification, linking, and freshness rules.
- Research reuse is not yet enforced before creating new research.
- Approved research is not consistently promoted into reusable Knowledge Objects.
- Decision proposals are not yet first-class outputs of research.
- Planner and Context Engine do not consistently consume only approved research and approved knowledge.

### Implement

- Add or extend Research Jobs as the entry point for research needs.
- Add persistent Research Objects as structured research results.
- Add structured Research Sources.
- Add structured Research Findings.
- Add Research Freshness tracking.
- Add Research Reuse rules.
- Add Knowledge Promotion from approved research.
- Add Decision Proposal generation from research.
- Make research implementation-oriented, not academic.
- Integrate approved Research Objects and Knowledge Objects into Planner.
- Integrate research context into Context Engine.
- Add MCP tools for the research workflow.
- Keep CLI commands minimal and mostly diagnostic.

### Research Jobs

A Research Job represents a need for research.

It can originate from:

- vision
- feature
- specific plan
- pending decision
- detected change
- user request
- impact analysis

It stores:

- topic
- objective
- scope
- research type
- status
- priority
- related project, plan, decision, or change
- created_at
- updated_at

### Research Objects

A Research Object is the persistent result of research.

It stores:

- executive summary
- full research
- sources
- findings
- common mistakes
- best practices
- comparisons
- alternatives
- risks
- final recommendation
- confidence level
- researched_at
- last_checked_at
- freshness_status
- links to Knowledge Objects and decisions

### Research Sources

Each source stores:

- URL or reference
- title
- source type
- publication date if known
- accessed_at
- confidence level
- useful excerpt
- related findings

### Research Findings

Findings must be stored separately, not only inside a large text blob.

Allowed finding types:

- `best_practice`
- `warning`
- `pitfall`
- `compatibility`
- `installation`
- `configuration`
- `performance`
- `security`
- `migration`
- `cost`
- `alternative`
- `recommendation`

### Research Freshness

Add freshness logic:

- `last_checked`
- `last_updated`
- `stale_after_days`
- `freshness_status`
- `reason_for_staleness`

Rules:

- fast-moving technologies expire sooner
- API documentation expires sooner
- stable concepts can stay fresh longer

### Research Reuse

Before creating new research, Plan-AI must search existing approved research.

Rules:

- approved, relevant, fresh research is reused
- approved but stale research is partially updated
- missing research creates a new Research Job
- Plan-AI must not reinvestigate the same topic twice without a reason

### Knowledge Promotion

Approved research can be promoted into reusable Knowledge Objects.

Examples:

- PostgreSQL multi-tenant
- Stripe billing
- Next.js App Router
- MCP architecture
- OpenCode integration

Knowledge Objects can then be reused across future plans.

### Decision Support

Important research can produce Decision Proposals.

Example:

```txt
Research Object: database choice
-> Decision Proposal:
   use PostgreSQL because...
   avoid MariaDB because...
```

Decision proposals are not auto-approved.

Plan-AI proposes. The user approves. Plan-AI stores the approved decision.

### Implementation-Oriented Research

Research must answer implementation questions:

- how to install
- how to configure
- how to deploy
- what version to use
- what errors to avoid
- what libraries to use
- what files to touch
- what tests to run
- how to validate
- how to rollback
- what risks to monitor

### Planning Integration

Planner must consume only approved Research Objects and approved Knowledge Objects.

Specific plans must include:

- research used
- decisions derived
- detected risks
- main sources
- reasons for the chosen solution

Planner must not rely on draft research or loose text.

### Context Engine Integration

Context Engine must serve research at multiple levels:

- short summary
- planning context
- implementation context
- full research context

Context must be derived from Research, Knowledge, Decisions, and Plans without duplicating content.

### MCP And CLI Integration

MCP should support:

- create research job
- search reusable research
- get research context
- list project research
- approve research
- promote research to knowledge
- check freshness
- request research update

CLI should stay minimal and mostly diagnostic.

### Key Files

- `internal/research/*`
- `internal/knowledge/*`
- `internal/planning/*`
- `internal/context/*`
- `internal/store/*research*`
- `internal/store/knowledge_repositories.go`
- `internal/store/store.go`
- `internal/mcp/tools.go`
- `internal/mcp/handlers.go`
- `cmd/plan-ai/data_commands.go`
- `docs/research-engine.md`
- `docs/architecture/research-engine.md`
- `docs/adr/0009-research-knowledge.md`

### Tests

- no duplicate research is created when approved fresh research exists
- stale research is marked correctly
- research finding is stored correctly
- research source links to Research Object
- approved research promotes to Knowledge Object
- planner only uses approved research
- context engine generates research context without duplication
- decision proposal derives from Research Object
- MCP can create and retrieve research jobs
- MCP can check freshness and request update

### Documentation

Update docs to explain:

- what the Research Intelligence Platform is
- Research -> Knowledge -> Decision -> Plan flow
- reuse rules
- freshness rules
- relationship with continuous planning
- how Plan-AI avoids reinvestigating the same topic

### Done When

- Research Jobs model research needs.
- Research Objects persist structured results.
- Sources and findings are first-class records.
- Freshness determines reuse, update, or new research behavior.
- Approved research can become reusable knowledge.
- Research can produce decision proposals.
- Planner uses approved research and approved knowledge only.
- Context Engine serves research context at multiple levels.
- MCP exposes the research workflow.
- Existing tests still pass.

## Phase 14: Universal Level Model And Change Lifecycle

### Objective

Close the gaps between the actual Plan-AI repository and the Universal Level Model defined in `docs/architecture/Plan-AI_Niveles_y_Gestion_de_Cambios.docx`.

Make Outcome a first-class entity, make Phases reachable through a real repository, make Decisions supersedable, make Snapshots rollback-capable, and unify the parallel v1/v2 implementations of Visions, Plans, Changes, Impacts, and Snapshots.

### Source Authority

`docs/architecture/Plan-AI_Niveles_y_Gestion_de_Cambios.docx` defines:

- 11 levels: Vision (0) → Outcome (1) → Requirements (2) → Research (3) → Knowledge (4) → Decisions (5) → Master Plan (6) → Specific Plans (7) → Implementation Documents (8) → Phases (9) → Tasks (10)
- 13 universal fields applied to every level
- 7 universal questions every level must answer
- 10-state Change Request lifecycle: `draft, submitted, analyzing, research_required, proposal_ready, waiting_approval, approved, applied, validated, rejected, cancelled`
- Impact Graph mapping `Decision → Research → Knowledge → Plan → Doc → Phase → Task → Files`
- Snapshots with version history and rollback
- 9 universal entity states: `draft, in_review, approved, rejected, blocked, implemented, validated, archived, superseded`
- 15 final system rules

### Current Gaps

- `phases` table exists (`internal/store/migrations.go:194`) but `store/repositories/phase_repository.go` is missing. `Repositories` struct at `internal/store/repositories.go:9-22` has no `Phase` field.
- `decisions` table has no `supersedes_id` or `superseded_by_id` column. `domain/decision.go:1-49` has no supersession state.
- `domain/change.go:9-14` declares only 5 ChangeRequest states (`draft, submitted, approved, rejected, applied`); DOCX requires 10.
- `change_events.status` is a free string at `internal/store/phase18_20_repositories.go:107-110` with no enum or transitions.
- `HandleRollbackSnapshot` is registered in `internal/mcp/dependencies.go:28` but the function definition does not exist.
- Three parallel change systems run: `change_requests` (`store/repositories/change_repository.go`), `change_events` (`store/phase18_20_repositories.go:62-135`), and `change_impact_reports_v2` (`store/v2_stage_d_repositories.go:10-58`).
- Six separate `approved_*` tables exist (`store.go:658-663`) instead of one unified `approved_context` table.
- No `outcomes` table; Outcome is a free-text field on `visions.expected_outcome`.
- No `subplans` or `parent_specific_plan_id` column to support subplan hierarchy.
- No `entity_links` table for cross-entity relationships (closest: `knowledge_relations`, which only links knowledge-to-knowledge).
- No universal history/audit table; only `decision_history` records per-entity history.
- `implementation_documents.context/delivery.go:347-375` (`buildImplementationContext`) emits only Constraints and Decisions; missing files, validations, known_risks, knowledge.
- Universal 13-field structure is partially implemented: `contexto`, `cambios`, `trazabilidad` fields are missing or unpersisted.
- 7 universal questions have no schema-level enforcement.

### Implement

- Add migration `0043_universal_structure` with:
  - `outcomes` table: `(id, project_id, statement, success_criteria TEXT[], approved, created_at, updated_at)`
  - ALTER `decisions` ADD COLUMN `supersedes_id TEXT NOT NULL DEFAULT ''`, `superseded_by_id TEXT NOT NULL DEFAULT ''`
  - ALTER `phases` ADD COLUMN `parent_phase_id TEXT NOT NULL DEFAULT ''`, `context TEXT NOT NULL DEFAULT ''`, `expected_outcome TEXT NOT NULL DEFAULT ''`, `risks TEXT NOT NULL DEFAULT '[]'`
  - ALTER `tasks` ADD COLUMN `dependencies TEXT NOT NULL DEFAULT '[]'`, `impact TEXT NOT NULL DEFAULT ''`, `expected_outcome TEXT NOT NULL DEFAULT ''`
  - ALTER `master_plans` ADD COLUMN `outcomes TEXT NOT NULL DEFAULT '[]'`, `validations TEXT NOT NULL DEFAULT '[]'`, `context TEXT NOT NULL DEFAULT ''`
  - ALTER `specific_plans` ADD COLUMN `parent_specific_plan_id TEXT NOT NULL DEFAULT ''`, `context TEXT NOT NULL DEFAULT ''`
  - ALTER `change_requests` ADD COLUMN `change_classification TEXT NOT NULL DEFAULT ''`, `analyzed_at`, `proposal_id`, `cancelled_at`, `validated_at`
  - CREATE `change_proposals` table: `(id, change_request_id, project_id, summary, plan_changes JSON, decision_changes JSON, status, created_at)`
  - CREATE `change_audit` table: `(id, change_request_id, project_id, actor, action, from_state, to_state, note, created_at)`
- Extend `domain.ChangeRequestStatus` to the full 10-state enum and add `ValidChangeRequestTransitions(from, to) bool`.
- Add `domain.Decision` fields `SupersedesID` and `SupersededByID`. Add `DecisionSuperseded` status. Extend `ValidDecisionTransitions` with `approved → superseded` and `superseded → (terminal)`.
- Create `internal/store/repositories/phase_repository.go` with `Save`, `GetByID`, `ListByPlan`, `ListByProject`, `UpdateStatus`, `Delete`. Wire into `Repositories` struct.
- Create `internal/store/repositories/outcome_repository.go`.
- Consolidate the three parallel change systems: make `change_requests` the single source of truth. Migrate `change_events` and `change_impact_reports_v2` writes behind `change_requests` with internal mirroring.
- Unify the 6 `approved_*` tables into one `approved_context` table. Add migration `0044_approved_context_unified` that copies data then drops the 6 legacy tables.
- Implement `HandleRollbackSnapshot` in `internal/mcp/handlers.go`: read `snapshots_v2.entity_snapshot`, deserialize stored decision/plan/task states, write back through the relevant repositories inside a single transaction.
- Extend `context.DeliveryEngine.buildImplementationContext` (`internal/context/delivery.go:347-375`) to include `expected_files`, `validations`, `known_risks`, `testing_strategy`, and `knowledge_objects.summary` linked to the implementation document.
- Add `entity_links` table for cross-entity traceability: `(id, project_id, source_type, source_id, target_type, target_id, link_type, created_at, UNIQUE(source_type,source_id,target_type,target_id,link_type))`.

### Key Files

- `internal/store/migrations.go`
- `internal/store/repositories.go`
- `internal/store/repositories/phase_repository.go` (new)
- `internal/store/repositories/outcome_repository.go` (new)
- `internal/store/repositories/change_repository.go`
- `internal/store/repositories/decision_repository.go`
- `internal/store/ingestion_vision_context_repositories.go`
- `internal/domain/change.go`
- `internal/domain/decision.go`
- `internal/domain/planning.go`
- `internal/change/service.go`
- `internal/change/impact.go`
- `internal/change/snapshots.go`
- `internal/context/delivery.go`
- `internal/mcp/handlers.go`
- `internal/mcp/dependencies.go`

### Tests

- `TestPhaseRepository_CRUD`: Save, GetByID, ListByPlan, UpdateStatus (using `ValidPhaseTransitions`), Delete
- `TestPhaseSubplan_Hierarchy`: create parent phase, create child with `parent_phase_id=parent.id`, query child by parent
- `TestOutcome_ApprovedBoolean`: create Outcome, approve, assert Approved=true; cannot unapprove
- `TestChangeRequestLifecycle_TenStates`: walk `draft→submitted→analyzing→research_required→proposal_ready→waiting_approval→approved→applied→validated`; reject illegal transitions (e.g., `draft→approved`)
- `TestDecisionSupersession`: create Decision A (approved), create Decision B with `supersedes=A`, assert `A.status == superseded` and `B.status == approved`
- `TestChangeAudit_RecordsTransitions`: walk a CR through 5 states, assert 5 `change_audit` rows with correct `from_state`/`to_state`
- `TestHandleRollbackSnapshot_RestoresState`: create decision→plan→task chain, take snapshot, change all three, rollback, assert state matches snapshot
- `TestApprovedContext_UnifiedTable`: Save `requirement` type and `preference` type, `ListApproved(projectID)` returns both from the unified table
- `TestL2ImplementationContext_IncludesFilesAndRisks`: create `implementation_documents` with `expected_files=["cmd/main.go"]`, `validations=["go test"]`, `known_risks=["db migration"]`; deliver L2 context; assert output contains all three
- `TestEntityLinks_TraceCrossEntity`: create Decision→Knowledge→Plan→Doc→Phase→Task links, query by source, assert all 5 targets returned
- `TestParallelChangeSystems_WriteToSameRow`: write a change via legacy `ChangeRepository` and via `change.Service.RegisterChange`, assert both surfaces see it

### Documentation

- Update `docs/architecture/domain-model.md` to reflect Outcome, Phases repo, unified approved context, decision supersession, and 10-state ChangeRequest lifecycle.
- Add ADR for Universal Level Model schema extensions.
- Supersede parts of `docs/architecture/storage.md` that describe parallel change systems.

### Done When

- `internal/store/repositories/` contains 15 files (add `phase_repository.go`, `outcome_repository.go`).
- `internal/domain/planning.go` defines `Outcome` struct.
- `domain.ChangeRequestStatus` has 10 states with valid transition map.
- `grep "HandleRollbackSnapshot" internal/mcp/` returns exactly 1 definition.
- `internal/store/store.go` has a single `approved_context` table; the 6 legacy `approved_*` tables are dropped.
- `internal/context/delivery.go` `buildImplementationContext` writes at least 5 distinct sections.
- `entity_links` table exists and is populated on save operations.
- All existing tests pass; new tests green.

## Phase 15: Architectural Authority Consolidation

### Objective

Eliminate the architectural bifurcation in Plan-AI where two parallel type systems exist (canonical services in `internal/{planning,vision,research,knowledge,change,requirements}/` and store-direct repositories in `internal/store/repositories/`), and where the MCP server (`internal/mcp/handlers.go`, 1642 LOC) bypasses all canonical services by writing directly to store.

Make the 6 fundamental principles of the v4 architectural spec enforced end-to-end through MCP, CLI, and conversation paths.

### Source Authority

`docs/architecture/PlanAI_Especificacion_Arquitectonica_Oficial_v4.docx` defines 21 layers and 6 fundamental principles:

1. Nada aprobado se vuelve a preguntar.
2. Nada investigado se vuelve a investigar.
3. Nada decidido se vuelve a decidir.
4. Nada planificado se vuelve a planificar.
5. Solo se actualiza lo afectado por un cambio aprobado.
6. El contexto siempre se deriva de entidades aprobadas.

The 21 layers are: Vision, Outcome, Requirements, Research Intelligence Platform, Knowledge Engine, Decision Engine, Master Plan, Specific Plans, Implementation Documents, Phase Engine, Task Engine, Change Engine, Impact Graph, Context Engine, Workflow Engine, Orchestrator, Model Strategy, Capability Registry, Store Layer, MCP Layer, OpenCode Integration.

### Current Gaps

- `internal/mcp/handlers.go` (1642 LOC) writes to `repos.Plan`, `repos.Research`, `repos.Knowledge`, `repos.Vision`, `repos.Decision`, `repos.Change`, `repos.Requirement` directly, bypassing the canonical services.
- Two parallel type systems exist:
  - Canonical: `planning.MasterPlan` (`internal/planning/master_plan_v2.go`), `research.ResearchEntry` (`internal/research/types.go`), `vision.Draft` (`internal/vision/types.go`)
  - Store-direct: `domain.MasterPlan`, `domain.Research`, `domain.Vision` in `internal/domain/`
- `internal/mcp/handlers.go:1377-1500` instantiates `intentv3` directly, bypassing `planning.Service.CreateMasterPlan` and its `ApprovedContext` guard at `internal/planning/service.go:25-27`.
- `internal/intentv3/` exists solely as a wrapper to facilitate the MCP bypass.
- `internal/planning/master_plan.go:1` and `internal/planning/specific_plan.go:1` are 1-line stubs (`package planning`).
- The 6 v4 principles are all violated in runtime:
  - **Principle 1**: `mcp/handlers.go:193,235,276,438,498,1085,1137,1170,1182,1194,1206,1219,1266,1270,1290,1300,1323` writes to store without consulting `context.Registry.IsApproved`.
  - **Principle 2**: MCP saves `domain.Research` directly without deduplicating by `topic_key`.
  - **Principle 3**: MCP writes `domain.Decision` directly without enforcement of "ya decidido".
  - **Principle 4**: MCP uses `intentv3` instead of `planning.Service`, bypassing the approved-context guard.
  - **Principle 5**: `internal/change/analyzer.go:38-42` returns literal `"Template analysis - entities will be resolved against actual data"`.
  - **Principle 6**: MCP reads store directly instead of `context.DeliveryEngine`.

### Decision Required

Choose the source-of-truth type system:

- **Option A (recommended)**: Make canonical types (`planning.MasterPlan`, `research.ResearchEntry`, etc.) the source of truth. Store repositories implement the canonical interfaces. Store-direct types (`domain.MasterPlan`, etc.) are deprecated.
- **Option B**: Enrich store-direct types with the canonical fields and make them the source of truth. Canonical packages become thin adapters.

This phase assumes **Option A** is approved. If Option B is preferred, swap the consolidation target but keep the rest of the phase intact.

### Implement

- Make `internal/store/repositories/*_repository.go` implement the canonical interfaces (`planning.Repository`, `research.Repository`, `knowledge.Repository`, `vision.Repository`, `decision.Repository`, `change.Repository`, `requirement.Repository`).
- Rewrite `internal/mcp/handlers.go:137-1500` so that every write operation calls the canonical service instead of the repository directly. This reactivates `planning.Service:25-27` approved-context guard automatically.
- Consolidate `internal/planning/master_plan_v2.go`, `internal/planning/specific_plan_v2.go`, `internal/planning/evolution_v3.go` into a single canonical module. Delete the 1-line stubs `master_plan.go` and `specific_plan.go`.
- Either delete `internal/intentv3/` or downgrade it to a thin adapter that delegates to canonical services.
- Add middleware in `internal/mcp/sdk_server.go` that wraps every tool call: before write, call `context.Registry.IsApproved(...)` to enforce Principle 1.
- Add middleware that deduplicates research by `topic_key` before write, enforcing Principle 2.
- Add middleware that checks `decisions` table for existing approved decisions with the same key before allowing a new write, enforcing Principle 3.
- Add middleware that requires `planning.Service.IsContextApproved` for any plan creation, enforcing Principle 4.
- Replace `internal/change/analyzer.go:38-42` placeholder text with a real implementation that walks the `entity_links` table (added in Phase 14), enforcing Principle 5.
- Replace direct store reads in `mcp/handlers.go` with `context.DeliveryEngine` calls, enforcing Principle 6.
- Add CLI/MCP equivalence tests: for every tool, prove that the CLI command and the MCP tool produce the same persisted state.

### Key Files

- `internal/mcp/handlers.go` (rewrite from 1642 LOC to ~600 LOC)
- `internal/mcp/sdk_server.go` (add middleware)
- `internal/mcp/dependencies.go`
- `internal/mcp/tools.go`
- `internal/store/repositories/*.go` (implement canonical interfaces)
- `internal/planning/service.go`
- `internal/planning/master_plan_v2.go` (consolidate)
- `internal/planning/specific_plan_v2.go` (consolidate)
- `internal/planning/evolution_v3.go` (consolidate)
- `internal/planning/master_plan.go` (delete stub)
- `internal/planning/specific_plan.go` (delete stub)
- `internal/research/service.go`
- `internal/knowledge/service.go`
- `internal/change/analyzer.go` (replace placeholder)
- `internal/context/registry.go` (consumed by middleware)
- `internal/intentv3/` (delete or downgrade to adapter)
- `cmd/plan-ai/` (verify CLI uses canonical services, not direct store writes)

### Tests

- `TestMCPNoLongerBypassesServices`: assert that for every `mcp/handlers.go` write, the call goes through a canonical service. `grep "repos\.\(Plan\|Research\|Knowledge\|Vision\|Decision\|Change\|Requirement\)\." internal/mcp/handlers.go` returns 0.
- `TestPrinciple1_ApprovedContextNotReasked`: MCP write to `master_plans` is rejected if approved context for the same `vision_reference` is missing.
- `TestPrinciple2_ResearchDeduplication`: saving Research with same `topic_key` returns existing record.
- `TestPrinciple3_DecisionNotRedecided`: saving Decision with same `key` as an approved decision returns the approved one.
- `TestPrinciple4_PlanningRequiresApprovedContext`: MCP `create_master_plan` without approved context returns 400 with discovery prompt.
- `TestPrinciple5_AnalyzerReturnsRealEntities`: `change.Service.AnalyzeChange(changeID)` returns `ImpactAnalysis.AffectedEntities` with concrete IDs (not `"Template analysis"`).
- `TestPrinciple6_ContextDeliveredFromApprovedEntities`: `mcp/handlers.go` read paths use `context.DeliveryEngine`, not direct store reads.
- `TestCLIEqualsMCP`: for 10 representative operations, CLI and MCP produce the same persisted state.
- `TestIntentV3_Deprecated`: `internal/intentv3/service.go` is either deleted or its public methods are thin pass-throughs to canonical services.
- `TestPlanningStubs_Deleted`: `internal/planning/master_plan.go` and `specific_plan.go` are no longer stubs.

### Documentation

- Add ADR: "Single source of truth for entities: canonical services own writes, stores implement interfaces."
- Add ADR: "MCP as transport adapter, not alternate business logic path."
- Add ADR: "Six v4 principles as enforced invariants."
- Update `docs/architecture/overview.md` to document the consolidated architecture.
- Update `docs/architecture/storage.md` to remove references to store-direct types as a separate API.
- Update `docs/mcp-server.md` and `docs/mcp-reference.md` to document the new middleware behavior.
- Update `docs/cli-reference.md` to clarify CLI/MCP equivalence.

### Done When

- `grep "repos\.\(Plan\|Research\|Knowledge\|Vision\|Decision\|Change\|Requirement\)\." internal/mcp/handlers.go` returns 0.
- `internal/intentv3/` is deleted or marked as deprecated adapter.
- `internal/planning/master_plan.go` and `specific_plan.go` are no longer 1-line stubs.
- The 6 v4 principles each have at least one passing test that validates them end-to-end via MCP.
- CLI and MCP produce equivalent persisted state for 10 representative operations.
- All existing tests pass; new tests green.

## Phase 16: ImpactGraph, Workflows, And Capabilities

### Objective

Implement the architectural layers that v4 requires but the repo only has as placeholders or in-memory stubs:

- **Impact Graph**: real graph persistence and traversal (currently `grep -rn "ImpactGraph\|impact_graph"` = 0 matches).
- **Workflow Engine**: real step execution (currently `workflows/registry.go:44-56` marks completed without executing `Steps`).
- **Capability Registry**: SQL-backed persistence (currently in-memory `sync.RWMutex` + map at `internal/capabilities/registry.go`).
- **Orchestrator**: become the actual MCP/CLI entry point (currently `orchestrator/orchestrator.go` is never invoked).
- **OpenCode authority**: consolidate the split between `internal/installer/sync.go:27` and `internal/opencode/setup.go`.

### Source Authority

`docs/architecture/PlanAI_Especificacion_Arquitectonica_Oficial_v4.docx` layers 14 (Impact Graph), 16 (Workflow Engine), 17 (Orchestrator), 19 (Capability Registry), 21 (OpenCode Integration).

### Current Gaps

- `internal/change/impact.go` uses static `DefaultInvalidationRules` at `change/types.go:81-93`; not graph-derived.
- `internal/workflows/registry.go:44-56` `ExecuteWorkflow` creates the run and marks it completed without executing `Steps`.
- `internal/capabilities/registry.go` is in-memory; `capabilities.NewDefaultRegistry()` is hardcoded at `types.go:9-20`.
- `internal/orchestrator/orchestrator.go` has `CreateJob`, `SelectCapability`, `LoadContext` but MCP and CLI do not call it.
- `internal/installer/sync.go:24-27` delegates to `opencode.SetupMCPConfig`; `internal/installer/installer.go:539` uses `opencode.SetupService`; `cmd/plan-ai/setup_commands.go:93,103,182,351,413,433` calls both. Authority is divided.
- `internal/skills/doc.go` declares a decision to not implement; this blocks future skills work.

### Dependencies

- Requires Phase 15 (Architectural Authority Consolidation) to be complete so the Orchestrator can be invoked through canonical services.
- Requires Phase 14 (Universal Level Model) so `entity_links` exists for the Impact Graph to traverse.

### Implement

- Create `internal/impact/` package:
  - `graph.go`: graph data structure (nodes = entities, edges = relationships).
  - `traversal.go`: BFS/DFS from a `ChangeRequest` root.
  - `store.go`: persists edges in `impact_edges` table.
  - Migration `0045_impact_edges`: `(id, project_id, source_type, source_id, target_type, target_id, edge_type, weight, created_at, UNIQUE(...))`.
  - `builder.go`: scans `entity_links` (added in Phase 14) to populate the graph on demand.
- Replace `internal/change/analyzer.go:38-42` with a real implementation that calls `impact.Graph.Traverse(changeID)` and returns concrete `AffectedEntities` with IDs and reasons.
- Refactor `internal/workflows/execution.go` `ExecuteWorkflow` to iterate over `Steps[]`, execute each step (sequentially or in parallel as declared), persist state after each step, and emit events. The four `*_workflow.go` files (`vision_workflow`, `research_workflow`, `planning_workflow`, `approval_workflow`) become real step definitions.
- Make `internal/orchestrator/orchestrator.go` the MCP/CLI entry point. Add MCP tool `plan_ai.run_workflow` that calls `Orchestrator.CreateJob` with the selected capability and workflow.
- Create `internal/capabilities/store.go` and `internal/capabilities/repository.go`. Migration `0046_capabilities`: `(id, name, description, schema, version, enabled, created_at)`. Replace `NewDefaultRegistry()` with a loader that reads from DB and seeds on first run.
- Consolidate OpenCode authority into `internal/opencode/`. `internal/installer/sync.go` becomes a thin wrapper. `internal/installer/installer.go:524-560` (duplicate OpenCode logic) is removed.
- Update `cmd/plan-ai/setup_commands.go` to call only the consolidated `opencode` package.
- Decide the scope of `internal/skills/`. Either remove the placeholder `doc.go` or implement the minimal skills registry. Document the decision in an ADR.

### Key Files

- `internal/impact/{graph,traversal,store,builder,migration}.go` (new, ~400 LOC)
- `internal/change/analyzer.go` (replace placeholder)
- `internal/change/impact.go` (use graph)
- `internal/workflows/execution.go` (rewrite)
- `internal/workflows/vision_workflow.go` (real steps)
- `internal/workflows/research_workflow.go` (real steps)
- `internal/workflows/planning_workflow.go` (real steps)
- `internal/workflows/approval_workflow.go` (real steps)
- `internal/orchestrator/orchestrator.go` (wiring)
- `internal/mcp/server.go` (use Orchestrator as entry point)
- `internal/capabilities/{registry,store,repository,types}.go`
- `internal/opencode/setup.go` (consolidate)
- `internal/installer/sync.go` (simplify to wrapper)
- `internal/installer/installer.go:524-560` (remove)
- `cmd/plan-ai/setup_commands.go`
- `internal/skills/doc.go` (decide and document)

### Tests

- `TestImpactGraph_DecisionAffectsTransitivePlans`: create Decision→Knowledge→Plan→Doc→Phase→Task, mutate the decision, assert all 5 entity types appear in `AffectedEntities`.
- `TestImpactGraph_TraversalIsolation`: two unrelated change requests do not share affected entities.
- `TestImpactGraph_BuildFromEntityLinks`: after Phase 14's `entity_links` are populated, `impact.Graph.Build()` returns the correct graph.
- `TestWorkflowEngine_ExecutesSteps`: a workflow with 3 steps, after `ExecuteWorkflow`, all 3 step results are persisted.
- `TestWorkflowEngine_FailsWithRollback`: if step 2 fails, step 1 is rolled back and the workflow is marked failed.
- `TestOrchestrator_IsMCPEntryPoint`: MCP tool `plan_ai.run_workflow` calls `Orchestrator.CreateJob` and returns the run ID.
- `TestCapabilities_LoadFromDB`: after migration, `Registry.List()` returns the seeded capabilities, not the hardcoded ones.
- `TestCapabilities_DisabledNotListed`: a disabled capability is not returned.
- `TestOpenCode_AuthorityConsolidated`: only `internal/opencode/` writes OpenCode config; `internal/installer/` is a wrapper.
- `TestCLIEqualsMCP_WorkflowExecution`: CLI `plan-ai workflow run` and MCP `plan_ai.run_workflow` produce the same execution state.

### Documentation

- Add ADR: "Impact Graph as the authority for change propagation."
- Add ADR: "Workflow Engine executes steps with persistence and rollback."
- Add ADR: "Capability Registry backed by SQL, not in-memory."
- Add ADR: "OpenCode integration owned by `internal/opencode/` package."
- Update `docs/architecture/overview.md` with the consolidated architecture.
- Update `docs/opencode-integration.md` to reflect the new authority model.
- Document the skills decision in `docs/adr/`.

### Done When

- `grep -rn "ImpactGraph\|impact_graph" /root/plan-ai` returns real references in `internal/impact/` and tests.
- `analyzer.go:38-42` no longer returns `"Template analysis"`.
- `workflows/registry.go:44-56` executes real steps and persists state.
- `capabilities` package has SQL persistence and seeded data.
- `Orchestrator.CreateJob` is the only MCP/CLI entry point for workflow execution.
- `internal/opencode/` is the only writer of OpenCode config.
- All existing tests pass; new tests green.

## Implementation Dependency Graph

```txt
Phase 1 -> Phase 2 -> Phase 4 -> Phase 5 -> Phase 6 -> Phase 7 -> Phase 9 -> Phase 12
              \          \          \          \          \          \
               -> Phase 3 -> Phase 4 -> Phase 10 -> Phase 12

Phase 5 -> Phase 6 -> Phase 13 -> Phase 9 -> Phase 12

Phase 14 -> Phase 15 -> Phase 16

Phase 14 should run after Phase 6 (Approved Context Authority) and Phase 9 (Continuous Planning Loop) are in place, because it consolidates their schemas.

Phase 15 (Architectural Authority Consolidation) must run before Phase 16. The MCP consolidation in Phase 15 is the prerequisite for the Orchestrator to be the real entry point in Phase 16.

Phase 11 should run alongside each phase as docs/ADRs become necessary.
```

## Highest Priority Start

Start with Phase 1.

Do not start with the conversation gateway before external storage exists. Conversation depends on stable project identity and a single source of truth.

Do not start Phase 15 before Phase 14 is complete. Phase 15's MCP consolidation exposes the schema gaps that Phase 14 closes (Outcome, Phases, unified approved context, decision supersession).

Do not start Phase 16 before Phase 15 is complete. The Orchestrator in Phase 16 needs the canonical services consolidated in Phase 15.

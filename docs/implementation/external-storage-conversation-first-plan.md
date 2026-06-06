# Plan-AI Implementation Study: External Storage And Conversation-First UX

**Status:** Approved for implementation planning
**Source:** `docs/architecture/plan-ai-product-vision.md`
**Primary decision:** External project storage by default; project-local storage only by explicit opt-in.

## Approved Product Decision

Plan-AI will use external project storage by default.

Default behavior:

```sh
plan-ai init
```

Stores project data in:

```txt
~/.plan-ai/projects/<project-id>/project.db
```

It does not write inside the repository by default.

Explicit local mode:

```sh
plan-ai init --local
```

Stores project data in:

```txt
<repo>/.plan-ai/project.db
```

Legacy local stores are not used silently.

If Plan-AI detects an existing legacy local store:

```txt
<repo>/.plan-ai/project.db
```

It should guide the user toward one of these explicit actions:

```sh
plan-ai migrate local-to-global
plan-ai init --local
```

## Implementation Order

1. External Project Store + Global Registry
2. Install Once + Update/Doctor/Uninstall
3. Conversation Gateway
4. Discovery-First Guardrail
5. Approved Context Authority
6. Research Reuse
7. Permanent Memory Recorder
8. Continuous Planning Loop
9. ADR/docs updates
10. Tests

## Phase 1: External Project Store + Global Registry

### Goal

Move Plan-AI's default project persistence out of the repository and into the global Plan-AI home.

### Key Files

- `internal/config/config.go`
- `internal/store/store.go`
- `internal/store/project_store.go`
- `internal/store/global_store.go`
- `cmd/plan-ai/main.go`
- `internal/mcp/handlers.go`

### Current Behavior

- `config.ProjectDir(projectDir)` resolves to `<project>/.plan-ai`.
- `store.EnsureProjectLayout(projectRoot)` creates project-local directories.
- `openInitializedProjectStore()` opens `<project>/.plan-ai/project.db`.
- `store.ProjectID(rootPath)` returns `project:<path>`, so identity changes if the repository moves.
- MCP handlers call `store.OpenProjectStore(projectRoot)`, which also uses project-local storage.

### Implementation Path

Add a registry-backed project resolution seam in `internal/store`.

Suggested modules:

- `ProjectResolver`
- `ProjectRegistryRepository`
- `ExternalProjectLayout`

Suggested behavior:

- Resolve current project by path or existing registry entry.
- Register unknown projects in global storage.
- Generate stable project IDs independent of filesystem path.
- Open project stores under `~/.plan-ai/projects/<project-id>/`.
- Keep `<repo>/.plan-ai` as explicit `--local` compatibility mode only.

### Proposed Global Tables

```sql
CREATE TABLE IF NOT EXISTS projects_registry (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  canonical_path TEXT NOT NULL,
  store_path TEXT NOT NULL,
  storage_mode TEXT NOT NULL DEFAULT 'external',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_seen_at TEXT
);
```

```sql
CREATE TABLE IF NOT EXISTS project_paths (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  path_type TEXT NOT NULL DEFAULT 'canonical',
  first_seen_at TEXT NOT NULL,
  last_seen_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects_registry(id)
);
```

### Tests

- `TestInitCommandDoesNotWriteProjectLocalByDefault`
- `TestEnsureExternalProjectLayoutCreatesExpectedPaths`
- `TestRegisterProjectCreatesStableID`
- `TestMCPInitProjectUsesGlobalRegistry`
- `TestInitCommandLocalModeCreatesProjectLocalStore`

## Phase 2: Install Once

### Goal

Make `plan-ai install` the single global setup path for machine-level installation and integrations.

### Key Files

- `cmd/plan-ai/setup_commands.go`
- `internal/installer/installer.go`
- `internal/installer/sync.go`
- `internal/opencode/setup.go`
- `internal/config/mcp.go`

### Current Behavior

- `plan-ai install` has both legacy and newer installer paths.
- There is no top-level `plan-ai update` command.
- `SetupMCPConfig()` ignores `OPENCODE_CONFIG_DIR`.
- OpenCode setup authority is split across installer and opencode packages.
- Full uninstall does not always remove owned OpenCode integration artifacts.

### Implementation Path

- Make `Installer.Install()` the only install path.
- Add `plan-ai update` as the idempotent repair/sync path.
- Centralize OpenCode config mutation in `internal/opencode`.
- Make install follow a staged pipeline: detect, plan, backup, apply, verify.
- Make uninstall remove Plan-AI-owned artifacts only.
- Make `doctor` verify binary path, MCP tool validity, OpenCode registration, duplicate registrations, and stale config.

### ADR Impact

`docs/adr/0017-opencode-integration.md` currently says Plan-AI detects OpenCode but never writes to it. This product vision supersedes that constraint for explicit install/setup flows.

Add a new ADR that defines safe auto-configuration with backups, ownership markers, and uninstall cleanup.

## Phase 3: Conversation Gateway

### Goal

Make conversation the primary UX while preserving commands for automation and power users.

### Key Files

- `internal/agent/service.go`
- `internal/agent/intent.go`
- `internal/agent/router.go`
- `internal/discoveryv3/service.go`
- `internal/intentv3/service.go`
- `internal/mcp/handlers.go`
- `internal/mcp/tools.go`

### Current Behavior

- CLI `agent process` exists.
- MCP `plan_ai.agent_message` exists but currently routes to a stub handler.
- Natural requests such as `create a SaaS`, `analyze this project`, and `create database plan` are not reliably routed.
- The agent records jobs but does not execute the discovery-first product flow.

### Implementation Path

Create `internal/conversation` as the product-level gateway.

Suggested interface:

```go
type Service interface {
    Process(projectRoot string, message string) (Response, error)
}
```

Responsibilities:

- Resolve project identity through the global registry.
- Load approved context, product intent, discovery state, plans, tasks, and research summaries.
- Classify natural-language requests.
- Route to discovery, project analysis, next step, impact analysis, or planning.
- Persist conversation runs/messages.
- Return clear next actions.

Supported conversational examples:

- `Plan-AI, analyze this project.`
- `Plan-AI, I want to create a SaaS.`
- `Plan-AI, tell me what is next.`
- `Plan-AI, create the database plan.`
- `Plan-AI, analyze the impact of this change.`

### MCP Changes

- Replace the MCP `HandleAgentProcess()` stub.
- Route `plan_ai.agent_message` through `internal/conversation.Service`.
- Add or expose discovery tools for next question, answer question, and discovery status.

## Phase 4: Discovery-First Guardrail

### Goal

Prevent weak plans by requiring enough approved intent/discovery context before planning.

### Rule

No plan should be created until Plan-AI has enough approved intent/discovery context.

If context is missing, Plan-AI should ask the next discovery question instead of generating a plan.

### Key Files

- `internal/discoveryv3/service.go`
- `internal/intentv3/service.go`
- `cmd/plan-ai/plan_commands.go`
- `internal/planning/service.go`
- `internal/mcp/handlers.go`

### Implementation Path

- Use `intentv3.Service.IsApprovedProductIntent()` as the core guard.
- Add a `PlanningGuard` seam used by CLI, MCP, and conversation paths.
- Return discovery questions when the guard fails.
- Keep any explicit bypass separate and visible if needed for automation.

## Phase 5: Approved Context Authority

### Goal

Make approved context the primary source of truth for planning and conversation.

### Key Files

- `internal/context/*`
- `internal/store/ingestion_vision_context_repositories.go`
- `cmd/plan-ai/context_commands.go`
- `cmd/plan-ai/plan_commands.go`

### Current Behavior

- Approved context exists but is split across multiple tables.
- Some paths still read raw research, knowledge, or direct SQL.
- Planning does not consistently enforce approved-only inputs.

### Implementation Path

- Add `context.AuthorityService`.
- Add a unified projection table for approved context.
- Add FTS search over approved context.
- Make planning consume approved facts only.
- Emit continuous events and memory records when facts are approved.

### Proposed Table

```sql
CREATE TABLE IF NOT EXISTS approved_context_items (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  item_type TEXT NOT NULL,
  source_id TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  normalized_content TEXT NOT NULL,
  state TEXT NOT NULL DEFAULT 'approved',
  confidence REAL NOT NULL DEFAULT 1.0,
  supersedes_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, item_type, normalized_content)
);
```

## Phase 6: Research Reuse

### Goal

Research important topics once and reuse approved findings in future plans.

### Key Files

- `internal/research/service.go`
- `internal/research/orchestrator.go`
- `internal/store/research_repositories.go`
- `internal/knowledge/service.go`
- `internal/store/knowledge_repositories.go`

### Current Behavior

- Research and knowledge are structured.
- Knowledge has FTS-backed search.
- Research reuse is not mandatory before creating new research.

### Implementation Path

- Add `research.ReuseService`.
- Add `FindReusableResearch(projectID, topic)`.
- Add FTS-backed research search.
- Exclude draft/unapproved research from planning.
- Promote approved research into reusable knowledge.

### Rule

New research is created only when no approved reusable research exists or when the existing research is stale/insufficient.

## Phase 7: Permanent Memory Recorder

### Goal

Make Plan-AI's memory durable and automatic, not dependent on the current model session.

### Key Files

- `internal/memory/memory.go`
- `internal/store/memory_repository.go`
- `internal/store/store.go`

### Current Behavior

- Project memory exists.
- FTS exists for memory.
- Important events do not automatically write memory entries.

### Implementation Path

- Add `memory.Recorder`.
- Record memory automatically from:
  - approved context
  - approved research
  - approved plans
  - applied proposals
  - change events
- Add artifact links and topic keys.
- Use FTS-backed memory lookup for question answering and context reuse.

Suggested fields:

- `topic_key`
- `confidence`
- `artifact_type`
- `artifact_id`
- `supersedes_id`
- `scope`

## Phase 8: Continuous Planning Loop

### Goal

Make Plan-AI accompany the project lifecycle from idea to implementation changes and replanning.

### Key Files

- `internal/continuous/*`
- `internal/change/*`
- `internal/mcp/handlers.go`

### Current Behavior

- Detector, planner, updater, status, context generation, and MCP tools exist.
- The full loop is not wired end-to-end.
- Some MCP handlers bypass service modules and use direct SQL.

### Implementation Path

Add `continuous.LoopService`.

Flow:

```txt
detect -> analyze -> propose -> wait for approval -> apply
```

Rules:

- Plan changes must never happen silently.
- Applying a proposal must be idempotent.
- Rejected/superseded facts must not drive planning context.
- Continuous status and context generation should use service modules, not duplicated SQL in handlers.

## ADRs Needed

- Supersede `docs/adr/0002-storage-layer.md` with external project storage.
- Supersede/update `docs/adr/0017-opencode-integration.md` for safe install-time auto-configuration.
- Add ADR for conversation-first gateway.
- Add ADR for approved context as authority boundary if the unified projection is accepted.

## Non-Negotiable Invariants

- The repository stays clean by default.
- The source of truth lives in Plan-AI, not in the agent or model session.
- Approved context is never re-asked unless ambiguous or superseded.
- Research is reused before new research is requested.
- Planning starts only after enough discovery/intent context exists.
- Continuous planning proposes changes; it does not silently mutate approved work.
- CLI, MCP, and conversation paths must share the same services and invariants.

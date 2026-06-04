# Plan-AI

**Local-first continuous implementation planning for AI-assisted projects.**

Plan-AI prepares implementation plans. It does not implement code.

Plan-AI owns the durable truth of a project — approved context, decisions, research, knowledge, plans, tasks, snapshots, and exported documents — in local SQLite stores. LLMs may help analyze or summarize, but Plan-AI's store is the source of truth.

## Quick start

```bash
# Install global persistence
go run ./cmd/plan-ai install

# Initialize for the current project
go run ./cmd/plan-ai init

# View status
go run ./cmd/plan-ai status

# Scan the project
go run ./cmd/plan-ai scan

# Ingest input and create vision
go run ./cmd/plan-ai ingest --type prompt --content "Build a planning assistant."
go run ./cmd/plan-ai vision draft

# Store approved context
go run ./cmd/plan-ai approved add --type requirement "The app must save planning drafts"

# Add research and knowledge
go run ./cmd/plan-ai research add --topic "Research topic" --summary "..."
go run ./cmd/plan-ai knowledge add --topic "Architecture pattern" --content "..."

# Generate plan artifacts
go run ./cmd/plan-ai plan

# Check integration health
go run ./cmd/plan-ai doctor
```

## Architecture

Plan-AI is organized in layers:

| Layer | Package | Purpose |
|-------|---------|---------|
| CLI | `cmd/plan-ai/` | Cobra command tree (20+ commands) |
| MCP Server | `cmd/mcp-server/` | stdio-based MCP interface (30 tools) |
| Core | `internal/core/` | App metadata, version |
| Config | `internal/config/` | Global/project config paths |
| Domain | `internal/domain/` | Canonical entity model |
| Store | `internal/store/` | SQLite persistence, migrations, repositories |
| Scanner | `internal/scanner/` | Deterministic project scanner |
| Ingestion | `internal/ingestion/` | Input classification and ingestion |
| Vision | `internal/vision/` | Vision draft creation and approval |
| Context | `internal/context/` | Approved context management |
| Research | `internal/research/` | Research entries, findings, sources, conclusions |
| Knowledge | `internal/knowledge/` | Reusable technical knowledge base |
| Planning | `internal/planning/` | Master plans, specific plans, implementation docs |
| Change | `internal/change/` | Change detection, impact analysis, snapshots |
| Workflows | `internal/workflows/` | Workflow execution registry |
| Agent | `internal/agent/` | Intent detection, routing, delegation |
| Continuous | `internal/continuous/` | Event detection, plan update proposals |
| MCP | `internal/mcp/` | Tool definitions and handlers |
| OpenCode | `internal/opencode/` | Optional OpenCode integration artifacts |
| Model Strategy | `internal/modelstrategy/` | LLM provider registry and budget tracking |
| Capabilities | `internal/capabilities/` | Capability registry |
| Orchestrator | `internal/orchestrator/` | Job queue and orchestration |
| Discovery | `internal/vision/` | Vision discovery sessions |
| Delivery | `internal/context/` | Context delivery (L0-L4) |
| Scanner | `internal/scanner/` | Stack and dependency detection |
| Validation | `internal/validation/` | Validation resources |
| Integrations | `internal/integrations/` | Integration surface |
| Skills | `internal/skills/` | Skills/resources |

## CLI

Full reference: [docs/cli-reference.md](docs/cli-reference.md)

Key commands:
```
plan-ai install        Install global persistence
plan-ai init           Initialize project store
plan-ai status         Show persistence and domain status
plan-ai scan           Deterministic project scan
plan-ai ingest         Classify and store input
plan-ai vision         Create/approve vision drafts
plan-ai approved       Manage approved context
plan-ai research       Research entries with findings/sources/conclusions
plan-ai knowledge      Reusable knowledge base
plan-ai plan           Generate planning artifacts
plan-ai context        Executive context overview
plan-ai capabilities   List registered capabilities
plan-ai doctor         Check store paths, migrations, OpenCode health
plan-ai agent          Agent system (status, process, list)
plan-ai continuous     Continuous planning (status, events, proposals)
plan-ai next           Get next pending task
plan-ai setup opencode Generate OpenCode integration artifacts
plan-ai validate       Run V2 validation suites
plan-ai dev            Development inspection helpers
```

## MCP

Plan-AI exposes a stdio MCP server at `cmd/mcp-server/` with 30 tools covering:
- Project initialization and status
- Master/specific plan creation and approval
- Research, knowledge, and context management
- Agent processing and continuous planning
- Change detection, snapshots, and export

Full reference: [docs/mcp-reference.md](docs/mcp-reference.md)

## Sandbox safety

**Never test against real user paths.** Use sandbox environment variables:

```bash
HOME="$PWD/.tmp/home" \
PLAN_AI_HOME="$PWD/.tmp/home" \
PLAN_AI_PROJECT_ROOT="$PWD/.tmp/project" \
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" \
go run ./cmd/plan-ai status
```

Full sandbox validation:
```bash
bash scripts/test-sandbox.sh
```

Sandbox markers verified:
- `REAL_GLOBAL_ABSENT` — no real `~/.plan-ai` exists
- `REAL_PROJECT_ABSENT` — no real `.plan-ai` in project root
- `REAL_OPENCODE_ABSENT` — no real `~/.config/opencode` exists
- `SANDBOX_CLEANED` — sandbox cleaned after test

## Storage model

Plan-AI uses two SQLite databases per machine:

- **Global store** (`~/.plan-ai/global.db`): known projects
- **Project store** (`<project>/.plan-ai/project.db`): all domain data

Both are created and migrated via `plan-ai install` / `plan-ai init`.

## OpenCode integration

Optional, sandbox-scoped. Generate integration artifacts:

```bash
plan-ai setup opencode
```

Artifacts (all under `$OPENCODE_CONFIG_DIR`):
- `opencode.json` — minimal OpenCode config
- `mcp-registry.json` — MCP tool registry
- `agents/plan-ai.json` — agent descriptor
- `profiles.json` — integration profiles
- `prompts.json` — prompt templates
- `<project>/.plan-ai/opencode-sync.json` — sync marker

## Development

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
```

## External docs

- [Architecture](docs/architecture.md)
- [CLI reference](docs/cli-reference.md)
- [MCP reference](docs/mcp-reference.md)
- [OpenCode integration guide](docs/opencode-integration-guide.md)
- [Installation guide](docs/installation.md)
- [Project structure](docs/project-structure.md)
- [Release notes](RELEASE_NOTES.md)

## License

MIT

# Plan-AI

**Local-first continuous implementation planning for AI-assisted projects.**

Plan-AI prepares implementation plans. It does not implement code. It stores approved context, decisions, research, knowledge, plans, tasks, snapshots, and exported documents in local SQLite stores so the project has a durable source of truth before AI agents start coding.

## Why Plan-AI exists

AI coding fails when the plan lives only in chat. Plan-AI gives the project its own planning memory:

- Product intent before implementation.
- Progressive discovery before task generation.
- Ambiguity and confidence checks before coding.
- Alignment reports that connect tasks back to the approved intent.
- Local-first persistence, no required hosted service.

## Install

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
bash scripts/install.sh
```

Then make sure the install prefix is on PATH if needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Verify:

```bash
plan-ai doctor
plan-ai init
plan-ai status
```

More details: [docs/install.md](docs/install.md).

## Quickstart

```bash
plan-ai ingest --type prompt --content "Build a planning assistant for product teams. It must use SQLite."
plan-ai vision draft
plan-ai approved add --type requirement "Plans must be stored locally."
plan-ai approved add --type decision "Use SQLite as the source of truth."
plan-ai plan master
plan-ai context
```

V3 Product Intent flow:

```bash
plan-ai intent discover "Quiero crear un CRM para talleres mecánicos"

plan-ai intent create \
  --description "CRM for mechanic workshops" \
  --expected-outcome "Workshops can track customers, vehicles, jobs, and follow-ups" \
  --desired-experience "Simple, fast, Spanish-first workflow" \
  --desired-result "A clear implementation plan before coding" \
  --success-definition "A workshop owner can manage jobs without spreadsheets" \
  --failure-definition "The tool becomes a generic CRM with no workshop-specific workflow"

plan-ai intent list
plan-ai discovery init --intent <pintent_id>
plan-ai ambiguity analyze --intent <pintent_id>
plan-ai confidence evaluate --intent <pintent_id>
plan-ai alignment framework --intent <pintent_id>
```

Full guide: [docs/quickstart.md](docs/quickstart.md).

## Core commands

```text
plan-ai install        Install or migrate global persistence
plan-ai init           Initialize project persistence
plan-ai status         Show store and project status
plan-ai doctor         Check stores, migrations, and integrations
plan-ai scan           Deterministic project scan
plan-ai ingest         Store user input for planning
plan-ai vision         Create and approve vision artifacts
plan-ai approved       Manage approved project context
plan-ai research       Store research and findings
plan-ai knowledge      Store reusable project knowledge
plan-ai plan           Generate planning artifacts
plan-ai intent         Detect V2 intent and manage V3 Product Intent
plan-ai discovery      Run progressive discovery for a Product Intent
plan-ai ambiguity      Analyze missing information and assumptions
plan-ai confidence     Score how well Plan-AI understands intent
plan-ai alignment      Review implementation alignment to intent
plan-ai setup opencode Generate safe OpenCode integration artifacts
plan-ai validate       Run deterministic validation suites
```

Full reference: [docs/cli-reference.md](docs/cli-reference.md).

## Storage model

Plan-AI uses SQLite:

- Global store: `~/.plan-ai/global.db`
- Project store: `<project>/.plan-ai/project.db`

Runtime data is intentionally ignored by git. Do not commit `.plan-ai/`, SQLite databases, logs, `.env` files, tokens, or generated binaries.

## OpenCode integration

Safe sandbox mode:

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai setup opencode
```

Real OpenCode config writes require explicit opt-in:

```bash
plan-ai setup opencode --allow-real-opencode
```

Guide: [docs/opencode-integration.md](docs/opencode-integration.md).

## Validation

Development gate:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
bash scripts/test-vps-clean.sh
bash scripts/release-check.sh
```

Manual scenario: [docs/manual-validation.md](docs/manual-validation.md).  
Clean VPS guide: [docs/vps-validation.md](docs/vps-validation.md).

## Architecture

| Layer | Package | Purpose |
|-------|---------|---------|
| CLI | `cmd/plan-ai/` | Cobra command tree |
| MCP Server | `cmd/mcp-server/` | stdio MCP interface |
| Core | `internal/core/` | App metadata |
| Config | `internal/config/` | Global/project config paths |
| Store | `internal/store/` | SQLite persistence, migrations, repositories |
| Planning | `internal/planning/` | Master plans, specific plans, implementation docs |
| Intent V3 | `internal/intentv3/` | Product Intent and deterministic discovery |
| Discovery V3 | `internal/discoveryv3/` | Progressive discovery questions |
| Ambiguity V3 | `internal/ambiguityv3/` | Missing information and assumption analysis |
| Confidence V3 | `internal/confidencev3/` | Intent confidence scoring |
| Alignment V3 | `internal/alignmentv3/` | Intent-to-implementation alignment reports |
| OpenCode | `internal/opencode/` | Optional OpenCode integration artifacts |
| MCP | `internal/mcp/` | Tool definitions and handlers |

## Documentation

- [Installation](docs/install.md)
- [Quickstart](docs/quickstart.md)
- [CLI reference](docs/cli-reference.md)
- [Manual validation](docs/manual-validation.md)
- [VPS validation](docs/vps-validation.md)
- [OpenCode integration](docs/opencode-integration.md)
- [Troubleshooting](docs/troubleshooting.md)
- [MCP reference](docs/mcp-reference.md)
- [Architecture](docs/architecture.md)
- [Project structure](docs/project-structure.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md), and [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).

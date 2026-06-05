# CLI Reference

Plan-AI exposes a Cobra command tree. Commands run as `plan-ai <command>` after installation or as `go run ./cmd/plan-ai <command>` from the repository.

## Global flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for any command |
| `--config` | Config file path, defaulting under `~/.plan-ai` |

## Runtime stores

- Global store: `~/.plan-ai/global.db`
- Project store: `<project>/.plan-ai/project.db`

Use `PLAN_AI_HOME` and `PLAN_AI_PROJECT_ROOT` for sandboxed validation.

## Commands

### `plan-ai install`

Create or migrate global persistence.

```bash
plan-ai install
```

Idempotent. Creates the global store if missing and applies migrations.

### `plan-ai init`

Create or migrate project persistence.

```bash
plan-ai init
```

### `plan-ai bootstrap`

Install/migrate global persistence, initialize the current project, and generate OpenCode/MCP artifacts.

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai bootstrap
plan-ai bootstrap --allow-real-opencode
```

Registers the current project and creates `<project>/.plan-ai/project.db`.

### `plan-ai status`

Show persistence and domain status.

```bash
plan-ai status
```

Includes store status, scan state, and domain counts.

### `plan-ai doctor`

Check global/project stores, migrations, and integration health.

```bash
plan-ai doctor
```

Exit code is non-zero when checks fail.

### `plan-ai scan`

Run deterministic project scan.

```bash
plan-ai scan
```

Stores project metadata, file counts, stack hints, and fingerprint information.

### `plan-ai ingest`

Classify and store input.

```bash
plan-ai ingest --type prompt --content "Build a planning assistant" --source cli
```

Supported types include `prompt`, `requirement`, `spec`, `ticket`, and `feedback`.

### `plan-ai vision`

Create and manage vision artifacts.

```bash
plan-ai vision draft
plan-ai vision list --limit 10
plan-ai vision get <id>
plan-ai vision approve <id>
plan-ai vision status
plan-ai vision begin
plan-ai vision discover
plan-ai vision conclude
plan-ai vision finalize
plan-ai vision document --intent <intent_id>
plan-ai vision documents
plan-ai vision document-show <document_id>
plan-ai vision approve-document <document_id>
```

### `plan-ai approved`

Manage approved context.

```bash
plan-ai approved add --type requirement "The app must save planning drafts"
plan-ai approved add --type decision "Use SQLite for local persistence"
plan-ai approved list
```

### `plan-ai research`

Manage research entries.

```bash
plan-ai research add --topic "Local-first planning" --summary "..." --source "manual"
plan-ai research list --limit 10
plan-ai research get <id>
plan-ai research findings add --research-id <id> --finding "..."
plan-ai research sources add --research-id <id> --url "https://example.com" --description "..."
plan-ai research conclusions add --research-id <id> --conclusion "..."
```

### `plan-ai knowledge`

Manage reusable project knowledge.

```bash
plan-ai knowledge add --topic "Architecture" --content "Use local SQLite stores"
plan-ai knowledge list
plan-ai knowledge show <id>
plan-ai knowledge search sqlite
plan-ai knowledge reuse <id>
```

### `plan-ai plan`

Generate planning artifacts.

```bash
plan-ai plan master
plan-ai plan specific
plan-ai plan impl-doc
plan-ai plan approve
plan-ai plan list
```

### `plan-ai context`

Show executive context overview.

```bash
plan-ai context
```

### `plan-ai capabilities`

List registered capabilities.

```bash
plan-ai capabilities
```

### `plan-ai intent`

Manage V2 intent profiles and V3 Product Intents.

V2 profile detection:

```bash
plan-ai intent detect --content "quiero un SaaS CRM"
plan-ai intent latest
plan-ai intent show <intent_id>
plan-ai intent approve <intent_id>
```

V3 Product Intent:

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
plan-ai intent submit <pintent_id>
plan-ai intent approve <pintent_id>
plan-ai intent show <pintent_id>
```

There is no `intent analyze` command in this release; use `intent discover` for raw ideas and `ambiguity analyze` for analysis.

### `plan-ai discovery`

Run progressive discovery for a V3 Product Intent.

```bash
plan-ai discovery init --intent <pintent_id>
plan-ai discovery next --intent <pintent_id>
plan-ai discovery next --intent <pintent_id> --level master_plan
plan-ai discovery answer --intent <pintent_id> --question <question_id> --answer "..."
plan-ai discovery v3-status --intent <pintent_id>
```

`init` creates deterministic questions across project, master plan, specific plan, phase, and task levels.

### `plan-ai ambiguity`

Analyze knowns, unknowns, assumptions, conflicts, and missing information.

```bash
plan-ai ambiguity analyze --input "Quiero un CRM moderno"
plan-ai ambiguity analyze --intent <pintent_id>
```

### `plan-ai confidence`

Evaluate how well Plan-AI understands a Product Intent.

```bash
plan-ai confidence evaluate --intent <pintent_id>
```

Output includes component scores and final intent confidence.

### `plan-ai alignment`

Review whether outcomes, plans, and tasks remain aligned to Product Intent.

```bash
plan-ai alignment review --intent <pintent_id> --outcome "..." --plan "..." --task "..."
plan-ai alignment context --intent <pintent_id>
plan-ai alignment references
plan-ai alignment framework --intent <pintent_id>
```

### `plan-ai agent`

Agent system management.

```bash
plan-ai agent status
plan-ai agent process
plan-ai agent list
```

### `plan-ai continuous`

Continuous planning engine.

```bash
plan-ai continuous status
plan-ai continuous events
plan-ai continuous proposals
```

### `plan-ai next`

Show the highest-priority unfinished task.

```bash
plan-ai next
```

### `plan-ai setup opencode`

Generate OpenCode integration artifacts.

Safe mode:

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai setup opencode
```

Real OpenCode config requires explicit opt-in:

```bash
plan-ai setup opencode --allow-real-opencode
```

### `plan-ai validate`

Run deterministic validation suites.

```bash
plan-ai validate v2
plan-ai validate cases
```

### `plan-ai dev`

Development inspection helpers.

```bash
plan-ai dev seed-domain
plan-ai dev list-domain
plan-ai dev seed-knowledge
plan-ai dev store <key>
plan-ai dev reset
```

`dev reset` is destructive for the current project store. Do not use it against user data.

## Release validation commands

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
bash scripts/test-vps-clean.sh
bash scripts/release-check.sh
```

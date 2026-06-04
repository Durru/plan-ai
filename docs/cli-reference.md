# CLI Reference

Plan-AI exposes a Cobra command tree with 20+ commands. All commands run via `go run ./cmd/plan-ai <command>` or the built binary `plan-ai <command>`.

## Global flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for any command |
| `--config` | Config file (default `~/.plan-ai/config.yaml`) |

## Commands

### `plan-ai install`

Install global persistence.

```bash
plan-ai install
```

Creates `~/.plan-ai/` with global SQLite store and schema migrations. Safe to re-run — idempotent. Exits with error if global store exists but migration fails.

### `plan-ai init`

Initialize project store.

```bash
plan-ai init
```

Creates `<project>/.plan-ai/` with project-scoped SQLite store and registers the project in the global store. Idempotent.

### `plan-ai status`

Show persistence and domain status.

```bash
plan-ai status
```

Output includes:
- Global store path and size
- Project store path and size
- Whether both stores are healthy
- Scanned project info

### `plan-ai scan`

Deterministic project scan.

```bash
plan-ai scan
```

Scans the project directory for:
- Go module name and version
- File counts and language distribution
- Main entry points
- Stack and dependency detection

Output is deterministic JSON stored in the project store.

### `plan-ai ingest`

Classify and store input as a vision artifact.

```bash
plan-ai ingest --type <type> --content <content> [--source <source>]
```

Arguments:
| Flag | Description | Required |
|------|-------------|----------|
| `--type` | Input type (`prompt`, `requirement`, `spec`, `ticket`, `feedback`) | Yes |
| `--content` | Raw content | Yes |
| `--source` | Source identifier | No |

### `plan-ai vision`

Create and manage vision drafts.

```bash
plan-ai vision draft                    # Create new vision draft from ingested items
plan-ai vision list [--limit N]         # List vision drafts
plan-ai vision get <id>                 # Get a specific vision draft
plan-ai vision approve <id>             # Approve a vision draft
plan-ai vision status                   # Show discovery session status
plan-ai vision begin                    # Begin discovery session
plan-ai vision discover                 # Run discovery iteration
plan-ai vision conclude                 # Conclude discovery
plan-ai vision finalize                 # Finalize vision
```

### `plan-ai approved`

Manage approved context.

```bash
plan-ai approved list [--type <type>]               # List approved items
plan-ai approved add --type <type> --content <c>    # Add approved item
```

### `plan-ai research`

Manage research entries.

```bash
plan-ai research add --topic <topic> --summary <summary> [--source <source>]
plan-ai research list [--topic <topic>] [--limit N]
plan-ai research get <id>
plan-ai research findings add --research-id <id> --finding <finding> [--source <source>]
plan-ai research sources add --research-id <id> --url <url> --description <desc>
plan-ai research conclusions add --research-id <id> --conclusion <conclusion>
```

### `plan-ai knowledge`

Manage the reusable knowledge base.

```bash
plan-ai knowledge add --topic <topic> --content <content> [--source <source>]
plan-ai knowledge list [--topic <topic>] [--limit N]
plan-ai knowledge get <id>
```

### `plan-ai plan`

Generate planning artifacts.

```bash
plan-ai plan master        # Generate master plan
plan-ai plan specific      # Generate specific plan
plan-ai plan impl-doc      # Generate implementation document
plan-ai plan approve       # Approve plan
plan-ai plan list          # List plans
```

### `plan-ai context`

Executive context overview.

```bash
plan-ai context
```

Shows:
- Vision status
- Total approved items
- Plan count and approval status
- Research and knowledge counts
- Project info

### `plan-ai capabilities`

List registered capabilities.

```bash
plan-ai capabilities
```

Outputs all registered capability IDs with descriptions.

### `plan-ai doctor`

Check store paths, migrations, and OpenCode integration health.

```bash
plan-ai doctor
```

Runs the following checks:
- Global store file exists and is readable
- Global store has all required migrations applied
- Project store file exists and is readable
- Project store has all required migrations applied
- OpenCode artifact directory exists
- All expected OpenCode artifacts are present

Exit code is 0 if all checks pass, 1 if any check fails. Each check outputs `[PASS]` or `[FAIL]`.

### `plan-ai agent`

Agent system management.

```bash
plan-ai agent status            # Show agent status
plan-ai agent process           # Trigger agent processing
plan-ai agent list              # List agents and their capabilities
```

### `plan-ai continuous`

Continuous planning engine.

```bash
plan-ai continuous status       # Show continuous planning status
plan-ai continuous events       # List detected events
plan-ai continuous proposals    # List generated proposals
```

### `plan-ai next`

Get the next pending task.

```bash
plan-ai next
```

Returns the highest-priority unfinished task or `info: no pending tasks`.

### `plan-ai setup opencode`

Generate OpenCode integration artifacts.

```bash
plan-ai setup opencode
```

Generates all six integration artifacts under `$OPENCODE_CONFIG_DIR`:
1. `opencode.json` — minimal config
2. `mcp-registry.json` — tool registry
3. `agents/plan-ai.json` — agent descriptor
4. `profiles.json` — integration profiles
5. `prompts.json` — prompt templates
6. `.plan-ai/opencode-sync.json` — sync marker

Exits with error if `OPENCODE_CONFIG_DIR` is not set.

### `plan-ai validate`

V2 validation suites.

```bash
plan-ai validate v2           # Run all 63 V2 validation checks (7 cases × 9 stages)
plan-ai validate cases        # List all 7 project categories used in V2 validation
```

The `validate v2` command runs a deterministic in-memory validation engine that checks every project case through every V2 stage. Reports total/passed/failed counts. Exits with error if any check fails.

The `validate cases` command lists each project category (SaaS, Ecommerce, Landing Page, MCP Server, Mobile App, API, CRM) with description, idea, and intent count.

### `plan-ai dev`

Development inspection helpers.

```bash
plan-ai dev store <key>         # Inspect store contents
plan-ai dev reset               # Reset project store
```

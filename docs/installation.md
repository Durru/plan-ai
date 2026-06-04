# Installation Guide

## Prerequisites

- Go 1.21+
- SQLite3 (libsqlite3-dev or equivalent)
- A project directory (existing or new)

## Install from source

```bash
# Clone the repository
git clone <repo-url> plan-ai
cd plan-ai

# Build the binary
go build -o plan-ai ./cmd/plan-ai
go build -o plan-ai-mcp-server ./cmd/mcp-server

# (Optional) Install to PATH
sudo mv plan-ai /usr/local/bin/
sudo mv plan-ai-mcp-server /usr/local/bin/
```

## Quick start

```bash
# 1. Install global persistence
plan-ai install

# 2. Initialize project
cd /path/to/your/project
plan-ai init

# 3. Scan the project
plan-ai scan

# 4. Ingest initial context
plan-ai ingest --type prompt --content "Build a planning assistant."

# 5. Create vision
plan-ai vision draft
plan-ai vision approve $(plan-ai vision list | head -1)

# 6. Add research and context
plan-ai research add --topic "Architecture" --summary "Hexagonal architecture"
plan-ai approved add --type requirement "The system must be local-first"

# 7. Generate plan
plan-ai plan master
plan-ai plan specific
```

## Verify installation

```bash
plan-ai doctor
```

Expected output: all checks `[PASS]`.

## Development

```bash
# Format code
gofmt -w cmd internal

# Run tests
go test ./...

# Run vet
go vet ./...

# Build
go build ./...

# Sandbox validation
bash scripts/test-sandbox.sh
```

## Sandbox testing

For safe testing without touching real user directories:

```bash
export HOME="$PWD/.tmp/home"
export PLAN_AI_HOME="$PWD/.tmp/home"
export PLAN_AI_PROJECT_ROOT="$PWD/.tmp/project"
export OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config"

plan-ai install
plan-ai init
plan-ai scan
plan-ai setup opencode
plan-ai doctor
```

Or use the pre-built sandbox script:

```bash
bash scripts/test-sandbox.sh
```

The sandbox script verifies:
- All CLI commands exit 0
- Store files are created in expected sandbox paths
- OpenCode integration artifacts are generated
- E2E flow (install → init → scan → ingest → plan) succeeds
- Continuous planning scenario completes
- No real user paths are touched (sandbox marker verification)

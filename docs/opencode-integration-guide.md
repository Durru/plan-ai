# OpenCode Integration Guide

Plan-AI can optionally integrate with [OpenCode](https://opencode.ai) for enhanced AI-assisted planning. This integration is **optional, sandbox-scoped, and zero-dependency** — Plan-AI works fully without it.

## Architecture

The integration generates static JSON artifacts that describe Plan-AI's capabilities to an OpenCode environment. Artifacts are generated into `$OPENCODE_CONFIG_DIR` and include:

```
$OPENCODE_CONFIG_DIR/
├── opencode.json            # Minimal OpenCode config
├── mcp-registry.json        # MCP tool registry
├── agents/plan-ai.json      # Agent descriptor
├── profiles.json            # Integration profiles
└── prompts.json             # Prompt templates

<project>/.plan-ai/
└── opencode-sync.json       # Sync marker
```

## Generating artifacts

```bash
export OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config"
plan-ai setup opencode
```

**Important:** `OPENCODE_CONFIG_DIR` must be set. The command exits with an error if it is not.

## Verification

```bash
plan-ai doctor
```

The `doctor` command checks:
- OpenCode artifact directory exists
- All expected artifacts are present

## Artifact details

### `opencode.json`

Minimal configuration that registers Plan-AI's MCP server and agents.

### `mcp-registry.json`

Describes all 28 MCP tools with input schemas for tool discovery.

### `agents/plan-ai.json`

Agent descriptor with capabilities, triggers, and slot configuration.

### `profiles.json`

Integration profiles mapping Plan-AI features to OpenCode workflows.

### `prompts.json`

Prompt templates for common planning workflows.

### `opencode-sync.json`

Sync marker recording the last integration artifact generation timestamp.

## No-dependency design

Plan-AI **never** imports any OpenCode packages, reads any OpenCode config, or depends on OpenCode's runtime. The integration is purely:
- CLI command (`plan-ai setup opencode`) that writes static JSON files
- Doctor check that verifies artifacts exist on disk

This ensures Plan-AI remains fully usable without any external integration.

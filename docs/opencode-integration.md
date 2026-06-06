# OpenCode Integration

Plan-AI can generate OpenCode integration artifacts so an OpenCode setup can call Plan-AI planning tools through MCP.

The integration is safe by default: it does **not** write to a real OpenCode config unless the user explicitly opts in.

## Safe sandbox setup

Recommended for tests and validation:

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai bootstrap
```

Generated artifacts:

- `$OPENCODE_CONFIG_DIR/opencode.json`
- `$OPENCODE_CONFIG_DIR/mcp-registry.json`
- `$OPENCODE_CONFIG_DIR/agents/plan-ai.json`
- `$OPENCODE_CONFIG_DIR/profiles.json`
- `$OPENCODE_CONFIG_DIR/prompts.json`
- `$OPENCODE_CONFIG_DIR/plan-ai-workflows.json`
- `<project>/.plan-ai/opencode-sync.json`

## Real OpenCode setup

Only use this when you intentionally want to update the real OpenCode config area:

```bash
plan-ai bootstrap --allow-real-opencode
```

If the project is already initialized and you only want to regenerate integration artifacts, use:

```bash
plan-ai setup opencode --allow-real-opencode
```

The generated `opencode.json` includes a local MCP entry:

```json
{
  "mcp": {
    "plan-ai": {
      "type": "local",
      "enabled": true,
      "command": ["plan-ai", "mcp", "serve"]
    }
  }
}
```

If `OPENCODE_CONFIG_DIR` is not set and `--allow-real-opencode` is not passed, the command exits with an error. This protects `~/.config/opencode` from accidental test writes.

## Integration modes

| Mode | Description |
|------|-------------|
| `standalone` | Plan-AI runs independently, no MCP bridge |
| `tool` | Plan-AI exposes tools for OpenCode MCP |
| `hybrid` | Both modes active simultaneously |

## Doctor

```bash
plan-ai doctor
```

Doctor checks store paths, migrations, and OpenCode integration health when integration state exists.

## Architecture

Relevant packages:

- `internal/opencode/` — detection, config, registry, doctor checks, artifact generation.
- `internal/mcp/` — MCP tool definitions and handlers.
- `cmd/plan-ai/` — CLI entry point; `plan-ai mcp serve` serves MCP over stdio.

## Safety rules

- Use `OPENCODE_CONFIG_DIR` in tests.
- Do not commit generated OpenCode config containing local paths or secrets.
- Do not run with `--allow-real-opencode` in CI.
- Keep `.plan-ai/` out of git; the sync marker is runtime state.

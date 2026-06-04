# OpenCode Integration

Plan-AI integrates with OpenCode to detect existing configuration and
expose planning tools via MCP. The integration is **read-only by default**
and operates in one of three modes.

## Integration Modes

| Mode | Description |
|------|-------------|
| `standalone` | Plan-AI runs independently, no MCP bridge |
| `tool` | Plan-AI exposes tools for OpenCode MCP (default) |
| `hybrid` | Both modes active simultaneously |

## Components

### Detector

The OpenCode detector (`internal/opencode/detector.go`) searches for
OpenCode configuration in:

1. `opencode.json` in project root
2. `opencode.jsonc` in project root
3. `.opencode/opencode.json` (or .jsonc) in project root
4. Parent directory `opencode.json[c]` for workspace-level config

It extracts: agent name, agent role, skill count, and self-init capability.

### Config

The integration config (`internal/opencode/config.go`) is stored at
`<plan-ai-home>/opencode-integration.json`:

```json
{
  "enabled": true,
  "mode": "tool",
  "auto_detect": true,
  "warn_on_conflict": true,
  "read_only": true,
  "doctor_checks": ["version", "config", "mcp"]
}
```

### Tool Registry

The MCP tool registry (`internal/opencode/registry.go`) maintains the list
of tools that Plan-AI exposes to OpenCode, tracking how many plans each
tool has built.

### Doctor

The doctor (`internal/opencode/doctor.go`) runs integration health checks:

- **version** — checks OpenCode config compatibility
- **config** — validates integration config integrity
- **mcp** — verifies MCP tool registration

### Database Tables

The OpenCode integration persists to:
- `opencode_detections` — detection history
- `opencode_integration_state` — current integration mode
- `opencode_doctor_checks` — health check results

## Doctor CLI

```sh
plan-ai doctor              # includes opencode checks when integration is enabled
plan-ai doctor --help       # show all check options
```

## Architecture Decision

Per ADR-0017, the integration is deliberately **read-only**: Plan-AI detects
OpenCode configuration but never modifies it. This prevents accidental
corruption of OpenCode setup while still enabling co-existence.

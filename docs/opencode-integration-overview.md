# OpenCode Integration Overview

OpenCode integration is optional and comes after Plan-AI is useful independently.

## Current state

Plan-AI does not modify OpenCode configuration. No `setup opencode` command is active yet.

## Direction

Phase 20 will add OpenCode integration after the definitive MCP layer exists.

Expected behavior:

- detect OpenCode safely
- generate optional MCP/agent configuration in sandbox first
- ask before touching real config
- keep Plan-AI usable without OpenCode

## Safety rule

Never modify real `~/.config/opencode` during tests. Use sandbox paths such as:

```bash
HOME="$PWD/.tmp/home"
PLAN_AI_HOME="$PWD/.tmp/home"
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config"
```

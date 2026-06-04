# Plan-AI Agent Instructions

Plan-AI prepares implementation. It does not improvise implementation.

## Rules

- Respect the current roadmap in `docs/roadmap.md`.
- Do not continue archived old phases unless explicitly requested.
- Keep CLI, Core, Store, MCP, Integration, and Agent layers separate.
- Use sandbox paths for install/init/status/scan tests.
- Do not touch real `~/.plan-ai`.
- Do not touch real `~/.config/opencode`.
- Do not configure OpenCode unless the current task explicitly asks for that phase.
- Document important architectural changes.
- Run relevant tests before claiming completion.

## Sandbox contract

Prefer repo-local sandbox paths:

```bash
HOME="$PWD/.tmp/home"
PLAN_AI_HOME="$PWD/.tmp/home"
PLAN_AI_PROJECT_ROOT="$PWD/.tmp/project"
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config"
```

## Current next phase

The next implementation phase is **Phase 7 — Definitive Domain Model**.

Do not implement Skill Intelligence, MCP, OpenCode integration, or agents before the roadmap reaches those phases.

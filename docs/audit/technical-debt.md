# Technical Debt

## Compatibility debt

Several phases introduced compatibility views or mirrored writes because earlier tables already existed with overlapping names. Examples include `agent_runs`, `tool_runs`, `tool_audit`, and context delivery compatibility views.

## Migration debt

Runtime migrations remain inline in `internal/store/store.go` with SQL mirror files under `internal/migrations/project/`. This matches current project convention but should eventually be extracted into a single source of truth.

## Repository debt

Some repositories mirror writes between legacy and definitive tables to maintain compatibility. This should be revisited once a stable v1 schema is frozen.

## CLI debt

The CLI is intentionally minimal but large in one file. It should later be split by command groups.

## MCP debt

MCP tool registration and validation are functional, but service binding is incomplete for production use. Stub handlers must be replaced by full service-layer handlers before external use.

## Audit provenance debt

The repository lacks a tracked baseline commit, making diff-based audit impossible. A baseline commit is strongly recommended before future work.

# MVP Gaps

## Functional gaps

- Real LLM provider execution is not connected; Plan-AI prepares contracts, schemas, retries, budgets, and orchestration.
- MCP server tools are registered and validated, but some handlers remain MVP stubs until bound to full service workflows.
- OpenCode integration is read-only detection, not bidirectional configuration or automatic install.
- Vision Discovery and plan generation are AI-first scaffolds; they do not replace LLM reasoning.

## Testing gaps

- Sandbox validates local SQLite/CLI/MCP paths, but not real external providers.
- Stress/benchmark suites are not yet dedicated commands.
- No baseline Git commit exists, so future touched-file audits cannot rely on diffs.

## Product gaps

- No TUI.
- No advanced Skill Intelligence.
- No Engram deep integration.
- No production packaging/release workflow.

## Integration gaps

- No real OpenCode mutation by design.
- No real network URL/image/document fetch pipeline by design; structures exist for future adapters.

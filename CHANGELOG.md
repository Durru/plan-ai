# Changelog

All notable changes to Plan-AI are documented here.

## Unreleased

### Added

- Open source release documentation and validation flow.
- Safe install and uninstall scripts.
- Clean VPS validation script.
- Release gate script that runs formatting, tests, vet, build, sandbox validation, and VPS-style validation.
- GitHub Actions CI workflow.

### Safety

- Release validation checks that local runtime data, SQLite databases, logs, `.env` files, tokens, and root binaries are not accidentally committed.
- OpenCode integration remains opt-in for real config mutation.

## v3 planning layer

### Added

- Product Intent Engine.
- Intent Discovery Engine.
- Progressive Discovery System.
- Ambiguity Detection Engine.
- Intent Confidence Engine.
- Deterministic Alignment Framework.

## v2 planning layer

### Added

- Project Memory System.
- Model Compatibility Layer.
- Real Project Validation.
- V2 release/audit hardening.

## MVP

### Added

- Local-first SQLite persistence.
- CLI and MCP surfaces for planning, context, research, knowledge, change tracking, and OpenCode integration.

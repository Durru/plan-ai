# Contributing to Plan-AI

Thanks for helping improve Plan-AI. This project is local-first planning infrastructure for AI-assisted software work.

## Development setup

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
go mod download
go test ./...
```

## Quality gate

Before opening a pull request, run:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
bash scripts/test-sandbox.sh
bash scripts/test-vps-clean.sh
bash scripts/release-check.sh
```

## Safety rules

- Do not commit `.plan-ai/`, SQLite databases, logs, `.env` files, tokens, or generated binaries.
- Do not write to a real OpenCode config during tests. Use `OPENCODE_CONFIG_DIR`.
- Keep changes additive unless a migration or compatibility note explains the break.
- Prefer deterministic behavior over implicit LLM behavior.

## Pull request checklist

- [ ] Tests pass.
- [ ] Documentation is updated when user-facing behavior changes.
- [ ] New commands are covered in `docs/cli-reference.md`.
- [ ] Runtime data and secrets are not committed.
- [ ] The change does not mutate real user configuration during validation.

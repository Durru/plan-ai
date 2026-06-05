# Troubleshooting

## `plan-ai: command not found`

The install prefix is not on PATH.

For non-root installs:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

For custom prefixes:

```bash
export PATH="$PLAN_AI_INSTALL_PREFIX/bin:$PATH"
```

## `project store not initialized`

Run from the project directory:

```bash
plan-ai init
```

## `global store not initialized`

Run:

```bash
plan-ai install
```

or re-run:

```bash
bash scripts/install.sh
```

## OpenCode setup refuses to write

This is intentional. By default `plan-ai setup opencode` requires `OPENCODE_CONFIG_DIR` so tests do not mutate real OpenCode config.

Safe sandbox example:

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai setup opencode
```

Only opt into real OpenCode config when you mean it:

```bash
plan-ai setup opencode --allow-real-opencode
```

## Sandbox leaves files behind

`test-sandbox.sh` and `test-vps-clean.sh` clean by default. If you passed `--keep`, remove the printed temporary directory manually.

The release gate also checks that these real paths are absent:

```bash
test ! -e /root/plan-ai/.plan-ai
test ! -e /root/.plan-ai
```

## Accidentally created runtime data in the repo

Do not commit it. Remove only the project runtime directory if you are sure it is test data:

```bash
rm -rf .plan-ai
```

Never remove a user's real `~/.plan-ai` without confirmation.

## Go build or test fails

Run:

```bash
go mod download
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./...
```

If the failure only appears in sandbox scripts, re-run with `--keep` and inspect the preserved temp directory.

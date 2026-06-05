# Installation

Plan-AI is distributed as a Go CLI. The installer builds local binaries from the repository and places them in a user-selectable prefix.

## Requirements

- Linux or macOS shell environment
- Go 1.22+ as declared by `go.mod`
- Git

Ubuntu packages:

```bash
sudo apt-get update
sudo apt-get install -y git golang-go ca-certificates
```

## Install from source

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
bash scripts/install.sh
```

The installer:

1. Builds `plan-ai` from `./cmd/plan-ai`.
2. Builds `plan-ai-mcp-server` from `./cmd/mcp-server`.
3. Installs both into:
   - `/usr/local/bin` when run as root.
   - `$HOME/.local/bin` when run as a normal user.
   - `$PLAN_AI_INSTALL_PREFIX/bin` when `PLAN_AI_INSTALL_PREFIX` is set.
4. Runs `plan-ai install` to create/migrate the global store.

## Non-root install

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
bash scripts/install.sh
export PATH="$HOME/.local/bin:$PATH"
plan-ai doctor
```

## Root install

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
sudo bash scripts/install.sh
plan-ai doctor
```

Root installation writes binaries to `/usr/local/bin`. It does not delete existing Plan-AI data.

## Custom prefix

```bash
PLAN_AI_INSTALL_PREFIX="$HOME/.plan-ai-bin" bash scripts/install.sh
export PATH="$HOME/.plan-ai-bin/bin:$PATH"
```

## Skip store initialization

For packaging or CI:

```bash
PLAN_AI_SKIP_INIT=1 bash scripts/install.sh
```

## Verify

```bash
plan-ai doctor
plan-ai bootstrap
plan-ai status
```

For real OpenCode integration, run from the project you want Plan-AI to manage:

```bash
plan-ai bootstrap --allow-real-opencode
```

For sandbox validation, keep OpenCode writes isolated:

```bash
OPENCODE_CONFIG_DIR="$PWD/.tmp/opencode-config" plan-ai bootstrap
```

## Uninstall

```bash
bash scripts/uninstall.sh
```

By default uninstall removes installed binaries and preserves `~/.plan-ai`. To delete data, use `--purge-data`; the script asks for confirmation unless `--yes` is also supplied.

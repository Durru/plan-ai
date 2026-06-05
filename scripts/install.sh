#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
  printf 'error: Go is required but was not found on PATH\n' >&2
  exit 1
fi

if [[ -n "${PLAN_AI_INSTALL_PREFIX:-}" ]]; then
  PREFIX="$PLAN_AI_INSTALL_PREFIX"
elif [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  PREFIX="/usr/local"
else
  PREFIX="$HOME/.local"
fi

BIN_DIR="$PREFIX/bin"
mkdir -p "$BIN_DIR"

PLAN_AI_BIN="$BIN_DIR/plan-ai"
MCP_BIN="$BIN_DIR/plan-ai-mcp-server"

printf 'Installing Plan-AI from %s\n' "$ROOT_DIR"
printf 'Install prefix: %s\n' "$PREFIX"

(
  cd "$ROOT_DIR"
  go build -o "$PLAN_AI_BIN" ./cmd/plan-ai
  go build -o "$MCP_BIN" ./cmd/mcp-server
)

chmod 0755 "$PLAN_AI_BIN" "$MCP_BIN"

printf 'Installed: %s\n' "$PLAN_AI_BIN"
printf 'Installed: %s\n' "$MCP_BIN"

if [[ "${PLAN_AI_SKIP_INIT:-0}" != "1" ]]; then
  "$PLAN_AI_BIN" install
  # Gentle-AI installer: create state.json with full preset.
  # To wire OpenCode integration, run: plan-ai sync --allow-real-opencode
  "$PLAN_AI_BIN" install --preset full-plan-ai --bin-dir "$BIN_DIR" 2>/dev/null || true
else
  printf 'Skipped global store initialization because PLAN_AI_SKIP_INIT=1\n'
fi

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    printf 'PATH hint: export PATH="%s:$PATH"\n' "$BIN_DIR"
    ;;
esac

printf 'Plan-AI installation complete.\n'

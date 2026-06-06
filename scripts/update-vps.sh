#!/usr/bin/env bash
set -euo pipefail

# Plan-AI VPS update script
#
# Updates Plan-AI on a remote VPS by SSHing in, pulling the latest
# code, building, and installing.
#
# Usage:
#   VPS_HOST=example.com bash scripts/update-vps.sh
#   VPS_HOST=example.com VPS_USER=deploy VPS_PORT=2222 bash scripts/update-vps.sh
#   VPS_HOST=example.com VPS_KEY=~/.ssh/id_ed25519_vps bash scripts/update-vps.sh
#
# All settings can also be passed as positional args:
#   bash scripts/update-vps.sh <host> [user] [port] [key-path]

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

HOST="${VPS_HOST:-${1:-}}"
USER="${VPS_USER:-${2:-root}}"
PORT="${VPS_PORT:-${3:-22}}"
KEY="${VPS_KEY:-${4:-}}"

if [[ -z "$HOST" ]]; then
  printf 'error: VPS_HOST is required\n' >&2
  printf '\n' >&2
  printf 'Usage:\n' >&2
  printf '  VPS_HOST=example.com bash scripts/update-vps.sh\n' >&2
  printf '  bash scripts/update-vps.sh <host> [user] [port] [key-path]\n' >&2
  exit 1
fi

# Build SSH args
SSH_OPTS="-o StrictHostKeyChecking=accept-new -o LogLevel=ERROR -p $PORT"
if [[ -n "$KEY" ]]; then
  SSH_OPTS="$SSH_OPTS -i $KEY"
fi

printf 'Updating Plan-AI on %s@%s (port %s)...\n' "$USER" "$HOST" "$PORT"

# The remote script runs on the VPS via SSH stdin.
# It clones/pulls, builds, installs, and verifies.
ssh $SSH_OPTS "$USER@$HOST" bash -s << 'REMOTE'
  set -euo pipefail

  PLAN_AI_DIR="${PLAN_AI_DIR:-${HOME}/plan-ai}"
  BIN_DIR="${BIN_DIR:-${HOME}/.local/bin}"

  # Clone or pull
  if [[ -d "$PLAN_AI_DIR/.git" ]]; then
    printf 'Pulling latest in %s...\n' "$PLAN_AI_DIR"
    cd "$PLAN_AI_DIR"
    git pull --ff-only
  else
    printf 'Cloning Plan-AI to %s...\n' "$PLAN_AI_DIR"
    mkdir -p "$PLAN_AI_DIR"
    git clone https://github.com/Durru/plan-ai.git "$PLAN_AI_DIR"
    cd "$PLAN_AI_DIR"
  fi

  # Build
  printf 'Building...\n'
  go build -o "$BIN_DIR/plan-ai" ./cmd/plan-ai
  chmod 0755 "$BIN_DIR/plan-ai"

  # Verify
  INSTALLED="$("$BIN_DIR/plan-ai" --version 2>/dev/null || "$BIN_DIR/plan-ai" version 2>/dev/null || true)"
  printf 'Installed: %s/plan-ai\n' "$BIN_DIR"
  printf 'Version: %s\n' "${INSTALLED:-plan-ai}"

  # Sync install state (preserves existing config)
  "$BIN_DIR/plan-ai" install 2>/dev/null || true
  printf 'UPDATE_VPS_OK\n'
REMOTE

printf 'Plan-AI update on %s complete.\n' "$HOST"

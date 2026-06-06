#!/usr/bin/env bash
set -euo pipefail

YES=0
PURGE_DATA=0

for arg in "$@"; do
  case "$arg" in
    --yes|-y) YES=1 ;;
    --purge-data) PURGE_DATA=1 ;;
    --help|-h)
      cat <<'USAGE'
Usage: bash scripts/uninstall.sh [--yes] [--purge-data]

Removes installed Plan-AI binaries from the install prefix.
Data under ~/.plan-ai is preserved unless --purge-data is supplied.
USAGE
      exit 0
      ;;
    *)
      printf 'error: unknown argument: %s\n' "$arg" >&2
      exit 1
      ;;
  esac
done

if [[ -n "${PLAN_AI_INSTALL_PREFIX:-}" ]]; then
  PREFIX="$PLAN_AI_INSTALL_PREFIX"
elif [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  PREFIX="/usr/local"
else
  PREFIX="$HOME/.local"
fi

BIN_DIR="$PREFIX/bin"
DATA_DIR="${PLAN_AI_DATA_DIR:-${HOME}/.plan-ai}"

# Try Gentle-AI uninstaller first (handles opencode config cleanup)
if command -v "$BIN_DIR/plan-ai" >/dev/null 2>&1; then
  "$BIN_DIR/plan-ai" uninstall 2>/dev/null || true
fi

for bin in "$BIN_DIR/plan-ai"; do
  if [[ -e "$bin" ]]; then
    rm -f "$bin"
    printf 'Removed: %s\n' "$bin"
  else
    printf 'Not found: %s\n' "$bin"
  fi
done

if [[ "$PURGE_DATA" -eq 1 ]]; then
  if [[ ! -e "$DATA_DIR" ]]; then
    printf 'Data directory not found: %s\n' "$DATA_DIR"
  elif [[ "$YES" -eq 1 ]]; then
    rm -rf "$DATA_DIR"
    printf 'Removed data directory: %s\n' "$DATA_DIR"
  else
    printf 'Delete Plan-AI data directory %s? Type yes to continue: ' "$DATA_DIR"
    read -r answer
    if [[ "$answer" == "yes" ]]; then
      rm -rf "$DATA_DIR"
      printf 'Removed data directory: %s\n' "$DATA_DIR"
    else
      printf 'Preserved data directory: %s\n' "$DATA_DIR"
    fi
  fi
else
  printf 'Preserved data directory: %s\n' "$DATA_DIR"
fi

printf 'Plan-AI uninstall complete.\n'

#!/usr/bin/env bash
# Plan-AI legacy install — builds from local source. For end-users, use:
#   curl -fsSL https://raw.githubusercontent.com/Durru/plan-ai/main/scripts/install.sh | bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PREFIX="${PLAN_AI_INSTALL_PREFIX:-/usr/local}"

echo "Installing Plan-AI from $PROJECT_ROOT"
echo "Install prefix: $PREFIX"

cd "$PROJECT_ROOT"
go build -o plan-ai ./cmd/plan-ai/ || { echo "Build failed"; exit 1; }

mkdir -p "$PREFIX/bin"
cp plan-ai "$PREFIX/bin/plan-ai"
chmod +x "$PREFIX/bin/plan-ai"
echo "Installed: $PREFIX/bin/plan-ai"

if [[ ":$PATH:" != *":$PREFIX/bin:"* ]]; then
    echo "Add to PATH: export PATH=\"\$PATH:$PREFIX/bin\""
fi

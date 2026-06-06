#!/usr/bin/env bash
# Test OpenCode MCP minimal mode.
# Creates a temp HOME, generates opencode.json, runs framed MCP initialize and
# tools/list with PLAN_AI_MCP_MINIMAL=true, verifies only minimal tools are
# exposed, and calls plan_ai.project_status.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

SANDBOX="$(mktemp -d)"
KEEP=0

if [[ "${1:-}" == "--keep" ]]; then
  KEEP=1
fi

cleanup() {
  if [[ "$KEEP" -eq 0 ]]; then
    rm -rf "$SANDBOX"
    printf 'SANDBOX_CLEANED\n'
  else
    printf 'Sandbox kept at %s\n' "$SANDBOX"
  fi
}
trap cleanup EXIT

export HOME="$SANDBOX/home"
export PLAN_AI_MCP_MINIMAL="true"

mkdir -p "$HOME"

printf '==> Generating opencode.json with minimal mode...\n'

# Simulate what bootstrap does: write a minimal opencode.json with minimal mode
mkdir -p "$HOME/.config/opencode"
cat > "$HOME/.config/opencode/opencode.json" <<'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "plan-ai": {
      "type": "local",
      "command": ["plan-ai", "mcp", "serve"],
      "enabled": true,
      "env": {
        "PLAN_AI_PROJECT_ROOT": "/tmp",
        "PLAN_AI_MCP_MINIMAL": "true"
      }
    }
  }
}
EOF

printf '  opencode.json written\n'

# Build the plan-ai CLI
printf '==> Building plan-ai...\n'
go build -o "$SANDBOX/plan-ai" ./cmd/plan-ai
printf '  plan-ai built\n'

# Helper: encode an MCP JSON-RPC message as a framed message.
encode_mcp_message() {
  local msg="$1"
  local len="${#msg}"
  printf 'Content-Length: %d\r\n\r\n%s' "$len" "$msg"
}

# Run framed initialize via stdio
printf '==> Running initialize...\n'
INIT_MSG=$(encode_mcp_message '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}')
INIT_OUTPUT=$(printf '%s' "$INIT_MSG" | PLAN_AI_MCP_MINIMAL=true "$SANDBOX/plan-ai" mcp serve 2>&1)
printf '%s\n' "$INIT_OUTPUT"

if [[ "$INIT_OUTPUT" != *'"protocolVersion"'* ]]; then
  printf 'FAIL: initialize response missing protocolVersion\n' >&2
  exit 1
fi
printf '  initialize: OK\n'

# Run framed tools/list
printf '==> Running tools/list...\n'
LIST_MSG=$(encode_mcp_message '{"jsonrpc":"2.0","id":2,"method":"tools/list"}')
TOOLS_OUTPUT=$(printf '%s' "$LIST_MSG" | PLAN_AI_MCP_MINIMAL=true "$SANDBOX/plan-ai" mcp serve 2>&1)
printf '%s\n' "$TOOLS_OUTPUT"

# Check that only minimal tools are present
if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.project_status"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.project_status\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.discover_intent"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.discover_intent\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.create_product_intent"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.create_product_intent\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.list_product_intents"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.list_product_intents\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.get_context"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.get_context\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" != *'"name":"plan_ai.get_next_task"'* ]]; then
  printf 'FAIL: tools/list missing plan_ai.get_next_task\n' >&2
  exit 1
fi

# Verify non-minimal tools are NOT present
if [[ "$TOOLS_OUTPUT" == *'"name":"plan_ai.init_project"'* ]]; then
  printf 'FAIL: tools/list exposed non-minimal tool plan_ai.init_project\n' >&2
  exit 1
fi

if [[ "$TOOLS_OUTPUT" == *'"name":"plan_ai.agent_message"'* ]]; then
  printf 'FAIL: tools/list exposed non-minimal tool plan_ai.agent_message\n' >&2
  exit 1
fi

# Count the number of tools
TOOL_COUNT=$(printf '%s\n' "$TOOLS_OUTPUT" | grep -o '"name":"plan_ai\.' | wc -l)
if [[ "$TOOL_COUNT" -ne 6 ]]; then
  printf 'FAIL: expected 6 minimal tools, got %d\n' "$TOOL_COUNT" >&2
  exit 1
fi
printf '  tools/list: OK (6 minimal tools)\n'

# Run framed tools/call plan_ai.project_status
printf '==> Running tools/call plan_ai.project_status...\n'
CALL_MSG=$(encode_mcp_message '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"plan_ai.project_status","arguments":{}}}')
CALL_OUTPUT=$(printf '%s' "$CALL_MSG" | PLAN_AI_MCP_MINIMAL=true "$SANDBOX/plan-ai" mcp serve 2>&1)
printf '%s\n' "$CALL_OUTPUT"

if [[ "$CALL_OUTPUT" != *'"isError":false'* && "$CALL_OUTPUT" != *'"success":true'* ]]; then
  printf 'FAIL: tools/call plan_ai.project_status did not succeed\n' >&2
  exit 1
fi
printf '  tools/call plan_ai.project_status: OK\n'

printf '\nTEST_OPENCODE_MCP_MINIMAL_OK\n'

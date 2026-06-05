#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SANDBOX="${PLAN_AI_VPS_SANDBOX:-$(mktemp -d /tmp/plan-ai-vps-clean.XXXXXX)}"
KEEP=0

if [[ "${1:-}" == "--keep" ]]; then
  KEEP=1
fi

cleanup() {
  if [[ "$KEEP" -eq 0 ]]; then
    if [[ -e "$SANDBOX" ]]; then
      chmod -R u+w "$SANDBOX" 2>/dev/null || true
      rm -rf "$SANDBOX"
    fi
    printf 'VPS_SANDBOX_CLEANED\n'
  else
    printf 'VPS sandbox kept at %s\n' "$SANDBOX"
  fi
}
trap cleanup EXIT

mkdir -p "$SANDBOX/home" "$SANDBOX/project" "$SANDBOX/prefix" "$SANDBOX/opencode-config"

export HOME="$SANDBOX/home"
export PLAN_AI_HOME="$SANDBOX/home"
export PLAN_AI_PROJECT_ROOT="$SANDBOX/project"
export PLAN_AI_INSTALL_PREFIX="$SANDBOX/prefix"
export OPENCODE_CONFIG_DIR="$SANDBOX/opencode-config"
export PATH="$PLAN_AI_INSTALL_PREFIX/bin:$PATH"

(
  cd "$ROOT_DIR"
  bash scripts/install.sh
)

command -v plan-ai >/dev/null
plan-ai doctor || true
plan-ai init
STATUS_OUTPUT=$(plan-ai status)
printf '%s\n' "$STATUS_OUTPUT"

DISCOVERY_OUTPUT=$(plan-ai intent discover "Quiero crear un CRM para talleres mecánicos")
printf '%s\n' "$DISCOVERY_OUTPUT"

CREATE_OUTPUT=$(plan-ai intent create \
  --description "CRM for mechanic workshops" \
  --expected-outcome "Workshop staff can manage customers, vehicles, repair orders, reminders, and status updates" \
  --desired-experience "Fast daily workflow for non-technical Spanish-speaking users" \
  --desired-result "A validated plan before implementation" \
  --success-definition "A mechanic can find a vehicle, see active work, and notify the customer quickly" \
  --failure-definition "The product becomes a generic CRM that ignores workshop operations")
printf '%s\n' "$CREATE_OUTPUT"
INTENT_ID=$(printf '%s\n' "$CREATE_OUTPUT" | awk '/Product Intent created:/ {print $4; exit}')

if [[ -z "$INTENT_ID" ]]; then
  printf 'failed to capture Product Intent id\n' >&2
  exit 1
fi

plan-ai intent list
plan-ai intent submit "$INTENT_ID"
plan-ai intent approve "$INTENT_ID"
plan-ai intent show "$INTENT_ID"
plan-ai discovery init --intent "$INTENT_ID"
plan-ai discovery next --intent "$INTENT_ID"
plan-ai ambiguity analyze --intent "$INTENT_ID"
plan-ai confidence evaluate --intent "$INTENT_ID"
plan-ai alignment context --intent "$INTENT_ID"
plan-ai alignment review --intent "$INTENT_ID" --outcome "Workshop CRM MVP" --plan "Build customers, vehicles, repair orders, and status tracking" --task "Create customer and vehicle schema"
plan-ai alignment framework --intent "$INTENT_ID"
plan-ai setup opencode
DOCTOR_AFTER_SETUP_OUTPUT=$(plan-ai doctor)
printf '%s\n' "$DOCTOR_AFTER_SETUP_OUTPUT"

if [[ "$DOCTOR_AFTER_SETUP_OUTPUT" != *'Agent: plan-ai'* ]]; then
  printf 'doctor did not detect the generated Plan-AI OpenCode agent\n' >&2
  exit 1
fi

python3 - <<'PY'
import json, os, pathlib, subprocess, sys

registry = json.loads(pathlib.Path(os.environ["OPENCODE_CONFIG_DIR"], "mcp-registry.json").read_text())
if registry.get("command", [None])[0] != "plan-ai-mcp-server":
    raise SystemExit(f"unexpected MCP command: {registry.get('command')!r}")
names = {tool.get("name") for tool in registry.get("tools", [])}
for expected in ["plan_ai.project_status", "plan_ai.agent_process", "plan_ai.create_product_intent"]:
    if expected not in names:
        raise SystemExit(f"missing MCP tool {expected}")

payload = json.dumps({"jsonrpc":"2.0","id":1,"method":"tools/list"}).encode()
wire = b"Content-Length: " + str(len(payload)).encode() + b"\r\n\r\n" + payload
proc = subprocess.run(["plan-ai-mcp-server"], input=wire, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)
raw = proc.stdout.decode()
header, body = raw.split("\r\n\r\n", 1)
response = json.loads(body)
tool_names = {tool["name"] for tool in response["result"]["tools"]}
if "plan_ai.project_status" not in tool_names:
    raise SystemExit("stdio MCP tools/list did not expose plan_ai.project_status")
PY

test -x "$PLAN_AI_INSTALL_PREFIX/bin/plan-ai"
test -x "$PLAN_AI_INSTALL_PREFIX/bin/plan-ai-mcp-server"
test -f "$HOME/.plan-ai/global.db"
test -f "$PLAN_AI_PROJECT_ROOT/.plan-ai/project.db"
test -f "$OPENCODE_CONFIG_DIR/opencode.json"

(
  cd "$ROOT_DIR"
  bash scripts/uninstall.sh --yes
)

test ! -e "$PLAN_AI_INSTALL_PREFIX/bin/plan-ai"
test ! -e "$PLAN_AI_INSTALL_PREFIX/bin/plan-ai-mcp-server"
test -f "$HOME/.plan-ai/global.db"

test ! -e "$ROOT_DIR/.plan-ai"
test ! -e "/root/.plan-ai"

printf 'VPS_CLEAN_VALIDATION_OK\n'

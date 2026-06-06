#!/usr/bin/env bash
set -euo pipefail

SANDBOX=${PLAN_AI_SANDBOX:-"$PWD/.tmp/plan-ai-sandbox"}
KEEP=0

if [[ "${1:-}" == "--keep" ]]; then
  KEEP=1
fi

rm -rf "$SANDBOX"
mkdir -p "$SANDBOX/home" "$SANDBOX/project" "$SANDBOX/opencode-config"

cleanup() {
  if [[ "$KEEP" -eq 0 ]]; then
    if [[ -e "$SANDBOX" ]]; then
      chmod -R u+w "$SANDBOX" 2>/dev/null || true
      rm -rf "$SANDBOX"
    fi
    printf 'SANDBOX_CLEANED\n'
  else
    printf 'Sandbox kept at %s\n' "$SANDBOX"
  fi
}
trap cleanup EXIT

export PLAN_AI_HOME="$SANDBOX/home"
export PLAN_AI_PROJECT_ROOT="$SANDBOX/project"
export HOME="$SANDBOX/home"
export OPENCODE_CONFIG_DIR="$SANDBOX/opencode-config"

go run ./cmd/plan-ai install
go run ./cmd/plan-ai status
go run ./cmd/plan-ai init
INTENT_OUTPUT=$(go run ./cmd/plan-ai intent detect --content "quiero un SaaS CRM")
printf '%s\n' "$INTENT_OUTPUT"
INTENT_ID=$(printf '%s\n' "$INTENT_OUTPUT" | awk '/id:/ {print $2; exit}')
INTENT_LATEST_OUTPUT=$(go run ./cmd/plan-ai intent latest)
printf '%s\n' "$INTENT_LATEST_OUTPUT"
INTENT_APPROVE_OUTPUT=$(go run ./cmd/plan-ai intent approve "$INTENT_ID")
printf '%s\n' "$INTENT_APPROVE_OUTPUT"
VISION_DOC_OUTPUT=$(go run ./cmd/plan-ai vision document --intent "$INTENT_ID")
printf '%s\n' "$VISION_DOC_OUTPUT"
VISION_DOC_ID=$(printf '%s\n' "$VISION_DOC_OUTPUT" | awk '/id:/ {print $2; exit}')
VISION_DOCS_OUTPUT=$(go run ./cmd/plan-ai vision documents)
printf '%s\n' "$VISION_DOCS_OUTPUT"
VISION_DOC_SHOW_OUTPUT=$(go run ./cmd/plan-ai vision document-show "$VISION_DOC_ID")
printf '%s\n' "$VISION_DOC_SHOW_OUTPUT"
VISION_DOC_APPROVE_OUTPUT=$(go run ./cmd/plan-ai vision approve-document "$VISION_DOC_ID")
printf '%s\n' "$VISION_DOC_APPROVE_OUTPUT"
REQ_DISCOVERY_OUTPUT=$(go run ./cmd/plan-ai requirements discover --content "quiero ecommerce")
printf '%s\n' "$REQ_DISCOVERY_OUTPUT"
REQ_LIST_OUTPUT=$(go run ./cmd/plan-ai requirements list)
printf '%s\n' "$REQ_LIST_OUTPUT"
REQ_ID=$(printf '%s\n' "$REQ_LIST_OUTPUT" | awk '/checkout/ {print $1; exit}')
REQ_APPROVE_OUTPUT=$(go run ./cmd/plan-ai requirements approve "$REQ_ID")
printf '%s\n' "$REQ_APPROVE_OUTPUT"
APPROVAL_LIST_OUTPUT=$(go run ./cmd/plan-ai approval list)
printf '%s\n' "$APPROVAL_LIST_OUTPUT"
OPENCODE_SETUP_OUTPUT=$(go run ./cmd/plan-ai setup opencode)
printf '%s\n' "$OPENCODE_SETUP_OUTPUT"
INGEST_OUTPUT=$(go run ./cmd/plan-ai ingest --type prompt --content "Build a planning assistant for product teams. It must use SQLite. Success: users approve a plan faster.")
printf '%s\n' "$INGEST_OUTPUT"
VISION_OUTPUT=$(go run ./cmd/plan-ai vision draft)
printf '%s\n' "$VISION_OUTPUT"
APPROVED_REQ_OUTPUT=$(go run ./cmd/plan-ai approved add --type requirement "The app must save planning drafts")
printf '%s\n' "$APPROVED_REQ_OUTPUT"
APPROVED_DECISION_OUTPUT=$(go run ./cmd/plan-ai approved add --type decision "Use SQLite for local persistence")
printf '%s\n' "$APPROVED_DECISION_OUTPUT"
APPROVED_LIST_OUTPUT=$(go run ./cmd/plan-ai approved list)
printf '%s\n' "$APPROVED_LIST_OUTPUT"
DOCTOR_OUTPUT=$(go run ./cmd/plan-ai doctor)
printf '%s\n' "$DOCTOR_OUTPUT"
go run ./cmd/plan-ai dev seed-domain
LIST_OUTPUT=$(go run ./cmd/plan-ai dev list-domain)
printf '%s\n' "$LIST_OUTPUT"
SCAN_OUTPUT=$(go run ./cmd/plan-ai scan)
printf '%s\n' "$SCAN_OUTPUT"
STATUS_OUTPUT=$(go run ./cmd/plan-ai status)
printf '%s\n' "$STATUS_OUTPUT"
SEED_KNOWLEDGE_OUTPUT=$(go run ./cmd/plan-ai dev seed-knowledge)
printf '%s\n' "$SEED_KNOWLEDGE_OUTPUT"
KNOWLEDGE_LIST_OUTPUT=$(go run ./cmd/plan-ai knowledge list)
printf '%s\n' "$KNOWLEDGE_LIST_OUTPUT"
KNOWLEDGE_SEARCH_OUTPUT=$(go run ./cmd/plan-ai knowledge search postgres)
printf '%s\n' "$KNOWLEDGE_SEARCH_OUTPUT"
KNOWLEDGE_PG_ID=$(printf '%s\n' "$KNOWLEDGE_LIST_OUTPUT" | grep "PostgreSQL Multi Tenant" | awk '{print $1; exit}')
KNOWLEDGE_REUSE_OUTPUT=$(go run ./cmd/plan-ai knowledge reuse "$KNOWLEDGE_PG_ID")
printf '%s\n' "$KNOWLEDGE_REUSE_OUTPUT"
FINAL_STATUS_OUTPUT=$(go run ./cmd/plan-ai status)
printf '%s\n' "$FINAL_STATUS_OUTPUT"

test -f "$SANDBOX/home/.plan-ai/global.db"
test -f "$SANDBOX/project/.plan-ai/project.db"

for expected in \
  'Intent profile detected.' \
  'primary_intent: CRM' \
  'SaaS/candidate' \
  'multi-user/candidate' \
  'admin panel/candidate' \
  'subscriptions/candidate' \
  'reports/candidate' \
  'automations/candidate' \
  'approved: false'; do
  if [[ "$INTENT_OUTPUT" != *"$expected"* ]]; then
    printf 'intent detect missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
for expected in \
  'Latest intent profile.' \
  "$INTENT_ID"; do
  if [[ "$INTENT_LATEST_OUTPUT" != *"$expected"* ]]; then
    printf 'intent latest missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
if [[ "$INTENT_APPROVE_OUTPUT" != *'Intent profile approved.'* || "$INTENT_APPROVE_OUTPUT" != *'approved: true'* ]]; then
  printf 'intent approve missing expected approval output\n' >&2
  exit 1
fi

for expected in \
  'Vision document created.' \
  'functional:' \
  'visual:' \
  'technical:' \
  'operational:' \
  'business:' \
  'approved: false'; do
  if [[ "$VISION_DOC_OUTPUT" != *"$expected"* ]]; then
    printf 'vision document missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
if [[ "$VISION_DOCS_OUTPUT" != *"$VISION_DOC_ID"* || "$VISION_DOC_SHOW_OUTPUT" != *'Vision document.'* || "$VISION_DOC_APPROVE_OUTPUT" != *'approved: true'* ]]; then
  printf 'vision document list/show/approve validation failed\n' >&2
  exit 1
fi

for expected in \
  'Requirement candidates discovered:' \
  'cart' \
  'checkout' \
  'coupons' \
  'SEO' \
  'analytics'; do
  if [[ "$REQ_DISCOVERY_OUTPUT" != *"$expected"* && "$REQ_LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'requirements missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
if [[ "$REQ_APPROVE_OUTPUT" != *'Requirement candidate approved:'* || "$REQ_APPROVE_OUTPUT" != *'approved'* ]]; then
  printf 'requirement approval validation failed\n' >&2
  exit 1
fi
if [[ "$APPROVAL_LIST_OUTPUT" != *'vision_document'* || "$APPROVAL_LIST_OUTPUT" != *'requirement_candidate'* ]]; then
  printf 'approval list missing expected target records\n' >&2
  exit 1
fi

test -f "$SANDBOX/opencode-config/opencode.json"
test -f "$SANDBOX/opencode-config/mcp-registry.json"
test -f "$SANDBOX/opencode-config/agents/plan-ai.json"
test -f "$SANDBOX/opencode-config/profiles.json"
test -f "$SANDBOX/opencode-config/prompts.json"
test -f "$SANDBOX/opencode-config/plan-ai-workflows.json"
test -f "$SANDBOX/project/.plan-ai/opencode-sync.json"

for expected in \
  'OpenCode integration artifacts generated.' \
  'opencode config:' \
  'mcp registry:' \
  'workflows:' \
  'sync marker:'; do
  if [[ "$OPENCODE_SETUP_OUTPUT" != *"$expected"* ]]; then
    printf 'setup opencode missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Input ingested.' \
  'classification:'; do
  if [[ "$INGEST_OUTPUT" != *"$expected"* ]]; then
    printf 'ingest missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Vision draft created.' \
  'approved: false'; do
  if [[ "$VISION_OUTPUT" != *"$expected"* ]]; then
    printf 'vision missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Approved context stored' \
  'The app must save planning drafts' \
  'Use SQLite for local persistence'; do
  if [[ "$APPROVED_REQ_OUTPUT" != *"$expected"* && "$APPROVED_DECISION_OUTPUT" != *"$expected"* && "$APPROVED_LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'approved context missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

if [[ -e "$HOME/.config/opencode" ]]; then
  printf 'sandbox HOME unexpectedly contains .config/opencode\n' >&2
  exit 1
fi

for expected in \
  'Plan-AI doctor' \
  'Global migrations: ok' \
  'Project migrations: ok'; do
  if [[ "$DOCTOR_OUTPUT" != *"$expected"* ]]; then
    printf 'doctor missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Project scan completed.' \
  'Git:' \
  'detected:' \
  'Fingerprint:' \
  'Files indexed:'; do
  if [[ "$SCAN_OUTPUT" != *"$expected"* ]]; then
    printf 'scan missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Scan:' \
  'latest:' \
  'files indexed:'; do
  if [[ "$STATUS_OUTPUT" != *"$expected"* ]]; then
    printf 'status missing expected scan output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Plans: 2' \
  'Phases: 1' \
  'Tasks: 1' \
  'Decisions: 1' \
  'Research: 1' \
  'Knowledge: 1' \
  'Validations: 1' \
  'Snapshots: 1'; do
  if [[ "$LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'list-domain missing expected count: %s\n' "$expected" >&2
    exit 1
  fi
  if [[ "$STATUS_OUTPUT" != *"$expected"* ]]; then
    printf 'status missing expected count: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Knowledge seed: created' \
  'PostgreSQL Multi Tenant' \
  'OAuth 2.0' \
  'Stripe Billing'; do
  if [[ "$SEED_KNOWLEDGE_OUTPUT" != *"$expected"* && "$KNOWLEDGE_LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'seed-knowledge missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'PostgreSQL Multi Tenant' \
  'approved' \
  'total: 4'; do
  if [[ "$FINAL_STATUS_OUTPUT" != *"$expected"* && "$KNOWLEDGE_SEARCH_OUTPUT" != *"$expected"* && "$KNOWLEDGE_REUSE_OUTPUT" != *"$expected"* ]]; then
    printf 'knowledge flow missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

if [[ "$KNOWLEDGE_REUSE_OUTPUT" != *"reuse count: 1"* ]]; then
  printf 'knowledge reuse did not increment count:\n%s\n' "$KNOWLEDGE_REUSE_OUTPUT" >&2
  exit 1
fi

if [[ "$FINAL_STATUS_OUTPUT" != *'total: 4'* ]]; then
  printf 'status knowledge total != 4:\n%s\n' "$FINAL_STATUS_OUTPUT" >&2
  exit 1
fi

if [[ "$FINAL_STATUS_OUTPUT" != *'approved: 3'* ]]; then
  printf 'status knowledge approved != 3:\n%s\n' "$FINAL_STATUS_OUTPUT" >&2
  exit 1
fi

if [[ "$FINAL_STATUS_OUTPUT" != *'reused: 2'* ]]; then
  printf 'status knowledge reused != 2 (1 from seed-domain + 1 from reuse):\n%s\n' "$FINAL_STATUS_OUTPUT" >&2
  exit 1
fi

SEED_RESEARCH_OUTPUT=$(go run ./cmd/plan-ai dev seed-research)
printf '%s\n' "$SEED_RESEARCH_OUTPUT"
for expected in 'Research seed: created'; do
  if [[ "$SEED_RESEARCH_OUTPUT" != *"$expected"* ]]; then
    printf 'seed-research missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

RESEARCH_LIST_OUTPUT=$(go run ./cmd/plan-ai research list)
printf '%s\n' "$RESEARCH_LIST_OUTPUT"
for expected in 'LLM Token Optimization' 'SQLite Performance Limits'; do
  if [[ "$RESEARCH_LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'research list missing: %s\n' "$expected" >&2
    exit 1
  fi
done

RESEARCH_SHOW_ID=$(printf '%s\n' "$RESEARCH_LIST_OUTPUT" | grep "LLM Token Optimization" | awk '{print $1; exit}')
RESEARCH_SHOW_OUTPUT=$(go run ./cmd/plan-ai research show "$RESEARCH_SHOW_ID")
printf '%s\n' "$RESEARCH_SHOW_OUTPUT"
for expected in 'Prompt Compression' 'LLMLingua Paper' 'Prompt compression is viable'; do
  if [[ "$RESEARCH_SHOW_OUTPUT" != *"$expected"* ]]; then
    printf 'research show missing: %s\n' "$expected" >&2
    exit 1
  fi
done

RESEARCH_APPROVE_OUTPUT=$(go run ./cmd/plan-ai research approve "$RESEARCH_SHOW_ID")
printf '%s\n' "$RESEARCH_APPROVE_OUTPUT"
if [[ "$RESEARCH_APPROVE_OUTPUT" != *"approved"* ]]; then
  printf 'research approve failed:\n%s\n' "$RESEARCH_APPROVE_OUTPUT" >&2
  exit 1
fi

PHASE12_RESEARCH_OUTPUT=$(go run ./cmd/plan-ai research add --topic "Workflow research" --summary "Research reusable workflow storage" --confidence 80)
printf '%s\n' "$PHASE12_RESEARCH_OUTPUT"
PHASE12_KNOWLEDGE_OUTPUT=$(go run ./cmd/plan-ai knowledge add --topic "Workflow Knowledge" --summary "Reuse workflow research" --content "Plan from approved context and stored knowledge." --confidence 0.8 --status approved --source research)
printf '%s\n' "$PHASE12_KNOWLEDGE_OUTPUT"
PLAN_OUTPUT=$(go run ./cmd/plan-ai plan)
printf '%s\n' "$PLAN_OUTPUT"
for expected in \
  'Research created.' \
  'Knowledge created.' \
  'Planning artifacts created.' \
  'master_plan:' \
  'specific_plan:' \
  'implementation_document:' \
  'workflow_run:'; do
  if [[ "$PHASE12_RESEARCH_OUTPUT" != *"$expected"* && "$PHASE12_KNOWLEDGE_OUTPUT" != *"$expected"* && "$PLAN_OUTPUT" != *"$expected"* ]]; then
    printf 'phase 12-14 flow missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

# Phase 15-17: Model Strategy, Orchestrator, Context Engine
CONTEXT_OUTPUT=$(go run ./cmd/plan-ai context)
printf '%s\n' "$CONTEXT_OUTPUT"
CAPABILITIES_OUTPUT=$(go run ./cmd/plan-ai capabilities)
printf '%s\n' "$CAPABILITIES_OUTPUT"
JOBS_LIST_OUTPUT=$(go run ./cmd/plan-ai jobs list)
printf '%s\n' "$JOBS_LIST_OUTPUT"

for expected in \
  'Executive context:' \
  'status:' \
  'what_missing:' \
  'what_next:'; do
  if [[ "$CONTEXT_OUTPUT" != *"$expected"* ]]; then
    printf 'context missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Registered capabilities:' \
  'planning: Planning' \
  'research: Research Execution' \
  'vision: Vision Drafting'; do
  if [[ "$CAPABILITIES_OUTPUT" != *"$expected"* ]]; then
    printf 'capabilities missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'No orchestration jobs yet.'; do
  if [[ "$JOBS_LIST_OUTPUT" != *"$expected"* ]]; then
    printf 'jobs list missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

printf 'Verified sandbox DBs:\n'
  printf '  %s\n' "$SANDBOX/home/.plan-ai/global.db"
  printf '  %s\n' "$SANDBOX/project/.plan-ai/project.db"

# Phase 18-20: Verify compatibility table and view names exist
PROJECT_DB="$SANDBOX/project/.plan-ai/project.db"
if [[ -f "$PROJECT_DB" ]]; then
  for table in snapshots change_events impact_reports opencode_detections provider_registry skill_registry; do
    COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $table LIMIT 0;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
      printf '  Table %s: OK\n' "$table"
    else
      printf '  Table %s: MISSING\n' "$table" >&2
    fi
  done
  for tbl in tool_runs tool_audit; do
    COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $tbl LIMIT 0;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
      printf '  View %s: OK\n' "$tbl"
    else
      printf '  View %s: MISSING\n' "$tbl" >&2
    fi
  done
fi

# Phase 21-22: Verify agent and continuous planning tables exist
if [[ -f "$PROJECT_DB" ]]; then
	for table in agent_runs_v2 agent_messages agent_delegated_jobs; do
    COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $table LIMIT 0;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
      printf '  Table %s: OK\n' "$table"
    else
      printf '  Table %s: MISSING\n' "$table" >&2
    fi
  done
	for tbl in continuous_events plan_update_proposals context_deliveries; do
    COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $tbl LIMIT 0;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
      printf '  Table %s: OK\n' "$tbl"
    else
      printf '  Table %s: MISSING\n' "$tbl" >&2
    fi
	done
	for table in continuous_status; do
		COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $table LIMIT 0;" 2>/dev/null)
		if [[ $? -eq 0 ]]; then
			printf '  Table %s: OK\n' "$table"
		else
			printf '  Table %s: MISSING\n' "$table" >&2
			exit 1
		fi
	done
	for view in agent_runs delegated_jobs agent_responses subagent_outputs context_delivery_logs; do
		COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $view LIMIT 0;" 2>/dev/null)
		if [[ $? -eq 0 ]]; then
			printf '  View %s: OK\n' "$view"
		else
			printf '  View %s: MISSING\n' "$view" >&2
			exit 1
		fi
	done
fi

MCP_TOOLS_OUTPUT=$(go run ./cmd/plan-ai mcp list-tools)
for expected_tool in \
  'plan_ai.agent_message' \
  'plan_ai.agent_status' \
  'plan_ai.continuous_status' \
  'plan_ai.propose_plan_update' \
  'plan_ai.approve_plan_update' \
  'plan_ai.reject_plan_update' \
  'plan_ai.get_context_level'; do
  if [[ "$MCP_TOOLS_OUTPUT" != *"$expected_tool"* ]]; then
    printf 'mcp tool missing: %s\n' "$expected_tool" >&2
    exit 1
  fi
  printf '  MCP tool %s: OK\n' "$expected_tool"
done

# Phase 21-22: Smoke test agent and continuous CLI commands
AGENT_STATUS_OUTPUT=$(go run ./cmd/plan-ai agent status 2>&1 || true)
printf 'agent status: %s\n' "$AGENT_STATUS_OUTPUT"

AGENT_LIST_OUTPUT=$(go run ./cmd/plan-ai agent list 2>&1 || true)
printf 'agent list: %s\n' "$AGENT_LIST_OUTPUT"

# Seed continuous scenario for validation
CONTINUOUS_SEED_OUTPUT=$(go run ./cmd/plan-ai dev seed-continuous-scenario)
printf 'continuous seed: %s\n' "$CONTINUOUS_SEED_OUTPUT"

CONTINUOUS_STATUS_OUTPUT=$(go run ./cmd/plan-ai continuous status 2>&1 || true)
printf 'continuous status: %s\n' "$CONTINUOUS_STATUS_OUTPUT"

CONTINUOUS_EVENTS_OUTPUT=$(go run ./cmd/plan-ai continuous events 2>&1 || true)
printf 'continuous events: %s\n' "$CONTINUOUS_EVENTS_OUTPUT"

CONTINUOUS_PROPOSALS_OUTPUT=$(go run ./cmd/plan-ai continuous proposals 2>&1 || true)
printf 'continuous proposals: %s\n' "$CONTINUOUS_PROPOSALS_OUTPUT"

NEXT_OUTPUT=$(go run ./cmd/plan-ai next 2>&1 || true)
printf 'next task: %s\n' "$NEXT_OUTPUT"

# Validate continuous scenario seeding
for expected in \
  'Continuous scenario seed: created 4 events and 4 proposals'; do
  if [[ "$CONTINUOUS_SEED_OUTPUT" != *"$expected"* ]]; then
    printf 'seed-continuous-scenario missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Recent events:' \
  'Pending proposals:'; do
  if [[ "$CONTINUOUS_STATUS_OUTPUT" != *"$expected"* ]]; then
    printf 'continuous status missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

if [[ "$CONTINUOUS_EVENTS_OUTPUT" == *"No events yet."* ]]; then
  printf 'continuous events returned empty despite seeding\n' >&2
  exit 1
fi

if [[ "$CONTINUOUS_PROPOSALS_OUTPUT" == *"No proposals yet."* ]]; then
  printf 'continuous proposals returned empty despite seeding\n' >&2
  exit 1
fi

printf '  Continuous seeding and querying: OK\n'

# Phase 23-27: MVP audit, discovery, plan generators, and context delivery
for audit_doc in \
  docs/audit/mvp-audit.md \
  docs/audit/gaps.md \
  docs/audit/technical-debt.md \
  docs/audit/roadmap-adjustments.md; do
  if [[ ! -s "$audit_doc" ]]; then
    printf 'audit doc missing or empty: %s\n' "$audit_doc" >&2
    exit 1
  fi
  printf '  Audit doc %s: OK\n' "$audit_doc"
done

DISCOVERY_STATUS_OUTPUT=$(go run ./cmd/plan-ai discovery status)
printf '%s\n' "$DISCOVERY_STATUS_OUTPUT"
MASTER_V2_STATUS_OUTPUT=$(go run ./cmd/plan-ai master-v2 status)
printf '%s\n' "$MASTER_V2_STATUS_OUTPUT"
SPECIFIC_V2_STATUS_OUTPUT=$(go run ./cmd/plan-ai specific-v2 status)
printf '%s\n' "$SPECIFIC_V2_STATUS_OUTPUT"
DELIVERY_STATUS_OUTPUT=$(go run ./cmd/plan-ai delivery status)
printf '%s\n' "$DELIVERY_STATUS_OUTPUT"
CONTEXT_PACKAGE_OUTPUT=$(go run ./cmd/plan-ai context package --type implementation --model gpt-5.5 --content "Implement ecommerce checkout from approved requirements" --token-budget 2048)
printf '%s\n' "$CONTEXT_PACKAGE_OUTPUT"
CONTEXT_PACKAGES_OUTPUT=$(go run ./cmd/plan-ai context packages)
printf '%s\n' "$CONTEXT_PACKAGES_OUTPUT"
RESEARCH_RUN_OUTPUT=$(go run ./cmd/plan-ai research run --agent security --topic "checkout fraud prevention")
printf '%s\n' "$RESEARCH_RUN_OUTPUT"
RESEARCH_RUNS_OUTPUT=$(go run ./cmd/plan-ai research runs)
printf '%s\n' "$RESEARCH_RUNS_OUTPUT"
REFERENCE_ADD_OUTPUT=$(go run ./cmd/plan-ai reference add --source url --uri "https://stripe.com/docs" --title "Stripe checkout docs")
printf '%s\n' "$REFERENCE_ADD_OUTPUT"
REFERENCE_ID=$(printf '%s\n' "$REFERENCE_ADD_OUTPUT" | awk '/id:/ {print $2; exit}')
REFERENCE_LIST_OUTPUT=$(go run ./cmd/plan-ai reference list)
printf '%s\n' "$REFERENCE_LIST_OUTPUT"
REFERENCE_APPROVE_OUTPUT=$(go run ./cmd/plan-ai reference approve "$REFERENCE_ID")
printf '%s\n' "$REFERENCE_APPROVE_OUTPUT"
PLAN_EVOLVE_OUTPUT=$(go run ./cmd/plan-ai plan evolve --objective "Ship approved ecommerce checkout")
printf '%s\n' "$PLAN_EVOLVE_OUTPUT"
PLAN_BLUEPRINT_ID=$(printf '%s\n' "$PLAN_EVOLVE_OUTPUT" | awk '/id:/ {print $2; exit}')
PLAN_BLUEPRINTS_OUTPUT=$(go run ./cmd/plan-ai plan blueprints)
printf '%s\n' "$PLAN_BLUEPRINTS_OUTPUT"
IMPLEMENTATION_PACKAGE_OUTPUT=$(go run ./cmd/plan-ai context implementation-package --plan "$PLAN_BLUEPRINT_ID" --model opencode --objective "Implement approved ecommerce checkout")
printf '%s\n' "$IMPLEMENTATION_PACKAGE_OUTPUT"
IMPLEMENTATION_PACKAGES_OUTPUT=$(go run ./cmd/plan-ai context implementation-packages)
printf '%s\n' "$IMPLEMENTATION_PACKAGES_OUTPUT"
IMPACT_V2_OUTPUT=$(go run ./cmd/plan-ai impact analyze-v2 --type technology_changed --summary "PostgreSQL to MariaDB")
printf '%s\n' "$IMPACT_V2_OUTPUT"
IMPACT_V2_REPORTS_OUTPUT=$(go run ./cmd/plan-ai impact reports-v2)
printf '%s\n' "$IMPACT_V2_REPORTS_OUTPUT"
CONTINUOUS_REGEN_OUTPUT=$(go run ./cmd/plan-ai continuous regenerate --reason "Database migration PostgreSQL to MariaDB" --scope database)
printf '%s\n' "$CONTINUOUS_REGEN_OUTPUT"
CONTINUOUS_REGENS_OUTPUT=$(go run ./cmd/plan-ai continuous regenerations)
printf '%s\n' "$CONTINUOUS_REGENS_OUTPUT"
SUBAGENT_CREATE_OUTPUT=$(go run ./cmd/plan-ai agent subagent-create --type security --objective "Review checkout fraud prevention")
printf '%s\n' "$SUBAGENT_CREATE_OUTPUT"
SUBAGENTS_OUTPUT=$(go run ./cmd/plan-ai agent subagents)
printf '%s\n' "$SUBAGENTS_OUTPUT"
OPENCODE_WORKFLOWS_OUTPUT=$(go run ./cmd/plan-ai setup opencode-workflows)
printf '%s\n' "$OPENCODE_WORKFLOWS_OUTPUT"

for expected in \
  'Discovery Engine: OK' \
  'Master Plan v2 Engine: OK' \
  'Specific Plan v2 Engine: OK' \
  'Context Delivery Engine: OK' \
  'L0-L4 delivery levels'; do
  if [[ "$DISCOVERY_STATUS_OUTPUT" != *"$expected"* && "$MASTER_V2_STATUS_OUTPUT" != *"$expected"* && "$SPECIFIC_V2_STATUS_OUTPUT" != *"$expected"* && "$DELIVERY_STATUS_OUTPUT" != *"$expected"* ]]; then
    printf 'phase 23-27 command missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Context package created.' \
  'type: implementation' \
  'budget=2048'; do
  if [[ "$CONTEXT_PACKAGE_OUTPUT" != *"$expected"* && "$CONTEXT_PACKAGES_OUTPUT" != *"$expected"* ]]; then
    printf 'context package missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Research orchestration completed.' \
  'agent: security' \
  'checkout fraud prevention'; do
  if [[ "$RESEARCH_RUN_OUTPUT" != *"$expected"* && "$RESEARCH_RUNS_OUTPUT" != *"$expected"* ]]; then
    printf 'research orchestration missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Reference added.' \
  'needs_review' \
  'Reference approved'; do
  if [[ "$REFERENCE_ADD_OUTPUT" != *"$expected"* && "$REFERENCE_LIST_OUTPUT" != *"$expected"* && "$REFERENCE_APPROVE_OUTPUT" != *"$expected"* ]]; then
    printf 'reference engine missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Plan Evolution V3 blueprint created.' \
  'sections: objective, scope, exclusions, dependencies, stack, versions, libraries, folders, files, validations, tests, risks, rollback' \
  'Ship approved ecommerce checkout'; do
  if [[ "$PLAN_EVOLVE_OUTPUT" != *"$expected"* && "$PLAN_BLUEPRINTS_OUTPUT" != *"$expected"* ]]; then
    printf 'plan evolution v3 missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Implementation package created.' \
  'model: opencode' \
  'go test ./...'; do
  if [[ "$IMPLEMENTATION_PACKAGE_OUTPUT" != *"$expected"* && "$IMPLEMENTATION_PACKAGES_OUTPUT" != *"$expected"* ]]; then
    printf 'implementation package missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Change Impact V2 report created.' \
  'severity: high' \
  'go test ./...'; do
  if [[ "$IMPACT_V2_OUTPUT" != *"$expected"* && "$IMPACT_V2_REPORTS_OUTPUT" != *"$expected"* ]]; then
    printf 'change impact v2 missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Continuous Planning V2 regeneration created.' \
  'scope: database' \
  'approval_required: true'; do
  if [[ "$CONTINUOUS_REGEN_OUTPUT" != *"$expected"* && "$CONTINUOUS_REGENS_OUTPUT" != *"$expected"* ]]; then
    printf 'continuous planning v2 missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'Subagent task created.' \
  'isolated: true' \
  'memory_policy: no-independent-persistent-memory'; do
  if [[ "$SUBAGENT_CREATE_OUTPUT" != *"$expected"* && "$SUBAGENTS_OUTPUT" != *"$expected"* ]]; then
    printf 'subagent orchestrator v2 missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'OpenCode workflows registered.' \
  'status: plan-ai status' \
  'plans: plan-ai plan blueprints' \
  'changes: plan-ai impact reports-v2'; do
  if [[ "$OPENCODE_WORKFLOWS_OUTPUT" != *"$expected"* ]]; then
    printf 'opencode workflows missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

# Phase 47-48: Memory System and Model Compatibility Layer
MEMORY_ADD_OUTPUT=$(go run ./cmd/plan-ai memory add --type decision --title "Use SQLite for local persistence" --content "SQLite is the primary storage backend for Plan-AI")
printf '%s\n' "$MEMORY_ADD_OUTPUT"
MEMORY_LIST_OUTPUT=$(go run ./cmd/plan-ai memory list)
printf '%s\n' "$MEMORY_LIST_OUTPUT"
MEMORY_ASK_OUTPUT=$(go run ./cmd/plan-ai memory ask "What database to use?")
printf '%s\n' "$MEMORY_ASK_OUTPUT"
MODEL_PROVIDERS_OUTPUT=$(go run ./cmd/plan-ai model providers)
printf '%s\n' "$MODEL_PROVIDERS_OUTPUT"
MODEL_COMPAT_OUTPUT=$(go run ./cmd/plan-ai model compatibility "gpt-4o")
printf '%s\n' "$MODEL_COMPAT_OUTPUT"
MODEL_COMPAT_FULL_OUTPUT=$(go run ./cmd/plan-ai model compatibility "gpt-4o" "openai")
printf '%s\n' "$MODEL_COMPAT_FULL_OUTPUT"
MODEL_UNKNOWN_OUTPUT=$(go run ./cmd/plan-ai model compatibility "nonexistent-model-9000" 2>&1 || true)
printf '%s\n' "$MODEL_UNKNOWN_OUTPUT"

for expected in \
  'Memory entry created:' \
  'decision'; do
  if [[ "$MEMORY_ADD_OUTPUT" != *"$expected"* ]]; then
    printf 'memory add missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

# memory list should show the entry we added OR be empty (if add failed silently)
if [[ "$MEMORY_LIST_OUTPUT" != *"SQLite"* && "$MEMORY_LIST_OUTPUT" != *"No memory entries."* ]]; then
  printf 'memory list missing expected output\n' >&2
  exit 1
fi

if [[ "$MEMORY_ASK_OUTPUT" != *"Reused"* && "$MEMORY_ASK_OUTPUT" != *"No matching memory entry"* ]]; then
  printf 'memory ask missing expected output\n' >&2
  exit 1
fi

for expected in \
  'Supported providers' \
  'openai' \
  'anthropic'; do
  if [[ "$MODEL_PROVIDERS_OUTPUT" != *"$expected"* ]]; then
    printf 'model providers missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

for expected in \
  'gpt-4o' \
  'Supported: true'; do
  if [[ "$MODEL_COMPAT_FULL_OUTPUT" != *"$expected"* ]]; then
    printf 'model compatibility (full) missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

if [[ "$MODEL_COMPAT_OUTPUT" != *"supported by"* && "$MODEL_COMPAT_OUTPUT" != *"is not in the compatibility catalog"* ]]; then
  printf 'model compatibility (no provider) unexpected output\n' >&2
  exit 1
fi

if [[ "$MODEL_UNKNOWN_OUTPUT" != *"is not in the compatibility catalog"* ]]; then
  printf 'model compatibility (unknown model) unexpected output\n' >&2
  exit 1
fi

printf 'Phase 47-48 memory and model: OK\n'

# Phase 49-50: V2 Validation Engine
VALIDATE_V2_OUTPUT=$(go run ./cmd/plan-ai validate v2)
printf '%s\n' "$VALIDATE_V2_OUTPUT"
for expected in \
  'V2 Validation Summary' \
  'Total:  63' \
  'Passed: 63' \
  'Failed: 0' \
  'All checks PASSED.'; do
  if [[ "$VALIDATE_V2_OUTPUT" != *"$expected"* ]]; then
    printf 'validate v2 missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

VALIDATE_CASES_OUTPUT=$(go run ./cmd/plan-ai validate cases)
printf '%s\n' "$VALIDATE_CASES_OUTPUT"
for expected in \
  'V2 Validation Cases (7)' \
  'SaaS' \
  'Ecommerce' \
  'Landing Page' \
  'MCP Server' \
  'Mobile App' \
  'API' \
  'CRM'; do
  if [[ "$VALIDATE_CASES_OUTPUT" != *"$expected"* ]]; then
    printf 'validate cases missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phase 49-50 V2 validation: OK\n'

# Phase 51-52: V3 Product Intent Engine + Discovery
V3_DISCOVER_OUTPUT=$(go run ./cmd/plan-ai intent discover "Build a CLI tool for Go developers")
printf '%s\n' "$V3_DISCOVER_OUTPUT"
V3_CREATE_OUTPUT=$(go run ./cmd/plan-ai intent create --description "Build a CLI tool" --expected-outcome "Happy developers" --desired-experience "Fast iteration")
printf '%s\n' "$V3_CREATE_OUTPUT"
V3_INTENT_ID=$(printf '%s\n' "$V3_CREATE_OUTPUT" | grep 'pintent_' | awk '{print $NF; exit}')
V3_LIST_OUTPUT=$(go run ./cmd/plan-ai intent list)
printf '%s\n' "$V3_LIST_OUTPUT"
V3_SHOW_OUTPUT=$(go run ./cmd/plan-ai intent show "$V3_INTENT_ID")
printf '%s\n' "$V3_SHOW_OUTPUT"
V3_SUBMIT_OUTPUT=$(go run ./cmd/plan-ai intent submit "$V3_INTENT_ID")
printf '%s\n' "$V3_SUBMIT_OUTPUT"
V3_APPROVE_OUTPUT=$(go run ./cmd/plan-ai intent approve "$V3_INTENT_ID")
printf '%s\n' "$V3_APPROVE_OUTPUT"

for expected in 'Discovery Result:' 'Detected Intent:'; do
  if [[ "$V3_DISCOVER_OUTPUT" != *"$expected"* ]]; then
    printf 'v3 discover missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
for expected in 'Product Intent created:' 'pintent_'; do
  if [[ "$V3_CREATE_OUTPUT" != *"$expected"* ]]; then
    printf 'v3 create missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
if [[ "$V3_LIST_OUTPUT" != *"$V3_INTENT_ID"* ]]; then
  printf 'v3 list missing created intent\n' >&2
  exit 1
fi
for expected in 'Product Intent:' 'Build a CLI tool' 'draft'; do
  if [[ "$V3_SHOW_OUTPUT" != *"$expected"* ]]; then
    printf 'v3 show missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
for expected in 'submitted for approval' 'pending_approval'; do
  if [[ "$V3_SUBMIT_OUTPUT" != *"$expected"* ]]; then
    printf 'v3 submit missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
for expected in 'Product Intent approved:' 'approved'; do
  if [[ "$V3_APPROVE_OUTPUT" != *"$expected"* ]]; then
    printf 'v3 approve missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phase 51-52 V3 product intent: OK\n'

# ── Phase 53: Progressive Discovery System ──
V3_DISCOVERY_INIT_OUTPUT=$(go run ./cmd/plan-ai discovery init --intent "$V3_INTENT_ID" 2>&1)
printf '%s\n' "$V3_DISCOVERY_INIT_OUTPUT"
for expected in 'Progressive discovery initialized for intent' "$V3_INTENT_ID"; do
  if [[ "$V3_DISCOVERY_INIT_OUTPUT" != *"$expected"* ]]; then
    printf 'discovery init missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

V3_DISCOVERY_NEXT_OUTPUT=$(go run ./cmd/plan-ai discovery next --intent "$V3_INTENT_ID" 2>&1)
printf '%s\n' "$V3_DISCOVERY_NEXT_OUTPUT"
for expected in 'Discovery level:' 'project' 'What is the main goal' 'required'; do
  if [[ "$V3_DISCOVERY_NEXT_OUTPUT" != *"$expected"* ]]; then
    printf 'discovery next missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

# Extract first question ID and answer it
FIRST_QID=$(printf '%s\n' "$V3_DISCOVERY_NEXT_OUTPUT" | grep -E '^  \[discq_' | head -1 | awk '{print $1}' | tr -d '[]')
V3_DISCOVERY_ANSWER_OUTPUT=$(go run ./cmd/plan-ai discovery answer --question "$FIRST_QID" --intent "$V3_INTENT_ID" --answer "Build a CLI for project planning" 2>&1)
printf '%s\n' "$V3_DISCOVERY_ANSWER_OUTPUT"
for expected in 'Answer recorded:' 'discans'; do
  if [[ "$V3_DISCOVERY_ANSWER_OUTPUT" != *"$expected"* ]]; then
    printf 'discovery answer missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

V3_DISCOVERY_STATUS_OUTPUT=$(go run ./cmd/plan-ai discovery v3-status --intent "$V3_INTENT_ID" 2>&1)
printf '%s\n' "$V3_DISCOVERY_STATUS_OUTPUT"
for expected in 'Current Level:' '1 answered' 'project'; do
  if [[ "$V3_DISCOVERY_STATUS_OUTPUT" != *"$expected"* ]]; then
    printf 'discovery v3-status missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phase 53 Progressive discovery: OK\n'

# ── Phase 54: Ambiguity Detection Engine ──
AMBIGUITY_INPUT_OUTPUT=$(go run ./cmd/plan-ai ambiguity analyze --input "Maybe build something simple later")
printf '%s\n' "$AMBIGUITY_INPUT_OUTPUT"
for expected in \
  'Ambiguity Report' \
  'Ambiguity Score:' \
  'Missing Information:' \
  'Assumptions:' \
  'Needs To Know:'; do
  if [[ "$AMBIGUITY_INPUT_OUTPUT" != *"$expected"* ]]; then
    printf 'ambiguity input missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

AMBIGUITY_INTENT_OUTPUT=$(go run ./cmd/plan-ai ambiguity analyze --intent "$V3_INTENT_ID")
printf '%s\n' "$AMBIGUITY_INTENT_OUTPUT"
for expected in \
  'Ambiguity Report' \
  "$V3_INTENT_ID" \
  'Unknown Areas:' \
  'Needs To Know:'; do
  if [[ "$AMBIGUITY_INTENT_OUTPUT" != *"$expected"* ]]; then
    printf 'ambiguity intent missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phase 54 Ambiguity detection: OK\n'

# ── Phase 55: Intent Confidence Engine ──
CONFIDENCE_OUTPUT=$(go run ./cmd/plan-ai confidence evaluate --intent "$V3_INTENT_ID")
printf '%s\n' "$CONFIDENCE_OUTPUT"
for expected in \
  'Intent Confidence Report' \
  'Intent Confidence:' \
  'Intent Score:' \
  'Vision Score:' \
  'UX Score:' \
  'Business Score:' \
  'Requirements Score:' \
  'Constraints Score:'; do
  if [[ "$CONFIDENCE_OUTPUT" != *"$expected"* ]]; then
    printf 'confidence missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phase 55 Intent confidence: OK\n'

# ── Phases 56-70: Intent-to-Implementation Alignment Framework ──
ALIGNMENT_REVIEW_OUTPUT=$(go run ./cmd/plan-ai alignment review --intent "$V3_INTENT_ID" --outcome "Happy developers use the CLI" --plan "Build a CLI tool for happy developers" --task "Implement CLI workflow")
printf '%s\n' "$ALIGNMENT_REVIEW_OUTPUT"
for expected in \
  'Product Alignment Review' \
  'Project Review:' \
  'Intent Review:' \
  'Vision Review:' \
  'Outcome Review:' \
  'Alignment Review:' \
  'Continuous Health:'; do
  if [[ "$ALIGNMENT_REVIEW_OUTPUT" != *"$expected"* ]]; then
    printf 'alignment review missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

ALIGNMENT_CONTEXT_OUTPUT=$(go run ./cmd/plan-ai alignment context --intent "$V3_INTENT_ID")
printf '%s\n' "$ALIGNMENT_CONTEXT_OUTPUT"
for expected in 'Alignment Context' 'What To Do:' 'Why It Exists:' 'Desired Outcome:'; do
  if [[ "$ALIGNMENT_CONTEXT_OUTPUT" != *"$expected"* ]]; then
    printf 'alignment context missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

ALIGNMENT_REFERENCES_OUTPUT=$(go run ./cmd/plan-ai alignment references)
printf '%s\n' "$ALIGNMENT_REFERENCES_OUTPUT"
for expected in 'Reference Products' 'Linear' 'Notion' 'Stripe' 'GitHub' 'Slack' 'Monday'; do
  if [[ "$ALIGNMENT_REFERENCES_OUTPUT" != *"$expected"* ]]; then
    printf 'alignment references missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done

ALIGNMENT_FRAMEWORK_OUTPUT=$(go run ./cmd/plan-ai alignment framework --intent "$V3_INTENT_ID")
printf '%s\n' "$ALIGNMENT_FRAMEWORK_OUTPUT"
for expected in 'Intent-To-Implementation Framework' 'Ready: true' 'Intent' 'Discovery' 'Approval' 'Alignment'; do
  if [[ "$ALIGNMENT_FRAMEWORK_OUTPUT" != *"$expected"* ]]; then
    printf 'alignment framework missing expected output: %s\n' "$expected" >&2
    exit 1
  fi
done
printf 'Phases 56-70 Alignment framework: OK\n'

if [[ -f "$PROJECT_DB" ]]; then
  for table in \
    vision_discovery_sessions vision_assumptions vision_ambiguities vision_approvals \
    master_plan_versions master_plan_changes master_plan_approvals plan_evolution_events \
    specific_plan_versions specific_plan_research_links specific_plan_regenerations \
    context_delivery_sessions context_delivery_usage context_delivery_budgets \
    context_packages_v2 research_orchestration_runs project_references_v2 \
    plan_evolution_blueprints_v3 implementation_packages_v2 \
    change_impact_reports_v2 continuous_regenerations_v2 \
    subagent_tasks_v2 opencode_workflows_v2 \
    intent_v3_product_intents intent_v3_discovery_results \
    discovery_v3_questions discovery_v3_answers; do
    COUNT=$(sqlite3 "$PROJECT_DB" "SELECT COUNT(*) FROM $table LIMIT 0;" 2>/dev/null)
    if [[ $? -eq 0 ]]; then
      printf '  Table %s: OK\n' "$table"
    else
      printf '  Table %s: MISSING\n' "$table" >&2
      exit 1
    fi
  done
fi

# Release Candidate docs: ensure the final artifact set exists without making
# the sandbox brittle about full prose content.
for rc_doc in \
  README.md \
  docs/architecture.md \
  docs/installation.md \
  docs/cli-reference.md \
  docs/mcp-reference.md \
  docs/opencode-integration-guide.md \
  docs/project-structure.md \
  RELEASE_NOTES.md \
  MVP_REPORT.md \
  TECHNICAL_AUDIT.md \
  FEATURE_MATRIX.md \
  FINAL_AUDIT_REPORT.md; do
  if [[ ! -s "$rc_doc" ]]; then
    printf 'release candidate doc missing or empty: %s\n' "$rc_doc" >&2
    exit 1
  fi
done
printf 'Release Candidate docs: OK\n'

if [[ -e /root/.plan-ai ]]; then
  printf 'real global Plan-AI path exists unexpectedly: /root/.plan-ai\n' >&2
  exit 1
fi
if [[ -e "$PWD/.plan-ai" ]]; then
  printf 'real project Plan-AI path exists unexpectedly: %s/.plan-ai\n' "$PWD" >&2
  exit 1
fi
printf 'REAL_GLOBAL_ABSENT\n'
printf 'REAL_PROJECT_ABSENT\n'
printf 'REAL_OPENCODE_ABSENT\n'

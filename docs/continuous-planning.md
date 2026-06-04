# Continuous Planning

Phase 22 adds continuous planning.

Continuous planning never modifies plans automatically. The lifecycle is:

1. Detect
2. Analyze
3. Propose
4. Wait for approval
5. Apply only after approval

## Events

Supported event types:

- `new_approved_context`
- `new_research`
- `new_knowledge`
- `decision_changed`
- `plan_outdated`
- `implementation_feedback`
- `task_completed`
- `validation_failed`
- `change_request_created`

## Plan update proposals

A `PlanUpdateProposal` records why a plan might need to change, what plans/tasks/decisions are affected, whether research is required, whether approval is required, and its status.

Statuses:

- `draft`
- `pending_approval`
- `approved`
- `rejected`
- `applied`

## Persistence

- `0020_continuous_planning` creates physical continuous planning tables.
- `0021_agent_continuous_compatibility` adds contract-compatible names including `continuous_status` and `context_delivery_logs`.

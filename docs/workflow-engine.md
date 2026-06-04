# Workflow Engine

Phase 14 formalizes workflows without adding intelligence.

Registered workflows:
- Vision Workflow: Input → Vision → Approval
- Research Workflow: Topic → Research → Knowledge
- Planning Workflow: Vision + Approved Context + Research + Knowledge → Master Plan → Specific Plan
- Approval Workflow: Draft → Review → Approved/Rejected

The engine persists `workflow_runs` with `id`, `workflow_type`, `status`, `started_at`, and `finished_at`.

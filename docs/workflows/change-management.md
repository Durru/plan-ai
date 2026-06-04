# Change Management Workflow

Plan-AI responds to project changes by analyzing impact before rewriting plans. The goal is to keep planning state aligned with project reality without losing useful history.

## Change Categories

Plan-AI recognizes five common categories of change:

- stack changes;
- architecture changes;
- library changes;
- decision changes;
- scope changes.

Each category may affect plans, phases, tasks, validations, risks, dependencies, and approval status.

## Stack Change

A stack change modifies the core technology foundation, such as language, framework, runtime, storage, deployment model, or primary tooling.

Plan-AI should:

1. Identify which plans assume the previous stack.
2. Reevaluate architecture and integration assumptions.
3. Update affected tasks and validations.
4. Record a decision revision.
5. Request renewed approval when cost, risk, or scope changes.

## Architecture Change

An architecture change modifies system boundaries, module responsibilities, data flow, control flow, or integration patterns.

Plan-AI should:

1. Map affected modules and planning objects.
2. Identify tasks that now target the wrong boundary.
3. Update phase ordering if dependencies changed.
4. Revise risks and validation strategy.
5. Preserve the reason the previous architecture was replaced.

## Library Change

A library change modifies an implementation dependency or planned tool. Library changes can be local or project-wide.

Plan-AI should:

1. Determine whether the change is isolated or architectural.
2. Review compatibility, maintenance, licensing, security, and migration impact.
3. Update research and knowledge objects.
4. Revise tasks that mention the previous library.
5. Request approval if the library choice affects project direction or risk.

## Decision Change

A decision change revises a recorded project choice. This is always a planning event.

Plan-AI should:

1. Version the decision.
2. Explain why the previous decision changed.
3. Trace affected plans, phases, tasks, and documents.
4. Update downstream planning objects.
5. Mark whether renewed approval is required.

## Scope Change

A scope change adds, removes, or redefines what the project is preparing to implement.

Plan-AI should:

1. Update scope and non-scope explicitly.
2. Add, remove, split, or merge affected phases and tasks.
3. Reevaluate roadmap and dependencies.
4. Update risks and success criteria.
5. Request approval when expectations, timeline, or deliverables change.

## Change Severity

Plan-AI classifies changes by severity:

| Severity | Meaning | Response |
|---|---|---|
| Informational | Adds context without changing the plan | Record knowledge if reusable |
| Local | Affects one task or small area | Update affected task |
| Plan-affecting | Changes phases, validations, or dependencies | Revise plan section |
| Decision-affecting | Changes a committed choice | Version decision and analyze impact |
| Approval-affecting | Changes scope, risk, cost, or direction | Require renewed approval |

## Replanning Rule

Plan-AI should never rewrite more than the change justifies. Preserve valid planning work, revise affected objects, and explain what remains unchanged.

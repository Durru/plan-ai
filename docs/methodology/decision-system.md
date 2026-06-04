# Decision System

Plan-AI treats decisions as durable planning records. A decision explains why the project moves in one direction instead of another and what parts of the plan are affected.

## What a Decision Is

A decision is a committed choice with project impact. It is stronger than an opinion and more durable than a conversation note.

A choice should be recorded as a decision when it affects:

- architecture or system boundaries;
- technology stack or libraries;
- scope or non-scope;
- security, data, or privacy posture;
- validation strategy;
- integration strategy;
- task ordering or phase structure;
- future reversibility or migration cost.

## Decision Creation

Plan-AI creates a decision record when analysis shows that a choice materially constrains the plan. A decision record should include context, problem, options, chosen decision, consequences, impact, and confidence.

Creation flow:

1. Identify the planning pressure that requires a choice.
2. State the problem in neutral terms.
3. Compare realistic options.
4. Recommend or record the chosen option.
5. Describe consequences and tradeoffs.
6. Map impact to plans, phases, tasks, and documents.
7. Mark approval status.

## Decision Modification

Changing a decision is not a simple edit. It is a project event because downstream plans may depend on the previous choice.

Modification flow:

1. State what changed and why the previous decision is no longer sufficient.
2. Identify affected planning objects.
3. Evaluate whether the original options are still relevant.
4. Record the new decision and preserve the previous rationale.
5. Update impacted plans, phases, tasks, knowledge objects, and documentation.
6. Request renewed approval when scope, risk, cost, or architecture changes.

## Project Effects

Each decision should describe its effects in practical terms:

- **Plan effect:** which plan sections change.
- **Phase effect:** which phase ordering or boundaries change.
- **Task effect:** which tasks are added, removed, split, or revised.
- **Validation effect:** which checks become necessary or obsolete.
- **Knowledge effect:** which research or recommendations are superseded.
- **Approval effect:** whether human approval must be renewed.

## Versioning

Decisions should preserve history. A decision may have revisions, but earlier reasoning remains useful for understanding why the project changed direction.

A versioned decision should track:

- decision identifier;
- current status;
- creation date or event;
- revision events;
- changed fields;
- reason for revision;
- impacted planning objects.

Versioning prevents silent drift. Future developers should be able to see not only what the project decided, but how the decision evolved.

## Impact

Decision impact is the connection between a choice and the rest of the planning system. Plan-AI should never treat major decisions as isolated notes.

Impact analysis should answer:

- Which assumptions rely on this decision?
- Which tasks become invalid if the decision changes?
- Which risks are introduced or reduced?
- Which validations are required because of the decision?
- Which stakeholders or users are affected?
- Which documents must be updated?

The larger the impact, the more explicit the approval and documentation requirements.

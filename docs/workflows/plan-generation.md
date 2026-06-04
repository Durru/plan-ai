# Plan Generation Workflow

Plan-AI generates plans in layers. It starts from broad intent and ends with implementation-ready tasks while preserving rationale and validation.

## Output Layers

Plan generation produces four main layers:

1. **Master Plan:** the project-level direction.
2. **Specific Plan:** the plan for a feature, capability, change, or workstream.
3. **Phases:** ordered groups of coherent work.
4. **Tasks:** executable implementation-preparation units.

## Step 1: Establish Planning Inputs

Plan-AI begins by collecting the minimum reliable inputs:

- idea or objective;
- existing context;
- known constraints;
- relevant decisions;
- research findings;
- risks and dependencies;
- approval expectations.

If a missing input would materially change the plan, Plan-AI asks a clarification question or performs research before proceeding.

## Step 2: Generate the Master Plan

The Master Plan defines the project direction. It should describe vision, objective, scope, non-scope, users, architecture, risks, dependencies, roadmap, and success criteria.

The Master Plan answers: what are we preparing, why does it matter, what boundaries apply, and how will we know the project is ready to implement?

## Step 3: Generate Specific Plans

A Specific Plan narrows the Master Plan into a feature, module, workflow, integration, or change. It includes objective, scope, non-scope, research, alternatives, decision taken, structure, validations, risks, dependencies, tasks, and impact.

The Specific Plan answers: how should this workstream be prepared under the approved project direction?

## Step 4: Derive Phases

Plan-AI converts a Specific Plan into phases based on dependency order, integration points, validation boundaries, and reviewability.

Each phase should define:

- objective;
- required inputs;
- expected outputs;
- validations;
- dependencies.

Phases should create meaningful checkpoints. A phase that cannot be validated is not a useful planning boundary.

## Step 5: Derive Tasks

Plan-AI decomposes each phase into tasks. Tasks should be specific enough for a developer or implementation agent to execute without rediscovering the plan.

Each task should define:

- objective;
- context;
- affected files or areas;
- restrictions;
- ordered steps;
- validations;
- expected result.

## Step 6: Validate the Plan Structure

Before presenting the plan for approval, Plan-AI checks that:

- every task traces to a phase;
- every phase supports a plan objective;
- every major decision is recorded;
- scope and non-scope are explicit;
- validations exist for meaningful outcomes;
- risks and dependencies are visible;
- implementation boundaries are respected.

## Step 7: Prepare for Approval

Plan-AI presents the plan with enough detail for review. It should identify decisions needing confirmation, risks requiring acceptance, and any assumptions that remain unresolved.

Approval turns a candidate plan into the reference for implementation preparation.

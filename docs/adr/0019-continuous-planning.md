# ADR 0019: Continuous Planning

## Status

Accepted

## Context

After plans, workflows, change analysis, and agent routing exist, Plan-AI must keep project state coherent as new decisions, research, knowledge, validation results, and implementation feedback arrive.

## Decision

Add Phase 22 Continuous Planning with a strict lifecycle:

1. detect;
2. analyze;
3. propose;
4. wait for approval;
5. apply only after approval.

Continuous planning produces plan update proposals, continuous events, continuous status, and final context delivery levels L0-L4.

## Persistence

Because migration `0018` was already used by an earlier compatibility layer, Phase 22 runtime persistence uses `0020_continuous_planning`, and `0021_agent_continuous_compatibility` adds contract-compatible names. This ADR remains `0019-continuous-planning` to match the roadmap phase documentation.

## Consequences

Plan-AI can identify outdated project state and propose safe updates without silently modifying approved work.

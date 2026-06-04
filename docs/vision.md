# Vision Engine

The vision engine creates an incomplete, user-reviewable vision draft from
ingested sources.

## Responsibilities

- Extract the main objective.
- Detect target users when explicitly mentioned.
- Capture mentioned features and goals.
- Capture constraints and preferences.
- Capture visual references and success criteria.
- Record missing information instead of inventing answers.

## Approval rule

A vision draft is never approved automatically. Approval requires an explicit
user action via the service or CLI.

## Persistence

Vision drafts are stored in `visions` with structured JSON list columns for
target users, goals, constraints, assumptions, missing information, visual
references, and success criteria.

# Approved Context Engine

The approved context engine is the durable registry for facts the user has
approved. Its fundamental rule is:

> Once approved, do not ask it again.

## Approved item types

- `requirement`
- `constraint`
- `decision`
- `preference`
- `goal`
- `reference`

## Registry API

- `StoreApproved()` stores approved knowledge and deduplicates by project and
  content.
- `GetApproved()` retrieves one approved item by type and id.
- `ListApproved()` retrieves approved context for a project.
- `FindApproved()` searches approved context without relying on chat history.

Only approved knowledge is persisted in the approved context tables. Drafts,
research, and unapproved observations remain outside this registry.

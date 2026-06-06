# Snapshots

Snapshots capture the state of planning entities at a point in time.
They serve as checkpoints before applying changes, enabling rollback
and audit of the planning process.

## V1 Snapshots

The original `snapshots` table (Phase 2) stores basic snapshots with
project ID, reason, summary, and timestamps.

Repository: `store.SnapshotRepository` in `internal/store/domain_repositories.go`

## V2 Snapshots

The `snapshots_v2` table (Phase 18) extends snapshots with structured
entity snapshots (JSON), enabling granular tracking of which entities
were captured.

Repository: `store.SnapshotV2Repository` in `internal/store/phase18_20_repositories.go`

## Creating Snapshots

### Via Go API

```go
svc := change.NewService(changeStore, snapshotStore)
snapID, err := svc.CreateSnapshot("pre-upgrade checkpoint")
```

### Via CLI

```sh
plan-ai snapshot create "pre-upgrade checkpoint"
plan-ai snapshot list
```

### Via MCP

```sh
plan-ai mcp call-tool plan_ai.create_snapshot '{"reason": "before architecture change"}'
```

## Snapshot Lifecycle

1. Change event occurs (vision update, requirement change, etc.)
2. Service automatically creates a snapshot before applying invalidation
3. Impact analysis runs against the snapshot as baseline
4. Entity states are updated with invalidation status
5. Report is generated linking the change, snapshot, and affected entities

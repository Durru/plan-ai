package change

import "time"

// SnapshotManager handles the creation, listing, and retrieval
// of project state snapshots.
type SnapshotManager struct {
	store SnapshotStore
}

// NewSnapshotManager creates a new snapshot manager.
func NewSnapshotManager(store SnapshotStore) *SnapshotManager {
	return &SnapshotManager{store: store}
}

// Create captures a snapshot of the current project state.
func (m *SnapshotManager) Create(projectID, reason string, entities []AffectedEntity) (*Snapshot, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	snap := &Snapshot{
		ID:               generateID("snap"),
		ProjectID:        projectID,
		Timestamp:        now,
		Reason:           reason,
		AffectedEntities: entities,
		CreatedAt:        now,
	}
	if err := m.store.Save(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

// Get retrieves a snapshot by ID.
func (m *SnapshotManager) Get(id string) (*Snapshot, error) {
	return m.store.Get(id)
}

// List returns recent snapshots for a project.
func (m *SnapshotManager) List(projectID string, limit int) ([]Snapshot, error) {
	if limit <= 0 {
		limit = 50
	}
	return m.store.List(projectID, limit)
}

func generateID(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000")
}

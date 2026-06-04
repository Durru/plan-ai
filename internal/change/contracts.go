package change

// ChangeReport summarizes what changed, what it affects, what must be reviewed,
// and what can continue unaffected.
type ChangeReport struct {
	Change       ChangeEvent     `json:"change"`
	Impact       *ImpactAnalysis `json:"impact,omitempty"`
	MustReview   []EntityState   `json:"must_review"`
	CanContinue  []string        `json:"can_continue"`
	Invalidation []EntityState   `json:"invalidation"`
	SnapshotID   string          `json:"snapshot_id,omitempty"`
}

// ImpactAnalysis describes the entities affected by a change.
type ImpactAnalysis struct {
	ChangeID        string              `json:"change_id"`
	AffectedTypes   map[string][]string `json:"affected_types"` // entity_type -> []entity_id
	AffectedByType  []AffectedGroup     `json:"affected_by_type"`
	Summary         string              `json:"summary"`
	ReviewRequired  bool                `json:"review_required"`
	BlockingChanges []string            `json:"blocking_changes,omitempty"`
}

// AffectedGroup groups affected entities by type.
type AffectedGroup struct {
	EntityType string   `json:"entity_type"`
	EntityIDs  []string `json:"entity_ids"`
}

// ChangeEngine is the core interface for the change management subsystem.
type ChangeEngine interface {
	RegisterChange(ChangeEvent) (*ChangeReport, error)
	AnalyzeImpact(changeID string) (*ImpactAnalysis, error)
	CreateSnapshot(reason string) (string, error)
	ListChanges(projectID string, limit int) ([]ChangeEvent, error)
	GetChange(changeID string) (*ChangeEvent, error)
	GetReport(changeID string) (*ChangeReport, error)
	InvalidateEntities(changeID string) ([]EntityState, error)
	GetEntityState(entityType, entityID string) (*EntityState, error)
}

// SnapshotStore defines persistence for snapshots.
type SnapshotStore interface {
	Save(snapshot *Snapshot) error
	Get(id string) (*Snapshot, error)
	List(projectID string, limit int) ([]Snapshot, error)
}

// Snapshot represents a point-in-time capture of project state.
type Snapshot struct {
	ID               string           `json:"id"`
	ProjectID        string           `json:"project_id"`
	Timestamp        string           `json:"timestamp"`
	Reason           string           `json:"reason"`
	AffectedEntities []AffectedEntity `json:"affected_entities,omitempty"`
	ChangeID         string           `json:"change_id,omitempty"`
	CreatedAt        string           `json:"created_at"`
}

// AffectedEntity identifies a single entity in a snapshot.
type AffectedEntity struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
}

// ChangeStore defines persistence for change events.
type ChangeStore interface {
	SaveEvent(event *ChangeEvent) error
	GetEvent(id string) (*ChangeEvent, error)
	ListEvents(projectID string, limit int) ([]ChangeEvent, error)
	SaveImpact(analysis *ImpactAnalysis) error
	GetImpact(changeID string) (*ImpactAnalysis, error)
	SaveEntityState(state *EntityState) error
	GetEntityState(entityType, entityID string) (*EntityState, error)
	ListEntityStates(entityType string) ([]EntityState, error)
}

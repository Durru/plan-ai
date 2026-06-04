package change

// VersionManager tracks plan versioning and invalidation state.
// When a change affects a plan, its version is considered stale
// until explicitly reviewed and re-approved.
type VersionManager struct {
	store ChangeStore
}

// NewVersionManager creates a version manager.
func NewVersionManager(store ChangeStore) *VersionManager {
	return &VersionManager{store: store}
}

// InvalidatePlan marks a planning entity as needing review or blocked.
func (vm *VersionManager) InvalidatePlan(entityType, entityID, changeID, reason string, status EntityStatus) (*EntityState, error) {
	state := &EntityState{
		EntityType:   entityType,
		EntityID:     entityID,
		Status:       status,
		LastChangeID: changeID,
		Reason:       reason,
	}
	if err := vm.store.SaveEntityState(state); err != nil {
		return nil, err
	}
	return state, nil
}

// CurrentStatus returns the current invalidation state of an entity.
func (vm *VersionManager) CurrentStatus(entityType, entityID string) (*EntityState, error) {
	return vm.store.GetEntityState(entityType, entityID)
}

// NeedsReview checks whether an entity requires review before use.
func (vm *VersionManager) NeedsReview(entityType, entityID string) (bool, error) {
	state, err := vm.store.GetEntityState(entityType, entityID)
	if err != nil {
		return true, err // conservatively assume review needed
	}
	return state.Status == EntityNeedsReview || state.Status == EntityOutdated || state.Status == EntityBlocked, nil
}

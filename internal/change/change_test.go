package change

import (
	"fmt"
	"testing"
)

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	types := r.List()
	if len(types) != 10 {
		t.Fatalf("expected 10 change types, got %d", len(types))
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	meta, ok := r.Get(VisionChanged)
	if !ok {
		t.Fatal("expected VisionChanged to be registered")
	}
	if meta.DisplayName != "Vision Changed" {
		t.Fatalf("expected 'Vision Changed', got %q", meta.DisplayName)
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get(ChangeType("nonexistent"))
	if ok {
		t.Fatal("expected nonexistent type to not be found")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	r.Register(TypeMeta{
		Type:        "custom_type",
		DisplayName: "Custom Type",
		Description: "A custom change type",
		Severity:    SeverityLow,
	})
	meta, ok := r.Get("custom_type")
	if !ok {
		t.Fatal("expected custom type to be registered")
	}
	if meta.DisplayName != "Custom Type" {
		t.Fatalf("expected 'Custom Type', got %q", meta.DisplayName)
	}
}

func TestClassifySeverity(t *testing.T) {
	tests := []struct {
		ct       ChangeType
		expected Severity
	}{
		{VisionChanged, SeverityHigh},
		{RequirementRemoved, SeverityHigh},
		{PlanChanged, SeverityHigh},
		{RequirementAdded, SeverityMedium},
		{ConstraintChanged, SeverityMedium},
		{DecisionChanged, SeverityMedium},
		{ResearchUpdated, SeverityLow},
		{KnowledgeUpdated, SeverityLow},
		{TechnologyChanged, SeverityLow},
		{ImplementationFeedback, SeverityLow},
	}
	for _, tt := range tests {
		got := ClassifySeverity(tt.ct)
		if got != tt.expected {
			t.Errorf("ClassifySeverity(%q) = %q, want %q", tt.ct, got, tt.expected)
		}
	}
}

func TestEntityAnalyzer_AnalyzeChange(t *testing.T) {
	a := NewAnalyzer(nil)
	impact := a.AnalyzeChange(VisionChanged)
	if !impact.ReviewRequired {
		t.Fatal("VisionChanged should require review")
	}
	if len(impact.AffectedByType) == 0 {
		t.Fatal("expected at least one affected entity type")
	}
}

func TestEntityAnalyzer_AffectedByChangeType(t *testing.T) {
	a := NewAnalyzer(nil)
	types := a.AffectedByChangeType(RequirementAdded)
	if len(types) == 0 {
		t.Fatal("RequirementAdded should affect at least one entity type")
	}
	// Requirement changes affect these entity types
	expected := []string{"master_plan", "specific_plan", "phase", "requirement"}
	for _, exp := range expected {
		found := false
		for _, t2 := range types {
			if t2 == exp {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("RequirementAdded should affect %q entities", exp)
		}
	}
}

func TestImpactBuilder_Build(t *testing.T) {
	b := NewImpactBuilder()
	ev := &ChangeEvent{
		ID:         "ev_001",
		ChangeType: PlanChanged,
		Summary:    "Test plan change",
	}
	analysis := b.Build(ev)
	if analysis.ChangeID != "ev_001" {
		t.Fatalf("expected ChangeID 'ev_001', got %q", analysis.ChangeID)
	}
}

func TestSnapshotManager(t *testing.T) {
	store := &mockSnapshotStore{snapshots: map[string]*Snapshot{}}
	mgr := NewSnapshotManager(store)

	snap, err := mgr.Create("proj_1", "test snapshot", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ProjectID != "proj_1" {
		t.Fatalf("expected project ID 'proj_1', got %q", snap.ProjectID)
	}
	if snap.Reason != "test snapshot" {
		t.Fatalf("expected reason 'test snapshot', got %q", snap.Reason)
	}

	// Get
	got, err := mgr.Get(snap.ID)
	if err != nil {
		t.Fatalf("unexpected error getting snapshot: %v", err)
	}
	if got.ID != snap.ID {
		t.Fatalf("expected ID %q, got %q", snap.ID, got.ID)
	}

	// List
	list, err := mgr.List("proj_1", 10)
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(list))
	}
}

func TestVersionManager_NeedsReview(t *testing.T) {
	store := &mockChangeStore{states: map[string]*EntityState{}}
	vm := NewVersionManager(store)

	// No state yet — returns error
	_, err := vm.NeedsReview("plan", "plan_1")
	if err == nil {
		t.Fatal("expected error for unknown entity")
	}
}

func TestVersionManager_InvalidatePlan(t *testing.T) {
	store := &mockChangeStore{states: map[string]*EntityState{}}
	vm := NewVersionManager(store)

	state, err := vm.InvalidatePlan("plan", "plan_1", "change_1", "test invalidation", EntityNeedsReview)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != EntityNeedsReview {
		t.Fatalf("expected status %q, got %q", EntityNeedsReview, state.Status)
	}
}

// ──────────────────────────────────────────────
// Mocks
// ──────────────────────────────────────────────

type mockSnapshotStore struct {
	snapshots map[string]*Snapshot
}

func (m *mockSnapshotStore) Save(snap *Snapshot) error {
	key := snap.ID
	if _, exists := m.snapshots[key]; exists {
		return nil // upsert is fine
	}
	m.snapshots[key] = snap
	return nil
}

func (m *mockSnapshotStore) Get(id string) (*Snapshot, error) {
	snap, ok := m.snapshots[id]
	if !ok {
		return nil, snapshotNotFound(id)
	}
	return snap, nil
}

func (m *mockSnapshotStore) List(projectID string, limit int) ([]Snapshot, error) {
	var snaps []Snapshot
	for _, s := range m.snapshots {
		if s.ProjectID == projectID {
			snaps = append(snaps, *s)
		}
	}
	if len(snaps) > limit {
		snaps = snaps[:limit]
	}
	return snaps, nil
}

func snapshotNotFound(id string) error {
	// Return a sentinel-like error; in production this would be sql.ErrNoRows
	return nil
}

type mockChangeStore struct {
	states map[string]*EntityState
}

func (m *mockChangeStore) SaveEntityState(state *EntityState) error {
	m.states[state.EntityType+":"+state.EntityID] = state
	return nil
}

func (m *mockChangeStore) GetEntityState(entityType, entityID string) (*EntityState, error) {
	state, ok := m.states[entityType+":"+entityID]
	if !ok {
		return nil, errNotFound
	}
	return state, nil
}

var errNotFound = fmt.Errorf("not found")

func (m *mockChangeStore) Save(*Snapshot) error                           { return nil }
func (m *mockChangeStore) Get(string) (*Snapshot, error)                  { return nil, nil }
func (m *mockChangeStore) List(string, int) ([]Snapshot, error)           { return nil, nil }
func (m *mockChangeStore) SaveEvent(*ChangeEvent) error                   { return nil }
func (m *mockChangeStore) GetEvent(string) (*ChangeEvent, error)          { return nil, nil }
func (m *mockChangeStore) ListEvents(string, int) ([]ChangeEvent, error)  { return nil, nil }
func (m *mockChangeStore) SaveImpact(*ImpactAnalysis) error               { return nil }
func (m *mockChangeStore) GetImpact(string) (*ImpactAnalysis, error)      { return nil, nil }
func (m *mockChangeStore) ListEntityStates(string) ([]EntityState, error) { return nil, nil }

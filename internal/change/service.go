package change

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// Service implements ChangeEngine and orchestrates the full change lifecycle.
type Service struct {
	store    ChangeStore
	snaps    SnapshotStore
	db       *sql.DB
	analyzer *EntityAnalyzer
	rules    []InvalidationRule
}

// NewService creates a change service with the given stores, database, and default rules.
func NewService(store ChangeStore, snaps SnapshotStore, db *sql.DB) *Service {
	return &Service{
		store:    store,
		snaps:    snaps,
		db:       db,
		analyzer: NewAnalyzer(db),
		rules:    DefaultInvalidationRules,
	}
}

// RegisterChange records a change event, creates a snapshot, analyzes impact,
// invalidates affected entities, and returns a full report.
func (s *Service) RegisterChange(ev ChangeEvent) (*ChangeReport, error) {
	if ev.ID == "" {
		ev.ID = domain.NewID("chg")
	}
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now().UTC()
	}
	if ev.Severity == "" {
		ev.Severity = SeverityMedium
	}

	if err := s.store.SaveEvent(&ev); err != nil {
		return nil, fmt.Errorf("save change event: %w", err)
	}

	// Auto-snapshot before applying
	snapID, err := s.CreateSnapshot(fmt.Sprintf("auto: %s - %s", ev.ChangeType, ev.Summary))
	if err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}

	// Analyze impact
	impact, err := s.analyze(&ev)
	if err != nil {
		return nil, fmt.Errorf("analyze impact: %w", err)
	}
	if err := s.store.SaveImpact(impact); err != nil {
		return nil, fmt.Errorf("save impact: %w", err)
	}

	// Invalidate entities
	invalidation, err := s.invalidate(&ev, impact)
	if err != nil {
		return nil, fmt.Errorf("invalidate entities: %w", err)
	}

	report := &ChangeReport{
		Change:       ev,
		Impact:       impact,
		SnapshotID:   snapID,
		Invalidation: invalidation,
	}

	// Separate must-review vs can-continue
	for _, st := range invalidation {
		if st.Status == EntityBlocked || st.Status == EntityNeedsReview {
			report.MustReview = append(report.MustReview, st)
		} else {
			report.CanContinue = append(report.CanContinue, st.EntityID)
		}
	}

	return report, nil
}

// AnalyzeImpact computes the impact for an already-registered change.
func (s *Service) AnalyzeImpact(changeID string) (*ImpactAnalysis, error) {
	ev, err := s.store.GetEvent(changeID)
	if err != nil {
		return nil, fmt.Errorf("get change %s: %w", changeID, err)
	}
	return s.analyze(ev)
}

// analyze computes the impact of a change event based on invalidation rules
// and entity_links data for transitive resolution.
func (s *Service) analyze(ev *ChangeEvent) (*ImpactAnalysis, error) {
	affected := &ImpactAnalysis{
		ChangeID:       ev.ID,
		AffectedTypes:  make(map[string][]string),
		AffectedByType: []AffectedGroup{},
	}

	for _, rule := range s.rules {
		for _, ct := range rule.AffectedBy {
			if ct == ev.ChangeType {
				affected.AffectedTypes[rule.EntityType] = append(affected.AffectedTypes[rule.EntityType], ev.EntityID)
				if rule.ResultStatus == EntityBlocked || rule.ResultStatus == EntityNeedsReview {
					affected.ReviewRequired = true
				}
				break
			}
		}
	}

	// Resolve real entity IDs from entity_links via the impact graph
	resolved, err := s.analyzer.AnalyzeEntityLinks(ev.ProjectID, ev.EntityType, ev.EntityID)
	if err != nil {
		return nil, err
	}
	for _, group := range resolved {
		existing := affected.AffectedTypes[group.EntityType]
		for _, id := range group.EntityIDs {
			found := false
			for _, eid := range existing {
				if eid == id {
					found = true
					break
				}
			}
			if !found {
				affected.AffectedTypes[group.EntityType] = append(affected.AffectedTypes[group.EntityType], id)
			}
		}
	}

	// Build the grouped list
	for etype, ids := range affected.AffectedTypes {
		affected.AffectedByType = append(affected.AffectedByType, AffectedGroup{
			EntityType: etype,
			EntityIDs:  ids,
		})
	}

	if len(affected.AffectedByType) == 0 {
		affected.Summary = fmt.Sprintf("Change %q has no impact on tracked entities", ev.ChangeType)
	} else {
		affected.Summary = fmt.Sprintf("Change %q affects %d entity types", ev.ChangeType, len(affected.AffectedByType))
	}

	return affected, nil
}

// invalidate marks affected entities with the appropriate status.
func (s *Service) invalidate(ev *ChangeEvent, impact *ImpactAnalysis) ([]EntityState, error) {
	var states []EntityState

	for _, group := range impact.AffectedByType {
		rule := s.findRule(group.EntityType, ev.ChangeType)
		if rule == nil {
			continue
		}

		for _, eid := range group.EntityIDs {
			state := EntityState{
				ID:           domain.NewID("est"),
				EntityType:   group.EntityType,
				EntityID:     eid,
				Status:       rule.ResultStatus,
				LastChangeID: ev.ID,
				Reason:       fmt.Sprintf("Change %s (%s) -> %s", ev.ChangeType, ev.Summary, rule.ResultStatus),
				UpdatedAt:    time.Now().UTC(),
			}
			if err := s.store.SaveEntityState(&state); err != nil {
				return nil, fmt.Errorf("save entity state: %w", err)
			}
			states = append(states, state)
		}
	}

	return states, nil
}

func (s *Service) findRule(entityType string, changeType ChangeType) *InvalidationRule {
	for _, rule := range s.rules {
		if rule.EntityType != entityType {
			continue
		}
		for _, ct := range rule.AffectedBy {
			if ct == changeType {
				return &rule
			}
		}
	}
	return nil
}

// CreateSnapshot creates a point-in-time snapshot, optionally linked to a change event.
func (s *Service) CreateSnapshot(reason string) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	snap := &Snapshot{
		ID:        domain.NewID("snap"),
		Timestamp: now,
		Reason:    reason,
		CreatedAt: now,
	}
	if err := s.snaps.Save(snap); err != nil {
		return "", fmt.Errorf("save snapshot: %w", err)
	}
	return snap.ID, nil
}

// ListChanges returns recent change events for a project.
func (s *Service) ListChanges(projectID string, limit int) ([]ChangeEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.store.ListEvents(projectID, limit)
}

// GetChange returns a single change event by ID.
func (s *Service) GetChange(changeID string) (*ChangeEvent, error) {
	return s.store.GetEvent(changeID)
}

// GetReport returns the full report for a change event.
func (s *Service) GetReport(changeID string) (*ChangeReport, error) {
	ev, err := s.store.GetEvent(changeID)
	if err != nil {
		return nil, err
	}
	impact, err := s.store.GetImpact(changeID)
	if err != nil {
		return nil, err
	}
	return &ChangeReport{
		Change: *ev,
		Impact: impact,
	}, nil
}

// InvalidateEntities re-computes and persists invalidation for an existing change.
func (s *Service) InvalidateEntities(changeID string) ([]EntityState, error) {
	ev, err := s.store.GetEvent(changeID)
	if err != nil {
		return nil, err
	}
	impact, err := s.store.GetImpact(changeID)
	if err != nil {
		return nil, err
	}
	return s.invalidate(ev, impact)
}

// GetEntityState returns the current invalidation state for an entity.
func (s *Service) GetEntityState(entityType, entityID string) (*EntityState, error) {
	return s.store.GetEntityState(entityType, entityID)
}

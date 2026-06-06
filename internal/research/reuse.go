package research

import (
	"fmt"
	"time"
)

// ReuseService makes approved research reusable. It provides:
//
//   - FindReusable — search for existing approved research by topic
//   - IsReusable — check if a research entry is approved and can be reused
//   - PromoteToKnowledge — promote approved research into a reusable knowledge object
//   - IncrementReuseCount — track how many times research has been reused
//
// This implements Phase 7 (Research Reuse).
type ReuseService struct {
	repo Repository
	now  func() time.Time
}

// NewReuseService creates a ReuseService backed by the given repository.
func NewReuseService(repo Repository) *ReuseService {
	return &ReuseService{repo: repo, now: time.Now}
}

// FindReusable searches for approved research entries matching the topic.
func (s *ReuseService) FindReusable(projectID string, topic string) ([]ResearchEntry, error) {
	all, err := s.repo.SearchEntries(topic)
	if err != nil {
		return nil, err
	}
	var reusable []ResearchEntry
	for _, entry := range all {
		if entry.ProjectID == projectID && entry.Status == ResearchStatusApproved {
			reusable = append(reusable, entry)
		}
	}
	return reusable, nil
}

// IncrementReuseCount increments the reuse counter for a research entry
// and updates the last-reused timestamp. Only applies to approved entries.
func (s *ReuseService) IncrementReuseCount(id string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository for reuse tracking")
	}
	return s.repo.IncrementReuseCount(id)
}

// EnsureFTS creates the FTS5 triggers for research_entries_fts if they
// don't already exist. Idempotent.
func (s *ReuseService) EnsureFTS() error {
	if s.repo == nil {
		return fmt.Errorf("no repository for FTS")
	}
	return s.repo.EnsureFTS()
}

// PromoteToKnowledge creates a knowledge object from an approved research
// entry. Delegates to the repository.
func (s *ReuseService) PromoteToKnowledge(researchID string) (knowledgeID string, err error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository for knowledge promotion")
	}
	return s.repo.PromoteToKnowledge(researchID)
}

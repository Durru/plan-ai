package research

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
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
	db          *sql.DB
	researchRepo Repository
	now         func() time.Time
}

// NewReuseService creates a ReuseService backed by the given database.
func NewReuseService(db *sql.DB, researchRepo Repository) *ReuseService {
	return &ReuseService{db: db, researchRepo: researchRepo, now: time.Now}
}

// FindReusable searches for approved research entries matching the topic.
func (s *ReuseService) FindReusable(projectID string, topic string) ([]ResearchEntry, error) {
	if s.researchRepo != nil {
		all, err := s.researchRepo.SearchEntries(topic)
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

	if s.db == nil {
		return nil, nil
	}

	like := "%" + strings.ToLower(strings.TrimSpace(topic)) + "%"
	rows, err := s.db.Query(`SELECT id, project_id, topic, summary, status, confidence, created_at, updated_at FROM research_entries WHERE project_id = ? AND status = ? AND (LOWER(topic) LIKE ? OR LOWER(summary) LIKE ?) ORDER BY COALESCE(reuse_count, 0) DESC, created_at DESC`,
		projectID, string(ResearchStatusApproved), like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ResearchEntry
	for rows.Next() {
		var e ResearchEntry
		var c, u string
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Topic, &e.Summary, &e.Status, &e.Confidence, &c, &u); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// IncrementReuseCount increments the reuse counter for a research entry
// and updates the last-reused timestamp. Only applies to approved entries.
func (s *ReuseService) IncrementReuseCount(id string) error {
	if s.db == nil {
		return fmt.Errorf("no database for reuse tracking")
	}
	now := s.now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`UPDATE research_entries SET reuse_count = COALESCE(reuse_count, 0) + 1, reused_at = ? WHERE id = ? AND status = ?`, now, id, string(ResearchStatusApproved))
	return err
}

// EnsureFTS creates the FTS5 triggers for research_entries_fts if they
// don't already exist. Idempotent.
func (s *ReuseService) EnsureFTS() error {
	if s.db == nil {
		return fmt.Errorf("no database for FTS")
	}
	_, err := s.db.Exec(researchFTSTriggers)
	return err
}

// ── knowledge promotion ──

// PromoteToKnowledge creates a knowledge object from an approved research
// entry. The knowledge object inherits the research topic (as title),
// summary, and confidence. A research-knowledge link is also created.
//
// Returns the knowledge object ID, or an error if the research entry is
// not approved.
func PromoteToKnowledge(db *sql.DB, researchID string) (knowledgeID string, err error) {
	var topic, summary string
	var confidence int
	var status, projectID string

	err = db.QueryRow(`SELECT project_id, topic, summary, confidence, status FROM research_entries WHERE id = ?`, researchID).
		Scan(&projectID, &topic, &summary, &confidence, &status)
	if err != nil {
		return "", fmt.Errorf("find research %s: %w", researchID, err)
	}
	if status != string(ResearchStatusApproved) {
		return "", fmt.Errorf("research %s is not approved (status: %s)", researchID, status)
	}

	category := Classify(topic)

	knowledgeID = domain.NewID("knowledge")
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := db.Exec(`INSERT INTO knowledge_objects (id, project_id, topic, category, summary, confidence, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		knowledgeID, projectID, topic, string(category), summary, float64(confidence), "approved", now, now); err != nil {
		return "", fmt.Errorf("create knowledge: %w", err)
	}

	// Link research to knowledge
	linkID := domain.NewID("rklink")
	if _, err := db.Exec(`INSERT INTO research_knowledge_links (id, research_id, knowledge_id, created_at) VALUES (?, ?, ?, ?)`,
		linkID, researchID, knowledgeID, now); err != nil {
		return knowledgeID, fmt.Errorf("link research->knowledge: %w", err)
	}

	return knowledgeID, nil
}

// sanitizeFTS5 escapes FTS5 special characters in a query string.
func sanitizeFTS5(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	q = strings.ReplaceAll(q, `"`, `""`)
	return `"` + q + `"`
}

// researchFTSTriggers creates the triggers that keep research_entries_fts
// in sync with the research_entries table. The virtual table is declared
// in the store's schema, but the Phase 7 triggers populate it.
const researchFTSTriggers = `
INSERT OR IGNORE INTO research_entries_fts(research_entries_fts) VALUES('rebuild');

CREATE TRIGGER IF NOT EXISTS research_entries_fts_ai AFTER INSERT ON research_entries BEGIN
  INSERT INTO research_entries_fts(rowid, id, topic, objective, summary)
  VALUES (new.rowid, new.id, new.topic, new.topic, new.summary);
END;

CREATE TRIGGER IF NOT EXISTS research_entries_fts_ad AFTER DELETE ON research_entries BEGIN
  DELETE FROM research_entries_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS research_entries_fts_au AFTER UPDATE ON research_entries BEGIN
  DELETE FROM research_entries_fts WHERE rowid = old.rowid;
  INSERT INTO research_entries_fts(rowid, id, topic, objective, summary)
  VALUES (new.rowid, new.id, new.topic, new.topic, new.summary);
END;
`

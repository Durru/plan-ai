package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/research"
)

// ResearchRepository persists ResearchEntries, findings, sources,
// conclusions, tags, and knowledge links in the project store.
// It implements the research.Repository contract.
type ResearchRepository struct{ db *sql.DB }

// NewResearchRepository creates a new ResearchRepository.
func NewResearchRepository(db *sql.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

// Compile-time assertion that the concrete repository satisfies the
// research package contract.
var _ research.Repository = (*ResearchRepository)(nil)

// researchEntryColumns includes the legacy source and conclusion columns
// to satisfy NOT NULL constraints from migration 0003. The new code uses
// category, status, and the research sub-tables instead.
const researchEntryColumns = `id, project_id, topic, source, category, summary, conclusion, status, confidence, created_at, updated_at`

func scanResearchEntry(row interface {
	Scan(dest ...any) error
}) (research.ResearchEntry, error) {
	var entry research.ResearchEntry
	var createdAt, updatedAt string
	var category, status string
	var confidence float64
	var legacySource, legacyConclusion string // consumed but not used
	if err := row.Scan(&entry.ID, &entry.ProjectID, &entry.Topic, &legacySource, &category, &entry.Summary, &legacyConclusion, &status, &confidence, &createdAt, &updatedAt); err != nil {
		return entry, err
	}
	entry.Category = research.ResearchCategory(category)
	entry.Status = research.ResearchStatus(status)
	entry.Confidence = int(confidence)
	entry.CreatedAt = parseTime(createdAt)
	entry.UpdatedAt = parseTime(updatedAt)
	return entry, nil
}

func scanResearchFinding(row interface {
	Scan(dest ...any) error
}) (research.ResearchFinding, error) {
	var f research.ResearchFinding
	var createdAt string
	if err := row.Scan(&f.ID, &f.ResearchID, &f.Title, &f.Content, &f.Importance, &createdAt); err != nil {
		return f, err
	}
	f.CreatedAt = parseTime(createdAt)
	return f, nil
}

func scanResearchSource(row interface {
	Scan(dest ...any) error
}) (research.ResearchSource, error) {
	var s research.ResearchSource
	var createdAt, sourceType string
	if err := row.Scan(&s.ID, &s.ResearchID, &s.Title, &s.URL, &sourceType, &createdAt); err != nil {
		return s, err
	}
	s.SourceType = research.ResearchSourceType(sourceType)
	s.CreatedAt = parseTime(createdAt)
	return s, nil
}

func scanResearchConclusion(row interface {
	Scan(dest ...any) error
}) (research.ResearchConclusion, error) {
	var c research.ResearchConclusion
	var createdAt string
	if err := row.Scan(&c.ID, &c.ResearchID, &c.Content, &c.Confidence, &createdAt); err != nil {
		return c, err
	}
	c.CreatedAt = parseTime(createdAt)
	return c, nil
}

func scanResearchTag(row interface {
	Scan(dest ...any) error
}) (research.ResearchTag, error) {
	var t research.ResearchTag
	if err := row.Scan(&t.ID, &t.ResearchID, &t.Tag); err != nil {
		return t, err
	}
	return t, nil
}

func scanResearchKnowledgeLink(row interface {
	Scan(dest ...any) error
}) (research.ResearchKnowledgeLink, error) {
	var l research.ResearchKnowledgeLink
	var createdAt string
	if err := row.Scan(&l.ID, &l.ResearchID, &l.KnowledgeID, &createdAt); err != nil {
		return l, err
	}
	l.CreatedAt = parseTime(createdAt)
	return l, nil
}

// CreateEntry persists a new ResearchEntry.
func (r *ResearchRepository) CreateEntry(entry research.ResearchEntry) error {
	entry.ID = ensureID(entry.ID, "research")
	if entry.Category == "" {
		entry.Category = research.CategoryGeneral
	}
	if entry.Status == "" {
		entry.Status = research.ResearchStatusDraft
	}
	createdAt, updatedAt := ensureTimestamps(entry.CreatedAt, entry.UpdatedAt)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(fmt.Sprintf(`INSERT INTO research_entries (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, researchEntryColumns),
		entry.ID, entry.ProjectID, entry.Topic, "", string(entry.Category), entry.Summary,
		"", string(entry.Status), float64(entry.Confidence), createdAt, updatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO research_jobs (id, project_id, topic, summary, confidence, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, topic=excluded.topic, summary=excluded.summary, confidence=excluded.confidence, status=excluded.status`,
		entry.ID, entry.ProjectID, entry.Topic, entry.Summary, float64(entry.Confidence)/100, string(entry.Status), createdAt); err != nil {
		return err
	}
	return tx.Commit()
}

// GetEntry returns a ResearchEntry by ID.
func (r *ResearchRepository) GetEntry(id string) (research.ResearchEntry, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM research_entries WHERE id = ?`, researchEntryColumns), id)
	return scanResearchEntry(row)
}

// ListEntries returns all ResearchEntries ordered by creation time.
func (r *ResearchRepository) ListEntries() ([]research.ResearchEntry, error) {
	return r.queryEntries(fmt.Sprintf(`SELECT %s FROM research_entries ORDER BY created_at, id`, researchEntryColumns))
}

// SearchEntries searches research entries by topic and summary.
func (r *ResearchRepository) SearchEntries(query string) ([]research.ResearchEntry, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	return r.queryEntries(fmt.Sprintf(`SELECT %s FROM research_entries
WHERE LOWER(topic) LIKE ? OR LOWER(summary) LIKE ?
ORDER BY created_at, id`, researchEntryColumns), like, like)
}

// UpdateEntryStatus updates the status of a ResearchEntry.
func (r *ResearchRepository) UpdateEntryStatus(id string, status research.ResearchStatus) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE research_entries SET status = ?, updated_at = ? WHERE id = ?`, string(status), now, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE research_jobs SET status = ? WHERE id = ?`, string(status), id); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteEntry deletes a ResearchEntry and all associated data.
func (r *ResearchRepository) DeleteEntry(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, table := range []string{"research_findings", "research_sources", "research_conclusions", "research_tags", "research_knowledge_links", "research_recommendations"} {
		if _, err := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE research_id = ?`, table), id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM research_jobs WHERE id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM knowledge_links WHERE link_type = 'research' AND target_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM research_entries WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateFinding persists a finding.
func (r *ResearchRepository) CreateFinding(finding research.ResearchFinding) error {
	finding.ID = ensureID(finding.ID, "finding")
	createdAt, _ := ensureTimestamps(finding.CreatedAt, time.Time{})
	_, err := r.db.Exec(`INSERT INTO research_findings (id, research_id, title, content, importance, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, finding.ID, finding.ResearchID, finding.Title, finding.Content, finding.Importance, createdAt)
	return err
}

// ListFindings returns all findings for a research entry.
func (r *ResearchRepository) ListFindings(researchID string) ([]research.ResearchFinding, error) {
	rows, err := r.db.Query(`SELECT id, research_id, title, content, importance, created_at FROM research_findings
WHERE research_id = ? ORDER BY importance DESC, created_at, id`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var findings []research.ResearchFinding
	for rows.Next() {
		f, err := scanResearchFinding(rows)
		if err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

// CreateSource persists a source.
func (r *ResearchRepository) CreateSource(source research.ResearchSource) error {
	source.ID = ensureID(source.ID, "source")
	createdAt, _ := ensureTimestamps(source.CreatedAt, time.Time{})
	_, err := r.db.Exec(`INSERT INTO research_sources (id, research_id, title, url, source_type, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, source.ID, source.ResearchID, source.Title, source.URL, string(source.SourceType), createdAt)
	return err
}

// ListSources returns all sources for a research entry.
func (r *ResearchRepository) ListSources(researchID string) ([]research.ResearchSource, error) {
	rows, err := r.db.Query(`SELECT id, research_id, title, url, source_type, created_at FROM research_sources
WHERE research_id = ? ORDER BY created_at, id`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []research.ResearchSource
	for rows.Next() {
		s, err := scanResearchSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

// CreateConclusion persists a conclusion.
func (r *ResearchRepository) CreateConclusion(conclusion research.ResearchConclusion) error {
	conclusion.ID = ensureID(conclusion.ID, "conclusion")
	createdAt, _ := ensureTimestamps(conclusion.CreatedAt, time.Time{})
	_, err := r.db.Exec(`INSERT INTO research_conclusions (id, research_id, content, confidence, created_at)
VALUES (?, ?, ?, ?, ?)`, conclusion.ID, conclusion.ResearchID, conclusion.Content, conclusion.Confidence, createdAt)
	return err
}

// ListConclusions returns all conclusions for a research entry.
func (r *ResearchRepository) ListConclusions(researchID string) ([]research.ResearchConclusion, error) {
	rows, err := r.db.Query(`SELECT id, research_id, content, confidence, created_at FROM research_conclusions
WHERE research_id = ? ORDER BY confidence DESC, created_at, id`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conclusions []research.ResearchConclusion
	for rows.Next() {
		c, err := scanResearchConclusion(rows)
		if err != nil {
			return nil, err
		}
		conclusions = append(conclusions, c)
	}
	return conclusions, rows.Err()
}

// AddTag adds a tag to a research entry. Duplicates are silently ignored.
func (r *ResearchRepository) AddTag(researchID, tag string) error {
	_, err := r.db.Exec(`INSERT INTO research_tags (id, research_id, tag)
VALUES (?, ?, ?)
ON CONFLICT(research_id, tag) DO NOTHING`, ensureID("", "rtag"), researchID, tag)
	return err
}

// ListTags returns all tags for a research entry.
func (r *ResearchRepository) ListTags(researchID string) ([]research.ResearchTag, error) {
	rows, err := r.db.Query(`SELECT id, research_id, tag FROM research_tags WHERE research_id = ? ORDER BY tag`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []research.ResearchTag
	for rows.Next() {
		t, err := scanResearchTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// LinkKnowledge links a research entry to a knowledge object. Duplicates are silently ignored.
func (r *ResearchRepository) LinkKnowledge(researchID, knowledgeID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO research_knowledge_links (id, research_id, knowledge_id, created_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(research_id, knowledge_id) DO NOTHING`, ensureID("", "rlink"), researchID, knowledgeID, now); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO knowledge_links (id, knowledge_id, link_type, target_id, created_at)
VALUES (?, ?, 'research', ?, ?)
ON CONFLICT(knowledge_id, link_type, target_id) DO NOTHING`, ensureID("", "klink"), knowledgeID, researchID, now); err != nil {
		return err
	}
	return tx.Commit()
}

// ListKnowledgeLinks returns all knowledge links for a research entry.
func (r *ResearchRepository) ListKnowledgeLinks(researchID string) ([]research.ResearchKnowledgeLink, error) {
	rows, err := r.db.Query(`SELECT id, research_id, knowledge_id, created_at FROM research_knowledge_links
WHERE research_id = ? ORDER BY created_at, id`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []research.ResearchKnowledgeLink
	for rows.Next() {
		l, err := scanResearchKnowledgeLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// Summary returns aggregate counts for the status command.
func (r *ResearchRepository) Summary() (research.ResearchSummary, error) {
	var summary research.ResearchSummary

	rows, err := r.db.Query(`SELECT status, COUNT(*) FROM research_entries GROUP BY status`)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return summary, err
		}
		summary.Total += count
		switch research.ResearchStatus(status) {
		case research.ResearchStatusDraft:
			summary.Draft += count
		case research.ResearchStatusInReview:
			summary.InReview += count
		case research.ResearchStatusApproved:
			summary.Approved += count
		case research.ResearchStatusRejected:
			summary.Rejected += count
		case research.ResearchStatusArchived:
			summary.Archived += count
		}
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM research_findings`).Scan(&summary.Findings); err != nil {
		return summary, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM research_sources`).Scan(&summary.Sources); err != nil {
		return summary, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM research_conclusions`).Scan(&summary.Conclusions); err != nil {
		return summary, err
	}

	return summary, nil
}

func (r *ResearchRepository) queryEntries(query string, args ...any) ([]research.ResearchEntry, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []research.ResearchEntry
	for rows.Next() {
		entry, err := scanResearchEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

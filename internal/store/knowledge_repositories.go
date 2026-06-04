package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/knowledge"
)

// KnowledgeRepository persists KnowledgeObjects, tags, relations, and
// references in the project store. The public surface is shaped by the
// knowledge.Repository contract.
type KnowledgeRepository struct{ db *sql.DB }

func NewKnowledgeRepository(db *sql.DB) KnowledgeRepository { return KnowledgeRepository{db: db} }

// Compile-time assertion that the concrete repository satisfies the
// knowledge package contract. If a method is missing or its signature
// drifts, the build fails here instead of at the call site.
var _ knowledge.Repository = KnowledgeRepository{}

// knowledgeColumns lists every column of knowledge_objects in the
// canonical insertion / projection order.
const knowledgeColumns = `id, topic, category, summary, content, confidence, source_type, reuse_count, status, created_at, updated_at`

func scanKnowledgeObject(row interface {
	Scan(dest ...any) error
}) (domain.KnowledgeObject, error) {
	var object domain.KnowledgeObject
	var createdAt, updatedAt string
	var category, status, sourceType string
	if err := row.Scan(&object.ID, &object.Topic, &category, &object.Summary, &object.Content, &object.Confidence, &sourceType, &object.ReuseCount, &status, &createdAt, &updatedAt); err != nil {
		return object, err
	}
	object.Category = domain.KnowledgeCategory(category)
	object.Status = domain.KnowledgeStatus(status)
	object.SourceType = domain.KnowledgeSourceType(sourceType)
	object.CreatedAt = parseTime(createdAt)
	object.UpdatedAt = parseTime(updatedAt)
	return object, nil
}

func (r KnowledgeRepository) Create(object domain.KnowledgeObject) error {
	object.ID = ensureID(object.ID, "knowledge")
	if object.Category == "" {
		object.Category = domain.KnowledgeCategoryGeneral
	}
	if object.Status == "" {
		object.Status = domain.KnowledgeStatusDraft
	}
	if object.SourceType == "" {
		object.SourceType = domain.KnowledgeSourceManual
	}
	createdAt, updatedAt := ensureTimestamps(object.CreatedAt, object.UpdatedAt)
	_, err := r.db.Exec(fmt.Sprintf(`INSERT INTO knowledge_objects (%s)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, knowledgeColumns),
		object.ID, object.Topic, object.Category, object.Summary, object.Content,
		object.Confidence, object.SourceType, object.ReuseCount, object.Status,
		createdAt, updatedAt)
	return err
}

func (r KnowledgeRepository) Update(object domain.KnowledgeObject) error {
	now := time.Now().UTC().Format(time.RFC3339)
	object.UpdatedAt = parseTime(now)
	_, err := r.db.Exec(`UPDATE knowledge_objects
SET topic = ?, category = ?, summary = ?, content = ?, confidence = ?, source_type = ?, reuse_count = ?, status = ?, updated_at = ?
WHERE id = ?`,
		object.Topic, object.Category, object.Summary, object.Content, object.Confidence,
		object.SourceType, object.ReuseCount, object.Status, now, object.ID)
	return err
}

func (r KnowledgeRepository) GetByID(id string) (domain.KnowledgeObject, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM knowledge_objects WHERE id = ?`, knowledgeColumns), id)
	return scanKnowledgeObject(row)
}

func (r KnowledgeRepository) GetByTopic(topic string) (domain.KnowledgeObject, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM knowledge_objects WHERE topic = ? ORDER BY created_at, id LIMIT 1`, knowledgeColumns), topic)
	return scanKnowledgeObject(row)
}

func (r KnowledgeRepository) List() ([]domain.KnowledgeObject, error) {
	return r.queryObjects(fmt.Sprintf(`SELECT %s FROM knowledge_objects ORDER BY created_at, id`, knowledgeColumns))
}

func (r KnowledgeRepository) ListByCategory(category domain.KnowledgeCategory) ([]domain.KnowledgeObject, error) {
	return r.queryObjects(fmt.Sprintf(`SELECT %s FROM knowledge_objects WHERE category = ? ORDER BY created_at, id`, knowledgeColumns), string(category))
}

func (r KnowledgeRepository) Search(query string) ([]domain.KnowledgeObject, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	return r.queryObjects(fmt.Sprintf(`SELECT %s FROM knowledge_objects
WHERE LOWER(topic) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(content) LIKE ?
ORDER BY created_at, id`, knowledgeColumns), like, like, like)
}

func (r KnowledgeRepository) queryObjects(query string, args ...any) ([]domain.KnowledgeObject, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var objects []domain.KnowledgeObject
	for rows.Next() {
		object, err := scanKnowledgeObject(rows)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

func (r KnowledgeRepository) IncrementReuseCount(id string) (domain.KnowledgeObject, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := r.db.Exec(`UPDATE knowledge_objects SET reuse_count = reuse_count + 1, updated_at = ? WHERE id = ?`, now, id); err != nil {
		return domain.KnowledgeObject{}, err
	}
	return r.GetByID(id)
}

func (r KnowledgeRepository) AddTag(knowledgeID, tag string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO knowledge_tags (id, knowledge_id, tag, created_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(knowledge_id, tag) DO NOTHING`, domain.NewID("tag"), knowledgeID, tag, now)
	return err
}

func (r KnowledgeRepository) ListTags(knowledgeID string) ([]knowledge.Tag, error) {
	rows, err := r.db.Query(`SELECT id, knowledge_id, tag, created_at FROM knowledge_tags WHERE knowledge_id = ? ORDER BY tag`, knowledgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []knowledge.Tag
	for rows.Next() {
		var tag knowledge.Tag
		if err := rows.Scan(&tag.ID, &tag.KnowledgeID, &tag.Tag, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r KnowledgeRepository) AddRelation(sourceID, targetID string, relationType domain.KnowledgeRelationType) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO knowledge_relations (id, source_id, target_id, relation_type, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(source_id, target_id, relation_type) DO NOTHING`, domain.NewID("rel"), sourceID, targetID, string(relationType), now)
	return err
}

func (r KnowledgeRepository) ListRelations(knowledgeID string) ([]knowledge.Relation, error) {
	rows, err := r.db.Query(`SELECT id, source_id, target_id, relation_type, created_at
FROM knowledge_relations
WHERE source_id = ? OR target_id = ?
ORDER BY created_at, id`, knowledgeID, knowledgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var relations []knowledge.Relation
	for rows.Next() {
		var relation knowledge.Relation
		var relationType string
		if err := rows.Scan(&relation.ID, &relation.SourceID, &relation.TargetID, &relationType, &relation.CreatedAt); err != nil {
			return nil, err
		}
		relation.RelationType = domain.KnowledgeRelationType(relationType)
		relations = append(relations, relation)
	}
	return relations, rows.Err()
}

func (r KnowledgeRepository) AddReference(knowledgeID string, referenceType domain.KnowledgeReferenceType, referenceID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`INSERT INTO knowledge_references (id, knowledge_id, reference_type, reference_id, created_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(knowledge_id, reference_type, reference_id) DO NOTHING`, domain.NewID("ref"), knowledgeID, string(referenceType), referenceID, now)
	return err
}

func (r KnowledgeRepository) ListReferences(knowledgeID string) ([]knowledge.Reference, error) {
	rows, err := r.db.Query(`SELECT id, knowledge_id, reference_type, reference_id, created_at
FROM knowledge_references
WHERE knowledge_id = ?
ORDER BY reference_type, reference_id`, knowledgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var references []knowledge.Reference
	for rows.Next() {
		var reference knowledge.Reference
		var referenceType string
		if err := rows.Scan(&reference.ID, &reference.KnowledgeID, &referenceType, &reference.ReferenceID, &reference.CreatedAt); err != nil {
			return nil, err
		}
		reference.ReferenceType = domain.KnowledgeReferenceType(referenceType)
		references = append(references, reference)
	}
	return references, rows.Err()
}

func (r KnowledgeRepository) Summary() (knowledge.Summary, error) {
	rows, err := r.db.Query(`SELECT status, COUNT(*), COALESCE(SUM(reuse_count), 0) FROM knowledge_objects GROUP BY status`)
	if err != nil {
		return knowledge.Summary{}, err
	}
	defer rows.Close()
	var summary knowledge.Summary
	for rows.Next() {
		var status string
		var count, reuse int
		if err := rows.Scan(&status, &count, &reuse); err != nil {
			return summary, err
		}
		summary.Total += count
		summary.Reused += reuse
		switch domain.KnowledgeStatus(status) {
		case domain.KnowledgeStatusDraft:
			summary.Draft += count
		case domain.KnowledgeStatusReviewed:
			summary.Reviewed += count
		case domain.KnowledgeStatusApproved:
			summary.Approved += count
		case domain.KnowledgeStatusArchived:
			summary.Archived += count
		}
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}
	// Reused can overlap with the status counts, so it does not change
	// the per-status totals above; it is reported separately.
	return summary, nil
}

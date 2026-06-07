package repositories

import (
	"database/sql"
	"github.com/Durru/plan-ai/internal/domain"
	"strings"
)

type KnowledgeRepository struct{ db *sql.DB }

func NewKnowledgeRepository(db *sql.DB) KnowledgeRepository { return KnowledgeRepository{db: db} }

var _ domain.KnowledgeRepository = KnowledgeRepository{}

func (r KnowledgeRepository) Save(x domain.KnowledgeObject) error   { return r.upsert(x) }
func (r KnowledgeRepository) Update(x domain.KnowledgeObject) error { return r.upsert(x) }
func (r KnowledgeRepository) upsert(x domain.KnowledgeObject) error {
	x.ID = ensureID(x.ID, "knowledge")
	if x.Category == "" {
		x.Category = domain.KnowledgeCategoryGeneral
	}
	if x.Type == "" {
		x.Type = domain.KnowledgeTypeReference
	}
	if x.SourceType == "" {
		x.SourceType = domain.KnowledgeSourceManual
	}
	if x.Status == "" {
		x.Status = domain.KnowledgeStatusDraft
	}
	c, u := times(x.CreatedAt, x.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO knowledge_objects (id, topic, category, type, summary, content, confidence, source_type, reuse_count, status, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET topic=excluded.topic,category=excluded.category,type=excluded.type,summary=excluded.summary,content=excluded.content,confidence=excluded.confidence,source_type=excluded.source_type,reuse_count=excluded.reuse_count,status=excluded.status,updated_at=excluded.updated_at`, x.ID, x.Topic, x.Category, x.Type, x.Summary, x.Content, x.Confidence, x.SourceType, x.ReuseCount, x.Status, c, u)
	return err
}
func (r KnowledgeRepository) GetByID(id string) (domain.KnowledgeObject, error) {
	xs, err := r.list(`WHERE id=?`, id)
	if err != nil {
		return domain.KnowledgeObject{}, err
	}
	if len(xs) == 0 {
		return domain.KnowledgeObject{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r KnowledgeRepository) GetByTopic(topic string) (domain.KnowledgeObject, error) {
	xs, err := r.list(`WHERE topic=? ORDER BY created_at, id`, topic)
	if err != nil {
		return domain.KnowledgeObject{}, err
	}
	if len(xs) == 0 {
		return domain.KnowledgeObject{}, sql.ErrNoRows
	}
	return xs[0], nil
}
func (r KnowledgeRepository) List() ([]domain.KnowledgeObject, error) {
	return r.list(`ORDER BY created_at, id`)
}
func (r KnowledgeRepository) ListByCategory(c domain.KnowledgeCategory) ([]domain.KnowledgeObject, error) {
	return r.list(`WHERE category=? ORDER BY created_at, id`, c)
}
func (r KnowledgeRepository) Search(q string) ([]domain.KnowledgeObject, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
	return r.list(`WHERE LOWER(topic) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(content) LIKE ? ORDER BY created_at, id`, like, like, like)
}
func (r KnowledgeRepository) list(where string, args ...any) ([]domain.KnowledgeObject, error) {
	rows, err := r.db.Query(`SELECT id, topic, category, type, summary, content, confidence, source_type, reuse_count, status, created_at, updated_at FROM knowledge_objects `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.KnowledgeObject
	for rows.Next() {
		var x domain.KnowledgeObject
		var cat, typ, src, st, c, u string
		if err := rows.Scan(&x.ID, &x.Topic, &cat, &typ, &x.Summary, &x.Content, &x.Confidence, &src, &x.ReuseCount, &st, &c, &u); err != nil {
			return nil, err
		}
		x.Category = domain.KnowledgeCategory(cat)
		x.Type = domain.KnowledgeType(typ)
		x.SourceType = domain.KnowledgeSourceType(src)
		x.Status = domain.KnowledgeStatus(st)
		x.CreatedAt = parse(c)
		x.UpdatedAt = parse(u)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r KnowledgeRepository) IncrementReuseCount(id string) (domain.KnowledgeObject, error) {
	_, err := r.db.Exec(`UPDATE knowledge_objects SET reuse_count=reuse_count+1, updated_at=? WHERE id=?`, now(), id)
	if err != nil {
		return domain.KnowledgeObject{}, err
	}
	return r.GetByID(id)
}
func (r KnowledgeRepository) AddTag(id, tag string) error {
	_, err := r.db.Exec(`INSERT INTO knowledge_tags (id, knowledge_id, tag, created_at) VALUES (?,?,?,?) ON CONFLICT(knowledge_id, tag) DO NOTHING`, ensureID("", "ktag"), id, tag, now())
	return err
}
func (r KnowledgeRepository) ListTags(id string) ([]domain.KnowledgeTag, error) {
	rows, err := r.db.Query(`SELECT id, knowledge_id, tag, created_at FROM knowledge_tags WHERE knowledge_id=? ORDER BY tag`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.KnowledgeTag
	for rows.Next() {
		var x domain.KnowledgeTag
		var c string
		if err := rows.Scan(&x.ID, &x.KnowledgeID, &x.Tag, &c); err != nil {
			return nil, err
		}
		x.CreatedAt = parse(c)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r KnowledgeRepository) AddRelation(s, t string, typ domain.KnowledgeRelationType) error {
	_, err := r.db.Exec(`INSERT INTO knowledge_relations (id, source_id, target_id, relation_type, created_at) VALUES (?,?,?,?,?) ON CONFLICT(source_id,target_id,relation_type) DO NOTHING`, ensureID("", "krel"), s, t, typ, now())
	return err
}
func (r KnowledgeRepository) ListRelations(id string) ([]domain.KnowledgeRelation, error) {
	rows, err := r.db.Query(`SELECT id, source_id, target_id, relation_type, created_at FROM knowledge_relations WHERE source_id=? OR target_id=?`, id, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.KnowledgeRelation
	for rows.Next() {
		var x domain.KnowledgeRelation
		var typ, c string
		if err := rows.Scan(&x.ID, &x.SourceID, &x.TargetID, &typ, &c); err != nil {
			return nil, err
		}
		x.RelationType = domain.KnowledgeRelationType(typ)
		x.CreatedAt = parse(c)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r KnowledgeRepository) AddReference(id string, typ domain.KnowledgeReferenceType, ref string) error {
	_, err := r.db.Exec(`INSERT INTO knowledge_references (id, knowledge_id, reference_type, reference_id, created_at) VALUES (?,?,?,?,?) ON CONFLICT(knowledge_id,reference_type,reference_id) DO NOTHING`, ensureID("", "kref"), id, typ, ref, now())
	return err
}
func (r KnowledgeRepository) ListReferences(id string) ([]domain.KnowledgeReference, error) {
	rows, err := r.db.Query(`SELECT id, knowledge_id, reference_type, reference_id, created_at FROM knowledge_references WHERE knowledge_id=?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.KnowledgeReference
	for rows.Next() {
		var x domain.KnowledgeReference
		var typ, c string
		if err := rows.Scan(&x.ID, &x.KnowledgeID, &typ, &x.ReferenceID, &c); err != nil {
			return nil, err
		}
		x.ReferenceType = domain.KnowledgeReferenceType(typ)
		x.CreatedAt = parse(c)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r KnowledgeRepository) Delete(id string) error {
	return deleteByID(r.db, "knowledge_objects", id)
}

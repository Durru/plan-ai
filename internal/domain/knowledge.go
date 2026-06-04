package domain

import "time"

// KnowledgeType classifies the nature of a KnowledgeObject.
type KnowledgeType string

const (
	KnowledgeTypeResearch    KnowledgeType = "research"
	KnowledgeTypeDecision    KnowledgeType = "decision"
	KnowledgeTypeRequirement KnowledgeType = "requirement"
	KnowledgeTypeConstraint  KnowledgeType = "constraint"
	KnowledgeTypeReference   KnowledgeType = "reference"
	KnowledgeTypePattern     KnowledgeType = "pattern"
)

// KnowledgeTag is a single label attached to a KnowledgeObject.
type KnowledgeTag struct {
	ID          string
	KnowledgeID string
	Tag         string
	CreatedAt   time.Time
}

// KnowledgeRelation is a directed typed link between two KnowledgeObjects.
type KnowledgeRelation struct {
	ID           string
	SourceID     string
	TargetID     string
	RelationType KnowledgeRelationType
	CreatedAt    time.Time
}

// KnowledgeReference is a link from a KnowledgeObject to an external
// entity (plan, decision, research, technology, etc.).
type KnowledgeReference struct {
	ID            string
	KnowledgeID   string
	ReferenceType KnowledgeReferenceType
	ReferenceID   string
	CreatedAt     time.Time
}

// KnowledgeObject is the central reusable knowledge entity. It captures
// distilled project knowledge that has been curated, approved, and is
// ready for reuse across planning and implementation phases.
type KnowledgeObject struct {
	ID         string
	Topic      string
	Category   KnowledgeCategory
	Type       KnowledgeType
	Summary    string
	Content    string
	Confidence float64
	SourceType KnowledgeSourceType
	ReuseCount int
	Status     KnowledgeStatus
	Tags       []KnowledgeTag
	Relations  []KnowledgeRelation
	References []KnowledgeReference
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

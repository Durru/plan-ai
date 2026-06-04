package knowledge

import (
	"github.com/plan-ai/plan-ai/internal/domain"
)

// Re-exported domain enum types for callers that prefer the knowledge
// package as the import surface.
type (
	Category      = domain.KnowledgeCategory
	Status        = domain.KnowledgeStatus
	SourceType    = domain.KnowledgeSourceType
	RelationType  = domain.KnowledgeRelationType
	ReferenceType = domain.KnowledgeReferenceType
)

const (
	CategoryDatabase       = domain.KnowledgeCategoryDatabase
	CategoryAuthentication = domain.KnowledgeCategoryAuthentication
	CategoryBilling        = domain.KnowledgeCategoryBilling
	CategoryFrontend       = domain.KnowledgeCategoryFrontend
	CategoryBackend        = domain.KnowledgeCategoryBackend
	CategorySecurity       = domain.KnowledgeCategorySecurity
	CategoryDeployment     = domain.KnowledgeCategoryDeployment
	CategoryArchitecture   = domain.KnowledgeCategoryArchitecture
	CategoryTesting        = domain.KnowledgeCategoryTesting
	CategoryMCP            = domain.KnowledgeCategoryMCP
	CategoryAgents         = domain.KnowledgeCategoryAgents
	CategoryAI             = domain.KnowledgeCategoryAI
	CategoryDevops         = domain.KnowledgeCategoryDevops
	CategoryIntegration    = domain.KnowledgeCategoryIntegration
	CategoryGeneral        = domain.KnowledgeCategoryGeneral
)

// Phase 12 knowledge categories. Authentication/billing aliases remain for
// the older knowledge-base surface, while auth/payments match the new plan.
const (
	CategoryAuth     domain.KnowledgeCategory = "auth"
	CategoryPayments domain.KnowledgeCategory = "payments"
)

type KnowledgeObject struct {
	ID                  string
	ProjectID           string
	Title               string
	Category            domain.KnowledgeCategory
	Summary             string
	ResearchIDs         []string
	RelatedDecisions    []string
	RelatedRequirements []string
	RelatedConstraints  []string
	Confidence          float64
	CreatedAt           string
	UpdatedAt           string
}

type CreateKnowledgeRequest struct {
	ProjectID           string
	Title               string
	Category            domain.KnowledgeCategory
	Summary             string
	ResearchIDs         []string
	RelatedDecisions    []string
	RelatedRequirements []string
	RelatedConstraints  []string
	Confidence          float64
}

const (
	StatusDraft    = domain.KnowledgeStatusDraft
	StatusReviewed = domain.KnowledgeStatusReviewed
	StatusApproved = domain.KnowledgeStatusApproved
	StatusArchived = domain.KnowledgeStatusArchived
)

const (
	SourceManual    = domain.KnowledgeSourceManual
	SourceResearch  = domain.KnowledgeSourceResearch
	SourceImported  = domain.KnowledgeSourceImported
	SourceGenerated = domain.KnowledgeSourceGenerated
)

const (
	RelationRelated       = domain.KnowledgeRelationRelated
	RelationDependsOn     = domain.KnowledgeRelationDependsOn
	RelationAlternativeTo = domain.KnowledgeRelationAlternativeTo
	RelationExtends       = domain.KnowledgeRelationExtends
)

const (
	ReferencePlan       = domain.KnowledgeReferencePlan
	ReferenceDecision   = domain.KnowledgeReferenceDecision
	ReferenceResearch   = domain.KnowledgeReferenceResearch
	ReferenceTechnology = domain.KnowledgeReferenceTechnology
)

// Tag is a single label attached to a KnowledgeObject.
type Tag struct {
	ID          string
	KnowledgeID string
	Tag         string
	CreatedAt   string
}

// Relation is a directed link between two KnowledgeObjects.
type Relation struct {
	ID           string
	SourceID     string
	TargetID     string
	RelationType domain.KnowledgeRelationType
	CreatedAt    string
}

// Reference is a link from a KnowledgeObject to a plan / decision /
// research / technology record in the same project store.
type Reference struct {
	ID            string
	KnowledgeID   string
	ReferenceType domain.KnowledgeReferenceType
	ReferenceID   string
	CreatedAt     string
}

// Summary is a flat projection used by status / dashboards.
type Summary struct {
	Total    int
	Approved int
	Reviewed int
	Draft    int
	Archived int
	Reused   int
}

// CreateInput is the validated payload accepted by Service.CreateKnowledge.
type CreateInput struct {
	ID         string
	Topic      string
	Category   domain.KnowledgeCategory
	Summary    string
	Content    string
	Confidence float64
	SourceType domain.KnowledgeSourceType
	Tags       []string
	Status     domain.KnowledgeStatus
}

// UpdateInput is the validated payload accepted by Service.UpdateKnowledge.
type UpdateInput struct {
	Topic      string
	Category   domain.KnowledgeCategory
	Summary    string
	Content    string
	Confidence float64
	SourceType domain.KnowledgeSourceType
	Status     domain.KnowledgeStatus
}

package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ──────────────────────────────────────────────
// Shared Status type (used across all entities)
// ──────────────────────────────────────────────

// Status is a general-purpose lifecycle status. Each entity type
// defines its canonical set of valid Status values.
type Status string

const (
	StatusDraft       Status = "draft"
	StatusInReview    Status = "in_review"
	StatusApproved    Status = "approved"
	StatusRejected    Status = "rejected"
	StatusBlocked     Status = "blocked"
	StatusImplemented Status = "implemented"
	StatusValidated   Status = "validated"
	StatusArchived    Status = "archived"
	StatusCompleted   Status = "completed"
)

// ──────────────────────────────────────────────
// Plan types
// ──────────────────────────────────────────────

type PlanType string

const (
	PlanTypeMaster   PlanType = "master"
	PlanTypeSpecific PlanType = "specific"
)

// ──────────────────────────────────────────────
// Context size for task execution
// ──────────────────────────────────────────────

type ContextSize string

const (
	ContextSizeShort  ContextSize = "short"
	ContextSizeMedium ContextSize = "medium"
	ContextSizeFull   ContextSize = "full"
)

// ──────────────────────────────────────────────
// Validation targets
// ──────────────────────────────────────────────

type ValidationTargetType string

const (
	ValidationTargetPlan     ValidationTargetType = "plan"
	ValidationTargetPhase    ValidationTargetType = "phase"
	ValidationTargetTask     ValidationTargetType = "task"
	ValidationTargetDecision ValidationTargetType = "decision"
)

// ──────────────────────────────────────────────
// Research-specific types (legacy compatibility)
// ──────────────────────────────────────────────

type ResearchStatus string

const (
	ResearchStatusDraft    ResearchStatus = "draft"
	ResearchStatusInReview ResearchStatus = "in_review"
	ResearchStatusApproved ResearchStatus = "approved"
	ResearchStatusRejected ResearchStatus = "rejected"
	ResearchStatusArchived ResearchStatus = "archived"
)

type ResearchSourceType string

const (
	ResearchSourceManual        ResearchSourceType = "manual"
	ResearchSourceDocumentation ResearchSourceType = "documentation"
	ResearchSourceArticle       ResearchSourceType = "article"
	ResearchSourceRepository    ResearchSourceType = "repository"
	ResearchSourceSpecification ResearchSourceType = "specification"
	ResearchSourceBenchmark     ResearchSourceType = "benchmark"
	ResearchSourceInternal      ResearchSourceType = "internal"
)

// ──────────────────────────────────────────────
// Knowledge-specific types (used by internal/knowledge)
// ──────────────────────────────────────────────

type KnowledgeCategory string

const (
	KnowledgeCategoryDatabase       KnowledgeCategory = "database"
	KnowledgeCategoryAuthentication KnowledgeCategory = "authentication"
	KnowledgeCategoryBilling        KnowledgeCategory = "billing"
	KnowledgeCategoryFrontend       KnowledgeCategory = "frontend"
	KnowledgeCategoryBackend        KnowledgeCategory = "backend"
	KnowledgeCategorySecurity       KnowledgeCategory = "security"
	KnowledgeCategoryDeployment     KnowledgeCategory = "deployment"
	KnowledgeCategoryArchitecture   KnowledgeCategory = "architecture"
	KnowledgeCategoryTesting        KnowledgeCategory = "testing"
	KnowledgeCategoryMCP            KnowledgeCategory = "mcp"
	KnowledgeCategoryAgents         KnowledgeCategory = "agents"
	KnowledgeCategoryAI             KnowledgeCategory = "ai"
	KnowledgeCategoryDevops         KnowledgeCategory = "devops"
	KnowledgeCategoryIntegration    KnowledgeCategory = "integration"
	KnowledgeCategoryGeneral        KnowledgeCategory = "general"
)

type KnowledgeStatus string

const (
	KnowledgeStatusDraft    KnowledgeStatus = "draft"
	KnowledgeStatusReviewed KnowledgeStatus = "reviewed"
	KnowledgeStatusApproved KnowledgeStatus = "approved"
	KnowledgeStatusArchived KnowledgeStatus = "archived"
)

type KnowledgeSourceType string

const (
	KnowledgeSourceManual    KnowledgeSourceType = "manual"
	KnowledgeSourceResearch  KnowledgeSourceType = "research"
	KnowledgeSourceImported  KnowledgeSourceType = "imported"
	KnowledgeSourceGenerated KnowledgeSourceType = "generated"
)

type KnowledgeRelationType string

const (
	KnowledgeRelationRelated       KnowledgeRelationType = "related"
	KnowledgeRelationDependsOn     KnowledgeRelationType = "depends_on"
	KnowledgeRelationAlternativeTo KnowledgeRelationType = "alternative_to"
	KnowledgeRelationExtends       KnowledgeRelationType = "extends"
)

type KnowledgeReferenceType string

const (
	KnowledgeReferencePlan       KnowledgeReferenceType = "plan"
	KnowledgeReferenceDecision   KnowledgeReferenceType = "decision"
	KnowledgeReferenceResearch   KnowledgeReferenceType = "research"
	KnowledgeReferenceTechnology KnowledgeReferenceType = "technology"
)

// ──────────────────────────────────────────────
// ID generation
// ──────────────────────────────────────────────

func NewID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes[:]))
}

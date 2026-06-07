package vision

import (
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/intent"
)

type DocumentStatus string

const (
	DocumentDraft    DocumentStatus = "draft"
	DocumentApproved DocumentStatus = "approved"
)

type Document struct {
	ID                string
	ProjectID         string
	IntentProfileID   string
	Source            string
	FunctionalVision  string
	VisualVision      string
	TechnicalVision   string
	OperationalVision string
	BusinessVision    string
	Status            DocumentStatus
	Approved          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type DocumentRepository interface {
	SaveDocument(Document) (Document, error)
	GetDocument(id string) (Document, error)
	ListDocuments(projectID string) ([]Document, error)
	ApproveDocument(id string) (Document, error)
}

type DocumentService struct{ repo DocumentRepository }

func NewDocumentService(repo DocumentRepository) DocumentService { return DocumentService{repo: repo} }

func (s DocumentService) CreateFromContent(projectID, content string) (Document, error) {
	return s.repo.SaveDocument(BuildDocumentFromContent(projectID, content))
}

func (s DocumentService) CreateFromIntent(profile intent.Profile) (Document, error) {
	return s.repo.SaveDocument(BuildDocumentFromIntent(profile))
}

func (s DocumentService) Get(id string) (Document, error) { return s.repo.GetDocument(id) }

func (s DocumentService) List(projectID string) ([]Document, error) {
	return s.repo.ListDocuments(projectID)
}

func (s DocumentService) Approve(id string) (Document, error) { return s.repo.ApproveDocument(id) }

func BuildDocumentFromContent(projectID, content string) Document {
	normalized := strings.ToLower(content)
	doc := Document{ProjectID: projectID, Source: content, Status: DocumentDraft}
	doc.FunctionalVision = "Clarify core user workflows before planning."
	doc.VisualVision = "Visual style needs user approval before implementation planning."
	doc.TechnicalVision = "Technical stack remains undecided until approved constraints are available."
	doc.OperationalVision = "Operational model needs ownership, roles, and support expectations."
	doc.BusinessVision = "Business model needs explicit approval before scope is frozen."
	if strings.Contains(normalized, "ecommerce") || strings.Contains(normalized, "tienda") {
		doc.FunctionalVision = "Ecommerce experience with catalog, product detail, cart, checkout, and order flow candidates."
		doc.BusinessVision = "Online sales business model with payments, fulfillment, and customer retention candidates."
	}
	if strings.Contains(normalized, "crm") {
		doc.FunctionalVision = "CRM experience for managing contacts, pipelines, reports, and automations as candidates."
		doc.BusinessVision = "SaaS CRM business model with subscriptions and admin operations as candidates."
	}
	return doc
}

func BuildDocumentFromIntent(profile intent.Profile) Document {
	doc := BuildDocumentFromContent(profile.ProjectID, profile.Source)
	doc.IntentProfileID = profile.ID
	if profile.PrimaryIntent.Name != "" {
		doc.FunctionalVision = profile.PrimaryIntent.Name + " functional vision requires user-approved workflows before planning."
	}
	return doc
}

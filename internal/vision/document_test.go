package vision

import "testing"

func TestBuildDocumentFromContentCreatesFiveVisionDimensions(t *testing.T) {
	doc := BuildDocumentFromContent("project", "quiero un ecommerce")
	if doc.FunctionalVision == "" || doc.VisualVision == "" || doc.TechnicalVision == "" || doc.OperationalVision == "" || doc.BusinessVision == "" {
		t.Fatalf("expected all vision dimensions, got %#v", doc)
	}
	if doc.Approved || doc.Status != DocumentDraft {
		t.Fatalf("document must start unapproved draft: %#v", doc)
	}
}

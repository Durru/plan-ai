package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/knowledge"
)

func openKnowledgeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "knowledge.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("run project migrations: %v", err)
	}
	return db
}

func newKnowledgeService(t *testing.T) (*knowledge.Service, *sql.DB) {
	t.Helper()
	db := openKnowledgeTestDB(t)
	return knowledge.NewService(NewKnowledgeRepository(db)), db
}

func TestKnowledgeRepositoryCreateGetAndList(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	created, err := service.CreateKnowledge(knowledge.CreateInput{
		Topic:      "PostgreSQL Multi Tenant",
		Summary:    "Tenant isolation patterns",
		Content:    "Schema per tenant is a good default",
		Tags:       []string{"postgres", "tenant", "Postgres"},
		SourceType: knowledge.SourceManual,
		Status:     knowledge.StatusApproved,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Category != domain.KnowledgeCategoryDatabase {
		t.Fatalf("category = %q, want database", created.Category)
	}
	if created.Status != domain.KnowledgeStatusApproved {
		t.Fatalf("status = %q, want approved", created.Status)
	}
	if created.ReuseCount != 0 {
		t.Fatalf("reuse count = %d, want 0", created.ReuseCount)
	}

	got, err := service.GetKnowledge(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Topic != created.Topic {
		t.Fatalf("topic mismatch: %q vs %q", got.Topic, created.Topic)
	}

	listed, err := service.ListKnowledge()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("listed count = %d, want 1", len(listed))
	}
}

func TestKnowledgeRepositoryClassifiesWhenCategoryOmitted(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	object, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Stripe Webhooks"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if object.Category != domain.KnowledgeCategoryBilling {
		t.Fatalf("auto-category = %q, want billing", object.Category)
	}
}

func TestKnowledgeRepositoryAddTagNormalizesAndDeduplicates(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	object, err := service.CreateKnowledge(knowledge.CreateInput{
		Topic: "OAuth 2.0",
		Tags:  []string{"auth", "Auth", "  AUTH  "},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, _, _, _, err = service.Describe(object.ID)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	tagRows, err := NewKnowledgeRepository(db).ListTags(object.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tagRows) != 1 {
		t.Fatalf("tag count = %d, want 1 (dedup)", len(tagRows))
	}
	if tagRows[0].Tag != "auth" {
		t.Fatalf("tag = %q, want %q", tagRows[0].Tag, "auth")
	}
}

func TestKnowledgeRepositoryReuseIncrementsAndUpdatesTimestamp(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	created, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Sample"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	originalStored, err := NewKnowledgeRepository(db).GetByID(created.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	reused, err := service.ReuseKnowledge(created.ID)
	if err != nil {
		t.Fatalf("reuse: %v", err)
	}
	if reused.ReuseCount != 1 {
		t.Fatalf("reuse count = %d, want 1", reused.ReuseCount)
	}
	if reused.UpdatedAt.Before(originalStored.UpdatedAt) {
		t.Fatalf("updated_at must not move backwards; original=%s reused=%s", originalStored.UpdatedAt, reused.UpdatedAt)
	}
	reused2, err := service.ReuseKnowledge(created.ID)
	if err != nil {
		t.Fatalf("reuse 2: %v", err)
	}
	if reused2.ReuseCount != 2 {
		t.Fatalf("reuse count = %d, want 2", reused2.ReuseCount)
	}
}

func TestKnowledgeRepositoryListByCategoryAndSearch(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	for _, topic := range []string{"PostgreSQL Multi Tenant", "OAuth 2.0", "Stripe Billing"} {
		if _, err := service.CreateKnowledge(knowledge.CreateInput{Topic: topic}); err != nil {
			t.Fatalf("create %s: %v", topic, err)
		}
	}

	dbEntries, err := service.ListByCategory(domain.KnowledgeCategoryDatabase)
	if err != nil {
		t.Fatalf("list by category: %v", err)
	}
	if len(dbEntries) != 1 {
		t.Fatalf("db category count = %d, want 1", len(dbEntries))
	}

	results, err := service.SearchKnowledge("billing")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("search count = %d, want 1", len(results))
	}
	if results[0].Topic != "Stripe Billing" {
		t.Fatalf("search result topic = %q, want %q", results[0].Topic, "Stripe Billing")
	}

	resultsEmpty, err := service.SearchKnowledge("   ")
	if err != nil {
		t.Fatalf("search empty: %v", err)
	}
	if len(resultsEmpty) != 3 {
		t.Fatalf("empty search count = %d, want 3 (full list)", len(resultsEmpty))
	}
}

func TestKnowledgeRepositoryRelationsAndReferences(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	pg, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "PostgreSQL Multi Tenant"})
	if err != nil {
		t.Fatalf("create pg: %v", err)
	}
	oauth, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "OAuth 2.0"})
	if err != nil {
		t.Fatalf("create oauth: %v", err)
	}
	stripe, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Stripe Billing"})
	if err != nil {
		t.Fatalf("create stripe: %v", err)
	}

	if err := service.LinkKnowledge(pg.ID, oauth.ID, knowledge.RelationDependsOn); err != nil {
		t.Fatalf("link: %v", err)
	}
	if err := service.LinkKnowledge(stripe.ID, pg.ID, knowledge.RelationRelated); err != nil {
		t.Fatalf("link stripe: %v", err)
	}
	if err := service.LinkKnowledge(pg.ID, pg.ID, knowledge.RelationRelated); err == nil {
		t.Fatalf("expected error for self-link")
	}
	if err := service.LinkKnowledge(pg.ID, "missing", knowledge.RelationRelated); err == nil {
		t.Fatalf("expected error for missing target")
	}

	_, _, relations, _, err := service.Describe(pg.ID)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if len(relations) != 2 {
		t.Fatalf("pg relations = %d, want 2 (1 out, 1 in)", len(relations))
	}

	if err := service.AttachReference(pg.ID, knowledge.ReferencePlan, "plan_123"); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if err := service.AttachReference(pg.ID, knowledge.ReferencePlan, "plan_123"); err != nil {
		t.Fatalf("attach duplicate should be a no-op: %v", err)
	}
	_, _, _, references, err := service.Describe(pg.ID)
	if err != nil {
		t.Fatalf("describe refs: %v", err)
	}
	if len(references) != 1 {
		t.Fatalf("references = %d, want 1", len(references))
	}
	if references[0].ReferenceID != "plan_123" {
		t.Fatalf("reference id = %q, want plan_123", references[0].ReferenceID)
	}
}

func TestKnowledgeRepositorySummaryCountsByStatus(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	if _, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Draft 1", Status: knowledge.StatusDraft}); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	approved, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Approved 1", Status: knowledge.StatusApproved})
	if err != nil {
		t.Fatalf("create approved: %v", err)
	}
	approved2, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "Approved 2", Status: knowledge.StatusApproved})
	if err != nil {
		t.Fatalf("create approved 2: %v", err)
	}
	if _, err := service.ReuseKnowledge(approved.ID); err != nil {
		t.Fatalf("reuse: %v", err)
	}
	if _, err := service.ReuseKnowledge(approved.ID); err != nil {
		t.Fatalf("reuse 2: %v", err)
	}
	if _, err := service.ReuseKnowledge(approved2.ID); err != nil {
		t.Fatalf("reuse 3: %v", err)
	}

	summary, err := service.GetSummary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Total != 3 {
		t.Fatalf("total = %d, want 3", summary.Total)
	}
	if summary.Draft != 1 {
		t.Fatalf("draft = %d, want 1", summary.Draft)
	}
	if summary.Approved != 2 {
		t.Fatalf("approved = %d, want 2", summary.Approved)
	}
	if summary.Reused != 3 {
		t.Fatalf("reused = %d, want 3", summary.Reused)
	}
}

func TestKnowledgeServiceValidation(t *testing.T) {
	service, db := newKnowledgeService(t)
	defer db.Close()

	if _, err := service.CreateKnowledge(knowledge.CreateInput{}); err == nil {
		t.Fatalf("expected error for empty topic")
	}
	if _, err := service.GetKnowledge(""); err == nil {
		t.Fatalf("expected error for empty id")
	}
	if _, err := service.ReuseKnowledge(""); err == nil {
		t.Fatalf("expected error for empty reuse id")
	}
	if err := service.LinkKnowledge("", "", knowledge.RelationRelated); err == nil {
		t.Fatalf("expected error for empty ids")
	}
	if err := service.AttachReference("", knowledge.ReferencePlan, "x"); err == nil {
		t.Fatalf("expected error for empty knowledge id")
	}
	if err := service.AddTag("", "tag"); err == nil {
		t.Fatalf("expected error for empty knowledge id in AddTag")
	}
	if err := service.AddTag("knowledge_x", "  "); err == nil {
		t.Fatalf("expected error for empty tag")
	}
	if _, err := service.CreateKnowledge(knowledge.CreateInput{Topic: "x", SourceType: domain.KnowledgeSourceType("weird")}); err == nil {
		t.Fatalf("expected error for unknown source type")
	}
}

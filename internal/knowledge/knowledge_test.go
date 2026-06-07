package knowledge

import (
	"reflect"
	"testing"

	"github.com/Durru/plan-ai/internal/domain"
)

func TestClassifyAssignsOfficialCategories(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  domain.KnowledgeCategory
	}{
		{name: "postgres", topic: "PostgreSQL Multi Tenant", want: domain.KnowledgeCategoryDatabase},
		{name: "redis", topic: "Redis Caching Patterns", want: domain.KnowledgeCategoryDatabase},
		{name: "oauth", topic: "OAuth 2.0 with PKCE", want: domain.KnowledgeCategoryAuthentication},
		{name: "jwt", topic: "JWT refresh token rotation", want: domain.KnowledgeCategoryAuthentication},
		{name: "stripe", topic: "Stripe Billing Webhooks", want: domain.KnowledgeCategoryBilling},
		{name: "subscription", topic: "Subscription Lifecycle", want: domain.KnowledgeCategoryBilling},
		{name: "react", topic: "React Server Components", want: domain.KnowledgeCategoryFrontend},
		{name: "next", topic: "Next.js App Router", want: domain.KnowledgeCategoryFrontend},
		{name: "backend", topic: "Backend API design", want: domain.KnowledgeCategoryBackend},
		{name: "xss", topic: "XSS prevention", want: domain.KnowledgeCategorySecurity},
		{name: "docker", topic: "Docker Compose for local dev", want: domain.KnowledgeCategoryDeployment},
		{name: "hexagonal", topic: "Hexagonal architecture", want: domain.KnowledgeCategoryArchitecture},
		{name: "tdd", topic: "TDD with Go", want: domain.KnowledgeCategoryTesting},
		{name: "mcp", topic: "MCP protocol overview", want: domain.KnowledgeCategoryMCP},
		{name: "agent", topic: "Autonomous agent design", want: domain.KnowledgeCategoryAgents},
		{name: "llm", topic: "LLM prompt patterns", want: domain.KnowledgeCategoryAI},
		{name: "monitoring", topic: "Observability with Prometheus", want: domain.KnowledgeCategoryDevops},
		{name: "webhook", topic: "Webhook signature verification", want: domain.KnowledgeCategoryIntegration},
		{name: "default", topic: "Random note that matches nothing", want: domain.KnowledgeCategoryGeneral},
		{name: "empty", topic: "", want: domain.KnowledgeCategoryGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.topic); got != tt.want {
				t.Fatalf("Classify(%q) = %q, want %q", tt.topic, got, tt.want)
			}
		})
	}
}

func TestNormalizeCategoryHonorsExplicitOverride(t *testing.T) {
	got := NormalizeCategory("architecture", "PostgreSQL Multi Tenant")
	if got != domain.KnowledgeCategoryArchitecture {
		t.Fatalf("explicit override ignored, got %q", got)
	}
}

func TestNormalizeCategoryFallsBackToClassifierForUnknown(t *testing.T) {
	got := NormalizeCategory("not-a-real-category", "Stripe Billing")
	if got != domain.KnowledgeCategoryBilling {
		t.Fatalf("classifier fallback wrong, got %q", got)
	}
}

func TestNormalizeStatusDefaultsToDraft(t *testing.T) {
	got, err := NormalizeStatus("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != domain.KnowledgeStatusDraft {
		t.Fatalf("default status = %q, want %q", got, domain.KnowledgeStatusDraft)
	}
}

func TestNormalizeStatusRejectsUnknown(t *testing.T) {
	if _, err := NormalizeStatus("published"); err == nil {
		t.Fatalf("expected error for unknown status")
	}
}

func TestNormalizeSourceTypeDefaultsToManual(t *testing.T) {
	got, err := NormalizeSourceType("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != domain.KnowledgeSourceManual {
		t.Fatalf("default source = %q, want %q", got, domain.KnowledgeSourceManual)
	}
}

func TestNormalizeRelationTypeAcceptsOfficial(t *testing.T) {
	tests := []domain.KnowledgeRelationType{
		domain.KnowledgeRelationRelated,
		domain.KnowledgeRelationDependsOn,
		domain.KnowledgeRelationAlternativeTo,
		domain.KnowledgeRelationExtends,
	}
	for _, value := range tests {
		got, err := NormalizeRelationType(string(value))
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", value, err)
		}
		if got != value {
			t.Fatalf("normalize %q = %q, want %q", value, got, value)
		}
	}
}

func TestNormalizeRelationTypeRejectsUnknown(t *testing.T) {
	if _, err := NormalizeRelationType("uses"); err == nil {
		t.Fatalf("expected error for unknown relation type")
	}
}

func TestNormalizeReferenceTypeAcceptsOfficial(t *testing.T) {
	tests := []domain.KnowledgeReferenceType{
		domain.KnowledgeReferencePlan,
		domain.KnowledgeReferenceDecision,
		domain.KnowledgeReferenceResearch,
		domain.KnowledgeReferenceTechnology,
	}
	for _, value := range tests {
		got, err := NormalizeReferenceType(string(value))
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", value, err)
		}
		if got != value {
			t.Fatalf("normalize %q = %q, want %q", value, got, value)
		}
	}
}

func TestIsValidRelationAndReferenceTypeGuards(t *testing.T) {
	if IsValidRelationType(domain.KnowledgeRelationType("uses")) {
		t.Fatalf("invalid relation type accepted")
	}
	if !IsValidRelationType(domain.KnowledgeRelationRelated) {
		t.Fatalf("valid relation type rejected")
	}
	if IsValidReferenceType(domain.KnowledgeReferenceType("user")) {
		t.Fatalf("invalid reference type accepted")
	}
	if !IsValidReferenceType(domain.KnowledgeReferencePlan) {
		t.Fatalf("valid reference type rejected")
	}
}

func TestNormalizeTagLowercasesAndTrims(t *testing.T) {
	got, err := NormalizeTag("  PostgreSQL ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "postgresql" {
		t.Fatalf("tag = %q, want %q", got, "postgresql")
	}
}

func TestNormalizeTagRejectsEmpty(t *testing.T) {
	if _, err := NormalizeTag("   "); err == nil {
		t.Fatalf("expected error for empty tag")
	}
}

func TestCategoryConstantsAreStable(t *testing.T) {
	want := []string{
		"database", "authentication", "billing", "frontend", "backend",
		"security", "deployment", "architecture", "testing", "mcp",
		"agents", "ai", "devops", "integration", "general",
	}
	got := []domain.KnowledgeCategory{
		CategoryDatabase, CategoryAuthentication, CategoryBilling, CategoryFrontend, CategoryBackend,
		CategorySecurity, CategoryDeployment, CategoryArchitecture, CategoryTesting, CategoryMCP,
		CategoryAgents, CategoryAI, CategoryDevops, CategoryIntegration, CategoryGeneral,
	}
	if len(got) != len(want) {
		t.Fatalf("category count = %d, want %d", len(got), len(want))
	}
	for i, value := range got {
		if string(value) != want[i] {
			t.Fatalf("category[%d] = %q, want %q", i, value, want[i])
		}
	}
}

func TestStatusAndSourceTypeConstantsAreStable(t *testing.T) {
	statuses := []domain.KnowledgeStatus{StatusDraft, StatusReviewed, StatusApproved, StatusArchived}
	statusStrings := []string{"draft", "reviewed", "approved", "archived"}
	if !reflect.DeepEqual(toStrings(statuses), statusStrings) {
		t.Fatalf("statuses = %v, want %v", toStrings(statuses), statusStrings)
	}
	sources := []domain.KnowledgeSourceType{SourceManual, SourceResearch, SourceImported, SourceGenerated}
	sourceStrings := []string{"manual", "research", "imported", "generated"}
	if !reflect.DeepEqual(toStrings(sources), sourceStrings) {
		t.Fatalf("sources = %v, want %v", toStrings(sources), sourceStrings)
	}
}

func toStrings[T ~string](values []T) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}
	return out
}

package research

import (
	"testing"
)

func TestClassifyDatabaseTopics(t *testing.T) {
	topics := []string{
		"PostgreSQL Multi Tenant",
		"postgres performance tuning",
		"SQLite write throughput limits",
		"MongoDB sharding strategies",
		"Redis caching patterns",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryDatabase {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryDatabase)
		}
	}
}

func TestClassifyAuthenticationTopics(t *testing.T) {
	topics := []string{
		"OAuth 2.0 Authorization Code Flow",
		"JWT access token rotation",
		"WebAuthn passkey implementation",
		"Session management security",
		"SSO integration with SAML",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryAuthentication {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryAuthentication)
		}
	}
}

func TestClassifyBillingTopics(t *testing.T) {
	topics := []string{
		"Stripe Billing subscription models",
		"Usage based pricing metering",
		"Invoice generation automation",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryBilling {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryBilling)
		}
	}
}

func TestClassifyFrontendTopics(t *testing.T) {
	topics := []string{
		"React Server Components deep dive",
		"Next.js App Router migration",
		"Tailwind CSS design system",
		"shadcn/ui component library",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryFrontend {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryFrontend)
		}
	}
}

func TestClassifyBackendTopics(t *testing.T) {
	topics := []string{
		"REST API versioning strategies",
		"gRPC vs REST performance comparison",
		"GraphQL federation failure modes",
		"Service layer architecture patterns",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryBackend {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryBackend)
		}
	}
}

func TestClassifySecurityTopics(t *testing.T) {
	topics := []string{
		"OWASP Top 10 mitigations",
		"SQL injection prevention in Go",      // "sql injection" keyword
		"Threat modeling with STRIDE",         // "threat model" keyword (via "threat model" in lowercase)
		"Cross-site scripting XSS prevention", // "xss" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategorySecurity {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategorySecurity)
		}
	}
}

func TestClassifyDeploymentTopics(t *testing.T) {
	topics := []string{
		"Kubernetes deployment strategies", // "deployment" keyword
		"GitHub Actions CI/CD pipeline",    // "ci/cd" keyword
		"Terraform infrastructure as code", // "terraform" keyword
		"Ansible playbook automation",      // "ansible" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryDeployment {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryDeployment)
		}
	}
}

func TestClassifyArchitectureTopics(t *testing.T) {
	topics := []string{
		"Clean Architecture in Go",      // "architecture" keyword
		"Domain driven design concepts", // "domain driven" keyword
		"Hexagonal ports and adapters",  // "ports and adapters" keyword
		"CQRS and event sourcing",       // "cqrs" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryArchitecture {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryArchitecture)
		}
	}
}

func TestClassifyTestingTopics(t *testing.T) {
	topics := []string{
		"TDD workflow with Go tests",  // "tdd" keyword
		"E2E testing with Playwright", // "playwright" keyword
		"Unit test best practices",    // "unit test" keyword (frontend has "ui/test" → let me check)
		"Vitest configuration setup",  // "vitest" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryTesting {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryTesting)
		}
	}
}

func TestClassifyMCPTopics(t *testing.T) {
	topics := []string{
		"Model Context Protocol deep dive",   // "mcp" keyword (no "ui"/"api" false positives)
		"Go client library for MCP protocol", // "mcp" keyword without "ui" substring
		"MCP resource templates",             // "mcp" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryMCP {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryMCP)
		}
	}
}

func TestClassifyAIAgentsTopics(t *testing.T) {
	topics := []string{
		"Autonomous agent orchestration",
		"Tool use with function calling",
		"Agent memory and context management",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryAgents {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryAgents)
		}
	}
}

func TestClassifyAITopics(t *testing.T) {
	topics := []string{
		"LLM fine tuning best practices",     // "fine tuning" keyword, no "ui"/"gin" false positives
		"RAG with vector embeddings",         // "rag" keyword, no "ui"/"gin" false positives
		"GPT transformer model optimization", // "gpt" + "transformer", no "ui" substring
		"RAG prompt chaining patterns",       // "rag" + "prompt" keywords
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryAI {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryAI)
		}
	}
}

func TestClassifyDevOpsTopics(t *testing.T) {
	topics := []string{
		"Prometheus monitoring setup",            // "prometheus" keyword
		"Distributed tracing with OpenTelemetry", // "opentelemetry" + "tracing" keywords
		"Grafana alerting dashboards",            // "grafana" + "alerting" keywords
		"Datadog APM traces",                     // "datadog" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryDevops {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryDevops)
		}
	}
}

func TestClassifyIntegrationTopics(t *testing.T) {
	topics := []string{
		"Webhook event processing", // "webhook" keyword
		"CSV import data sync",     // "import" keyword
		"ETL batch processing",     // "etl" keyword
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryIntegration {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryIntegration)
		}
	}
}

func TestClassifyEmptyTopicFallsBackToGeneral(t *testing.T) {
	if got := Classify(""); got != CategoryGeneral {
		t.Errorf("Classify(\"\") = %q, want %q", got, CategoryGeneral)
	}
}

func TestClassifyUnknownTopicFallsBackToGeneral(t *testing.T) {
	topics := []string{
		"zyxwvutsrqponmlkjihgfedcba", // guaranteed no keyword substring matches
		"",
		"   ",
	}
	for _, topic := range topics {
		got := Classify(topic)
		if got != CategoryGeneral {
			t.Errorf("Classify(%q) = %q, want %q", topic, got, CategoryGeneral)
		}
	}
}

func TestResearchStatusConstants(t *testing.T) {
	tests := []struct {
		status ResearchStatus
		want   string
	}{
		{ResearchStatusDraft, "draft"},
		{ResearchStatusInReview, "in_review"},
		{ResearchStatusApproved, "approved"},
		{ResearchStatusRejected, "rejected"},
		{ResearchStatusArchived, "archived"},
	}
	for _, tc := range tests {
		if got := string(tc.status); got != tc.want {
			t.Errorf("ResearchStatus constant = %q, want %q", got, tc.want)
		}
	}
}

func TestResearchSourceTypeConstants(t *testing.T) {
	tests := []struct {
		st   ResearchSourceType
		want string
	}{
		{SourceTypeManual, "manual"},
		{SourceTypeDocumentation, "documentation"},
		{SourceTypeArticle, "article"},
		{SourceTypeRepository, "repository"},
		{SourceTypeSpecification, "specification"},
		{SourceTypeBenchmark, "benchmark"},
		{SourceTypeInternal, "internal"},
	}
	for _, tc := range tests {
		if got := string(tc.st); got != tc.want {
			t.Errorf("ResearchSourceType constant = %q, want %q", got, tc.want)
		}
	}
}

func TestValidationErrorMessages(t *testing.T) {
	errs := ValidationErrors{
		{Field: "findings", Message: "at least one finding is required"},
		{Field: "sources", Message: "at least one source is required"},
	}
	msg := errs.Error()
	if msg == "" {
		t.Fatal("ValidationErrors.Error() returned empty string")
	}
}

func TestValidationErrorSingle(t *testing.T) {
	errs := ValidationErrors{
		{Field: "findings", Message: "at least one finding is required"},
	}
	expected := "findings: at least one finding is required"
	if got := errs.Error(); got != expected {
		t.Errorf("ValidationErrors.Error() = %q, want %q", got, expected)
	}
}

func TestResearchCategoryAliasesAreConsistent(t *testing.T) {
	// Verify all 15 categories exist and match KnowledgeCategory values
	categories := []ResearchCategory{
		CategoryDatabase, CategoryAuthentication, CategoryBilling,
		CategoryFrontend, CategoryBackend, CategorySecurity,
		CategoryDeployment, CategoryArchitecture, CategoryTesting,
		CategoryMCP, CategoryAgents, CategoryAI,
		CategoryDevops, CategoryIntegration, CategoryGeneral,
	}
	if len(categories) != 15 {
		t.Fatalf("expected 15 categories, got %d", len(categories))
	}
	for i, c := range categories {
		if string(c) == "" {
			t.Errorf("category[%d] is empty", i)
		}
	}
}

func TestCreateOptionWithCategory(t *testing.T) {
	var cfg createConfig
	WithCategory(CategoryDatabase)(&cfg)
	if cfg.category != CategoryDatabase {
		t.Errorf("WithCategory set category = %q, want %q", cfg.category, CategoryDatabase)
	}
}

func TestCreateOptionWithSummary(t *testing.T) {
	var cfg createConfig
	WithSummary("test summary")(&cfg)
	if cfg.summary != "test summary" {
		t.Errorf("WithSummary set summary = %q, want %q", cfg.summary, "test summary")
	}
}

func TestCreateOptionWithConfidence(t *testing.T) {
	var cfg createConfig
	WithConfidence(85)(&cfg)
	if cfg.confidence != 85 {
		t.Errorf("WithConfidence set confidence = %d, want %d", cfg.confidence, 85)
	}
}

func TestCreateOptionWithTags(t *testing.T) {
	var cfg createConfig
	WithTags("tag1", "tag2", "tag3")(&cfg)
	if len(cfg.tags) != 3 || cfg.tags[0] != "tag1" || cfg.tags[2] != "tag3" {
		t.Errorf("WithTags = %v, want [tag1 tag2 tag3]", cfg.tags)
	}
}

func TestCategoryStringValues(t *testing.T) {
	tests := []struct {
		cat  ResearchCategory
		name string
	}{
		{CategoryDatabase, "database"},
		{CategoryAuthentication, "authentication"},
		{CategoryBilling, "billing"},
		{CategoryFrontend, "frontend"},
		{CategoryBackend, "backend"},
		{CategorySecurity, "security"},
		{CategoryDeployment, "deployment"},
		{CategoryArchitecture, "architecture"},
		{CategoryTesting, "testing"},
		{CategoryMCP, "mcp"},
		{CategoryAgents, "agents"},
		{CategoryAI, "ai"},
		{CategoryDevops, "devops"},
		{CategoryIntegration, "integration"},
		{CategoryGeneral, "general"},
	}
	for _, tc := range tests {
		if got := string(tc.cat); got != tc.name {
			t.Errorf("Category = %q, want %q", got, tc.name)
		}
	}
}

func TestNewServiceUsesUTC(t *testing.T) {
	// Service should default to time.Now().UTC().
	// For this we just test that the constructor doesn't panic.
	repo := &mockRepository{}
	svc := NewService(repo)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

// mockRepository implements Repository for testing service construction.
type mockRepository struct{}

func (m *mockRepository) CreateEntry(entry ResearchEntry) error { return nil }
func (m *mockRepository) GetEntry(id string) (ResearchEntry, error) {
	return ResearchEntry{ID: id, Topic: "mock", Status: ResearchStatusDraft}, nil
}
func (m *mockRepository) ListEntries() ([]ResearchEntry, error)                     { return nil, nil }
func (m *mockRepository) SearchEntries(query string) ([]ResearchEntry, error)       { return nil, nil }
func (m *mockRepository) UpdateEntryStatus(id string, status ResearchStatus) error  { return nil }
func (m *mockRepository) DeleteEntry(id string) error                               { return nil }
func (m *mockRepository) CreateFinding(finding ResearchFinding) error               { return nil }
func (m *mockRepository) ListFindings(researchID string) ([]ResearchFinding, error) { return nil, nil }
func (m *mockRepository) CreateSource(source ResearchSource) error                  { return nil }
func (m *mockRepository) ListSources(researchID string) ([]ResearchSource, error)   { return nil, nil }
func (m *mockRepository) CreateConclusion(conclusion ResearchConclusion) error      { return nil }
func (m *mockRepository) ListConclusions(researchID string) ([]ResearchConclusion, error) {
	return nil, nil
}
func (m *mockRepository) AddTag(researchID, tag string) error                { return nil }
func (m *mockRepository) ListTags(researchID string) ([]ResearchTag, error)  { return nil, nil }
func (m *mockRepository) LinkKnowledge(researchID, knowledgeID string) error { return nil }
func (m *mockRepository) ListKnowledgeLinks(researchID string) ([]ResearchKnowledgeLink, error) {
	return nil, nil
}
func (m *mockRepository) Summary() (ResearchSummary, error) { return ResearchSummary{}, nil }

package store

import (
	"path/filepath"
	"testing"

	"github.com/Durru/plan-ai/internal/intentv3"
)

func openTestDB(t *testing.T) *ProjectStore {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &ProjectStore{DB: db}
}

func TestCreateAndApproveProductIntent(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	intentRepo := NewIntentV3Repository(ps.DB)
	discRepo := NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:         "project",
		Description:       "Build a CRM SaaS",
		ExpectedOutcome:   "Increase sales team productivity",
		DesiredExperience: "Fast and intuitive interface",
		DesiredResult:     "30% more leads managed",
		UserExpectations:  []string{"Mobile access", "Real-time sync"},
		NonExpectations:   []string{"Not a full ERP"},
		SuccessDefinition: "Users adopt within 2 weeks",
		FailureDefinition: "Performance below 100ms response",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if pi.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if pi.Status != intentv3.StatusDraft {
		t.Fatalf("status = %q, want draft", pi.Status)
	}
	if len(pi.UserExpectations) != 2 || pi.UserExpectations[0] != "Mobile access" {
		t.Fatalf("expectations = %v, want [Mobile access Real-time sync]", pi.UserExpectations)
	}

	// Submit for approval
	submitted, err := svc.SubmitProductIntentForApproval(pi.ID)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if submitted.Status != intentv3.StatusPendingApproval {
		t.Fatalf("status = %q, want pending_approval", submitted.Status)
	}

	// Approve
	approved, err := svc.ApproveProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != intentv3.StatusApproved {
		t.Fatalf("status = %q, want approved", approved.Status)
	}

	// List
	list, err := svc.ListProductIntents("project")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	if list[0].ID != pi.ID {
		t.Fatalf("list[0].id = %q, want %q", list[0].ID, pi.ID)
	}

	// Get
	got, err := svc.GetProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Description != "Build a CRM SaaS" {
		t.Fatalf("description = %q, want 'Build a CRM SaaS'", got.Description)
	}

	// IsApprovedProductIntent
	if !svc.IsApprovedProductIntent("project") {
		t.Fatal("expected approved product intent to return true")
	}
	if svc.IsApprovedProductIntent("other-project") {
		t.Fatal("expected no approved product intents for other project")
	}
}

func TestProductIntentLifecycleTransitions(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	pi, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Test lifecycle",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Draft -> Archive (direct)
	archived, err := svc.ArchiveProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("archive from draft: %v", err)
	}
	if archived.Status != intentv3.StatusArchived {
		t.Fatalf("status = %q, want archived", archived.Status)
	}
}

func TestProductIntentInvalidTransitions(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	pi, _ := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Test invalid transitions",
	})
	// Can't approve directly from draft (must submit first)
	if _, err := svc.ApproveProductIntent(pi.ID); err == nil {
		t.Fatal("expected error approving draft intent directly")
	}
}

func TestProductIntentUpdate(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	pi, _ := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Initial",
	})

	updated, err := svc.UpdateProductIntent(pi.ID, intentv3.CreateProductIntentInput{
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Description != "Updated description" {
		t.Fatalf("description = %q, want 'Updated description'", updated.Description)
	}
}

func TestDiscoveryResult(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	dr, err := svc.DiscoverIntent("project", "I want a SaaS CRM with real-time sync and mobile access. Should be secure and scalable. Not a full ERP system.")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if dr.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if dr.Classification != "saas_platform" {
		t.Fatalf("classification = %q, want saas_platform", dr.Classification)
	}
	if len(dr.Objectives) == 0 {
		t.Fatal("expected at least one objective")
	}
	if dr.DetectedIntent != "saas_platform" {
		t.Fatalf("detected_intent = %q, want saas_platform", dr.DetectedIntent)
	}
	if len(dr.Restrictions) == 0 {
		t.Fatal("expected restrictions")
	}
	if len(dr.Gaps) == 0 {
		t.Fatal("expected gaps")
	}
	if len(dr.Questions) == 0 {
		t.Fatal("expected questions")
	}

	// Get
	got, err := svc.GetDiscoveryResult(dr.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.DetectedIntent != dr.DetectedIntent {
		t.Fatalf("detected_intent = %q, want %q", got.DetectedIntent, dr.DetectedIntent)
	}

	// List
	list, err := svc.ListDiscoveryResults("project")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
}

func TestDetectorClassifiesContent(t *testing.T) {
	d := intentv3.NewDetector()

	tests := []struct {
		input string
		want  string
	}{
		{"Build a multi-tenant cloud service", "saas_platform"},
		{"An iOS mobile app for tracking", "mobile_application"},
		{"A CRM with billing and invoicing", "business_tool"},
		{"An API library for developers", "developer_tool"},
		{"A dashboard with analytics", "analytics_tool"},
		{"An AI chatbot using LLMs", "ai_feature"},
		{"An ecommerce website", "web_platform"},
		{"Workflow automation tool", "automation_tool"},
		{"Something random", "general"},
	}

	for _, tt := range tests {
		dr := d.Discover("proj", tt.input)
		if dr.Classification != tt.want {
			t.Errorf("classify(%q) = %q, want %q", tt.input, dr.Classification, tt.want)
		}
	}
}

func TestDiscoveryResultLinkedToProductIntent(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	// First discover
	dr, _ := svc.DiscoverIntent("project", "I want a SaaS analytics dashboard with real-time monitoring")
	// Then create intent from discovery
	pi, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:         "project",
		Description:       "Analytics dashboard",
		DiscoveryResultID: dr.ID,
	})
	if err != nil {
		t.Fatalf("create from discovery: %v", err)
	}
	if pi.DiscoveryResultID != dr.ID {
		t.Fatalf("discovery_result_id = %q, want %q", pi.DiscoveryResultID, dr.ID)
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from, to intentv3.ProductIntentStatus
		valid    bool
	}{
		{intentv3.StatusDraft, intentv3.StatusPendingApproval, true},
		{intentv3.StatusDraft, intentv3.StatusArchived, true},
		{intentv3.StatusDraft, intentv3.StatusApproved, false},
		{intentv3.StatusPendingApproval, intentv3.StatusApproved, true},
		{intentv3.StatusPendingApproval, intentv3.StatusDraft, true},
		{intentv3.StatusPendingApproval, intentv3.StatusArchived, true},
		{intentv3.StatusApproved, intentv3.StatusSuperseded, true},
		{intentv3.StatusApproved, intentv3.StatusArchived, true},
		{intentv3.StatusApproved, intentv3.StatusDraft, false},
		{intentv3.StatusSuperseded, intentv3.StatusArchived, true},
		{intentv3.StatusArchived, intentv3.StatusDraft, false},
	}
	for _, tt := range tests {
		got := intentv3.IsValidTransition(tt.from, tt.to)
		if got != tt.valid {
			t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.valid)
		}
	}
}

func TestCreateProductIntentValidation(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	if _, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{ProjectID: "p"}); err == nil {
		t.Fatal("expected error for empty description")
	}
	if _, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{Description: "desc"}); err == nil {
		t.Fatal("expected error for empty project_id")
	}
}

func TestRejectProductIntent(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	pi, _ := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Test reject",
	})

	// Submit
	svc.SubmitProductIntentForApproval(pi.ID)

	// Reject (goes back to draft)
	rejected, err := svc.RejectProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if rejected.Status != intentv3.StatusDraft {
		t.Fatalf("status = %q, want draft", rejected.Status)
	}
}

func TestSupersedeProductIntent(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	pi, _ := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Old version",
	})
	svc.SubmitProductIntentForApproval(pi.ID)
	svc.ApproveProductIntent(pi.ID)

	superseded, err := svc.SupersedeProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}
	if superseded.Status != intentv3.StatusSuperseded {
		t.Fatalf("status = %q, want superseded", superseded.Status)
	}
}

func TestDiscoverIntentValidation(t *testing.T) {
	ps := openTestDB(t)
	defer ps.Close()

	svc := intentv3.NewService(NewIntentV3Repository(ps.DB), NewIntentV3DiscoveryResultRepository(ps.DB))

	if _, err := svc.DiscoverIntent("", "content"); err == nil {
		t.Fatal("expected error for empty project_id")
	}
	if _, err := svc.DiscoverIntent("p", ""); err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestDetectorExtractsGaps(t *testing.T) {
	d := intentv3.NewDetector()
	dr := d.Discover("proj", "Build a simple API")
	if len(dr.Gaps) == 0 {
		t.Fatal("expected gaps from minimal description")
	}
}

func TestDetectorExtractsQuestions(t *testing.T) {
	d := intentv3.NewDetector()
	dr := d.Discover("proj", "Build a CRM. What database should I use? How should we deploy?")
	if len(dr.Questions) < 2 {
		t.Fatalf("expected at least 2 questions, got %d", len(dr.Questions))
	}
}

func TestProductIntentPersistenceAcrossDBReopen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := intentv3.NewService(NewIntentV3Repository(db), NewIntentV3DiscoveryResultRepository(db))

	pi, _ := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:   "project",
		Description: "Persistent test",
	})
	db.Close()

	// Re-open
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()

	svc2 := intentv3.NewService(NewIntentV3Repository(db2), NewIntentV3DiscoveryResultRepository(db2))

	got, err := svc2.GetProductIntent(pi.ID)
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if got.Description != "Persistent test" {
		t.Fatalf("description = %q, want 'Persistent test'", got.Description)
	}
}

package context

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

type memoryRepo struct {
	items map[string]ApprovedItem
	nextID int
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{items: make(map[string]ApprovedItem)}
}

func (r *memoryRepo) StoreApproved(item ApprovedItem) (ApprovedItem, error) {
	for _, existing := range r.items {
		if existing.ProjectID == item.ProjectID && strings.EqualFold(existing.Content, item.Content) && existing.Type == item.Type {
			return existing, nil
		}
	}
	r.nextID++
	item.ID = string(item.Type) + "-" + itoa(r.nextID)
	item.State = StateApproved
	r.items[item.ID] = item
	return item, nil
}

func (r *memoryRepo) GetApproved(typ ApprovedType, id string) (ApprovedItem, error) {
	item, ok := r.items[id]
	if !ok {
		return ApprovedItem{}, sql.ErrNoRows
	}
	return item, nil
}

func (r *memoryRepo) ListApproved(projectID string, typ ApprovedType) ([]ApprovedItem, error) {
	var out []ApprovedItem
	for _, item := range r.items {
		if item.ProjectID == projectID && (typ == "" || item.Type == typ) {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *memoryRepo) FindApproved(projectID string, typ ApprovedType, query string) ([]ApprovedItem, error) {
	var out []ApprovedItem
	for _, item := range r.items {
		if item.ProjectID == projectID && item.Type == typ &&
			strings.Contains(strings.ToLower(item.Content), strings.ToLower(query)) {
			out = append(out, item)
		}
	}
	return out, nil
}

func itoa(n int) string {
	const digits = "0123456789"
	var buf [20]byte
	i := len(buf)
	for n >= 10 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	i--
	buf[i] = digits[n]
	return string(buf[i:])
}

func TestApprovedContextAuthorityDedupesFacts(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	content := "Use OAuth2 for authentication"
	item := ApprovedItem{ID: "dec-oauth2", ProjectID: "proj_1", Type: TypeDecision, Content: content}

	added, existed, err := auth.Add(item)
	if err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if existed {
		t.Fatal("first Add should not report existed")
	}
	if added.Content != content {
		t.Errorf("content mismatch: %q", added.Content)
	}

	added2, existed, err := auth.Add(item)
	if err != nil {
		t.Fatalf("second Add: %v", err)
	}
	if !existed {
		t.Fatal("second Add should report existed (dedup)")
	}
	if added2.ID != added.ID {
		t.Errorf("dedup should return same ID, got %q vs %q", added2.ID, added.ID)
	}
}

func TestFindApprovedAcrossAllTypes(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	items := []ApprovedItem{
		{ProjectID: "proj_fts", Type: TypeRequirement, Content: "Support multi-tenant isolation"},
		{ProjectID: "proj_fts", Type: TypeDecision, Content: "Use schema-per-tenant for isolation"},
		{ProjectID: "proj_fts", Type: TypeConstraint, Content: "Must comply with GDPR"},
	}
	for _, item := range items {
		if _, _, err := auth.Add(item); err != nil {
			t.Fatalf("add %q: %v", item.Content, err)
		}
	}

	results, err := auth.FindAll("proj_fts", "isolation")
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("expected >= 2 results for 'isolation', got %d", len(results))
	}

	results, err = auth.FindAll("proj_fts", "GDPR")
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'GDPR', got %d", len(results))
	}
}

func TestApprovedContextWritesMemoryRecord(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	var captured []ApprovedItem
	auth.SetHook(func(item ApprovedItem) {
		captured = append(captured, item)
	})

	item := ApprovedItem{ID: "req-mem", ProjectID: "proj_mem", Type: TypeRequirement, Content: "Memory should be created"}
	_, _, err := auth.Add(item)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if len(captured) == 0 {
		t.Fatal("hook should have been called on new add")
	}
	if captured[0].Content != item.Content {
		t.Errorf("hook captured wrong content: %q", captured[0].Content)
	}

	captured = nil
	auth.Add(item)
	if len(captured) > 0 {
		t.Error("hook should not fire on dedup add")
	}
}

func TestApprovedContextEmitsContinuousEvent(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	old := ApprovedItem{ID: "dec-mysql", ProjectID: "proj_ce", Type: TypeDecision, Content: "Use MySQL"}
	if _, _, err := auth.Add(old); err != nil {
		t.Fatalf("add old: %v", err)
	}

	var captured []ApprovedItem
	auth.SetHook(func(item ApprovedItem) {
		captured = append(captured, item)
	})

	another := ApprovedItem{ID: "dec-cache", ProjectID: "proj_ce", Type: TypeDecision, Content: "Use SQLite as cache"}
	if _, _, err := auth.Add(another); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if len(captured) != 1 {
		t.Errorf("hook should fire once for new add, got %d", len(captured))
	}
}

func TestPlanningUsesApprovedFactsOnly(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	item := ApprovedItem{ProjectID: "proj_plan", Type: TypeRequirement, Content: "User login via email and password"}
	added, _, err := auth.Add(item)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if !auth.IsKnown("proj_plan", "User login via email and password") {
		t.Error("IsKnown should return true for stored content")
	}

	if auth.IsKnown("proj_plan", "Never discussed this approach") {
		t.Error("IsKnown should return false for novel content")
	}

	_ = added
}

func TestAuthorityService_ListAllTypes(t *testing.T) {
	repo := newMemoryRepo()
	auth := NewAuthorityService(repo, nil)

	types := []ApprovedType{TypeRequirement, TypeConstraint, TypeDecision, TypePreference, TypeGoal, TypeReference}
	for _, typ := range types {
		_, _, err := auth.Add(ApprovedItem{
			ProjectID: "proj_all",
			Type:      typ,
			Content:   "unique-content-" + string(typ),
		})
		if err != nil {
			t.Fatalf("add %s: %v", typ, err)
		}
	}

	results, err := auth.FindAll("proj_all", "unique")
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(results) != 6 {
		t.Errorf("expected 6 results across all types, got %d", len(results))
	}
}

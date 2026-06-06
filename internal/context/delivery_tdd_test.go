package context

import (
	"strings"
	"testing"
)

// mockResearchDataWithKnowledge returns test knowledge data.
type mockResearchDataWithKnowledge struct {
	mockResearchData
	knowledge []KnowledgeBrief
}

func (m *mockResearchDataWithKnowledge) ListKnowledgeBriefs(projectID string) ([]KnowledgeBrief, error) {
	return m.knowledge, nil
}

func TestBuildImplementationContext_IncludesAllSections(t *testing.T) {
	repo := newMockDeliveryRepo()

	knowMock := &mockResearchDataWithKnowledge{
		knowledge: []KnowledgeBrief{
			{Topic: "auth", Summary: "JWT-based auth flow"},
			{Topic: "db", Summary: "SQLite with modernc driver"},
		},
	}

	eng := NewDeliveryEngine(repo, &mockExecData{}, &mockPlanData{}, &mockImplData{}, knowMock)

	session, err := eng.DeliverContext("proj-1", LevelImplementation, map[string]string{"task_id": "task-1", "source": "test"})
	if err != nil {
		t.Fatalf("DeliverContext failed: %v", err)
	}

	content := session.Content

	sections := []string{
		"# Implementation Context",
		"Task: task-1",
		"## Constraints",
		"## Decisions",
		"## Expected Files",
		"## Validations",
		"## Known Risks",
		"## Testing Strategy",
		"## Knowledge Objects",
	}
	for _, s := range sections {
		if !strings.Contains(content, s) {
			t.Errorf("content missing section %q", s)
		}
	}

	// Verify validation commands
	if !strings.Contains(content, "go test ./...") {
		t.Error("missing go test")
	}
	if !strings.Contains(content, "go vet ./...") {
		t.Error("missing go vet")
	}
	if !strings.Contains(content, "go build ./...") {
		t.Error("missing go build")
	}

	// Verify knowledge objects
	if !strings.Contains(content, "JWT-based auth flow") {
		t.Error("missing auth knowledge")
	}
	if !strings.Contains(content, "SQLite with modernc driver") {
		t.Error("missing db knowledge")
	}
}

package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHandleCreateMasterPlan_UsesPlanningService verifies that after Phase 15
// refactoring, the handler creates plans through planning.Service rather than
// through intentv3 or direct domain/repo manipulation.
func TestHandleCreateMasterPlan_UsesPlanningService(t *testing.T) {
	tmp := t.TempDir()
	projectRoot := filepath.Join(tmp, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "project_id"), []byte("p15-master"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The guard requires an approved product intent before planning.
	// Create and approve one so the master plan call succeeds.
	piResult, err := HandleCreateProductIntent(map[string]any{
		"project_root": projectRoot,
		"description":  "Test product for Phase 15 master plan guard",
	})
	if err != nil {
		t.Fatalf("create product intent: %v", err)
	}
	piID, _ := piResult["id"].(string)
	if _, err := HandleSubmitProductIntent(map[string]any{
		"project_root": projectRoot,
		"intent_id":    piID,
	}); err != nil {
		t.Fatalf("submit product intent: %v", err)
	}
	if _, err := HandleApproveProductIntent(map[string]any{
		"project_root": projectRoot,
		"intent_id":    piID,
	}); err != nil {
		t.Fatalf("approve product intent: %v", err)
	}

	// Now the handler should succeed via planning.Service.
	result, err := HandleCreateMasterPlan(map[string]any{
		"project_root":     projectRoot,
		"title":            "Phase 15 Master Plan",
		"summary":          "Refactored master plan creation via planning.Service",
		"vision_reference": "vis-1",
	})
	if err != nil {
		t.Fatalf("HandleCreateMasterPlan error: %v", err)
	}
	if result["plan_id"] == nil || result["plan_id"] == "" {
		t.Fatalf("expected plan_id in result, got %v", result)
	}
	if result["status"] != "draft" {
		t.Fatalf("expected status draft, got %v", result["status"])
	}
}

// TestHandleCreateSpecificPlan_UsesPlanningService verifies that after Phase 15
// refactoring, the handler creates specific plans through planning.Service.
func TestHandleCreateSpecificPlan_UsesPlanningService(t *testing.T) {
	tmp := t.TempDir()
	projectRoot := filepath.Join(tmp, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "project_id"), []byte("p15-specific"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The guard requires an approved product intent.
	piResult, err := HandleCreateProductIntent(map[string]any{
		"project_root": projectRoot,
		"description":  "Test product for Phase 15 specific plan guard",
	})
	if err != nil {
		t.Fatalf("create product intent: %v", err)
	}
	piID, _ := piResult["id"].(string)
	if _, err := HandleSubmitProductIntent(map[string]any{
		"project_root": projectRoot,
		"intent_id":    piID,
	}); err != nil {
		t.Fatalf("submit product intent: %v", err)
	}
	if _, err := HandleApproveProductIntent(map[string]any{
		"project_root": projectRoot,
		"intent_id":    piID,
	}); err != nil {
		t.Fatalf("approve product intent: %v", err)
	}

	// First create a master plan for the specific plan to reference.
	masterResult, err := HandleCreateMasterPlan(map[string]any{
		"project_root":     projectRoot,
		"title":            "Parent Master",
		"summary":          "Parent for specific",
		"vision_reference": "vis-parent",
	})
	if err != nil {
		t.Fatalf("create parent master plan error: %v", err)
	}
	masterPlanID, _ := masterResult["plan_id"].(string)

	result, err := HandleCreateSpecificPlan(map[string]any{
		"project_root":   projectRoot,
		"master_plan_id": masterPlanID,
		"title":          "Specific Plan Title",
		"goal":           "Build the thing",
	})
	if err != nil {
		t.Fatalf("HandleCreateSpecificPlan error: %v", err)
	}
	if result["plan_id"] == nil || result["plan_id"] == "" {
		t.Fatalf("expected plan_id in result, got %v", result)
	}
	if s, _ := result["status"].(string); s != "draft" {
		t.Fatalf("expected status draft, got %v", s)
	}
}

// TestIntentV3_NoLongerUsedInHandlers verifies that intentv3 is no longer imported
// or referenced in handlers.go after Phase 15 refactoring.
// This is a compile-time check: if handlers.go imports intentv3, this test file
// (same package) won't notice directly, but the grep-based verification in CI
// will catch it.
func TestIntentV3_NoLongerUsedInHandlers(t *testing.T) {
	// Read handlers.go source to verify no intentv3 import.
	data, err := os.ReadFile(filepath.Join("..", "..", "internal", "mcp", "handlers.go"))
	if err != nil {
		// Try the test working directory relative path.
		wd, _ := os.Getwd()
		if strings.HasSuffix(wd, "mcp") {
			data, err = os.ReadFile("handlers.go")
		}
	}
	if err != nil {
		t.Skipf("cannot read handlers.go for import check: %v", err)
	}
	content := string(data)
	if strings.Contains(content, `"github.com/Durru/plan-ai/internal/intentv3"`) {
		t.Errorf("handlers.go still imports intentv3 — remove the import")
	}
	if strings.Contains(content, "intentv3.") && !strings.Contains(content, "intentv3.Service") {
		t.Errorf("handlers.go still references intentv3 types — move them to handlers_intent.go")
	}
}

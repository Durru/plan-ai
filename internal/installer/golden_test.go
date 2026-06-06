package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenDir is the directory containing golden test data.
var goldenDir = filepath.Join("testdata", "golden")

// updateGoldens, when set via GOLDEN_UPDATE=1, rewrites golden files
// with the current output.
const envGoldenUpdate = "GOLDEN_UPDATE"

func TestGoldenOpenCodeConfig_FreshInstall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping golden test in short mode")
	}

	// Use a stable path for reproducible output
	sandbox := t.TempDir()
	ocDir := filepath.Join(sandbox, "config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Fix the project root for reproducibility
	t.Setenv("PLAN_AI_PROJECT_ROOT", "/test/project")

	content, err := generateOpenCodeConfigContent(ocDir, "")
	if err != nil {
		t.Fatalf("generateOpenCodeConfigContent: %v", err)
	}

	goldenPath := filepath.Join(goldenDir, "opencode_fresh.json")

	if os.Getenv(envGoldenUpdate) == "1" {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, content, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v — run with GOLDEN_UPDATE=1 to create", goldenPath, err)
	}

	// Normalize paths for comparison so platform-specific differences
	// (like /tmp vs /var/folders) don't cause false failures.
	got := normalizeForComparison(string(content), ocDir)
	want := normalizeForComparison(string(golden), ocDir)

	if got != want {
		t.Fatalf("generated config doesn't match golden\n--- want (%s)\n+++ got\n%s", goldenPath, diffLines(want, got))
	}
}

func TestGoldenOpenCodeConfig_WithExisting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping golden test in short mode")
	}

	sandbox := t.TempDir()
	ocDir := filepath.Join(sandbox, "config", "opencode")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PLAN_AI_PROJECT_ROOT", "/test/project")

	// Create an existing config with a third-party MCP server
	existingPath := filepath.Join(ocDir, "opencode.json")
	if err := os.WriteFile(existingPath, []byte(`{"agent_name":"my-app","mcp":{"other":{"type":"local","command":["other-srv"]}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := generateOpenCodeConfigContent(ocDir, "")
	if err != nil {
		t.Fatalf("generateOpenCodeConfigContent: %v", err)
	}

	goldenPath := filepath.Join(goldenDir, "opencode_with_existing.json")

	if os.Getenv(envGoldenUpdate) == "1" {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, content, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v — run with GOLDEN_UPDATE=1 to create", goldenPath, err)
	}

	got := normalizeForComparison(string(content), ocDir)
	want := normalizeForComparison(string(golden), ocDir)

	if got != want {
		t.Fatalf("generated config doesn't match golden\n--- want (%s)\n+++ got\n%s", goldenPath, diffLines(want, got))
	}
}

// normalizeForComparison replaces the temp dir prefix and whitespace
// so golden comparisons are portable across machines.
func normalizeForComparison(s, ocDir string) string {
	// Replace the sandbox root with a stable marker
	const stableRoot = "/home/user"
	// Replace all occurrences of the ocDir path
	result := strings.ReplaceAll(s, ocDir, stableRoot)
	// The project root is already fixed by env var
	return strings.TrimSpace(result)
}

// diffLines returns a simple line diff of two strings.
func diffLines(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	var b strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}

	for i := 0; i < max; i++ {
		var wLine, gLine string
		if i < len(wantLines) {
			wLine = wantLines[i]
		}
		if i < len(gotLines) {
			gLine = gotLines[i]
		}
		if wLine != gLine {
			if wLine != "" {
				b.WriteString("- " + wLine + "\n")
			}
			if gLine != "" {
				b.WriteString("+ " + gLine + "\n")
			}
		}
	}
	return b.String()
}

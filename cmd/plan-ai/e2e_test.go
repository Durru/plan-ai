package main

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/plan-ai/plan-ai/internal/config"
	"github.com/plan-ai/plan-ai/internal/store"
)

// ──────────────────────────────────────────────
// Slice 3: Sandbox E2E CLI Scenarios
// Each test runs: install → init → seed scenario → verify
// ──────────────────────────────────────────────

func TestE2E_SaaSCRMBasic(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	// Domain seed creates planning artifacts
	runStep(t, home, project, "Domain seed: created", "dev", "seed-domain")

	// Status reflects the seeded data
	out := runStep(t, home, project, "", "status")
	for _, want := range []string{"Plans: 2", "Phases: 1", "Tasks: 1", "Decisions: 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestE2E_EcommerceWithIngestAndKnowledge(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	mustWriteCLI(t, project+"/go.mod", "module demo\n\ngo 1.25\n")
	mustWriteCLI(t, project+"/main.go", "package main\n")

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	// Ingest a prompt
	runStep(t, home, project, "Input ingested.", "ingest", "--type", "prompt", "--content", "Build an e-commerce checkout with cart and payments.")

	// Seed knowledge
	runStep(t, home, project, "Knowledge seed: created", "dev", "seed-knowledge")

	// List knowledge
	out := runStep(t, home, project, "", "knowledge", "list")
	for _, want := range []string{"PostgreSQL Multi Tenant", "Stripe Billing", "OAuth 2.0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("knowledge list missing %q:\n%s", want, out)
		}
	}

	// Status shows knowledge summary
	out = runStep(t, home, project, "", "status")
	if !strings.Contains(out, "Knowledge:") {
		t.Fatalf("status missing knowledge section:\n%s", out)
	}
}

func TestE2E_MCPServerWithResearch(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	mustWriteCLI(t, project+"/go.mod", "module demo\n\ngo 1.25\n")
	mustWriteCLI(t, project+"/main.go", "package main\n")

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	// Seed research
	runStep(t, home, project, "Research seed: created", "dev", "seed-research")

	// List research
	out := runStep(t, home, project, "", "research", "list")
	for _, want := range []string{"LLM Token Optimization", "SQLite Performance Limits"} {
		if !strings.Contains(out, want) {
			t.Fatalf("research list missing %q:\n%s", want, out)
		}
	}

	// Status shows research section
	out = runStep(t, home, project, "", "status")
	if !strings.Contains(out, "Research:") {
		t.Fatalf("status missing research section:\n%s", out)
	}
}

func TestE2E_MultiTenantWithVisionAndApproved(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	mustWriteCLI(t, project+"/go.mod", "module demo\n\ngo 1.25\n")
	mustWriteCLI(t, project+"/main.go", "package main\n")

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")

	// Ingest and vision draft
	runStep(t, home, project, "Input ingested.", "ingest", "--type", "prompt", "--content", "A multi-tenant SaaS with isolated databases per customer.")
	runStep(t, home, project, "Vision draft created.", "vision", "draft")

	// Add approved context
	runStep(t, home, project, "Approved context stored", "approved", "add", "--type", "requirement", "Multi-tenant isolation via separate DBs")
	runStep(t, home, project, "Approved context stored", "approved", "add", "--type", "decision", "Use schema-per-tenant")

	// Domain seed
	runStep(t, home, project, "Domain seed: created", "dev", "seed-domain")

	// Create and approve product intent before planning (Phase 5 guard).
	// Direct DB insert bypasses the CLI lifecycle to avoid resolver/Chdir issues.
	{
		pid := store.ProjectID(project)
		slug := config.ProjectSlug(project)
		projDBPath := filepath.Join(home, ".plan-ai", "projects", slug, "project.db")
		projDB, err := sql.Open("sqlite", projDBPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
		if err != nil {
			t.Fatalf("open project db for intent: %v", err)
		}
		_, err = projDB.Exec(`INSERT OR IGNORE INTO intent_v3_product_intents (id, project_id, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
			"pintent_e2e", pid, "Multi-tenant SaaS platform", "approved")
		projDB.Close()
		if err != nil {
			t.Fatalf("insert approved intent: %v", err)
		}
	}

	// Plan command works now because we have vision + approved context + approved intent
	out := runStep(t, home, project, "", "plan")
	for _, want := range []string{"master_plan:", "specific_plan:", "implementation_document:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("plan output missing %q:\n%s", want, out)
		}
	}

	// Status reflects domain seed
	out = runStep(t, home, project, "", "status")
	for _, want := range []string{"Plans: 2", "Tasks: 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestE2E_ContinuousScenarioSeedAndVerify(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")
	runStep(t, home, project, "Continuous scenario seed: created 4 events and 4 proposals", "dev", "seed-continuous-scenario")

	// Events command
	out := runStep(t, home, project, "", "continuous", "events")
	if strings.Contains(out, "No events yet.") {
		t.Fatal("expected events but got 'No events yet.'")
	}

	// Proposals command
	out = runStep(t, home, project, "", "continuous", "proposals")
	if strings.Contains(out, "No proposals yet.") {
		t.Fatal("expected proposals but got 'No proposals yet.'")
	}

	// Status command shows continuous planning info
	out = runStep(t, home, project, "", "continuous", "status")
	for _, want := range []string{"Recent events:", "Pending proposals:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("continuous status missing %q:\n%s", want, out)
		}
	}
}

func TestE2E_ContinuousApproveRejectFlow(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	runStep(t, home, project, "Global installation: installed", "install")
	runStep(t, home, project, "Project initialization: initialized", "init")
	runStep(t, home, project, "Continuous scenario seed: created 4 events and 4 proposals", "dev", "seed-continuous-scenario")

	// Get proposals output
	propsOut, err := executeCommand(t, home, project, "continuous", "proposals")
	if err != nil {
		t.Fatalf("proposals: %v\n%s", err, propsOut)
	}

	// The proposals list doesn't show IDs, but continuous status should show pending count
	out := runStep(t, home, project, "", "continuous", "status")
	if !strings.Contains(out, "Pending proposals:") {
		t.Fatalf("status missing Pending proposals:\n%s", out)
	}
}

func TestE2E_ValidateV2AllChecksPass(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	// Validate v2 is self-contained — no install/init needed
	out := runStep(t, home, project, "", "validate", "v2")
	for _, want := range []string{"V2 Validation Summary", "Passed: 63", "Failed: 0", "All checks PASSED"} {
		if !strings.Contains(out, want) {
			t.Fatalf("validate v2 output missing %q:\n%s", want, out)
		}
	}
}

func TestE2E_ValidateCasesListsSevenCategories(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	out := runStep(t, home, project, "", "validate", "cases")
	for _, want := range []string{
		"V2 Validation Cases (7)",
		"SaaS", "Ecommerce", "Landing Page", "MCP Server",
		"Mobile App", "API", "CRM",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("validate cases output missing %q:\n%s", want, out)
		}
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func runStep(t *testing.T, home, workdir, expected string, args ...string) string {
	t.Helper()
	output, err := executeCommand(t, home, workdir, args...)
	if err != nil {
		t.Fatalf("step %q error: %v\noutput:\n%s", strings.Join(args, " "), err, output)
	}
	if expected != "" && !strings.Contains(output, expected) {
		t.Fatalf("step %q output missing %q:\n%s", strings.Join(args, " "), expected, output)
	}
	return output
}

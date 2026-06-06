package main

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpserver "github.com/plan-ai/plan-ai/internal/mcp"

	_ "modernc.org/sqlite"
)

func TestInstallCommandCreatesGlobalStoreIdempotently(t *testing.T) {
	home := t.TempDir()

	for i := 0; i < 2; i++ {
		output, err := executeCommand(t, home, t.TempDir(), "install")
		if err != nil {
			t.Fatalf("install pass %d: %v\n%s", i+1, err, output)
		}
		if !strings.Contains(output, "Global installation: installed") {
			t.Fatalf("install output should summarize installation, got:\n%s", output)
		}
	}

	assertPathExists(t, filepath.Join(home, ".plan-ai", "config.json"))
	assertPathExists(t, filepath.Join(home, ".plan-ai", "global.db"))
	for _, name := range []string{"cache", "skills", "logs", "data", "backups"} {
		assertPathExists(t, filepath.Join(home, ".plan-ai", name))
	}
}

func TestInitCommandCreatesProjectStoreAndRegistersKnownProject(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	output, err := executeCommand(t, home, project, "init", "--local")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Project initialization: initialized") {
		t.Fatalf("init output should summarize initialization, got:\n%s", output)
	}
	assertPathExists(t, filepath.Join(project, ".plan-ai", "config.json"))
	assertPathExists(t, filepath.Join(project, ".plan-ai", "project.db"))
	for _, name := range []string{"cache", "snapshots", "exports", "docs", "locks", "backups"} {
		assertPathExists(t, filepath.Join(project, ".plan-ai", name))
	}

	db, err := sql.Open("sqlite", filepath.Join(home, ".plan-ai", "global.db"))
	if err != nil {
		t.Fatalf("open global db: %v", err)
	}
	defer db.Close()
	var got string
	if err := db.QueryRow(`SELECT path FROM known_projects WHERE path = ?`, project).Scan(&got); err != nil {
		t.Fatalf("known project was not registered: %v", err)
	}
	if got != project {
		t.Fatalf("known project path = %q, want %q", got, project)
	}
}

func TestInitCommandWorksWithoutGlobalInstallation(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	output, err := executeCommand(t, home, project, "init", "--local")
	if err != nil {
		t.Fatalf("init without global install: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Global installation: missing") {
		t.Fatalf("init should report missing global installation, got:\n%s", output)
	}
	assertPathExists(t, filepath.Join(project, ".plan-ai", "project.db"))
}

func TestStatusCommandReportsGlobalAndProjectRoutes(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	output, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, output)
	}

	for _, want := range []string{
		"Plan-AI v2.0.0",
		"Global installation: installed",
		"Project initialization: initialized",
		filepath.Join(home, ".plan-ai"),
		filepath.Join(home, ".plan-ai", "global.db"),
		filepath.Join(project, ".plan-ai"),
		filepath.Join(project, ".plan-ai", "project.db"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorDetectsSandboxOpenCodeConfigAfterSetup(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	opencodeDir := t.TempDir()
	env := map[string]string{"HOME": home, "OPENCODE_CONFIG_DIR": opencodeDir}

	if output, err := executeCommandWithEnv(t, env, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommandWithEnv(t, env, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	if output, err := executeCommandWithEnv(t, env, project, "setup", "opencode"); err != nil {
		t.Fatalf("setup opencode: %v\n%s", err, output)
	}

	output, err := executeCommandWithEnv(t, env, project, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, output)
	}
	for _, want := range []string{
		"Config found: " + filepath.Join(opencodeDir, "opencode.json"),
		"Agent: plan-ai",
		"[pass] OpenCode config detected, Plan-AI integration version compatible",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, output)
		}
	}
}

func TestBootstrapInitializesProjectAndOpenCodeIntegration(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	opencodeDir := t.TempDir()
	env := map[string]string{"HOME": home, "OPENCODE_CONFIG_DIR": opencodeDir}

	output, err := executeCommandWithEnv(t, env, project, "bootstrap")
	if err != nil {
		t.Fatalf("bootstrap: %v\n%s", err, output)
	}
	for _, want := range []string{
		"Global installation: installed",
		"Project initialization: initialized",
		"OpenCode integration artifacts generated.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("bootstrap output missing %q:\n%s", want, output)
		}
	}
	assertPathExists(t, filepath.Join(home, ".plan-ai", "global.db"))
	assertPathExists(t, filepath.Join(project, ".plan-ai", "project.db"))
	assertPathExists(t, filepath.Join(opencodeDir, "opencode.json"))
	assertPathExists(t, filepath.Join(opencodeDir, "mcp-registry.json"))
	assertPathExists(t, filepath.Join(opencodeDir, "agents", "plan-ai.json"))

	doctorOutput, err := executeCommandWithEnv(t, env, project, "doctor")
	if err != nil {
		t.Fatalf("doctor after bootstrap: %v\n%s", err, doctorOutput)
	}
	if !strings.Contains(doctorOutput, "Agent: plan-ai") {
		t.Fatalf("doctor should detect plan-ai agent after bootstrap:\n%s", doctorOutput)
	}
}

func TestRootCommandHasMCPServeSubcommand(t *testing.T) {
	cmd := newRootCommand()
	mcpCmd, _, err := cmd.Find([]string{"mcp", "serve"})
	if err != nil {
		t.Fatalf("find mcp serve: %v", err)
	}
	if mcpCmd == nil || mcpCmd.Name() != "serve" {
		t.Fatalf("expected mcp serve command, got %#v", mcpCmd)
	}
}

func TestMCPServeRespectsExplicitMinimalMode(t *testing.T) {
	srv, err := mcpserver.NewSDKServer(mcpserver.ToolContext{}, mcpserver.DefaultToolDependencies(), true)
	if err != nil {
		t.Fatalf("NewSDKServer: %v", err)
	}
	seen := map[string]bool{}
	for name := range srv.ListTools() {
		seen[name] = true
	}
	if !seen["plan_ai.project_status"] {
		t.Fatalf("minimal tools missing project_status: %#v", seen)
	}
	if seen["plan_ai.create_master_plan"] {
		t.Fatalf("minimal tools should not include create_master_plan: %#v", seen)
	}
}

func TestStatusCommandReportsEmptyDomainCountsForInitializedProject(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	output, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, output)
	}

	for _, want := range []string{
		"Plans: 0",
		"Phases: 0",
		"Tasks: 0",
		"Decisions: 0",
		"Research: 0",
		"Knowledge: 0",
		"Validations: 0",
		"Snapshots: 0",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output missing %q:\n%s", want, output)
		}
	}
}

func TestDevSeedDomainAndListDomainUseSandboxProjectStore(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	seedOutput, err := executeCommand(t, home, project, "dev", "seed-domain")
	if err != nil {
		t.Fatalf("seed domain: %v\n%s", err, seedOutput)
	}
	if !strings.Contains(seedOutput, "Domain seed: created") {
		t.Fatalf("seed output should summarize creation, got:\n%s", seedOutput)
	}

	listOutput, err := executeCommand(t, home, project, "dev", "list-domain")
	if err != nil {
		t.Fatalf("list domain: %v\n%s", err, listOutput)
	}
	statusOutput, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status after seed: %v\n%s", err, statusOutput)
	}

	for _, output := range []string{listOutput, statusOutput} {
		for _, want := range []string{
			"Plans: 2",
			"Phases: 1",
			"Tasks: 1",
			"Decisions: 1",
			"Research: 1",
			"Knowledge: 1",
			"Validations: 1",
			"Snapshots: 1",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("output missing %q:\n%s", want, output)
			}
		}
	}
}

func TestScanCommandRequiresInitializedProject(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	output, err := executeCommand(t, home, project, "scan")
	if err == nil {
		t.Fatalf("scan without init should fail, output:\n%s", output)
	}
	if !strings.Contains(err.Error(), "project is not initialized") {
		t.Fatalf("scan error = %v, output:\n%s", err, output)
	}
}

func TestScanCommandPersistsResultAndStatusReflectsIt(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	mustWriteCLI(t, filepath.Join(project, "go.mod"), `module demo

go 1.25

require (
	github.com/spf13/cobra v1.9.1
	modernc.org/sqlite v1.51.0
)
`)
	mustWriteCLI(t, filepath.Join(project, "README.md"), "# Demo\n")
	mustWriteCLI(t, filepath.Join(project, "main.go"), "package main\n")

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	scanOutput, err := executeCommand(t, home, project, "scan")
	if err != nil {
		t.Fatalf("scan: %v\n%s", err, scanOutput)
	}
	fingerprint := extractFingerprint(t, scanOutput)

	statusOutput, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, statusOutput)
	}
	for _, want := range []string{"Scan:", "latest:", "files indexed:", "git:", "Go", "Markdown"} {
		if !strings.Contains(statusOutput, want) {
			t.Fatalf("status missing %q:\n%s", want, statusOutput)
		}
	}
	if !strings.Contains(scanOutput, fingerprint) {
		t.Fatalf("scan output missing extracted fingerprint %q:\n%s", fingerprint, scanOutput)
	}

	db, err := sql.Open("sqlite", filepath.Join(project, ".plan-ai", "project.db"))
	if err != nil {
		t.Fatalf("open project db: %v", err)
	}
	defer db.Close()
	var stored string
	if err := db.QueryRow(`SELECT fingerprint FROM project_scans ORDER BY created_at DESC LIMIT 1`).Scan(&stored); err != nil {
		t.Fatalf("read stored fingerprint: %v", err)
	}
	if stored != fingerprint {
		t.Fatalf("stored fingerprint = %q, want %q", stored, fingerprint)
	}
}

func TestScanCommandIgnoresNodeModulesAndPlanAIDir(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	mustWriteCLI(t, filepath.Join(project, "main.go"), "package main\n")
	mustWriteCLI(t, filepath.Join(project, "node_modules", "foo.js"), "module.exports = 1\n")
	mustWriteCLI(t, filepath.Join(project, ".plan-ai", "bar.go"), "package leak\n")

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "scan"); err != nil {
		t.Fatalf("scan: %v\n%s", err, output)
	}

	db, err := sql.Open("sqlite", filepath.Join(project, ".plan-ai", "project.db"))
	if err != nil {
		t.Fatalf("open project db: %v", err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT path FROM project_scan_files`)
	if err != nil {
		t.Fatalf("query scan files: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			t.Fatalf("scan path: %v", err)
		}
		if strings.HasPrefix(path, "node_modules/") || strings.HasPrefix(path, ".plan-ai/") {
			t.Fatalf("ignored path persisted: %q", path)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
}

func TestEnvironmentOverridesHomeAndProjectRoot(t *testing.T) {
	realHome := t.TempDir()
	realWorkdir := t.TempDir()
	sandboxHome := t.TempDir()
	sandboxProject := t.TempDir()

	env := map[string]string{
		"HOME":                 realHome,
		"PLAN_AI_HOME":         sandboxHome,
		"PLAN_AI_PROJECT_ROOT": sandboxProject,
	}

	if output, err := executeCommandWithEnv(t, env, realWorkdir, "install"); err != nil {
		t.Fatalf("install with env overrides: %v\n%s", err, output)
	}
	if output, err := executeCommandWithEnv(t, env, realWorkdir, "init", "--local"); err != nil {
		t.Fatalf("init with env overrides: %v\n%s", err, output)
	}
	output, err := executeCommandWithEnv(t, env, realWorkdir, "status")
	if err != nil {
		t.Fatalf("status with env overrides: %v\n%s", err, output)
	}

	for _, want := range []string{
		filepath.Join(sandboxHome, ".plan-ai", "global.db"),
		filepath.Join(sandboxProject, ".plan-ai", "project.db"),
	} {
		assertPathExists(t, want)
		if !strings.Contains(output, want) {
			t.Fatalf("status output missing %q:\n%s", want, output)
		}
	}
	assertPathAbsent(t, filepath.Join(realHome, ".plan-ai"))
	assertPathAbsent(t, filepath.Join(realWorkdir, ".plan-ai"))
}

func TestVersionCommandOutputStaysExact(t *testing.T) {
	output, err := executeCommand(t, t.TempDir(), t.TempDir(), "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if output != "Plan-AI v2.0.0\n" {
		t.Fatalf("version output = %q", output)
	}
}

func TestKnowledgeAddListShowSearchAndReuseFlow(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	for _, args := range [][]string{
		{"install"}, {"init"},
	} {
		if output, err := executeCommand(t, home, project, args...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output)
		}
	}

	addOutput, err := executeCommand(t, home, project, "knowledge", "add",
		"--topic", "PostgreSQL Multi Tenant",
		"--summary", "Tenant isolation patterns",
		"--content", "Schema per tenant is a good default",
		"--tag", "postgres",
		"--tag", "multi-tenant",
		"--status", "approved",
		"--source", "manual",
		"--confidence", "0.9",
	)
	if err != nil {
		t.Fatalf("knowledge add: %v\n%s", err, addOutput)
	}
	for _, want := range []string{"Knowledge created.", "category: database", "status: approved", "tags: postgres, multi-tenant"} {
		if !strings.Contains(addOutput, want) {
			t.Fatalf("knowledge add output missing %q:\n%s", want, addOutput)
		}
	}

	listOutput, err := executeCommand(t, home, project, "knowledge", "list")
	if err != nil {
		t.Fatalf("knowledge list: %v\n%s", err, listOutput)
	}
	if !strings.Contains(listOutput, "PostgreSQL Multi Tenant") {
		t.Fatalf("list output missing topic:\n%s", listOutput)
	}

	listFiltered, err := executeCommand(t, home, project, "knowledge", "list", "--category", "architecture")
	if err != nil {
		t.Fatalf("knowledge list filtered: %v\n%s", err, listFiltered)
	}
	if !strings.Contains(listFiltered, "No knowledge objects yet.") {
		t.Fatalf("expected empty list for architecture filter, got:\n%s", listFiltered)
	}

	id := extractKnowledgeID(t, listOutput)

	showOutput, err := executeCommand(t, home, project, "knowledge", "show", id)
	if err != nil {
		t.Fatalf("knowledge show: %v\n%s", err, showOutput)
	}
	for _, want := range []string{"topic: PostgreSQL Multi Tenant", "category: database", "tags: multi-tenant, postgres", "summary: Tenant isolation patterns", "content:"} {
		if !strings.Contains(showOutput, want) {
			t.Fatalf("knowledge show output missing %q:\n%s", want, showOutput)
		}
	}

	searchOutput, err := executeCommand(t, home, project, "knowledge", "search", "tenant")
	if err != nil {
		t.Fatalf("knowledge search: %v\n%s", err, searchOutput)
	}
	if !strings.Contains(searchOutput, "PostgreSQL Multi Tenant") {
		t.Fatalf("search did not find topic:\n%s", searchOutput)
	}

	emptySearch, err := executeCommand(t, home, project, "knowledge", "search", "nothing-matches-this-needle")
	if err != nil {
		t.Fatalf("knowledge search empty: %v\n%s", err, emptySearch)
	}
	if !strings.Contains(emptySearch, "No knowledge matches") {
		t.Fatalf("expected empty search message, got:\n%s", emptySearch)
	}

	reuseOutput, err := executeCommand(t, home, project, "knowledge", "reuse", id)
	if err != nil {
		t.Fatalf("knowledge reuse: %v\n%s", err, reuseOutput)
	}
	if !strings.Contains(reuseOutput, "Knowledge reused: PostgreSQL Multi Tenant (reuse count: 1)") {
		t.Fatalf("reuse output unexpected:\n%s", reuseOutput)
	}

	statusOutput, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, statusOutput)
	}
	for _, want := range []string{"Knowledge:", "total: 1", "approved: 1", "reused: 1"} {
		if !strings.Contains(statusOutput, want) {
			t.Fatalf("status output missing %q:\n%s", want, statusOutput)
		}
	}
}

func TestKnowledgeAddRequiresTopic(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "knowledge", "add"); err == nil {
		t.Fatalf("expected error for missing --topic, got:\n%s", output)
	}
}

func TestKnowledgeCommandsRequireInitializedProject(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	// Phase 1: the resolver lazily provisions an external project store on
	// first access, so `knowledge list` on a fresh project returns the empty
	// state instead of an error. The semantic guarantee is now "an external
	// project store is auto-created and migrations are applied" rather than
	// "command fails before init."
	out, err := executeCommand(t, home, project, "knowledge", "list")
	if err != nil {
		t.Fatalf("knowledge list on fresh project: %v\n%s", err, out)
	}
	if !strings.Contains(out, "No knowledge objects yet.") {
		t.Fatalf("expected empty knowledge state, got:\n%s", out)
	}
}

func TestDevSeedKnowledgeCreatesSeededEntries(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	for _, args := range [][]string{
		{"install"}, {"init"},
	} {
		if output, err := executeCommand(t, home, project, args...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output)
		}
	}
	seedOutput, err := executeCommand(t, home, project, "dev", "seed-knowledge")
	if err != nil {
		t.Fatalf("dev seed-knowledge: %v\n%s", err, seedOutput)
	}
	if !strings.Contains(seedOutput, "Knowledge seed: created") {
		t.Fatalf("seed output missing summary:\n%s", seedOutput)
	}

	listOutput, err := executeCommand(t, home, project, "knowledge", "list")
	if err != nil {
		t.Fatalf("knowledge list: %v\n%s", err, listOutput)
	}
	for _, topic := range []string{"PostgreSQL Multi Tenant", "OAuth 2.0", "Stripe Billing"} {
		if !strings.Contains(listOutput, topic) {
			t.Fatalf("seed list missing %q:\n%s", topic, listOutput)
		}
	}

	statusOutput, err := executeCommand(t, home, project, "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, statusOutput)
	}
	for _, want := range []string{"total: 3", "approved: 3", "reused: 0"} {
		if !strings.Contains(statusOutput, want) {
			t.Fatalf("status after seed missing %q:\n%s", want, statusOutput)
		}
	}
}

func extractKnowledgeID(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "knowledge_") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	t.Fatalf("knowledge id not found in output:\n%s", output)
	return ""
}

func executeCommand(t *testing.T, home, workdir string, args ...string) (string, error) {
	t.Helper()
	// Always point the OpenCode config at a sandbox under the test HOME so
	// the installer's refuse-to-write-real-config safety check does not
	// trip in tests. Callers that want to drive OPENCODE_CONFIG_DIR
	// themselves should use executeCommandWithEnv.
	ocDir := filepath.Join(home, ".config", "opencode")
	return executeCommandWithEnv(t, map[string]string{
		"HOME":                home,
		"OPENCODE_CONFIG_DIR": ocDir,
	}, workdir, args...)
}

func executeCommandWithEnv(t *testing.T, env map[string]string, workdir string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("PLAN_AI_HOME", "")
	t.Setenv("PLAN_AI_PROJECT_ROOT", "")
	for key, value := range env {
		t.Setenv(key, value)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	cmd := newRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), err
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertPathAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be absent, err=%v", path, err)
	}
}

func mustWriteCLI(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestIntentV3Discover(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	output, err := executeCommand(t, home, project, "intent", "discover", "Build a CLI tool for Go developers")
	if err != nil {
		t.Fatalf("intent discover: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Discovery Result:") {
		t.Fatalf("discover output should contain 'Discovery Result:', got:\n%s", output)
	}
	if !strings.Contains(output, "Detected Intent:") {
		t.Fatalf("discover output should contain 'Detected Intent:', got:\n%s", output)
	}
}

func TestIntentV3CreateListShowApprove(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	// Create a product intent
	output, err := executeCommand(t, home, project,
		"intent", "create",
		"--description", "Build a CLI tool",
		"--expected-outcome", "Happy developers",
		"--desired-experience", "Fast iteration",
	)
	if err != nil {
		t.Fatalf("intent create: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Product Intent created: pintent_") {
		t.Fatalf("create output should contain 'Product Intent created: pintent_', got:\n%s", output)
	}

	// Extract the ID
	var intentID string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Product Intent created:") {
			parts := strings.Split(line, "pintent_")
			if len(parts) == 2 {
				intentID = "pintent_" + parts[1]
			}
		}
	}
	if intentID == "" {
		t.Fatalf("could not extract intent ID from output:\n%s", output)
	}

	// List
	output, err = executeCommand(t, home, project, "intent", "list")
	if err != nil {
		t.Fatalf("intent list: %v\n%s", err, output)
	}
	if !strings.Contains(output, intentID) {
		t.Fatalf("list output should contain the created intent ID %q, got:\n%s", intentID, output)
	}

	// Show
	output, err = executeCommand(t, home, project, "intent", "show", intentID)
	if err != nil {
		t.Fatalf("intent show: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Product Intent:") {
		t.Fatalf("show output should contain 'Product Intent:', got:\n%s", output)
	}
	if !strings.Contains(output, "Build a CLI tool") {
		t.Fatalf("show output should contain description, got:\n%s", output)
	}

	// Submit for approval (draft -> pending_approval)
	output, err = executeCommand(t, home, project, "intent", "submit", intentID)
	if err != nil {
		t.Fatalf("intent submit: %v\n%s", err, output)
	}
	if !strings.Contains(output, "submitted for approval") {
		t.Fatalf("submit output should contain 'submitted for approval', got:\n%s", output)
	}

	// Approve (pending_approval -> approved)
	output, err = executeCommand(t, home, project, "intent", "approve", intentID)
	if err != nil {
		t.Fatalf("intent approve: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Product Intent approved:") {
		t.Fatalf("approve output should contain 'Product Intent approved:', got:\n%s", output)
	}
}

func TestIntentV3ListEmpty(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	output, err := executeCommand(t, home, project, "intent", "list")
	if err != nil {
		t.Fatalf("intent list (empty): %v\n%s", err, output)
	}
	if !strings.Contains(output, "No V3 product intents found.") {
		t.Fatalf("list output should contain 'No V3 product intents found.', got:\n%s", output)
	}
}

func TestConfidenceEvaluateIntent(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()

	if output, err := executeCommand(t, home, project, "install"); err != nil {
		t.Fatalf("install: %v\n%s", err, output)
	}
	if output, err := executeCommand(t, home, project, "init", "--local"); err != nil {
		t.Fatalf("init: %v\n%s", err, output)
	}

	output, err := executeCommand(t, home, project,
		"intent", "create",
		"--description", "Build a CLI tool",
		"--expected-outcome", "Happy developers",
		"--desired-experience", "Fast iteration",
		"--desired-result", "Developers ship faster",
		"--success-definition", "Developers complete setup",
		"--failure-definition", "Developers abandon the tool",
	)
	if err != nil {
		t.Fatalf("intent create: %v\n%s", err, output)
	}
	intentID := extractProductIntentID(t, output)

	output, err = executeCommand(t, home, project, "confidence", "evaluate", "--intent", intentID)
	if err != nil {
		t.Fatalf("confidence evaluate: %v\n%s", err, output)
	}
	for _, want := range []string{"Intent Confidence Report", "Intent Confidence:", "Intent Score:", "Vision Score:", "UX Score:", "Business Score:", "Requirements Score:", "Constraints Score:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("confidence output missing %q:\n%s", want, output)
		}
	}
}

func extractProductIntentID(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Product Intent created:") {
			parts := strings.Split(line, "pintent_")
			if len(parts) == 2 {
				return "pintent_" + strings.TrimSpace(parts[1])
			}
		}
	}
	t.Fatalf("could not extract product intent ID from output:\n%s", output)
	return ""
}

func extractFingerprint(t *testing.T, output string) string {
	t.Helper()
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "Fingerprint:" && i+1 < len(lines) {
			fingerprint := strings.TrimSpace(lines[i+1])
			if len(fingerprint) != 32 {
				t.Fatalf("fingerprint length = %d in output:\n%s", len(fingerprint), output)
			}
			return fingerprint
		}
	}
	t.Fatalf("fingerprint not found in output:\n%s", output)
	return ""
}

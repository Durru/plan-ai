package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestDetectGitFindsBranchFromHead(t *testing.T) {
	root := t.TempDir()
	gitDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	detected, branch := DetectGit(root)
	if !detected {
		t.Fatalf("expected detected=true, got false")
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}
}

func TestDetectGitReturnsFalseWhenNoGit(t *testing.T) {
	root := t.TempDir()
	detected, branch := DetectGit(root)
	if detected {
		t.Fatalf("expected detected=false on plain directory, got %+v", detected)
	}
	if branch != "" {
		t.Fatalf("branch = %q, want empty", branch)
	}
}

func TestScanIntegration(t *testing.T) {
	root := t.TempDir()

	// git init fake.
	gitDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feature/test\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	// Realistic project layout.
	mustWrite(t, filepath.Join(root, "go.mod"), `module demo

go 1.25

require (
	github.com/spf13/cobra v1.9.1
	modernc.org/sqlite v1.51.0
)
`)
	mustWrite(t, filepath.Join(root, "package.json"), `{
  "name": "demo",
  "dependencies": {
    "react": "18.3.1"
  }
}`)
	mustWrite(t, filepath.Join(root, "README.md"), "# Demo\n")
	mustWrite(t, filepath.Join(root, "main.go"), "package main\n")
	mustMkdir(t, filepath.Join(root, "node_modules"))
	mustWrite(t, filepath.Join(root, "node_modules", "a.js"), "module.exports = 1\n")

	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if !scan.GitDetected {
		t.Fatalf("GitDetected = false, want true")
	}
	if scan.GitBranch != "feature/test" {
		t.Fatalf("GitBranch = %q, want feature/test", scan.GitBranch)
	}

	// Languages should include Go and Markdown.
	langs := map[string]int{}
	for _, l := range scan.Languages {
		langs[l.Language] = l.Files
	}
	if langs["Go"] < 1 {
		t.Fatalf("expected Go >= 1, got %+v", langs)
	}
	if langs["Markdown"] < 1 {
		t.Fatalf("expected Markdown >= 1, got %+v", langs)
	}

	// Frameworks should include Go and Cobra at minimum.
	frameworkNames := make([]string, 0, len(scan.Frameworks))
	for _, f := range scan.Frameworks {
		frameworkNames = append(frameworkNames, f.Name)
	}
	sort.Strings(frameworkNames)
	if !containsString(frameworkNames, "Go") {
		t.Fatalf("expected Go framework, got %v", frameworkNames)
	}
	if !containsString(frameworkNames, "Cobra") {
		t.Fatalf("expected Cobra framework, got %v", frameworkNames)
	}

	// Package managers should include "go" at minimum (go.mod is the spec's
	// signal for the Go package manager; a bare package.json is not).
	pms := make([]string, 0, len(scan.PackageManagers))
	for _, pm := range scan.PackageManagers {
		pms = append(pms, pm.Name)
	}
	sort.Strings(pms)
	if !containsString(pms, "go") {
		t.Fatalf("expected go package manager, got %v", pms)
	}

	// node_modules/a.js must NOT be in the file list.
	for _, f := range scan.Files {
		if strings.HasPrefix(f.Path, "node_modules/") {
			t.Fatalf("node_modules file leaked into scan: %q", f.Path)
		}
	}

	// Fingerprint is non-empty and hex.
	if len(scan.Fingerprint) != 32 {
		t.Fatalf("fingerprint length = %d, want 32", len(scan.Fingerprint))
	}

	// Files should at least include go.mod, package.json, README.md, main.go.
	paths := map[string]bool{}
	for _, f := range scan.Files {
		paths[f.Path] = true
	}
	for _, want := range []string{"go.mod", "package.json", "README.md", "main.go"} {
		if !paths[want] {
			t.Fatalf("expected %q in files, got %v", want, scan.Files)
		}
	}

	// Summary is non-empty.
	if scan.Summary == "" {
		t.Fatalf("summary is empty")
	}
}

func TestScanIgnoresHiddenPlanAIDir(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, ".plan-ai"))
	mustWrite(t, filepath.Join(root, ".plan-ai", "leak.go"), "package planai\n")
	mustWrite(t, filepath.Join(root, "main.go"), "package main\n")

	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	for _, f := range scan.Files {
		if strings.HasPrefix(f.Path, ".plan-ai/") {
			t.Fatalf(".plan-ai file leaked: %q", f.Path)
		}
	}
}

func TestScanSummaryHasExpectedShape(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.go"), "package a\n")
	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !strings.Contains(scan.Summary, "1 files") {
		t.Fatalf("summary = %q, want 1 files", scan.Summary)
	}
	if !strings.Contains(scan.Summary, "1 languages") {
		t.Fatalf("summary = %q, want 1 languages", scan.Summary)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestSortSliceStableByName(t *testing.T) {
	items := []string{"b", "a", "c"}
	sort.Strings(items)
	if !reflect.DeepEqual(items, []string{"a", "b", "c"}) {
		t.Fatalf("sorted = %v", items)
	}
}

package store

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestScanRepositoryCreateAndReadLatest(t *testing.T) {
	db := openTestProjectDB(t)
	repo := NewScanRepository(db)
	createdAt := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	id, err := repo.CreateScan(Scan{
		ProjectRoot: "/tmp/project",
		GitDetected: true,
		GitBranch:   "main",
		Fingerprint: "abc123",
		Summary:     "2 files, 2 languages, 1 frameworks, 1 dependencies",
		CreatedAt:   createdAt,
		Languages: []ScanLanguage{
			{Language: "Go", FilesCount: 1},
			{Language: "Markdown", FilesCount: 1},
		},
		Frameworks:      []ScanFramework{{Framework: "Cobra", Evidence: "go.mod:github.com/spf13/cobra"}},
		PackageManagers: []ScanPackageManager{{Manager: "go", Evidence: "go.mod"}},
		Dependencies:    []ScanDependency{{Name: "github.com/spf13/cobra", Version: "v1.9.1", Source: "go.mod"}},
		Files: []ScanFile{
			{Path: "README.md", Kind: "doc", SizeBytes: 7},
			{Path: "main.go", Kind: "source", SizeBytes: 13},
		},
	})
	if err != nil {
		t.Fatalf("CreateScan: %v", err)
	}
	if id == "" {
		t.Fatalf("CreateScan returned empty id")
	}

	latest, err := repo.GetLatestScan()
	if err != nil {
		t.Fatalf("GetLatestScan: %v", err)
	}
	if latest.ID != id {
		t.Fatalf("latest id = %q, want %q", latest.ID, id)
	}
	if !latest.GitDetected || latest.GitBranch != "main" || latest.Fingerprint != "abc123" {
		t.Fatalf("latest scan mismatch: %+v", latest)
	}
	if len(latest.Languages) != 2 || len(latest.Files) != 2 || len(latest.Dependencies) != 1 {
		t.Fatalf("children were not loaded: %+v", latest)
	}
}

func TestScanRepositorySummaryAndLists(t *testing.T) {
	db := openTestProjectDB(t)
	repo := NewScanRepository(db)

	oldID, err := repo.CreateScan(Scan{ProjectRoot: "/tmp/old", Fingerprint: "old", Summary: "old", CreatedAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("CreateScan old: %v", err)
	}
	newID, err := repo.CreateScan(Scan{
		ProjectRoot:     "/tmp/new",
		GitDetected:     false,
		Fingerprint:     "new",
		Summary:         "summary",
		CreatedAt:       time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		Languages:       []ScanLanguage{{Language: "Go", FilesCount: 2}},
		Frameworks:      []ScanFramework{{Framework: "Go", Evidence: "go.mod"}},
		PackageManagers: []ScanPackageManager{{Manager: "go", Evidence: "go.mod"}},
		Dependencies:    []ScanDependency{{Name: "modernc.org/sqlite", Version: "v1.51.0", Source: "go.mod"}},
		Files:           []ScanFile{{Path: "go.mod", Kind: "config", SizeBytes: 12}},
	})
	if err != nil {
		t.Fatalf("CreateScan new: %v", err)
	}
	if oldID == newID {
		t.Fatalf("scan ids should differ")
	}

	summary, err := repo.GetScanSummary()
	if err != nil {
		t.Fatalf("GetScanSummary: %v", err)
	}
	if summary.ID != newID || summary.ProjectRoot != "/tmp/new" || summary.FileCount != 1 {
		t.Fatalf("summary mismatch: %+v", summary)
	}
	if !reflect.DeepEqual(summary.LanguageNames, []string{"Go"}) {
		t.Fatalf("language names = %+v", summary.LanguageNames)
	}
	if !reflect.DeepEqual(summary.FrameworkNames, []string{"Go"}) {
		t.Fatalf("framework names = %+v", summary.FrameworkNames)
	}
	if !reflect.DeepEqual(summary.PackageManagerNames, []string{"go"}) {
		t.Fatalf("pm names = %+v", summary.PackageManagerNames)
	}

	languages, err := repo.ListLanguages(newID)
	if err != nil || len(languages) != 1 || languages[0].Language != "Go" {
		t.Fatalf("ListLanguages = %+v, %v", languages, err)
	}
	frameworks, err := repo.ListFrameworks(newID)
	if err != nil || len(frameworks) != 1 || frameworks[0].Framework != "Go" {
		t.Fatalf("ListFrameworks = %+v, %v", frameworks, err)
	}
	managers, err := repo.ListPackageManagers(newID)
	if err != nil || len(managers) != 1 || managers[0].Manager != "go" {
		t.Fatalf("ListPackageManagers = %+v, %v", managers, err)
	}
	deps, err := repo.ListDependencies(newID)
	if err != nil || len(deps) != 1 || deps[0].Name != "modernc.org/sqlite" {
		t.Fatalf("ListDependencies = %+v, %v", deps, err)
	}
	files, err := repo.ListFiles(newID)
	if err != nil || len(files) != 1 || files[0].Path != "go.mod" {
		t.Fatalf("ListFiles = %+v, %v", files, err)
	}
}

func TestScanRepositoryLatestNoRows(t *testing.T) {
	db := openTestProjectDB(t)
	repo := NewScanRepository(db)
	if _, err := repo.GetLatestScan(); err != sql.ErrNoRows {
		t.Fatalf("GetLatestScan err = %v, want sql.ErrNoRows", err)
	}
	if _, err := repo.GetScanSummary(); err != sql.ErrNoRows {
		t.Fatalf("GetScanSummary err = %v, want sql.ErrNoRows", err)
	}
}

func openTestProjectDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(t.TempDir() + "/project.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("RunProjectMigrations: %v", err)
	}
	return db
}

package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plan-ai/plan-ai/internal/config"
)

// newTestResolver builds a ProjectResolver against a temp home directory and
// temp global database. The global DB has the global migrations applied.
func newTestResolver(t *testing.T) (*ProjectResolver, *sql.DB) {
	t.Helper()
	home := t.TempDir()
	globalDBPath := filepath.Join(home, "global.db")
	globalDB, err := Open(globalDBPath)
	if err != nil {
		t.Fatalf("open global db: %v", err)
	}
	t.Cleanup(func() { _ = globalDB.Close() })
	if err := RunGlobalMigrations(globalDB); err != nil {
		t.Fatalf("run global migrations: %v", err)
	}
	return NewProjectResolver(home, globalDB), globalDB
}

func TestEnsureExternalProjectLayoutCreatesExpectedPaths(t *testing.T) {
	home := t.TempDir()
	layout, err := EnsureExternalProjectLayout(home, "test_slug")
	if err != nil {
		t.Fatalf("EnsureExternalProjectLayout: %v", err)
	}

	// EnsureExternalProjectLayout creates the on-disk directories but not
	// the config.json file itself (that is written later when the project
	// is actually registered). Verify each directory path exists and the
	// parent of ConfigPath exists.
	dirPaths := []string{
		layout.Dir,
		layout.CacheDir,
		layout.SnapshotsDir,
		layout.ExportsDir,
		layout.DocsDir,
		layout.LocksDir,
		layout.BackupsDir,
	}
	for _, p := range dirPaths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Dir(layout.ConfigPath)); err != nil {
		t.Errorf("expected parent of ConfigPath %q to exist: %v", layout.ConfigPath, err)
	}

	expectedDBPath := config.ExternalProjectDBPath(home, "test_slug")
	if layout.DBPath != expectedDBPath {
		t.Errorf("layout.DBPath = %q, want %q", layout.DBPath, expectedDBPath)
	}

	db, err := Open(layout.DBPath)
	if err != nil {
		t.Fatalf("open layout db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Errorf("db.Ping failed: %v", err)
	}
}

func TestProjectResolverReturnsExternalLocationByDefault(t *testing.T) {
	resolver, globalDB := newTestResolver(t)
	rootPath := t.TempDir()
	slug := config.ProjectSlug(rootPath)

	repo := NewProjectRegistryRepository(globalDB)
	if _, err := repo.Register(ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
		Slug:     slug,
		Mode:     config.ProjectModeExternal,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	loc, err := resolver.Resolve(rootPath)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if loc.Mode != config.ProjectModeExternal {
		t.Errorf("Mode = %q, want %q", loc.Mode, config.ProjectModeExternal)
	}
	if loc.DBPath != config.ExternalProjectDBPath(resolver.HomeDir(), slug) {
		t.Errorf("DBPath = %q, want %q", loc.DBPath, config.ExternalProjectDBPath(resolver.HomeDir(), slug))
	}
	if loc.ProjectID != ProjectID(rootPath) {
		t.Errorf("ProjectID = %q, want %q", loc.ProjectID, ProjectID(rootPath))
	}
	if loc.Slug == "" {
		t.Errorf("Slug is empty")
	}
	expectedDir := config.ExternalProjectDir(resolver.HomeDir(), slug)
	if loc.Layout.Dir != expectedDir {
		t.Errorf("Layout.Dir = %q, want %q", loc.Layout.Dir, expectedDir)
	}
}

func TestProjectResolverReturnsLocalLocationForLocalMode(t *testing.T) {
	resolver, globalDB := newTestResolver(t)
	rootPath := t.TempDir()
	slug := config.ProjectSlug(rootPath)

	repo := NewProjectRegistryRepository(globalDB)
	if _, err := repo.Register(ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
		Slug:     slug,
		Mode:     config.ProjectModeLocal,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	loc, err := resolver.Resolve(rootPath)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if loc.Mode != config.ProjectModeLocal {
		t.Errorf("Mode = %q, want %q", loc.Mode, config.ProjectModeLocal)
	}
	if loc.DBPath != config.ProjectDBPath(rootPath) {
		t.Errorf("DBPath = %q, want %q", loc.DBPath, config.ProjectDBPath(rootPath))
	}
	if loc.Layout.Dir != config.ProjectDir(rootPath) {
		t.Errorf("Layout.Dir = %q, want %q", loc.Layout.Dir, config.ProjectDir(rootPath))
	}
}

func TestProjectResolverDetectsLegacyLocalStore(t *testing.T) {
	resolver, _ := newTestResolver(t)
	rootPath := t.TempDir()

	// Create a legacy local store at <root>/.plan-ai/project.db and run
	// project migrations on it so it is a valid SQLite file.
	if _, err := EnsureProjectLayout(rootPath); err != nil {
		t.Fatalf("EnsureProjectLayout: %v", err)
	}
	dbPath := config.ProjectDBPath(rootPath)
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if err := RunProjectMigrations(db); err != nil {
		db.Close()
		t.Fatalf("RunProjectMigrations: %v", err)
	}
	db.Close()

	_, err = resolver.Resolve(rootPath)
	if !errors.Is(err, ErrLegacyLocalStoreFound) {
		t.Fatalf("Resolve err = %v, want ErrLegacyLocalStoreFound", err)
	}
}

func TestProjectResolverReturnsDraftExternalWhenUnregisteredAndNoLegacy(t *testing.T) {
	resolver, _ := newTestResolver(t)
	rootPath := t.TempDir()

	loc, err := resolver.Resolve(rootPath)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if loc.Mode != config.ProjectModeExternal {
		t.Errorf("Mode = %q, want %q", loc.Mode, config.ProjectModeExternal)
	}
	slug := config.ProjectSlug(rootPath)
	if loc.Slug != slug {
		t.Errorf("Slug = %q, want %q", loc.Slug, slug)
	}
	expectedDir := config.ExternalProjectDir(resolver.HomeDir(), slug)
	if loc.Layout.Dir != expectedDir {
		t.Errorf("Layout.Dir = %q, want %q", loc.Layout.Dir, expectedDir)
	}
	if loc.DBPath != config.ExternalProjectDBPath(resolver.HomeDir(), slug) {
		t.Errorf("DBPath = %q, want %q", loc.DBPath, config.ExternalProjectDBPath(resolver.HomeDir(), slug))
	}
}

func TestProjectResolverOpenStoreOpensDBAndTouchesLastSeen(t *testing.T) {
	resolver, globalDB := newTestResolver(t)
	rootPath := t.TempDir()
	slug := config.ProjectSlug(rootPath)

	repo := NewProjectRegistryRepository(globalDB)
	if _, err := repo.Register(ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
		Slug:     slug,
		Mode:     config.ProjectModeExternal,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	db, loc, err := resolver.OpenStore(rootPath)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	if db == nil {
		t.Fatalf("OpenStore returned nil *sql.DB")
	}
	if loc.ProjectID == "" {
		t.Errorf("loc.ProjectID is empty")
	}
	if loc.Mode != config.ProjectModeExternal {
		t.Errorf("loc.Mode = %q, want %q", loc.Mode, config.ProjectModeExternal)
	}
	_ = db.Close()

	// Give the second-precision clock a moment, then verify TouchLastSeen
	// updated the row. We use the path lookup (not GetByID) because the
	// resolver's path canonicalization may produce a different absolute path
	// than what was passed in.
	enteredAt, err := repo.GetBySlug(slug)
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if enteredAt.LastSeenAt.IsZero() {
		t.Fatalf("LastSeenAt is zero after OpenStore")
	}
	if time.Since(enteredAt.LastSeenAt) > 10*time.Second {
		t.Errorf("LastSeenAt = %v is not recent (should be within 10s)", enteredAt.LastSeenAt)
	}
}

func TestRegisterProjectCreatesStableIDAcrossResolverCalls(t *testing.T) {
	resolver, globalDB := newTestResolver(t)
	rootPath := t.TempDir()
	slug := config.ProjectSlug(rootPath)

	repo := NewProjectRegistryRepository(globalDB)
	if _, err := repo.Register(ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     filepath.Base(rootPath),
		RootPath: rootPath,
		Slug:     slug,
		Mode:     config.ProjectModeExternal,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	loc1, err := resolver.Resolve(rootPath)
	if err != nil {
		t.Fatalf("resolve 1: %v", err)
	}
	loc2, err := resolver.Resolve(rootPath)
	if err != nil {
		t.Fatalf("resolve 2: %v", err)
	}

	if loc1.ProjectID != loc2.ProjectID {
		t.Errorf("ProjectID not stable: %q vs %q", loc1.ProjectID, loc2.ProjectID)
	}
	if loc1.Slug != loc2.Slug {
		t.Errorf("Slug not stable: %q vs %q", loc1.Slug, loc2.Slug)
	}
	if loc1.ProjectID == "" {
		t.Errorf("ProjectID is empty")
	}
	if loc1.Slug == "" {
		t.Errorf("Slug is empty")
	}
}

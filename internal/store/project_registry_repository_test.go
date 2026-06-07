package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/config"
)

func newTestRegistry(t *testing.T) (*sql.DB, *ProjectRegistryRepository) {
	t.Helper()
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "global.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open global db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := RunGlobalMigrations(db); err != nil {
		t.Fatalf("run global migrations: %v", err)
	}
	return db, NewProjectRegistryRepository(db)
}

func TestRegisterProjectCreatesStableID(t *testing.T) {
	_, repo := newTestRegistry(t)
	rootPath := "/tmp/myproj"
	expectedID := ProjectID(rootPath)
	slug := config.ProjectSlug(rootPath)

	entry := ProjectRegistryEntry{
		ID:       expectedID,
		Name:     "myproj",
		RootPath: rootPath,
		Slug:     slug,
		Mode:     config.ProjectModeExternal,
	}
	if _, err := repo.Register(entry); err != nil {
		t.Fatalf("register: %v", err)
	}

	got, err := repo.GetByPath(rootPath)
	if err != nil {
		t.Fatalf("get by path: %v", err)
	}
	if got.ID != expectedID {
		t.Errorf("ID = %q, want %q", got.ID, expectedID)
	}
	if got.Name != "myproj" {
		t.Errorf("Name = %q, want %q", got.Name, "myproj")
	}
	if got.RootPath != rootPath {
		t.Errorf("RootPath = %q, want %q", got.RootPath, rootPath)
	}
	if got.Slug != slug {
		t.Errorf("Slug = %q, want %q", got.Slug, slug)
	}
	if got.Mode != config.ProjectModeExternal {
		t.Errorf("Mode = %q, want %q", got.Mode, config.ProjectModeExternal)
	}

	gotByID, err := repo.GetByID(expectedID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if gotByID.ID != expectedID {
		t.Errorf("GetByID ID = %q, want %q", gotByID.ID, expectedID)
	}

	gotBySlug, err := repo.GetBySlug(slug)
	if err != nil {
		t.Fatalf("get by slug: %v", err)
	}
	if gotBySlug.Slug != slug {
		t.Errorf("GetBySlug Slug = %q, want %q", gotBySlug.Slug, slug)
	}

	list, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List len = %d, want 1", len(list))
	}
	if list[0].ID != expectedID {
		t.Errorf("List[0].ID = %q, want %q", list[0].ID, expectedID)
	}
}

func TestRegisterProjectUpdatesLastSeen(t *testing.T) {
	_, repo := newTestRegistry(t)
	rootPath := "/tmp/myproj"
	entry := ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     "myproj",
		RootPath: rootPath,
		Slug:     config.ProjectSlug(rootPath),
		Mode:     config.ProjectModeExternal,
	}
	registered, err := repo.Register(entry)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	firstSeen := registered.LastSeenAt

	// TouchLastSeen formats with RFC3339 (second precision), so we must
	// sleep more than a second to ensure the new timestamp is strictly newer
	// than the one captured at Register time.
	time.Sleep(1100 * time.Millisecond)

	if err := repo.TouchLastSeen(registered.ID); err != nil {
		t.Fatalf("touch last seen: %v", err)
	}

	got, err := repo.GetByPath(rootPath)
	if err != nil {
		t.Fatalf("get by path: %v", err)
	}
	if got.LastSeenAt.IsZero() {
		t.Fatalf("LastSeenAt is zero")
	}
	if !got.LastSeenAt.After(firstSeen) {
		t.Errorf("LastSeenAt = %v, want > %v", got.LastSeenAt, firstSeen)
	}
	if time.Since(got.LastSeenAt) > 5*time.Second {
		t.Errorf("LastSeenAt = %v is not recent (should be within 5s)", got.LastSeenAt)
	}
}

func TestUnregisterRemovesEntry(t *testing.T) {
	_, repo := newTestRegistry(t)
	rootPath := "/tmp/myproj"
	entry := ProjectRegistryEntry{
		ID:       ProjectID(rootPath),
		Name:     "myproj",
		RootPath: rootPath,
		Slug:     config.ProjectSlug(rootPath),
		Mode:     config.ProjectModeExternal,
	}
	registered, err := repo.Register(entry)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := repo.Unregister(registered.ID); err != nil {
		t.Fatalf("unregister: %v", err)
	}

	_, err = repo.GetByPath(rootPath)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetByPath after Unregister err = %v, want sql.ErrNoRows", err)
	}
}

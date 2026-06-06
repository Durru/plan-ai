package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plan-ai/plan-ai/internal/config"
)

// ProjectLocation describes a resolved project: its identity, the on-disk
// layout, the mode it was registered in, and the on-disk path of the SQLite
// database to open.
type ProjectLocation struct {
	ProjectID string
	Name      string
	RootPath  string
	Slug      string
	Mode      string
	Layout    ProjectLocationLayout
	DBPath    string
}

// ProjectLocationLayout captures the on-disk directories for a project, so
// callers can operate on the project without caring whether it lives in
// external (global) or local mode.
type ProjectLocationLayout struct {
	Dir          string
	ConfigPath   string
	CacheDir     string
	SnapshotsDir string
	ExportsDir   string
	DocsDir      string
	LocksDir     string
	BackupsDir   string
}

// ErrLegacyLocalStoreFound is returned by ProjectResolver.Resolve when a
// legacy <root>/.plan-ai/project.db exists but the project is not yet
// registered (or is registered in external mode). Callers should treat this
// as an explicit prompt: either re-run with `plan-ai init --local` or run
// `plan-ai migrate local-to-global`.
var ErrLegacyLocalStoreFound = errors.New("project resolver: legacy project-local store found at <root>/.plan-ai/project.db — use `plan-ai init --local` to keep it or `plan-ai migrate local-to-global` to migrate it")

// ProjectResolver decides where a given project root should store its data
// (external by default, local only if explicitly registered as such) and
// produces a ProjectLocation describing the chosen layout.
type ProjectResolver struct {
	homeDir string
	db      *sql.DB
}

// NewProjectResolver constructs a resolver backed by the given global DB
// (the database that holds known_projects).
func NewProjectResolver(homeDir string, globalDB *sql.DB) *ProjectResolver {
	return &ProjectResolver{homeDir: homeDir, db: globalDB}
}

// HomeDir returns the configured global home directory.
func (r *ProjectResolver) HomeDir() string { return r.homeDir }

// DB returns the underlying global database handle.
func (r *ProjectResolver) DB() *sql.DB { return r.db }

// Resolve returns the project location for the given root path. The flow is:
//
//  1. If the project is registered in the global registry, honor its mode
//     (external or local).
//  2. If the project is NOT registered but a legacy <root>/.plan-ai/project.db
//     exists, return ErrLegacyLocalStoreFound so the caller can prompt the
//     user to choose between `init --local` and `migrate local-to-global`.
//  3. If the project is NOT registered and no legacy store exists, return
//     a draft external ProjectLocation so callers can register it.
func (r *ProjectResolver) Resolve(rootPath string) (ProjectLocation, error) {
	cleaned, err := filepath.Abs(rootPath)
	if err != nil {
		return ProjectLocation{}, fmt.Errorf("resolve project root: %w", err)
	}
	cleaned = filepath.Clean(cleaned)

	registry := NewProjectRegistryRepository(r.db)
	if entry, err := registry.GetByPath(cleaned); err == nil {
		return r.locationForEntry(entry, cleaned)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return ProjectLocation{}, err
	}

	// Not registered. Check for legacy local store.
	if legacyExists(cleaned) {
		return ProjectLocation{}, ErrLegacyLocalStoreFound
	}

	slug := config.ProjectSlug(cleaned)
	layout, err := EnsureExternalProjectLayout(r.homeDir, slug)
	if err != nil {
		return ProjectLocation{}, err
	}
	return ProjectLocation{
		ProjectID: ProjectID(cleaned),
		Name:      filepath.Base(cleaned),
		RootPath:  cleaned,
		Slug:      slug,
		Mode:      config.ProjectModeExternal,
		Layout: ProjectLocationLayout{
			Dir:          layout.Dir,
			ConfigPath:   layout.ConfigPath,
			CacheDir:     layout.CacheDir,
			SnapshotsDir: layout.SnapshotsDir,
			ExportsDir:   layout.ExportsDir,
			DocsDir:      layout.DocsDir,
			LocksDir:     layout.LocksDir,
			BackupsDir:   layout.BackupsDir,
		},
		DBPath: layout.DBPath,
	}, nil
}

// OpenStore resolves the project and opens its database, applying project
// migrations. The caller is responsible for closing the returned *sql.DB.
func (r *ProjectResolver) OpenStore(rootPath string) (*sql.DB, ProjectLocation, error) {
	loc, err := r.Resolve(rootPath)
	if err != nil {
		return nil, ProjectLocation{}, err
	}
	db, err := Open(loc.DBPath)
	if err != nil {
		return nil, loc, err
	}
	if err := RunProjectMigrations(db); err != nil {
		_ = db.Close()
		return nil, loc, err
	}
	if loc.Mode == config.ProjectModeExternal {
		registry := NewProjectRegistryRepository(r.db)
		_ = registry.TouchLastSeen(loc.ProjectID)
	}
	return db, loc, nil
}

func (r *ProjectResolver) locationForEntry(entry ProjectRegistryEntry, rootPath string) (ProjectLocation, error) {
	loc := ProjectLocation{
		ProjectID: entry.ID,
		Name:      entry.Name,
		RootPath:  rootPath,
		Slug:      entry.Slug,
		Mode:      entry.Mode,
	}
	switch entry.Mode {
	case config.ProjectModeLocal:
		layout, err := EnsureProjectLayout(rootPath)
		if err != nil {
			return ProjectLocation{}, err
		}
		loc.Layout = ProjectLocationLayout{
			Dir:          layout.Dir,
			ConfigPath:   layout.ConfigPath,
			CacheDir:     layout.CacheDir,
			SnapshotsDir: layout.SnapshotsDir,
			ExportsDir:   layout.ExportsDir,
			DocsDir:      layout.DocsDir,
			LocksDir:     layout.LocksDir,
			BackupsDir:   layout.BackupsDir,
		}
		loc.DBPath = layout.DBPath
	case config.ProjectModeExternal, "":
		slug := entry.Slug
		if slug == "" {
			slug = config.ProjectSlug(rootPath)
		}
		layout, err := EnsureExternalProjectLayout(r.homeDir, slug)
		if err != nil {
			return ProjectLocation{}, err
		}
		loc.Mode = config.ProjectModeExternal
		loc.Slug = slug
		loc.Layout = ProjectLocationLayout{
			Dir:          layout.Dir,
			ConfigPath:   layout.ConfigPath,
			CacheDir:     layout.CacheDir,
			SnapshotsDir: layout.SnapshotsDir,
			ExportsDir:   layout.ExportsDir,
			DocsDir:      layout.DocsDir,
			LocksDir:     layout.LocksDir,
			BackupsDir:   layout.BackupsDir,
		}
		loc.DBPath = layout.DBPath
	default:
		return ProjectLocation{}, fmt.Errorf("project resolver: unknown mode %q for project %q", entry.Mode, entry.RootPath)
	}
	return loc, nil
}

func legacyExists(rootPath string) bool {
	_, err := os.Stat(config.ProjectDBPath(rootPath))
	return err == nil
}

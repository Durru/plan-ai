package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ProjectRegistryEntry is the persistent record for a project in the global
// registry. The slug is the filesystem-safe identifier used for the external
// project directory under ~/.plan-ai/projects/<slug>/.
type ProjectRegistryEntry struct {
	ID         string
	Name       string
	RootPath   string
	Slug       string
	Mode       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LastSeenAt time.Time
}

// ProjectRegistryRepository manages the global project registry stored in
// the known_projects table of the global Plan-AI database.
type ProjectRegistryRepository struct {
	db *sql.DB
}

// NewProjectRegistryRepository constructs a ProjectRegistryRepository that
// uses the provided global database handle.
func NewProjectRegistryRepository(db *sql.DB) *ProjectRegistryRepository {
	return &ProjectRegistryRepository{db: db}
}

// Register inserts or refreshes a project entry. It returns the resolved
// entry (with timestamps populated from the database) and a sentinel error
// (ErrProjectRegistryNotInitialized) if the global migrations have not been
// run yet (no known_projects table present).
func (r *ProjectRegistryRepository) Register(entry ProjectRegistryEntry) (ProjectRegistryEntry, error) {
	if r == nil || r.db == nil {
		return ProjectRegistryEntry{}, errors.New("project registry: nil database handle")
	}
	if entry.Mode == "" {
		entry.Mode = "external"
	}
	if entry.Slug == "" {
		entry.Slug = entry.Name
	}
	now := time.Now().UTC()
	if err := UpsertKnownProjectWithMode(r.db, entry.ID, entry.Name, entry.RootPath, entry.Slug, entry.Mode); err != nil {
		return ProjectRegistryEntry{}, fmt.Errorf("upsert project %q: %w", entry.RootPath, err)
	}
	entry.CreatedAt = now
	entry.UpdatedAt = now
	entry.LastSeenAt = now
	return entry, nil
}

// GetByPath returns the registry entry for the project whose RootPath matches
// the given path, or sql.ErrNoRows if no such project is registered.
func (r *ProjectRegistryRepository) GetByPath(rootPath string) (ProjectRegistryEntry, error) {
	return r.queryOne(`SELECT id, name, path, slug, mode, created_at, updated_at, last_seen_at FROM known_projects WHERE path = ?`, rootPath)
}

// GetByID returns the registry entry for the given project ID, or
// sql.ErrNoRows if no such project is registered.
func (r *ProjectRegistryRepository) GetByID(id string) (ProjectRegistryEntry, error) {
	return r.queryOne(`SELECT id, name, path, slug, mode, created_at, updated_at, last_seen_at FROM known_projects WHERE id = ?`, id)
}

// GetBySlug returns the registry entry for the given slug, or sql.ErrNoRows
// if no such project is registered.
func (r *ProjectRegistryRepository) GetBySlug(slug string) (ProjectRegistryEntry, error) {
	return r.queryOne(`SELECT id, name, path, slug, mode, created_at, updated_at, last_seen_at FROM known_projects WHERE slug = ?`, slug)
}

// List returns all registered projects, ordered by name.
func (r *ProjectRegistryRepository) List() ([]ProjectRegistryEntry, error) {
	rows, err := r.db.Query(`SELECT id, name, path, slug, mode, created_at, updated_at, last_seen_at FROM known_projects ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	out := []ProjectRegistryEntry{}
	for rows.Next() {
		entry, err := scanProjectRegistryEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// TouchLastSeen updates the last_seen_at timestamp for the given project.
func (r *ProjectRegistryRepository) TouchLastSeen(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(`UPDATE known_projects SET last_seen_at = ?, updated_at = ? WHERE id = ?`, now, now, id)
	return err
}

// Unregister removes a project entry from the registry. It does NOT delete
// the project's on-disk data.
func (r *ProjectRegistryRepository) Unregister(id string) error {
	_, err := r.db.Exec(`DELETE FROM known_projects WHERE id = ?`, id)
	return err
}

func (r *ProjectRegistryRepository) queryOne(query string, args ...any) (ProjectRegistryEntry, error) {
	row := r.db.QueryRow(query, args...)
	return scanProjectRegistryEntry(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProjectRegistryEntry(row rowScanner) (ProjectRegistryEntry, error) {
	var (
		entry        ProjectRegistryEntry
		createdAt    string
		updatedAt    string
		lastSeenAt   string
	)
	if err := row.Scan(&entry.ID, &entry.Name, &entry.RootPath, &entry.Slug, &entry.Mode, &createdAt, &updatedAt, &lastSeenAt); err != nil {
		return ProjectRegistryEntry{}, err
	}
	if t, ok := parseRegistryTime(createdAt); ok {
		entry.CreatedAt = t
	}
	if t, ok := parseRegistryTime(updatedAt); ok {
		entry.UpdatedAt = t
	}
	if t, ok := parseRegistryTime(lastSeenAt); ok {
		entry.LastSeenAt = t
	}
	return entry, nil
}

func parseRegistryTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, true
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

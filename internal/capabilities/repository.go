package capabilities

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(c Capability) error {
	if c.ID == "" {
		c.ID = domain.NewID("cap")
	}
	if c.Name == "" {
		c.Name = string(c.Type)
	}
	if c.SchemaInfo == "" {
		c.SchemaInfo = "{}"
	}
	if c.Version == "" {
		c.Version = "1.0"
	}
	enabled := 0
	if c.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO capabilities_v2 (id, name, description, schema_info, version, enabled, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET
		   description = excluded.description,
		   schema_info = excluded.schema_info,
		   version = excluded.version,
		   enabled = excluded.enabled`,
		c.ID, c.Name, c.Description, c.SchemaInfo, c.Version, enabled, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) List() ([]Capability, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, schema_info, version, enabled, created_at
		 FROM capabilities_v2 WHERE enabled = 1 ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Capability
	for rows.Next() {
		var c Capability
		var created string
		var enabled int
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.SchemaInfo, &c.Version, &enabled, &created); err != nil {
			return nil, err
		}
		c.Type = CapabilityType(c.Name)
		c.Enabled = enabled == 1
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *Repository) GetByName(name string) (Capability, error) {
	row := r.db.QueryRow(
		`SELECT id, name, description, schema_info, version, enabled, created_at
		 FROM capabilities_v2 WHERE name = ?`, name,
	)
	var c Capability
	var created string
	var enabled int
	if err := row.Scan(&c.ID, &c.Name, &c.Description, &c.SchemaInfo, &c.Version, &enabled, &created); err != nil {
		return Capability{}, fmt.Errorf("capability %q not found: %w", name, err)
	}
	c.Type = CapabilityType(c.Name)
	c.Enabled = enabled == 1
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return c, nil
}

func (r *Repository) Delete(name string) error {
	_, err := r.db.Exec(`DELETE FROM capabilities_v2 WHERE name = ?`, name)
	return err
}

// SeedDefaults inserts the default capabilities into capabilities_v2.
// Uses INSERT OR IGNORE for idempotency.
func SeedDefaults(db *sql.DB) error {
	now := time.Now().UTC().Format(time.RFC3339)
	defaults := []struct {
		id, name, desc string
	}{
		{domain.NewID("cap"), "vision", "Create vision drafts from ingested inputs"},
		{domain.NewID("cap"), "research", "Perform research on topics with sources and findings"},
		{domain.NewID("cap"), "planning", "Create master plans, specific plans, and implementation documents"},
		{domain.NewID("cap"), "architecture", "Design system architecture and component relationships"},
		{domain.NewID("cap"), "database", "Design database schemas, migrations, and queries"},
		{domain.NewID("cap"), "backend", "Implement server-side logic and APIs"},
		{domain.NewID("cap"), "frontend", "Implement client-side interfaces and interactions"},
		{domain.NewID("cap"), "security", "Review code and architecture for security issues"},
		{domain.NewID("cap"), "testing", "Write and execute tests for verification"},
		{domain.NewID("cap"), "impact_analysis", "Analyze impact of changes on plans and decisions"},
		{domain.NewID("cap"), "validation", "Validate plans, tasks, and implementations against criteria"},
	}
	for _, c := range defaults {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO capabilities_v2 (id, name, description, schema_info, version, enabled, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			c.id, c.name, c.desc, "{}", "1.0", 1, now,
		); err != nil {
			return fmt.Errorf("seed capability %q: %w", c.name, err)
		}
	}
	return nil
}

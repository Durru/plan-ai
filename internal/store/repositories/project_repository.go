package repositories

import (
	"database/sql"

	"github.com/Durru/plan-ai/internal/domain"
)

type ProjectRepository struct{ db *sql.DB }

func NewProjectRepository(db *sql.DB) ProjectRepository { return ProjectRepository{db: db} }

var _ domain.ProjectRepository = ProjectRepository{}

func (r ProjectRepository) Save(p domain.Project) error {
	p.ID = ensureID(p.ID, "project")
	if p.Status == "" {
		p.Status = domain.ProjectStatusDraft
	}
	c, u := times(p.CreatedAt, p.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO projects (id, name, root_path, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET name=excluded.name, root_path=excluded.root_path, description=excluded.description, status=excluded.status, updated_at=excluded.updated_at`,
		p.ID, p.Name, p.RootPath, p.Description, p.Status, c, u)
	return err
}

func (r ProjectRepository) GetByID(id string) (domain.Project, error) {
	return r.get(`WHERE id = ?`, id)
}
func (r ProjectRepository) FindByName(name string) (domain.Project, error) {
	return r.get(`WHERE name = ?`, name)
}

func (r ProjectRepository) get(where string, args ...any) (domain.Project, error) {
	var p domain.Project
	var c, u string
	var status string
	err := r.db.QueryRow(`SELECT id, name, root_path, description, status, created_at, updated_at FROM projects `+where, args...).Scan(&p.ID, &p.Name, &p.RootPath, &p.Description, &status, &c, &u)
	p.Status = domain.ProjectStatus(status)
	p.CreatedAt = parse(c)
	p.UpdatedAt = parse(u)
	return p, err
}

func (r ProjectRepository) List() ([]domain.Project, error) {
	rows, err := r.db.Query(`SELECT id, name, root_path, description, status, created_at, updated_at FROM projects ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		var c, u, status string
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.Description, &status, &c, &u); err != nil {
			return nil, err
		}
		p.Status = domain.ProjectStatus(status)
		p.CreatedAt = parse(c)
		p.UpdatedAt = parse(u)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r ProjectRepository) UpdateStatus(id string, status domain.ProjectStatus) error {
	return updateStatus(r.db, "projects", id, status)
}
func (r ProjectRepository) Delete(id string) error { return deleteByID(r.db, "projects", id) }

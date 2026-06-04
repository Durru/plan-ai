package store

import "database/sql"

type ProjectStore struct {
	Layout ProjectLayout
	DB     *sql.DB
}

func OpenProjectStore(projectRoot string) (*ProjectStore, error) {
	layout, err := EnsureProjectLayout(projectRoot)
	if err != nil {
		return nil, err
	}
	db, err := Open(layout.DBPath)
	if err != nil {
		return nil, err
	}
	if err := RunProjectMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return &ProjectStore{Layout: layout, DB: db}, nil
}

func (s *ProjectStore) Close() error { return s.DB.Close() }

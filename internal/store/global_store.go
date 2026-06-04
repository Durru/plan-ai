package store

import "database/sql"

type GlobalStore struct {
	Layout GlobalLayout
	DB     *sql.DB
}

func OpenGlobalStore(homeRoot string) (*GlobalStore, error) {
	layout, err := EnsureGlobalLayout(homeRoot)
	if err != nil {
		return nil, err
	}
	db, err := Open(layout.DBPath)
	if err != nil {
		return nil, err
	}
	if err := RunGlobalMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return &GlobalStore{Layout: layout, DB: db}, nil
}

func (s *GlobalStore) Close() error { return s.DB.Close() }

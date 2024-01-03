package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	op         = "storage.sqlite.New"
	driverName = "sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewWithContext(_ context.Context, storagePath string) (*Storage, error) {
	db, err := sql.Open(driverName, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

func (s *Storage) Check() (interface{}, error) {
	return s.db.Stats(), s.db.Ping()
}

func (s *Storage) Prepare(query string) (*sql.Stmt, error) {
	return s.db.Prepare(query)
}

func (s *Storage) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

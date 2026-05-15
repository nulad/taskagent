package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB     *sql.DB
	logger *slog.Logger
}

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrPermissionDenied = errors.New("permission denied")
)

func NewStore(dbPath string, logger *slog.Logger) (*Store, error) {
	dsn := dbPath
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Store{DB: db, logger: logger}

	if err := s.runMigrations(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

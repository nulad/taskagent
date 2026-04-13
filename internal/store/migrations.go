package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/nulad/taskagent/migrations"
)

func (s *Store) runMigrations() error {
	tx, err := s.DB.Begin()

	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS _migrations (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		applied_at TEXT NOT NULL
	);`); err != nil {
		return fmt.Errorf("create _migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := entry.Name()

		var exists int
		err := tx.QueryRow("SELECT 1 FROM _migrations WHERE name = ? LIMIT 1", name).Scan(&exists)
		if err == nil {
			continue
		}

		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		sqlBytes, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO _migrations (name, applied_at) VALUES (?, ?)`,
			name,
			time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}

	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}

	return nil
}

package store

import "testing"

func TestNewStore_AppliesMigrations(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	// Core tables expected from 001_initial.sql + runner bootstrap.
	expectedTables := []string{
		"projects",
		"statuses",
		"tasks",
		"users",
		"api_keys",
		"user_project_permissions",
		"_migrations",
	}
	for _, table := range expectedTables {
		assertTableExists(t, s, table)
	}

	// Seeded statuses should exist.
	var statusCount int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM statuses`).Scan(&statusCount); err != nil {
		t.Fatalf("query statuses count: %v", err)
	}
	if statusCount != 6 {
		t.Fatalf("expected 6 statuses, got %d", statusCount)
	}

	// Initial migration should be recorded.
	var migrationCount int
	if err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM _migrations WHERE name = ?`,
		"001_initial.sql",
	).Scan(&migrationCount); err != nil {
		t.Fatalf("query _migrations: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected 001_initial.sql recorded once, got %d", migrationCount)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	var before int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM _migrations`).Scan(&before); err != nil {
		t.Fatalf("count _migrations before rerun: %v", err)
	}

	// Re-running should not fail and should not duplicate migration records.
	if err := s.runMigrations(); err != nil {
		t.Fatalf("runMigrations() second run error = %v", err)
	}

	var after int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM _migrations`).Scan(&after); err != nil {
		t.Fatalf("count _migrations after rerun: %v", err)
	}

	if after != before {
		t.Fatalf("expected migration count to remain %d, got %d", before, after)
	}
}

func assertTableExists(t *testing.T, s *Store, table string) {
	t.Helper()

	var name string
	err := s.DB.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
		table,
	).Scan(&name)
	if err != nil {
		t.Fatalf("table %q not found: %v", table, err)
	}
}

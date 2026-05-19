package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("TASKAGENT_LISTEN_ADDR", "")
	t.Setenv("TASKAGENT_DB_PATH", "")
	t.Setenv("TASKAGENT_LOG_LEVEL", "")

	cfg := Load()

	if cfg.ListenAddr != defaultListenAddr {
		t.Errorf("expected listen addr %s, got %s", defaultListenAddr, cfg.ListenAddr)
	}
	if cfg.DatabasePath != defaultDatabasePath {
		t.Errorf("expected database path %s, got %s", defaultDatabasePath, cfg.DatabasePath)
	}
	if cfg.LogLevel != defaultLogLevel {
		t.Errorf("expected log level %s, got %s", defaultLogLevel, cfg.LogLevel)
	}
}

func TestLoadOverridesFromEnv(t *testing.T) {
	t.Setenv("TASKAGENT_LISTEN_ADDR", ":9090")
	t.Setenv("TASKAGENT_DB_PATH", "/tmp/test.db")
	t.Setenv("TASKAGENT_LOG_LEVEL", "debug")

	cfg := Load()

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected listen addr :9090, got %s", cfg.ListenAddr)
	}
	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("expected database path /tmp/test.db, got %s", cfg.DatabasePath)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
}

func TestLoadCORSDDefaultsToEmpty(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "")

	cfg := Load()

	if cfg.CORSOrigins != nil {
		t.Errorf("expected nil CORS origins by default, got %v", cfg.CORSOrigins)
	}
}

func TestLoadCORSOriginsSingleOrigin(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "https://example.com")

	cfg := Load()

	expected := []string{"https://example.com"}
	if len(cfg.CORSOrigins) != len(expected) {
		t.Fatalf("expected %d origins, got %d: %v", len(expected), len(cfg.CORSOrigins), cfg.CORSOrigins)
	}
	if cfg.CORSOrigins[0] != expected[0] {
		t.Errorf("expected origin %s, got %s", expected[0], cfg.CORSOrigins[0])
	}
}

func TestLoadCORSOriginsMultipleOrigins(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "https://a.com,https://b.com,https://c.com")

	cfg := Load()

	expected := []string{"https://a.com", "https://b.com", "https://c.com"}
	if len(cfg.CORSOrigins) != len(expected) {
		t.Fatalf("expected %d origins, got %d: %v", len(expected), len(cfg.CORSOrigins), cfg.CORSOrigins)
	}
	for i, exp := range expected {
		if cfg.CORSOrigins[i] != exp {
			t.Errorf("expected origin[%d] = %s, got %s", i, exp, cfg.CORSOrigins[i])
		}
	}
}

func TestLoadCORSOriginsTrimsWhitespace(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "  https://a.com  ,  https://b.com  ")

	cfg := Load()

	expected := []string{"https://a.com", "https://b.com"}
	if len(cfg.CORSOrigins) != len(expected) {
		t.Fatalf("expected %d origins, got %d: %v", len(expected), len(cfg.CORSOrigins), cfg.CORSOrigins)
	}
	for i, exp := range expected {
		if cfg.CORSOrigins[i] != exp {
			t.Errorf("expected origin[%d] = %s, got %s", i, exp, cfg.CORSOrigins[i])
		}
	}
}

func TestLoadCORSOriginsSkipsEmptyEntries(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "https://a.com,,https://b.com,  ,https://c.com")

	cfg := Load()

	expected := []string{"https://a.com", "https://b.com", "https://c.com"}
	if len(cfg.CORSOrigins) != len(expected) {
		t.Fatalf("expected %d origins, got %d: %v", len(expected), len(cfg.CORSOrigins), cfg.CORSOrigins)
	}
	for i, exp := range expected {
		if cfg.CORSOrigins[i] != exp {
			t.Errorf("expected origin[%d] = %s, got %s", i, exp, cfg.CORSOrigins[i])
		}
	}
}

func TestLoadCORSOriginsAllWhitespaceYieldsEmpty(t *testing.T) {
	t.Setenv("TASKAGENT_CORS_ORIGINS", "  ,  ,  ")

	cfg := Load()

	if cfg.CORSOrigins != nil {
		t.Errorf("expected nil CORS origins when all entries are whitespace, got %v", cfg.CORSOrigins)
	}
}

func TestLoadCORSSetsOrigins(t *testing.T) {
	// Verify that LoadCORSOrigins works as expected with the strings helper
	got := loadCORSOrigins()
	if got != nil {
		t.Errorf("expected nil when env is unset, got %v", got)
	}
	t.Setenv(envCORSOrigins, "https://trusted.com")
	got = loadCORSOrigins()
	if !reflect.DeepEqual(got, []string{"https://trusted.com"}) {
		t.Errorf("expected [https://trusted.com], got %v", got)
	}
	t.Setenv(envCORSOrigins, "https://x.com,, ,https://y.com")
	got = loadCORSOrigins()
	if !reflect.DeepEqual(got, []string{"https://x.com", "https://y.com"}) {
		t.Errorf("expected [https://x.com https://y.com], got %v", got)
	}
}

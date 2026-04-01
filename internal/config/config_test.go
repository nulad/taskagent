package config

import "testing"

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
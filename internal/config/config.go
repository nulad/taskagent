package config

import (
	"os"
	"strings"
)

const (
	defaultListenAddr   = ":8080"
	defaultDatabasePath = "taskagent.db"
	defaultLogLevel     = "info"

	envListen       = "TASKAGENT_LISTEN_ADDR"
	envDatabasePath = "TASKAGENT_DB_PATH"
	envLogLevel     = "TASKAGENT_LOG_LEVEL"
	envCORSOrigins  = "TASKAGENT_CORS_ORIGINS"
)

type Config struct {
	// ListenAddr is the address to listen on for HTTP requests.
	ListenAddr string
	// DatabasePath is the path to the database file.
	DatabasePath string
	// LogLevel is the level of logging to use.
	LogLevel string
	// CORSOrigins is an allowlist of allowed CORS origins.
	CORSOrigins []string
}

func Load() Config {
	cfg := Config{
		ListenAddr:   defaultListenAddr,
		DatabasePath: defaultDatabasePath,
		LogLevel:     defaultLogLevel,
	}
	if env := os.Getenv(envListen); env != "" {
		cfg.ListenAddr = env
	}
	if dbPath := os.Getenv(envDatabasePath); dbPath != "" {
		cfg.DatabasePath = dbPath
	}
	if logLevel := os.Getenv(envLogLevel); logLevel != "" {
		cfg.LogLevel = logLevel
	}
	cfg.CORSOrigins = loadCORSOrigins()
	return cfg
}

func loadCORSOrigins() []string {
	raw := os.Getenv(envCORSOrigins)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var origins []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

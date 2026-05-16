package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/nulad/taskagent/internal/middleware"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

type e2eHarness struct {
	server *httptest.Server
	client  *http.Client
	apiKey  string
}

func newE2EServer(t *testing.T) *e2eHarness {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Quiet test logger
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	appStore, err := store.NewStore(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() {
		_ = appStore.Close()
	})

	// Seed an admin user and raw API key
	ctx := context.Background()
	adminUser, err := appStore.CreateUser(ctx, "admin", true)
	if err != nil {
		t.Fatalf("failed to seed admin user: %v", err)
	}

	_, apiKey, err := appStore.CreateApiKey(ctx, "bootstrap", adminUser.ID)
	if err != nil {
		t.Fatalf("failed to seed api key: %v", err)
	}

	// Build services
	projectService := service.NewProjectService(appStore, logger)
	taskService := service.NewTaskService(appStore, logger)

	// Build handlers
	projectHandler := NewProjectHandler(projectService, logger)
	taskHandler := NewTaskHandler(taskService, logger)
	authHandler := NewAuthHandler(appStore, logger)

	// Build muxes and middleware to match cmd/server/main.go
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	protectedMux := http.NewServeMux()
	RegisterProjectRoutes(protectedMux, projectHandler)
	RegisterTaskRoutes(protectedMux, taskHandler)
	RegisterAuthRoutes(protectedMux, authHandler)

	protectedAPI := middleware.AuthMiddleware(appStore)(protectedMux)
	mux.Handle("/", protectedAPI)

	finalHandler := middleware.RequestIDMiddleware()(
		middleware.RequestLoggingMiddleware(logger)(mux),
	)

	server := httptest.NewServer(finalHandler)
	t.Cleanup(func() {
		server.Close()
	})

	return &e2eHarness{
		server: server,
		client: server.Client(),
		apiKey: apiKey,
	}
}

func TestE2ESmoke(t *testing.T) {
	h := newE2EServer(t)

	req, err := http.NewRequest("GET", h.server.URL+"/projects", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-API-Key", h.apiKey)

	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var projects []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify it's an array (even if empty)
	if projects == nil {
		t.Errorf("expected projects to be a slice, got nil")
	}
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nulad/taskagent/internal/middleware"
	"github.com/nulad/taskagent/internal/model"
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

func (h *e2eHarness) newJSONRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	url := h.server.URL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("X-API-Key", h.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

func (h *e2eHarness) doJSONRequest(t *testing.T, method, path string, body interface{}, expectedStatus int, dest interface{}) {
	t.Helper()
	req := h.newJSONRequest(t, method, path, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	}
}

func (h *e2eHarness) doNoContentRequest(t *testing.T, method, path string, body interface{}, expectedStatus int) {
	t.Helper()
	req := h.newJSONRequest(t, method, path, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func (h *e2eHarness) assertJSONError(t *testing.T, method, path string, body interface{}, expectedStatus int, expectedErrorMessage string) {
	t.Helper()
	req := h.newJSONRequest(t, method, path, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatalf("expected error message in response, but got none")
	}

	if expectedErrorMessage != "" && !strings.Contains(errResp.Error, expectedErrorMessage) {
		t.Fatalf("expected error message to contain %q, got %q", expectedErrorMessage, errResp.Error)
	}
}

func TestE2ESmoke(t *testing.T) {
	h := newE2EServer(t)

	var projects []interface{}
	h.doJSONRequest(t, "GET", "/projects", nil, http.StatusOK, &projects)

	// Verify it's an array (even if empty)
	if projects == nil {
		t.Errorf("expected projects to be a slice, got nil")
	}
}

func TestProjectLifecycleE2E(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create Project
	newProject := model.Project{
		Name:        "Test Project",
		Description: "Test description",
	}
	var createdProject model.Project
	h.doJSONRequest(t, "POST", "/projects", newProject, http.StatusCreated, &createdProject)

	if createdProject.ID == "" {
		t.Error("expected non-empty project ID")
	}
	if createdProject.Name != newProject.Name {
		t.Errorf("expected name %q, got %q", newProject.Name, createdProject.Name)
	}
	if createdProject.Description != newProject.Description {
		t.Errorf("expected description %q, got %q", newProject.Description, createdProject.Description)
	}
	if createdProject.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}
	if createdProject.UpdatedAt == "" {
		t.Error("expected non-empty UpdatedAt")
	}

	// 2. Get Project
	var fetchedProject model.Project
	h.doJSONRequest(t, "GET", "/projects/"+createdProject.ID, nil, http.StatusOK, &fetchedProject)

	if fetchedProject.ID != createdProject.ID {
		t.Errorf("expected ID %q, got %q", createdProject.ID, fetchedProject.ID)
	}
	if fetchedProject.Name != createdProject.Name {
		t.Errorf("expected name %q, got %q", createdProject.Name, fetchedProject.Name)
	}

	// 3. List Projects
	var projects []model.Project
	h.doJSONRequest(t, "GET", "/projects", nil, http.StatusOK, &projects)

	found := false
	for _, p := range projects {
		if p.ID == createdProject.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find project %q in list", createdProject.ID)
	}

	// 4. Delete Project
	h.doNoContentRequest(t, "DELETE", "/projects/"+createdProject.ID, nil, http.StatusNoContent)

	// 5. Verify deletion
	h.assertJSONError(t, "GET", "/projects/"+createdProject.ID, nil, http.StatusNotFound, "not found")
}

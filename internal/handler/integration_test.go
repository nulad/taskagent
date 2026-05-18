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

func TestTaskListFilteringAndEmptyArrays(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create a project
	newProject := model.Project{
		Name:        "Filtering Test Project",
		Description: "Project for testing list filtering and empty arrays",
	}
	var createdProject model.Project
	h.doJSONRequest(t, "POST", "/projects", newProject, http.StatusCreated, &createdProject)

	// 2. Create three tasks for this project
	taskTitles := []string{"Task A", "Task B", "Task C"}
	tasks := make([]model.Task, len(taskTitles))

	for i, title := range taskTitles {
		taskReq := model.Task{
			ProjectID: createdProject.ID,
			Title:     title,
		}
		h.doJSONRequest(t, "POST", "/tasks", taskReq, http.StatusCreated, &tasks[i])
	}

	// 3. Move one task to 'done' and leave the others in 'backlog'
	taskToComplete := tasks[0]

	// Move: backlog -> todo -> in-progress -> review -> done
	statuses := []model.TaskStatus{
		model.StatusTodo,
		model.StatusInProgress,
		model.StatusReview,
		model.StatusDone,
	}
	for _, nextStatus := range statuses {
		moveReq := map[string]model.TaskStatus{"status": nextStatus}
		var updatedTask model.Task
		h.doJSONRequest(t, "PATCH", "/tasks/"+taskToComplete.ID+"/move", moveReq, http.StatusOK, &updatedTask)
		if updatedTask.Status != nextStatus {
			t.Errorf("expected status %q, got %q", nextStatus, updatedTask.Status)
		}
	}

	// Verify the moved task is done
	var finalTask model.Task
	h.doJSONRequest(t, "GET", "/tasks/"+taskToComplete.ID, nil, http.StatusOK, &finalTask)
	if finalTask.Status != model.StatusDone {
		t.Errorf("expected task %q to be %q, got %q", taskToComplete.ID, model.StatusDone, finalTask.Status)
	}

	// --- Step 3: GET /tasks?project_id={projectID} returns all three tasks ---
	var allTasks []model.Task
	h.doJSONRequest(t, "GET", "/tasks?project_id="+createdProject.ID, nil, http.StatusOK, &allTasks)

	if allTasks == nil {
		t.Error("expected non-nil slice for GET /tasks?project_id, got nil")
	}
	if len(allTasks) != 3 {
		t.Errorf("expected 3 tasks for project %q, got %d", createdProject.ID, len(allTasks))
	}

	// --- Step 4: GET /tasks?project_id={projectID}&status=done returns only the done task ---
	var doneTasks []model.Task
	h.doJSONRequest(t, "GET", "/tasks?project_id="+createdProject.ID+"&status=done", nil, http.StatusOK, &doneTasks)

	if doneTasks == nil {
		t.Error("expected non-nil slice for GET /tasks?project_id&status=done, got nil")
	}
	if len(doneTasks) != 1 {
		t.Errorf("expected 1 done task for project %q, got %d", createdProject.ID, len(doneTasks))
	}
	if len(doneTasks) > 0 && doneTasks[0].ID != taskToComplete.ID {
		t.Errorf("expected done task ID %q, got %q", taskToComplete.ID, doneTasks[0].ID)
	}

	// --- Step 5: GET /tasks?project_id={projectID}&status=review returns non-nil empty slice ---
	var reviewTasks []model.Task
	h.doJSONRequest(t, "GET", "/tasks?project_id="+createdProject.ID+"&status=review", nil, http.StatusOK, &reviewTasks)

	if reviewTasks == nil {
		t.Error("expected non-nil empty slice for GET /tasks?project_id&status=review, got nil")
	}
	if len(reviewTasks) != 0 {
		t.Errorf("expected 0 review tasks for project %q, got %d", createdProject.ID, len(reviewTasks))
	}

	// --- Step 6: GET /projects in a fresh harness returns non-nil empty slice ---
	freshHarness := newE2EServer(t)
	var emptyProjects []model.Project
	freshHarness.doJSONRequest(t, "GET", "/projects", nil, http.StatusOK, &emptyProjects)

	if emptyProjects == nil {
		t.Error("expected non-nil empty slice for GET /projects on fresh harness, got nil")
	}
	if len(emptyProjects) != 0 {
		t.Errorf("expected 0 projects on fresh harness, got %d", len(emptyProjects))
	}
}

func TestTaskLifecycleE2E(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create Project
	newProject := model.Project{
		Name:        "Task Workflow Project",
		Description: "Project to test task status workflow",
	}
	var createdProject model.Project
	h.doJSONRequest(t, "POST", "/projects", newProject, http.StatusCreated, &createdProject)

	// 2. Create three tasks
	taskTitles := []string{"Task 1", "Task 2", "Task 3"}
	tasks := make([]model.Task, len(taskTitles))

	for i, title := range taskTitles {
		taskReq := model.Task{
			ProjectID: createdProject.ID,
			Title:     title,
		}
		h.doJSONRequest(t, "POST", "/tasks", taskReq, http.StatusCreated, &tasks[i])

		// Assertions for each task creation
		if tasks[i].ID == "" {
			t.Error("expected non-empty task ID")
		}
		if tasks[i].ProjectID != createdProject.ID {
			t.Errorf("expected project ID %q, got %q", createdProject.ID, tasks[i].ProjectID)
		}
		if tasks[i].Title != title {
			t.Errorf("expected title %q, got %q", title, tasks[i].Title)
		}
		if tasks[i].Status != model.StatusBacklog {
			t.Errorf("expected initial status %q, got %q", model.StatusBacklog, tasks[i].Status)
		}
	}

	// 3. Move one task through the sequence
	// Sequence: backlog -> todo -> in-progress -> review -> done
	statuses := []model.TaskStatus{
		model.StatusTodo,
		model.StatusInProgress,
		model.StatusReview,
		model.StatusDone,
	}

	taskToMove := tasks[0]
	for _, nextStatus := range statuses {
		moveReq := map[string]model.TaskStatus{
			"status": nextStatus,
		}
		var updatedTask model.Task
		h.doJSONRequest(t, "PATCH", "/tasks/"+taskToMove.ID+"/move", moveReq, http.StatusOK, &updatedTask)

		// Verify the response body reports the requested status
		if updatedTask.Status != nextStatus {
			t.Errorf("expected status %q, got %q", nextStatus, updatedTask.Status)
		}
	}

	// 4. Fetch the moved task and assert its final status is 'done'
	var finalTask model.Task
	h.doJSONRequest(t, "GET", "/tasks/"+taskToMove.ID, nil, http.StatusOK, &finalTask)

	if finalTask.Status != model.StatusDone {
		t.Errorf("expected final status %q, got %q", model.StatusDone, finalTask.Status)
	}
}

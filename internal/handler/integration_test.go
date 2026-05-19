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
	client *http.Client
	apiKey string
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

	// Wrap the mux with CORS to match cmd/server/main.go
	corsHandler := middleware.CORSMiddleware([]string{"http://localhost"})(mux)

	finalHandler := middleware.RequestIDMiddleware()(
		middleware.RequestLoggingMiddleware(logger)(corsHandler),
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

// newUnauthenticatedRequest creates a request without an X-API-Key header.
func (h *e2eHarness) newUnauthenticatedRequest(t *testing.T, method, path string, body interface{}) *http.Request {
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

	// Deliberately no X-API-Key header set
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// newOverrideAPIKeyRequest creates a request with an overridden X-API-Key value.
func (h *e2eHarness) newOverrideAPIKeyRequest(t *testing.T, method, path string, apiKey string, body interface{}) *http.Request {
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

	req.Header.Set("X-API-Key", apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// doUnauthenticatedRequest executes an unauthenticated request and validates the status code.
func (h *e2eHarness) doUnauthenticatedRequest(t *testing.T, method, path string, body interface{}, expectedStatus int, dest interface{}) {
	t.Helper()
	req := h.newUnauthenticatedRequest(t, method, path, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do unauthenticated request: %v", err)
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

// doOverrideKeyRequest executes a request with an overridden API key and validates the status code.
func (h *e2eHarness) doOverrideKeyRequest(t *testing.T, method, path string, apiKey string, body interface{}, expectedStatus int, dest interface{}) {
	t.Helper()
	req := h.newOverrideAPIKeyRequest(t, method, path, apiKey, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request with overridden key: %v", err)
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

func TestTaskDeleteWorkflowCleanupE2E(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create a project
	newProject := model.Project{
		Name:        "Delete Workflow Project",
		Description: "Project to test task deletion before project deletion",
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

		if tasks[i].ID == "" {
			t.Error("expected non-empty task ID")
		}
		if tasks[i].ProjectID != createdProject.ID {
			t.Errorf("expected project ID %q, got %q", createdProject.ID, tasks[i].ProjectID)
		}
	}

	// 3. Delete each task with DELETE /tasks/{id} and assert 204
	for _, task := range tasks {
		h.doNoContentRequest(t, "DELETE", "/tasks/"+task.ID, nil, http.StatusNoContent)
	}

	// 4. Fetch one deleted task and assert 404 with a JSON error body
	h.assertJSONError(t, "GET", "/tasks/"+tasks[0].ID, nil, http.StatusNotFound, "not found")

	// 5. Delete the project with DELETE /projects/{id} and assert 204
	h.doNoContentRequest(t, "DELETE", "/projects/"+createdProject.ID, nil, http.StatusNoContent)

	// 6. Fetch the deleted project and assert 404 with a JSON error body
	h.assertJSONError(t, "GET", "/projects/"+createdProject.ID, nil, http.StatusNotFound, "not found")
}

func TestTaskInvalidTransitionE2E(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create Project
	newProject := model.Project{
		Name:        "Invalid Transition Project",
		Description: "Project to test invalid task status transitions",
	}
	var createdProject model.Project
	h.doJSONRequest(t, "POST", "/projects", newProject, http.StatusCreated, &createdProject)

	// 2. Create a task
	taskReq := model.Task{
		ProjectID: createdProject.ID,
		Title:     "Task to move invalidly",
	}
	var createdTask model.Task
	h.doJSONRequest(t, "POST", "/tasks", taskReq, http.StatusCreated, &createdTask)

	// 3. Attempt to move the task directly from backlog to done
	moveReq := map[string]model.TaskStatus{
		"status": model.StatusDone,
	}

	// 4. Assert the response status is 422 Unprocessable Entity and contains an error message
	h.assertJSONError(t, "PATCH", "/tasks/"+createdTask.ID+"/move", moveReq, http.StatusUnprocessableEntity, "")

	// 5. Fetch the task and assert its status remains backlog
	var fetchedTask model.Task
	h.doJSONRequest(t, "GET", "/tasks/"+createdTask.ID, nil, http.StatusOK, &fetchedTask)

	if fetchedTask.Status != model.StatusBacklog {
		t.Errorf("expected status to remain %q, got %q", model.StatusBacklog, fetchedTask.Status)
	}
}

// assertJSONErrorContract asserts that a request returns the expected status code
// and the response body is a JSON object with a non-empty "error" field.
// It also verifies the Content-Type header is application/json.
func (h *e2eHarness) assertJSONErrorContract(t *testing.T, method, path string, body interface{},
	expectedStatus int, expectedErrorMessage string,
) {
	t.Helper()
	req := h.newJSONRequest(t, method, path, body)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d; body: %s", expectedStatus, resp.StatusCode, string(bodyBytes))
	}

	// Verify Content-Type is application/json
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	// Verify the body is a valid JSON object with a non-empty "error" field
	var jsonErr map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonErr); err != nil {
		t.Fatalf("expected valid JSON error response, decode error: %v", err)
	}

	errorVal, ok := jsonErr["error"]
	if !ok {
		t.Fatalf("expected JSON object to contain an \"error\" field")
	}
	errorStr, ok := errorVal.(string)
	if !ok || errorStr == "" {
		t.Fatalf("expected \"error\" field to be a non-empty string, got %v", errorVal)
	}

	if expectedErrorMessage != "" && !strings.Contains(errorStr, expectedErrorMessage) {
		t.Fatalf("expected error message to contain %q, got %q", expectedErrorMessage, errorStr)
	}
}

func TestAuthFailureE2E(t *testing.T) {
	h := newE2EServer(t)

	// --- Test 1: Missing API key returns 401 with JSON error body ---
	var missingKeyErrResp struct {
		Error string `json:"error"`
	}
	h.doUnauthenticatedRequest(t, "GET", "/projects", nil, http.StatusUnauthorized, &missingKeyErrResp)

	if missingKeyErrResp.Error == "" {
		t.Fatalf("expected non-empty error message for missing API key, got empty string")
	}

	// --- Test 2: Invalid API key returns 401 with JSON error body ---
	var invalidKeyErrResp struct {
		Error string `json:"error"`
	}
	h.doOverrideKeyRequest(t, "GET", "/projects", "invalid-key-12345", nil, http.StatusUnauthorized, &invalidKeyErrResp)

	if invalidKeyErrResp.Error == "" {
		t.Fatalf("expected non-empty error message for invalid API key, got empty string")
	}

	// --- Test 3: Verify auth is required for other protected routes too ---
	var createProjectErrResp struct {
		Error string `json:"error"`
	}
	h.doUnauthenticatedRequest(t, "POST", "/projects",
		map[string]string{"name": "unauthorized project"},
		http.StatusUnauthorized, &createProjectErrResp)

	if createProjectErrResp.Error == "" {
		t.Fatalf("expected non-empty error message for unauthenticated POST, got empty string")
	}

	var createTaskErrResp struct {
		Error string `json:"error"`
	}
	h.doOverrideKeyRequest(t, "POST", "/tasks",
		"wrong-key",
		map[string]string{"title": "unauthorized task"},
		http.StatusUnauthorized, &createTaskErrResp)

	if createTaskErrResp.Error == "" {
		t.Fatalf("expected non-empty error message for unauthenticated POST, got empty string")
	}
}

// TestMalformedJSONErrorContract verifies that a protected route returns 400
// with a JSON error body when the request body is not valid JSON.
func TestMalformedJSONErrorContract(t *testing.T) {
	h := newE2EServer(t)

	// Send raw invalid JSON (no Content-Type set, so readJSON will fail)
	req := httptest.NewRequest("POST", "/projects", strings.NewReader(`{not valid json}`))
	req.Header.Set("X-API-Key", h.apiKey)

	rec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 400, got %d; body: %s", resp.StatusCode, string(body))
	}

	// Verify Content-Type
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	// Verify JSON error body with non-empty error field
	var jsonErr map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonErr); err != nil {
		t.Fatalf("expected valid JSON error response, decode error: %v", err)
	}

	errorVal, ok := jsonErr["error"]
	if !ok {
		t.Fatal("expected JSON object to contain an \"error\" field")
	}
	errorStr, ok := errorVal.(string)
	if !ok || errorStr == "" {
		t.Fatalf("expected \"error\" field to be a non-empty string, got %v", errorVal)
	}
}

// TestValidationErrorsOnProjectCreate verifies that creating a project with
// missing required fields returns 400 with a JSON error body.
func TestValidationErrorsOnProjectCreate(t *testing.T) {
	h := newE2EServer(t)

	// Create a project with an empty name — should fail validation
	createReq := map[string]string{
		"name":        "",
		"description": "some description",
	}
	h.assertJSONErrorContract(t, "POST", "/projects", createReq,
		http.StatusBadRequest, "name")

	// Verify project was not created
	var projects []model.Project
	h.doJSONRequest(t, "GET", "/projects", nil, http.StatusOK, &projects)
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

// TestConflictOnDeleteProjectWithTasks verifies that deleting a project
// that still has tasks returns 409 with a JSON error body.
func TestConflictOnDeleteProjectWithTasks(t *testing.T) {
	h := newE2EServer(t)

	// 1. Create a project
	newProject := model.Project{
		Name:        "Conflict Project",
		Description: "Project that retains tasks for 409 testing",
	}
	var createdProject model.Project
	h.doJSONRequest(t, "POST", "/projects", newProject, http.StatusCreated, &createdProject)

	// 2. Create a task in that project
	taskReq := model.Task{
		ProjectID: createdProject.ID,
		Title:     "Task that prevents project deletion",
	}
	var createdTask model.Task
	h.doJSONRequest(t, "POST", "/tasks", taskReq, http.StatusCreated, &createdTask)

	// 3. Attempt to delete the project — should return 409 Conflict
	h.assertJSONErrorContract(t, "DELETE", "/projects/"+createdProject.ID, nil,
		http.StatusConflict, "has 1 associated task(s)")

	// 4. Verify the project still exists
	var fetchedProject model.Project
	h.doJSONRequest(t, "GET", "/projects/"+createdProject.ID, nil, http.StatusOK, &fetchedProject)
	if fetchedProject.ID != createdProject.ID {
		t.Errorf("expected project ID %q, got %q", createdProject.ID, fetchedProject.ID)
	}
}

// TestCORSIntegration verifies CORS behavior across preflight and actual requests.
func TestCORSIntegration(t *testing.T) {
	h := newE2EServer(t)

	// --- Test 1: Allowed preflight (OPTIONS) from allowed origin ---
	preflightReq := httptest.NewRequest(http.MethodOptions, "/projects", nil)
	preflightReq.Header.Set("Origin", "http://localhost")
	preflightReq.Header.Set("Access-Control-Request-Method", "POST")
	preflightReq.Header.Set("Access-Control-Request-Headers", "X-API-Key, Content-Type")

	preflightRec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(preflightRec, preflightReq)

	preflightResp := preflightRec.Result()
	defer preflightResp.Body.Close()

	if preflightResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected preflight status 204, got %d", preflightResp.StatusCode)
	}

	if origin := preflightResp.Header.Get("Access-Control-Allow-Origin"); origin != "http://localhost" {
		t.Fatalf("expected ACAO header 'http://localhost', got %q", origin)
	}

	if methods := preflightResp.Header.Get("Access-Control-Allow-Methods"); methods == "" {
		t.Fatal("expected Access-Control-Allow-Methods header on preflight response")
	}

	if headers := preflightResp.Header.Get("Access-Control-Allow-Headers"); headers == "" {
		t.Fatal("expected Access-Control-Allow-Headers header on preflight response")
	}

	// --- Test 2: Disallowed preflight from disallowed origin ---
	disallowedPreflight := httptest.NewRequest(http.MethodOptions, "/projects", nil)
	disallowedPreflight.Header.Set("Origin", "http://evil.com")

	disallowedRec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(disallowedRec, disallowedPreflight)

	disallowedResult := disallowedRec.Result()
	defer disallowedResult.Body.Close()

	if disallowedResult.StatusCode != http.StatusForbidden {
		t.Fatalf("expected disallowed preflight status 403, got %d", disallowedResult.StatusCode)
	}

	// Verify error body
	var errResp map[string]interface{}
	if err := json.NewDecoder(disallowedResult.Body).Decode(&errResp); err != nil {
		t.Fatalf("expected JSON error body, decode error: %v", err)
	}
	if errMsg, ok := errResp["error"]; !ok || errMsg != "origin not allowed" {
		t.Fatalf("expected error 'origin not allowed', got %v", errMsg)
	}

	// --- Test 3: Disallowed preflight should NOT set ACAO header ---
	if origin := disallowedResult.Header.Get("Access-Control-Allow-Origin"); origin != "" {
		t.Fatalf("disallowed preflight should not set ACAO, got %q", origin)
	}

	// --- Test 4: Actual POST request from allowed origin receives CORS headers ---
	actualReq := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"name":"CORS Test Project"}`))
	actualReq.Header.Set("Origin", "http://localhost")
	actualReq.Header.Set("Content-Type", "application/json")
	actualReq.Header.Set("X-API-Key", h.apiKey)

	actualRec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(actualRec, actualReq)

	actualResult := actualRec.Result()
	defer actualResult.Body.Close()

	if actualResult.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(actualResult.Body)
		t.Fatalf("expected 201, got %d; body: %s", actualResult.StatusCode, string(body))
	}

	if origin := actualResult.Header.Get("Access-Control-Allow-Origin"); origin != "http://localhost" {
		t.Fatalf("expected ACAO 'http://localhost' on actual response, got %q", origin)
	}

	if vary := actualResult.Header.Get("Vary"); vary != "Origin" {
		t.Fatalf("expected Vary: Origin, got %q", vary)
	}

	// --- Test 5: Actual request from disallowed origin gets 403 ---
	disallowedActual := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"name":"Evil Project"}`))
	disallowedActual.Header.Set("Origin", "http://evil.com")
	disallowedActual.Header.Set("Content-Type", "application/json")
	disallowedActual.Header.Set("X-API-Key", h.apiKey)

	disallowedActualRec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(disallowedActualRec, disallowedActual)

	disallowedActualResult := disallowedActualRec.Result()
	defer disallowedActualResult.Body.Close()

	if disallowedActualResult.StatusCode != http.StatusForbidden {
		t.Fatalf("expected disallowed actual request status 403, got %d", disallowedActualResult.StatusCode)
	}

	// --- Test 6: Request without Origin header passes through unchanged (no CORS overhead) ---
	noOriginReq := httptest.NewRequest(http.MethodGet, "/projects", nil)
	noOriginReq.Header.Set("X-API-Key", h.apiKey)

	noOriginRec := httptest.NewRecorder()
	h.server.Config.Handler.ServeHTTP(noOriginRec, noOriginReq)

	noOriginResult := noOriginRec.Result()
	defer noOriginResult.Body.Close()

	if noOriginResult.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for request without Origin, got %d", noOriginResult.StatusCode)
	}

	// Should NOT have ACAO header when no Origin was sent
	if origin := noOriginResult.Header.Get("Access-Control-Allow-Origin"); origin != "" {
		t.Fatalf("expected no ACAO when no Origin header, got %q", origin)
	}
}

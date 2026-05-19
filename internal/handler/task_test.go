package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

type taskHandlerSeed struct {
	projectID string
	taskID    string
}

type taskHandlerCase struct {
	name       string
	method     string
	path       func(baseURL string, seed *taskHandlerSeed) string
	body       any
	setup      func(t *testing.T, s *store.Store, body any) *taskHandlerSeed
	wantStatus int
	assert     func(t *testing.T, s *store.Store, seed *taskHandlerSeed, status int, body []byte)
}

func newTaskHandlerTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()

	testLogger := testLogger()

	s := newHandlerTestStore(t)
	svc := service.NewTaskService(s, testLogger)
	h := NewTaskHandler(svc, testLogger)

	mux := http.NewServeMux()
	RegisterTaskRoutes(mux, h)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, s
}

func seedTask(t *testing.T, s *store.Store, projectID string, title string) string {
	t.Helper()
	var newTask = &model.Task{
		ProjectID: projectID,
		Title:     title,
		Status:    model.StatusBacklog,
	}

	err := s.CreateTask(context.Background(), newTask)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	return newTask.ID
}

func decodeTaskResponse(t *testing.T, body []byte) model.Task {
	t.Helper()

	var task model.Task
	if err := json.Unmarshal(body, &task); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return task
}

func decodeTasksResponse(t *testing.T, body []byte) []model.Task {
	t.Helper()

	var tasks []model.Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return tasks
}

func TestTaskHandler(t *testing.T) {
	cases := []taskHandlerCase{
		{
			name:   "create task",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body: map[string]string{
				// ProjectID will be set in setup
				"title":  "Test Task",
				"status": string(model.StatusBacklog),
			},
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				body.(map[string]string)["project_id"] = projectID
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, s *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusCreated {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusCreated, string(body))
				}
				task := decodeTaskResponse(t, body)
				if task.ID == "" || task.Title != "Test Task" {
					t.Fatalf("unexpected created task: %+v", task)
				}
				if task.CreatedAt == "" || task.UpdatedAt == "" {
					t.Fatalf("expected timestamps to be set: %+v", task)
				}

				got, err := s.GetTask(context.Background(), task.ID)
				if err != nil {
					t.Fatalf("GetTask() error = %v", err)
				}
				if got.ID != task.ID || got.Title != task.Title {
					t.Fatalf("unexpected persisted task: %+v", got)
				}
			},
		},
		{
			name:   "create task missing fields",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body:       map[string]string{"title": "no project"},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
				if errBody := decodeErrorResponse(t, body); errBody["error"] == "" {
					t.Fatalf("expected error body, got %+v", errBody)
				}
			},
		},
		{
			name:   "create task title too long",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body: func() any {
				return map[string]string{
					"title": strings.Repeat("x", 501),
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				body.(map[string]string)["project_id"] = projectID
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "title") {
					t.Fatalf("expected title error, got %+v", errBody)
				}
			},
		},
		{
			name:   "create task description too long",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body: func() any {
				return map[string]string{
					"description": strings.Repeat("x", 5001),
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				body.(map[string]string)["project_id"] = projectID
				body.(map[string]string)["title"] = "ok title"
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "description") {
					t.Fatalf("expected description error, got %+v", errBody)
				}
			},
		},
		{
			name:   "create task too many tags",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body: func() any {
				tags := make([]string, 21)
				for i := 0; i < 21; i++ {
					tags[i] = "tag" + strconv.Itoa(i)
				}
				return map[string]any{
					"project_id": "", // set by setup
					"title":      "ok",
					"tags":       tags,
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				body.(map[string]any)["project_id"] = projectID
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "tags") {
					t.Fatalf("expected tags error, got %+v", errBody)
				}
			},
		},
		{
			name:   "create task invalid status",
			method: http.MethodPost,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			body: func() any {
				return map[string]string{
					"status": "bogus",
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				body.(map[string]string)["project_id"] = projectID
				body.(map[string]string)["title"] = "ok"
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "status") {
					t.Fatalf("expected status error, got %+v", errBody)
				}
			},
		},
		{
			name:   "update task title too long",
			method: http.MethodPut,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			body: func() any {
				return map[string]string{
					"title": strings.Repeat("x", 501),
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Alpha")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "title") {
					t.Fatalf("expected title error, got %+v", errBody)
				}
			},
		},
		{
			name:   "update task description too long",
			method: http.MethodPut,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			body: func() any {
				return map[string]string{
					"description": strings.Repeat("x", 5001),
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Alpha")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "description") {
					t.Fatalf("expected description error, got %+v", errBody)
				}
			},
		},
		{
			name:   "update task too many tags",
			method: http.MethodPut,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			body: func() any {
				tags := make([]string, 21)
				for i := 0; i < 21; i++ {
					tags[i] = "tag" + strconv.Itoa(i)
				}
				return map[string]any{
					"title": "ok",
					"tags":  tags,
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Alpha")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "tags") {
					t.Fatalf("expected tags error, got %+v", errBody)
				}
			},
		},
		{
			name:   "update task invalid status",
			method: http.MethodPut,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			body: func() any {
				return map[string]string{
					"status": "bogus",
				}
			}(),
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Alpha")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusBadRequest, string(body))
				}
				if errBody := decodeErrorResponse(t, body); !strings.Contains(errBody["error"], "status") {
					t.Fatalf("expected status error, got %+v", errBody)
				}
			},
		},
		{
			name:   "get task",
			method: http.MethodGet,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Read me")
				taskID := seedTask(t, s, projectID, "Alpha")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, s *store.Store, seed *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				task := decodeTaskResponse(t, body)
				if task.ID == "" || task.Title != "Alpha" {
					t.Fatalf("unexpected created task: %+v", task)
				}
				if task.CreatedAt == "" || task.UpdatedAt == "" {
					t.Fatalf("expected timestamps to be set: %+v", task)
				}

				got, err := s.GetTask(context.Background(), task.ID)
				if err != nil {
					t.Fatalf("GetTask() error = %v", err)
				}
				if got.ID != task.ID || got.Title != task.Title {
					t.Fatalf("unexpected persisted task: %+v", got)
				}
			},
		},
		{
			name:   "get missing task",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks/missing-id"
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
				}
				if errBody := decodeErrorResponse(t, body); errBody["error"] != "task not found" {
					t.Fatalf("error body = %+v, want task not found", errBody)
				}
			},
		},
		{
			name:   "update task",
			method: http.MethodPut,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			body: map[string]string{
				"id":     "WRONG-ID-SHOULD-BE-IGNORED",
				"title":  "Alpha Updated",
				"status": string(model.StatusBacklog),
			},
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Alpha")
				body.(map[string]string)["project_id"] = projectID
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, s *store.Store, seed *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusOK, string(body))
				}
				task := decodeTaskResponse(t, body)
				if task.ID != seed.taskID {
					t.Fatalf("response id = %q, want %q (path id must override body id)", task.ID, seed.taskID)
				}
				if task.Title != "Alpha Updated" {
					t.Fatalf("title = %q, want Alpha Updated", task.Title)
				}

				got, err := s.GetTask(context.Background(), seed.taskID)
				if err != nil {
					t.Fatalf("GetTask() error = %v", err)
				}
				if got.Title != "Alpha Updated" {
					t.Fatalf("persisted title = %q, want Alpha Updated", got.Title)
				}

				if _, err := s.GetTask(context.Background(), "WRONG-ID-SHOULD-BE-IGNORED"); err == nil {
					t.Fatal("body id should not have created/updated a separate task")
				}
			},
		},
		{
			name:   "update missing task",
			method: http.MethodPut,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks/missing-id"
			},
			body: map[string]string{
				"title":  "X",
				"status": string(model.StatusBacklog),
			},
			setup: func(t *testing.T, s *store.Store, body any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				body.(map[string]string)["project_id"] = projectID
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusNotFound, string(body))
				}
			},
		},
		{
			name:   "move task valid transition",
			method: http.MethodPatch,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID + "/move"
			},
			body: map[string]string{"status": string(model.StatusTodo)},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "MoveMe")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, s *store.Store, seed *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusOK, string(body))
				}
				task := decodeTaskResponse(t, body)
				if task.Status != model.StatusTodo {
					t.Fatalf("response status = %q, want %q", task.Status, model.StatusTodo)
				}
				got, err := s.GetTask(context.Background(), seed.taskID)
				if err != nil {
					t.Fatalf("GetTask() error = %v", err)
				}
				if got.Status != model.StatusTodo {
					t.Fatalf("persisted status = %q, want %q", got.Status, model.StatusTodo)
				}
			},
		},
		{
			name:   "move task invalid transition",
			method: http.MethodPatch,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID + "/move"
			},
			body: map[string]string{"status": string(model.StatusDone)},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "Stuck")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusUnprocessableEntity {
					t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
				}
				if errBody := decodeErrorResponse(t, body); errBody["error"] == "" {
					t.Fatalf("expected error body, got %+v", errBody)
				}
			},
		},
		{
			name:   "move missing task",
			method: http.MethodPatch,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks/missing-id/move"
			},
			body:       map[string]string{"status": string(model.StatusTodo)},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusNotFound, string(body))
				}
			},
		},
		{
			name:   "move task missing status",
			method: http.MethodPatch,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID + "/move"
			},
			body:       map[string]string{},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "MoveNoStatus")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
				if errBody := decodeErrorResponse(t, body); errBody["error"] != "status is required" {
					t.Fatalf("error body = %+v, want 'status is required'", errBody)
				}
			},
		},
		{
			name:   "move task invalid status",
			method: http.MethodPatch,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID + "/move"
			},
			body: map[string]string{"status": "bogus-status"},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "MoveBadStatus")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
				if errBody := decodeErrorResponse(t, body); errBody["error"] != "invalid status" {
					t.Fatalf("error body = %+v, want 'invalid status'", errBody)
				}
			},
		},
		{
			name:   "delete task",
			method: http.MethodDelete,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks/" + seed.taskID
			},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "Proj")
				taskID := seedTask(t, s, projectID, "GoneSoon")
				return &taskHandlerSeed{projectID: projectID, taskID: taskID}
			},
			wantStatus: http.StatusNoContent,
			assert: func(t *testing.T, s *store.Store, seed *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusNoContent {
					t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
				}
				if len(body) != 0 {
					t.Fatalf("body = %q, want empty", string(body))
				}
				if _, err := s.GetTask(context.Background(), seed.taskID); err == nil {
					t.Fatal("expected deleted task to be missing")
				}
			},
		},
		{
			name:   "delete missing task",
			method: http.MethodDelete,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks/missing-id"
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
				}
			},
		},
		{
			name:   "list tasks empty",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks"
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				if string(body) != "[]" {
					t.Fatalf("body = %s, want []", string(body))
				}
			},
		},
		{
			name:   "list tasks filter by project",
			method: http.MethodGet,
			path: func(baseURL string, seed *taskHandlerSeed) string {
				return baseURL + "/tasks?project_id=" + seed.projectID
			},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projA := seedProject(t, s, "A")
				projB := seedProject(t, s, "B")
				seedTask(t, s, projA, "a1")
				seedTask(t, s, projA, "a2")
				seedTask(t, s, projB, "b1")
				return &taskHandlerSeed{projectID: projA}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, seed *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusOK, string(body))
				}
				tasks := decodeTasksResponse(t, body)
				if len(tasks) != 2 {
					t.Fatalf("len(tasks) = %d, want 2", len(tasks))
				}
				for _, tk := range tasks {
					if tk.ProjectID != seed.projectID {
						t.Fatalf("task project_id = %q, want %q", tk.ProjectID, seed.projectID)
					}
				}
			},
		},
		{
			name:   "list tasks filter by status",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks?status=todo"
			},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				ctx := context.Background()
				projectID := seedProject(t, s, "P")
				seedTask(t, s, projectID, "back")
				todoTaskID := seedTask(t, s, projectID, "todo1")
				if err := s.UpdateTaskStatus(ctx, todoTaskID, model.StatusTodo); err != nil {
					t.Fatalf("UpdateTaskStatus() error = %v", err)
				}
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusOK, string(body))
				}
				tasks := decodeTasksResponse(t, body)
				if len(tasks) != 1 {
					t.Fatalf("len(tasks) = %d, want 1", len(tasks))
				}
				if tasks[0].Status != model.StatusTodo {
					t.Fatalf("status = %q, want todo", tasks[0].Status)
				}
			},
		},
		{
			name:   "list tasks limit and offset",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks?limit=2&offset=1"
			},
			setup: func(t *testing.T, s *store.Store, _ any) *taskHandlerSeed {
				projectID := seedProject(t, s, "P")
				seedTask(t, s, projectID, "t1")
				seedTask(t, s, projectID, "t2")
				seedTask(t, s, projectID, "t3")
				seedTask(t, s, projectID, "t4")
				return &taskHandlerSeed{projectID: projectID}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusOK, string(body))
				}
				tasks := decodeTasksResponse(t, body)
				if len(tasks) != 2 {
					t.Fatalf("len(tasks) = %d, want 2", len(tasks))
				}
			},
		},
		{
			name:   "list tasks invalid limit",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks?limit=abc"
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
			},
		},
		{
			name:   "list tasks negative offset",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks?offset=-1"
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
			},
		},
		{
			name:   "list tasks invalid status",
			method: http.MethodGet,
			path: func(baseURL string, _ *taskHandlerSeed) string {
				return baseURL + "/tasks?status=bogus"
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *taskHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			server, s := newTaskHandlerTestServer(t)
			client := server.Client()

			var seed *taskHandlerSeed
			if tt.setup != nil {
				seed = tt.setup(t, s, tt.body)
			}

			resp, body := doJSONRequest(t, client, tt.method, tt.path(server.URL, seed), tt.body)
			if tt.assert != nil {
				tt.assert(t, s, seed, resp.StatusCode, body)
				return
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

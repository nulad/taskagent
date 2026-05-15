package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

type projectHandlerSeed struct {
	projectID string
}

type projectHandlerCase struct {
	name       string
	method     string
	path       func(baseURL string, seed *projectHandlerSeed) string
	body       any
	setup      func(t *testing.T, s *store.Store) *projectHandlerSeed
	wantStatus int
	assert     func(t *testing.T, s *store.Store, seed *projectHandlerSeed, status int, body []byte)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newHandlerTestStore(t *testing.T) *store.Store {
	t.Helper()

	s, err := store.NewStore(":memory:", testLogger())
	if err != nil {
		t.Fatalf("store.NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	return s
}

func newProjectHandlerTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()

	testLogger := testLogger()

	s := newHandlerTestStore(t)
	svc := service.NewProjectService(s, testLogger)
	h := NewProjectHandler(svc, testLogger)

	mux := http.NewServeMux()
	RegisterProjectRoutes(mux, h)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, s
}

func seedProject(t *testing.T, s *store.Store, name string) string {
	t.Helper()

	id, err := s.CreateProject(context.Background(), &model.Project{Name: name})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	return id
}

func seedProjectWithTask(t *testing.T, s *store.Store, name string) *projectHandlerSeed {
	t.Helper()

	projectID := seedProject(t, s, name)
	task := &model.Task{
		ProjectID: projectID,
		Title:     "Task for conflict",
		Status:    model.StatusBacklog,
	}
	if err := s.CreateTask(context.Background(), task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	return &projectHandlerSeed{projectID: projectID}
}

func doJSONRequest(t *testing.T, client *http.Client, method, url string, body any) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}

	return resp, respBody
}

func decodeProjectResponse(t *testing.T, body []byte) model.Project {
	t.Helper()

	var project model.Project
	if err := json.Unmarshal(body, &project); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return project
}

func decodeProjectsResponse(t *testing.T, body []byte) []model.Project {
	t.Helper()

	var projects []model.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return projects
}

func decodeErrorResponse(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return resp
}

func TestProjectHandler(t *testing.T) {
	cases := []projectHandlerCase{
		{
			name:   "create project",
			method: http.MethodPost,
			path: func(baseURL string, _ *projectHandlerSeed) string {
				return baseURL + "/projects"
			},
			body: map[string]string{
				"name":        "Alpha",
				"description": "first project",
			},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, s *store.Store, _ *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusCreated {
					t.Fatalf("status = %d, want %d", status, http.StatusCreated)
				}
				project := decodeProjectResponse(t, body)
				if project.ID == "" || project.Name != "Alpha" || project.Description != "first project" {
					t.Fatalf("unexpected created project: %+v", project)
				}
				if project.CreatedAt == "" || project.UpdatedAt == "" {
					t.Fatalf("expected timestamps to be set: %+v", project)
				}

				got, err := s.GetProject(context.Background(), project.ID)
				if err != nil {
					t.Fatalf("GetProject() error = %v", err)
				}
				if got.ID != project.ID || got.Name != project.Name || got.Description != project.Description {
					t.Fatalf("unexpected persisted project: %+v", got)
				}
			},
		},
		{
			name:   "get project",
			method: http.MethodGet,
			path: func(baseURL string, seed *projectHandlerSeed) string {
				return baseURL + "/projects/" + seed.projectID
			},
			setup: func(t *testing.T, s *store.Store) *projectHandlerSeed {
				return &projectHandlerSeed{projectID: seedProject(t, s, "Read me")}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, seed *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				project := decodeProjectResponse(t, body)
				if project.ID != seed.projectID || project.Name != "Read me" {
					t.Fatalf("unexpected project: %+v", project)
				}
			},
		},
		{
			name:   "list projects empty",
			method: http.MethodGet,
			path: func(baseURL string, _ *projectHandlerSeed) string {
				return baseURL + "/projects"
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, _ *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				projects := decodeProjectsResponse(t, body)
				if len(projects) != 0 {
					t.Fatalf("len(projects) = %d, want 0", len(projects))
				}
				if string(body) != "[]" {
					t.Fatalf("body = %s, want []", string(body))
				}
			},
		},
		{
			name:   "update project",
			method: http.MethodPut,
			path: func(baseURL string, seed *projectHandlerSeed) string {
				return baseURL + "/projects/" + seed.projectID
			},
			body: map[string]string{
				"name":        "Alpha Updated",
				"description": "updated project",
			},
			setup: func(t *testing.T, s *store.Store) *projectHandlerSeed {
				return &projectHandlerSeed{projectID: seedProject(t, s, "Alpha")}
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, s *store.Store, seed *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				project := decodeProjectResponse(t, body)
				if project.ID != seed.projectID || project.Name != "Alpha Updated" || project.Description != "updated project" {
					t.Fatalf("unexpected updated project: %+v", project)
				}
				if project.UpdatedAt == "" {
					t.Fatalf("expected updated_at to be set: %+v", project)
				}

				got, err := s.GetProject(context.Background(), seed.projectID)
				if err != nil {
					t.Fatalf("GetProject() error = %v", err)
				}
				if got.Name != "Alpha Updated" || got.Description != "updated project" {
					t.Fatalf("unexpected persisted project after update: %+v", got)
				}
			},
		},
		{
			name:   "delete project",
			method: http.MethodDelete,
			path: func(baseURL string, seed *projectHandlerSeed) string {
				return baseURL + "/projects/" + seed.projectID
			},
			setup: func(t *testing.T, s *store.Store) *projectHandlerSeed {
				return &projectHandlerSeed{projectID: seedProject(t, s, "Delete me")}
			},
			wantStatus: http.StatusNoContent,
			assert: func(t *testing.T, s *store.Store, seed *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusNoContent {
					t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
				}
				if len(body) != 0 {
					t.Fatalf("body = %q, want empty body", string(body))
				}

				_, err := s.GetProject(context.Background(), seed.projectID)
				if err == nil {
					t.Fatal("expected deleted project to be missing")
				}
			},
		},
		{
			name:   "get missing project",
			method: http.MethodGet,
			path: func(baseURL string, _ *projectHandlerSeed) string {
				return baseURL + "/projects/missing-id"
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
				}
				errBody := decodeErrorResponse(t, body)
				if errBody["error"] != "project not found" {
					t.Fatalf("error body = %+v, want project not found", errBody)
				}
			},
		},
		{
			name:   "delete project with tasks",
			method: http.MethodDelete,
			path: func(baseURL string, seed *projectHandlerSeed) string {
				return baseURL + "/projects/" + seed.projectID
			},
			setup: func(t *testing.T, s *store.Store) *projectHandlerSeed {
				return seedProjectWithTask(t, s, "Blocked")
			},
			wantStatus: http.StatusConflict,
			assert: func(t *testing.T, _ *store.Store, _ *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusConflict {
					t.Fatalf("status = %d, want %d", status, http.StatusConflict)
				}
				errBody := decodeErrorResponse(t, body)
				if errBody["error"] == "" {
					t.Fatalf("expected conflict error body, got %+v", errBody)
				}
			},
		},
		{
			name:   "delete missing project",
			method: http.MethodDelete,
			path: func(baseURL string, _ *projectHandlerSeed) string {
				return baseURL + "/projects/missing-id"
			},
			wantStatus: http.StatusNotFound,
			assert: func(t *testing.T, _ *store.Store, _ *projectHandlerSeed, status int, body []byte) {
				if status != http.StatusNotFound {
					t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
				}
				errBody := decodeErrorResponse(t, body)
				if errBody["error"] != "project not found" {
					t.Fatalf("error body = %+v, want project not found", errBody)
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			server, s := newProjectHandlerTestServer(t)
			client := server.Client()

			var seed *projectHandlerSeed
			if tt.setup != nil {
				seed = tt.setup(t, s)
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

package store

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTaskStore_FullCRUDCycle(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	projectID, err := s.CreateProject(ctx, &Project{Name: "Task CRUD Project"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	task := &Task{
		ProjectID:   projectID,
		Title:       "Initial title",
		Description: "Initial description",
		Status:      "todo",
		Tags:        []string{"backend", "priority-high"},
	}

	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.ID == "" {
		t.Fatal("expected task ID to be set")
	}
	if task.CreatedAt == "" || task.UpdatedAt == "" {
		t.Fatal("expected timestamps to be set")
	}

	got, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if got.ProjectID != projectID || got.Title != "Initial title" || got.Status != "todo" {
		t.Fatalf("unexpected task after create: %+v", got)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "backend" || got.Tags[1] != "priority-high" {
		t.Fatalf("unexpected tags after create: %+v", got.Tags)
	}

	task.Title = "Updated title"
	task.Description = "Updated description"
	task.Status = "in-progress"
	task.Tags = []string{"backend", "updated"}
	if err := s.UpdateTask(ctx, task); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	updated, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() after update error = %v", err)
	}
	if updated.Title != "Updated title" || updated.Description != "Updated description" || updated.Status != "in-progress" {
		t.Fatalf("unexpected task after update: %+v", updated)
	}
	if len(updated.Tags) != 2 || updated.Tags[1] != "updated" {
		t.Fatalf("unexpected tags after update: %+v", updated.Tags)
	}

	if err := s.DeleteTask(ctx, task.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	_, err = s.GetTask(ctx, task.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetTask() after delete error = %v, want %v", err, ErrNotFound)
	}
}

func TestTaskStore_ListTasks_WithFilters(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	projectA, err := s.CreateProject(ctx, &Project{Name: "List Project A"})
	if err != nil {
		t.Fatalf("CreateProject(A) error = %v", err)
	}
	projectB, err := s.CreateProject(ctx, &Project{Name: "List Project B"})
	if err != nil {
		t.Fatalf("CreateProject(B) error = %v", err)
	}

	fixtures := []Task{
		{ProjectID: projectA, Title: "A-backlog", Tags: []string{"a", "backlog"}}, // defaults to backlog
		{ProjectID: projectA, Title: "A-todo", Status: "todo", Tags: []string{"a", "todo"}},
		{ProjectID: projectA, Title: "A-done", Status: "done", Tags: []string{"a", "done"}},
		{ProjectID: projectB, Title: "B-review", Status: "review", Tags: []string{"b", "review"}},
	}
	for i := range fixtures {
		if err := s.CreateTask(ctx, &fixtures[i]); err != nil {
			t.Fatalf("CreateTask(fixture %d) error = %v", i, err)
		}
	}

	todo := "todo"
	review := "review"
	projectAFilter := projectA
	missingStatus := "missing-status"

	tests := []struct {
		name         string
		filter       TaskFilter
		wantLen      int
		wantProject  *string
		wantStatus   *string
	}{
		{
			name:    "no filters",
			filter:  TaskFilter{},
			wantLen: 4,
		},
		{
			name:        "filter by project",
			filter:      TaskFilter{ProjectID: &projectAFilter},
			wantLen:     3,
			wantProject: &projectAFilter,
		},
		{
			name:       "filter by status",
			filter:     TaskFilter{Status: &review},
			wantLen:    1,
			wantStatus: &review,
		},
		{
			name:        "filter by project and status",
			filter:      TaskFilter{ProjectID: &projectAFilter, Status: &todo},
			wantLen:     1,
			wantProject: &projectAFilter,
			wantStatus:  &todo,
		},
		{
			name:    "limit and offset",
			filter:  TaskFilter{Limit: 2, Offset: 1},
			wantLen: 2,
		},
		{
			name:    "unknown status returns empty",
			filter:  TaskFilter{Status: &missingStatus},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ListTasks(ctx, tt.filter)
			if err != nil {
				t.Fatalf("ListTasks() error = %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len(ListTasks()) = %d, want %d", len(got), tt.wantLen)
			}

			for _, task := range got {
				if tt.wantProject != nil && task.ProjectID != *tt.wantProject {
					t.Fatalf("task project = %q, want %q", task.ProjectID, *tt.wantProject)
				}
				if tt.wantStatus != nil && task.Status != *tt.wantStatus {
					t.Fatalf("task status = %q, want %q", task.Status, *tt.wantStatus)
				}
			}
		})
	}
}

func TestTaskStore_UpdateTaskStatus(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	projectID, err := s.CreateProject(ctx, &Project{Name: "Status Project"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	task := &Task{ProjectID: projectID, Title: "Move me"}
	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if err := s.UpdateTaskStatus(ctx, task.ID, "todo"); err != nil {
		t.Fatalf("UpdateTaskStatus() error = %v", err)
	}

	got, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if got.Status != "todo" {
		t.Fatalf("status = %q, want %q", got.Status, "todo")
	}
}

func TestTaskStore_UpdateTaskStatus_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		status  string
		wantErr error
	}{
		{name: "empty id", id: "", status: "todo", wantErr: ErrInvalidInput},
		{name: "empty status", id: "task-id", status: "", wantErr: ErrInvalidInput},
		{name: "invalid status", id: "task-id", status: "invalid", wantErr: ErrInvalidInput},
		{name: "missing task", id: "missing-id", status: "todo", wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateTaskStatus(ctx, tt.id, tt.status)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateTaskStatus() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskStore_CreateTask_NonExistentProject_ReturnsForeignKeyError(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := &Task{
		ProjectID: "missing-project-id",
		Title:     "orphan task",
	}

	err := s.CreateTask(ctx, task)
	if err == nil {
		t.Fatal("CreateTask() expected foreign key error, got nil")
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "FOREIGN KEY") {
		t.Fatalf("CreateTask() error = %v, want foreign key error", err)
	}
}

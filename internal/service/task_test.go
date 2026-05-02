package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

func newTestService(t *testing.T) (*TaskService, *store.Store) {
	t.Helper()

	s, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	return NewTaskService(s), s
}

func seedProject(t *testing.T, s *store.Store) string {
	t.Helper()
	ctx := context.Background()
	id, err := s.CreateProject(ctx, &model.Project{Name: "Test Project"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	return id
}

func seedTask(t *testing.T, s *store.Store, projectID string, status model.TaskStatus) string {
	t.Helper()
	ctx := context.Background()
	task := &model.Task{
		ProjectID: projectID,
		Title:     "Test Task",
		Status:    status,
	}
	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	return task.ID
}

func TestTaskService_MoveTask_ValidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from model.TaskStatus
		to   model.TaskStatus
	}{
		{
			name: "backlog to todo",
			from: model.StatusBacklog,
			to:   model.StatusTodo,
		},
		{
			name: "backlog to closed",
			from: model.StatusBacklog,
			to:   model.StatusClosed,
		},
		{
			name: "todo to in-progress",
			from: model.StatusTodo,
			to:   model.StatusInProgress,
		},
		{
			name: "todo to backlog",
			from: model.StatusTodo,
			to:   model.StatusBacklog,
		},
		{
			name: "todo to closed",
			from: model.StatusTodo,
			to:   model.StatusClosed,
		},
		{
			name: "in-progress to review",
			from: model.StatusInProgress,
			to:   model.StatusReview,
		},
		{
			name: "in-progress to todo",
			from: model.StatusInProgress,
			to:   model.StatusTodo,
		},
		{
			name: "in-progress to closed",
			from: model.StatusInProgress,
			to:   model.StatusClosed,
		},
		{
			name: "review to done",
			from: model.StatusReview,
			to:   model.StatusDone,
		},
		{
			name: "review to in-progress",
			from: model.StatusReview,
			to:   model.StatusInProgress,
		},
		{
			name: "review to closed",
			from: model.StatusReview,
			to:   model.StatusClosed,
		},
		{
			name: "done to closed",
			from: model.StatusDone,
			to:   model.StatusClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			svc, s := newTestService(t)
			ctx := context.Background()
			projectID := seedProject(t, s)
			taskID := seedTask(t, s, projectID, tt.from)

			// act
			err := svc.MoveTask(ctx, taskID, tt.to)

			// assert
			if err != nil {
				t.Fatalf("MoveTask() error = %v, want nil", err)
			}
			updatedTask, err := s.GetTask(ctx, taskID)
			if err != nil {
				t.Fatalf("GetTask() error = %v", err)
			}
			if updatedTask.Status != tt.to {
				t.Fatalf("MoveTask() status = %v, want %v", updatedTask.Status, tt.to)
			}
		})
	}
}

func TestTaskService_MoveTask_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from model.TaskStatus
		to   model.TaskStatus
	}{
		{
			name: "backlog to done",
			from: model.StatusBacklog,
			to:   model.StatusDone,
		},
		{
			name: "backlog to review",
			from: model.StatusBacklog,
			to:   model.StatusReview,
		},
		{
			name: "closed is terminal",
			from: model.StatusClosed,
			to:   model.StatusTodo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			svc, s := newTestService(t)
			ctx := context.Background()
			projectID := seedProject(t, s)
			taskID := seedTask(t, s, projectID, tt.from)

			// act
			err := svc.MoveTask(ctx, taskID, tt.to)

			// assert
			var target *InvalidTransitionError
			if !errors.As(err, &target) {
				t.Fatalf("MoveTask() error = %v, want InvalidTransitionError", err)
			}
		})
	}
}

func TestTaskService_MoveTask_TaskNotFound(t *testing.T) {
	// arrange
	svc, _ := newTestService(t)
	ctx := context.Background()

	// act
	err := svc.MoveTask(ctx, "-1", model.StatusInProgress)

	// assert
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("MoveTask() error = %v, want ErrNotFound", err)
	}
}

func TestTaskService_CreateTask_ProjectExists(t *testing.T) {
	svc, s := newTestService(t)
	ctx := context.Background()
	projectID := seedProject(t, s)

	task := &model.Task{
		ProjectID: projectID,
		Title:     "New Task",
		Status:    model.StatusTodo,
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v, want nil", err)
	}
	if task.ID == "" {
		t.Fatalf("CreateTask() did not set task.ID")
	}
}

func TestTaskService_CreateTask_ProjectMissing(t *testing.T) {
	svc, s := newTestService(t)
	ctx := context.Background()
	_ = seedProject(t, s)

	task := &model.Task{
		ProjectID: "bogus-id",
		Title:     "New Task",
		Status:    model.StatusTodo,
	}
	err := svc.CreateTask(ctx, task)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("CreateTask() error = %v, want ErrNotFound", err)
	}
}

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

func newProjectTestService(t *testing.T) (*ProjectService, *store.Store) {
	t.Helper()

	s, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	return NewProjectService(s), s
}

func TestProjectService_DeleteProject_NoTasks(t *testing.T) {
	svc, s := newProjectTestService(t)
	ctx := context.Background()

	// Create a project
	projectID := seedProject(t, s)

	// Attempt to delete the project
	err := svc.DeleteProject(ctx, projectID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify the project was deleted
	_, err = svc.GetProject(ctx, projectID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("GetProject() expected error %v, got %v", store.ErrNotFound, err)
	}
}

func TestProjectService_DeleteProject_WithTasks(t *testing.T) {
	svc, s := newProjectTestService(t)
	ctx := context.Background()

	// Create a project
	projectID := seedProject(t, s)

	// Create a task associated with the project
	task := &model.Task{
		ProjectID: projectID,
		Title:     "Test Task",
		Status:    model.StatusBacklog,
	}
	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Attempt to delete the project
	err := svc.DeleteProject(ctx, projectID)
	var target *ProjectHasTasksError
	if !errors.As(err, &target) {
		t.Fatalf("DeleteProject() expected error ProjectHasTasksError, got %v", err)
	}

	errMsg := target.Error()
	expectedMsg := "project " + projectID + " has 1 associated task(s)"
	if errMsg != expectedMsg {
		t.Fatalf("DeleteProject() expected error message %q, got %q", expectedMsg, errMsg)
	}
}

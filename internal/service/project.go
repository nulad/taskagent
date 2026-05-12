package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

type ProjectService struct {
	store  *store.Store
	logger *slog.Logger
}

func NewProjectService(store *store.Store, logger *slog.Logger) *ProjectService {
	return &ProjectService{store: store, logger: logger}
}

type ProjectHasTasksError struct {
	TasksCount int
	ID         string
}

func (e *ProjectHasTasksError) Error() string {
	return fmt.Sprintf("project %s has %d associated task(s)", e.ID, e.TasksCount)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *model.Project) (string, error) {
	return s.store.CreateProject(ctx, project)
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (model.Project, error) {
	return s.store.GetProject(ctx, id)
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *model.Project) error {
	return s.store.UpdateProject(ctx, project)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	tasks, err := s.store.ListTasks(ctx, model.TaskFilter{ProjectID: &id})
	if err != nil {
		return err
	}

	if len(tasks) > 0 {
		return &ProjectHasTasksError{TasksCount: len(tasks), ID: id}
	}
	return s.store.DeleteProject(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]model.Project, error) {
	return s.store.ListProjects(ctx)
}

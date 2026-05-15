package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nulad/taskagent/internal/logging"
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
	id, err := s.store.CreateProject(ctx, project)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to create project", err, slog.String("name", project.Name))
		return "", err
	}
	return id, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (model.Project, error) {
	project, err := s.store.GetProject(ctx, id)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get project", err, slog.String("id", id))
		return model.Project{}, err
	}
	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *model.Project) error {
	err := s.store.UpdateProject(ctx, project)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to update project", err, slog.String("id", project.ID))
		return err
	}
	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	tasks, err := s.store.ListTasks(ctx, model.TaskFilter{ProjectID: &id})
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to list tasks for project", err, slog.String("project_id", id))
		return err
	}

	if len(tasks) > 0 {
		return &ProjectHasTasksError{TasksCount: len(tasks), ID: id}
	}
	err = s.store.DeleteProject(ctx, id)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to delete project", err, slog.String("project_id", id))
		return err
	}
	return nil
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]model.Project, error) {
	projects, err := s.store.ListProjects(ctx)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to list projects", err)
		return nil, err
	}
	return projects, nil
}

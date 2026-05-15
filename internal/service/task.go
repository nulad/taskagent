package service

import (
	"context"
	"log/slog"

	"github.com/nulad/taskagent/internal/logging"
	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

type TaskService struct {
	store  *store.Store
	logger *slog.Logger
}

func NewTaskService(store *store.Store, logger *slog.Logger) *TaskService {
	return &TaskService{store: store, logger: logger}
}

type InvalidTransitionError struct {
	From model.TaskStatus
	To   model.TaskStatus
}

func (e *InvalidTransitionError) Error() string {
	return "invalid status transition from " + string(e.From) + " to " + string(e.To)
}

var validStatusFlows = map[model.TaskStatus][]model.TaskStatus{
	model.StatusBacklog:    {model.StatusTodo, model.StatusClosed},
	model.StatusTodo:       {model.StatusBacklog, model.StatusInProgress, model.StatusClosed},
	model.StatusInProgress: {model.StatusTodo, model.StatusReview, model.StatusClosed},
	model.StatusReview:     {model.StatusInProgress, model.StatusDone, model.StatusClosed},
	model.StatusDone:       {model.StatusClosed},
}

func isValidStatusTransition(current, nextStatus model.TaskStatus) bool {
	allowedNextStatuses, exists := validStatusFlows[current]
	if !exists {
		return false
	}
	for _, allowed := range allowedNextStatuses {
		if nextStatus == allowed {
			return true
		}
	}
	return false
}

func (s *TaskService) MoveTask(ctx context.Context, id string, newStatus model.TaskStatus) error {
	task, err := s.store.GetTask(ctx, id)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get task for move", err, slog.String("task_id", id))
		return err
	}
	currentStatus := task.Status
	if !isValidStatusTransition(currentStatus, newStatus) {
		logging.LogWithError(ctx, s.logger, "failed to move task", &InvalidTransitionError{From: currentStatus, To: newStatus}, slog.String("task_id", id))
		return &InvalidTransitionError{From: currentStatus, To: newStatus}
	}
	return s.store.UpdateTaskStatus(ctx, id, newStatus)
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) error {
	_, err := s.store.GetProject(ctx, task.ProjectID)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get project for task", err, slog.String("project_id", task.ProjectID))
		return err
	}

	err = s.store.CreateTask(ctx, task)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to create task", err, slog.String("project_id", task.ProjectID))
		return err
	}
	return nil
}
func (s *TaskService) GetTask(ctx context.Context, id string) (model.Task, error) {
	task, err := s.store.GetTask(ctx, id)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get task", err, slog.String("task_id", id))
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	tasks, err := s.store.ListTasks(ctx, filter)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to list tasks", err, slog.Any("filter", filter))
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, task *model.Task) error {
	err := s.store.UpdateTask(ctx, task)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to update task", err, slog.String("task_id", task.ID))
		return err
	}
	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	err := s.store.DeleteTask(ctx, id)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to delete task", err, slog.String("task_id", id))
		return err
	}
	return nil
}

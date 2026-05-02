package service

import (
	"context"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

type TaskService struct {
	store *store.Store
}

func NewTaskService(store *store.Store) *TaskService {
	return &TaskService{store: store}
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
		return err
	}
	currentStatus := task.Status
	if !isValidStatusTransition(currentStatus, newStatus) {
		return &InvalidTransitionError{From: currentStatus, To: newStatus}
	}
	return s.store.UpdateTaskStatus(ctx, id, newStatus)
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) error {
	_, err := s.store.GetProject(ctx, task.ProjectID)
	if err != nil {
		return err
	}

	return s.store.CreateTask(ctx, task)
}
func (s *TaskService) GetTask(ctx context.Context, id string) (model.Task, error) {
	return s.store.GetTask(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	return s.store.ListTasks(ctx, filter)
}

func (s *TaskService) UpdateTask(ctx context.Context, task *model.Task) error {
	return s.store.UpdateTask(ctx, task)
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	return s.store.DeleteTask(ctx, id)
}

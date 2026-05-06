package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

// TODO log properly the error instead of just writing a generic message to the client

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func RegisterTaskRoutes(mux *http.ServeMux, handler *TaskHandler) {
	mux.HandleFunc("POST /tasks", handler.handleCreate)
	mux.HandleFunc("GET /tasks/{id}", handler.handleGet)
	mux.HandleFunc("PUT /tasks/{id}", handler.handleUpdate)
	mux.HandleFunc("PATCH /tasks/{id}/move", handler.handleMove)
	mux.HandleFunc("DELETE /tasks/{id}", handler.handleDelete)
	mux.HandleFunc("GET /tasks", handler.handleList)
}

func (h *TaskHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var task model.Task
	err := readJSON(r, &task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if task.ProjectID == "" || task.Title == "" {
		writeError(w, http.StatusUnprocessableEntity, "project_id and title are required")
		return
	}

	err = h.service.CreateTask(r.Context(), &task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	resultTask, err := h.service.GetTask(r.Context(), task.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}
	writeJSON(w, http.StatusCreated, resultTask)

}

func (h *TaskHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	var taskID = pathParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var taskID = pathParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	var task model.Task
	err := readJSON(r, &task)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task.ID = taskID

	err = h.service.UpdateTask(r.Context(), &task)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, store.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid task data")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	resultTask, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}
	writeJSON(w, http.StatusOK, resultTask)
}

type moveRequest struct {
	Status model.TaskStatus `json:"status"`
}

func (h *TaskHandler) handleMove(w http.ResponseWriter, r *http.Request) {
	var taskID = pathParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	var moveStatus moveRequest
	err := readJSON(r, &moveStatus)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.MoveTask(r.Context(), taskID, moveStatus.Status)
	if err != nil {
		var target *service.InvalidTransitionError
		if errors.As(err, &target) {
			writeError(w, http.StatusUnprocessableEntity, target.Error())
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to move task")
		return
	}

	resultTask, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}
	writeJSON(w, http.StatusOK, resultTask)
}

func (h *TaskHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var taskID = pathParam(r, "id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	err := h.service.DeleteTask(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *TaskHandler) handleList(w http.ResponseWriter, r *http.Request) {

	var projectID = r.URL.Query().Get("project_id")
	var status = r.URL.Query().Get("status")
	var limit = r.URL.Query().Get("limit")
	var offset = r.URL.Query().Get("offset")

	var filter = model.TaskFilter{}

	if projectID != "" {
		filter.ProjectID = &projectID
	}

	filter.Limit = 50
	if limit != "" {
		lim, err := strconv.Atoi(limit)
		if err != nil || lim <= 0 {
			writeError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		filter.Limit = lim
	}

	filter.Offset = 0
	if offset != "" {
		off, err := strconv.Atoi(offset)
		if err != nil || off < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		filter.Offset = off
	}

	if status != "" {
		if !model.ValidStatus(status) {
			writeError(w, http.StatusBadRequest, "invalid status parameter")
			return
		}
		taskStatus := model.TaskStatus(status)
		filter.Status = &taskStatus
	}

	tasks, err := h.service.ListTasks(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	if len(tasks) == 0 {
		tasks = []model.Task{} // Return an empty array instead of null
	}
	writeJSON(w, http.StatusOK, tasks)
}

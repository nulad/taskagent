package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/service"
	"github.com/nulad/taskagent/internal/store"
)

type ProjectHandler struct {
	service *service.ProjectService
	logger  *slog.Logger
}

func NewProjectHandler(service *service.ProjectService, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{service: service, logger: logger}
}

func RegisterProjectRoutes(mux *http.ServeMux, handler *ProjectHandler) {
	mux.HandleFunc("POST /projects", handler.handleCreate)
	mux.HandleFunc("GET /projects/{id}", handler.handleGet)
	mux.HandleFunc("PUT /projects/{id}", handler.handleUpdate)
	mux.HandleFunc("DELETE /projects/{id}", handler.handleDelete)
	mux.HandleFunc("GET /projects", handler.handleList)
}

func (h *ProjectHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var project model.Project
	err := readJSON(r, &project)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if project.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "project name is required")
		return
	}

	createdProjectID, err := h.service.CreateProject(r.Context(), &project)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create project",
			"error", err,
			"name", project.Name,
		)
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	resultProject, err := h.service.GetProject(r.Context(), createdProjectID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to retrieve project",
			"error", err,
			"project_id", createdProjectID,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve project")
		return
	}
	writeJSON(w, http.StatusCreated, resultProject)
}

func (h *ProjectHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	var projectID = pathParam(r, "id")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project ID is required")
		return
	}

	project, err := h.service.GetProject(r.Context(), projectID)
	if err != nil {

		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
		} else {
			h.logger.ErrorContext(r.Context(), "failed to retrieve project",
				"error", err,
				"project_id", projectID,
			)
			writeError(w, http.StatusInternalServerError, "failed to retrieve project")
		}
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var projectID = pathParam(r, "id")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project ID is required")
		return
	}

	var project model.Project
	err := readJSON(r, &project)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	project.ID = projectID

	err = h.service.UpdateProject(r.Context(), &project)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
		} else if errors.Is(err, store.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid project data")
		} else {
			h.logger.ErrorContext(r.Context(), "failed to update project",
				"error", err,
				"project_id", projectID,
			)
			writeError(w, http.StatusInternalServerError, "failed to update project")
		}
		return
	}

	project, err = h.service.GetProject(r.Context(), projectID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to retrieve project",
			"error", err,
			"project_id", projectID,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var projectID = pathParam(r, "id")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project ID is required")
		return
	}

	err := h.service.DeleteProject(r.Context(), projectID)
	var projectHasTasksErr *service.ProjectHasTasksError
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		} else if errors.As(err, &projectHasTasksErr) {
			writeError(w, http.StatusConflict, err.Error())
			return
		} else {
			h.logger.ErrorContext(r.Context(), "failed to delete project",
				"error", err,
				"project_id", projectID,
			)
			writeError(w, http.StatusInternalServerError, "failed to delete project")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) handleList(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list projects",
			"error", err,
		)
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	if len(projects) == 0 {
		projects = []model.Project{} // Return an empty array instead of null
	}
	writeJSON(w, http.StatusOK, projects)
}

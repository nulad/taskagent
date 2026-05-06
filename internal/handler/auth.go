package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/nulad/taskagent/internal/middleware"
	"github.com/nulad/taskagent/internal/store"
)

// TODO log properly the error instead of just writing a generic message to the client

type AuthHandler struct {
	store *store.Store
}

func NewAuthHandler(store *store.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

func RegisterAuthRoutes(mux *http.ServeMux, handler *AuthHandler) {
	mux.HandleFunc("POST /auth/keys", handler.handleCreate)
	mux.HandleFunc("GET /auth/keys", handler.handleGet)
	mux.HandleFunc("DELETE /auth/keys/{id}", handler.handleDelete)
}

type createApiKeyRequest struct {
	Label    string `json:"label"`
	UserName string `json:"user_name"`
}

type createApiKeyResponse struct {
	Key   string `json:"key"`
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

func (h *AuthHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createApiKeyRequest
	err := readJSON(r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Label == "" || req.UserName == "" {
		writeError(w, http.StatusUnprocessableEntity, "label and user_name are required")
		return
	}
	user, err := h.store.GetUserByName(r.Context(), req.UserName)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	id, rawKey, err := h.store.CreateApiKey(r.Context(), req.Label, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	writeJSON(w, http.StatusCreated, createApiKeyResponse{Key: rawKey, ID: id, Label: req.Label})
}

type getKeyResponse struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
}

func (h *AuthHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	keys, err := h.store.ListApiKeys(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}
	var response []getKeyResponse
	if len(keys) == 0 {
		response = []getKeyResponse{} // ensure we return an empty array instead of null
	}
	for i := range keys {
		response = append(response, getKeyResponse{
			ID:        keys[i].ID,
			Label:     keys[i].Label,
			UserName:  keys[i].UserName,
			CreatedAt: keys[i].CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	var keyID = pathParam(r, "id")
	if keyID == "" {
		writeError(w, http.StatusBadRequest, "API key ID is required")
		return
	}

	keyIDInt, err := strconv.ParseInt(keyID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "API key ID is invalid")
		return
	}
	caller, ok := middleware.GetApiKey(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "cannot retrieve API key from context")
		return
	}

	if caller.ID == keyIDInt {
		writeError(w, http.StatusBadRequest, "cannot delete own API key")
		return
	}

	err = h.store.DeleteApiKey(r.Context(), keyIDInt)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "API key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

type ctxKey struct{}

var apiKeyCtxKey = ctxKey{}

func AuthMiddleware(s *store.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing API key")
				return
			}

			key, err := s.ValidateKey(r.Context(), apiKey)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					writeJSONError(w, http.StatusUnauthorized, "invalid API key")
				} else {
					writeJSONError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, apiKeyCtxKey, key)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func GetApiKey(ctx context.Context) (model.ApiKey, bool) {
	key, ok := ctx.Value(apiKeyCtxKey).(model.ApiKey)
	return key, ok
}

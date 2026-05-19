package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/nulad/taskagent/internal/logging"
)

func writeMiddlewareError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}

func RequestIDMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			randByte := make([]byte, 16)
			if _, err := rand.Read(randByte); err != nil {
				writeMiddlewareError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			requestID := hex.EncodeToString(randByte)

			w.Header().Set("X-Request-ID", requestID)
			ctx := logging.WithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/nulad/taskagent/internal/logging"
)

func RequestIDMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			randByte := make([]byte, 16)
			if _, err := rand.Read(randByte); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			requestID := hex.EncodeToString(randByte)

			w.Header().Set("X-Request-ID", requestID)
			ctx := logging.WithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

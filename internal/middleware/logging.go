package middleware

import (
	"log/slog"
	"net/http"
	"time"

	appLogging "github.com/nulad/taskagent/internal/logging"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default if WriteHeader is never called
	}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	// If the handler only calls Write, net/http will implicitly use 200.
	// We already defaulted to 200, so this just delegates.
	return r.ResponseWriter.Write(b)
}

func RequestLoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := newStatusRecorder(w)

			next.ServeHTTP(recorder, r)

			requestID, _ := appLogging.RequestIDFromContext(r.Context())

			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", recorder.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestID,
			)
		})
	}
}

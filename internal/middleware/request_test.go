package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nulad/taskagent/internal/logging"
	"github.com/nulad/taskagent/internal/store"
)

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		wantHeader    bool
		wantInContext bool
		wantLength    int
	}{
		{
			name:          "request ID generated and added to header and context",
			wantHeader:    true,
			wantInContext: true,
			wantLength:    32, // 16 bytes = 32 hex chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logs
			var logBuffer bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

			// Create middleware chain
			middleware := RequestIDMiddleware()

			// Create handler that checks context and logs
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if request ID is in context
				if reqID, ok := logging.RequestIDFromContext(r.Context()); ok {
					if !tt.wantInContext {
						t.Errorf("RequestIDFromContext() returned value, expected none")
					}
					if len(reqID) != tt.wantLength {
						t.Errorf("RequestID length = %d, want %d", len(reqID), tt.wantLength)
					}
				} else if tt.wantInContext {
					t.Error("RequestIDFromContext() returned false, expected true")
				}

				// Log with error to test request ID propagation
				logging.LogWithError(r.Context(), logger, "test error", errors.New("test error"), slog.String("test", "value"))
			})

			// Wrap handler with middleware
			wrapped := middleware(handler)

			// Create request
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			// Serve request
			wrapped.ServeHTTP(w, req)

			// Check response header
			header := w.Header().Get("X-Request-ID")
			if tt.wantHeader && header == "" {
				t.Error("Expected X-Request-ID header, got empty")
			} else if !tt.wantHeader && header != "" {
				t.Errorf("Expected no X-Request-ID header, got %s", header)
			}

			// Parse log output to check for request_id
			logLines := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
			for _, line := range logLines {
				if line == "" {
					continue
				}
				var logEntry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
					t.Fatalf("Failed to parse log line: %v", err)
				}

				if reqID, ok := logEntry["request_id"].(string); ok {
					if len(reqID) != tt.wantLength {
						t.Errorf("Logged request_id length = %d, want %d", len(reqID), tt.wantLength)
					}
				} else if tt.wantInContext {
					t.Error("Expected request_id in log output, but not found")
				}
			}
		})
	}
}

func TestRequestLoggingMiddleware(t *testing.T) {
	// Capture logs
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create middleware
	middleware := RequestLoggingMiddleware(logger)

	// Create handler that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Add request ID first
	reqIDMiddleware := RequestIDMiddleware()
	wrapped := reqIDMiddleware(middleware(handler))

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve request
	wrapped.ServeHTTP(w, req)

	// Check that request was logged with required fields
	logOutput := logBuffer.String()
	if logOutput == "" {
		t.Fatal("Expected log output, got empty")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(logOutput), nil); err != nil {
		// Try parsing each line if multiple lines
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
				break
			}
		}
		if logEntry == nil {
			t.Fatalf("Failed to parse log output: %v", err)
		}
	}

	// Check required fields
	requiredFields := []string{"method", "path", "status_code", "duration_ms", "request_id"}
	for _, field := range requiredFields {
		if _, ok := logEntry[field]; !ok {
			t.Errorf("Expected field %s in log output, but not found", field)
		}
	}

	// Check values
	if logEntry["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", logEntry["method"])
	}
	if logEntry["path"] != "/test" {
		t.Errorf("Expected path /test, got %v", logEntry["path"])
	}
	if logEntry["status_code"] != float64(500) {
		t.Errorf("Expected status_code 500, got %v", logEntry["status_code"])
	}
}

func TestErrorLoggingWithRequestID(t *testing.T) {
	// Create in-memory store with logger
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s, err := store.NewStore(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Add request ID to context
	reqID := "test1234567890abcdef1234567890abcdef"
	ctx := logging.WithRequestID(context.Background(), reqID)

	// Trigger an error that will be logged
	_, err = s.GetProject(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check log output for request_id
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, reqID) {
		t.Errorf("Expected request_id %s in log output, but not found. Log: %s", reqID, logOutput)
	}

	// Parse log to ensure request_id is a top-level field
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
			if logEntry["request_id"] != reqID {
				t.Errorf("Expected request_id %s, got %v", reqID, logEntry["request_id"])
			}
			break
		}
	}
}

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nulad/taskagent/internal/model"
	"github.com/nulad/taskagent/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	s, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	return s
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	AuthMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
	expectedBody := `{"error":"missing API key"}` + "\n"
	if rec.Body.String() != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestAuthMiddleware_InvalidKey(t *testing.T) {
	s := newTestStore(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	AuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
	var errorResponse ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&errorResponse)
	if err != nil {
		t.Fatalf("error decoding response body: %v", err)
	}
	if errorResponse.Error != "invalid API key" {
		t.Fatalf("expected error message %q, got %q", "invalid API key", errorResponse.Error)
	}
}

func TestAuthMiddleware_ValidKey(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()

	user, err := s.CreateUser(ctx, "test-user", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	rawKey, err := s.CreateApiKey(ctx, "test-key", user.ID)
	if err != nil {
		t.Fatalf("CreateApiKey() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", rawKey)
	var gotKey model.ApiKey
	AuthMiddleware(s)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, ok := GetApiKey(r.Context())
		if !ok {
			t.Fatal("expected API key in context")
		}
		gotKey = key
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if gotKey.UserID != user.ID {
		t.Fatalf("ValidateKey() UserID = %d, want %d", gotKey.UserID, user.ID)
	}
	if gotKey.UserName != user.Name {
		t.Fatalf("ValidateKey() UserName = %q, want %q", gotKey.UserName, user.Name)
	}
	if gotKey.CreatedAt == "" {
		t.Fatal("ValidateKey() expected CreatedAt to be set")
	}
}

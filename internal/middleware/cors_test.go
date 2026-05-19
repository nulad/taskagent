package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// No Origin header set

	nextCalled := false
	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called when no Origin header")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS header, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_AllowedOrigin_ActualRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	nextCalled := false
	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called for allowed origin")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "https://example.com", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("expected Vary %q, got %q", "Origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, PATCH, DELETE" {
		t.Fatalf("expected Access-Control-Allow-Methods %q, got %q", "GET, POST, PUT, PATCH, DELETE", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "X-API-Key, Content-Type" {
		t.Fatalf("expected Access-Control-Allow-Headers %q, got %q", "X-API-Key, Content-Type", got)
	}
}

func TestCORSMiddleware_DisallowedOrigin_ActualRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://malicious.com")

	nextCalled := false
	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		t.Fatal("next handler should not be called for disallowed origin")
	})).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("expected next handler NOT to be called for disallowed origin")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS header for disallowed origin, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
	var errResp ErrorResponse
	if err := jsonDecode(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("error decoding error response: %v", err)
	}
	if errResp.Error != "origin not allowed" {
		t.Fatalf("expected error %q, got %q", "origin not allowed", errResp.Error)
	}
}

func TestCORSMiddleware_AllowedOrigin_Preflight(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "https://example.com", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, PATCH, DELETE" {
		t.Fatalf("expected Access-Control-Allow-Methods %q, got %q", "GET, POST, PUT, PATCH, DELETE", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "X-API-Key, Content-Type" {
		t.Fatalf("expected Access-Control-Allow-Headers %q, got %q", "X-API-Key, Content-Type", got)
	}
}

func TestCORSMiddleware_DisallowedOrigin_Preflight(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://malicious.com")

	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for disallowed preflight")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS header for disallowed origin, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
	var errResp ErrorResponse
	if err := jsonDecode(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("error decoding error response: %v", err)
	}
	if errResp.Error != "origin not allowed" {
		t.Fatalf("expected error %q, got %q", "origin not allowed", errResp.Error)
	}
}

func TestCORSMiddleware_MultipleAllowedOrigins(t *testing.T) {
	origins := []string{"https://example.com", "https://app.example.com"}
	testCases := []struct {
		name          string
		origin        string
		expectAllowed bool
	}{
		{"exact match first", "https://example.com", true},
		{"exact match second", "https://app.example.com", true},
		{"not in list", "https://evil.com", false},
		{"similar but not match", "https://example.com.evil.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Origin", tc.origin)

			nextCalled := false
			CORSMiddleware(origins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rec, req)

			if tc.expectAllowed {
				if !nextCalled {
					t.Fatal("expected next handler to be called")
				}
				if rec.Code != http.StatusOK {
					t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
				}
				if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.origin {
					t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", tc.origin, got)
				}
			} else {
				if nextCalled {
					t.Fatal("expected next handler NOT to be called")
				}
				if rec.Code != http.StatusForbidden {
					t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
				}
			}
		})
	}
}

func TestCORSMiddleware_CORSHeadersOnDisallowedRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://malicious.com")

	CORSMiddleware([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)

	// Verify that no CORS headers leak to disallowed origins
	for _, header := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	} {
		if got := rec.Header().Get(header); got != "" {
			t.Errorf("unexpected CORS header %q: %q on disallowed origin response", header, got)
		}
	}
}

// jsonDecode is a helper to decode JSON from bytes without importing encoding/json in the test func signature
func jsonDecode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

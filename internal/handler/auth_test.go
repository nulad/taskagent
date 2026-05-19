package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/nulad/taskagent/internal/middleware"
	"github.com/nulad/taskagent/internal/store"
)

type authHandlerSeed struct {
	bootstrapRawKey string
	bootstrapKeyID  int64
	bootstrapUserID int64
	// Tests can stash extra IDs they care about here.
	targetKeyID int64
}

type authHandlerCase struct {
	name       string
	method     string
	path       func(baseURL string, seed *authHandlerSeed) string
	body       any
	setup      func(t *testing.T, s *store.Store, seed *authHandlerSeed)
	wantStatus int
	assert     func(t *testing.T, s *store.Store, seed *authHandlerSeed, status int, body []byte)
}

func newAuthHandlerTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()

	testLogger := testLogger()

	s := newHandlerTestStore(t)
	h := NewAuthHandler(s, testLogger)

	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, h)

	server := httptest.NewServer(middleware.AuthMiddleware(s)(mux))
	t.Cleanup(server.Close)

	return server, s
}

func seedUser(t *testing.T, s *store.Store, username string) int64 {
	t.Helper()

	newUser, err := s.CreateUser(context.Background(), username, false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	return newUser.ID
}

func seedApiKey(t *testing.T, s *store.Store, label string, userID int64) (int64, string) {
	t.Helper()

	id, raw, err := s.CreateApiKey(context.Background(), label, userID)
	if err != nil {
		t.Fatalf("CreateApiKey() error = %v", err)
	}
	return id, raw
}

func seedBootstrap(t *testing.T, s *store.Store) *authHandlerSeed {
	t.Helper()

	userID := seedUser(t, s, "bootstrap")
	keyID, raw := seedApiKey(t, s, "bootstrap", userID)
	return &authHandlerSeed{
		bootstrapRawKey: raw,
		bootstrapKeyID:  keyID,
		bootstrapUserID: userID,
	}
}

func decodeCreateApiKeyResponse(t *testing.T, body []byte) createApiKeyResponse {
	t.Helper()

	var response createApiKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return response
}

func decodeKeysResponse(t *testing.T, body []byte) []getKeyResponse {
	t.Helper()

	var keys []getKeyResponse
	if err := json.Unmarshal(body, &keys); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return keys
}

func doAuthedJSONRequest(t *testing.T, client *http.Client, method, url, apiKey string, body any) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}

	return resp, respBody
}

func TestAuthHandler(t *testing.T) {
	cases := []authHandlerCase{
		{
			name:   "create API key",
			method: http.MethodPost,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			body: createApiKeyRequest{
				Label:    "Test Key",
				UserName: "testuser",
			},
			setup: func(t *testing.T, s *store.Store, _ *authHandlerSeed) {
				seedUser(t, s, "testuser")
			},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, s *store.Store, _ *authHandlerSeed, status int, body []byte) {
				if status != http.StatusCreated {
					t.Fatalf("status = %d, want %d, body: %s", status, http.StatusCreated, string(body))
				}
				response := decodeCreateApiKeyResponse(t, body)
				if response.Key == "" || response.Label != "Test Key" {
					t.Fatalf("unexpected created API key: %+v", response)
				}

				got, err := s.ValidateKey(context.Background(), response.Key)
				if err != nil {
					t.Fatalf("ValidateKey() error = %v", err)
				}
				if got.ID != response.ID || got.Label != response.Label {
					t.Fatalf("unexpected persisted API key: %+v", got)
				}
			},
		},
		{
			name:   "create API key missing user_name",
			method: http.MethodPost,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			body:       createApiKeyRequest{Label: "no user"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "create API key missing label",
			method: http.MethodPost,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			body:       createApiKeyRequest{UserName: "ghost"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "create API key missing both fields",
			method: http.MethodPost,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			body:       createApiKeyRequest{},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, _ *store.Store, _ *authHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
				var errResp map[string]string
				if err := json.Unmarshal(body, &errResp); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				msg, ok := errResp["error"]
				if !ok {
					t.Fatalf("missing error field in response: %s", string(body))
				}
				if !strings.Contains(msg, "label") || !strings.Contains(msg, "user_name") {
					t.Fatalf("expected both field errors in message, got: %s", msg)
				}
			},
		},
		{
			name:   "create API key unknown user",
			method: http.MethodPost,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			body: createApiKeyRequest{
				Label:    "ghost",
				UserName: "nobody",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "list API keys",
			method: http.MethodGet,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys"
			},
			setup: func(t *testing.T, s *store.Store, seed *authHandlerSeed) {
				extraUser := seedUser(t, s, "extra")
				seedApiKey(t, s, "extra-key", extraUser)
			},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *store.Store, _ *authHandlerSeed, status int, body []byte) {
				if status != http.StatusOK {
					t.Fatalf("status = %d, want %d", status, http.StatusOK)
				}
				keys := decodeKeysResponse(t, body)
				if len(keys) != 2 {
					t.Fatalf("len(keys) = %d, want 2", len(keys))
				}
				for _, k := range keys {
					if k.Label == "" {
						t.Fatalf("incomplete metadata: %+v", k)
					}
				}
			},
		},
		{
			name:   "delete API key",
			method: http.MethodDelete,
			path: func(baseURL string, seed *authHandlerSeed) string {
				return baseURL + "/auth/keys/" + strconv.FormatInt(seed.targetKeyID, 10)
			},
			setup: func(t *testing.T, s *store.Store, seed *authHandlerSeed) {
				extraUser := seedUser(t, s, "extra")
				id, _ := seedApiKey(t, s, "doomed", extraUser)
				seed.targetKeyID = id
			},
			wantStatus: http.StatusNoContent,
			assert: func(t *testing.T, s *store.Store, seed *authHandlerSeed, status int, body []byte) {
				if status != http.StatusNoContent {
					t.Fatalf("status = %d, want %d", status, http.StatusNoContent)
				}
				if len(body) != 0 {
					t.Fatalf("body = %q, want empty", string(body))
				}
				if err := s.DeleteApiKey(context.Background(), seed.targetKeyID); err == nil {
					t.Fatal("expected key to already be deleted")
				}
			},
		},
		{
			name:   "delete own API key is rejected",
			method: http.MethodDelete,
			path: func(baseURL string, seed *authHandlerSeed) string {
				return baseURL + "/auth/keys/" + strconv.FormatInt(seed.bootstrapKeyID, 10)
			},
			wantStatus: http.StatusBadRequest,
			assert: func(t *testing.T, s *store.Store, seed *authHandlerSeed, status int, body []byte) {
				if status != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
				}
				if _, err := s.ValidateKey(context.Background(), seed.bootstrapRawKey); err != nil {
					t.Fatalf("bootstrap key should still be valid: %v", err)
				}
			},
		},
		{
			name:   "delete missing API key",
			method: http.MethodDelete,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys/999999"
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "delete API key with non-numeric id",
			method: http.MethodDelete,
			path: func(baseURL string, _ *authHandlerSeed) string {
				return baseURL + "/auth/keys/not-a-number"
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			server, s := newAuthHandlerTestServer(t)
			client := server.Client()

			seed := seedBootstrap(t, s)
			if tt.setup != nil {
				tt.setup(t, s, seed)
			}

			resp, body := doAuthedJSONRequest(t, client, tt.method, tt.path(server.URL, seed), seed.bootstrapRawKey, tt.body)
			if tt.assert != nil {
				tt.assert(t, s, seed, resp.StatusCode, body)
				return
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", resp.StatusCode, tt.wantStatus, string(body))
			}
		})
	}
}

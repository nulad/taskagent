package store

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/nulad/taskagent/internal/model"
)

func TestApiKeyStore_CreateAndValidate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	user, err := s.CreateUser(ctx, "apikey-user", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	rawKey, err := s.CreateApiKey(ctx, "default", user.ID)
	if err != nil {
		t.Fatalf("CreateApiKey() error = %v", err)
	}

	pattern := regexp.MustCompile(`^ta_[0-9a-f]{32}$`)
	if !pattern.MatchString(rawKey) {
		t.Fatalf("CreateApiKey() raw key %q does not match expected format", rawKey)
	}

	got, err := s.ValidateKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("ValidateKey() error = %v", err)
	}

	if got.ID <= 0 {
		t.Fatalf("ValidateKey() ID = %d, want > 0", got.ID)
	}
	if got.Label != "default" {
		t.Fatalf("ValidateKey() Label = %q, want %q", got.Label, "default")
	}
	if got.UserID != user.ID {
		t.Fatalf("ValidateKey() UserID = %d, want %d", got.UserID, user.ID)
	}
	if got.UserName != user.Name {
		t.Fatalf("ValidateKey() UserName = %q, want %q", got.UserName, user.Name)
	}
	if got.CreatedAt == "" {
		t.Fatal("ValidateKey() expected CreatedAt to be set")
	}
}

func TestApiKeyStore_ValidateKey_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		rawKey string
	}{
		{name: "empty key", rawKey: ""},
		{name: "missing prefix", rawKey: "not-prefixed"},
		{name: "wrong but prefixed", rawKey: "ta_0123456789abcdef0123456789abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.ValidateKey(ctx, tt.rawKey)
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("ValidateKey() error = %v, want %v", err, ErrNotFound)
			}
		})
	}
}

func TestApiKeyStore_DeleteThenValidateFails(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	user, err := s.CreateUser(ctx, "delete-key-user", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	rawKey, err := s.CreateApiKey(ctx, "temp", user.ID)
	if err != nil {
		t.Fatalf("CreateApiKey() error = %v", err)
	}

	validated, err := s.ValidateKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("ValidateKey() before delete error = %v", err)
	}

	if err := s.DeleteApiKey(ctx, validated.ID); err != nil {
		t.Fatalf("DeleteApiKey() error = %v", err)
	}

	_, err = s.ValidateKey(ctx, rawKey)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("ValidateKey() after delete error = %v, want %v", err, ErrNotFound)
	}
}

func TestApiKeyStore_ListApiKeys_MetadataOnly(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	userA, err := s.CreateUser(ctx, "user-a", false)
	if err != nil {
		t.Fatalf("CreateUser(user-a) error = %v", err)
	}
	userB, err := s.CreateUser(ctx, "user-b", true)
	if err != nil {
		t.Fatalf("CreateUser(user-b) error = %v", err)
	}

	_, err = s.CreateApiKey(ctx, "a-key", userA.ID)
	if err != nil {
		t.Fatalf("CreateApiKey(a-key) error = %v", err)
	}
	_, err = s.CreateApiKey(ctx, "b-key", userB.ID)
	if err != nil {
		t.Fatalf("CreateApiKey(b-key) error = %v", err)
	}

	got, err := s.ListApiKeys(ctx)
	if err != nil {
		t.Fatalf("ListApiKeys() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(ListApiKeys()) = %d, want 2", len(got))
	}

	for _, k := range got {
		if k.ID <= 0 || k.UserID <= 0 || k.UserName == "" || k.CreatedAt == "" {
			t.Fatalf("ListApiKeys() returned incomplete metadata: %+v", k)
		}
	}

	if _, ok := reflect.TypeOf(model.ApiKey{}).FieldByName("KeyHash"); ok {
		t.Fatal("ApiKey struct should not expose key hash")
	}
	if _, ok := reflect.TypeOf(model.ApiKey{}).FieldByName("RawKey"); ok {
		t.Fatal("ApiKey struct should not expose raw key")
	}
}

func TestApiKeyStore_DeleteApiKey_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      int64
		wantErr error
	}{
		{name: "invalid id", id: 0, wantErr: ErrInvalidInput},
		{name: "missing key", id: 999999, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.DeleteApiKey(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DeleteApiKey() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

package store

import (
	"context"
	"errors"
	"testing"
)

func TestUserStore_CreateAndGetByName(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created, err := s.CreateUser(ctx, "alice", true)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("CreateUser() ID = %d, want > 0", created.ID)
	}
	if created.Name != "alice" {
		t.Fatalf("CreateUser() Name = %q, want %q", created.Name, "alice")
	}
	if !created.Active {
		t.Fatal("CreateUser() Active = false, want true")
	}
	if !created.IsAdmin {
		t.Fatal("CreateUser() IsAdmin = false, want true")
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatal("CreateUser() expected timestamps to be set")
	}

	got, err := s.GetUserByName(ctx, "alice")
	if err != nil {
		t.Fatalf("GetUserByName() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("GetUserByName() ID = %d, want %d", got.ID, created.ID)
	}
	if got.Name != created.Name || got.Active != created.Active || got.IsAdmin != created.IsAdmin {
		t.Fatalf("GetUserByName() user mismatch: got %+v, created %+v", got, created)
	}
}

func TestUserStore_CreateUser_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		userName string
		wantErr error
	}{
		{name: "empty", userName: "", wantErr: ErrInvalidInput},
		{name: "whitespace-only", userName: "   ", wantErr: ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.CreateUser(ctx, tt.userName, false)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStore_GetUserByName_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		wantErr  error
	}{
		{name: "empty", userName: "", wantErr: ErrInvalidInput},
		{name: "missing", userName: "missing-user", wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.GetUserByName(ctx, tt.userName)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserByName() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}


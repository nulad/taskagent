package store

import (
	"context"
	"errors"
	"testing"

	"github.com/nulad/taskagent/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	return s
}

func TestProjectStore_CreateAndRetrieve(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateProject(ctx, &model.Project{
		Name:        "Test Project",
		Description: "A test description",
	})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id")
	}

	got, err := s.GetProject(ctx, id)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if got.ID != id || got.Name != "Test Project" || got.Description != "A test description" {
		t.Fatalf("unexpected project: %+v", got)
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Fatal("expected timestamps to be set")
	}
}

func TestProjectStore_CreateDuplicateName(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.CreateProject(ctx, &model.Project{Name: "Duplicate"})
	if err != nil {
		t.Fatalf("first CreateProject() error = %v", err)
	}

	_, err = s.CreateProject(ctx, &model.Project{Name: "Duplicate"})
	if err == nil {
		t.Fatal("expected duplicate name error, got nil")
	}
}

func TestProjectStore_GetProject_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "empty id",
			id:      "",
			wantErr: ErrInvalidInput,
		},
		{
			name:    "non-existent",
			id:      "missing-id",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.GetProject(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetProject() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectStore_CreateProject_Errors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		project *model.Project
		wantErr error
	}{
		{
			name:    "nil project",
			project: nil,
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty name",
			project: &model.Project{Name: ""},
			wantErr: ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.CreateProject(ctx, tt.project)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateProject() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectStore_UpdateProject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateProject(ctx, &model.Project{Name: "Original"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	tests := []struct {
		name    string
		project *model.Project
		wantErr error
	}{
		{
			name:    "update existing",
			project: &model.Project{ID: id, Name: "Updated", Description: "New desc"},
			wantErr: nil,
		},
		{
			name:    "nil project",
			project: nil,
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty id",
			project: &model.Project{ID: "", Name: "Name"},
			wantErr: ErrInvalidInput,
		},
		{
			name:    "empty name",
			project: &model.Project{ID: id, Name: ""},
			wantErr: ErrInvalidInput,
		},
		{
			name:    "update non-existent",
			project: &model.Project{ID: "missing-id", Name: "Name"},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateProject(ctx, tt.project)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("UpdateProject() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateProject() error = %v", err)
			}

			if tt.name == "update existing" {
				p, err := s.GetProject(ctx, id)
				if err != nil {
					t.Fatalf("GetProject() error = %v", err)
				}
				if p.Name != "Updated" {
					t.Errorf("Name = %v, want Updated", p.Name)
				}
			}
		})
	}
}

func TestProjectStore_DeleteProject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateProject(ctx, &model.Project{Name: "To Delete"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "delete existing",
			id:      id,
			wantErr: nil,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: ErrInvalidInput,
		},
		{
			name:    "delete non-existent",
			id:      "missing-id",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.DeleteProject(ctx, tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DeleteProject() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DeleteProject() error = %v", err)
			}

			if tt.name == "delete existing" {
				_, err := s.GetProject(ctx, id)
				if !errors.Is(err, ErrNotFound) {
					t.Error("expected ErrNotFound after deletion")
				}
			}
		})
	}
}

func TestProjectStore_ListProjects(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Store, context.Context) (string, error)
		wantLen int
	}{
		{
			name:    "list empty",
			setup:   nil,
			wantLen: 0,
		},
		{
			name: "list multiple",
			setup: func(s *Store, ctx context.Context) (string, error) {
				for _, name := range []string{"Project1", "Project2", "Project3"} {
					if _, err := s.CreateProject(ctx, &model.Project{Name: name}); err != nil {
						return "", err
					}
				}
				return "", nil
			},
			wantLen: 3,
		},
		{
			name: "list multiple then delete one",
			setup: func(s *Store, ctx context.Context) (string, error) {
				for _, name := range []string{"ListProj1", "ListProj2"} {
					if _, err := s.CreateProject(ctx, &model.Project{Name: name}); err != nil {
						return "", err
					}
				}
				projects, err := s.ListProjects(ctx)
				if err != nil {
					return "", err
				}
				if len(projects) > 0 {
					return projects[0].ID, s.DeleteProject(ctx, projects[0].ID)
				}
				return "", nil
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			ctx := context.Background()

			if tt.setup != nil {
				deletedID, err := tt.setup(s, ctx)
				if err != nil {
					t.Fatalf("setup() error = %v", err)
				}
				if deletedID != "" && tt.name == "list multiple then delete one" {
					t.Logf("deleted project %s", deletedID)
				}
			}

			projects, err := s.ListProjects(ctx)
			if err != nil {
				t.Fatalf("ListProjects() error = %v", err)
			}

			if len(projects) != tt.wantLen {
				t.Errorf("len(projects) = %d, want %d", len(projects), tt.wantLen)
			}

			if tt.wantLen > 0 {
				names := make(map[string]bool)
				for _, p := range projects {
					names[p.Name] = true
				}
				if tt.name == "list multiple" {
					if !names["Project1"] || !names["Project2"] || !names["Project3"] {
						t.Error("expected all project names")
					}
				}
				if tt.name == "list multiple then delete one" {
					if names["ListProj1"] || !names["ListProj2"] {
						t.Error("expected ListProj2 only")
					}
				}
			}
		})
	}
}

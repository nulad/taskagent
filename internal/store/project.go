package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (s *Store) CreateProject(ctx context.Context, project *Project) (string, error) {
	if project == nil || project.Name == "" {
		return "", ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)

	id := uuid.New().String()
	query := `
		INSERT INTO projects (uuid, name, description, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
		`

	_, err := s.DB.ExecContext(ctx, query, id, project.Name, project.Description, currTimestamp, currTimestamp)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (Project, error) {
	if id == "" {
		return Project{}, ErrInvalidInput
	}

	query := `
		SELECT uuid, name, description, created_at, updated_at 
		FROM projects 
		WHERE uuid = ?
		`

	var project Project
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, err
	}
	return project, nil
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	query := `
		SELECT uuid, name, description, created_at, updated_at 
		FROM projects 
		`

	var projects []Project
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var project Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *Store) UpdateProject(ctx context.Context, project *Project) error {

	if project == nil || project.ID == "" || project.Name == "" {
		return ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE projects 
		SET name = ?, description = ?, updated_at = ? 
		WHERE uuid = ?
		`

	result, err := s.DB.ExecContext(ctx, query, project.Name, project.Description, currTimestamp, project.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	query := `
		DELETE FROM projects 
		WHERE uuid = ?
		`

	result, err := s.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

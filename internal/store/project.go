package store

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/nulad/taskagent/internal/logging"
	"github.com/nulad/taskagent/internal/model"
)

func (s *Store) CreateProject(ctx context.Context, project *model.Project) (string, error) {
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
		logging.LogWithError(ctx, s.logger, "failed to create project", err, slog.String("name", project.Name))
		return "", err
	}
	return id, nil
}

func (s *Store) GetProject(ctx context.Context, id string) (model.Project, error) {
	if id == "" {
		return model.Project{}, ErrInvalidInput
	}

	query := `
		SELECT uuid, name, description, created_at, updated_at 
		FROM projects 
		WHERE uuid = ?
		`

	var project model.Project
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logging.LogWithError(ctx, s.logger, "project not found", err, slog.String("id", id))
			return model.Project{}, ErrNotFound
		}
		logging.LogWithError(ctx, s.logger, "failed to get project", err, slog.String("id", id))
		return model.Project{}, err
	}
	return project, nil
}

func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	query := `
		SELECT uuid, name, description, created_at, updated_at 
		FROM projects 
		`

	var projects []model.Project
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to list projects", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var project model.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			logging.LogWithError(ctx, s.logger, "failed to scan project", err)
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		logging.LogWithError(ctx, s.logger, "failed to iterate projects", err)
		return nil, err
	}

	return projects, nil
}

func (s *Store) UpdateProject(ctx context.Context, project *model.Project) error {

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
		logging.LogWithError(ctx, s.logger, "failed to update project", err, slog.String("id", project.ID))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get rows affected", err)
		return err
	}
	if rowsAffected == 0 {
		logging.LogWithError(ctx, s.logger, "project not found", ErrNotFound, slog.String("id", project.ID))
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
		logging.LogWithError(ctx, s.logger, "failed to delete project", err, slog.String("id", id))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get rows affected", err)
		return err
	}
	if rowsAffected == 0 {
		logging.LogWithError(ctx, s.logger, "project not found", ErrNotFound, slog.String("id", id))
		return ErrNotFound
	}
	return nil
}

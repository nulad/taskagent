package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nulad/taskagent/internal/model"
)

func (s *Store) statusExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM statuses WHERE name = ?)", name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) CreateTask(ctx context.Context, task *model.Task) error {
	if task == nil || task.ProjectID == "" || task.Title == "" {
		return ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)
	task.ID = uuid.New().String()
	task.CreatedAt = currTimestamp
	task.UpdatedAt = currTimestamp
	if task.Status == "" {
		task.Status = "backlog"
	}
	ok, err := s.statusExists(ctx, task.Status.String())
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidInput
	}

	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO tasks (uuid, project_uuid, name, description, status_id, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, 
		(SELECT id FROM statuses WHERE name = ?), 
		?, ?, ?)
		`

	_, err = s.DB.ExecContext(
		ctx,
		query,
		task.ID,
		task.ProjectID,
		task.Title,
		task.Description,
		task.Status,
		string(tagsJSON),
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetTask(ctx context.Context, id string) (model.Task, error) {
	if id == "" {
		return model.Task{}, ErrInvalidInput
	}

	var tagsJSON string

	query := `
		SELECT
			t.uuid,
			t.project_uuid,
			t.name,
			COALESCE(t.description, ''),
			s.name,
			COALESCE(t.tags, '[]'),
			t.created_at,
			COALESCE(t.updated_at, '')
		FROM tasks t	
		JOIN statuses s ON s.id = t.status_id
		WHERE t.uuid = ?
		`

	var task model.Task
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&tagsJSON,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Task{}, ErrNotFound
		}
		return model.Task{}, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
		return model.Task{}, err
	}

	return task, nil
}

func (s *Store) ListTasks(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	tasks := []model.Task{}

	query := `
		SELECT
			t.uuid,
			t.project_uuid,
			t.name,
			COALESCE(t.description, ''),
			s.name,
			COALESCE(t.tags, '[]'),
			t.created_at,
			COALESCE(t.updated_at, '')
		FROM tasks t	
		JOIN statuses s ON s.id = t.status_id
		`

	var conditions []string
	var args []any

	if filter.ProjectID != nil {
		conditions = append(conditions, "t.project_uuid = ?")
		args = append(args, *filter.ProjectID)
	}

	if filter.Status != nil {
		conditions = append(conditions, "s.name = ?")
		args = append(args, *filter.Status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagsJSON string
		var task model.Task

		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.Title,
			&task.Description,
			&task.Status,
			&tagsJSON,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Store) UpdateTask(ctx context.Context, task *model.Task) error {
	if task == nil || task.ID == "" || task.Title == "" || task.ProjectID == "" || task.Status == "" {
		return ErrInvalidInput
	}

	ok, err := s.statusExists(ctx, task.Status.String())
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)
	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return err
	}

	query := `
		UPDATE tasks
		SET
			project_uuid = ?,
			name = ?,
			description = ?,
			status_id = (SELECT id FROM statuses WHERE name = ?),
			tags = ?,
			updated_at = ?
		WHERE uuid = ?
		`

	result, err := s.DB.ExecContext(
		ctx,
		query,
		task.ProjectID,
		task.Title,
		task.Description,
		task.Status,
		string(tagsJSON),
		currTimestamp,
		task.ID,
	)
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

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	query := `
		DELETE FROM tasks
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

func (s *Store) UpdateTaskStatus(ctx context.Context, id string, status string) error {
	if id == "" || status == "" {
		return ErrInvalidInput
	}

	ok, err := s.statusExists(ctx, status)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)

	query := `
		UPDATE tasks
		SET
			status_id = (SELECT id FROM statuses WHERE name = ?),
			updated_at = ?
		WHERE uuid = ?
		`

	result, err := s.DB.ExecContext(
		ctx,
		query,
		status,
		currTimestamp,
		id,
	)
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

package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/nulad/taskagent/internal/model"
)

func (s *Store) CreateUser(ctx context.Context, name string, isAdmin bool) (model.User, error) {
	if strings.TrimSpace(name) == "" {
		return model.User{}, ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO users (name, is_admin, created_at, updated_at)
		VALUES (?,?,?,?)
	`

	res, err := s.DB.ExecContext(ctx, query, name, isAdmin, currTimestamp, currTimestamp)
	if err != nil {
		return model.User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	return model.User{
		ID:        id,
		Name:      name,
		Active:    true,
		IsAdmin:   isAdmin,
		CreatedAt: currTimestamp,
		UpdatedAt: currTimestamp,
	}, nil
}

func (s *Store) GetUserByName(ctx context.Context, name string) (model.User, error) {
	if name == "" {
		return model.User{}, ErrInvalidInput
	}

	query := `
		SELECT id, name, active, is_admin, created_at, COALESCE(updated_at, '')
		FROM users
		WHERE name = ?
	`

	var user model.User
	err := s.DB.QueryRowContext(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.Active,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

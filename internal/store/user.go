package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) CreateUser(ctx context.Context, name string, isAdmin bool) (User, error) {
	if strings.TrimSpace(name) == "" {
		return User{}, ErrInvalidInput
	}

	currTimestamp := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO users (name, is_admin, created_at, updated_at)
		VALUES (?,?,?,?)
	`

	res, err := s.DB.ExecContext(ctx, query, name, isAdmin, currTimestamp, currTimestamp)
	if err != nil {
		return User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return User{
		ID:        id,
		Name:      name,
		Active:    true,
		IsAdmin:   isAdmin,
		CreatedAt: currTimestamp,
		UpdatedAt: currTimestamp,
	}, nil
}

func (s *Store) GetUserByName(ctx context.Context, name string) (User, error) {
	if name == "" {
		return User{}, ErrInvalidInput
	}

	query := `
		SELECT id, name, active, is_admin, created_at, COALESCE(updated_at, '')
		FROM users
		WHERE name = ?
	`

	var user User
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
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	return user, nil
}

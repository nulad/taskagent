package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type ApiKey struct {
	ID        int64  `json:"id"`
	Label     string `json:"label"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateApiKey(ctx context.Context, label string, userID int64) (string, error) {
	if userID <= 0 {
		return "", ErrInvalidInput
	}

	rawKey, err := generateRawKey()
	if err != nil {
		return "", err
	}

	hashedKey := hashKey(rawKey)
	currTimestamp := time.Now().UTC().Format(time.RFC3339)

	query := `
		INSERT INTO api_keys (user_id, key_hash, label, created_at)
		VALUES (?, ?, ?, ?)
	`
	_, err = s.DB.ExecContext(ctx, query, userID, hashedKey, label, currTimestamp)
	if err != nil {
		return "", err
	}
	return rawKey, nil
}

func (s *Store) ValidateKey(ctx context.Context, rawKey string) (ApiKey, error) {
	if rawKey == "" || !strings.HasPrefix(rawKey, "ta_") {
		return ApiKey{}, ErrNotFound
	}

	hashed := hashKey(rawKey)
	query := `
		SELECT k.id, COALESCE(k.label, ''), k.user_id, u.name, k.created_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.key_hash = ?
	`
	var apiKey ApiKey
	err := s.DB.QueryRowContext(ctx, query, hashed).Scan(&apiKey.ID, &apiKey.Label, &apiKey.UserID, &apiKey.UserName, &apiKey.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApiKey{}, ErrNotFound
		}
		return ApiKey{}, err
	}
	return apiKey, nil
}

func (s *Store) ListApiKeys(ctx context.Context) ([]ApiKey, error) {
	query := `
		SELECT k.id, COALESCE(k.label, ''), k.user_id, u.name, k.created_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		ORDER BY k.created_at DESC
	`
	var apiKeys []ApiKey
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var apiKey ApiKey
		err := rows.Scan(&apiKey.ID, &apiKey.Label, &apiKey.UserID, &apiKey.UserName, &apiKey.CreatedAt)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Store) DeleteApiKey(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	query := `
		DELETE FROM api_keys
		WHERE id = ?
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

func generateRawKey() (string, error) {
	b := make([]byte, 16) // 16 bytes => 32 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "ta_" + hex.EncodeToString(b), nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

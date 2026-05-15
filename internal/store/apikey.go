package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/nulad/taskagent/internal/logging"
	"github.com/nulad/taskagent/internal/model"
)

func (s *Store) CreateApiKey(ctx context.Context, label string, userID int64) (int64, string, error) {
	if userID <= 0 {
		logging.LogWithError(ctx, s.logger, "failed to create API key", ErrInvalidInput, slog.Int64("user_id", userID))
		return 0, "", ErrInvalidInput
	}

	rawKey, err := generateRawKey()
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to generate raw key", err)
		return 0, "", err
	}

	hashedKey := hashKey(rawKey)
	currTimestamp := time.Now().UTC().Format(time.RFC3339)

	query := `
		INSERT INTO api_keys (user_id, key_hash, label, created_at)
		VALUES (?, ?, ?, ?)
	`
	result, err := s.DB.ExecContext(ctx, query, userID, hashedKey, label, currTimestamp)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to insert API key", err)
		return 0, "", err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get last insert ID", err)
		return 0, "", err
	}

	return id, rawKey, nil
}

func (s *Store) ValidateKey(ctx context.Context, rawKey string) (model.ApiKey, error) {
	if rawKey == "" || !strings.HasPrefix(rawKey, "ta_") {
		return model.ApiKey{}, ErrNotFound
	}

	hashed := hashKey(rawKey)
	query := `
		SELECT k.id, COALESCE(k.label, ''), k.user_id, u.name, k.created_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.key_hash = ?
	`
	var apiKey model.ApiKey
	err := s.DB.QueryRowContext(ctx, query, hashed).Scan(&apiKey.ID, &apiKey.Label, &apiKey.UserID, &apiKey.UserName, &apiKey.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logging.LogWithError(ctx, s.logger, "API key not found", err)
			return model.ApiKey{}, ErrNotFound
		}
		logging.LogWithError(ctx, s.logger, "failed to query API key", err)
		return model.ApiKey{}, err
	}
	return apiKey, nil
}

func (s *Store) ListApiKeys(ctx context.Context) ([]model.ApiKey, error) {
	query := `
		SELECT k.id, COALESCE(k.label, ''), k.user_id, u.name, k.created_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		ORDER BY k.created_at DESC
	`
	var apiKeys []model.ApiKey
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to list API keys", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var apiKey model.ApiKey
		err := rows.Scan(&apiKey.ID, &apiKey.Label, &apiKey.UserID, &apiKey.UserName, &apiKey.CreatedAt)
		if err != nil {
			logging.LogWithError(ctx, s.logger, "failed to scan API key", err)
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}
	if err := rows.Err(); err != nil {
		logging.LogWithError(ctx, s.logger, "failed to iterate API keys", err)
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
		logging.LogWithError(ctx, s.logger, "failed to delete API key", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.LogWithError(ctx, s.logger, "failed to get rows affected", err)
		return err
	}
	if rowsAffected == 0 {
		logging.LogWithError(ctx, s.logger, "failed to delete API key", ErrNotFound, slog.Int64("api_key_id", id))
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

package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type IdempotencyStore struct {
	db *sql.DB
}

func NewIdempotencyStore(db *sql.DB) *IdempotencyStore {
	return &IdempotencyStore{db: db}
}

func (s *IdempotencyStore) EnsureTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS idempotency_keys (
		id SERIAL PRIMARY KEY,
		client_generated_id UUID NOT NULL UNIQUE,
		event_type VARCHAR(100) NOT NULL,
		user_id UUID NOT NULL,
		processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_idempotency_keys_client_id ON idempotency_keys(client_generated_id);
	CREATE INDEX IF NOT EXISTS idx_idempotency_keys_event_type ON idempotency_keys(event_type);
	CREATE INDEX IF NOT EXISTS idx_idempotency_keys_created_at ON idempotency_keys(created_at);
	`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create idempotency_keys table: %w", err)
	}

	return nil
}

func (s *IdempotencyStore) CheckAndMarkProcessed(ctx context.Context, clientGeneratedID, eventType, userID string) (bool, error) {
	query := `
	INSERT INTO idempotency_keys (client_generated_id, event_type, user_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (client_generated_id) DO NOTHING
	RETURNING id
	`

	var id int
	err := s.db.QueryRowContext(ctx, query, clientGeneratedID, eventType, userID).Scan(&id)

	if err == sql.ErrNoRows {
		log.Warn().
			Str("client_generated_id", clientGeneratedID).
			Str("event_type", eventType).
			Msg("Event already processed (duplicate)")
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check idempotency: %w", err)
	}

	log.Info().
		Str("client_generated_id", clientGeneratedID).
		Str("event_type", eventType).
		Msg("Event marked as processed")

	return true, nil
}

func (s *IdempotencyStore) IsProcessed(ctx context.Context, clientGeneratedID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM idempotency_keys WHERE client_generated_id = $1)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, clientGeneratedID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if processed: %w", err)
	}

	return exists, nil
}

func (s *IdempotencyStore) CleanupOldKeys(ctx context.Context, olderThanDays int) (int64, error) {
	query := `DELETE FROM idempotency_keys WHERE created_at < $1`

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	result, err := s.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old keys: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	log.Info().
		Int64("rows_deleted", rowsAffected).
		Int("older_than_days", olderThanDays).
		Msg("Cleaned up old idempotency keys")

	return rowsAffected, nil
}

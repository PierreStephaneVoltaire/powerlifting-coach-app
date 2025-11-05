package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type EventHandler func(ctx context.Context, payload []byte) error

type BaseEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
}

type IdempotencyMiddleware struct {
	db *sql.DB
}

func NewIdempotencyMiddleware(db *sql.DB) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{db: db}
}

func (m *IdempotencyMiddleware) EnsureIdempotencyTable(ctx context.Context) error {
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

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create idempotency_keys table: %w", err)
	}

	return nil
}

func (m *IdempotencyMiddleware) WrapHandler(handler EventHandler) EventHandler {
	return func(ctx context.Context, payload []byte) error {
		var baseEvent BaseEvent
		if err := json.Unmarshal(payload, &baseEvent); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal base event for idempotency check")
			return fmt.Errorf("invalid event format: %w", err)
		}

		if baseEvent.ClientGeneratedID == "" {
			log.Error().Str("event_type", baseEvent.EventType).Msg("Event missing client_generated_id")
			return fmt.Errorf("event missing client_generated_id")
		}

		isNew, err := m.checkAndMarkProcessed(ctx, baseEvent)
		if err != nil {
			return fmt.Errorf("idempotency check failed: %w", err)
		}

		if !isNew {
			log.Info().
				Str("client_generated_id", baseEvent.ClientGeneratedID).
				Str("event_type", baseEvent.EventType).
				Msg("Skipping duplicate event")
			return nil
		}

		return handler(ctx, payload)
	}
}

func (m *IdempotencyMiddleware) checkAndMarkProcessed(ctx context.Context, event BaseEvent) (bool, error) {
	query := `
	INSERT INTO idempotency_keys (client_generated_id, event_type, user_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (client_generated_id) DO NOTHING
	RETURNING id
	`

	var id int
	err := m.db.QueryRowContext(ctx, query, event.ClientGeneratedID, event.EventType, event.UserID).Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check idempotency: %w", err)
	}

	log.Info().
		Str("client_generated_id", event.ClientGeneratedID).
		Str("event_type", event.EventType).
		Msg("New event marked as processed")

	return true, nil
}

func (m *IdempotencyMiddleware) WithTransaction(handler EventHandler) EventHandler {
	return func(ctx context.Context, payload []byte) error {
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				log.Error().Interface("panic", p).Msg("Panic in event handler, rolling back transaction")
				panic(p)
			}
		}()

		err = handler(ctx, payload)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error().Err(rbErr).Msg("Failed to rollback transaction")
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Info().Msg("Event processed and committed successfully")
		return nil
	}
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MachineEventHandlers struct {
	db *sql.DB
}

func NewMachineEventHandlers(db *sql.DB) *MachineEventHandlers {
	return &MachineEventHandlers{db: db}
}

type MachineNotesSubmittedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		NoteID      string `json:"note_id"`
		Brand       string `json:"brand"`
		Model       string `json:"model"`
		MachineType string `json:"machine_type"`
		Settings    string `json:"settings"`
		Visibility  string `json:"visibility"`
	} `json:"data"`
}

func (h *MachineEventHandlers) HandleMachineNotesSubmitted(ctx context.Context, payload []byte) error {
	var event MachineNotesSubmittedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	noteID, err := uuid.Parse(event.Data.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note_id: %w", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkQuery := `SELECT COUNT(*) FROM idempotency_keys WHERE key = $1`
	var count int
	err = tx.QueryRowContext(ctx, checkQuery, event.ClientGeneratedID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if count > 0 {
		log.Info().Str("client_generated_id", event.ClientGeneratedID).Msg("Event already processed")
		return nil
	}

	query := `
	INSERT INTO machine_notes (note_id, user_id, brand, model, machine_type, settings, visibility)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, query, noteID, userID, event.Data.Brand, event.Data.Model, event.Data.MachineType, event.Data.Settings, event.Data.Visibility)
	if err != nil {
		return fmt.Errorf("failed to insert machine note: %w", err)
	}

	idempotencyQuery := `INSERT INTO idempotency_keys (key, event_type, processed_at) VALUES ($1, $2, NOW())`
	_, err = tx.ExecContext(ctx, idempotencyQuery, event.ClientGeneratedID, event.EventType)
	if err != nil {
		return fmt.Errorf("failed to insert idempotency key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("note_id", noteID.String()).
		Str("user_id", event.UserID).
		Str("machine_type", event.Data.MachineType).
		Msg("Machine note persisted")

	return nil
}

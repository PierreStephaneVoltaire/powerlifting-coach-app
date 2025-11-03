package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type EventPublisher interface {
	PublishEvent(routingKey string, event interface{}) error
}

type DMEventHandlers struct {
	db        *sql.DB
	publisher EventPublisher
}

func NewDMEventHandlers(db *sql.DB, publisher EventPublisher) *DMEventHandlers {
	return &DMEventHandlers{
		db:        db,
		publisher: publisher,
	}
}

type DMMessageSentEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		ConversationID string        `json:"conversation_id"`
		SenderID       string        `json:"sender_id"`
		RecipientID    string        `json:"recipient_id"`
		MessageBody    string        `json:"message_body"`
		Attachments    []interface{} `json:"attachments"`
	} `json:"data"`
}

func (h *DMEventHandlers) HandleDMMessageSent(ctx context.Context, payload []byte) error {
	var event DMMessageSentEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	senderID, err := uuid.Parse(event.Data.SenderID)
	if err != nil {
		return fmt.Errorf("invalid sender_id: %w", err)
	}

	recipientID, err := uuid.Parse(event.Data.RecipientID)
	if err != nil {
		return fmt.Errorf("invalid recipient_id: %w", err)
	}

	messageID := uuid.New()

	attachmentsJSON, err := json.Marshal(event.Data.Attachments)
	if err != nil {
		return fmt.Errorf("failed to marshal attachments: %w", err)
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

	var conversationID uuid.UUID
	if event.Data.ConversationID != "" {
		conversationID, err = uuid.Parse(event.Data.ConversationID)
		if err != nil {
			return fmt.Errorf("invalid conversation_id: %w", err)
		}
	} else {
		participant1, participant2 := senderID, recipientID
		if senderID.String() > recipientID.String() {
			participant1, participant2 = recipientID, senderID
		}

		conversationQuery := `
		INSERT INTO conversations (participant_1_id, participant_2_id, last_message_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (participant_1_id, participant_2_id)
		DO UPDATE SET last_message_at = NOW()
		RETURNING conversation_id
		`
		err = tx.QueryRowContext(ctx, conversationQuery, participant1, participant2).Scan(&conversationID)
		if err != nil {
			return fmt.Errorf("failed to create/update conversation: %w", err)
		}
	}

	messageQuery := `
	INSERT INTO messages (message_id, conversation_id, sender_id, recipient_id, message_body, attachments)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, messageQuery, messageID, conversationID, senderID, recipientID, event.Data.MessageBody, attachmentsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
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
		Str("message_id", messageID.String()).
		Str("conversation_id", conversationID.String()).
		Str("sender_id", event.Data.SenderID).
		Msg("DM message persisted")

	if h.publisher != nil {
		persistedEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "dm.message.persisted",
			"client_generated_id": event.ClientGeneratedID,
			"user_id":             event.UserID,
			"timestamp":           event.Timestamp,
			"source_service":      "dm-service",
			"data": map[string]interface{}{
				"message_id":      messageID.String(),
				"conversation_id": conversationID.String(),
				"sender_id":       event.Data.SenderID,
				"recipient_id":    event.Data.RecipientID,
				"message_body":    event.Data.MessageBody,
			},
		}
		if err := h.publisher.PublishEvent("dm.message.persisted", persistedEvent); err != nil {
			log.Error().Err(err).Msg("Failed to publish dm.message.persisted event")
		}
	}

	return nil
}

type DMPinAttemptsEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		ConversationID string `json:"conversation_id"`
	} `json:"data"`
}

func (h *DMEventHandlers) HandleDMPinAttempts(ctx context.Context, payload []byte) error {
	var event DMPinAttemptsEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	conversationID, err := uuid.Parse(event.Data.ConversationID)
	if err != nil {
		return fmt.Errorf("invalid conversation_id: %w", err)
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

	pinnedData := map[string]string{
		"pinned_at": event.Timestamp,
		"pinned_by": event.UserID,
	}
	pinnedJSON, err := json.Marshal([]interface{}{pinnedData})
	if err != nil {
		return fmt.Errorf("failed to marshal pinned data: %w", err)
	}

	query := `
	UPDATE conversations
	SET pinned_attempts = pinned_attempts || $1::jsonb,
	    updated_at = NOW()
	WHERE conversation_id = $2
	`

	_, err = tx.ExecContext(ctx, query, pinnedJSON, conversationID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
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
		Str("conversation_id", conversationID.String()).
		Str("user_id", event.UserID).
		Msg("Attempts pinned to conversation")

	return nil
}

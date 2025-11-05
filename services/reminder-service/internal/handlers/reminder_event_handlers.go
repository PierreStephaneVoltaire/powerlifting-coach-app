package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type EventPublisher interface {
	PublishEvent(routingKey string, event interface{}) error
}

type ReminderEventHandlers struct {
	db        *sql.DB
	publisher EventPublisher
}

func NewReminderEventHandlers(db *sql.DB, publisher EventPublisher) *ReminderEventHandlers {
	return &ReminderEventHandlers{
		db:        db,
		publisher: publisher,
	}
}

type ProgramPlanPersistedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		ProgramID           string `json:"program_id"`
		Name                string `json:"name"`
		StartDate           string `json:"start_date"`
		CompDate            string `json:"comp_date"`
		TrainingDaysPerWeek int    `json:"training_days_per_week"`
	} `json:"data"`
}

func (h *ReminderEventHandlers) HandleProgramPlanPersisted(ctx context.Context, payload []byte) error {
	var event ProgramPlanPersistedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	programID, err := uuid.Parse(event.Data.ProgramID)
	if err != nil {
		return fmt.Errorf("invalid program_id: %w", err)
	}

	compDate, err := time.Parse("2006-01-02", event.Data.CompDate)
	if err != nil {
		return fmt.Errorf("invalid comp_date: %w", err)
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

	reminders := []struct {
		reminderType string
		daysBeforeComp int
		message      string
	}{
		{"comp_week", 7, "Your competition is in 1 week! Time to taper and prepare."},
		{"comp_2weeks", 14, "2 weeks until competition. Focus on recovery and technique."},
		{"comp_4weeks", 28, "4 weeks out from competition. Peak strength phase."},
	}

	for _, reminder := range reminders {
		scheduledFor := compDate.AddDate(0, 0, -reminder.daysBeforeComp)
		if scheduledFor.Before(time.Now()) {
			continue
		}

		reminderID := uuid.New()
		metadata := map[string]interface{}{
			"program_name":          event.Data.Name,
			"comp_date":             event.Data.CompDate,
			"training_days_per_week": event.Data.TrainingDaysPerWeek,
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		query := `
		INSERT INTO reminders (reminder_id, user_id, program_id, reminder_type, scheduled_for, message, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		`

		_, err = tx.ExecContext(ctx, query, reminderID, userID, programID, reminder.reminderType, scheduledFor, reminder.message, metadataJSON)
		if err != nil {
			return fmt.Errorf("failed to insert reminder: %w", err)
		}

		log.Info().
			Str("reminder_id", reminderID.String()).
			Str("reminder_type", reminder.reminderType).
			Time("scheduled_for", scheduledFor).
			Msg("Reminder created")
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
		Str("program_id", programID.String()).
		Str("user_id", event.UserID).
		Msg("Program reminders scheduled")

	return nil
}

func (h *ReminderEventHandlers) HandleProgramPlanUpdated(ctx context.Context, payload []byte) error {
	var event ProgramPlanPersistedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	programID, err := uuid.Parse(event.Data.ProgramID)
	if err != nil {
		return fmt.Errorf("invalid program_id: %w", err)
	}

	compDate, err := time.Parse("2006-01-02", event.Data.CompDate)
	if err != nil {
		return fmt.Errorf("invalid comp_date: %w", err)
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

	cancelQuery := `
	UPDATE reminders
	SET status = 'cancelled', updated_at = NOW()
	WHERE program_id = $1 AND user_id = $2 AND status = 'pending'
	`
	_, err = tx.ExecContext(ctx, cancelQuery, programID, userID)
	if err != nil {
		return fmt.Errorf("failed to cancel existing reminders: %w", err)
	}

	reminders := []struct {
		reminderType   string
		daysBeforeComp int
		message        string
	}{
		{"comp_week", 7, "Your competition is in 1 week! Time to taper and prepare."},
		{"comp_2weeks", 14, "2 weeks until competition. Focus on recovery and technique."},
		{"comp_4weeks", 28, "4 weeks out from competition. Peak strength phase."},
	}

	for _, reminder := range reminders {
		scheduledFor := compDate.AddDate(0, 0, -reminder.daysBeforeComp)
		if scheduledFor.Before(time.Now()) {
			continue
		}

		reminderID := uuid.New()
		metadata := map[string]interface{}{
			"program_name":           event.Data.Name,
			"comp_date":              event.Data.CompDate,
			"training_days_per_week": event.Data.TrainingDaysPerWeek,
		}
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		query := `
		INSERT INTO reminders (reminder_id, user_id, program_id, reminder_type, scheduled_for, message, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		`

		_, err = tx.ExecContext(ctx, query, reminderID, userID, programID, reminder.reminderType, scheduledFor, reminder.message, metadataJSON)
		if err != nil {
			return fmt.Errorf("failed to insert reminder: %w", err)
		}

		log.Info().
			Str("reminder_id", reminderID.String()).
			Str("reminder_type", reminder.reminderType).
			Time("scheduled_for", scheduledFor).
			Msg("Reminder rescheduled")
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
		Str("program_id", programID.String()).
		Str("user_id", event.UserID).
		Msg("Program reminders updated")

	return nil
}

func (h *ReminderEventHandlers) ProcessPendingReminders(ctx context.Context) error {
	query := `
	SELECT reminder_id, user_id, program_id, reminder_type, message, metadata
	FROM reminders
	WHERE status = 'pending'
	  AND scheduled_for <= NOW()
	LIMIT 100
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query pending reminders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reminderID, userID, programID uuid.UUID
		var reminderType, message string
		var metadataJSON []byte

		err := rows.Scan(&reminderID, &userID, &programID, &reminderType, &message, &metadataJSON)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan reminder row")
			continue
		}

		if err := h.sendReminder(ctx, reminderID, userID, programID, reminderType, message, metadataJSON); err != nil {
			log.Error().
				Err(err).
				Str("reminder_id", reminderID.String()).
				Msg("Failed to send reminder")
		}
	}

	return rows.Err()
}

func (h *ReminderEventHandlers) sendReminder(ctx context.Context, reminderID, userID, programID uuid.UUID, reminderType, message string, metadataJSON []byte) error {
	if h.publisher != nil {
		reminderEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "reminder.sent",
			"client_generated_id": uuid.New().String(),
			"user_id":             userID.String(),
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"source_service":      "reminder-service",
			"data": map[string]interface{}{
				"reminder_id":   reminderID.String(),
				"program_id":    programID.String(),
				"reminder_type": reminderType,
				"message":       message,
			},
		}
		if err := h.publisher.PublishEvent("reminder.sent", reminderEvent); err != nil {
			return fmt.Errorf("failed to publish reminder.sent event: %w", err)
		}
	}

	updateQuery := `
	UPDATE reminders
	SET status = 'sent', sent_at = NOW(), updated_at = NOW()
	WHERE reminder_id = $1
	`
	_, err := h.db.ExecContext(ctx, updateQuery, reminderID)
	if err != nil {
		return fmt.Errorf("failed to update reminder status: %w", err)
	}

	log.Info().
		Str("reminder_id", reminderID.String()).
		Str("user_id", userID.String()).
		Str("reminder_type", reminderType).
		Msg("Reminder sent")

	return nil
}

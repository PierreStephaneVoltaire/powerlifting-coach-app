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

type ProgramEventHandlers struct {
	db        *sql.DB
	publisher EventPublisher
}

func NewProgramEventHandlers(db *sql.DB, publisher EventPublisher) *ProgramEventHandlers {
	return &ProgramEventHandlers{
		db:        db,
		publisher: publisher,
	}
}

type ProgramPlanCreatedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		Name                 string `json:"name"`
		StartDate            string `json:"start_date"`
		CompDate             string `json:"comp_date"`
		TrainingDaysPerWeek  int    `json:"training_days_per_week"`
		Notes                string `json:"notes"`
	} `json:"data"`
}

func (h *ProgramEventHandlers) HandleProgramPlanCreated(ctx context.Context, payload []byte) error {
	var event ProgramPlanCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	programID := uuid.New()

	startDate, err := time.Parse("2006-01-02", event.Data.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start_date: %w", err)
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

	query := `
	INSERT INTO programs (program_id, user_id, name, start_date, comp_date, training_days_per_week, notes, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
	`

	_, err = tx.ExecContext(ctx, query, programID, userID, event.Data.Name, startDate, compDate, event.Data.TrainingDaysPerWeek, event.Data.Notes)
	if err != nil {
		return fmt.Errorf("failed to insert program: %w", err)
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
		Str("name", event.Data.Name).
		Msg("Program plan created")

	if h.publisher != nil {
		persistedEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "program.plan.persisted",
			"client_generated_id": event.ClientGeneratedID,
			"user_id":             event.UserID,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"source_service":      "program-service",
			"data": map[string]interface{}{
				"program_id":            programID.String(),
				"name":                  event.Data.Name,
				"start_date":            event.Data.StartDate,
				"comp_date":             event.Data.CompDate,
				"training_days_per_week": event.Data.TrainingDaysPerWeek,
				"notes":                 event.Data.Notes,
			},
		}
		if err := h.publisher.PublishEvent("program.plan.persisted", persistedEvent); err != nil {
			log.Error().Err(err).Msg("Failed to publish program.plan.persisted event")
		}
	}

	return nil
}

type WorkoutStartedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		WorkoutID      string `json:"workout_id"`
		StartTimestamp string `json:"start_timestamp"`
	} `json:"data"`
}

func (h *ProgramEventHandlers) HandleWorkoutStarted(ctx context.Context, payload []byte) error {
	var event WorkoutStartedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	workoutID, err := uuid.Parse(event.Data.WorkoutID)
	if err != nil {
		return fmt.Errorf("invalid workout_id: %w", err)
	}

	startTime, err := time.Parse(time.RFC3339, event.Data.StartTimestamp)
	if err != nil {
		return fmt.Errorf("invalid start_timestamp: %w", err)
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
	INSERT INTO workout_sessions (workout_session_id, user_id, started_at, status)
	VALUES ($1, $2, $3, 'in_progress')
	ON CONFLICT (workout_session_id) DO NOTHING
	`

	_, err = tx.ExecContext(ctx, query, workoutID, userID, startTime)
	if err != nil {
		return fmt.Errorf("failed to insert workout session: %w", err)
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
		Str("workout_id", workoutID.String()).
		Str("user_id", event.UserID).
		Msg("Workout started")

	return nil
}

type WorkoutCompletedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		WorkoutID        string      `json:"workout_id"`
		DurationMinutes  int         `json:"duration_minutes"`
		ExercisesSummary interface{} `json:"exercises_summary"`
		Notes            string      `json:"notes"`
	} `json:"data"`
}

func (h *ProgramEventHandlers) HandleWorkoutCompleted(ctx context.Context, payload []byte) error {
	var event WorkoutCompletedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	workoutID, err := uuid.Parse(event.Data.WorkoutID)
	if err != nil {
		return fmt.Errorf("invalid workout_id: %w", err)
	}

	summaryJSON, err := json.Marshal(event.Data.ExercisesSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal exercises summary: %w", err)
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
	UPDATE workout_sessions
	SET completed_at = NOW(), duration_minutes = $1, exercises_summary = $2, notes = $3, status = 'completed'
	WHERE workout_session_id = $4 AND user_id = $5
	`

	result, err := tx.ExecContext(ctx, query, event.Data.DurationMinutes, summaryJSON, event.Data.Notes, workoutID, userID)
	if err != nil {
		return fmt.Errorf("failed to update workout session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		insertQuery := `
		INSERT INTO workout_sessions (workout_session_id, user_id, completed_at, duration_minutes, exercises_summary, notes, status)
		VALUES ($1, $2, NOW(), $3, $4, $5, 'completed')
		`
		_, err = tx.ExecContext(ctx, insertQuery, workoutID, userID, event.Data.DurationMinutes, summaryJSON, event.Data.Notes)
		if err != nil {
			return fmt.Errorf("failed to insert workout session: %w", err)
		}
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
		Str("workout_id", workoutID.String()).
		Str("user_id", event.UserID).
		Int("duration_minutes", event.Data.DurationMinutes).
		Msg("Workout completed")

	return nil
}

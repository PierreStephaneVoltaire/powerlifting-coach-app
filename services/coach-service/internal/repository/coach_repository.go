package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/powerlifting-coach-app/coach-service/internal/models"
)

type CoachRepository struct {
	db *sql.DB
}

func NewCoachRepository(db *sql.DB) *CoachRepository {
	return &CoachRepository{db: db}
}

func (r *CoachRepository) CreateFeedback(feedback *models.CoachFeedback) error {
	query := `
		INSERT INTO coach_feedback (coach_id, athlete_id, feedback_type, priority, title, 
		                           content, reference_type, reference_id, tags, is_private)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		feedback.CoachID, feedback.AthleteID, feedback.FeedbackType, feedback.Priority,
		feedback.Title, feedback.Content, feedback.ReferenceType, feedback.ReferenceID,
		pq.Array(feedback.Tags), feedback.IsPrivate,
	).Scan(&feedback.ID, &feedback.CreatedAt, &feedback.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}

	return nil
}

func (r *CoachRepository) GetFeedbackByID(id uuid.UUID) (*models.CoachFeedback, error) {
	query := `
		SELECT id, coach_id, athlete_id, feedback_type, priority, title, content,
		       reference_type, reference_id, tags, is_private, incorporated_by_ai,
		       incorporated_at, created_at, updated_at
		FROM coach_feedback WHERE id = $1`

	feedback := &models.CoachFeedback{}
	err := r.db.QueryRow(query, id).Scan(
		&feedback.ID, &feedback.CoachID, &feedback.AthleteID, &feedback.FeedbackType,
		&feedback.Priority, &feedback.Title, &feedback.Content, &feedback.ReferenceType,
		&feedback.ReferenceID, pq.Array(&feedback.Tags), &feedback.IsPrivate,
		&feedback.IncorporatedByAI, &feedback.IncorporatedAt,
		&feedback.CreatedAt, &feedback.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feedback not found")
		}
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	return feedback, nil
}

func (r *CoachRepository) GetFeedbackByCoachID(coachID uuid.UUID, page, pageSize int, athleteID *uuid.UUID, feedbackType *string) ([]models.CoachFeedback, int, error) {
	offset := (page - 1) * pageSize

	// Build WHERE clause
	whereClauses := []string{"coach_id = $1"}
	args := []interface{}{coachID}
	argIndex := 2

	if athleteID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("athlete_id = $%d", argIndex))
		args = append(args, *athleteID)
		argIndex++
	}

	if feedbackType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("feedback_type = $%d", argIndex))
		args = append(args, *feedbackType)
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM coach_feedback WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feedback count: %w", err)
	}

	// Get feedback
	query := fmt.Sprintf(`
		SELECT id, coach_id, athlete_id, feedback_type, priority, title, content,
		       reference_type, reference_id, tags, is_private, incorporated_by_ai,
		       incorporated_at, created_at, updated_at
		FROM coach_feedback 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []models.CoachFeedback
	for rows.Next() {
		var feedback models.CoachFeedback
		err := rows.Scan(
			&feedback.ID, &feedback.CoachID, &feedback.AthleteID, &feedback.FeedbackType,
			&feedback.Priority, &feedback.Title, &feedback.Content, &feedback.ReferenceType,
			&feedback.ReferenceID, pq.Array(&feedback.Tags), &feedback.IsPrivate,
			&feedback.IncorporatedByAI, &feedback.IncorporatedAt,
			&feedback.CreatedAt, &feedback.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feedback: %w", err)
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, totalCount, nil
}

func (r *CoachRepository) GetFeedbackByAthleteID(athleteID uuid.UUID, page, pageSize int) ([]models.CoachFeedback, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := `SELECT COUNT(*) FROM coach_feedback WHERE athlete_id = $1 AND is_private = false`
	var totalCount int
	err := r.db.QueryRow(countQuery, athleteID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feedback count: %w", err)
	}

	// Get feedback
	query := `
		SELECT id, coach_id, athlete_id, feedback_type, priority, title, content,
		       reference_type, reference_id, tags, is_private, incorporated_by_ai,
		       incorporated_at, created_at, updated_at
		FROM coach_feedback 
		WHERE athlete_id = $1 AND is_private = false 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, athleteID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []models.CoachFeedback
	for rows.Next() {
		var feedback models.CoachFeedback
		err := rows.Scan(
			&feedback.ID, &feedback.CoachID, &feedback.AthleteID, &feedback.FeedbackType,
			&feedback.Priority, &feedback.Title, &feedback.Content, &feedback.ReferenceType,
			&feedback.ReferenceID, pq.Array(&feedback.Tags), &feedback.IsPrivate,
			&feedback.IncorporatedByAI, &feedback.IncorporatedAt,
			&feedback.CreatedAt, &feedback.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feedback: %w", err)
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, totalCount, nil
}

func (r *CoachRepository) UpdateFeedback(feedback *models.CoachFeedback) error {
	query := `
		UPDATE coach_feedback SET
			feedback_type = $2, priority = $3, title = $4, content = $5,
			reference_type = $6, reference_id = $7, tags = $8, is_private = $9
		WHERE id = $1`

	_, err := r.db.Exec(query,
		feedback.ID, feedback.FeedbackType, feedback.Priority, feedback.Title,
		feedback.Content, feedback.ReferenceType, feedback.ReferenceID,
		pq.Array(feedback.Tags), feedback.IsPrivate,
	)

	if err != nil {
		return fmt.Errorf("failed to update feedback: %w", err)
	}

	return nil
}

func (r *CoachRepository) MarkFeedbackIncorporated(feedbackID uuid.UUID) error {
	query := `
		UPDATE coach_feedback SET
			incorporated_by_ai = true,
			incorporated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(query, feedbackID)
	if err != nil {
		return fmt.Errorf("failed to mark feedback as incorporated: %w", err)
	}

	return nil
}

func (r *CoachRepository) CreateNote(note *models.CoachAthleteNote) error {
	query := `
		INSERT INTO coach_athlete_notes (coach_id, athlete_id, note_type, title, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		note.CoachID, note.AthleteID, note.NoteType, note.Title, note.Content,
	).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

func (r *CoachRepository) GetNotesByAthleteID(coachID, athleteID uuid.UUID) ([]models.CoachAthleteNote, error) {
	query := `
		SELECT id, coach_id, athlete_id, note_type, title, content, is_archived,
		       created_at, updated_at
		FROM coach_athlete_notes 
		WHERE coach_id = $1 AND athlete_id = $2 AND is_archived = false
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, coachID, athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	defer rows.Close()

	var notes []models.CoachAthleteNote
	for rows.Next() {
		var note models.CoachAthleteNote
		err := rows.Scan(
			&note.ID, &note.CoachID, &note.AthleteID, &note.NoteType,
			&note.Title, &note.Content, &note.IsArchived,
			&note.CreatedAt, &note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func (r *CoachRepository) CreateFeedbackResponse(response *models.FeedbackResponse) error {
	query := `
		INSERT INTO feedback_responses (feedback_id, athlete_id, response_text, is_acknowledgment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		response.FeedbackID, response.AthleteID, response.ResponseText, response.IsAcknowledgment,
	).Scan(&response.ID, &response.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create feedback response: %w", err)
	}

	return nil
}

func (r *CoachRepository) GetFeedbackResponses(feedbackID uuid.UUID) ([]models.FeedbackResponse, error) {
	query := `
		SELECT id, feedback_id, athlete_id, response_text, is_acknowledgment, created_at
		FROM feedback_responses 
		WHERE feedback_id = $1 
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, feedbackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback responses: %w", err)
	}
	defer rows.Close()

	var responses []models.FeedbackResponse
	for rows.Next() {
		var response models.FeedbackResponse
		err := rows.Scan(
			&response.ID, &response.FeedbackID, &response.AthleteID,
			&response.ResponseText, &response.IsAcknowledgment, &response.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback response: %w", err)
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (r *CoachRepository) CreateNotification(notification *models.CoachNotification) error {
	query := `
		INSERT INTO coach_notifications (coach_id, athlete_id, notification_type, title, 
		                                message, reference_type, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		notification.CoachID, notification.AthleteID, notification.NotificationType,
		notification.Title, notification.Message, notification.ReferenceType,
		notification.ReferenceID,
	).Scan(&notification.ID, &notification.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *CoachRepository) GetNotificationsByCoachID(coachID uuid.UUID, limit int, unreadOnly bool) ([]models.CoachNotification, error) {
	var query string
	if unreadOnly {
		query = `
			SELECT id, coach_id, athlete_id, notification_type, title, message,
			       reference_type, reference_id, is_read, is_archived, created_at, read_at
			FROM coach_notifications 
			WHERE coach_id = $1 AND is_read = false AND is_archived = false
			ORDER BY created_at DESC 
			LIMIT $2`
	} else {
		query = `
			SELECT id, coach_id, athlete_id, notification_type, title, message,
			       reference_type, reference_id, is_read, is_archived, created_at, read_at
			FROM coach_notifications 
			WHERE coach_id = $1 AND is_archived = false
			ORDER BY created_at DESC 
			LIMIT $2`
	}

	rows, err := r.db.Query(query, coachID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.CoachNotification
	for rows.Next() {
		var notification models.CoachNotification
		err := rows.Scan(
			&notification.ID, &notification.CoachID, &notification.AthleteID,
			&notification.NotificationType, &notification.Title, &notification.Message,
			&notification.ReferenceType, &notification.ReferenceID,
			&notification.IsRead, &notification.IsArchived,
			&notification.CreatedAt, &notification.ReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (r *CoachRepository) MarkNotificationAsRead(notificationID uuid.UUID) error {
	query := `
		UPDATE coach_notifications SET
			is_read = true,
			read_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (r *CoachRepository) TrackAthleteProgress(progress *models.AthleteProgressTracking) error {
	measurementsJSON, _ := json.Marshal(progress.Measurements)

	query := `
		INSERT INTO athlete_progress_tracking (coach_id, athlete_id, tracking_date,
		                                      body_weight_kg, squat_max_kg, bench_max_kg,
		                                      deadlift_max_kg, total_kg, notes, measurements)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (coach_id, athlete_id, tracking_date)
		DO UPDATE SET
			body_weight_kg = EXCLUDED.body_weight_kg,
			squat_max_kg = EXCLUDED.squat_max_kg,
			bench_max_kg = EXCLUDED.bench_max_kg,
			deadlift_max_kg = EXCLUDED.deadlift_max_kg,
			total_kg = EXCLUDED.total_kg,
			notes = EXCLUDED.notes,
			measurements = EXCLUDED.measurements
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		progress.CoachID, progress.AthleteID, progress.TrackingDate,
		progress.BodyWeightKg, progress.SquatMaxKg, progress.BenchMaxKg,
		progress.DeadliftMaxKg, progress.TotalKg, progress.Notes, measurementsJSON,
	).Scan(&progress.ID, &progress.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to track athlete progress: %w", err)
	}

	return nil
}

func (r *CoachRepository) GetAthleteProgress(coachID, athleteID uuid.UUID, limit int) ([]models.AthleteProgressTracking, error) {
	query := `
		SELECT id, coach_id, athlete_id, tracking_date, body_weight_kg,
		       squat_max_kg, bench_max_kg, deadlift_max_kg, total_kg,
		       notes, measurements, created_at
		FROM athlete_progress_tracking 
		WHERE coach_id = $1 AND athlete_id = $2
		ORDER BY tracking_date DESC 
		LIMIT $3`

	rows, err := r.db.Query(query, coachID, athleteID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get athlete progress: %w", err)
	}
	defer rows.Close()

	var progress []models.AthleteProgressTracking
	for rows.Next() {
		var p models.AthleteProgressTracking
		var measurementsJSON []byte

		err := rows.Scan(
			&p.ID, &p.CoachID, &p.AthleteID, &p.TrackingDate,
			&p.BodyWeightKg, &p.SquatMaxKg, &p.BenchMaxKg,
			&p.DeadliftMaxKg, &p.TotalKg, &p.Notes,
			&measurementsJSON, &p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan athlete progress: %w", err)
		}

		if len(measurementsJSON) > 0 {
			json.Unmarshal(measurementsJSON, &p.Measurements)
		}

		progress = append(progress, p)
	}

	return progress, nil
}
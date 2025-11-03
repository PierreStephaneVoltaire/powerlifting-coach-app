package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/powerlifting-coach-app/program-service/internal/models"
)

type ProgramRepository struct {
	db *sql.DB
}

func NewProgramRepository(db *sql.DB) *ProgramRepository {
	return &ProgramRepository{db: db}
}

func (r *ProgramRepository) CreateProgram(program *models.Program) error {
	programDataJSON, _ := json.Marshal(program.ProgramData)

	query := `
		INSERT INTO programs (athlete_id, coach_id, name, description, phase, start_date, 
		                     end_date, weeks_total, days_per_week, program_data, ai_generated, 
		                     ai_model, ai_prompt)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		program.AthleteID, program.CoachID, program.Name, program.Description,
		program.Phase, program.StartDate, program.EndDate, program.WeeksTotal,
		program.DaysPerWeek, programDataJSON, program.AIGenerated,
		program.AIModel, program.AIPrompt,
	).Scan(&program.ID, &program.CreatedAt, &program.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create program: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetProgramByID(id uuid.UUID) (*models.Program, error) {
	query := `
		SELECT id, athlete_id, coach_id, name, description, phase, start_date, end_date,
		       weeks_total, days_per_week, program_data, ai_generated, ai_model, ai_prompt,
		       is_active, created_at, updated_at
		FROM programs WHERE id = $1`

	program := &models.Program{}
	var programDataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&program.ID, &program.AthleteID, &program.CoachID, &program.Name,
		&program.Description, &program.Phase, &program.StartDate, &program.EndDate,
		&program.WeeksTotal, &program.DaysPerWeek, &programDataJSON,
		&program.AIGenerated, &program.AIModel, &program.AIPrompt,
		&program.IsActive, &program.CreatedAt, &program.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("program not found")
		}
		return nil, fmt.Errorf("failed to get program: %w", err)
	}

	if len(programDataJSON) > 0 {
		json.Unmarshal(programDataJSON, &program.ProgramData)
	}

	return program, nil
}

func (r *ProgramRepository) GetProgramsByAthleteID(athleteID uuid.UUID) ([]models.Program, error) {
	query := `
		SELECT id, athlete_id, coach_id, name, description, phase, start_date, end_date,
		       weeks_total, days_per_week, program_data, ai_generated, ai_model, ai_prompt,
		       is_active, created_at, updated_at
		FROM programs 
		WHERE athlete_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get programs: %w", err)
	}
	defer rows.Close()

	var programs []models.Program
	for rows.Next() {
		var program models.Program
		var programDataJSON []byte

		err := rows.Scan(
			&program.ID, &program.AthleteID, &program.CoachID, &program.Name,
			&program.Description, &program.Phase, &program.StartDate, &program.EndDate,
			&program.WeeksTotal, &program.DaysPerWeek, &programDataJSON,
			&program.AIGenerated, &program.AIModel, &program.AIPrompt,
			&program.IsActive, &program.CreatedAt, &program.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan program: %w", err)
		}

		if len(programDataJSON) > 0 {
			json.Unmarshal(programDataJSON, &program.ProgramData)
		}

		programs = append(programs, program)
	}

	return programs, nil
}

func (r *ProgramRepository) UpdateProgram(program *models.Program) error {
	programDataJSON, _ := json.Marshal(program.ProgramData)

	query := `
		UPDATE programs SET
			name = $2, description = $3, phase = $4, start_date = $5,
			end_date = $6, weeks_total = $7, days_per_week = $8,
			program_data = $9, is_active = $10
		WHERE id = $1`

	_, err := r.db.Exec(query,
		program.ID, program.Name, program.Description, program.Phase,
		program.StartDate, program.EndDate, program.WeeksTotal,
		program.DaysPerWeek, programDataJSON, program.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update program: %w", err)
	}

	return nil
}

func (r *ProgramRepository) CreateTrainingSession(session *models.TrainingSession) error {
	query := `
		INSERT INTO training_sessions (program_id, athlete_id, week_number, day_number, 
		                              session_name, scheduled_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		session.ProgramID, session.AthleteID, session.WeekNumber, session.DayNumber,
		session.SessionName, session.ScheduledDate, session.Notes,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create training session: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetSessionsByProgramID(programID uuid.UUID) ([]models.TrainingSession, error) {
	query := `
		SELECT id, program_id, athlete_id, week_number, day_number, session_name,
		       scheduled_date, completed_at, notes, rpe_rating, duration_minutes,
		       created_at, updated_at
		FROM training_sessions 
		WHERE program_id = $1 
		ORDER BY week_number, day_number`

	rows, err := r.db.Query(query, programID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.TrainingSession
	for rows.Next() {
		var session models.TrainingSession

		err := rows.Scan(
			&session.ID, &session.ProgramID, &session.AthleteID,
			&session.WeekNumber, &session.DayNumber, &session.SessionName,
			&session.ScheduledDate, &session.CompletedAt, &session.Notes,
			&session.RPERating, &session.DurationMins,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Load exercises for this session
		exercises, err := r.GetExercisesBySessionID(session.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get exercises for session %s: %w", session.ID, err)
		}
		session.Exercises = exercises

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *ProgramRepository) CreateExercise(exercise *models.Exercise) error {
	query := `
		INSERT INTO exercises (session_id, exercise_order, lift_type, exercise_name,
		                      target_sets, target_reps, target_weight_kg, target_rpe,
		                      target_percentage, rest_seconds, notes, tempo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		exercise.SessionID, exercise.ExerciseOrder, exercise.LiftType,
		exercise.ExerciseName, exercise.TargetSets, exercise.TargetReps,
		exercise.TargetWeightKg, exercise.TargetRPE, exercise.TargetPercentage,
		exercise.RestSeconds, exercise.Notes, exercise.Tempo,
	).Scan(&exercise.ID, &exercise.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create exercise: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetExercisesBySessionID(sessionID uuid.UUID) ([]models.Exercise, error) {
	query := `
		SELECT id, session_id, exercise_order, lift_type, exercise_name,
		       target_sets, target_reps, target_weight_kg, target_rpe,
		       target_percentage, rest_seconds, notes, tempo, created_at
		FROM exercises 
		WHERE session_id = $1 
		ORDER BY exercise_order`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercises: %w", err)
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var exercise models.Exercise

		err := rows.Scan(
			&exercise.ID, &exercise.SessionID, &exercise.ExerciseOrder,
			&exercise.LiftType, &exercise.ExerciseName, &exercise.TargetSets,
			&exercise.TargetReps, &exercise.TargetWeightKg, &exercise.TargetRPE,
			&exercise.TargetPercentage, &exercise.RestSeconds, &exercise.Notes,
			&exercise.Tempo, &exercise.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}

		// Load completed sets for this exercise
		sets, err := r.GetCompletedSetsByExerciseID(exercise.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get completed sets for exercise %s: %w", exercise.ID, err)
		}
		exercise.CompletedSets = sets

		exercises = append(exercises, exercise)
	}

	return exercises, nil
}

func (r *ProgramRepository) LogCompletedSet(set *models.CompletedSet) error {
	query := `
		INSERT INTO completed_sets (exercise_id, set_number, reps_completed, weight_kg,
		                           rpe_actual, video_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, completed_at`

	err := r.db.QueryRow(query,
		set.ExerciseID, set.SetNumber, set.RepsCompleted, set.WeightKg,
		set.RPEActual, set.VideoID, set.Notes,
	).Scan(&set.ID, &set.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to log completed set: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetCompletedSetsByExerciseID(exerciseID uuid.UUID) ([]models.CompletedSet, error) {
	query := `
		SELECT id, exercise_id, set_number, reps_completed, weight_kg,
		       rpe_actual, video_id, notes, completed_at
		FROM completed_sets 
		WHERE exercise_id = $1 
		ORDER BY set_number`

	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed sets: %w", err)
	}
	defer rows.Close()

	var sets []models.CompletedSet
	for rows.Next() {
		var set models.CompletedSet

		err := rows.Scan(
			&set.ID, &set.ExerciseID, &set.SetNumber, &set.RepsCompleted,
			&set.WeightKg, &set.RPEActual, &set.VideoID, &set.Notes,
			&set.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan completed set: %w", err)
		}

		sets = append(sets, set)
	}

	return sets, nil
}

func (r *ProgramRepository) CompleteSession(sessionID uuid.UUID, notes *string, rpeRating *float64, durationMins *int) error {
	query := `
		UPDATE training_sessions SET
			completed_at = NOW(),
			notes = COALESCE($2, notes),
			rpe_rating = COALESCE($3, rpe_rating),
			duration_minutes = COALESCE($4, duration_minutes)
		WHERE id = $1`

	_, err := r.db.Exec(query, sessionID, notes, rpeRating, durationMins)
	if err != nil {
		return fmt.Errorf("failed to complete session: %w", err)
	}

	return nil
}

func (r *ProgramRepository) CreateAIConversation(conversation *models.AIConversation) error {
	messagesJSON, _ := json.Marshal(conversation.Messages)
	feedbackJSON, _ := json.Marshal(conversation.CoachFeedbackIncorp)

	query := `
		INSERT INTO ai_conversations (athlete_id, program_id, conversation_type, messages,
		                             coach_context_enabled, coach_feedback_incorporated)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		conversation.AthleteID, conversation.ProgramID, conversation.ConversationType,
		messagesJSON, conversation.CoachContextEnabled, feedbackJSON,
	).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create AI conversation: %w", err)
	}

	return nil
}

func (r *ProgramRepository) UpdateAIConversation(conversation *models.AIConversation) error {
	messagesJSON, _ := json.Marshal(conversation.Messages)
	feedbackJSON, _ := json.Marshal(conversation.CoachFeedbackIncorp)

	query := `
		UPDATE ai_conversations SET
			messages = $2,
			coach_feedback_incorporated = $3
		WHERE id = $1`

	_, err := r.db.Exec(query, conversation.ID, messagesJSON, feedbackJSON)
	if err != nil {
		return fmt.Errorf("failed to update AI conversation: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetAIConversationsByAthleteID(athleteID uuid.UUID, limit int) ([]models.AIConversation, error) {
	query := `
		SELECT id, athlete_id, program_id, conversation_type, messages,
		       coach_context_enabled, coach_feedback_incorporated,
		       created_at, updated_at
		FROM ai_conversations 
		WHERE athlete_id = $1 
		ORDER BY updated_at DESC 
		LIMIT $2`

	rows, err := r.db.Query(query, athleteID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIConversation
	for rows.Next() {
		var conversation models.AIConversation
		var messagesJSON, feedbackJSON []byte

		err := rows.Scan(
			&conversation.ID, &conversation.AthleteID, &conversation.ProgramID,
			&conversation.ConversationType, &messagesJSON,
			&conversation.CoachContextEnabled, &feedbackJSON,
			&conversation.CreatedAt, &conversation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AI conversation: %w", err)
		}

		if len(messagesJSON) > 0 {
			json.Unmarshal(messagesJSON, &conversation.Messages)
		}
		if len(feedbackJSON) > 0 {
			json.Unmarshal(feedbackJSON, &conversation.CoachFeedbackIncorp)
		}

		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

func (r *ProgramRepository) GetProgramTemplates(category string, experienceLevel string) ([]models.ProgramTemplate, error) {
	var query strings.Builder
	var args []interface{}
	argIndex := 1

	query.WriteString(`
		SELECT id, name, description, category, experience_level, phase,
		       weeks_duration, days_per_week, template_data, is_public,
		       created_by, created_at
		FROM program_templates 
		WHERE is_public = true`)

	if category != "" {
		query.WriteString(fmt.Sprintf(" AND category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if experienceLevel != "" {
		query.WriteString(fmt.Sprintf(" AND experience_level = $%d", argIndex))
		args = append(args, experienceLevel)
	}

	query.WriteString(" ORDER BY category, name")

	rows, err := r.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get program templates: %w", err)
	}
	defer rows.Close()

	var templates []models.ProgramTemplate
	for rows.Next() {
		var template models.ProgramTemplate
		var templateDataJSON []byte

		err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.Category,
			&template.ExperienceLevel, &template.Phase, &template.WeeksDuration,
			&template.DaysPerWeek, &templateDataJSON, &template.IsPublic,
			&template.CreatedBy, &template.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan program template: %w", err)
		}

		if len(templateDataJSON) > 0 {
			json.Unmarshal(templateDataJSON, &template.TemplateData)
		}

		templates = append(templates, template)
	}

	return templates, nil
}
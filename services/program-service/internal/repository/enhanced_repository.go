package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/program-service/internal/models"
)

// GetPreviousSetsForExercise retrieves historical sets for autofill
func (r *ProgramRepository) GetPreviousSetsForExercise(athleteID uuid.UUID, exerciseName string, limit int) ([]models.PreviousSetData, error) {
	if limit == 0 {
		limit = 5 // default to last 5 sessions
	}

	query := `
		SELECT
			e.exercise_name,
			ts.completed_at as session_date,
			cs.set_number,
			cs.reps_completed,
			cs.weight_kg,
			cs.rpe_actual,
			cs.set_type
		FROM completed_sets cs
		JOIN exercises e ON cs.exercise_id = e.id
		JOIN training_sessions ts ON e.session_id = ts.id
		WHERE ts.athlete_id = $1
		  AND LOWER(e.exercise_name) = LOWER($2)
		  AND ts.completed_at IS NOT NULL
		  AND ts.deleted_at IS NULL
		ORDER BY ts.completed_at DESC, cs.set_number
		LIMIT $3`

	rows, err := r.db.Query(query, athleteID, exerciseName, limit*10) // Get more rows to handle multiple sets per session
	if err != nil {
		return nil, fmt.Errorf("failed to get previous sets: %w", err)
	}
	defer rows.Close()

	var previousSets []models.PreviousSetData
	for rows.Next() {
		var set models.PreviousSetData
		err := rows.Scan(
			&set.ExerciseName,
			&set.SessionDate,
			&set.SetNumber,
			&set.RepsCompleted,
			&set.WeightKg,
			&set.RPEActual,
			&set.SetType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan previous set: %w", err)
		}
		previousSets = append(previousSets, set)
	}

	return previousSets, nil
}

// GenerateWarmupSets calculates warm-up sets based on working weight
func (r *ProgramRepository) GenerateWarmupSets(workingWeightKg float64, liftType string) []models.WarmupSet {
	// Standard powerlifting warm-up progression
	warmupPercentages := []struct {
		percentage float64
		reps       int
	}{
		{0.0, 5},   // Empty bar
		{0.4, 5},   // 40%
		{0.5, 5},   // 50%
		{0.6, 3},   // 60%
		{0.7, 2},   // 70%
		{0.85, 1},  // 85%
		{0.95, 1},  // 95%
	}

	// Bar weight (20kg for men, 15kg for women - using standard 20kg)
	barWeight := 20.0

	var warmups []models.WarmupSet
	for i, wp := range warmupPercentages {
		weight := workingWeightKg * wp.percentage
		if weight < barWeight {
			weight = barWeight
		}

		warmups = append(warmups, models.WarmupSet{
			SetNumber:       i + 1,
			WeightKg:        weight,
			Reps:            wp.reps,
			PercentageOfMax: wp.percentage * 100,
			PlateSetup:      calculatePlateSetup(weight, barWeight),
		})
	}

	return warmups
}

// calculatePlateSetup determines plate configuration for a given weight
func calculatePlateSetup(totalWeight, barWeight float64) string {
	if totalWeight <= barWeight {
		return "Empty bar"
	}

	weightPerSide := (totalWeight - barWeight) / 2.0

	// Standard plate sizes in kg
	plates := []float64{25, 20, 15, 10, 5, 2.5, 1.25, 0.5}
	var plateSetup []string
	remaining := weightPerSide

	for _, plate := range plates {
		count := int(remaining / plate)
		if count > 0 {
			plateSetup = append(plateSetup, fmt.Sprintf("%dx%.1fkg", count, plate))
			remaining -= float64(count) * plate
		}
	}

	if len(plateSetup) == 0 {
		return "Empty bar"
	}

	return fmt.Sprintf("%s per side", plateSetup)
}

// Exercise Library Methods

func (r *ProgramRepository) CreateExerciseLibrary(exercise *models.ExerciseLibrary) error {
	primaryMusclesJSON, _ := json.Marshal(exercise.PrimaryMuscles)
	secondaryMusclesJSON, _ := json.Marshal(exercise.SecondaryMuscles)
	equipmentJSON, _ := json.Marshal(exercise.EquipmentNeeded)
	formCuesJSON, _ := json.Marshal(exercise.FormCues)

	query := `
		INSERT INTO exercise_library (name, description, lift_type, primary_muscles,
		                              secondary_muscles, difficulty, equipment_needed,
		                              demo_video_url, instructions, form_cues, is_custom,
		                              created_by, is_public)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		exercise.Name, exercise.Description, exercise.LiftType,
		primaryMusclesJSON, secondaryMusclesJSON, exercise.Difficulty,
		equipmentJSON, exercise.DemoVideoURL, exercise.Instructions,
		formCuesJSON, exercise.IsCustom, exercise.CreatedBy, exercise.IsPublic,
	).Scan(&exercise.ID, &exercise.CreatedAt, &exercise.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create exercise library entry: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetExerciseLibrary(athleteID *uuid.UUID, liftType *models.LiftType) ([]models.ExerciseLibrary, error) {
	query := `
		SELECT id, name, description, lift_type, primary_muscles, secondary_muscles,
		       difficulty, equipment_needed, demo_video_url, instructions, form_cues,
		       is_custom, created_by, is_public, created_at, updated_at
		FROM exercise_library
		WHERE (is_public = TRUE OR created_by = $1)
		  AND ($2::lift_type IS NULL OR lift_type = $2)
		ORDER BY is_custom ASC, name ASC`

	rows, err := r.db.Query(query, athleteID, liftType)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise library: %w", err)
	}
	defer rows.Close()

	var exercises []models.ExerciseLibrary
	for rows.Next() {
		var ex models.ExerciseLibrary
		var primaryJSON, secondaryJSON, equipmentJSON, cuesJSON []byte

		err := rows.Scan(
			&ex.ID, &ex.Name, &ex.Description, &ex.LiftType,
			&primaryJSON, &secondaryJSON, &ex.Difficulty,
			&equipmentJSON, &ex.DemoVideoURL, &ex.Instructions,
			&cuesJSON, &ex.IsCustom, &ex.CreatedBy, &ex.IsPublic,
			&ex.CreatedAt, &ex.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}

		json.Unmarshal(primaryJSON, &ex.PrimaryMuscles)
		json.Unmarshal(secondaryJSON, &ex.SecondaryMuscles)
		json.Unmarshal(equipmentJSON, &ex.EquipmentNeeded)
		json.Unmarshal(cuesJSON, &ex.FormCues)

		exercises = append(exercises, ex)
	}

	return exercises, nil
}

// Workout Template Methods

func (r *ProgramRepository) CreateWorkoutTemplate(template *models.WorkoutTemplate) error {
	templateDataJSON, _ := json.Marshal(template.TemplateData)

	query := `
		INSERT INTO workout_templates (athlete_id, name, description, template_data, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		template.AthleteID, template.Name, template.Description,
		templateDataJSON, template.IsPublic,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create workout template: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetWorkoutTemplates(athleteID uuid.UUID) ([]models.WorkoutTemplate, error) {
	query := `
		SELECT id, athlete_id, name, description, template_data, is_public, times_used, created_at, updated_at
		FROM workout_templates
		WHERE athlete_id = $1 OR is_public = TRUE
		ORDER BY times_used DESC, created_at DESC`

	rows, err := r.db.Query(query, athleteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout templates: %w", err)
	}
	defer rows.Close()

	var templates []models.WorkoutTemplate
	for rows.Next() {
		var tmpl models.WorkoutTemplate
		var templateJSON []byte

		err := rows.Scan(
			&tmpl.ID, &tmpl.AthleteID, &tmpl.Name, &tmpl.Description,
			&templateJSON, &tmpl.IsPublic, &tmpl.TimesUsed,
			&tmpl.CreatedAt, &tmpl.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		json.Unmarshal(templateJSON, &tmpl.TemplateData)
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// Analytics Methods

func (r *ProgramRepository) GetVolumeData(athleteID uuid.UUID, startDate, endDate time.Time, exerciseName *string) ([]models.VolumeData, error) {
	query := `
		SELECT
			DATE(ts.completed_at) as date,
			e.exercise_name,
			COUNT(DISTINCT cs.id) as total_sets,
			SUM(cs.reps_completed) as total_reps,
			SUM(cs.reps_completed * cs.weight_kg) as total_volume,
			AVG(cs.weight_kg) as average_weight,
			AVG(cs.rpe_actual) as average_rpe
		FROM completed_sets cs
		JOIN exercises e ON cs.exercise_id = e.id
		JOIN training_sessions ts ON e.session_id = ts.id
		WHERE ts.athlete_id = $1
		  AND ts.completed_at BETWEEN $2 AND $3
		  AND ts.deleted_at IS NULL
		  AND ($4::text IS NULL OR LOWER(e.exercise_name) = LOWER($4))
		GROUP BY DATE(ts.completed_at), e.exercise_name
		ORDER BY date DESC, total_volume DESC`

	rows, err := r.db.Query(query, athleteID, startDate, endDate, exerciseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume data: %w", err)
	}
	defer rows.Close()

	var volumeData []models.VolumeData
	for rows.Next() {
		var vd models.VolumeData
		err := rows.Scan(
			&vd.Date, &vd.ExerciseName, &vd.TotalSets, &vd.TotalReps,
			&vd.TotalVolume, &vd.AverageWeight, &vd.AverageRPE,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan volume data: %w", err)
		}
		volumeData = append(volumeData, vd)
	}

	return volumeData, nil
}

func (r *ProgramRepository) GetE1RMData(athleteID uuid.UUID, startDate, endDate time.Time, liftType *models.LiftType) ([]models.E1RMData, error) {
	query := `
		SELECT
			DATE(ts.completed_at) as date,
			e.exercise_name,
			e.lift_type,
			cs.weight_kg,
			cs.reps_completed,
			cs.rpe_actual,
			-- Epley formula: weight * (1 + reps/30)
			cs.weight_kg * (1 + cs.reps_completed::float / 30.0) as estimated_1rm
		FROM completed_sets cs
		JOIN exercises e ON cs.exercise_id = e.id
		JOIN training_sessions ts ON e.session_id = ts.id
		WHERE ts.athlete_id = $1
		  AND ts.completed_at BETWEEN $2 AND $3
		  AND ts.deleted_at IS NULL
		  AND cs.reps_completed <= 10  -- Only sets under 10 reps for accuracy
		  AND cs.set_type IN ('working', 'amrap')  -- Only working sets
		  AND ($4::lift_type IS NULL OR e.lift_type = $4)
		ORDER BY date DESC, estimated_1rm DESC`

	rows, err := r.db.Query(query, athleteID, startDate, endDate, liftType)
	if err != nil {
		return nil, fmt.Errorf("failed to get e1RM data: %w", err)
	}
	defer rows.Close()

	var e1rmData []models.E1RMData
	for rows.Next() {
		var ed models.E1RMData
		err := rows.Scan(
			&ed.Date, &ed.ExerciseName, &ed.LiftType,
			&ed.WeightUsed, &ed.RepsAchieved, &ed.RPE,
			&ed.Estimated1RM,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan e1RM data: %w", err)
		}
		e1rmData = append(e1rmData, ed)
	}

	return e1rmData, nil
}

// Program Change Management (Git-like)

func (r *ProgramRepository) ProposeChange(change *models.ProgramChange) error {
	changesJSON, _ := json.Marshal(change.ProposedChanges)

	query := `
		INSERT INTO program_changes (program_id, change_type, proposed_changes,
		                             change_description, proposed_by, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		change.ProgramID, change.ChangeType, changesJSON,
		change.ChangeDescription, change.ProposedBy, change.Status,
	).Scan(&change.ID, &change.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to propose change: %w", err)
	}

	return nil
}

func (r *ProgramRepository) GetPendingChanges(programID uuid.UUID) ([]models.ProgramChange, error) {
	query := `
		SELECT id, program_id, change_type, proposed_changes, change_description,
		       proposed_by, status, created_at, applied_at
		FROM program_changes
		WHERE program_id = $1 AND status = 'pending'
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, programID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending changes: %w", err)
	}
	defer rows.Close()

	var changes []models.ProgramChange
	for rows.Next() {
		var change models.ProgramChange
		var changesJSON []byte

		err := rows.Scan(
			&change.ID, &change.ProgramID, &change.ChangeType,
			&changesJSON, &change.ChangeDescription, &change.ProposedBy,
			&change.Status, &change.CreatedAt, &change.AppliedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan change: %w", err)
		}

		json.Unmarshal(changesJSON, &change.ProposedChanges)
		changes = append(changes, change)
	}

	return changes, nil
}

func (r *ProgramRepository) ApplyChange(changeID uuid.UUID) error {
	query := `
		UPDATE program_changes
		SET status = 'applied', applied_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(query, changeID)
	if err != nil {
		return fmt.Errorf("failed to apply change: %w", err)
	}

	return nil
}

func (r *ProgramRepository) RejectChange(changeID uuid.UUID) error {
	query := `
		UPDATE program_changes
		SET status = 'rejected'
		WHERE id = $1`

	_, err := r.db.Exec(query, changeID)
	if err != nil {
		return fmt.Errorf("failed to reject change: %w", err)
	}

	return nil
}

// Historical Workout Management

func (r *ProgramRepository) GetSessionHistory(athleteID uuid.UUID, startDate, endDate time.Time, limit int) ([]models.TrainingSession, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, program_id, athlete_id, week_number, day_number, session_name,
		       scheduled_date, completed_at, notes, rpe_rating, duration_minutes,
		       created_at, updated_at
		FROM training_sessions
		WHERE athlete_id = $1
		  AND completed_at IS NOT NULL
		  AND deleted_at IS NULL
		  AND completed_at BETWEEN $2 AND $3
		ORDER BY completed_at DESC
		LIMIT $4`

	rows, err := r.db.Query(query, athleteID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session history: %w", err)
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
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *ProgramRepository) SoftDeleteSession(sessionID uuid.UUID, reason string) error {
	query := `
		UPDATE training_sessions
		SET deleted_at = NOW(), deleted_reason = $1
		WHERE id = $2`

	_, err := r.db.Exec(query, reason, sessionID)
	if err != nil {
		return fmt.Errorf("failed to soft delete session: %w", err)
	}

	return nil
}

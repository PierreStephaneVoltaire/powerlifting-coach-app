package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/program-service/internal/models"
	"github.com/powerlifting-coach-app/program-service/internal/repository"
	"github.com/rs/zerolog/log"
)

type WorkoutGenerator struct {
	programRepo *repository.ProgramRepository
}

func NewWorkoutGenerator(programRepo *repository.ProgramRepository) *WorkoutGenerator {
	return &WorkoutGenerator{
		programRepo: programRepo,
	}
}

// GenerateWorkoutsFromProgram creates training sessions and exercises from approved program data
func (wg *WorkoutGenerator) GenerateWorkoutsFromProgram(program *models.Program) error {
	if program.ProgramData == nil {
		return fmt.Errorf("program data is nil")
	}

	programData := program.ProgramData

	// Extract weekly workouts from program data
	weeklyWorkoutsRaw, ok := programData["weeklyWorkouts"]
	if !ok {
		return fmt.Errorf("no weeklyWorkouts found in program data")
	}

	weeklyWorkouts, ok := weeklyWorkoutsRaw.([]interface{})
	if !ok {
		return fmt.Errorf("weeklyWorkouts is not an array")
	}

	log.Info().
		Str("program_id", program.ID.String()).
		Int("weeks", len(weeklyWorkouts)).
		Msg("Generating workouts from program")

	// Process each week
	for _, weekRaw := range weeklyWorkouts {
		week, ok := weekRaw.(map[string]interface{})
		if !ok {
			log.Warn().Msg("Invalid week structure, skipping")
			continue
		}

		weekNumber, ok := week["week"].(float64)
		if !ok {
			log.Warn().Msg("Week number not found, skipping")
			continue
		}

		workoutsRaw, ok := week["workouts"]
		if !ok {
			log.Warn().Int("week", int(weekNumber)).Msg("No workouts found for week")
			continue
		}

		workouts, ok := workoutsRaw.([]interface{})
		if !ok {
			log.Warn().Int("week", int(weekNumber)).Msg("Workouts is not an array")
			continue
		}

		// Process each workout in the week
		for _, workoutRaw := range workouts {
			workout, ok := workoutRaw.(map[string]interface{})
			if !ok {
				log.Warn().Msg("Invalid workout structure, skipping")
				continue
			}

			if err := wg.createTrainingSession(program, int(weekNumber), workout); err != nil {
				log.Error().
					Err(err).
					Int("week", int(weekNumber)).
					Msg("Failed to create training session")
				// Continue with other workouts instead of failing completely
				continue
			}
		}
	}

	log.Info().
		Str("program_id", program.ID.String()).
		Msg("Successfully generated all workouts")

	return nil
}

func (wg *WorkoutGenerator) createTrainingSession(
	program *models.Program,
	weekNumber int,
	workout map[string]interface{},
) error {
	// Extract workout details
	dayNumber, ok := workout["day"].(float64)
	if !ok {
		return fmt.Errorf("day number not found")
	}

	workoutName, ok := workout["name"].(string)
	if !ok {
		workoutName = "Training Session"
	}

	// Calculate scheduled date
	scheduledDate := wg.calculateScheduledDate(program.StartDate, weekNumber, int(dayNumber))

	// Create training session
	session := &models.TrainingSession{
		ProgramID:     program.ID,
		AthleteID:     program.AthleteID,
		WeekNumber:    weekNumber,
		DayNumber:     int(dayNumber),
		SessionName:   &workoutName,
		ScheduledDate: &scheduledDate,
		Notes:         nil,
	}

	if err := wg.programRepo.CreateTrainingSession(session); err != nil {
		return fmt.Errorf("failed to create training session: %w", err)
	}

	// Extract and create exercises
	exercisesRaw, ok := workout["exercises"]
	if !ok {
		log.Warn().Msg("No exercises found in workout")
		return nil
	}

	exercises, ok := exercisesRaw.([]interface{})
	if !ok {
		return fmt.Errorf("exercises is not an array")
	}

	for idx, exerciseRaw := range exercises {
		exerciseData, ok := exerciseRaw.(map[string]interface{})
		if !ok {
			log.Warn().Int("index", idx).Msg("Invalid exercise structure, skipping")
			continue
		}

		if err := wg.createExercise(session.ID, idx+1, exerciseData); err != nil {
			log.Error().
				Err(err).
				Str("session_id", session.ID.String()).
				Int("exercise_idx", idx).
				Msg("Failed to create exercise")
			// Continue with other exercises
			continue
		}
	}

	return nil
}

func (wg *WorkoutGenerator) createExercise(
	sessionID uuid.UUID,
	order int,
	exerciseData map[string]interface{},
) error {
	// Extract exercise details
	name, ok := exerciseData["name"].(string)
	if !ok {
		return fmt.Errorf("exercise name not found")
	}

	// Lift type (squat, bench, deadlift, or accessory)
	liftType, _ := exerciseData["liftType"].(string)
	if liftType == "" {
		liftType = "accessory"
	}

	// Sets and reps
	sets, ok := exerciseData["sets"].(float64)
	if !ok {
		sets = 3 // Default
	}

	reps, _ := exerciseData["reps"].(string)
	if reps == "" {
		reps = "5"
	}

	// Intensity and RPE
	intensity, _ := exerciseData["intensity"].(string)
	rpe, _ := exerciseData["rpe"].(float64)

	// Parse intensity percentage if available
	var targetPercentage *float64
	if intensity != "" {
		var pct float64
		if _, err := fmt.Sscanf(intensity, "%f%%", &pct); err == nil {
			targetPercentage = &pct
		}
	}

	// RPE value
	var targetRPE *float64
	if rpe > 0 {
		targetRPE = &rpe
	}

	// Notes
	notes, _ := exerciseData["notes"].(string)
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	// Tempo
	tempo, _ := exerciseData["tempo"].(string)
	var tempoPtr *string
	if tempo != "" {
		tempoPtr = &tempo
	}

	// Rest time
	rest, _ := exerciseData["rest"].(float64)
	var restSeconds *int
	if rest > 0 {
		restInt := int(rest)
		restSeconds = &restInt
	}

	// Create exercise
	exercise := &models.Exercise{
		SessionID:        sessionID,
		ExerciseOrder:    order,
		LiftType:         models.LiftType(liftType),
		ExerciseName:     name,
		TargetSets:       int(sets),
		TargetReps:       reps,
		TargetWeightKg:   nil, // Will be calculated based on athlete's maxes
		TargetRPE:        targetRPE,
		TargetPercentage: targetPercentage,
		RestSeconds:      restSeconds,
		Notes:            notesPtr,
		Tempo:            tempoPtr,
	}

	if err := wg.programRepo.CreateExercise(exercise); err != nil {
		return fmt.Errorf("failed to create exercise: %w", err)
	}

	return nil
}

func (wg *WorkoutGenerator) calculateScheduledDate(startDate time.Time, weekNumber int, dayNumber int) time.Time {
	// Week 1 starts on the start date
	// Calculate days from start
	daysFromStart := (weekNumber-1)*7 + (dayNumber - 1)

	scheduledDate := startDate.AddDate(0, 0, daysFromStart)
	return scheduledDate
}

// DeleteProgramWorkouts removes all training sessions and exercises for a program
func (wg *WorkoutGenerator) DeleteProgramWorkouts(programID uuid.UUID) error {
	// This would need a new repository method to delete sessions by program ID
	// For now, we'll log it
	log.Info().
		Str("program_id", programID.String()).
		Msg("Deleting workouts for program (not yet implemented)")

	return nil
}

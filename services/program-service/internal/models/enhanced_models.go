package models

import (
	"time"
	"github.com/google/uuid"
)

// ExerciseLibrary represents a reusable exercise with metadata
type ExerciseLibrary struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description" db:"description"`
	LiftType         LiftType   `json:"lift_type" db:"lift_type"`
	PrimaryMuscles   []string   `json:"primary_muscles" db:"primary_muscles"`
	SecondaryMuscles []string   `json:"secondary_muscles" db:"secondary_muscles"`
	Difficulty       *string    `json:"difficulty" db:"difficulty"`
	EquipmentNeeded  []string   `json:"equipment_needed" db:"equipment_needed"`
	DemoVideoURL     *string    `json:"demo_video_url" db:"demo_video_url"`
	Instructions     *string    `json:"instructions" db:"instructions"`
	FormCues         []string   `json:"form_cues" db:"form_cues"`
	IsCustom         bool       `json:"is_custom" db:"is_custom"`
	CreatedBy        *uuid.UUID `json:"created_by" db:"created_by"`
	IsPublic         bool       `json:"is_public" db:"is_public"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// WorkoutTemplate represents a reusable workout structure
type WorkoutTemplate struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	AthleteID    uuid.UUID              `json:"athlete_id" db:"athlete_id"`
	Name         string                 `json:"name" db:"name"`
	Description  *string                `json:"description" db:"description"`
	TemplateData map[string]interface{} `json:"template_data" db:"template_data"`
	IsPublic     bool                   `json:"is_public" db:"is_public"`
	TimesUsed    int                    `json:"times_used" db:"times_used"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ProgramChange represents a proposed change to a program (git-like)
type ProgramChange struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	ProgramID         uuid.UUID              `json:"program_id" db:"program_id"`
	ChangeType        string                 `json:"change_type" db:"change_type"`
	ProposedChanges   map[string]interface{} `json:"proposed_changes" db:"proposed_changes"`
	ChangeDescription *string                `json:"change_description" db:"change_description"`
	ProposedBy        string                 `json:"proposed_by" db:"proposed_by"`
	Status            string                 `json:"status" db:"status"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	AppliedAt         *time.Time             `json:"applied_at" db:"applied_at"`
}

// PreviousSetData represents historical set data for autofill
type PreviousSetData struct {
	ExerciseName  string    `json:"exercise_name"`
	SessionDate   time.Time `json:"session_date"`
	SetNumber     int       `json:"set_number"`
	RepsCompleted int       `json:"reps_completed"`
	WeightKg      float64   `json:"weight_kg"`
	RPEActual     *float64  `json:"rpe_actual"`
	SetType       SetType   `json:"set_type"`
}

// WarmupSet represents a calculated warm-up set
type WarmupSet struct {
	SetNumber      int     `json:"set_number"`
	WeightKg       float64 `json:"weight_kg"`
	Reps           int     `json:"reps"`
	PercentageOfMax float64 `json:"percentage_of_max"`
	PlateSetup     string  `json:"plate_setup"`
}

// VolumeData represents volume tracking calculations
type VolumeData struct {
	Date          time.Time `json:"date"`
	ExerciseName  string    `json:"exercise_name"`
	TotalSets     int       `json:"total_sets"`
	TotalReps     int       `json:"total_reps"`
	TotalVolume   float64   `json:"total_volume"` // sets * reps * weight
	AverageWeight float64   `json:"average_weight"`
	AverageRPE    *float64  `json:"average_rpe"`
}

// E1RMData represents estimated 1RM calculations
type E1RMData struct {
	Date         time.Time `json:"date"`
	ExerciseName string    `json:"exercise_name"`
	LiftType     LiftType  `json:"lift_type"`
	Estimated1RM float64   `json:"estimated_1rm"` // Epley formula
	WeightUsed   float64   `json:"weight_used"`
	RepsAchieved int       `json:"reps_achieved"`
	RPE          *float64  `json:"rpe"`
}

// Request DTOs for new endpoints
type GetPreviousSetsRequest struct {
	ExerciseName string `json:"exercise_name" binding:"required"`
	Limit        int    `json:"limit"`
}

type GenerateWarmupsRequest struct {
	WorkingWeightKg float64 `json:"working_weight_kg" binding:"required"`
	LiftType        string  `json:"lift_type" binding:"required"`
}

type CreateExerciseLibraryRequest struct {
	Name             string   `json:"name" binding:"required"`
	Description      *string  `json:"description"`
	LiftType         LiftType `json:"lift_type" binding:"required"`
	PrimaryMuscles   []string `json:"primary_muscles"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Difficulty       *string  `json:"difficulty"`
	EquipmentNeeded  []string `json:"equipment_needed"`
	DemoVideoURL     *string  `json:"demo_video_url"`
	Instructions     *string  `json:"instructions"`
	FormCues         []string `json:"form_cues"`
}

type CreateWorkoutTemplateRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  *string                `json:"description"`
	TemplateData map[string]interface{} `json:"template_data" binding:"required"`
	IsPublic     bool                   `json:"is_public"`
}

type ProposeChangeRequest struct {
	ProgramID         uuid.UUID              `json:"program_id" binding:"required"`
	ProposedChanges   map[string]interface{} `json:"proposed_changes" binding:"required"`
	ChangeDescription *string                `json:"change_description"`
}

type GetVolumeDataRequest struct {
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	ExerciseName *string   `json:"exercise_name"`
	LiftType     *LiftType `json:"lift_type"`
}

type GetE1RMDataRequest struct {
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	LiftType     *LiftType `json:"lift_type"`
}

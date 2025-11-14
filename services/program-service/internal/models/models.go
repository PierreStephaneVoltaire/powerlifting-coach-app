package models

import (
	"time"
	"github.com/google/uuid"
)

type LiftType string
type ProgramPhase string

const (
	LiftTypeSquat     LiftType = "squat"
	LiftTypeBench     LiftType = "bench"
	LiftTypeDeadlift  LiftType = "deadlift"
	LiftTypeAccessory LiftType = "accessory"
)

const (
	PhaseHypertrophy ProgramPhase = "hypertrophy"
	PhaseStrength    ProgramPhase = "strength"
	PhasePeaking     ProgramPhase = "peaking"
	PhaseDeload      ProgramPhase = "deload"
	PhaseOffSeason   ProgramPhase = "off_season"
)

type ProgramStatus string

const (
	ProgramStatusDraft           ProgramStatus = "draft"
	ProgramStatusPendingApproval ProgramStatus = "pending_approval"
	ProgramStatusApproved        ProgramStatus = "approved"
	ProgramStatusRejected        ProgramStatus = "rejected"
)

type Program struct {
	ID                  uuid.UUID              `json:"id" db:"id"`
	AthleteID           uuid.UUID              `json:"athlete_id" db:"athlete_id"`
	CoachID             *uuid.UUID             `json:"coach_id" db:"coach_id"`
	Name                string                 `json:"name" db:"name"`
	Description         *string                `json:"description" db:"description"`
	Phase               ProgramPhase           `json:"phase" db:"phase"`
	StartDate           time.Time              `json:"start_date" db:"start_date"`
	EndDate             time.Time              `json:"end_date" db:"end_date"`
	WeeksTotal          int                    `json:"weeks_total" db:"weeks_total"`
	DaysPerWeek         int                    `json:"days_per_week" db:"days_per_week"`
	ProgramData         map[string]interface{} `json:"program_data" db:"program_data"`
	PendingProgramData  *map[string]interface{} `json:"pending_program_data,omitempty" db:"pending_program_data"`
	ProgramStatus       ProgramStatus          `json:"program_status" db:"program_status"`
	AIGenerated         bool                   `json:"ai_generated" db:"ai_generated"`
	AIModel             *string                `json:"ai_model" db:"ai_model"`
	AIPrompt            *string                `json:"ai_prompt" db:"ai_prompt"`
	IsActive            bool                   `json:"is_active" db:"is_active"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

type TrainingSession struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ProgramID     uuid.UUID  `json:"program_id" db:"program_id"`
	AthleteID     uuid.UUID  `json:"athlete_id" db:"athlete_id"`
	WeekNumber    int        `json:"week_number" db:"week_number"`
	DayNumber     int        `json:"day_number" db:"day_number"`
	SessionName   *string    `json:"session_name" db:"session_name"`
	ScheduledDate *time.Time `json:"scheduled_date" db:"scheduled_date"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	Notes         *string    `json:"notes" db:"notes"`
	RPERating     *float64   `json:"rpe_rating" db:"rpe_rating"`
	DurationMins  *int       `json:"duration_minutes" db:"duration_minutes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	Exercises     []Exercise `json:"exercises,omitempty"`
}

type Exercise struct {
	ID                uuid.UUID `json:"id" db:"id"`
	SessionID         uuid.UUID `json:"session_id" db:"session_id"`
	ExerciseOrder     int       `json:"exercise_order" db:"exercise_order"`
	LiftType          LiftType  `json:"lift_type" db:"lift_type"`
	ExerciseName      string    `json:"exercise_name" db:"exercise_name"`
	TargetSets        int       `json:"target_sets" db:"target_sets"`
	TargetReps        string    `json:"target_reps" db:"target_reps"`
	TargetWeightKg    *float64  `json:"target_weight_kg" db:"target_weight_kg"`
	TargetRPE         *float64  `json:"target_rpe" db:"target_rpe"`
	TargetPercentage  *float64  `json:"target_percentage" db:"target_percentage"`
	RestSeconds       *int      `json:"rest_seconds" db:"rest_seconds"`
	Notes             *string   `json:"notes" db:"notes"`
	Tempo             *string   `json:"tempo" db:"tempo"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	CompletedSets     []CompletedSet `json:"completed_sets,omitempty"`
}

type SetType string

const (
	SetTypeWarmUp   SetType = "warm_up"
	SetTypeWorking  SetType = "working"
	SetTypeBackoff  SetType = "backoff"
	SetTypeAMRAP    SetType = "amrap"
	SetTypeFailure  SetType = "failure"
	SetTypeDropSet  SetType = "drop_set"
	SetTypeCluster  SetType = "cluster"
	SetTypePause    SetType = "pause"
	SetTypeTempo    SetType = "tempo"
	SetTypeCustom   SetType = "custom"
)

type CompletedSet struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	ExerciseID    uuid.UUID       `json:"exercise_id" db:"exercise_id"`
	SetNumber     int             `json:"set_number" db:"set_number"`
	RepsCompleted int             `json:"reps_completed" db:"reps_completed"`
	WeightKg      float64         `json:"weight_kg" db:"weight_kg"`
	RPEActual     *float64        `json:"rpe_actual" db:"rpe_actual"`
	VideoID       *uuid.UUID      `json:"video_id" db:"video_id"`
	Notes         *string         `json:"notes" db:"notes"`
	SetType       SetType         `json:"set_type" db:"set_type"`
	MediaURLs     []string        `json:"media_urls" db:"media_urls"`
	ExerciseNotes *string         `json:"exercise_notes" db:"exercise_notes"`
	CompletedAt   time.Time       `json:"completed_at" db:"completed_at"`
}

type AIConversation struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	AthleteID               uuid.UUID              `json:"athlete_id" db:"athlete_id"`
	ProgramID               *uuid.UUID             `json:"program_id" db:"program_id"`
	ConversationType        string                 `json:"conversation_type" db:"conversation_type"`
	Messages                []Message              `json:"messages" db:"messages"`
	CoachContextEnabled     bool                   `json:"coach_context_enabled" db:"coach_context_enabled"`
	CoachFeedbackIncorp     []string               `json:"coach_feedback_incorporated" db:"coach_feedback_incorporated"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ProgramTemplate struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     *string                `json:"description" db:"description"`
	Category        string                 `json:"category" db:"category"`
	ExperienceLevel *string                `json:"experience_level" db:"experience_level"`
	Phase           ProgramPhase           `json:"phase" db:"phase"`
	WeeksDuration   int                    `json:"weeks_duration" db:"weeks_duration"`
	DaysPerWeek     int                    `json:"days_per_week" db:"days_per_week"`
	TemplateData    map[string]interface{} `json:"template_data" db:"template_data"`
	IsPublic        bool                   `json:"is_public" db:"is_public"`
	CreatedBy       *uuid.UUID             `json:"created_by" db:"created_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// Request/Response DTOs
type CreateProgramRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  *string                `json:"description"`
	Phase        ProgramPhase           `json:"phase" binding:"required"`
	StartDate    time.Time              `json:"start_date" binding:"required"`
	WeeksTotal   int                    `json:"weeks_total" binding:"required,min=1,max=52"`
	DaysPerWeek  int                    `json:"days_per_week" binding:"required,min=1,max=7"`
	ProgramData  map[string]interface{} `json:"program_data"`
	TemplateID   *uuid.UUID             `json:"template_id"`
}

type GenerateProgramRequest struct {
	Goals              string     `json:"goals" binding:"required"`
	ExperienceLevel    string     `json:"experience_level" binding:"required"`
	TrainingDays       int        `json:"training_days" binding:"required,min=1,max=7"`
	WeeksDuration      int        `json:"weeks_duration" binding:"required,min=1,max=52"`
	CompetitionDate    *time.Time `json:"competition_date"`
	CurrentMaxes       MaxLifts   `json:"current_maxes"`
	Injuries           *string    `json:"injuries"`
	Preferences        *string    `json:"preferences"`
	CoachContextEnable bool       `json:"coach_context_enable"`
}

type MaxLifts struct {
	SquatKg    *float64 `json:"squat_kg"`
	BenchKg    *float64 `json:"bench_kg"`
	DeadliftKg *float64 `json:"deadlift_kg"`
}

type ChatRequest struct {
	Message            string     `json:"message" binding:"required"`
	ProgramID          *uuid.UUID `json:"program_id"`
	CoachContextEnable bool       `json:"coach_context_enable"`
}

type LogWorkoutRequest struct {
	SessionID uuid.UUID             `json:"session_id" binding:"required"`
	Exercises []LoggedExercise      `json:"exercises" binding:"required"`
	Notes     *string               `json:"notes"`
	RPERating *float64              `json:"rpe_rating"`
	Duration  *int                  `json:"duration_minutes"`
}

type LoggedExercise struct {
	ExerciseID uuid.UUID    `json:"exercise_id" binding:"required"`
	Sets       []LoggedSet  `json:"sets" binding:"required"`
	Notes      *string      `json:"notes"`
}

type LoggedSet struct {
	SetNumber     int        `json:"set_number" binding:"required"`
	RepsCompleted int        `json:"reps_completed" binding:"required"`
	WeightKg      float64    `json:"weight_kg" binding:"required"`
	RPEActual     *float64   `json:"rpe_actual"`
	VideoID       *uuid.UUID `json:"video_id"`
	Notes         *string    `json:"notes"`
	SetType       SetType    `json:"set_type"`
	MediaURLs     []string   `json:"media_urls"`
	ExerciseNotes *string    `json:"exercise_notes"`
}

type ProgramResponse struct {
	Program  Program           `json:"program"`
	Sessions []TrainingSession `json:"sessions,omitempty"`
}

type ExportRequest struct {
	ProgramID uuid.UUID `json:"program_id" binding:"required"`
	Format    string    `json:"format" binding:"required,oneof=excel pdf"`
}
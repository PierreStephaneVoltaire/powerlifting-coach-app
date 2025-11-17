package models

import (
	"time"
	"github.com/google/uuid"
)

type UserType string
type ExperienceLevel string

const (
	UserTypeAthlete UserType = "athlete"
	UserTypeCoach   UserType = "coach"
)

const (
	ExperienceBeginner     ExperienceLevel = "beginner"
	ExperienceIntermediate ExperienceLevel = "intermediate"
	ExperienceAdvanced     ExperienceLevel = "advanced"
	ExperienceElite        ExperienceLevel = "elite"
)

type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	KeycloakID  string    `json:"keycloak_id" db:"keycloak_id"`
	Email       string    `json:"email" db:"email"`
	Name        string    `json:"name" db:"name"`
	UserType    UserType  `json:"user_type" db:"user_type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AthleteProfile struct {
	ID                   uuid.UUID        `json:"id" db:"id"`
	UserID               uuid.UUID        `json:"user_id" db:"user_id"`
	WeightKg             *float64         `json:"weight_kg" db:"weight_kg"`
	ExperienceLevel      *ExperienceLevel `json:"experience_level" db:"experience_level"`
	CompetitionDate      *time.Time       `json:"competition_date" db:"competition_date"`
	AccessCode           *string          `json:"access_code" db:"access_code"`
	AccessCodeExpiresAt  *time.Time       `json:"access_code_expires_at" db:"access_code_expires_at"`
	SquatMaxKg           *float64         `json:"squat_max_kg" db:"squat_max_kg"`
	BenchMaxKg           *float64         `json:"bench_max_kg" db:"bench_max_kg"`
	DeadliftMaxKg        *float64         `json:"deadlift_max_kg" db:"deadlift_max_kg"`
	TrainingFrequency    *int             `json:"training_frequency" db:"training_frequency"`
	Goals                *string          `json:"goals" db:"goals"`
	Injuries             *string          `json:"injuries" db:"injuries"`
	Bio                  *string          `json:"bio" db:"bio"`
	TargetWeightClass    *string          `json:"target_weight_class" db:"target_weight_class"`
	PreferredFederation  *string          `json:"preferred_federation" db:"preferred_federation"`
	CreatedAt            time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" db:"updated_at"`
}

type CoachProfile struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Bio             *string   `json:"bio" db:"bio"`
	Certifications  []string  `json:"certifications" db:"certifications"`
	YearsExperience *int      `json:"years_experience" db:"years_experience"`
	Specializations []string  `json:"specializations" db:"specializations"`
	HourlyRate      *float64  `json:"hourly_rate" db:"hourly_rate"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CoachAthleteAccess struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	CoachID    uuid.UUID  `json:"coach_id" db:"coach_id"`
	AthleteID  uuid.UUID  `json:"athlete_id" db:"athlete_id"`
	AccessCode string     `json:"access_code" db:"access_code"`
	GrantedAt  time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	IsActive   bool       `json:"is_active" db:"is_active"`
}

// Request/Response DTOs
type CreateUserRequest struct {
	KeycloakID string   `json:"keycloak_id" binding:"required"`
	Email      string   `json:"email" binding:"required,email"`
	Name       string   `json:"name" binding:"required"`
	UserType   UserType `json:"user_type" binding:"required"`
}

type UpdateAthleteProfileRequest struct {
	WeightKg            *float64         `json:"weight_kg"`
	ExperienceLevel     *ExperienceLevel `json:"experience_level"`
	CompetitionDate     *time.Time       `json:"competition_date"`
	SquatMaxKg          *float64         `json:"squat_max_kg"`
	BenchMaxKg          *float64         `json:"bench_max_kg"`
	DeadliftMaxKg       *float64         `json:"deadlift_max_kg"`
	TrainingFrequency   *int             `json:"training_frequency"`
	Goals               *string          `json:"goals"`
	Injuries            *string          `json:"injuries"`
}

type UpdateCoachProfileRequest struct {
	Bio             *string   `json:"bio"`
	Certifications  *[]string `json:"certifications"`
	YearsExperience *int      `json:"years_experience"`
	Specializations *[]string `json:"specializations"`
	HourlyRate      *float64  `json:"hourly_rate"`
}

type GrantAccessRequest struct {
	AccessCode string `json:"access_code" binding:"required,len=6"`
}

type GenerateAccessCodeRequest struct {
	ExpiresInWeeks *int `json:"expires_in_weeks"` // 0-12 weeks, nil for permanent
}

type UserResponse struct {
	User           User            `json:"user"`
	AthleteProfile *AthleteProfile `json:"athlete_profile,omitempty"`
	CoachProfile   *CoachProfile   `json:"coach_profile,omitempty"`
}
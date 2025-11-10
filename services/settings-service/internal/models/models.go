package models

import (
	"time"
	"github.com/google/uuid"
)

type UserSettings struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	UserID                  uuid.UUID              `json:"user_id" db:"user_id"`
	Theme                   string                 `json:"theme" db:"theme"`
	Language                string                 `json:"language" db:"language"`
	Timezone                string                 `json:"timezone" db:"timezone"`
	Units                   string                 `json:"units" db:"units"`
	Notifications           map[string]interface{} `json:"notifications" db:"notifications"`
	Privacy                 map[string]interface{} `json:"privacy" db:"privacy"`
	TrainingPreferences     map[string]interface{} `json:"training_preferences" db:"training_preferences"`
	WeightValue             *float64               `json:"weight_value,omitempty" db:"weight_value"`
	WeightUnit              *string                `json:"weight_unit,omitempty" db:"weight_unit"`
	Age                     *int                   `json:"age,omitempty" db:"age"`
	TargetWeightClass       *string                `json:"target_weight_class,omitempty" db:"target_weight_class"`
	CompetitionDate         *string                `json:"competition_date,omitempty" db:"competition_date"`
	SquatGoalValue          *float64               `json:"squat_goal_value,omitempty" db:"squat_goal_value"`
	SquatGoalUnit           *string                `json:"squat_goal_unit,omitempty" db:"squat_goal_unit"`
	BenchGoalValue          *float64               `json:"bench_goal_value,omitempty" db:"bench_goal_value"`
	BenchGoalUnit           *string                `json:"bench_goal_unit,omitempty" db:"bench_goal_unit"`
	DeadGoalValue           *float64               `json:"dead_goal_value,omitempty" db:"dead_goal_value"`
	DeadGoalUnit            *string                `json:"dead_goal_unit,omitempty" db:"dead_goal_unit"`
	MostImportantLift       *string                `json:"most_important_lift,omitempty" db:"most_important_lift"`
	LeastImportantLift      *string                `json:"least_important_lift,omitempty" db:"least_important_lift"`
	RecoveryRatingSquat     *int                   `json:"recovery_rating_squat,omitempty" db:"recovery_rating_squat"`
	RecoveryRatingBench     *int                   `json:"recovery_rating_bench,omitempty" db:"recovery_rating_bench"`
	RecoveryRatingDead      *int                   `json:"recovery_rating_dead,omitempty" db:"recovery_rating_dead"`
	TrainingDaysPerWeek     *int                   `json:"training_days_per_week,omitempty" db:"training_days_per_week"`
	SessionLengthMinutes    *int                   `json:"session_length_minutes,omitempty" db:"session_length_minutes"`
	WeightPlan              *string                `json:"weight_plan,omitempty" db:"weight_plan"`
	Injuries                *string                `json:"injuries,omitempty" db:"injuries"`
	KneeSleeve              *string                `json:"knee_sleeve,omitempty" db:"knee_sleeve"`
	DeadliftStyle           *string                `json:"deadlift_style,omitempty" db:"deadlift_style"`
	SquatStance             *string                `json:"squat_stance,omitempty" db:"squat_stance"`
	SquatBarPosition        *string                `json:"squat_bar_position,omitempty" db:"squat_bar_position"`
	VolumePreference        *string                `json:"volume_preference,omitempty" db:"volume_preference"`
	HeightValue             *float64               `json:"height_value,omitempty" db:"height_value"`
	HeightUnit              *string                `json:"height_unit,omitempty" db:"height_unit"`
	FeedVisibility          *string                `json:"feed_visibility,omitempty" db:"feed_visibility"`
	HasCompeted             *bool                  `json:"has_competed,omitempty" db:"has_competed"`
	BestSquatKg             *float64               `json:"best_squat_kg,omitempty" db:"best_squat_kg"`
	BestBenchKg             *float64               `json:"best_bench_kg,omitempty" db:"best_bench_kg"`
	BestDeadKg              *float64               `json:"best_dead_kg,omitempty" db:"best_dead_kg"`
	BestTotalKg             *float64               `json:"best_total_kg,omitempty" db:"best_total_kg"`
	CompPrDate              *string                `json:"comp_pr_date,omitempty" db:"comp_pr_date"`
	CompFederation          *string                `json:"comp_federation,omitempty" db:"comp_federation"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at" db:"updated_at"`
}

type AppSetting struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Key         string                 `json:"key" db:"key"`
	Value       map[string]interface{} `json:"value" db:"value"`
	Description *string                `json:"description" db:"description"`
	Category    string                 `json:"category" db:"category"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Request/Response DTOs
type UpdateUserSettingsRequest struct {
	Theme               *string                 `json:"theme"`
	Language            *string                 `json:"language"`
	Timezone            *string                 `json:"timezone"`
	Units               *string                 `json:"units"`
	Notifications       *map[string]interface{} `json:"notifications"`
	Privacy             *map[string]interface{} `json:"privacy"`
	TrainingPreferences *map[string]interface{} `json:"training_preferences"`
}

type UpdateAppSettingRequest struct {
	Value       map[string]interface{} `json:"value" binding:"required"`
	Description *string                `json:"description"`
}

type NotificationSettings struct {
	Email bool `json:"email"`
	Push  bool `json:"push"`
	SMS   bool `json:"sms"`
}

type PrivacySettings struct {
	ProfilePublic bool `json:"profile_public"`
	VideosPublic  bool `json:"videos_public"`
}

type TrainingPreferences struct {
	PreferredTrainingDays []string `json:"preferred_training_days"`
	SessionDurationMins   *int     `json:"session_duration_mins"`
	RestDaysBetween      *int     `json:"rest_days_between"`
	MaxSetsPerExercise   *int     `json:"max_sets_per_exercise"`
	PreferredTimeOfDay   *string  `json:"preferred_time_of_day"`
}
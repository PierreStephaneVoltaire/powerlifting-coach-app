package models

import (
	"time"
	"github.com/google/uuid"
)

type UserSettings struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	UserID               uuid.UUID              `json:"user_id" db:"user_id"`
	Theme                string                 `json:"theme" db:"theme"`
	Language             string                 `json:"language" db:"language"`
	Timezone             string                 `json:"timezone" db:"timezone"`
	Units                string                 `json:"units" db:"units"`
	Notifications        map[string]interface{} `json:"notifications" db:"notifications"`
	Privacy              map[string]interface{} `json:"privacy" db:"privacy"`
	TrainingPreferences  map[string]interface{} `json:"training_preferences" db:"training_preferences"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
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
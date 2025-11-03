package models

import (
	"time"
	"github.com/google/uuid"
)

type VideoStatus string

const (
	VideoStatusUploading  VideoStatus = "uploading"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusReady      VideoStatus = "ready"
	VideoStatusFailed     VideoStatus = "failed"
)

type Video struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	AthleteID          uuid.UUID  `json:"athlete_id" db:"athlete_id"`
	Filename           string     `json:"filename" db:"filename"`
	OriginalFilename   string     `json:"original_filename" db:"original_filename"`
	FileSize           int64      `json:"file_size" db:"file_size"`
	ContentType        string     `json:"content_type" db:"content_type"`
	DurationSeconds    *float64   `json:"duration_seconds" db:"duration_seconds"`
	OriginalURL        *string    `json:"original_url" db:"original_url"`
	ProcessedURL       *string    `json:"processed_url" db:"processed_url"`
	ThumbnailURL       *string    `json:"thumbnail_url" db:"thumbnail_url"`
	PublicShareToken   string     `json:"public_share_token" db:"public_share_token"`
	Status             VideoStatus `json:"status" db:"status"`
	ProcessingError    *string    `json:"processing_error" db:"processing_error"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	ProcessedAt        *time.Time `json:"processed_at" db:"processed_at"`
}

type FormFeedback struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	VideoID         uuid.UUID              `json:"video_id" db:"video_id"`
	FeedbackText    string                 `json:"feedback_text" db:"feedback_text"`
	ConfidenceScore *float64               `json:"confidence_score" db:"confidence_score"`
	Issues          []FormIssue            `json:"issues" db:"issues"`
	AIModel         *string                `json:"ai_model" db:"ai_model"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

type FormIssue struct {
	Type              string   `json:"type"`
	Description       string   `json:"description"`
	TimestampSeconds  *float64 `json:"timestamp_seconds"`
	Severity          *float64 `json:"severity"` // 0-1 scale
}

type VideoShare struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	VideoID     uuid.UUID  `json:"video_id" db:"video_id"`
	SharedBy    uuid.UUID  `json:"shared_by" db:"shared_by"`
	AccessLevel string     `json:"access_level" db:"access_level"`
	ViewCount   int        `json:"view_count" db:"view_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}

// Request/Response DTOs
type UploadRequest struct {
	Filename string `json:"filename" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
}

type UploadResponse struct {
	VideoID     uuid.UUID `json:"video_id"`
	UploadURL   string    `json:"upload_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type VideoListResponse struct {
	Videos     []Video `json:"videos"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

type ShareVideoRequest struct {
	AccessLevel string     `json:"access_level" binding:"oneof=public coach_only"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type ProcessVideoMessage struct {
	VideoID uuid.UUID `json:"video_id"`
}

type AIFeedbackRequest struct {
	VideoID uuid.UUID `json:"video_id" binding:"required"`
	Prompt  string    `json:"prompt"`
}

type VideoMetadata struct {
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Bitrate     int     `json:"bitrate,omitempty"`
	Framerate   float64 `json:"framerate,omitempty"`
	Codec       string  `json:"codec,omitempty"`
	Resolution  string  `json:"resolution,omitempty"`
}
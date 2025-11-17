package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MediaEventHandlers struct {
	db        *sql.DB
	publisher EventPublisher
}

func NewMediaEventHandlers(db *sql.DB, publisher EventPublisher) *MediaEventHandlers {
	return &MediaEventHandlers{
		db:        db,
		publisher: publisher,
	}
}

type MediaUploadRequestedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		Filename      string  `json:"filename"`
		ContentType   string  `json:"content_type"`
		FileSize      int64   `json:"file_size"`
		MovementLabel string  `json:"movement_label"`
		Weight        float64 `json:"weight"`
		RPE           float64 `json:"rpe"`
		CommentText   string  `json:"comment_text"`
		Visibility    string  `json:"visibility"`
	} `json:"data"`
}

func (h *MediaEventHandlers) HandleMediaUploadRequested(ctx context.Context, payload []byte) error {
	var event MediaUploadRequestedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	checkQuery := `SELECT COUNT(*) FROM media_idempotency_keys WHERE key = $1`
	var count int
	err = tx.QueryRowContext(ctx, checkQuery, event.ClientGeneratedID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if count > 0 {
		log.Info().Str("client_generated_id", event.ClientGeneratedID).Msg("Event already processed")
		return nil
	}

	uploadID := uuid.New()
	videoID := uuid.New()

	presignedURL := fmt.Sprintf("https://storage.example.com/uploads/%s", uploadID.String())

	uploadQuery := `
	INSERT INTO media_uploads (upload_id, user_id, video_id, filename, content_type, file_size, upload_status, presigned_url)
	VALUES ($1, $2, $3, $4, $5, $6, 'requested', $7)
	`

	_, err = tx.ExecContext(ctx, uploadQuery, uploadID, userID, videoID, event.Data.Filename, event.Data.ContentType, event.Data.FileSize, presignedURL)
	if err != nil {
		return fmt.Errorf("failed to insert media upload: %w", err)
	}

	videoQuery := `
	INSERT INTO videos (id, athlete_id, filename, original_filename, file_size, content_type, status, movement_label, weight, rpe, comment_text, visibility)
	VALUES ($1, $2, $3, $4, $5, $6, 'uploading', $7, $8, $9, $10, $11)
	`

	_, err = tx.ExecContext(ctx, videoQuery, videoID, userID, event.Data.Filename, event.Data.Filename, event.Data.FileSize, event.Data.ContentType, event.Data.MovementLabel, event.Data.Weight, event.Data.RPE, event.Data.CommentText, event.Data.Visibility)
	if err != nil {
		return fmt.Errorf("failed to insert video: %w", err)
	}

	idempotencyQuery := `INSERT INTO media_idempotency_keys (key, event_type, upload_id, processed_at) VALUES ($1, $2, $3, NOW())`
	_, err = tx.ExecContext(ctx, idempotencyQuery, event.ClientGeneratedID, event.EventType, uploadID)
	if err != nil {
		return fmt.Errorf("failed to insert idempotency key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("upload_id", uploadID.String()).
		Str("video_id", videoID.String()).
		Str("user_id", event.UserID).
		Msg("Media upload requested")

	if h.publisher != nil {
		mediaUploadedEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "media.uploaded",
			"client_generated_id": event.ClientGeneratedID,
			"user_id":             event.UserID,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"source_service":      "video-service",
			"data": map[string]interface{}{
				"upload_id":   uploadID.String(),
				"video_id":    videoID.String(),
				"filename":    event.Data.Filename,
				"presigned_url": presignedURL,
			},
		}
		if err := h.publisher.PublishEvent("media.uploaded", mediaUploadedEvent); err != nil {
			log.Error().Err(err).Msg("Failed to publish media.uploaded event")
		}
	}

	return nil
}

type MediaUploadedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		UploadID   string `json:"upload_id"`
		VideoID    string `json:"video_id"`
		Filename   string `json:"filename"`
		PresignedURL string `json:"presigned_url"`
	} `json:"data"`
}

func (h *MediaEventHandlers) HandleMediaUploaded(ctx context.Context, payload []byte) error {
	var event MediaUploadedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	uploadID, err := uuid.Parse(event.Data.UploadID)
	if err != nil {
		return fmt.Errorf("invalid upload_id: %w", err)
	}

	videoID, err := uuid.Parse(event.Data.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video_id: %w", err)
	}

	updateUploadQuery := `
	UPDATE media_uploads
	SET upload_status = 'uploaded', uploaded_at = NOW()
	WHERE upload_id = $1
	`
	_, err = h.db.ExecContext(ctx, updateUploadQuery, uploadID)
	if err != nil {
		return fmt.Errorf("failed to update upload status: %w", err)
	}

	updateVideoQuery := `
	UPDATE videos
	SET status = 'processing', original_url = $1
	WHERE id = $2
	`
	originalURL := fmt.Sprintf("https://storage.example.com/videos/%s", videoID.String())
	_, err = h.db.ExecContext(ctx, updateVideoQuery, originalURL, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	log.Info().
		Str("upload_id", uploadID.String()).
		Str("video_id", videoID.String()).
		Msg("Media uploaded, queuing for processing")

	// Queue video for processing by media-processor-service
	var filename string
	filenameQuery := `SELECT filename FROM videos WHERE id = $1`
	err = h.db.QueryRowContext(ctx, filenameQuery, videoID).Scan(&filename)
	if err != nil {
		return fmt.Errorf("failed to get filename: %w", err)
	}

	processMsg := map[string]interface{}{
		"video_id": videoID.String(),
		"user_id":  event.UserID,
		"filename": filename,
	}

	if h.publisher != nil {
		if err := h.publisher.PublishVideoProcessing(processMsg); err != nil {
			log.Error().Err(err).Msg("Failed to queue video for processing")
			// Update video status to failed
			failQuery := `UPDATE videos SET status = 'failed' WHERE id = $1`
			h.db.ExecContext(ctx, failQuery, videoID)
			return fmt.Errorf("failed to queue video processing: %w", err)
		}
	}

	log.Info().
		Str("video_id", videoID.String()).
		Msg("Video queued for processing")

	return nil
}

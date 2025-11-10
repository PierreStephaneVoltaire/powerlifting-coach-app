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

type MetadataHandlers struct {
	db        *sql.DB
	publisher EventPublisher
}

func NewMetadataHandlers(db *sql.DB, publisher EventPublisher) *MetadataHandlers {
	return &MetadataHandlers{
		db:        db,
		publisher: publisher,
	}
}

type VideoMetadataMessage struct {
	VideoID      string                 `json:"video_id"`
	ProcessedURL string                 `json:"processed_url"`
	ThumbnailURL string                 `json:"thumbnail_url"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func (h *MetadataHandlers) HandleVideoMetadata(payload []byte) error {
	var msg VideoMetadataMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	videoID, err := uuid.Parse(msg.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video_id: %w", err)
	}

	ctx := context.Background()

	// Update video with processed URLs and metadata
	updateQuery := `
	UPDATE videos
	SET status = 'ready', processed_url = $1, thumbnail_url = $2, processed_at = NOW(), metadata = $3
	WHERE id = $4
	`

	metadataJSON, err := json.Marshal(msg.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = h.db.ExecContext(ctx, updateQuery, msg.ProcessedURL, msg.ThumbnailURL, metadataJSON, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	log.Info().
		Str("video_id", msg.VideoID).
		Str("processed_url", msg.ProcessedURL).
		Str("thumbnail_url", msg.ThumbnailURL).
		Msg("Video processing completed")

	// Get video details for feed post creation
	var userID, movementLabel, commentText, visibility string
	var weight, rpe sql.NullFloat64

	videoQuery := `
	SELECT athlete_id, movement_label, weight, rpe, comment_text, visibility
	FROM videos
	WHERE id = $1
	`
	err = h.db.QueryRowContext(ctx, videoQuery, videoID).Scan(&userID, &movementLabel, &weight, &rpe, &commentText, &visibility)
	if err != nil {
		return fmt.Errorf("failed to fetch video details: %w", err)
	}

	// Publish media.processed event
	if h.publisher != nil {
		mediaProcessedEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "media.processed",
			"client_generated_id": uuid.New().String(),
			"user_id":             userID,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"source_service":      "video-service",
			"data": map[string]interface{}{
				"video_id":      msg.VideoID,
				"media_url":     msg.ProcessedURL,
				"thumbnail_url": msg.ThumbnailURL,
			},
		}
		if err := h.publisher.PublishEvent("media.processed", mediaProcessedEvent); err != nil {
			log.Error().Err(err).Msg("Failed to publish media.processed event")
		}

		// Publish feed.post.created event
		feedPostEvent := map[string]interface{}{
			"schema_version":      "1.0.0",
			"event_type":          "feed.post.created",
			"client_generated_id": uuid.New().String(),
			"user_id":             userID,
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"source_service":      "video-service",
			"data": map[string]interface{}{
				"post_id":        msg.VideoID,
				"content_type":   "video",
				"media_url":      msg.ProcessedURL,
				"thumbnail_url":  msg.ThumbnailURL,
				"movement_label": movementLabel,
				"weight":         weight.Float64,
				"rpe":            rpe.Float64,
				"caption":        commentText,
				"visibility":     visibility,
			},
		}
		if err := h.publisher.PublishEvent("feed.post.created", feedPostEvent); err != nil {
			log.Error().Err(err).Msg("Failed to publish feed.post.created event")
		}
	}

	return nil
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type FeedHandlers struct {
	db *sql.DB
}

func NewFeedHandlers(db *sql.DB) *FeedHandlers {
	return &FeedHandlers{db: db}
}

type FeedPostCreatedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		PostID        string  `json:"post_id"`
		MediaID       string  `json:"media_id"`
		MediaURL      string  `json:"media_url"`
		ThumbnailURL  string  `json:"thumbnail_url"`
		MovementLabel string  `json:"movement_label"`
		Weight        *struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		} `json:"weight"`
		RPE              *float64 `json:"rpe"`
		CommentText      string   `json:"comment_text"`
		Visibility       string   `json:"visibility"`
		PasscodeRequired bool     `json:"passcode_required"`
	} `json:"data"`
}

func (h *FeedHandlers) HandleFeedPostCreated(ctx context.Context, payload []byte) error {
	var event FeedPostCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	postID, err := uuid.Parse(event.Data.PostID)
	if err != nil {
		return fmt.Errorf("invalid post_id: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	var videoID *uuid.UUID
	if event.Data.MediaID != "" {
		id, err := uuid.Parse(event.Data.MediaID)
		if err != nil {
			return fmt.Errorf("invalid media_id: %w", err)
		}
		videoID = &id
	}

	query := `
	INSERT INTO feed_posts (
		post_id, user_id, video_id, visibility, movement_label,
		weight_value, weight_unit, rpe, comment_text
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (post_id) DO NOTHING
	`

	var weightValue *float64
	var weightUnit *string
	if event.Data.Weight != nil {
		weightValue = &event.Data.Weight.Value
		weightUnit = &event.Data.Weight.Unit
	}

	_, err = h.db.ExecContext(
		ctx, query,
		postID, userID, videoID, event.Data.Visibility, event.Data.MovementLabel,
		weightValue, weightUnit, event.Data.RPE, event.Data.CommentText,
	)
	if err != nil {
		return fmt.Errorf("failed to insert feed post: %w", err)
	}

	log.Info().
		Str("post_id", event.Data.PostID).
		Str("user_id", event.UserID).
		Msg("Feed post created")

	return nil
}

type FeedPostUpdatedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		PostID        string  `json:"post_id"`
		MovementLabel *string `json:"movement_label"`
		Weight        *struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		} `json:"weight"`
		RPE         *float64 `json:"rpe"`
		CommentText *string  `json:"comment_text"`
		Visibility  *string  `json:"visibility"`
	} `json:"data"`
}

func (h *FeedHandlers) HandleFeedPostUpdated(ctx context.Context, payload []byte) error {
	var event FeedPostUpdatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	postID, err := uuid.Parse(event.Data.PostID)
	if err != nil {
		return fmt.Errorf("invalid post_id: %w", err)
	}

	query := `
	UPDATE feed_posts SET
		movement_label = COALESCE($2, movement_label),
		weight_value = COALESCE($3, weight_value),
		weight_unit = COALESCE($4, weight_unit),
		rpe = COALESCE($5, rpe),
		comment_text = COALESCE($6, comment_text),
		visibility = COALESCE($7, visibility),
		updated_at = NOW()
	WHERE post_id = $1
	`

	var weightValue *float64
	var weightUnit *string
	if event.Data.Weight != nil {
		weightValue = &event.Data.Weight.Value
		weightUnit = &event.Data.Weight.Unit
	}

	_, err = h.db.ExecContext(
		ctx, query,
		postID, event.Data.MovementLabel,
		weightValue, weightUnit,
		event.Data.RPE, event.Data.CommentText, event.Data.Visibility,
	)
	if err != nil {
		return fmt.Errorf("failed to update feed post: %w", err)
	}

	log.Info().Str("post_id", event.Data.PostID).Msg("Feed post updated")
	return nil
}

type FeedPostDeletedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		PostID string `json:"post_id"`
		Reason string `json:"reason"`
	} `json:"data"`
}

func (h *FeedHandlers) HandleFeedPostDeleted(ctx context.Context, payload []byte) error {
	var event FeedPostDeletedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	postID, err := uuid.Parse(event.Data.PostID)
	if err != nil {
		return fmt.Errorf("invalid post_id: %w", err)
	}

	query := `DELETE FROM feed_posts WHERE post_id = $1`
	_, err = h.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("failed to delete feed post: %w", err)
	}

	log.Info().
		Str("post_id", event.Data.PostID).
		Str("reason", event.Data.Reason).
		Msg("Feed post deleted")

	return nil
}

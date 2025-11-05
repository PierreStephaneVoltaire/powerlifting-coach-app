package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type CommentHandlers struct {
	db *sql.DB
}

func NewCommentHandlers(db *sql.DB) *CommentHandlers {
	return &CommentHandlers{db: db}
}

type CommentCreatedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		PostID          string  `json:"post_id"`
		ParentCommentID *string `json:"parent_comment_id"`
		CommentText     string  `json:"comment_text"`
	} `json:"data"`
}

func (h *CommentHandlers) HandleCommentCreated(ctx context.Context, payload []byte) error {
	var event CommentCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	commentID := uuid.New()
	postID, err := uuid.Parse(event.Data.PostID)
	if err != nil {
		return fmt.Errorf("invalid post_id: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	var parentCommentID *uuid.UUID
	if event.Data.ParentCommentID != nil {
		id, err := uuid.Parse(*event.Data.ParentCommentID)
		if err != nil {
			return fmt.Errorf("invalid parent_comment_id: %w", err)
		}
		parentCommentID = &id
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
	INSERT INTO comments (comment_id, post_id, user_id, parent_comment_id, comment_text)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (comment_id) DO NOTHING
	`

	_, err = tx.ExecContext(ctx, query, commentID, postID, userID, parentCommentID, event.Data.CommentText)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	updateQuery := `UPDATE feed_posts SET comments_count = comments_count + 1 WHERE post_id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, postID)
	if err != nil {
		return fmt.Errorf("failed to update comment count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("comment_id", commentID.String()).
		Str("post_id", event.Data.PostID).
		Str("user_id", event.UserID).
		Msg("Comment created")

	return nil
}

type InteractionLikedEvent struct {
	SchemaVersion     string `json:"schema_version"`
	EventType         string `json:"event_type"`
	ClientGeneratedID string `json:"client_generated_id"`
	UserID            string `json:"user_id"`
	Timestamp         string `json:"timestamp"`
	SourceService     string `json:"source_service"`
	Data              struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Action     string `json:"action"`
	} `json:"data"`
}

func (h *CommentHandlers) HandleInteractionLiked(ctx context.Context, payload []byte) error {
	var event InteractionLikedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	targetID, err := uuid.Parse(event.Data.TargetID)
	if err != nil {
		return fmt.Errorf("invalid target_id: %w", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if event.Data.Action == "unlike" {
		query := `DELETE FROM likes WHERE user_id = $1 AND target_type = $2 AND target_id = $3`
		_, err = tx.ExecContext(ctx, query, userID, event.Data.TargetType, targetID)
		if err != nil {
			return fmt.Errorf("failed to delete like: %w", err)
		}

		if event.Data.TargetType == "post" {
			updateQuery := `UPDATE feed_posts SET likes_count = GREATEST(0, likes_count - 1) WHERE post_id = $1`
			_, err = tx.ExecContext(ctx, updateQuery, targetID)
			if err != nil {
				return fmt.Errorf("failed to update like count: %w", err)
			}
		}
	} else {
		query := `
		INSERT INTO likes (user_id, target_type, target_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, target_type, target_id) DO NOTHING
		`

		_, err = tx.ExecContext(ctx, query, userID, event.Data.TargetType, targetID)
		if err != nil {
			return fmt.Errorf("failed to insert like: %w", err)
		}

		if event.Data.TargetType == "post" {
			updateQuery := `UPDATE feed_posts SET likes_count = likes_count + 1 WHERE post_id = $1`
			_, err = tx.ExecContext(ctx, updateQuery, targetID)
			if err != nil {
				return fmt.Errorf("failed to update like count: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("user_id", event.UserID).
		Str("target_type", event.Data.TargetType).
		Str("target_id", event.Data.TargetID).
		Str("action", event.Data.Action).
		Msg("Interaction processed")

	return nil
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

type CommentResponse struct {
	ID              string    `json:"id"`
	CommentID       string    `json:"comment_id"`
	PostID          string    `json:"post_id"`
	UserID          string    `json:"user_id"`
	ParentCommentID *string   `json:"parent_comment_id"`
	CommentText     string    `json:"comment_text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LikeResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *CommentHandlers) GetPostComments(c *gin.Context) {
	postID := c.Param("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post_id"})
		return
	}

	query := `
	SELECT id, comment_id, post_id, user_id, parent_comment_id, comment_text, created_at, updated_at
	FROM comments
	WHERE post_id = $1
	ORDER BY created_at ASC
	`

	rows, err := h.db.QueryContext(c.Request.Context(), query, postUUID)
	if err != nil {
		log.Error().Err(err).Str("post_id", postID).Msg("Failed to query comments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}
	defer rows.Close()

	var comments []CommentResponse
	for rows.Next() {
		var comment CommentResponse
		var parentCommentID sql.NullString

		err := rows.Scan(
			&comment.ID, &comment.CommentID, &comment.PostID, &comment.UserID,
			&parentCommentID, &comment.CommentText, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan comment")
			continue
		}

		if parentCommentID.Valid {
			comment.ParentCommentID = &parentCommentID.String
		}

		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (h *CommentHandlers) GetPostLikes(c *gin.Context) {
	postID := c.Param("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post_id"})
		return
	}

	query := `
	SELECT id, user_id, target_type, target_id, created_at
	FROM likes
	WHERE target_id = $1 AND target_type = 'post'
	ORDER BY created_at DESC
	`

	rows, err := h.db.QueryContext(c.Request.Context(), query, postUUID)
	if err != nil {
		log.Error().Err(err).Str("post_id", postID).Msg("Failed to query likes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch likes"})
		return
	}
	defer rows.Close()

	var likes []LikeResponse
	for rows.Next() {
		var like LikeResponse

		err := rows.Scan(
			&like.ID, &like.UserID, &like.TargetType, &like.TargetID, &like.CreatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan like")
			continue
		}

		likes = append(likes, like)
	}

	c.JSON(http.StatusOK, gin.H{"likes": likes})
}

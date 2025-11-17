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

type FeedPostResponse struct {
	ID            string    `json:"id"`
	PostID        string    `json:"post_id"`
	UserID        string    `json:"user_id"`
	VideoID       *string   `json:"video_id"`
	Visibility    string    `json:"visibility"`
	MovementLabel string    `json:"movement_label"`
	Weight        *struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"weight"`
	RPE           *float64  `json:"rpe"`
	CommentText   string    `json:"comment_text"`
	CommentsCount int       `json:"comments_count"`
	LikesCount    int       `json:"likes_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *FeedHandlers) GetFeed(c *gin.Context) {
	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 20
		}
	}

	cursor := c.Query("cursor")
	var cursorTime time.Time
	if cursor != "" {
		var err error
		cursorTime, err = time.Parse(time.RFC3339, cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor format"})
			return
		}
	} else {
		cursorTime = time.Now()
	}

	visibility := c.Query("visibility")
	if visibility == "" {
		visibility = "public"
	}

	query := `
	SELECT id, post_id, user_id, video_id, visibility, movement_label,
		   weight_value, weight_unit, rpe, comment_text,
		   comments_count, likes_count, created_at, updated_at
	FROM feed_posts
	WHERE visibility = $1 AND created_at < $2
	ORDER BY created_at DESC
	LIMIT $3
	`

	rows, err := h.db.QueryContext(c.Request.Context(), query, visibility, cursorTime, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query feed posts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed"})
		return
	}
	defer rows.Close()

	var posts []FeedPostResponse
	for rows.Next() {
		var post FeedPostResponse
		var videoID sql.NullString
		var weightValue sql.NullFloat64
		var weightUnit sql.NullString
		var rpe sql.NullFloat64

		err := rows.Scan(
			&post.ID, &post.PostID, &post.UserID, &videoID, &post.Visibility,
			&post.MovementLabel, &weightValue, &weightUnit, &rpe, &post.CommentText,
			&post.CommentsCount, &post.LikesCount, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan feed post")
			continue
		}

		if videoID.Valid {
			post.VideoID = &videoID.String
		}

		if weightValue.Valid && weightUnit.Valid {
			post.Weight = &struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			}{
				Value: weightValue.Float64,
				Unit:  weightUnit.String,
			}
		}

		if rpe.Valid {
			post.RPE = &rpe.Float64
		}

		posts = append(posts, post)
	}

	var nextCursor *string
	if len(posts) == limit {
		cursorStr := posts[len(posts)-1].CreatedAt.Format(time.RFC3339)
		nextCursor = &cursorStr
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

func (h *FeedHandlers) GetFeedPost(c *gin.Context) {
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
	SELECT id, post_id, user_id, video_id, visibility, movement_label,
		   weight_value, weight_unit, rpe, comment_text,
		   comments_count, likes_count, created_at, updated_at
	FROM feed_posts
	WHERE post_id = $1
	`

	var post FeedPostResponse
	var videoID sql.NullString
	var weightValue sql.NullFloat64
	var weightUnit sql.NullString
	var rpe sql.NullFloat64

	err = h.db.QueryRowContext(c.Request.Context(), query, postUUID).Scan(
		&post.ID, &post.PostID, &post.UserID, &videoID, &post.Visibility,
		&post.MovementLabel, &weightValue, &weightUnit, &rpe, &post.CommentText,
		&post.CommentsCount, &post.LikesCount, &post.CreatedAt, &post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if err != nil {
		log.Error().Err(err).Str("post_id", postID).Msg("Failed to query feed post")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	if videoID.Valid {
		post.VideoID = &videoID.String
	}

	if weightValue.Valid && weightUnit.Valid {
		post.Weight = &struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		}{
			Value: weightValue.Float64,
			Unit:  weightUnit.String,
		}
	}

	if rpe.Valid {
		post.RPE = &rpe.Float64
	}

	c.JSON(http.StatusOK, post)
}

func (h *FeedHandlers) GetPrivacySettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"is_public":      true,
		"access_token":   "",
		"require_token":  false,
	})
}

func (h *FeedHandlers) UpdatePrivacySettings(c *gin.Context) {
	var req struct {
		IsPublic     bool   `json:"is_public"`
		RequireToken bool   `json:"require_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken := ""
	if req.RequireToken {
		accessToken = uuid.New().String()[:8]
	}

	c.JSON(http.StatusOK, gin.H{
		"is_public":     req.IsPublic,
		"access_token":  accessToken,
		"require_token": req.RequireToken,
		"message":       "Privacy settings updated",
	})
}

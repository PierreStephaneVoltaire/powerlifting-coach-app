package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/video-service/internal/models"
)

type VideoRepository struct {
	db *sql.DB
}

func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) CreateVideo(video *models.Video) error {
	metadataJSON, _ := json.Marshal(video.Metadata)

	query := `
		INSERT INTO videos (athlete_id, filename, original_filename, file_size, content_type, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, public_share_token, created_at, updated_at`

	err := r.db.QueryRow(query, 
		video.AthleteID, video.Filename, video.OriginalFilename, 
		video.FileSize, video.ContentType, video.Status, metadataJSON,
	).Scan(&video.ID, &video.PublicShareToken, &video.CreatedAt, &video.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetVideoByID(id uuid.UUID) (*models.Video, error) {
	query := `
		SELECT id, athlete_id, filename, original_filename, file_size, content_type,
		       duration_seconds, original_url, processed_url, thumbnail_url,
		       public_share_token, status, processing_error, metadata,
		       created_at, updated_at, processed_at
		FROM videos WHERE id = $1`

	video := &models.Video{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&video.ID, &video.AthleteID, &video.Filename, &video.OriginalFilename,
		&video.FileSize, &video.ContentType, &video.DurationSeconds,
		&video.OriginalURL, &video.ProcessedURL, &video.ThumbnailURL,
		&video.PublicShareToken, &video.Status, &video.ProcessingError,
		&metadataJSON, &video.CreatedAt, &video.UpdatedAt, &video.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &video.Metadata)
	}

	return video, nil
}

func (r *VideoRepository) GetVideoByShareToken(token string) (*models.Video, error) {
	query := `
		SELECT id, athlete_id, filename, original_filename, file_size, content_type,
		       duration_seconds, original_url, processed_url, thumbnail_url,
		       public_share_token, status, processing_error, metadata,
		       created_at, updated_at, processed_at
		FROM videos WHERE public_share_token = $1 AND status = 'ready'`

	video := &models.Video{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, token).Scan(
		&video.ID, &video.AthleteID, &video.Filename, &video.OriginalFilename,
		&video.FileSize, &video.ContentType, &video.DurationSeconds,
		&video.OriginalURL, &video.ProcessedURL, &video.ThumbnailURL,
		&video.PublicShareToken, &video.Status, &video.ProcessingError,
		&metadataJSON, &video.CreatedAt, &video.UpdatedAt, &video.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &video.Metadata)
	}

	return video, nil
}

func (r *VideoRepository) UpdateVideo(video *models.Video) error {
	metadataJSON, _ := json.Marshal(video.Metadata)

	query := `
		UPDATE videos SET
			duration_seconds = $2,
			original_url = $3,
			processed_url = $4,
			thumbnail_url = $5,
			status = $6,
			processing_error = $7,
			metadata = $8,
			processed_at = $9
		WHERE id = $1`

	_, err := r.db.Exec(query,
		video.ID, video.DurationSeconds, video.OriginalURL, video.ProcessedURL,
		video.ThumbnailURL, video.Status, video.ProcessingError,
		metadataJSON, video.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetVideosByAthleteID(athleteID uuid.UUID, page, pageSize int) ([]models.Video, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := `SELECT COUNT(*) FROM videos WHERE athlete_id = $1`
	var totalCount int
	err := r.db.QueryRow(countQuery, athleteID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get videos
	query := `
		SELECT id, athlete_id, filename, original_filename, file_size, content_type,
		       duration_seconds, original_url, processed_url, thumbnail_url,
		       public_share_token, status, processing_error, metadata,
		       created_at, updated_at, processed_at
		FROM videos 
		WHERE athlete_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, athleteID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}
	defer rows.Close()

	var videos []models.Video
	for rows.Next() {
		var video models.Video
		var metadataJSON []byte

		err := rows.Scan(
			&video.ID, &video.AthleteID, &video.Filename, &video.OriginalFilename,
			&video.FileSize, &video.ContentType, &video.DurationSeconds,
			&video.OriginalURL, &video.ProcessedURL, &video.ThumbnailURL,
			&video.PublicShareToken, &video.Status, &video.ProcessingError,
			&metadataJSON, &video.CreatedAt, &video.UpdatedAt, &video.ProcessedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &video.Metadata)
		}

		videos = append(videos, video)
	}

	return videos, totalCount, nil
}

func (r *VideoRepository) CreateFormFeedback(feedback *models.FormFeedback) error {
	issuesJSON, _ := json.Marshal(feedback.Issues)

	query := `
		INSERT INTO form_feedback (video_id, feedback_text, confidence_score, issues, ai_model)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		feedback.VideoID, feedback.FeedbackText, feedback.ConfidenceScore,
		issuesJSON, feedback.AIModel,
	).Scan(&feedback.ID, &feedback.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create form feedback: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetFormFeedbackByVideoID(videoID uuid.UUID) ([]models.FormFeedback, error) {
	query := `
		SELECT id, video_id, feedback_text, confidence_score, issues, ai_model, created_at
		FROM form_feedback 
		WHERE video_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get form feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []models.FormFeedback
	for rows.Next() {
		var feedback models.FormFeedback
		var issuesJSON []byte

		err := rows.Scan(
			&feedback.ID, &feedback.VideoID, &feedback.FeedbackText,
			&feedback.ConfidenceScore, &issuesJSON, &feedback.AIModel, &feedback.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan form feedback: %w", err)
		}

		if len(issuesJSON) > 0 {
			json.Unmarshal(issuesJSON, &feedback.Issues)
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

func (r *VideoRepository) CreateVideoShare(share *models.VideoShare) error {
	query := `
		INSERT INTO video_shares (video_id, shared_by, access_level, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.db.QueryRow(query,
		share.VideoID, share.SharedBy, share.AccessLevel, share.ExpiresAt,
	).Scan(&share.ID, &share.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create video share: %w", err)
	}

	return nil
}

func (r *VideoRepository) IncrementViewCount(shareToken string) error {
	query := `
		UPDATE video_shares 
		SET view_count = view_count + 1 
		WHERE video_id = (SELECT id FROM videos WHERE public_share_token = $1)`

	_, err := r.db.Exec(query, shareToken)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

func (r *VideoRepository) DeleteVideo(id uuid.UUID) error {
	query := `DELETE FROM videos WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("video not found")
	}

	return nil
}
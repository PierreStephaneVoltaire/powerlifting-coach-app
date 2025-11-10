package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/powerlifting-coach-app/video-service/internal/models"
	"github.com/powerlifting-coach-app/video-service/internal/queue"
	"github.com/powerlifting-coach-app/video-service/internal/repository"
	"github.com/powerlifting-coach-app/video-service/internal/storage"
	"github.com/rs/zerolog/log"
)

type VideoHandlers struct {
	videoRepo    *repository.VideoRepository
	spacesClient *storage.SpacesClient
	queueClient  *queue.RabbitMQClient
	maxFileSize  int64
	allowedExts  []string
}

func NewVideoHandlers(
	videoRepo *repository.VideoRepository,
	spacesClient *storage.SpacesClient,
	queueClient *queue.RabbitMQClient,
	maxFileSize int64,
	allowedExts []string,
) *VideoHandlers {
	return &VideoHandlers{
		videoRepo:    videoRepo,
		spacesClient: spacesClient,
		queueClient:  queueClient,
		maxFileSize:  maxFileSize,
		allowedExts:  allowedExts,
	}
}

func (h *VideoHandlers) GetUploadURL(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate file size
	if req.FileSize > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds maximum allowed size"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(req.Filename))
	isAllowed := false
	for _, allowedExt := range h.allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	userUUID, _ := uuid.Parse(userID)
	
	// Create video record
	video := &models.Video{
		AthleteID:        userUUID,
		Filename:         generateUniqueFilename(req.Filename),
		OriginalFilename: req.Filename,
		FileSize:         req.FileSize,
		ContentType:      getContentTypeFromExtension(ext),
		Status:           models.VideoStatusUploading,
		Metadata:         make(map[string]interface{}),
	}

	if err := h.videoRepo.CreateVideo(video); err != nil {
		log.Error().Err(err).Msg("Failed to create video record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video record"})
		return
	}

	// Generate presigned upload URL
	key := fmt.Sprintf("originals/%s/%s", userID, video.Filename)
	uploadURL, err := h.spacesClient.GeneratePresignedUploadURL(key, video.ContentType, time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate upload URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	response := models.UploadResponse{
		VideoID:   video.ID,
		UploadURL: uploadURL,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	c.JSON(http.StatusOK, response)
}

func (h *VideoHandlers) CompleteUpload(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	userID := middleware.GetUserID(c)
	video, err := h.videoRepo.GetVideoByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// Verify ownership
	if video.AthleteID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update video status and queue for processing
	video.Status = models.VideoStatusProcessing
	originalURL := h.spacesClient.GetFileURL(fmt.Sprintf("originals/%s/%s", userID, video.Filename))
	video.OriginalURL = &originalURL

	if err := h.videoRepo.UpdateVideo(video); err != nil {
		log.Error().Err(err).Msg("Failed to update video")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
		return
	}

	// Queue video for processing
	message := models.ProcessVideoMessage{VideoID: videoID}
	if err := h.queueClient.PublishVideoProcessing(message); err != nil {
		log.Error().Err(err).Msg("Failed to queue video processing")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload completed, video queued for processing",
		"video":   video,
	})
}

func (h *VideoHandlers) GetMyVideos(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userUUID, _ := uuid.Parse(userID)
	videos, totalCount, err := h.videoRepo.GetVideosByAthleteID(userUUID, page, pageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get videos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}

	response := models.VideoListResponse{
		Videos:     videos,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, response)
}

func (h *VideoHandlers) GetVideo(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.videoRepo.GetVideoByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	userID := middleware.GetUserID(c)
	
	// Check if user has access (owner or coach with access)
	if video.AthleteID.String() != userID {
		// TODO: Check if user is a coach with access to this athlete
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get form feedback
	feedback, err := h.videoRepo.GetFormFeedbackByVideoID(videoID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get form feedback")
	}

	c.JSON(http.StatusOK, gin.H{
		"video":    video,
		"feedback": feedback,
	})
}

func (h *VideoHandlers) GetSharedVideo(c *gin.Context) {
	shareToken := c.Param("token")
	
	video, err := h.videoRepo.GetVideoByShareToken(shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found or not available"})
		return
	}

	// Increment view count
	if err := h.videoRepo.IncrementViewCount(shareToken); err != nil {
		log.Warn().Err(err).Msg("Failed to increment view count")
	}

	// Get form feedback
	feedback, err := h.videoRepo.GetFormFeedbackByVideoID(video.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get form feedback")
	}

	c.JSON(http.StatusOK, gin.H{
		"video":    video,
		"feedback": feedback,
	})
}

func (h *VideoHandlers) DeleteVideo(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	userID := middleware.GetUserID(c)
	video, err := h.videoRepo.GetVideoByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// Verify ownership
	if video.AthleteID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete from storage
	if video.OriginalURL != nil {
		key := fmt.Sprintf("originals/%s/%s", userID, video.Filename)
		if err := h.spacesClient.DeleteFile(key); err != nil {
			log.Warn().Err(err).Msg("Failed to delete original file")
		}
	}

	if video.ProcessedURL != nil {
		key := fmt.Sprintf("feed/%s/%s", userID, video.Filename)
		if err := h.spacesClient.DeleteFile(key); err != nil {
			log.Warn().Err(err).Msg("Failed to delete processed file")
		}
	}

	if video.ThumbnailURL != nil {
		key := fmt.Sprintf("thumbnails/%s/%s.jpg", userID, strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename)))
		if err := h.spacesClient.DeleteFile(key); err != nil {
			log.Warn().Err(err).Msg("Failed to delete thumbnail")
		}
	}

	// Delete from database
	if err := h.videoRepo.DeleteVideo(videoID); err != nil {
		log.Error().Err(err).Msg("Failed to delete video from database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
}

func (h *VideoHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "video-service",
	})
}

// Helper functions
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("%d_%s%s", timestamp, uuid, ext)
}

func getContentTypeFromExtension(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	default:
		return "video/mp4"
	}
}
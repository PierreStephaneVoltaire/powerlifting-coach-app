package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/notification-service/internal/config"
	"github.com/powerlifting-coach-app/notification-service/internal/models"
	"github.com/powerlifting-coach-app/notification-service/internal/notification"
	"github.com/rs/zerolog/log"
)

type Handlers struct {
	sender *notification.Sender
	config *config.Config
}

func NewHandlers(sender *notification.Sender, cfg *config.Config) *Handlers {
	return &Handlers{
		sender: sender,
		config: cfg,
	}
}

type SendNotificationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Type        string                 `json:"type" binding:"required"`
	Channel     string                 `json:"channel" binding:"required"`
	Subject     string                 `json:"subject" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	Priority    int                    `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
}

func (h *Handlers) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notification := models.NotificationMessage{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        models.NotificationType(req.Type),
		Channel:     models.NotificationChannel(req.Channel),
		Subject:     req.Subject,
		Content:     req.Content,
		Data:        req.Data,
		Priority:    req.Priority,
		ScheduledAt: req.ScheduledAt,
		CreatedAt:   time.Now(),
	}

	if req.Priority == 0 {
		notification.Priority = 3 // Default priority
	}

	if err := h.sender.SendNotification(notification); err != nil {
		log.Error().Err(err).Msg("Failed to send notification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Notification sent successfully",
		"notification_id": notification.ID,
	})
}

func (h *Handlers) GetPreferences(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: Implement database lookup for user preferences
	// For now, return default preferences
	preferences := models.UserNotificationPreferences{
		UserID:              userID,
		EmailEnabled:        true,
		PushEnabled:         true,
		SMSEnabled:          false,
		NewVideoNotifs:      true,
		FeedbackNotifs:      true,
		ProgramNotifs:       true,
		ReminderNotifs:      true,
		MarketingEmails:     false,
		WeeklyDigest:        true,
	}

	c.JSON(http.StatusOK, preferences)
}

type UpdatePreferencesRequest struct {
	EmailEnabled    *bool `json:"email_enabled"`
	PushEnabled     *bool `json:"push_enabled"`
	SMSEnabled      *bool `json:"sms_enabled"`
	NewVideoNotifs  *bool `json:"new_video_notifications"`
	FeedbackNotifs  *bool `json:"feedback_notifications"`
	ProgramNotifs   *bool `json:"program_notifications"`
	ReminderNotifs  *bool `json:"reminder_notifications"`
	MarketingEmails *bool `json:"marketing_emails"`
	WeeklyDigest    *bool `json:"weekly_digest"`
}

func (h *Handlers) UpdatePreferences(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database update for user preferences
	// For now, just return success
	log.Info().Str("user_id", userID.String()).Msg("Updated notification preferences")

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}

type NotificationHistoryItem struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Subject   string                 `json:"subject"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data"`
	Priority  int                    `json:"priority"`
	SentAt    time.Time              `json:"sent_at"`
	Status    string                 `json:"status"`
}

func (h *Handlers) GetNotificationHistory(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// TODO: Implement database lookup for notification history
	// For now, return empty history
	history := []NotificationHistoryItem{}

	log.Info().
		Str("user_id", userID.String()).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Retrieved notification history")

	c.JSON(http.StatusOK, gin.H{
		"notifications": history,
		"total":         0,
		"limit":         limit,
		"offset":        offset,
	})
}
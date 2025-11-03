package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/notification-service/internal/queue"
	"github.com/rs/zerolog/log"
)

type EventsHandler struct {
	publisher *queue.Publisher
}

func NewEventsHandler(publisher *queue.Publisher) *EventsHandler {
	return &EventsHandler{
		publisher: publisher,
	}
}

type PublishEventRequest struct {
	SchemaVersion      string                 `json:"schema_version" binding:"required"`
	EventType          string                 `json:"event_type" binding:"required"`
	ClientGeneratedID  string                 `json:"client_generated_id" binding:"required"`
	UserID             string                 `json:"user_id" binding:"required"`
	Timestamp          *time.Time             `json:"timestamp"`
	SourceService      string                 `json:"source_service" binding:"required"`
	Data               map[string]interface{} `json:"data" binding:"required"`
}

type EventPayload struct {
	SchemaVersion      string                 `json:"schema_version"`
	EventType          string                 `json:"event_type"`
	ClientGeneratedID  string                 `json:"client_generated_id"`
	UserID             string                 `json:"user_id"`
	Timestamp          string                 `json:"timestamp"`
	SourceService      string                 `json:"source_service"`
	Data               map[string]interface{} `json:"data"`
}

func (h *EventsHandler) PublishEvent(c *gin.Context) {
	var req PublishEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("Invalid event request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := uuid.Parse(req.ClientGeneratedID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_generated_id must be a valid UUID"})
		return
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be a valid UUID"})
		return
	}

	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}

	payload := EventPayload{
		SchemaVersion:     req.SchemaVersion,
		EventType:         req.EventType,
		ClientGeneratedID: req.ClientGeneratedID,
		UserID:            req.UserID,
		Timestamp:         timestamp.Format(time.RFC3339),
		SourceService:     req.SourceService,
		Data:              req.Data,
	}

	if err := h.publisher.PublishEvent(req.EventType, payload); err != nil {
		log.Error().
			Err(err).
			Str("event_type", req.EventType).
			Str("client_generated_id", req.ClientGeneratedID).
			Msg("Failed to publish event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	log.Info().
		Str("event_type", req.EventType).
		Str("client_generated_id", req.ClientGeneratedID).
		Str("user_id", req.UserID).
		Msg("Event published successfully")

	c.JSON(http.StatusAccepted, gin.H{
		"message":             "Event published successfully",
		"event_type":          req.EventType,
		"client_generated_id": req.ClientGeneratedID,
	})
}

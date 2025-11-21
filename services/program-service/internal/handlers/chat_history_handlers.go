package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/rs/zerolog/log"
)

type ChatHistoryHandlers struct {
	db *sql.DB
}

func NewChatHistoryHandlers(db *sql.DB) *ChatHistoryHandlers {
	return &ChatHistoryHandlers{
		db: db,
	}
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type SaveMessagesRequest struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatHistoryResponse struct {
	ConversationID string        `json:"conversation_id"`
	Messages       []ChatMessage `json:"messages"`
}

func (h *ChatHistoryHandlers) GetChatHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var conversationID uuid.UUID
	err = h.db.QueryRow(`
		SELECT conversation_id
		FROM chat_conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, userUUID).Scan(&conversationID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, ChatHistoryResponse{
			ConversationID: "",
			Messages:       []ChatMessage{},
		})
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get conversation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation"})
		return
	}

	rows, err := h.db.Query(`
		SELECT message_id, role, content, created_at
		FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`, conversationID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get messages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan message")
			continue
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, ChatHistoryResponse{
		ConversationID: conversationID.String(),
		Messages:       messages,
	})
}

func (h *ChatHistoryHandlers) SaveMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req SaveMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save messages"})
		return
	}
	defer tx.Rollback()

	var conversationID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO chat_conversations (user_id, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
		RETURNING conversation_id
	`, userUUID).Scan(&conversationID)

	if err != nil {
		var existingConvID uuid.UUID
		err = tx.QueryRow(`
			SELECT conversation_id
			FROM chat_conversations
			WHERE user_id = $1
		`, userUUID).Scan(&existingConvID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to get or create conversation")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save messages"})
			return
		}
		conversationID = existingConvID
	}

	for _, msg := range req.Messages {
		var msgID uuid.UUID
		if msg.ID != "" {
			msgID, err = uuid.Parse(msg.ID)
			if err != nil {
				msgID = uuid.New()
			}
		} else {
			msgID = uuid.New()
		}

		_, err = tx.Exec(`
			INSERT INTO chat_messages (message_id, conversation_id, role, content, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (message_id) DO NOTHING
		`, msgID, conversationID, msg.Role, msg.Content, msg.CreatedAt)

		if err != nil {
			log.Error().Err(err).Msg("Failed to insert message")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save messages"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": conversationID.String(),
		"saved_count":     len(req.Messages),
	})
}

func (h *ChatHistoryHandlers) ClearHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	_, err = h.db.Exec(`
		DELETE FROM chat_conversations
		WHERE user_id = $1
	`, userUUID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to clear history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat history cleared"})
}

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powerlifting-coach-app/program-service/internal/config"
	"github.com/rs/zerolog/log"
)

type OpenAICompatHandlers struct {
	cfg *config.Config
}

func NewOpenAICompatHandlers(cfg *config.Config) *OpenAICompatHandlers {
	return &OpenAICompatHandlers{
		cfg: cfg,
	}
}

func (h *OpenAICompatHandlers) ChatCompletions(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
		return
	}
	defer c.Request.Body.Close()

	litellmURL := fmt.Sprintf("%s/v1/chat/completions", h.cfg.LiteLLMEndpoint)
	req, err := http.NewRequest(http.MethodPost, litellmURL, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create proxy request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to forward to LiteLLM")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to forward request to LiteLLM"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read LiteLLM response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read LiteLLM response"})
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Data(resp.StatusCode, "application/json", respBody)
}

func (h *OpenAICompatHandlers) ListModels(c *gin.Context) {
	litellmURL := fmt.Sprintf("%s/v1/models", h.cfg.LiteLLMEndpoint)
	req, err := http.NewRequest(http.MethodGet, litellmURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create proxy request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to forward to LiteLLM")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to forward request to LiteLLM"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read LiteLLM response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read LiteLLM response"})
		return
	}

	var modelsResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &modelsResponse); err != nil {
		log.Error().Err(err).Msg("Failed to parse models response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse models response"})
		return
	}

	c.JSON(resp.StatusCode, modelsResponse)
}

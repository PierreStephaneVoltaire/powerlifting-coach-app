package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/powerlifting-coach-app/llm-proxy-service/internal/config"
)

type ProxyHandler struct {
	config *config.Config
}

func NewProxyHandler(cfg *config.Config) *ProxyHandler {
	return &ProxyHandler{
		config: cfg,
	}
}

func (h *ProxyHandler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	litellmURL := fmt.Sprintf("%s/v1/chat/completions", h.config.LiteLLMEndpoint)
	req, err := http.NewRequest(http.MethodPost, litellmURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if h.config.LiteLLMAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.LiteLLMAPIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to LiteLLM: %v", err)
		http.Error(w, "Failed to forward request to LiteLLM", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading LiteLLM response: %v", err)
		http.Error(w, "Failed to read LiteLLM response", http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (h *ProxyHandler) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	litellmURL := fmt.Sprintf("%s/v1/models", h.config.LiteLLMEndpoint)
	req, err := http.NewRequest(http.MethodGet, litellmURL, nil)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	if h.config.LiteLLMAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.LiteLLMAPIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to LiteLLM: %v", err)
		http.Error(w, "Failed to forward request to LiteLLM", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading LiteLLM response: %v", err)
		http.Error(w, "Failed to read LiteLLM response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (h *ProxyHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "llm-proxy-service",
	})
}

func (h *ProxyHandler) HandleGenericProxy(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1")
	if path == "" || path == "/" {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	litellmURL := fmt.Sprintf("%s/v1%s", h.config.LiteLLMEndpoint, path)
	req, err := http.NewRequest(r.Method, litellmURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if h.config.LiteLLMAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.LiteLLMAPIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to LiteLLM: %v", err)
		http.Error(w, "Failed to forward request to LiteLLM", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading LiteLLM response: %v", err)
		http.Error(w, "Failed to read LiteLLM response", http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

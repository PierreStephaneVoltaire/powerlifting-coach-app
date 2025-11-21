package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/powerlifting-coach-app/llm-proxy-service/internal/config"
	"github.com/powerlifting-coach-app/llm-proxy-service/internal/handler"
	"github.com/powerlifting-coach-app/llm-proxy-service/internal/middleware"
	"github.com/rs/cors"
)

func main() {
	cfg := config.Load()

	proxyHandler := handler.NewProxyHandler(cfg)

	r := mux.NewRouter()
	r.Use(middleware.Logging)

	r.HandleFunc("/health", proxyHandler.HandleHealth).Methods("GET")
	r.HandleFunc("/v1/chat/completions", proxyHandler.HandleChatCompletions).Methods("POST")
	r.HandleFunc("/v1/models", proxyHandler.HandleModels).Methods("GET")
	r.PathPrefix("/v1/").HandlerFunc(proxyHandler.HandleGenericProxy)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	log.Printf("LLM Proxy Service starting on port %s", cfg.Port)
	log.Printf("Proxying to LiteLLM at: %s", cfg.LiteLLMEndpoint)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

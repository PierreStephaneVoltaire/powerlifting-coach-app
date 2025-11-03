package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/powerlifting-coach-app/notification-service/internal/config"
	"github.com/powerlifting-coach-app/notification-service/internal/handlers"
	"github.com/powerlifting-coach-app/notification-service/internal/notification"
	"github.com/powerlifting-coach-app/notification-service/internal/queue"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Set up structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.Environment == "development" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zlog.Info().Msg("Starting notification service")

	// Initialize notification sender
	sender := notification.NewSender(cfg.SendGridAPIKey, cfg.SendGridFromEmail)

	// Initialize queue consumer
	consumer, err := queue.NewConsumer(cfg.RabbitMQURL, sender)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to create queue consumer")
	}
	defer consumer.Close()

	// Start consuming messages
	if err := consumer.StartConsuming(); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to start consuming messages")
	}

	// Initialize HTTP handlers
	h := handlers.NewHandlers(sender, cfg)

	// Set up Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "notification-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/notifications/send", h.SendNotification)
		v1.GET("/notifications/preferences/:user_id", h.GetPreferences)
		v1.PUT("/notifications/preferences/:user_id", h.UpdatePreferences)
		v1.GET("/notifications/history/:user_id", h.GetNotificationHistory)
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	zlog.Info().Str("port", cfg.Port).Msg("Notification service started")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zlog.Info().Msg("Shutting down notification service...")

	// The context is used to inform the server it has 30 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zlog.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	zlog.Info().Msg("Notification service stopped")
}
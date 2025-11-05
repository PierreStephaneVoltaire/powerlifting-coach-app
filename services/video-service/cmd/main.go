package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/powerlifting-coach-app/video-service/internal/config"
	"github.com/powerlifting-coach-app/video-service/internal/database"
	"github.com/powerlifting-coach-app/video-service/internal/handlers"
	"github.com/powerlifting-coach-app/video-service/internal/queue"
	"github.com/powerlifting-coach-app/video-service/internal/repository"
	"github.com/powerlifting-coach-app/video-service/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	godotenv.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENVIRONMENT") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cfg := config.Load()
	log.Info().Str("port", cfg.Port).Msg("Starting video service")

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	spacesClient, err := storage.NewSpacesClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Spaces client")
	}

	queueClient, err := queue.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer queueClient.Close()

	eventConsumer, err := queue.NewEventConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create event consumer")
	}
	defer eventConsumer.Close()

	eventPublisher, err := queue.NewRabbitMQClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create event publisher")
	}
	defer eventPublisher.Close()

	feedHandlers := handlers.NewFeedHandlers(db.DB)
	commentHandlers := handlers.NewCommentHandlers(db.DB)
	commentHandlers.SetPublisher(eventPublisher)

	eventConsumer.RegisterHandler("feed.post.created", feedHandlers.HandleFeedPostCreated)
	eventConsumer.RegisterHandler("feed.post.updated", feedHandlers.HandleFeedPostUpdated)
	eventConsumer.RegisterHandler("feed.post.deleted", feedHandlers.HandleFeedPostDeleted)
	eventConsumer.RegisterHandler("comment.created", commentHandlers.HandleCommentCreated)
	eventConsumer.RegisterHandler("interaction.liked", commentHandlers.HandleInteractionLiked)

	routingKeys := []string{
		"feed.post.created",
		"feed.post.updated",
		"feed.post.deleted",
		"comment.created",
		"interaction.liked",
	}

	if err := eventConsumer.StartConsuming("video-service.events", routingKeys); err != nil {
		log.Fatal().Err(err).Msg("Failed to start event consumer")
	}

	videoRepo := repository.NewVideoRepository(db.DB)
	videoHandlers := handlers.NewVideoHandlers(
		videoRepo, spacesClient, queueClient,
		cfg.MaxFileSize, cfg.AllowedExtensions,
	)

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	authConfig := middleware.AuthConfig{
		AuthService:  cfg.AuthService,
		JWTSecret:    cfg.JWTSecret,
		SkipPaths:    []string{"/health", "/api/v1/videos/shared/"},
	}

	v1 := router.Group("/api/v1")
	{
		videos := v1.Group("/videos")
		{
			// Public endpoints
			videos.GET("/shared/:token", videoHandlers.GetSharedVideo)

			// Protected endpoints
			videos.Use(middleware.AuthMiddleware(authConfig))
			{
				videos.POST("/upload", videoHandlers.GetUploadURL)
				videos.POST("/:id/complete", videoHandlers.CompleteUpload)
				videos.GET("/", videoHandlers.GetMyVideos)
				videos.GET("/:id", videoHandlers.GetVideo)
				videos.DELETE("/:id", videoHandlers.DeleteVideo)
			}
		}

		feed := v1.Group("/feed")
		{
			feed.Use(middleware.AuthMiddleware(authConfig))
			{
				feed.GET("/", feedHandlers.GetFeed)
				feed.GET("/:post_id", feedHandlers.GetFeedPost)
			}
		}

		posts := v1.Group("/posts")
		{
			posts.Use(middleware.AuthMiddleware(authConfig))
			{
				posts.GET("/:post_id/comments", commentHandlers.GetPostComments)
				posts.GET("/:post_id/likes", commentHandlers.GetPostLikes)
			}
		}
	}

	router.GET("/health", videoHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Video service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
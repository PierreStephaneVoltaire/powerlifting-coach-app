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
	"github.com/powerlifting-coach-app/coach-service/internal/config"
	"github.com/powerlifting-coach-app/coach-service/internal/database"
	"github.com/powerlifting-coach-app/coach-service/internal/handlers"
	"github.com/powerlifting-coach-app/coach-service/internal/repository"
	"github.com/powerlifting-coach-app/shared/middleware"
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
	log.Info().Str("port", cfg.Port).Msg("Starting coach service")

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	coachRepo := repository.NewCoachRepository(db.DB)
	coachHandlers := handlers.NewCoachHandlers(coachRepo)

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
		SkipPaths:    []string{"/health"},
	}

	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(authConfig))
	{
		// Coach endpoints
		coaches := v1.Group("/coaches")
		{
			coaches.GET("/dashboard", coachHandlers.GetDashboard)
			coaches.GET("/notifications", coachHandlers.GetNotifications)
			coaches.PUT("/notifications/:id/read", coachHandlers.MarkNotificationRead)
			
			// Feedback management
			coaches.POST("/feedback", coachHandlers.CreateFeedback)
			coaches.GET("/feedback", coachHandlers.GetMyFeedback)
			coaches.GET("/feedback/:id", coachHandlers.GetFeedback)
			coaches.PUT("/feedback/:id", coachHandlers.UpdateFeedback)
			
			// Notes management
			coaches.POST("/notes", coachHandlers.CreateNote)
			coaches.GET("/athletes/:athlete_id/notes", coachHandlers.GetAthleteNotes)
			
			// Progress tracking
			coaches.POST("/progress", coachHandlers.TrackProgress)
			coaches.GET("/athletes/:athlete_id/progress", coachHandlers.GetAthleteProgress)
		}

		// Athlete endpoints for feedback
		athletes := v1.Group("/athletes")
		{
			athletes.GET("/feedback", coachHandlers.GetMyFeedbackAsAthlete)
			athletes.GET("/feedback/:id", coachHandlers.GetFeedback)
			athletes.POST("/feedback/:id/respond", coachHandlers.RespondToFeedback)
		}
	}

	router.GET("/health", coachHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Coach service started")

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
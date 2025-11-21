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
	"github.com/powerlifting-coach-app/program-service/internal/ai"
	"github.com/powerlifting-coach-app/program-service/internal/clients"
	"github.com/powerlifting-coach-app/program-service/internal/config"
	"github.com/powerlifting-coach-app/program-service/internal/database"
	"github.com/powerlifting-coach-app/program-service/internal/excel"
	"github.com/powerlifting-coach-app/program-service/internal/handlers"
	"github.com/powerlifting-coach-app/program-service/internal/repository"
	"github.com/powerlifting-coach-app/program-service/internal/services"
	"github.com/powerlifting-coach-app/program-service/internal/queue"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
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
	log.Info().Str("port", cfg.Port).Msg("Starting program service")

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	eventConsumer, err := queue.NewEventConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create event consumer")
	}
	defer eventConsumer.Close()

	programEventHandlers := handlers.NewProgramEventHandlers(db.DB, eventConsumer)

	eventConsumer.RegisterHandler("program.plan.created", programEventHandlers.HandleProgramPlanCreated)
	eventConsumer.RegisterHandler("program.plan.updated", programEventHandlers.HandleProgramPlanUpdated)
	eventConsumer.RegisterHandler("workout.started", programEventHandlers.HandleWorkoutStarted)
	eventConsumer.RegisterHandler("workout.completed", programEventHandlers.HandleWorkoutCompleted)

	routingKeys := []string{
		"program.plan.created",
		"program.plan.updated",
		"workout.started",
		"workout.completed",
	}

	if err := eventConsumer.StartConsuming("program-service.events", routingKeys); err != nil {
		log.Fatal().Err(err).Msg("Failed to start event consumer")
	}

	programRepo := repository.NewProgramRepository(db.DB)
	aiClient := ai.NewLiteLLMClient(cfg)
	excelExporter := excel.NewExcelExporter()
	workoutGenerator := services.NewWorkoutGenerator(programRepo)
	settingsClient := clients.NewSettingsClient(cfg.SettingsService)
	coachClient := clients.NewCoachClient(cfg.CoachService)

	programHandlers := handlers.NewProgramHandlers(programRepo, aiClient, excelExporter, workoutGenerator, settingsClient, coachClient)

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
		SkipPaths:    []string{"/health", "/api/v1/programs/templates"},
	}

	v1 := router.Group("/api/v1")
	{
		programs := v1.Group("/programs")
		{
			programs.GET("/templates", programHandlers.GetProgramTemplates)

			programs.Use(middleware.AuthMiddleware(authConfig))
			{
				programs.POST("/", programHandlers.CreateProgram)
				programs.POST("/generate", programHandlers.GenerateProgram)
				programs.POST("/from-chat", programHandlers.CreateProgramFromChat)
				programs.GET("/", programHandlers.GetMyPrograms)
				programs.GET("/active", programHandlers.GetActiveProgram)
				programs.GET("/pending", programHandlers.GetPendingProgram)
				programs.GET("/:id", programHandlers.GetProgram)
				programs.POST("/:id/approve", programHandlers.ApproveProgram)
				programs.POST("/:id/reject", programHandlers.RejectProgram)
				programs.GET("/:id/changes/pending", programHandlers.GetPendingChanges)
				programs.POST("/export", programHandlers.ExportProgram)
				programs.POST("/chat", programHandlers.ChatWithAI)
				programs.GET("/chat/conversation", programHandlers.GetAIConversation)
				programs.POST("/log-workout", programHandlers.LogWorkout)

				// Program change management (git-like)
				programs.POST("/changes/propose", programHandlers.ProposeChange)
				programs.POST("/changes/:changeId/apply", programHandlers.ApplyChange)
				programs.POST("/changes/:changeId/reject", programHandlers.RejectChange)
			}
		}

		// Exercise library endpoints
		exercises := v1.Group("/exercises")
		exercises.Use(middleware.AuthMiddleware(authConfig))
		{
			exercises.GET("/library", programHandlers.GetExerciseLibrary)
			exercises.POST("/library", programHandlers.CreateExerciseLibrary)
			exercises.GET("/:exerciseName/previous", programHandlers.GetPreviousSets)
			exercises.POST("/warmups/generate", programHandlers.GenerateWarmups)
		}

		// Workout template endpoints
		templates := v1.Group("/templates")
		templates.Use(middleware.AuthMiddleware(authConfig))
		{
			templates.GET("/workouts", programHandlers.GetWorkoutTemplates)
			templates.POST("/workouts", programHandlers.CreateWorkoutTemplate)
		}

		// Analytics endpoints
		analytics := v1.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(authConfig))
		{
			analytics.POST("/volume", programHandlers.GetVolumeData)
			analytics.POST("/e1rm", programHandlers.GetE1RMData)
		}

		// Session history endpoints
		sessions := v1.Group("/sessions")
		sessions.Use(middleware.AuthMiddleware(authConfig))
		{
			sessions.GET("/history", programHandlers.GetSessionHistory)
			sessions.DELETE("/:sessionId", programHandlers.DeleteSession)
		}
	}

	router.GET("/health", programHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Program service started")

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
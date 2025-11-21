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
	"github.com/powerlifting-coach-app/machine-service/internal/config"
	"github.com/powerlifting-coach-app/machine-service/internal/database"
	"github.com/powerlifting-coach-app/machine-service/internal/handlers"
	"github.com/powerlifting-coach-app/machine-service/internal/queue"
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
	log.Info().Str("port", cfg.Port).Msg("Starting Machine service")

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

	machineEventHandlers := handlers.NewMachineEventHandlers(db.DB, eventConsumer)

	eventConsumer.RegisterHandler("machine.notes.submitted", machineEventHandlers.HandleMachineNotesSubmitted)

	routingKeys := []string{
		"machine.notes.submitted",
	}

	if err := eventConsumer.StartConsuming("machine-service.events", routingKeys); err != nil {
		log.Fatal().Err(err).Msg("Failed to start event consumer")
	}

	router := gin.Default()

	router.Use(middleware.DefaultCORSMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Machine service started")

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

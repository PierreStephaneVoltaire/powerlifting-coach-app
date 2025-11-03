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
	"github.com/powerlifting-coach-app/settings-service/internal/config"
	"github.com/powerlifting-coach-app/settings-service/internal/database"
	"github.com/powerlifting-coach-app/settings-service/internal/handlers"
	"github.com/powerlifting-coach-app/settings-service/internal/repository"
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
	log.Info().Str("port", cfg.Port).Msg("Starting settings service")

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	settingsRepo := repository.NewSettingsRepository(db.DB)
	settingsHandlers := handlers.NewSettingsHandlers(settingsRepo)

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
		SkipPaths:    []string{"/health", "/api/v1/settings/app/public"},
	}

	v1 := router.Group("/api/v1")
	{
		settings := v1.Group("/settings")
		{
			// Public endpoints
			settings.GET("/app/public", settingsHandlers.GetPublicAppSettings)
			
			// Protected endpoints
			settings.Use(middleware.AuthMiddleware(authConfig))
			{
				// User settings
				settings.GET("/user", settingsHandlers.GetUserSettings)
				settings.PUT("/user", settingsHandlers.UpdateUserSettings)
				
				// App settings
				settings.GET("/app", settingsHandlers.GetAllAppSettings)
				settings.POST("/app", settingsHandlers.CreateAppSetting)
				settings.GET("/app/:key", settingsHandlers.GetAppSetting)
				settings.PUT("/app/:key", settingsHandlers.UpdateAppSetting)
				settings.DELETE("/app/:key", settingsHandlers.DeleteAppSetting)
			}
		}
	}

	router.GET("/health", settingsHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Settings service started")

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
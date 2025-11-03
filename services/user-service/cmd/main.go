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
	"github.com/powerlifting-coach-app/shared/middleware"
	"github.com/powerlifting-coach-app/user-service/internal/config"
	"github.com/powerlifting-coach-app/user-service/internal/database"
	"github.com/powerlifting-coach-app/user-service/internal/handlers"
	"github.com/powerlifting-coach-app/user-service/internal/repository"
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
	log.Info().Str("port", cfg.Port).Msg("Starting user service")

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	if err := db.RunMigrations("./migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	userRepo := repository.NewUserRepository(db.DB)
	userHandlers := handlers.NewUserHandlers(userRepo)

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
		SkipPaths:    []string{"/health", "/api/v1/users/create"},
	}

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/create", userHandlers.CreateUser)
			
			users.Use(middleware.AuthMiddleware(authConfig))
			{
				users.GET("/profile", userHandlers.GetProfile)
				users.GET("/:id", userHandlers.GetUserByID)
				users.PUT("/athlete/profile", userHandlers.UpdateAthleteProfile)
				users.PUT("/coach/profile", userHandlers.UpdateCoachProfile)
				users.POST("/athlete/access-code", userHandlers.GenerateAccessCode)
				users.POST("/coach/grant-access", userHandlers.GrantAccess)
				users.GET("/coach/athletes", userHandlers.GetMyAthletes)
			}
		}
	}

	router.GET("/health", userHandlers.HealthCheck)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("User service started")

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
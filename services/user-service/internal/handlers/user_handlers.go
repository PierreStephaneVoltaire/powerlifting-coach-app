package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/powerlifting-coach-app/user-service/internal/models"
	"github.com/powerlifting-coach-app/user-service/internal/repository"
	"github.com/rs/zerolog/log"
)

type UserHandlers struct {
	userRepo *repository.UserRepository
}

func NewUserHandlers(userRepo *repository.UserRepository) *UserHandlers {
	return &UserHandlers{
		userRepo: userRepo,
	}
}

func (h *UserHandlers) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		KeycloakID: req.KeycloakID,
		Email:      req.Email,
		Name:       req.Name,
		UserType:   req.UserType,
	}

	if err := h.userRepo.CreateUser(user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	response := h.buildUserResponse(user)
	c.JSON(http.StatusCreated, response)
}

func (h *UserHandlers) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := h.buildUserResponse(user)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandlers) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userIDStr).Msg("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := h.buildUserResponse(user)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandlers) UpdateAthleteProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil || user.UserType != models.UserTypeAthlete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req models.UpdateAthleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userRepo.UpdateAthleteProfile(userUUID, req); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to update athlete profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	response := h.buildUserResponse(user)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandlers) UpdateCoachProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil || user.UserType != models.UserTypeCoach {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req models.UpdateCoachProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userRepo.UpdateCoachProfile(userUUID, req); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to update coach profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	response := h.buildUserResponse(user)
	c.JSON(http.StatusOK, response)
}

func (h *UserHandlers) GenerateAccessCode(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil || user.UserType != models.UserTypeAthlete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only athletes can generate access codes"})
		return
	}

	var req models.GenerateAccessCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ExpiresInWeeks != nil && (*req.ExpiresInWeeks < 0 || *req.ExpiresInWeeks > 12) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expires in weeks must be between 0 and 12"})
		return
	}

	accessCode, err := h.userRepo.GenerateAccessCode(userUUID, req.ExpiresInWeeks)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to generate access code")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_code": accessCode})
}

func (h *UserHandlers) GrantAccess(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil || user.UserType != models.UserTypeCoach {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can request access"})
		return
	}

	var req models.GrantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	athlete, err := h.userRepo.GetAthleteByAccessCode(req.AccessCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired access code"})
		return
	}

	if err := h.userRepo.GrantCoachAccess(userUUID, athlete.ID, req.AccessCode); err != nil {
		log.Error().Err(err).Str("coach_id", userID).Str("athlete_id", athlete.ID.String()).Msg("Failed to grant access")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access granted successfully",
		"athlete": h.buildUserResponse(athlete),
	})
}

func (h *UserHandlers) GetMyAthletes(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(userUUID)
	if err != nil || user.UserType != models.UserTypeCoach {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view athletes"})
		return
	}

	athletes, err := h.userRepo.GetCoachAthletes(userUUID)
	if err != nil {
		log.Error().Err(err).Str("coach_id", userID).Msg("Failed to get coach athletes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get athletes"})
		return
	}

	var response []models.UserResponse
	for _, athlete := range athletes {
		response = append(response, h.buildUserResponse(&athlete))
	}

	c.JSON(http.StatusOK, gin.H{"athletes": response})
}

func (h *UserHandlers) GetAthletePublicProfile(c *gin.Context) {
	athleteIDStr := c.Param("id")
	athleteID, err := uuid.Parse(athleteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid athlete ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(athleteID)
	if err != nil {
		log.Error().Err(err).Str("athlete_id", athleteIDStr).Msg("Failed to get athlete")
		c.JSON(http.StatusNotFound, gin.H{"error": "Athlete not found"})
		return
	}

	if user.UserType != models.UserTypeAthlete {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not an athlete"})
		return
	}

	athleteProfile, _ := h.userRepo.GetAthleteProfile(athleteID)

	squat := 0.0
	bench := 0.0
	deadlift := 0.0
	var weightKg *float64
	var experienceLevel *models.ExperienceLevel

	if athleteProfile != nil {
		if athleteProfile.SquatMaxKg != nil {
			squat = *athleteProfile.SquatMaxKg
		}
		if athleteProfile.BenchMaxKg != nil {
			bench = *athleteProfile.BenchMaxKg
		}
		if athleteProfile.DeadliftMaxKg != nil {
			deadlift = *athleteProfile.DeadliftMaxKg
		}
		weightKg = athleteProfile.WeightKg
		experienceLevel = athleteProfile.ExperienceLevel
	}

	response := gin.H{
		"id":               user.ID,
		"name":             user.Name,
		"email":            user.Email,
		"weight_kg":        weightKg,
		"experience_level": experienceLevel,
		"joined_at":        user.CreatedAt,
		"stats": gin.H{
			"current_squat_max":    squat,
			"current_bench_max":    bench,
			"current_deadlift_max": deadlift,
			"total":                squat + bench + deadlift,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "user-service",
	})
}

func (h *UserHandlers) buildUserResponse(user *models.User) models.UserResponse {
	response := models.UserResponse{
		User: *user,
	}

	if user.UserType == models.UserTypeAthlete {
		if profile, err := h.userRepo.GetAthleteProfile(user.ID); err == nil {
			response.AthleteProfile = profile
		}
	} else if user.UserType == models.UserTypeCoach {
		if profile, err := h.userRepo.GetCoachProfile(user.ID); err == nil {
			response.CoachProfile = profile
		}
	}

	return response
}
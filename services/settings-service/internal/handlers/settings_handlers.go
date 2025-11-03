package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/shared/middleware"
	"github.com/powerlifting-coach-app/settings-service/internal/models"
	"github.com/powerlifting-coach-app/settings-service/internal/repository"
	"github.com/rs/zerolog/log"
)

type SettingsHandlers struct {
	settingsRepo *repository.SettingsRepository
}

func NewSettingsHandlers(settingsRepo *repository.SettingsRepository) *SettingsHandlers {
	return &SettingsHandlers{
		settingsRepo: settingsRepo,
	}
}

func (h *SettingsHandlers) GetUserSettings(c *gin.Context) {
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

	settings, err := h.settingsRepo.GetUserSettings(userUUID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandlers) UpdateUserSettings(c *gin.Context) {
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

	var req models.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate theme if provided
	if req.Theme != nil {
		validThemes := map[string]bool{"light": true, "dark": true, "auto": true}
		if !validThemes[*req.Theme] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme"})
			return
		}
	}

	// Validate units if provided
	if req.Units != nil {
		validUnits := map[string]bool{"metric": true, "imperial": true}
		if !validUnits[*req.Units] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid units"})
			return
		}
	}

	if err := h.settingsRepo.UpdateUserSettings(userUUID, req); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to update user settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user settings"})
		return
	}

	// Return updated settings
	settings, err := h.settingsRepo.GetUserSettings(userUUID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get updated user settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Settings updated but failed to retrieve"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandlers) GetPublicAppSettings(c *gin.Context) {
	settings, err := h.settingsRepo.GetPublicAppSettings()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get public app settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get app settings"})
		return
	}

	// Convert to a map for easier frontend consumption
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		settingsMap[setting.Key] = setting.Value
	}

	c.JSON(http.StatusOK, gin.H{"settings": settingsMap})
}

func (h *SettingsHandlers) GetAllAppSettings(c *gin.Context) {
	userType := middleware.GetUserType(c)
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	settings, err := h.settingsRepo.GetAllAppSettings()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all app settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get app settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (h *SettingsHandlers) GetAppSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	setting, err := h.settingsRepo.GetAppSetting(key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to get app setting")
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	// Check if setting is public or user has admin access
	userType := middleware.GetUserType(c)
	if !setting.IsPublic && userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *SettingsHandlers) UpdateAppSetting(c *gin.Context) {
	userType := middleware.GetUserType(c)
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var req models.UpdateAppSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsRepo.UpdateAppSetting(key, req); err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to update app setting")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app setting"})
		return
	}

	// Return updated setting
	setting, err := h.settingsRepo.GetAppSetting(key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to get updated app setting")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Setting updated but failed to retrieve"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *SettingsHandlers) CreateAppSetting(c *gin.Context) {
	userType := middleware.GetUserType(c)
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var setting models.AppSetting
	if err := c.ShouldBindJSON(&setting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if setting.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	if setting.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting category is required"})
		return
	}

	if err := h.settingsRepo.CreateAppSetting(&setting); err != nil {
		log.Error().Err(err).Interface("setting", setting).Msg("Failed to create app setting")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create app setting"})
		return
	}

	c.JSON(http.StatusCreated, setting)
}

func (h *SettingsHandlers) DeleteAppSetting(c *gin.Context) {
	userType := middleware.GetUserType(c)
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	if err := h.settingsRepo.DeleteAppSetting(key); err != nil {
		log.Error().Err(err).Str("key", key).Msg("Failed to delete app setting")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete app setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App setting deleted successfully"})
}

func (h *SettingsHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "settings-service",
	})
}
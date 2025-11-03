package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/settings-service/internal/models"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetUserSettings(userID uuid.UUID) (*models.UserSettings, error) {
	query := `
		SELECT id, user_id, theme, language, timezone, units, notifications, 
		       privacy, training_preferences, created_at, updated_at
		FROM user_settings WHERE user_id = $1`

	settings := &models.UserSettings{}
	var notificationsJSON, privacyJSON, trainingPrefsJSON []byte

	err := r.db.QueryRow(query, userID).Scan(
		&settings.ID, &settings.UserID, &settings.Theme, &settings.Language,
		&settings.Timezone, &settings.Units, &notificationsJSON,
		&privacyJSON, &trainingPrefsJSON, &settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create default settings if none exist
			return r.createDefaultUserSettings(userID)
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Parse JSON fields
	if len(notificationsJSON) > 0 {
		json.Unmarshal(notificationsJSON, &settings.Notifications)
	}
	if len(privacyJSON) > 0 {
		json.Unmarshal(privacyJSON, &settings.Privacy)
	}
	if len(trainingPrefsJSON) > 0 {
		json.Unmarshal(trainingPrefsJSON, &settings.TrainingPreferences)
	}

	return settings, nil
}

func (r *SettingsRepository) createDefaultUserSettings(userID uuid.UUID) (*models.UserSettings, error) {
	defaultNotifications := map[string]interface{}{
		"email": true,
		"push":  true,
		"sms":   false,
	}
	defaultPrivacy := map[string]interface{}{
		"profile_public": false,
		"videos_public":  false,
	}
	defaultTrainingPrefs := map[string]interface{}{}

	notificationsJSON, _ := json.Marshal(defaultNotifications)
	privacyJSON, _ := json.Marshal(defaultPrivacy)
	trainingPrefsJSON, _ := json.Marshal(defaultTrainingPrefs)

	query := `
		INSERT INTO user_settings (user_id, notifications, privacy, training_preferences)
		VALUES ($1, $2, $3, $4)
		RETURNING id, theme, language, timezone, units, created_at, updated_at`

	settings := &models.UserSettings{
		UserID:              userID,
		Notifications:       defaultNotifications,
		Privacy:             defaultPrivacy,
		TrainingPreferences: defaultTrainingPrefs,
	}

	err := r.db.QueryRow(query, userID, notificationsJSON, privacyJSON, trainingPrefsJSON).Scan(
		&settings.ID, &settings.Theme, &settings.Language, &settings.Timezone,
		&settings.Units, &settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create default user settings: %w", err)
	}

	return settings, nil
}

func (r *SettingsRepository) UpdateUserSettings(userID uuid.UUID, req models.UpdateUserSettingsRequest) error {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := []interface{}{userID}
	argIndex := 2

	if req.Theme != nil {
		setParts = append(setParts, fmt.Sprintf("theme = $%d", argIndex))
		args = append(args, *req.Theme)
		argIndex++
	}
	if req.Language != nil {
		setParts = append(setParts, fmt.Sprintf("language = $%d", argIndex))
		args = append(args, *req.Language)
		argIndex++
	}
	if req.Timezone != nil {
		setParts = append(setParts, fmt.Sprintf("timezone = $%d", argIndex))
		args = append(args, *req.Timezone)
		argIndex++
	}
	if req.Units != nil {
		setParts = append(setParts, fmt.Sprintf("units = $%d", argIndex))
		args = append(args, *req.Units)
		argIndex++
	}
	if req.Notifications != nil {
		notificationsJSON, _ := json.Marshal(*req.Notifications)
		setParts = append(setParts, fmt.Sprintf("notifications = $%d", argIndex))
		args = append(args, notificationsJSON)
		argIndex++
	}
	if req.Privacy != nil {
		privacyJSON, _ := json.Marshal(*req.Privacy)
		setParts = append(setParts, fmt.Sprintf("privacy = $%d", argIndex))
		args = append(args, privacyJSON)
		argIndex++
	}
	if req.TrainingPreferences != nil {
		trainingPrefsJSON, _ := json.Marshal(*req.TrainingPreferences)
		setParts = append(setParts, fmt.Sprintf("training_preferences = $%d", argIndex))
		args = append(args, trainingPrefsJSON)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE user_settings SET %s WHERE user_id = $1", 
		strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (r *SettingsRepository) GetAppSetting(key string) (*models.AppSetting, error) {
	query := `
		SELECT id, key, value, description, category, is_public, created_at, updated_at
		FROM app_settings WHERE key = $1`

	setting := &models.AppSetting{}
	var valueJSON []byte

	err := r.db.QueryRow(query, key).Scan(
		&setting.ID, &setting.Key, &valueJSON, &setting.Description,
		&setting.Category, &setting.IsPublic, &setting.CreatedAt, &setting.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("app setting not found")
		}
		return nil, fmt.Errorf("failed to get app setting: %w", err)
	}

	if len(valueJSON) > 0 {
		json.Unmarshal(valueJSON, &setting.Value)
	}

	return setting, nil
}

func (r *SettingsRepository) GetPublicAppSettings() ([]models.AppSetting, error) {
	query := `
		SELECT id, key, value, description, category, is_public, created_at, updated_at
		FROM app_settings WHERE is_public = true ORDER BY category, key`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get public app settings: %w", err)
	}
	defer rows.Close()

	var settings []models.AppSetting
	for rows.Next() {
		var setting models.AppSetting
		var valueJSON []byte

		err := rows.Scan(
			&setting.ID, &setting.Key, &valueJSON, &setting.Description,
			&setting.Category, &setting.IsPublic, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan app setting: %w", err)
		}

		if len(valueJSON) > 0 {
			json.Unmarshal(valueJSON, &setting.Value)
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SettingsRepository) GetAllAppSettings() ([]models.AppSetting, error) {
	query := `
		SELECT id, key, value, description, category, is_public, created_at, updated_at
		FROM app_settings ORDER BY category, key`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all app settings: %w", err)
	}
	defer rows.Close()

	var settings []models.AppSetting
	for rows.Next() {
		var setting models.AppSetting
		var valueJSON []byte

		err := rows.Scan(
			&setting.ID, &setting.Key, &valueJSON, &setting.Description,
			&setting.Category, &setting.IsPublic, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan app setting: %w", err)
		}

		if len(valueJSON) > 0 {
			json.Unmarshal(valueJSON, &setting.Value)
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SettingsRepository) UpdateAppSetting(key string, req models.UpdateAppSettingRequest) error {
	valueJSON, _ := json.Marshal(req.Value)

	query := `
		UPDATE app_settings 
		SET value = $2, description = COALESCE($3, description)
		WHERE key = $1`

	result, err := r.db.Exec(query, key, valueJSON, req.Description)
	if err != nil {
		return fmt.Errorf("failed to update app setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("app setting not found")
	}

	return nil
}

func (r *SettingsRepository) CreateAppSetting(setting *models.AppSetting) error {
	valueJSON, _ := json.Marshal(setting.Value)

	query := `
		INSERT INTO app_settings (key, value, description, category, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		setting.Key, valueJSON, setting.Description, setting.Category, setting.IsPublic,
	).Scan(&setting.ID, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create app setting: %w", err)
	}

	return nil
}

func (r *SettingsRepository) DeleteAppSetting(key string) error {
	query := `DELETE FROM app_settings WHERE key = $1`

	result, err := r.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete app setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("app setting not found")
	}

	return nil
}
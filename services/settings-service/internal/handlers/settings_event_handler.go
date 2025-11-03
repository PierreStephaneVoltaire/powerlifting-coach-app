package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/powerlifting-coach-app/settings-service/internal/queue"
	"golang.org/x/crypto/bcrypt"
)

type SettingsEventHandler struct {
	db        *sql.DB
	publisher *queue.Publisher
	jwtSecret string
}

func NewSettingsEventHandler(db *sql.DB, publisher *queue.Publisher, jwtSecret string) *SettingsEventHandler {
	return &SettingsEventHandler{
		db:        db,
		publisher: publisher,
		jwtSecret: jwtSecret,
	}
}

type UserSettingsSubmittedEvent struct {
	SchemaVersion     string    `json:"schema_version"`
	EventType         string    `json:"event_type"`
	ClientGeneratedID string    `json:"client_generated_id"`
	UserID            string    `json:"user_id"`
	Timestamp         time.Time `json:"timestamp"`
	SourceService     string    `json:"source_service"`
	Data              struct {
		Weight                   *struct{ Value float64; Unit string } `json:"weight"`
		Age                      int                                   `json:"age"`
		TargetWeightClass        *string                               `json:"target_weight_class"`
		WeeksUntilComp           *int                                  `json:"weeks_until_comp"`
		SquatGoal                *struct{ Value float64; Unit string } `json:"squat_goal"`
		BenchGoal                *struct{ Value float64; Unit string } `json:"bench_goal"`
		DeadGoal                 *struct{ Value float64; Unit string } `json:"dead_goal"`
		MostImportantLift        *string                               `json:"most_important_lift"`
		LeastImportantLift       *string                               `json:"least_important_lift"`
		RecoveryRatingSquat      *int                                  `json:"recovery_rating_squat"`
		RecoveryRatingBench      *int                                  `json:"recovery_rating_bench"`
		RecoveryRatingDead       *int                                  `json:"recovery_rating_dead"`
		TrainingDaysPerWeek      int                                   `json:"training_days_per_week"`
		SessionLengthMinutes     *int                                  `json:"session_length_minutes"`
		WeightPlan               *string                               `json:"weight_plan"`
		FormIssues               []string                              `json:"form_issues"`
		Injuries                 *string                               `json:"injuries"`
		EvaluateFeasibility      *bool                                 `json:"evaluate_feasibility"`
		Federation               *string                               `json:"federation"`
		KneeSleeve               *string                               `json:"knee_sleeve"`
		DeadliftStyle            *string                               `json:"deadlift_style"`
		SquatStance              *string                               `json:"squat_stance"`
		AddPerMonth              *string                               `json:"add_per_month"`
		VolumePreference         *string                               `json:"volume_preference"`
		RecoversFromHeavyDeads   *bool                                 `json:"recovers_from_heavy_deads"`
		Height                   *struct{ Value float64; Unit string } `json:"height"`
		PastCompetitions         []map[string]interface{}              `json:"past_competitions"`
		FeedVisibility           *string                               `json:"feed_visibility"`
		Passcode                 *string                               `json:"passcode"`
	} `json:"data"`
}

func (h *SettingsEventHandler) HandleUserSettingsSubmitted(ctx context.Context, payload []byte) error {
	var event UserSettingsSubmittedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal user.settings.submitted event")
		return h.emitFailedEvent(ctx, "", "", "VALIDATION_ERROR", "Invalid event payload", nil)
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", event.UserID).Msg("Invalid user_id")
		return h.emitFailedEvent(ctx, event.ClientGeneratedID, event.UserID, "VALIDATION_ERROR", "Invalid user_id", nil)
	}

	clientGenID, err := uuid.Parse(event.ClientGeneratedID)
	if err != nil {
		log.Error().Err(err).Str("client_generated_id", event.ClientGeneratedID).Msg("Invalid client_generated_id")
		return h.emitFailedEvent(ctx, event.ClientGeneratedID, event.UserID, "VALIDATION_ERROR", "Invalid client_generated_id", nil)
	}

	validationErrors := h.validateSettings(&event)
	if len(validationErrors) > 0 {
		log.Warn().Interface("validation_errors", validationErrors).Msg("Settings validation failed")
		return h.emitFailedEvent(ctx, event.ClientGeneratedID, event.UserID, "VALIDATION_ERROR", "Invalid settings data", validationErrors)
	}

	var passcodeHash *string
	if event.Data.Passcode != nil && *event.Data.Passcode != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*event.Data.Passcode), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash passcode")
			return h.emitFailedEvent(ctx, event.ClientGeneratedID, event.UserID, "UNKNOWN_ERROR", "Failed to process passcode", nil)
		}
		hashStr := string(hash)
		passcodeHash = &hashStr
	}

	formIssuesJSON, _ := json.Marshal(event.Data.FormIssues)
	pastCompsJSON, _ := json.Marshal(event.Data.PastCompetitions)

	query := `
		INSERT INTO user_settings (
			user_id, weight_value, weight_unit, age, target_weight_class, weeks_until_comp,
			squat_goal_value, squat_goal_unit, bench_goal_value, bench_goal_unit,
			dead_goal_value, dead_goal_unit, most_important_lift, least_important_lift,
			recovery_rating_squat, recovery_rating_bench, recovery_rating_dead,
			training_days_per_week, session_length_minutes, weight_plan,
			form_issues, injuries, evaluate_feasibility, federation, knee_sleeve,
			deadlift_style, squat_stance, add_per_month, volume_preference,
			recovers_from_heavy_deads, height_value, height_unit, past_competitions,
			feed_visibility, passcode_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35
		)
		ON CONFLICT (user_id) DO UPDATE SET
			weight_value = EXCLUDED.weight_value,
			weight_unit = EXCLUDED.weight_unit,
			age = EXCLUDED.age,
			target_weight_class = EXCLUDED.target_weight_class,
			weeks_until_comp = EXCLUDED.weeks_until_comp,
			squat_goal_value = EXCLUDED.squat_goal_value,
			squat_goal_unit = EXCLUDED.squat_goal_unit,
			bench_goal_value = EXCLUDED.bench_goal_value,
			bench_goal_unit = EXCLUDED.bench_goal_unit,
			dead_goal_value = EXCLUDED.dead_goal_value,
			dead_goal_unit = EXCLUDED.dead_goal_unit,
			most_important_lift = EXCLUDED.most_important_lift,
			least_important_lift = EXCLUDED.least_important_lift,
			recovery_rating_squat = EXCLUDED.recovery_rating_squat,
			recovery_rating_bench = EXCLUDED.recovery_rating_bench,
			recovery_rating_dead = EXCLUDED.recovery_rating_dead,
			training_days_per_week = EXCLUDED.training_days_per_week,
			session_length_minutes = EXCLUDED.session_length_minutes,
			weight_plan = EXCLUDED.weight_plan,
			form_issues = EXCLUDED.form_issues,
			injuries = EXCLUDED.injuries,
			evaluate_feasibility = EXCLUDED.evaluate_feasibility,
			federation = EXCLUDED.federation,
			knee_sleeve = EXCLUDED.knee_sleeve,
			deadlift_style = EXCLUDED.deadlift_style,
			squat_stance = EXCLUDED.squat_stance,
			add_per_month = EXCLUDED.add_per_month,
			volume_preference = EXCLUDED.volume_preference,
			recovers_from_heavy_deads = EXCLUDED.recovers_from_heavy_deads,
			height_value = EXCLUDED.height_value,
			height_unit = EXCLUDED.height_unit,
			past_competitions = EXCLUDED.past_competitions,
			feed_visibility = EXCLUDED.feed_visibility,
			passcode_hash = EXCLUDED.passcode_hash,
			updated_at = NOW()
		RETURNING id`

	var weightValue, weightUnit interface{}
	if event.Data.Weight != nil {
		weightValue = event.Data.Weight.Value
		weightUnit = event.Data.Weight.Unit
	}

	var squatGoalValue, squatGoalUnit interface{}
	if event.Data.SquatGoal != nil {
		squatGoalValue = event.Data.SquatGoal.Value
		squatGoalUnit = event.Data.SquatGoal.Unit
	}

	var benchGoalValue, benchGoalUnit interface{}
	if event.Data.BenchGoal != nil {
		benchGoalValue = event.Data.BenchGoal.Value
		benchGoalUnit = event.Data.BenchGoal.Unit
	}

	var deadGoalValue, deadGoalUnit interface{}
	if event.Data.DeadGoal != nil {
		deadGoalValue = event.Data.DeadGoal.Value
		deadGoalUnit = event.Data.DeadGoal.Unit
	}

	var heightValue, heightUnit interface{}
	if event.Data.Height != nil {
		heightValue = event.Data.Height.Value
		heightUnit = event.Data.Height.Unit
	}

	var settingsID uuid.UUID
	err = h.db.QueryRowContext(ctx, query,
		userID, weightValue, weightUnit, event.Data.Age, event.Data.TargetWeightClass,
		event.Data.WeeksUntilComp, squatGoalValue, squatGoalUnit, benchGoalValue, benchGoalUnit,
		deadGoalValue, deadGoalUnit, event.Data.MostImportantLift, event.Data.LeastImportantLift,
		event.Data.RecoveryRatingSquat, event.Data.RecoveryRatingBench, event.Data.RecoveryRatingDead,
		event.Data.TrainingDaysPerWeek, event.Data.SessionLengthMinutes, event.Data.WeightPlan,
		formIssuesJSON, event.Data.Injuries, event.Data.EvaluateFeasibility, event.Data.Federation,
		event.Data.KneeSleeve, event.Data.DeadliftStyle, event.Data.SquatStance, event.Data.AddPerMonth,
		event.Data.VolumePreference, event.Data.RecoversFromHeavyDeads, heightValue, heightUnit,
		pastCompsJSON, event.Data.FeedVisibility, passcodeHash,
	).Scan(&settingsID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to persist user settings")
		return h.emitFailedEvent(ctx, event.ClientGeneratedID, event.UserID, "DATABASE_ERROR", "Failed to save settings", nil)
	}

	log.Info().Str("user_id", event.UserID).Str("settings_id", settingsID.String()).Msg("User settings persisted")

	return h.emitPersistedEvent(ctx, event.ClientGeneratedID, event.UserID, settingsID.String(), clientGenID.String())
}

func (h *SettingsEventHandler) validateSettings(event *UserSettingsSubmittedEvent) []map[string]string {
	var errors []map[string]string

	if event.Data.Age < 13 || event.Data.Age > 120 {
		errors = append(errors, map[string]string{
			"field":   "age",
			"message": "Age must be between 13 and 120",
		})
	}

	if event.Data.TrainingDaysPerWeek < 1 || event.Data.TrainingDaysPerWeek > 7 {
		errors = append(errors, map[string]string{
			"field":   "training_days_per_week",
			"message": "Training days per week must be between 1 and 7",
		})
	}

	if event.Data.Weight == nil {
		errors = append(errors, map[string]string{
			"field":   "weight",
			"message": "Weight is required",
		})
	}

	if event.Data.RecoveryRatingSquat != nil && (*event.Data.RecoveryRatingSquat < 1 || *event.Data.RecoveryRatingSquat > 5) {
		errors = append(errors, map[string]string{
			"field":   "recovery_rating_squat",
			"message": "Recovery rating must be between 1 and 5",
		})
	}

	if event.Data.RecoveryRatingBench != nil && (*event.Data.RecoveryRatingBench < 1 || *event.Data.RecoveryRatingBench > 5) {
		errors = append(errors, map[string]string{
			"field":   "recovery_rating_bench",
			"message": "Recovery rating must be between 1 and 5",
		})
	}

	if event.Data.RecoveryRatingDead != nil && (*event.Data.RecoveryRatingDead < 1 || *event.Data.RecoveryRatingDead > 5) {
		errors = append(errors, map[string]string{
			"field":   "recovery_rating_dead",
			"message": "Recovery rating must be between 1 and 5",
		})
	}

	if event.Data.SessionLengthMinutes != nil && (*event.Data.SessionLengthMinutes < 15 || *event.Data.SessionLengthMinutes > 300) {
		errors = append(errors, map[string]string{
			"field":   "session_length_minutes",
			"message": "Session length must be between 15 and 300 minutes",
		})
	}

	if event.Data.FeedVisibility != nil && *event.Data.FeedVisibility == "passcode" && (event.Data.Passcode == nil || len(*event.Data.Passcode) < 4) {
		errors = append(errors, map[string]string{
			"field":   "passcode",
			"message": "Passcode must be at least 4 characters when feed visibility is passcode-protected",
		})
	}

	return errors
}

func (h *SettingsEventHandler) emitPersistedEvent(ctx context.Context, clientGenID, userID, settingsID, originalEventID string) error {
	event := map[string]interface{}{
		"schema_version":      "1.0.0",
		"event_type":          "user.settings.persisted",
		"client_generated_id": clientGenID,
		"user_id":             userID,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"source_service":      "settings-service",
		"data": map[string]string{
			"settings_id":       settingsID,
			"original_event_id": originalEventID,
		},
	}

	return h.publisher.PublishEvent(ctx, "user.settings.persisted", event)
}

func (h *SettingsEventHandler) emitFailedEvent(ctx context.Context, clientGenID, userID, errorCode, errorMessage string, validationErrors []map[string]string) error {
	event := map[string]interface{}{
		"schema_version":      "1.0.0",
		"event_type":          "user.settings.failed",
		"client_generated_id": clientGenID,
		"user_id":             userID,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"source_service":      "settings-service",
		"data": map[string]interface{}{
			"error_code":        errorCode,
			"error_message":     errorMessage,
			"validation_errors": validationErrors,
			"original_event_id": clientGenID,
		},
	}

	return h.publisher.PublishEvent(ctx, "user.settings.failed", event)
}

type FeedAccessAttemptEvent struct {
	SchemaVersion     string    `json:"schema_version"`
	EventType         string    `json:"event_type"`
	ClientGeneratedID string    `json:"client_generated_id"`
	UserID            string    `json:"user_id"`
	Timestamp         time.Time `json:"timestamp"`
	SourceService     string    `json:"source_service"`
	Data              struct {
		FeedOwnerID string `json:"feed_owner_id"`
		Passcode    string `json:"passcode"`
	} `json:"data"`
}

func (h *SettingsEventHandler) HandleFeedAccessAttempt(ctx context.Context, payload []byte) error {
	var event FeedAccessAttemptEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal feed.access.attempt event")
		return h.emitAccessDeniedEvent(ctx, "", "", "", "FEED_NOT_FOUND")
	}

	feedOwnerID, err := uuid.Parse(event.Data.FeedOwnerID)
	if err != nil {
		log.Error().Err(err).Str("feed_owner_id", event.Data.FeedOwnerID).Msg("Invalid feed_owner_id")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "FEED_NOT_FOUND")
	}

	query := `
	SELECT passcode_hash, feed_visibility
	FROM user_settings
	WHERE user_id = $1
	`

	var passcodeHash sql.NullString
	var feedVisibility sql.NullString
	err = h.db.QueryRowContext(ctx, query, feedOwnerID).Scan(&passcodeHash, &feedVisibility)
	if err == sql.ErrNoRows {
		log.Warn().Str("feed_owner_id", event.Data.FeedOwnerID).Msg("Feed owner not found")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "FEED_NOT_FOUND")
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to query feed settings")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "FEED_NOT_FOUND")
	}

	if !feedVisibility.Valid || feedVisibility.String != "passcode" {
		log.Warn().Str("feed_owner_id", event.Data.FeedOwnerID).Msg("Feed is not passcode-protected")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "FEED_NOT_PROTECTED")
	}

	if !passcodeHash.Valid || passcodeHash.String == "" {
		log.Warn().Str("feed_owner_id", event.Data.FeedOwnerID).Msg("Passcode hash not set")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "INVALID_PASSCODE")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passcodeHash.String), []byte(event.Data.Passcode))
	if err != nil {
		log.Warn().Str("feed_owner_id", event.Data.FeedOwnerID).Msg("Invalid passcode")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "INVALID_PASSCODE")
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := h.generateAccessToken(event.UserID, event.Data.FeedOwnerID, expiresAt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate access token")
		return h.emitAccessDeniedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, "INVALID_PASSCODE")
	}

	log.Info().
		Str("user_id", event.UserID).
		Str("feed_owner_id", event.Data.FeedOwnerID).
		Msg("Feed access granted")

	return h.emitAccessGrantedEvent(ctx, event.ClientGeneratedID, event.UserID, event.Data.FeedOwnerID, token, expiresAt)
}

func (h *SettingsEventHandler) generateAccessToken(userID, feedOwnerID string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"user_id":       userID,
		"feed_owner_id": feedOwnerID,
		"exp":           expiresAt.Unix(),
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *SettingsEventHandler) emitAccessGrantedEvent(ctx context.Context, clientGenID, userID, feedOwnerID, accessToken string, expiresAt time.Time) error {
	event := map[string]interface{}{
		"schema_version":      "1.0.0",
		"event_type":          "feed.access.granted",
		"client_generated_id": clientGenID,
		"user_id":             userID,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"source_service":      "settings-service",
		"data": map[string]interface{}{
			"feed_owner_id": feedOwnerID,
			"access_token":  accessToken,
			"expires_at":    expiresAt.Format(time.RFC3339),
		},
	}

	return h.publisher.PublishEvent(ctx, "feed.access.granted", event)
}

func (h *SettingsEventHandler) emitAccessDeniedEvent(ctx context.Context, clientGenID, userID, feedOwnerID, reason string) error {
	event := map[string]interface{}{
		"schema_version":      "1.0.0",
		"event_type":          "feed.access.denied",
		"client_generated_id": clientGenID,
		"user_id":             userID,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"source_service":      "settings-service",
		"data": map[string]interface{}{
			"feed_owner_id": feedOwnerID,
			"reason":        reason,
		},
	}

	return h.publisher.PublishEvent(ctx, "feed.access.denied", event)
}

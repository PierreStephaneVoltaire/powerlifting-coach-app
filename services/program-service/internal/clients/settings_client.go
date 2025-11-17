package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SettingsClient struct {
	baseURL    string
	httpClient *http.Client
}

type UserSettings struct {
	UserID               string   `json:"user_id"`
	WeightValue          *float64 `json:"weight_value,omitempty"`
	WeightUnit           *string  `json:"weight_unit,omitempty"`
	Age                  *int     `json:"age,omitempty"`
	TargetWeightClass    *string  `json:"target_weight_class,omitempty"`
	CompetitionDate      *string  `json:"competition_date,omitempty"`
	SquatGoalValue       *float64 `json:"squat_goal_value,omitempty"`
	SquatGoalUnit        *string  `json:"squat_goal_unit,omitempty"`
	BenchGoalValue       *float64 `json:"bench_goal_value,omitempty"`
	BenchGoalUnit        *string  `json:"bench_goal_unit,omitempty"`
	DeadGoalValue        *float64 `json:"dead_goal_value,omitempty"`
	DeadGoalUnit         *string  `json:"dead_goal_unit,omitempty"`
	MostImportantLift    *string  `json:"most_important_lift,omitempty"`
	LeastImportantLift   *string  `json:"least_important_lift,omitempty"`
	RecoveryRatingSquat  *int     `json:"recovery_rating_squat,omitempty"`
	RecoveryRatingBench  *int     `json:"recovery_rating_bench,omitempty"`
	RecoveryRatingDead   *int     `json:"recovery_rating_dead,omitempty"`
	TrainingDaysPerWeek  *int     `json:"training_days_per_week,omitempty"`
	SessionLengthMinutes *int     `json:"session_length_minutes,omitempty"`
	WeightPlan           *string  `json:"weight_plan,omitempty"`
	Injuries             *string  `json:"injuries,omitempty"`
	KneeSleeve           *string  `json:"knee_sleeve,omitempty"`
	DeadliftStyle        *string  `json:"deadlift_style,omitempty"`
	SquatStance          *string  `json:"squat_stance,omitempty"`
	SquatBarPosition     *string  `json:"squat_bar_position,omitempty"`
	VolumePreference     *string  `json:"volume_preference,omitempty"`
	HeightValue          *float64 `json:"height_value,omitempty"`
	HeightUnit           *string  `json:"height_unit,omitempty"`
	HasCompeted          *bool    `json:"has_competed,omitempty"`
	BestSquatKg          *float64 `json:"best_squat_kg,omitempty"`
	BestBenchKg          *float64 `json:"best_bench_kg,omitempty"`
	BestDeadKg           *float64 `json:"best_dead_kg,omitempty"`
	BestTotalKg          *float64 `json:"best_total_kg,omitempty"`
	CompPrDate           *string  `json:"comp_pr_date,omitempty"`
	CompFederation       *string  `json:"comp_federation,omitempty"`
}

func NewSettingsClient(baseURL string) *SettingsClient {
	return &SettingsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *SettingsClient) GetUserSettings(ctx context.Context, authToken string) (*UserSettings, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/settings/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("settings service returned status %d", resp.StatusCode)
	}

	var settings UserSettings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &settings, nil
}

// FormatAthleteProfile converts user settings into a human-readable profile string for AI
func (s *UserSettings) FormatAthleteProfile() string {
	profile := "## Athlete Profile\n\n"

	// Physical attributes
	if s.Age != nil {
		profile += fmt.Sprintf("**Age**: %d years old\n", *s.Age)
	}

	if s.WeightValue != nil && s.WeightUnit != nil {
		profile += fmt.Sprintf("**Body Weight**: %.1f %s\n", *s.WeightValue, *s.WeightUnit)
	}

	if s.HeightValue != nil && s.HeightUnit != nil {
		profile += fmt.Sprintf("**Height**: %.1f %s\n", *s.HeightValue, *s.HeightUnit)
	}

	if s.TargetWeightClass != nil {
		profile += fmt.Sprintf("**Target Weight Class**: %s\n", *s.TargetWeightClass)
	}

	if s.WeightPlan != nil {
		profile += fmt.Sprintf("**Weight Plan**: %s\n", *s.WeightPlan)
	}

	// Current best lifts
	profile += "\n### Current Best Lifts\n"
	if s.BestSquatKg != nil {
		profile += fmt.Sprintf("- **Squat**: %.1f kg\n", *s.BestSquatKg)
	}
	if s.BestBenchKg != nil {
		profile += fmt.Sprintf("- **Bench Press**: %.1f kg\n", *s.BestBenchKg)
	}
	if s.BestDeadKg != nil {
		profile += fmt.Sprintf("- **Deadlift**: %.1f kg\n", *s.BestDeadKg)
	}
	if s.BestTotalKg != nil {
		profile += fmt.Sprintf("- **Total**: %.1f kg\n", *s.BestTotalKg)
	}

	// Goal lifts
	profile += "\n### Competition Goals\n"
	if s.SquatGoalValue != nil && s.SquatGoalUnit != nil {
		profile += fmt.Sprintf("- **Squat Goal**: %.1f %s\n", *s.SquatGoalValue, *s.SquatGoalUnit)
	}
	if s.BenchGoalValue != nil && s.BenchGoalUnit != nil {
		profile += fmt.Sprintf("- **Bench Goal**: %.1f %s\n", *s.BenchGoalValue, *s.BenchGoalUnit)
	}
	if s.DeadGoalValue != nil && s.DeadGoalUnit != nil {
		profile += fmt.Sprintf("- **Deadlift Goal**: %.1f %s\n", *s.DeadGoalValue, *s.DeadGoalUnit)
	}

	// Competition date
	if s.CompetitionDate != nil {
		profile += fmt.Sprintf("\n**Competition Date**: %s\n", *s.CompetitionDate)
	}

	// Federation
	if s.CompFederation != nil {
		profile += fmt.Sprintf("**Federation**: %s\n", *s.CompFederation)
	}

	// Competition history
	if s.HasCompeted != nil {
		if *s.HasCompeted {
			profile += "\n**Competition Experience**: Has competed before\n"
			if s.CompPrDate != nil {
				profile += fmt.Sprintf("- Last PR Date: %s\n", *s.CompPrDate)
			}
		} else {
			profile += "\n**Competition Experience**: First-time competitor\n"
		}
	}

	// Training preferences
	profile += "\n### Training Preferences\n"
	if s.TrainingDaysPerWeek != nil {
		profile += fmt.Sprintf("- **Training Days**: %d days per week\n", *s.TrainingDaysPerWeek)
	}
	if s.SessionLengthMinutes != nil {
		profile += fmt.Sprintf("- **Session Length**: %d minutes\n", *s.SessionLengthMinutes)
	}
	if s.VolumePreference != nil {
		profile += fmt.Sprintf("- **Volume Preference**: %s\n", *s.VolumePreference)
	}

	// Lift priorities
	if s.MostImportantLift != nil {
		profile += fmt.Sprintf("- **Most Important Lift**: %s\n", *s.MostImportantLift)
	}
	if s.LeastImportantLift != nil {
		profile += fmt.Sprintf("- **Least Important Lift**: %s\n", *s.LeastImportantLift)
	}

	// Recovery ratings
	profile += "\n### Recovery Ratings (1-5 scale)\n"
	if s.RecoveryRatingSquat != nil {
		profile += fmt.Sprintf("- **Squat Recovery**: %d/5\n", *s.RecoveryRatingSquat)
	}
	if s.RecoveryRatingBench != nil {
		profile += fmt.Sprintf("- **Bench Recovery**: %d/5\n", *s.RecoveryRatingBench)
	}
	if s.RecoveryRatingDead != nil {
		profile += fmt.Sprintf("- **Deadlift Recovery**: %d/5\n", *s.RecoveryRatingDead)
	}

	// Technical preferences
	profile += "\n### Technical Preferences\n"
	if s.SquatStance != nil {
		profile += fmt.Sprintf("- **Squat Stance**: %s\n", *s.SquatStance)
	}
	if s.SquatBarPosition != nil {
		profile += fmt.Sprintf("- **Squat Bar Position**: %s\n", *s.SquatBarPosition)
	}
	if s.DeadliftStyle != nil {
		profile += fmt.Sprintf("- **Deadlift Style**: %s\n", *s.DeadliftStyle)
	}

	// Equipment
	if s.KneeSleeve != nil {
		profile += fmt.Sprintf("- **Knee Sleeves**: %s\n", *s.KneeSleeve)
	}

	// Injuries and limitations
	if s.Injuries != nil && *s.Injuries != "" {
		profile += fmt.Sprintf("\n### Injuries/Limitations\n%s\n", *s.Injuries)
	}

	return profile
}

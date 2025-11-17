package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CoachClient struct {
	baseURL    string
	httpClient *http.Client
}

type CoachFeedback struct {
	ID            uuid.UUID `json:"id"`
	CoachID       uuid.UUID `json:"coach_id"`
	AthleteID     uuid.UUID `json:"athlete_id"`
	FeedbackType  string    `json:"feedback_type"`
	Priority      string    `json:"priority"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	IsIncorporated bool     `json:"is_incorporated"`
	CreatedAt     time.Time `json:"created_at"`
}

type FeedbackListResponse struct {
	Feedback   []CoachFeedback `json:"feedback"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalCount int             `json:"total_count"`
}

type CoachingAssignment struct {
	ID         uuid.UUID `json:"id"`
	CoachID    uuid.UUID `json:"coach_id"`
	AthleteID  uuid.UUID `json:"athlete_id"`
	Status     string    `json:"status"`
	StartDate  time.Time `json:"start_date"`
}

type CoachingListResponse struct {
	Assignments []CoachingAssignment `json:"assignments"`
}

func NewCoachClient(baseURL string) *CoachClient {
	return &CoachClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *CoachClient) GetAthleteFeedback(ctx context.Context, authToken string, athleteID uuid.UUID) ([]CoachFeedback, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/athletes/feedback", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coach service returned status %d", resp.StatusCode)
	}

	var feedbackResp FeedbackListResponse
	if err := json.NewDecoder(resp.Body).Decode(&feedbackResp); err != nil {
		return nil, err
	}

	return feedbackResp.Feedback, nil
}

func (c *CoachClient) GetCoachAthletes(ctx context.Context, authToken string) ([]CoachingAssignment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/coaches/athletes", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coach service returned status %d", resp.StatusCode)
	}

	var assignmentsResp CoachingListResponse
	if err := json.NewDecoder(resp.Body).Decode(&assignmentsResp); err != nil {
		return nil, err
	}

	return assignmentsResp.Assignments, nil
}

func (c *CoachClient) HasCoachAccess(ctx context.Context, authToken string, coachID, athleteID uuid.UUID) (bool, error) {
	assignments, err := c.GetCoachAthletes(ctx, authToken)
	if err != nil {
		return false, err
	}

	for _, assignment := range assignments {
		if assignment.AthleteID == athleteID && assignment.Status == "active" {
			return true, nil
		}
	}

	return false, nil
}

func FormatCoachFeedback(feedback []CoachFeedback) string {
	if len(feedback) == 0 {
		return ""
	}

	result := "## Recent Coach Feedback\n\n"
	for i, fb := range feedback {
		if i >= 5 {
			break
		}
		if fb.IsIncorporated {
			continue
		}

		priorityEmoji := ""
		switch fb.Priority {
		case "urgent":
			priorityEmoji = "🚨"
		case "high":
			priorityEmoji = "⚠️"
		case "medium":
			priorityEmoji = "📝"
		case "low":
			priorityEmoji = "💡"
		}

		result += fmt.Sprintf("### %s %s (%s)\n", priorityEmoji, fb.Title, fb.FeedbackType)
		result += fb.Content + "\n"
		result += fmt.Sprintf("*Priority: %s | Date: %s*\n\n", fb.Priority, fb.CreatedAt.Format("Jan 2, 2006"))
	}

	return result
}

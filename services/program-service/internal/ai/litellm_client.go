package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/powerlifting-coach-app/program-service/internal/config"
	"github.com/powerlifting-coach-app/program-service/internal/models"
	"github.com/rs/zerolog/log"
)

type LiteLLMClient struct {
	baseURL    string
	httpClient *http.Client
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewLiteLLMClient(cfg *config.Config) *LiteLLMClient {
	return &LiteLLMClient{
		baseURL: cfg.LiteLLMEndpoint,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *LiteLLMClient) GenerateProgram(ctx context.Context, req models.GenerateProgramRequest, athleteProfile string, coachFeedback string) (string, error) {
	systemPrompt := c.buildProgramGenerationPrompt(req, athleteProfile, coachFeedback)
	
	userPrompt := fmt.Sprintf(`Generate a %d-week powerlifting program with the following requirements:
- Experience Level: %s
- Training Days: %d per week
- Goals: %s
- Phase: Based on competition date and goals

Please provide a detailed JSON program structure that includes:
1. Program overview and periodization
2. Weekly structure with training sessions
3. Exercise selection with sets, reps, and intensity
4. Progression scheme
5. Deload weeks if applicable

Current maxes: Squat: %.1fkg, Bench: %.1fkg, Deadlift: %.1fkg`,
		req.WeeksDuration,
		req.ExperienceLevel,
		req.TrainingDays,
		req.Goals,
		getFloatValue(req.CurrentMaxes.SquatKg),
		getFloatValue(req.CurrentMaxes.BenchKg),
		getFloatValue(req.CurrentMaxes.DeadliftKg))

	if req.Injuries != nil && *req.Injuries != "" {
		userPrompt += fmt.Sprintf("\n\nInjuries/Limitations: %s", *req.Injuries)
	}

	if req.Preferences != nil && *req.Preferences != "" {
		userPrompt += fmt.Sprintf("\n\nPreferences: %s", *req.Preferences)
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := c.chatCompletion(ctx, messages, "gpt-3.5-turbo")
	if err != nil {
		return "", fmt.Errorf("failed to generate program: %w", err)
	}

	return response, nil
}

func (c *LiteLLMClient) ChatWithAI(ctx context.Context, conversation models.AIConversation, newMessage string, athleteProfile string, coachFeedback string) (string, error) {
	systemPrompt := c.buildChatSystemPrompt(athleteProfile, coachFeedback)
	
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	// Add conversation history
	for _, msg := range conversation.Messages {
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add new user message
	messages = append(messages, Message{
		Role:    "user",
		Content: newMessage,
	})

	response, err := c.chatCompletion(ctx, messages, "gpt-3.5-turbo")
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	return response, nil
}

func (c *LiteLLMClient) AnalyzeFormVideo(ctx context.Context, videoURL string, exerciseName string) (string, error) {
	systemPrompt := `You are an expert powerlifting coach analyzing form videos. Provide detailed feedback on technique, safety, and areas for improvement. Focus on the main lifts: squat, bench press, and deadlift.`
	
	userPrompt := fmt.Sprintf(`Please analyze this %s video and provide form feedback. The video is available at: %s

Please provide:
1. Overall technique assessment
2. Specific issues or areas for improvement
3. Safety concerns (if any)
4. Recommendations for correction
5. Positive aspects of the lift`, exerciseName, videoURL)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := c.chatCompletion(ctx, messages, "gpt-4-vision-preview")
	if err != nil {
		// Fallback to text-only model if vision model fails
		log.Warn().Err(err).Msg("Vision model failed, falling back to text analysis")
		
		textPrompt := fmt.Sprintf(`Provide general form tips and common issues for the %s exercise. Since I cannot analyze the specific video, provide comprehensive guidance that would be helpful for most lifters.`, exerciseName)
		
		messages[1].Content = textPrompt
		response, err = c.chatCompletion(ctx, messages, "gpt-3.5-turbo")
		if err != nil {
			return "", fmt.Errorf("failed to analyze form: %w", err)
		}
	}

	return response, nil
}

func (c *LiteLLMClient) chatCompletion(ctx context.Context, messages []Message, model string) (string, error) {
	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *LiteLLMClient) buildProgramGenerationPrompt(req models.GenerateProgramRequest, athleteProfile, coachFeedback string) string {
	prompt := `You are an expert powerlifting coach with decades of experience in program design. You create effective, science-based training programs tailored to individual athletes.

Key principles to follow:
1. Progressive overload with appropriate periodization
2. Specificity to powerlifting (squat, bench, deadlift focus)
3. Adequate recovery and deload weeks
4. Individual customization based on experience level
5. Safety and injury prevention

Consider the athlete's profile, current abilities, and any coach feedback when designing the program.

Athlete Profile:
` + athleteProfile

	if coachFeedback != "" {
		prompt += "\n\nCoach Feedback to Incorporate:\n" + coachFeedback
	}

	prompt += `

Provide the response as a structured JSON object that can be easily parsed and implemented. Include detailed explanations for your programming choices.`

	return prompt
}

func (c *LiteLLMClient) buildChatSystemPrompt(athleteProfile, coachFeedback string) string {
	prompt := `You are an expert powerlifting coach and training assistant. You help athletes with:
- Training program questions and modifications
- Exercise technique and form advice
- Nutrition and recovery guidance
- Competition preparation
- Motivation and mindset coaching

Always provide evidence-based advice and prioritize safety. Be encouraging but realistic about expectations and timelines.

Athlete Profile:
` + athleteProfile

	if coachFeedback != "" {
		prompt += "\n\nRecent Coach Feedback:\n" + coachFeedback
	}

	return prompt
}

func getFloatValue(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
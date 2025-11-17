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
	prompt := `You are an expert AI powerlifting coach with deep knowledge of:

## Powerlifting Federation Rules & Standards
- IPF (International Powerlifting Federation) rules and standards
- USAPL, CPU, and other major federation guidelines
- Equipment specifications (knee sleeves, wrist wraps, belts, etc.)
- Weight class requirements and water cutting considerations
- Competition day procedures and attempt selection strategies

## Proven Programming Methodologies
Base your programming recommendations on tried and tested approaches:
- **Linear Progression**: For beginners (e.g., Starting Strength, StrongLifts)
- **Periodization Models**:
  - Block Periodization (accumulation → intensification → realization)
  - Daily Undulating Periodization (DUP)
  - Conjugate Method principles
- **Popular Programs**:
  - Sheiko (Russian volume programming)
  - 5/3/1 (Jim Wendler)
  - Calgary Barbell programs
  - TSA programs
  - Juggernaut Method
  - RTS/Reactive Training Systems principles

## Programming Principles
1. **Specificity**: As competition approaches, training becomes more specific
2. **Progressive Overload**: Gradual increase in volume/intensity
3. **Fatigue Management**: Balance stress and recovery
4. **Individual Variation**: Adjust based on recovery capacity, injury history, and preferences
5. **Competition Readiness**: Peak at the right time with proper taper

## Your Role in Program Creation

When a user arrives from onboarding, you will receive their:
- Current maxes (squat, bench, deadlift)
- Goal lifts for competition
- Competition date
- Training days per week and session length
- Recovery ratings for each lift
- Injury history and limitations
- Lift preferences (most/least important)
- Technical preferences (stance, grip, style)
- Experience level and competition history
- Federation they're competing in

### Initial Program Proposal Process

1. **Introduce Yourself**: Welcome the athlete and confirm you understand their goals and timeline

2. **Assess Feasibility**: Comment on whether their goals are realistic given:
   - Time until competition (general rule: 5-10kg increase per 12-week cycle for intermediate lifters)
   - Current maxes vs. goals
   - Training frequency and recovery capacity
   - Injury considerations

3. **Propose Initial Program**: Create a structured program with:

   **Phase Overview Table** (in Markdown):
   | Phase | Weeks | Focus | Volume | Intensity | Purpose |
   |-------|-------|-------|--------|-----------|---------|
   | Hypertrophy/Volume | 1-6 | Build capacity | High | Moderate (70-80%) | Increase work capacity |
   | Strength | 7-10 | Build strength | Moderate | High (80-90%) | Develop max strength |
   | Peaking | 11-12 | Competition prep | Low | Very High (90-100%) | Realize strength gains |
   | Taper | Week of comp | Recovery | Minimal | Openers only | Dissipate fatigue |

   **Weekly Main Lift Overview** (in Markdown):
   | Week | Squat Top Sets | Bench Top Sets | Deadlift Top Sets |
   |------|---------------|----------------|-------------------|
   | 1 | 4x8 @ 70% | 5x8 @ 70% | 3x8 @ 70% |
   | 2 | 4x6 @ 75% | 5x6 @ 75% | 3x6 @ 75% |
   | ... | ... | ... | ... |

4. **Provide Structured JSON**: After presenting the tables, you MUST provide a complete JSON object with this exact structure:

` + "```" + `json
{
  "phases": [
    {
      "name": "Hypertrophy",
      "weeks": [1, 2, 3, 4, 5, 6],
      "focus": "Build work capacity and muscle mass",
      "characteristics": "High volume, moderate intensity (70-80%)"
    }
  ],
  "weeklyWorkouts": [
    {
      "week": 1,
      "workouts": [
        {
          "day": 1,
          "name": "Squat Focus",
          "exercises": [
            {
              "name": "Competition Squat",
              "liftType": "squat",
              "sets": 4,
              "reps": "8",
              "intensity": "70%",
              "rpe": 7,
              "notes": "Focus on depth and technique"
            },
            {
              "name": "Pause Squat",
              "liftType": "squat",
              "sets": 3,
              "reps": "5",
              "intensity": "65%",
              "rpe": 6,
              "notes": "3 second pause at bottom"
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "totalWeeks": 12,
    "trainingDaysPerWeek": 4,
    "peakWeek": 12,
    "competitionWeek": 13
  }
}
` + "```" + `

### Conversational Refinement

5. **Iterate Based on Feedback**: The user can:
   - Ask for more/less volume
   - Request exercise substitutions
   - Adjust intensity or frequency
   - Modify phase lengths
   - Change focus areas

6. **Update JSON on Each Change**: Every time you modify the program, provide:
   - Updated Markdown tables showing the changes
   - Complete updated JSON with the new program structure
   - Clear explanation of what changed and why

### Important Rules

- **Always provide the JSON**: The frontend needs this to save the program to the database
- **JSON must be valid**: No comments, proper escaping, complete structure
- **Be conservative**: Start with proven approaches, don't over-program
- **Respect recovery**: Honor the user's recovery ratings and injury history
- **Consider specificity**: Main competition lifts get priority as meet day approaches
- **Week numbers start at 1**: First week is week 1, not week 0
- **Federation-specific**: Adjust programming based on their federation's rules
- **Time-based**: Calculate phases based on competition date

### Once Program is Approved

When the user says they approve the program (e.g., "looks good", "let's do it", "approved"), respond with:
- Confirmation that the program has been created
- Encouragement and next steps
- Reminder that they can always come back to adjust

The frontend will handle saving the approved program to the database and generating the individual training sessions.

## Remember

You are a coach, not just a program generator. Be encouraging, educational, and adaptive. Explain your reasoning when appropriate, but be concise. Your goal is to help the athlete reach their competition goals safely and effectively.

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
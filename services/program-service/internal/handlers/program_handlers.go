package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/program-service/internal/ai"
	"github.com/powerlifting-coach-app/program-service/internal/excel"
	"github.com/powerlifting-coach-app/program-service/internal/models"
	"github.com/powerlifting-coach-app/program-service/internal/repository"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/rs/zerolog/log"
)

type ProgramHandlers struct {
	programRepo   *repository.ProgramRepository
	aiClient      *ai.LiteLLMClient
	excelExporter *excel.ExcelExporter
}

func NewProgramHandlers(
	programRepo *repository.ProgramRepository,
	aiClient *ai.LiteLLMClient,
	excelExporter *excel.ExcelExporter,
) *ProgramHandlers {
	return &ProgramHandlers{
		programRepo:   programRepo,
		aiClient:      aiClient,
		excelExporter: excelExporter,
	}
}

func (h *ProgramHandlers) CreateProgram(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, _ := uuid.Parse(userID)
	endDate := req.StartDate.AddDate(0, 0, req.WeeksTotal*7)

	program := &models.Program{
		AthleteID:    userUUID,
		Name:         req.Name,
		Description:  req.Description,
		Phase:        req.Phase,
		StartDate:    req.StartDate,
		EndDate:      endDate,
		WeeksTotal:   req.WeeksTotal,
		DaysPerWeek:  req.DaysPerWeek,
		ProgramData:  req.ProgramData,
		AIGenerated:  false,
		IsActive:     true,
	}

	if err := h.programRepo.CreateProgram(program); err != nil {
		log.Error().Err(err).Msg("Failed to create program")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create program"})
		return
	}

	c.JSON(http.StatusCreated, program)
}

func (h *ProgramHandlers) GenerateProgram(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.GenerateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, _ := uuid.Parse(userID)

	// Get athlete profile for context
	athleteProfile := h.getAthleteProfileString(userUUID)
	
	// Get coach feedback if enabled
	var coachFeedback string
	if req.CoachContextEnable {
		coachFeedback = h.getCoachFeedbackString(userUUID)
	}

	// Generate program using AI
	programJSON, err := h.aiClient.GenerateProgram(c.Request.Context(), req, athleteProfile, coachFeedback)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate program with AI")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate program"})
		return
	}

	// Parse AI response and create program
	var programData map[string]interface{}
	if err := json.Unmarshal([]byte(programJSON), &programData); err != nil {
		log.Error().Err(err).Msg("Failed to parse AI program response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse generated program"})
		return
	}

	// Determine program phase based on competition date
	phase := h.determineProgramPhase(req.CompetitionDate)

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, req.WeeksDuration*7)

	program := &models.Program{
		AthleteID:    userUUID,
		Name:         fmt.Sprintf("AI Generated %s Program", req.ExperienceLevel),
		Description:  &req.Goals,
		Phase:        phase,
		StartDate:    startDate,
		EndDate:      endDate,
		WeeksTotal:   req.WeeksDuration,
		DaysPerWeek:  req.TrainingDays,
		ProgramData:  programData,
		AIGenerated:  true,
		AIModel:      stringPtr("gpt-3.5-turbo"),
		AIPrompt:     &req.Goals,
		IsActive:     true,
	}

	if err := h.programRepo.CreateProgram(program); err != nil {
		log.Error().Err(err).Msg("Failed to save generated program")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save program"})
		return
	}

	// Create AI conversation record
	conversation := &models.AIConversation{
		AthleteID:           userUUID,
		ProgramID:           &program.ID,
		ConversationType:    "program_generation",
		Messages: []models.Message{
			{
				ID:        uuid.New().String(),
				Role:      "user",
				Content:   req.Goals,
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New().String(),
				Role:      "assistant",
				Content:   programJSON,
				Timestamp: time.Now(),
			},
		},
		CoachContextEnabled: req.CoachContextEnable,
	}

	if err := h.programRepo.CreateAIConversation(conversation); err != nil {
		log.Warn().Err(err).Msg("Failed to save AI conversation")
	}

	c.JSON(http.StatusCreated, gin.H{
		"program":      program,
		"ai_response":  programJSON,
		"conversation": conversation,
	})
}

func (h *ProgramHandlers) GetMyPrograms(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, _ := uuid.Parse(userID)
	programs, err := h.programRepo.GetProgramsByAthleteID(userUUID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get programs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get programs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"programs": programs})
}

func (h *ProgramHandlers) GetProgram(c *gin.Context) {
	programIDStr := c.Param("id")
	programID, err := uuid.Parse(programIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}

	userID := middleware.GetUserID(c)
	program, err := h.programRepo.GetProgramByID(programID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get program")
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}

	// Check access permissions
	if !h.hasAccessToProgram(userID, program) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get training sessions
	sessions, err := h.programRepo.GetSessionsByProgramID(programID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get training sessions")
		sessions = []models.TrainingSession{} // Continue with empty sessions
	}

	response := models.ProgramResponse{
		Program:  *program,
		Sessions: sessions,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProgramHandlers) ChatWithAI(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, _ := uuid.Parse(userID)

	// Get or create conversation
	conversations, err := h.programRepo.GetAIConversationsByAthleteID(userUUID, 1)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get AI conversations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversation"})
		return
	}

	var conversation models.AIConversation
	if len(conversations) > 0 && conversations[0].ProgramID != nil && req.ProgramID != nil && *conversations[0].ProgramID == *req.ProgramID {
		conversation = conversations[0]
	} else {
		// Create new conversation
		conversation = models.AIConversation{
			AthleteID:           userUUID,
			ProgramID:           req.ProgramID,
			ConversationType:    "chat",
			Messages:            []models.Message{},
			CoachContextEnabled: req.CoachContextEnable,
		}
	}

	// Get context
	athleteProfile := h.getAthleteProfileString(userUUID)
	var coachFeedback string
	if req.CoachContextEnable {
		coachFeedback = h.getCoachFeedbackString(userUUID)
	}

	// Get AI response
	aiResponse, err := h.aiClient.ChatWithAI(c.Request.Context(), conversation, req.Message, athleteProfile, coachFeedback)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get AI response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI response"})
		return
	}

	// Add messages to conversation
	userMessage := models.Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	}

	aiMessage := models.Message{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}

	conversation.Messages = append(conversation.Messages, userMessage, aiMessage)

	// Save conversation
	if conversation.ID == uuid.Nil {
		if err := h.programRepo.CreateAIConversation(&conversation); err != nil {
			log.Error().Err(err).Msg("Failed to create AI conversation")
		}
	} else {
		if err := h.programRepo.UpdateAIConversation(&conversation); err != nil {
			log.Error().Err(err).Msg("Failed to update AI conversation")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      aiResponse,
		"conversation": conversation,
	})
}

func (h *ProgramHandlers) LogWorkout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.LogWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log all completed sets
	for _, exercise := range req.Exercises {
		for _, set := range exercise.Sets {
			completedSet := &models.CompletedSet{
				ExerciseID:    exercise.ExerciseID,
				SetNumber:     set.SetNumber,
				RepsCompleted: set.RepsCompleted,
				WeightKg:      set.WeightKg,
				RPEActual:     set.RPEActual,
				VideoID:       set.VideoID,
				Notes:         set.Notes,
			}

			if err := h.programRepo.LogCompletedSet(completedSet); err != nil {
				log.Error().Err(err).Msg("Failed to log completed set")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log workout"})
				return
			}
		}
	}

	// Complete the session
	if err := h.programRepo.CompleteSession(req.SessionID, req.Notes, req.RPERating, req.Duration); err != nil {
		log.Error().Err(err).Msg("Failed to complete session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workout logged successfully"})
}

func (h *ProgramHandlers) ExportProgram(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	program, err := h.programRepo.GetProgramByID(req.ProgramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}

	// Check access permissions
	if !h.hasAccessToProgram(userID, program) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	sessions, err := h.programRepo.GetSessionsByProgramID(req.ProgramID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get training sessions for export")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare export"})
		return
	}

	if req.Format == "excel" {
		var buf bytes.Buffer
		if err := h.excelExporter.ExportProgram(*program, sessions, &buf); err != nil {
			log.Error().Err(err).Msg("Failed to export to Excel")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export program"})
			return
		}

		filename := fmt.Sprintf("%s_%s.xlsx", program.Name, time.Now().Format("2006-01-02"))
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format"})
	}
}

func (h *ProgramHandlers) GetProgramTemplates(c *gin.Context) {
	category := c.Query("category")
	experienceLevel := c.Query("experience_level")

	templates, err := h.programRepo.GetProgramTemplates(category, experienceLevel)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get program templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *ProgramHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "program-service",
	})
}

// Helper functions
func (h *ProgramHandlers) hasAccessToProgram(userID string, program *models.Program) bool {
	userUUID, _ := uuid.Parse(userID)
	
	// Athlete can access their own programs
	if program.AthleteID == userUUID {
		return true
	}

	// Coach can access if they're assigned to the program
	if program.CoachID != nil && *program.CoachID == userUUID {
		return true
	}

	// TODO: Check if user is a coach with access to this athlete
	return false
}

func (h *ProgramHandlers) getAthleteProfileString(athleteID uuid.UUID) string {
	// TODO: Fetch athlete profile from user service
	return fmt.Sprintf("Athlete ID: %s", athleteID.String())
}

func (h *ProgramHandlers) getCoachFeedbackString(athleteID uuid.UUID) string {
	// TODO: Fetch recent coach feedback from coach service
	return ""
}

func (h *ProgramHandlers) determineProgramPhase(competitionDate *time.Time) models.ProgramPhase {
	if competitionDate == nil {
		return models.PhaseStrength
	}

	weeksUntilComp := int(time.Until(*competitionDate).Hours() / (24 * 7))
	
	if weeksUntilComp <= 4 {
		return models.PhasePeaking
	} else if weeksUntilComp <= 12 {
		return models.PhaseStrength
	} else {
		return models.PhaseHypertrophy
	}
}

func stringPtr(s string) *string {
	return &s
}
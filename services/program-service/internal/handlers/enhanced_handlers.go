package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/program-service/internal/models"
	"github.com/rs/zerolog/log"
)

// GetPreviousSets retrieves historical sets for autofill
func (h *ProgramHandlers) GetPreviousSets(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	exerciseName := c.Param("exerciseName")
	if exerciseName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise name is required"})
		return
	}

	limit := 5 // default
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	previousSets, err := h.programRepo.GetPreviousSetsForExercise(athleteID.(uuid.UUID), exerciseName, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get previous sets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get previous sets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"previous_sets": previousSets})
}

// GenerateWarmups calculates warm-up sets based on working weight
func (h *ProgramHandlers) GenerateWarmups(c *gin.Context) {
	var req models.GenerateWarmupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	warmups := h.programRepo.GenerateWarmupSets(req.WorkingWeightKg, req.LiftType)

	c.JSON(http.StatusOK, gin.H{"warmup_sets": warmups})
}

// Exercise Library Handlers

func (h *ProgramHandlers) CreateExerciseLibrary(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateExerciseLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdBy := athleteID.(uuid.UUID)
	exercise := &models.ExerciseLibrary{
		Name:             req.Name,
		Description:      req.Description,
		LiftType:         req.LiftType,
		PrimaryMuscles:   req.PrimaryMuscles,
		SecondaryMuscles: req.SecondaryMuscles,
		Difficulty:       req.Difficulty,
		EquipmentNeeded:  req.EquipmentNeeded,
		DemoVideoURL:     req.DemoVideoURL,
		Instructions:     req.Instructions,
		FormCues:         req.FormCues,
		IsCustom:         true,
		CreatedBy:        &createdBy,
		IsPublic:         false, // custom exercises are private by default
	}

	if err := h.programRepo.CreateExerciseLibrary(exercise); err != nil {
		log.Error().Err(err).Msg("Failed to create exercise library entry")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exercise"})
		return
	}

	c.JSON(http.StatusCreated, exercise)
}

func (h *ProgramHandlers) GetExerciseLibrary(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var liftType *models.LiftType
	if lt := c.Query("lift_type"); lt != "" {
		t := models.LiftType(lt)
		liftType = &t
	}

	userID := athleteID.(uuid.UUID)
	exercises, err := h.programRepo.GetExerciseLibrary(&userID, liftType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get exercise library")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exercises"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}

// Workout Template Handlers

func (h *ProgramHandlers) CreateWorkoutTemplate(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateWorkoutTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.WorkoutTemplate{
		AthleteID:    athleteID.(uuid.UUID),
		Name:         req.Name,
		Description:  req.Description,
		TemplateData: req.TemplateData,
		IsPublic:     req.IsPublic,
	}

	if err := h.programRepo.CreateWorkoutTemplate(template); err != nil {
		log.Error().Err(err).Msg("Failed to create workout template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *ProgramHandlers) GetWorkoutTemplates(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	templates, err := h.programRepo.GetWorkoutTemplates(athleteID.(uuid.UUID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workout templates")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// Analytics Handlers

func (h *ProgramHandlers) GetVolumeData(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.GetVolumeDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to last 30 days if not specified
	if req.StartDate.IsZero() {
		req.StartDate = time.Now().AddDate(0, 0, -30)
	}
	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	volumeData, err := h.programRepo.GetVolumeData(athleteID.(uuid.UUID), req.StartDate, req.EndDate, req.ExerciseName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get volume data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get volume data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"volume_data": volumeData})
}

func (h *ProgramHandlers) GetE1RMData(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.GetE1RMDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to last 90 days if not specified
	if req.StartDate.IsZero() {
		req.StartDate = time.Now().AddDate(0, 0, -90)
	}
	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	e1rmData, err := h.programRepo.GetE1RMData(athleteID.(uuid.UUID), req.StartDate, req.EndDate, req.LiftType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get e1RM data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get e1RM data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"e1rm_data": e1rmData})
}

// Program Change Management Handlers

func (h *ProgramHandlers) ProposeChange(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.ProposeChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the program belongs to the athlete
	program, err := h.programRepo.GetProgramByID(req.ProgramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}

	if program.AthleteID != athleteID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this program"})
		return
	}

	change := &models.ProgramChange{
		ProgramID:         req.ProgramID,
		ChangeType:        "propose",
		ProposedChanges:   req.ProposedChanges,
		ChangeDescription: req.ChangeDescription,
		ProposedBy:        "athlete",
		Status:            "pending",
	}

	if err := h.programRepo.ProposeChange(change); err != nil {
		log.Error().Err(err).Msg("Failed to propose change")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to propose change"})
		return
	}

	c.JSON(http.StatusCreated, change)
}

func (h *ProgramHandlers) GetPendingChanges(c *gin.Context) {
	programID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}

	changes, err := h.programRepo.GetPendingChanges(programID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending changes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"changes": changes})
}

func (h *ProgramHandlers) ApplyChange(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	changeID, err := uuid.Parse(c.Param("changeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid change ID"})
		return
	}

	// TODO: Verify ownership before applying
	// This would require joining with program table

	if err := h.programRepo.ApplyChange(changeID); err != nil {
		log.Error().Err(err).Msg("Failed to apply change")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply change"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Change applied successfully"})
}

func (h *ProgramHandlers) RejectChange(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	changeID, err := uuid.Parse(c.Param("changeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid change ID"})
		return
	}

	if err := h.programRepo.RejectChange(changeID); err != nil {
		log.Error().Err(err).Msg("Failed to reject change")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject change"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Change rejected"})
}

// Historical Workout Management

func (h *ProgramHandlers) GetSessionHistory(c *gin.Context) {
	athleteID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse query parameters
	startDate := time.Now().AddDate(0, -3, 0) // default 3 months ago
	endDate := time.Now()
	limit := 50

	if sd := c.Query("start_date"); sd != "" {
		startDate, _ = time.Parse(time.RFC3339, sd)
	}
	if ed := c.Query("end_date"); ed != "" {
		endDate, _ = time.Parse(time.RFC3339, ed)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	sessions, err := h.programRepo.GetSessionHistory(athleteID.(uuid.UUID), startDate, endDate, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get session history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session history"})
		return
	}

	// Get exercises for each session
	for i := range sessions {
		exercises, err := h.programRepo.GetExercisesBySessionID(sessions[i].ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get exercises for session")
			continue
		}

		// Get completed sets for each exercise
		for j := range exercises {
			sets, err := h.programRepo.GetCompletedSetsByExerciseID(exercises[j].ID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get completed sets")
				continue
			}
			exercises[j].CompletedSets = sets
		}

		sessions[i].Exercises = exercises
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions, "total": len(sessions)})
}

func (h *ProgramHandlers) DeleteSession(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionID, err := uuid.Parse(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Get reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	// TODO: Verify ownership before deleting

	if err := h.programRepo.SoftDeleteSession(sessionID, req.Reason); err != nil {
		log.Error().Err(err).Msg("Failed to delete session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
}

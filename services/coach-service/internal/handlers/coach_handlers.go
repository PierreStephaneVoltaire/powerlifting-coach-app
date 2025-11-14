package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/powerlifting-coach-app/coach-service/internal/models"
	"github.com/powerlifting-coach-app/coach-service/internal/repository"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/middleware"
	"github.com/rs/zerolog/log"
)

type CoachHandlers struct {
	coachRepo *repository.CoachRepository
}

func NewCoachHandlers(coachRepo *repository.CoachRepository) *CoachHandlers {
	return &CoachHandlers{
		coachRepo: coachRepo,
	}
}

func (h *CoachHandlers) CreateFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can create feedback"})
		return
	}

	var req models.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	feedback := &models.CoachFeedback{
		CoachID:       coachUUID,
		AthleteID:     req.AthleteID,
		FeedbackType:  req.FeedbackType,
		Priority:      req.Priority,
		Title:         req.Title,
		Content:       req.Content,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Tags:          req.Tags,
		IsPrivate:     req.IsPrivate,
	}

	if err := h.coachRepo.CreateFeedback(feedback); err != nil {
		log.Error().Err(err).Msg("Failed to create feedback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback"})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}

func (h *CoachHandlers) GetMyFeedback(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view feedback"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	athleteIDStr := c.Query("athlete_id")
	feedbackType := c.Query("feedback_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var athleteID *uuid.UUID
	if athleteIDStr != "" {
		if parsed, err := uuid.Parse(athleteIDStr); err == nil {
			athleteID = &parsed
		}
	}

	var feedbackTypePtr *string
	if feedbackType != "" {
		feedbackTypePtr = &feedbackType
	}

	coachUUID, _ := uuid.Parse(userID)
	feedbacks, totalCount, err := h.coachRepo.GetFeedbackByCoachID(coachUUID, page, pageSize, athleteID, feedbackTypePtr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feedback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feedback"})
		return
	}

	response := models.FeedbackListResponse{
		Feedback:   feedbacks,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CoachHandlers) GetMyFeedbackAsAthlete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "athlete" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only athletes can view their feedback"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	athleteUUID, _ := uuid.Parse(userID)
	feedbacks, totalCount, err := h.coachRepo.GetFeedbackByAthleteID(athleteUUID, page, pageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feedback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feedback"})
		return
	}

	response := models.FeedbackListResponse{
		Feedback:   feedbacks,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CoachHandlers) GetFeedback(c *gin.Context) {
	feedbackIDStr := c.Param("id")
	feedbackID, err := uuid.Parse(feedbackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback ID"})
		return
	}

	feedback, err := h.coachRepo.GetFeedbackByID(feedbackID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feedback")
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	userUUID, _ := uuid.Parse(userID)

	// Check permissions
	if userType == "coach" && feedback.CoachID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	if userType == "athlete" && (feedback.AthleteID != userUUID || feedback.IsPrivate) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get responses
	responses, err := h.coachRepo.GetFeedbackResponses(feedbackID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get feedback responses")
		responses = []models.FeedbackResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"feedback":  feedback,
		"responses": responses,
	})
}

func (h *CoachHandlers) UpdateFeedback(c *gin.Context) {
	feedbackIDStr := c.Param("id")
	feedbackID, err := uuid.Parse(feedbackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback ID"})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can update feedback"})
		return
	}

	feedback, err := h.coachRepo.GetFeedbackByID(feedbackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if feedback.CoachID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req models.UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.FeedbackType != nil {
		feedback.FeedbackType = *req.FeedbackType
	}
	if req.Priority != nil {
		feedback.Priority = *req.Priority
	}
	if req.Title != nil {
		feedback.Title = *req.Title
	}
	if req.Content != nil {
		feedback.Content = *req.Content
	}
	if req.ReferenceType != nil {
		feedback.ReferenceType = req.ReferenceType
	}
	if req.ReferenceID != nil {
		feedback.ReferenceID = req.ReferenceID
	}
	if req.Tags != nil {
		feedback.Tags = *req.Tags
	}
	if req.IsPrivate != nil {
		feedback.IsPrivate = *req.IsPrivate
	}

	if err := h.coachRepo.UpdateFeedback(feedback); err != nil {
		log.Error().Err(err).Msg("Failed to update feedback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feedback"})
		return
	}

	c.JSON(http.StatusOK, feedback)
}

func (h *CoachHandlers) RespondToFeedback(c *gin.Context) {
	feedbackIDStr := c.Param("id")
	feedbackID, err := uuid.Parse(feedbackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback ID"})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "athlete" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only athletes can respond to feedback"})
		return
	}

	// Verify feedback exists and athlete has access
	feedback, err := h.coachRepo.GetFeedbackByID(feedbackID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if feedback.AthleteID != userUUID || feedback.IsPrivate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req models.RespondToFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := &models.FeedbackResponse{
		FeedbackID:       feedbackID,
		AthleteID:        userUUID,
		ResponseText:     req.ResponseText,
		IsAcknowledgment: req.IsAcknowledgment,
	}

	if err := h.coachRepo.CreateFeedbackResponse(response); err != nil {
		log.Error().Err(err).Msg("Failed to create feedback response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *CoachHandlers) CreateNote(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can create notes"})
		return
	}

	var req models.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	note := &models.CoachAthleteNote{
		CoachID:   coachUUID,
		AthleteID: req.AthleteID,
		NoteType:  req.NoteType,
		Title:     req.Title,
		Content:   req.Content,
	}

	if err := h.coachRepo.CreateNote(note); err != nil {
		log.Error().Err(err).Msg("Failed to create note")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *CoachHandlers) GetAthleteNotes(c *gin.Context) {
	athleteIDStr := c.Param("athlete_id")
	athleteID, err := uuid.Parse(athleteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid athlete ID"})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view notes"})
		return
	}

	coachUUID, _ := uuid.Parse(userID)
	notes, err := h.coachRepo.GetNotesByAthleteID(coachUUID, athleteID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (h *CoachHandlers) TrackProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can track progress"})
		return
	}

	var req models.TrackProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	// Calculate total if all lifts are provided
	var totalKg *float64
	if req.SquatMaxKg != nil && req.BenchMaxKg != nil && req.DeadliftMaxKg != nil {
		total := *req.SquatMaxKg + *req.BenchMaxKg + *req.DeadliftMaxKg
		totalKg = &total
	}

	progress := &models.AthleteProgressTracking{
		CoachID:       coachUUID,
		AthleteID:     req.AthleteID,
		TrackingDate:  req.TrackingDate,
		BodyWeightKg:  req.BodyWeightKg,
		SquatMaxKg:    req.SquatMaxKg,
		BenchMaxKg:    req.BenchMaxKg,
		DeadliftMaxKg: req.DeadliftMaxKg,
		TotalKg:       totalKg,
		Notes:         req.Notes,
		Measurements:  req.Measurements,
	}

	if err := h.coachRepo.TrackAthleteProgress(progress); err != nil {
		log.Error().Err(err).Msg("Failed to track progress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track progress"})
		return
	}

	c.JSON(http.StatusCreated, progress)
}

func (h *CoachHandlers) GetAthleteProgress(c *gin.Context) {
	athleteIDStr := c.Param("athlete_id")
	athleteID, err := uuid.Parse(athleteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid athlete ID"})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view progress"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	coachUUID, _ := uuid.Parse(userID)
	progress, err := h.coachRepo.GetAthleteProgress(coachUUID, athleteID, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get athlete progress")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progress})
}

func (h *CoachHandlers) GetNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view notifications"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	unreadOnly := c.Query("unread_only") == "true"

	if limit < 1 || limit > 200 {
		limit = 50
	}

	coachUUID, _ := uuid.Parse(userID)
	notifications, err := h.coachRepo.GetNotificationsByCoachID(coachUUID, limit, unreadOnly)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *CoachHandlers) MarkNotificationRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userType := middleware.GetUserType(c)
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can mark notifications as read"})
		return
	}

	if err := h.coachRepo.MarkNotificationAsRead(notificationID); err != nil {
		log.Error().Err(err).Msg("Failed to mark notification as read")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *CoachHandlers) GetDashboard(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	
	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can view dashboard"})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	// Get recent notifications
	notifications, err := h.coachRepo.GetNotificationsByCoachID(coachUUID, 10, false)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notifications for dashboard")
		notifications = []models.CoachNotification{}
	}

	// Get pending feedback (non-incorporated AI feedback)
	pendingFeedback, _, err := h.coachRepo.GetFeedbackByCoachID(coachUUID, 1, 10, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending feedback for dashboard")
		pendingFeedback = []models.CoachFeedback{}
	}

	// TODO: Get athlete overviews from user service
	athletes := []models.AthleteOverview{}
	
	// Count unread notifications
	unreadCount := 0
	for _, notif := range notifications {
		if !notif.IsRead {
			unreadCount++
		}
	}

	dashboard := models.CoachDashboard{
		Athletes:            athletes,
		RecentNotifications: notifications,
		PendingFeedback:     pendingFeedback,
		TotalAthletes:       len(athletes),
		UnreadNotifications: unreadCount,
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *CoachHandlers) SendRelationshipRequest(c *gin.Context) {
	var req models.SendRelationshipRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	athleteUUID, _ := uuid.Parse(userID)

	relationship := &models.CoachAthleteRelationship{
		CoachID:        req.CoachID,
		AthleteID:      athleteUUID,
		RequestMessage: req.RequestMessage,
	}

	if err := h.coachRepo.SendRelationshipRequest(relationship); err != nil {
		log.Error().Err(err).Msg("Failed to send relationship request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send relationship request"})
		return
	}

	c.JSON(http.StatusCreated, relationship)
}

func (h *CoachHandlers) AcceptRelationshipRequest(c *gin.Context) {
	relationshipIDStr := c.Param("id")
	relationshipID, err := uuid.Parse(relationshipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relationship ID"})
		return
	}

	userID := middleware.GetUserID(c)
	acceptedBy, _ := uuid.Parse(userID)

	if err := h.coachRepo.AcceptRelationshipRequest(relationshipID, acceptedBy); err != nil {
		log.Error().Err(err).Msg("Failed to accept relationship request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	relationship, err := h.coachRepo.GetRelationshipByID(relationshipID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated relationship")
		c.JSON(http.StatusOK, gin.H{"message": "Relationship accepted"})
		return
	}

	c.JSON(http.StatusOK, relationship)
}

func (h *CoachHandlers) TerminateRelationship(c *gin.Context) {
	relationshipIDStr := c.Param("id")
	relationshipID, err := uuid.Parse(relationshipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relationship ID"})
		return
	}

	var req models.TerminateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	terminatedBy, _ := uuid.Parse(userID)

	if err := h.coachRepo.TerminateRelationship(relationshipID, terminatedBy, req.TerminationReason, req.CooldownDays); err != nil {
		log.Error().Err(err).Msg("Failed to terminate relationship")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Relationship terminated"})
}

func (h *CoachHandlers) GetMyRelationships(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)
	userUUID, _ := uuid.Parse(userID)

	statusStr := c.Query("status")
	var status *models.RelationshipStatus
	if statusStr != "" {
		s := models.RelationshipStatus(statusStr)
		status = &s
	}

	var relationships []models.CoachAthleteRelationship
	var err error

	if userType == "coach" {
		relationships, err = h.coachRepo.GetRelationshipsByCoachID(userUUID, status)
	} else {
		relationships, err = h.coachRepo.GetRelationshipsByAthleteID(userUUID, status)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get relationships")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get relationships"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"relationships": relationships})
}

func (h *CoachHandlers) GetRelationship(c *gin.Context) {
	relationshipIDStr := c.Param("id")
	relationshipID, err := uuid.Parse(relationshipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid relationship ID"})
		return
	}

	relationship, err := h.coachRepo.GetRelationshipByID(relationshipID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get relationship")
		c.JSON(http.StatusNotFound, gin.H{"error": "Relationship not found"})
		return
	}

	c.JSON(http.StatusOK, relationship)
}

func (h *CoachHandlers) CreateCertification(c *gin.Context) {
	var req models.CreateCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)

	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can add certifications"})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	cert := &models.CoachCertification{
		CoachID:                 coachUUID,
		CertificationName:       req.CertificationName,
		IssuingOrganization:     req.IssuingOrganization,
		IssueDate:               req.IssueDate,
		ExpiryDate:              req.ExpiryDate,
		VerificationDocumentURL: req.VerificationDocumentURL,
	}

	if err := h.coachRepo.CreateCertification(cert); err != nil {
		log.Error().Err(err).Msg("Failed to create certification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create certification"})
		return
	}

	c.JSON(http.StatusCreated, cert)
}

func (h *CoachHandlers) GetCertifications(c *gin.Context) {
	coachIDStr := c.Param("coach_id")
	coachID, err := uuid.Parse(coachIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coach ID"})
		return
	}

	certs, err := h.coachRepo.GetCertificationsByCoachID(coachID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get certifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get certifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"certifications": certs})
}

func (h *CoachHandlers) CreateSuccessStory(c *gin.Context) {
	var req models.CreateSuccessStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	userType := middleware.GetUserType(c)

	if userType != "coach" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only coaches can add success stories"})
		return
	}

	coachUUID, _ := uuid.Parse(userID)

	story := &models.CoachSuccessStory{
		CoachID:         coachUUID,
		AthleteName:     req.AthleteName,
		Achievement:     req.Achievement,
		CompetitionName: req.CompetitionName,
		CompetitionDate: req.CompetitionDate,
		TotalKg:         req.TotalKg,
		WeightClass:     req.WeightClass,
		Federation:      req.Federation,
		Placement:       req.Placement,
		IsPublic:        req.IsPublic,
	}

	if err := h.coachRepo.CreateSuccessStory(story); err != nil {
		log.Error().Err(err).Msg("Failed to create success story")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create success story"})
		return
	}

	c.JSON(http.StatusCreated, story)
}

func (h *CoachHandlers) GetSuccessStories(c *gin.Context) {
	coachIDStr := c.Param("coach_id")
	coachID, err := uuid.Parse(coachIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coach ID"})
		return
	}

	publicOnly := c.Query("public_only") == "true"

	stories, err := h.coachRepo.GetSuccessStoriesByCoachID(coachID, publicOnly)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get success stories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get success stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success_stories": stories})
}

func (h *CoachHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "coach-service",
	})
}
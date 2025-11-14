package models

import (
	"time"
	"github.com/google/uuid"
)

type FeedbackType string
type FeedbackPriority string

const (
	FeedbackTypeProgramAdjustment FeedbackType = "program_adjustment"
	FeedbackTypeFormCorrection    FeedbackType = "form_correction"
	FeedbackTypeGeneralNote       FeedbackType = "general_note"
	FeedbackTypeMotivation        FeedbackType = "motivation"
)

const (
	PriorityLow    FeedbackPriority = "low"
	PriorityMedium FeedbackPriority = "medium"
	PriorityHigh   FeedbackPriority = "high"
	PriorityUrgent FeedbackPriority = "urgent"
)

type CoachFeedback struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	CoachID          uuid.UUID        `json:"coach_id" db:"coach_id"`
	AthleteID        uuid.UUID        `json:"athlete_id" db:"athlete_id"`
	FeedbackType     FeedbackType     `json:"feedback_type" db:"feedback_type"`
	Priority         FeedbackPriority `json:"priority" db:"priority"`
	Title            string           `json:"title" db:"title"`
	Content          string           `json:"content" db:"content"`
	ReferenceType    *string          `json:"reference_type" db:"reference_type"`
	ReferenceID      *uuid.UUID       `json:"reference_id" db:"reference_id"`
	Tags             []string         `json:"tags" db:"tags"`
	IsPrivate        bool             `json:"is_private" db:"is_private"`
	IncorporatedByAI bool             `json:"incorporated_by_ai" db:"incorporated_by_ai"`
	IncorporatedAt   *time.Time       `json:"incorporated_at" db:"incorporated_at"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

type CoachAthleteNote struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	CoachID    uuid.UUID  `json:"coach_id" db:"coach_id"`
	AthleteID  uuid.UUID  `json:"athlete_id" db:"athlete_id"`
	NoteType   string     `json:"note_type" db:"note_type"`
	Title      *string    `json:"title" db:"title"`
	Content    string     `json:"content" db:"content"`
	IsArchived bool       `json:"is_archived" db:"is_archived"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

type FeedbackResponse struct {
	ID               uuid.UUID `json:"id" db:"id"`
	FeedbackID       uuid.UUID `json:"feedback_id" db:"feedback_id"`
	AthleteID        uuid.UUID `json:"athlete_id" db:"athlete_id"`
	ResponseText     string    `json:"response_text" db:"response_text"`
	IsAcknowledgment bool      `json:"is_acknowledgment" db:"is_acknowledgment"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

type CoachNotification struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	CoachID         uuid.UUID  `json:"coach_id" db:"coach_id"`
	AthleteID       uuid.UUID  `json:"athlete_id" db:"athlete_id"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
	Title           string     `json:"title" db:"title"`
	Message         string     `json:"message" db:"message"`
	ReferenceType   *string    `json:"reference_type" db:"reference_type"`
	ReferenceID     *uuid.UUID `json:"reference_id" db:"reference_id"`
	IsRead          bool       `json:"is_read" db:"is_read"`
	IsArchived      bool       `json:"is_archived" db:"is_archived"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ReadAt          *time.Time `json:"read_at" db:"read_at"`
}

type AthleteProgressTracking struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	CoachID       uuid.UUID              `json:"coach_id" db:"coach_id"`
	AthleteID     uuid.UUID              `json:"athlete_id" db:"athlete_id"`
	TrackingDate  time.Time              `json:"tracking_date" db:"tracking_date"`
	BodyWeightKg  *float64               `json:"body_weight_kg" db:"body_weight_kg"`
	SquatMaxKg    *float64               `json:"squat_max_kg" db:"squat_max_kg"`
	BenchMaxKg    *float64               `json:"bench_max_kg" db:"bench_max_kg"`
	DeadliftMaxKg *float64               `json:"deadlift_max_kg" db:"deadlift_max_kg"`
	TotalKg       *float64               `json:"total_kg" db:"total_kg"`
	Notes         *string                `json:"notes" db:"notes"`
	Measurements  map[string]interface{} `json:"measurements" db:"measurements"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// Request/Response DTOs
type CreateFeedbackRequest struct {
	AthleteID     uuid.UUID        `json:"athlete_id" binding:"required"`
	FeedbackType  FeedbackType     `json:"feedback_type" binding:"required"`
	Priority      FeedbackPriority `json:"priority" binding:"required"`
	Title         string           `json:"title" binding:"required"`
	Content       string           `json:"content" binding:"required"`
	ReferenceType *string          `json:"reference_type"`
	ReferenceID   *uuid.UUID       `json:"reference_id"`
	Tags          []string         `json:"tags"`
	IsPrivate     bool             `json:"is_private"`
}

type UpdateFeedbackRequest struct {
	FeedbackType  *FeedbackType     `json:"feedback_type"`
	Priority      *FeedbackPriority `json:"priority"`
	Title         *string           `json:"title"`
	Content       *string           `json:"content"`
	ReferenceType *string           `json:"reference_type"`
	ReferenceID   *uuid.UUID        `json:"reference_id"`
	Tags          *[]string         `json:"tags"`
	IsPrivate     *bool             `json:"is_private"`
}

type CreateNoteRequest struct {
	AthleteID uuid.UUID `json:"athlete_id" binding:"required"`
	NoteType  string    `json:"note_type" binding:"required"`
	Title     *string   `json:"title"`
	Content   string    `json:"content" binding:"required"`
}

type UpdateNoteRequest struct {
	NoteType   *string `json:"note_type"`
	Title      *string `json:"title"`
	Content    *string `json:"content"`
	IsArchived *bool   `json:"is_archived"`
}

type RespondToFeedbackRequest struct {
	ResponseText     string `json:"response_text" binding:"required"`
	IsAcknowledgment bool   `json:"is_acknowledgment"`
}

type TrackProgressRequest struct {
	AthleteID     uuid.UUID              `json:"athlete_id" binding:"required"`
	TrackingDate  time.Time              `json:"tracking_date" binding:"required"`
	BodyWeightKg  *float64               `json:"body_weight_kg"`
	SquatMaxKg    *float64               `json:"squat_max_kg"`
	BenchMaxKg    *float64               `json:"bench_max_kg"`
	DeadliftMaxKg *float64               `json:"deadlift_max_kg"`
	Notes         *string                `json:"notes"`
	Measurements  map[string]interface{} `json:"measurements"`
}

type FeedbackListResponse struct {
	Feedback   []CoachFeedback `json:"feedback"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

type AthleteOverview struct {
	AthleteID        uuid.UUID  `json:"athlete_id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	AccessCode       string     `json:"access_code"`
	AccessGrantedAt  time.Time  `json:"access_granted_at"`
	LastActiveAt     *time.Time `json:"last_active_at"`
	UnreadFeedback   int        `json:"unread_feedback_count"`
	ActivePrograms   int        `json:"active_programs_count"`
	RecentProgress   *AthleteProgressTracking `json:"recent_progress"`
}

type CoachDashboard struct {
	Athletes            []AthleteOverview   `json:"athletes"`
	RecentNotifications []CoachNotification `json:"recent_notifications"`
	PendingFeedback     []CoachFeedback     `json:"pending_feedback"`
	TotalAthletes       int                 `json:"total_athletes"`
	UnreadNotifications int                 `json:"unread_notifications"`
}

type RelationshipStatus string

const (
	StatusPending    RelationshipStatus = "pending"
	StatusActive     RelationshipStatus = "active"
	StatusTerminated RelationshipStatus = "terminated"
)

type CoachAthleteRelationship struct {
	ID                 uuid.UUID          `json:"id" db:"id"`
	CoachID            uuid.UUID          `json:"coach_id" db:"coach_id"`
	AthleteID          uuid.UUID          `json:"athlete_id" db:"athlete_id"`
	Status             RelationshipStatus `json:"status" db:"status"`
	RequestMessage     *string            `json:"request_message" db:"request_message"`
	RequestedAt        time.Time          `json:"requested_at" db:"requested_at"`
	AcceptedAt         *time.Time         `json:"accepted_at" db:"accepted_at"`
	TerminatedAt       *time.Time         `json:"terminated_at" db:"terminated_at"`
	TerminatedBy       *uuid.UUID         `json:"terminated_by" db:"terminated_by"`
	TerminationReason  *string            `json:"termination_reason" db:"termination_reason"`
	CooldownUntil      *time.Time         `json:"cooldown_until" db:"cooldown_until"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
}

type CoachCertification struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	CoachID                 uuid.UUID  `json:"coach_id" db:"coach_id"`
	CertificationName       string     `json:"certification_name" db:"certification_name"`
	IssuingOrganization     *string    `json:"issuing_organization" db:"issuing_organization"`
	IssueDate               *time.Time `json:"issue_date" db:"issue_date"`
	ExpiryDate              *time.Time `json:"expiry_date" db:"expiry_date"`
	VerificationStatus      string     `json:"verification_status" db:"verification_status"`
	VerificationDocumentURL *string    `json:"verification_document_url" db:"verification_document_url"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
}

type CoachSuccessStory struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	CoachID         uuid.UUID  `json:"coach_id" db:"coach_id"`
	AthleteName     *string    `json:"athlete_name" db:"athlete_name"`
	Achievement     string     `json:"achievement" db:"achievement"`
	CompetitionName *string    `json:"competition_name" db:"competition_name"`
	CompetitionDate *time.Time `json:"competition_date" db:"competition_date"`
	TotalKg         *float64   `json:"total_kg" db:"total_kg"`
	WeightClass     *string    `json:"weight_class" db:"weight_class"`
	Federation      *string    `json:"federation" db:"federation"`
	Placement       *int       `json:"placement" db:"placement"`
	IsPublic        bool       `json:"is_public" db:"is_public"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type CoachProfile struct {
	ID              uuid.UUID            `json:"id"`
	UserID          uuid.UUID            `json:"user_id"`
	Name            string               `json:"name"`
	Email           string               `json:"email"`
	Bio             *string              `json:"bio"`
	Certifications  []CoachCertification `json:"certifications"`
	SuccessStories  []CoachSuccessStory  `json:"success_stories"`
	TotalAthletes   int                  `json:"total_athletes"`
	CreatedAt       time.Time            `json:"created_at"`
}

type SendRelationshipRequestRequest struct {
	CoachID        uuid.UUID `json:"coach_id" binding:"required"`
	RequestMessage *string   `json:"request_message"`
}

type AcceptRelationshipRequest struct {
}

type TerminateRelationshipRequest struct {
	TerminationReason *string `json:"termination_reason"`
	CooldownDays      *int    `json:"cooldown_days"`
}

type CreateCertificationRequest struct {
	CertificationName       string     `json:"certification_name" binding:"required"`
	IssuingOrganization     *string    `json:"issuing_organization"`
	IssueDate               *time.Time `json:"issue_date"`
	ExpiryDate              *time.Time `json:"expiry_date"`
	VerificationDocumentURL *string    `json:"verification_document_url"`
}

type CreateSuccessStoryRequest struct {
	AthleteName     *string    `json:"athlete_name"`
	Achievement     string     `json:"achievement" binding:"required"`
	CompetitionName *string    `json:"competition_name"`
	CompetitionDate *time.Time `json:"competition_date"`
	TotalKg         *float64   `json:"total_kg"`
	WeightClass     *string    `json:"weight_class"`
	Federation      *string    `json:"federation"`
	Placement       *int       `json:"placement"`
	IsPublic        bool       `json:"is_public"`
}
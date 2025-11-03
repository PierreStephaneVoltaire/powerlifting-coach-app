package models

import (
	"time"
	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationNewVideo          NotificationType = "new_video"
	NotificationMissedSession     NotificationType = "missed_session"
	NotificationProgramCompleted  NotificationType = "program_completed"
	NotificationNewFeedback       NotificationType = "new_feedback"
	NotificationFeedbackResponse  NotificationType = "feedback_response"
	NotificationNewProgram        NotificationType = "new_program"
	NotificationAccessGranted     NotificationType = "access_granted"
	NotificationFormAnalysis      NotificationType = "form_analysis"
	NotificationWelcome           NotificationType = "welcome"
	NotificationReminder          NotificationType = "reminder"
)

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelPush  NotificationChannel = "push"
	ChannelSMS   NotificationChannel = "sms"
)

type NotificationMessage struct {
	ID          uuid.UUID            `json:"id"`
	UserID      uuid.UUID            `json:"user_id"`
	Type        NotificationType     `json:"type"`
	Channel     NotificationChannel  `json:"channel"`
	Subject     string               `json:"subject"`
	Content     string               `json:"content"`
	Data        map[string]interface{} `json:"data"`
	Priority    int                  `json:"priority"` // 1=low, 5=high
	ScheduledAt *time.Time           `json:"scheduled_at"`
	CreatedAt   time.Time            `json:"created_at"`
}

type EmailTemplate struct {
	TemplateID string
	Subject    string
	PlainText  string
	HTML       string
}

// Event messages from other services
type VideoUploadedEvent struct {
	VideoID     uuid.UUID `json:"video_id"`
	AthleteID   uuid.UUID `json:"athlete_id"`
	Filename    string    `json:"filename"`
	UploadedAt  time.Time `json:"uploaded_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

type FeedbackCreatedEvent struct {
	FeedbackID uuid.UUID `json:"feedback_id"`
	CoachID    uuid.UUID `json:"coach_id"`
	AthleteID  uuid.UUID `json:"athlete_id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Priority   string    `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

type ProgramCreatedEvent struct {
	ProgramID   uuid.UUID `json:"program_id"`
	AthleteID   uuid.UUID `json:"athlete_id"`
	CoachID     *uuid.UUID `json:"coach_id"`
	Name        string    `json:"name"`
	Phase       string    `json:"phase"`
	StartDate   time.Time `json:"start_date"`
	WeeksTotal  int       `json:"weeks_total"`
	AIGenerated bool      `json:"ai_generated"`
	CreatedAt   time.Time `json:"created_at"`
}

type SessionMissedEvent struct {
	SessionID    uuid.UUID `json:"session_id"`
	ProgramID    uuid.UUID `json:"program_id"`
	AthleteID    uuid.UUID `json:"athlete_id"`
	SessionName  string    `json:"session_name"`
	ScheduledFor time.Time `json:"scheduled_for"`
	MissedAt     time.Time `json:"missed_at"`
}

type UserRegisteredEvent struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	UserType string    `json:"user_type"`
	JoinedAt time.Time `json:"joined_at"`
}

type AccessGrantedEvent struct {
	CoachID    uuid.UUID `json:"coach_id"`
	AthleteID  uuid.UUID `json:"athlete_id"`
	AccessCode string    `json:"access_code"`
	GrantedAt  time.Time `json:"granted_at"`
}

type FormAnalyzedEvent struct {
	AnalysisID uuid.UUID `json:"analysis_id"`
	VideoID    uuid.UUID `json:"video_id"`
	AthleteID  uuid.UUID `json:"athlete_id"`
	Exercise   string    `json:"exercise"`
	Score      float64   `json:"score"`
	Feedback   string    `json:"feedback"`
	CreatedAt  time.Time `json:"created_at"`
}

// User preferences for notifications
type UserNotificationPreferences struct {
	UserID              uuid.UUID `json:"user_id"`
	EmailEnabled        bool      `json:"email_enabled"`
	PushEnabled         bool      `json:"push_enabled"`
	SMSEnabled          bool      `json:"sms_enabled"`
	NewVideoNotifs      bool      `json:"new_video_notifications"`
	FeedbackNotifs      bool      `json:"feedback_notifications"`
	ProgramNotifs       bool      `json:"program_notifications"`
	ReminderNotifs      bool      `json:"reminder_notifications"`
	MarketingEmails     bool      `json:"marketing_emails"`
	WeeklyDigest        bool      `json:"weekly_digest"`
}
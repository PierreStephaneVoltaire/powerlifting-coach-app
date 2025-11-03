package notification

import (
	"fmt"
	"log"

	"github.com/powerlifting-coach-app/notification-service/internal/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Sender struct {
	sendGridClient *sendgrid.Client
	fromEmail     string
}

func NewSender(apiKey, fromEmail string) *Sender {
	return &Sender{
		sendGridClient: sendgrid.NewSendClient(apiKey),
		fromEmail:     fromEmail,
	}
}

func (s *Sender) SendNotification(notification models.NotificationMessage) error {
	switch notification.Channel {
	case models.ChannelEmail:
		return s.sendEmail(notification)
	case models.ChannelPush:
		return s.sendPushNotification(notification)
	case models.ChannelSMS:
		return s.sendSMS(notification)
	default:
		return fmt.Errorf("unsupported notification channel: %s", notification.Channel)
	}
}

func (s *Sender) sendEmail(notification models.NotificationMessage) error {
	from := mail.NewEmail("Powerlifting Coach", s.fromEmail)
	to := mail.NewEmail("", notification.UserID.String()+"@temp.com") // This would need user email lookup
	
	message := mail.NewSingleEmail(from, notification.Subject, to, notification.Content, s.generateHTMLContent(notification))
	
	response, err := s.sendGridClient.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body)
	}
	
	log.Printf("Email sent successfully to user %s", notification.UserID)
	return nil
}

func (s *Sender) sendPushNotification(notification models.NotificationMessage) error {
	// TODO: Implement push notification logic
	log.Printf("Push notification would be sent to user %s: %s", notification.UserID, notification.Subject)
	return nil
}

func (s *Sender) sendSMS(notification models.NotificationMessage) error {
	// TODO: Implement SMS logic
	log.Printf("SMS would be sent to user %s: %s", notification.UserID, notification.Content)
	return nil
}

func (s *Sender) generateHTMLContent(notification models.NotificationMessage) string {
	// Generate basic HTML template based on notification type
	switch notification.Type {
	case models.NotificationNewVideo:
		return s.generateVideoHTML(notification)
	case models.NotificationNewFeedback:
		return s.generateFeedbackHTML(notification)
	case models.NotificationNewProgram:
		return s.generateProgramHTML(notification)
	case models.NotificationMissedSession:
		return s.generateSessionHTML(notification)
	case models.NotificationWelcome:
		return s.generateWelcomeHTML(notification)
	case models.NotificationAccessGranted:
		return s.generateAccessHTML(notification)
	case models.NotificationFormAnalysis:
		return s.generateFormAnalysisHTML(notification)
	default:
		return s.generateGenericHTML(notification)
	}
}

func (s *Sender) generateVideoHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">New Video Uploaded</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Video:</strong> %s</p>
					<p><strong>Athlete ID:</strong> %s</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Video</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content, 
		notification.Data["filename"], 
		notification.Data["athlete_id"])
}

func (s *Sender) generateFeedbackHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">New Feedback Received</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Type:</strong> %s</p>
					<p><strong>Priority:</strong> %s</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Feedback</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content,
		notification.Data["type"],
		notification.Data["priority"])
}

func (s *Sender) generateProgramHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">New Training Program</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Program:</strong> %s</p>
					<p><strong>Phase:</strong> %s</p>
					<p><strong>Duration:</strong> %v weeks</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #17a2b8; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Program</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content,
		notification.Data["name"],
		notification.Data["phase"],
		notification.Data["weeks_total"])
}

func (s *Sender) generateSessionHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #dc3545;">Missed Training Session</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Session:</strong> %s</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #ffc107; color: #212529; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reschedule Session</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content,
		notification.Data["session_name"])
}

func (s *Sender) generateWelcomeHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #28a745;">Welcome to Powerlifting Coach!</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p>Get started by:</p>
					<ul>
						<li>Completing your profile</li>
						<li>Generating your first AI training program</li>
						<li>Uploading your first lift video</li>
					</ul>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Get Started</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content)
}

func (s *Sender) generateAccessHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #17a2b8;">Access Granted</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Access Code:</strong> %s</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #6f42c1; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Dashboard</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content,
		notification.Data["access_code"])
}

func (s *Sender) generateFormAnalysisHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #28a745;">Form Analysis Complete</h2>
				<p>%s</p>
				<div style="background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p><strong>Exercise:</strong> %s</p>
					<p><strong>Score:</strong> %.1f/10</p>
					<p><strong>Feedback:</strong> %s</p>
				</div>
				<p style="margin-top: 30px;">
					<a href="#" style="background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Full Analysis</a>
				</p>
			</div>
		</body>
		</html>
	`, notification.Content,
		notification.Data["exercise"],
		notification.Data["score"],
		notification.Data["feedback"])
}

func (s *Sender) generateGenericHTML(notification models.NotificationMessage) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2c3e50;">%s</h2>
				<p>%s</p>
			</div>
		</body>
		</html>
	`, notification.Subject, notification.Content)
}
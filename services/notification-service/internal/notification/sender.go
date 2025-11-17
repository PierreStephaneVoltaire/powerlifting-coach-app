package notification

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"

	"github.com/powerlifting-coach-app/notification-service/internal/models"
)

type Sender struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewSender(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail string) *Sender {
	return &Sender{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
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
	// This would need user email lookup in production
	recipientEmail := notification.UserID.String() + "@temp.com"

	// Build email message
	subject := notification.Subject
	htmlContent := *s.generateHTMLContent(notification)

	// Construct the email message with headers
	message := []byte(
		"From: " + s.fromEmail + "\r\n" +
		"To: " + recipientEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlContent + "\r\n")

	// Set up authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Connect to the SMTP server with TLS
	serverAddr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}

	// Connect to the server, authenticate, and send the email
	client, err := smtp.Dial(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err = client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(recipientEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	if err = client.Quit(); err != nil {
		return fmt.Errorf("failed to quit: %w", err)
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

func (s *Sender) generateHTMLContent(notification models.NotificationMessage) *string {
	html := s.generateHTML(notification)
	return &html
}

func (s *Sender) generateHTML(notification models.NotificationMessage) string {
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
				<h2 style="color: #28a745;">Welcome to Coach Potato!</h2>
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
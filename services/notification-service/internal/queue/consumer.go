package queue

import (
	"encoding/json"
	"fmt"

	"github.com/powerlifting-coach-app/notification-service/internal/models"
	"github.com/powerlifting-coach-app/notification-service/internal/notification"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/utils"
)

type Consumer struct {
	conn               *amqp.Connection
	channel            *amqp.Channel
	notificationSender *notification.Sender
	exchangeName       string
}

func NewConsumer(rabbitmqURL string, notificationSender *notification.Sender) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &Consumer{
		conn:               conn,
		channel:            ch,
		notificationSender: notificationSender,
		exchangeName:       "app.events",
	}

	if err := consumer.setupExchangesAndQueues(); err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %w", err)
	}

	return consumer, nil
}

func (c *Consumer) setupExchangesAndQueues() error {
	// Setup Dead Letter Queue
	if err := utils.SetupDLQ(c.channel); err != nil {
		return fmt.Errorf("failed to setup DLQ: %w", err)
	}

	// Declare the main events exchange
	err := c.channel.ExchangeDeclare(
		c.exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Define queue configurations
	queues := []struct {
		name       string
		routingKey string
	}{
		{"notifications.video.uploaded", "video.uploaded"},
		{"notifications.feedback.created", "feedback.created"},
		{"notifications.program.created", "program.created"},
		{"notifications.session.missed", "session.missed"},
		{"notifications.user.registered", "user.registered"},
		{"notifications.access.granted", "access.granted"},
		{"notifications.form.analyzed", "form.analyzed"},
	}

	for _, q := range queues {
		// Declare queue with single active consumer
		_, err := utils.DeclareQueueWithSingleActiveConsumer(c.channel, q.name)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		// Bind queue to exchange
		err = c.channel.QueueBind(
			q.name,
			q.routingKey,
			c.exchangeName,
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}
	}

	return nil
}

func (c *Consumer) StartConsuming() error {
	// Set QoS
	err := c.channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming from each queue
	c.consumeVideoEvents()
	c.consumeFeedbackEvents()
	c.consumeProgramEvents()
	c.consumeSessionEvents()
	c.consumeUserEvents()
	c.consumeAccessEvents()
	c.consumeFormAnalysisEvents()

	log.Info().Msg("Started consuming notification events")
	return nil
}

func (c *Consumer) consumeVideoEvents() {
	msgs, err := c.channel.Consume(
		"notifications.video.uploaded",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming video events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.VideoUploadedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal video event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleVideoUploaded(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle video uploaded event")
				c.handleMessageFailureWithRetry(msg, "video.uploaded")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeFeedbackEvents() {
	msgs, err := c.channel.Consume(
		"notifications.feedback.created",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming feedback events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.FeedbackCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal feedback event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleFeedbackCreated(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle feedback created event")
				c.handleMessageFailureWithRetry(msg, "feedback.created")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeProgramEvents() {
	msgs, err := c.channel.Consume(
		"notifications.program.created",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming program events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.ProgramCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal program event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleProgramCreated(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle program created event")
				c.handleMessageFailureWithRetry(msg, "program.created")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeSessionEvents() {
	msgs, err := c.channel.Consume(
		"notifications.session.missed",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming session events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.SessionMissedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal session event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleSessionMissed(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle session missed event")
				c.handleMessageFailureWithRetry(msg, "session.missed")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeUserEvents() {
	msgs, err := c.channel.Consume(
		"notifications.user.registered",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming user events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.UserRegisteredEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal user event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleUserRegistered(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle user registered event")
				c.handleMessageFailureWithRetry(msg, "user.registered")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeAccessEvents() {
	msgs, err := c.channel.Consume(
		"notifications.access.granted",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming access events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.AccessGrantedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal access event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleAccessGranted(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle access granted event")
				c.handleMessageFailureWithRetry(msg, "access.granted")
				continue
			}

			msg.Ack(false)
		}
	}()
}

func (c *Consumer) consumeFormAnalysisEvents() {
	msgs, err := c.channel.Consume(
		"notifications.form.analyzed",
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming form analysis events")
		return
	}

	go func() {
		for msg := range msgs {
			var event models.FormAnalyzedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal form analysis event")
				msg.Nack(false, false)
				continue
			}

			if err := c.handleFormAnalyzed(event); err != nil {
				log.Error().Err(err).Msg("Failed to handle form analyzed event")
				c.handleMessageFailureWithRetry(msg, "form.analyzed")
				continue
			}

			msg.Ack(false)
		}
	}()
}

// Event handlers
func (c *Consumer) handleVideoUploaded(event models.VideoUploadedEvent) error {
	// Notify coaches who have access to this athlete
	notification := models.NotificationMessage{
		UserID:  event.AthleteID, // Will be sent to coaches who have access
		Type:    models.NotificationNewVideo,
		Channel: models.ChannelEmail,
		Subject: "New Video Uploaded",
		Content: fmt.Sprintf("A new video '%s' has been uploaded and is ready for review.", event.Filename),
		Data: map[string]interface{}{
			"video_id":  event.VideoID.String(),
			"athlete_id": event.AthleteID.String(),
			"filename":  event.Filename,
		},
		Priority: 3,
	}

	return c.notificationSender.SendNotification(notification)
}

func (c *Consumer) handleFeedbackCreated(event models.FeedbackCreatedEvent) error {
	notification := models.NotificationMessage{
		UserID:  event.AthleteID,
		Type:    models.NotificationNewFeedback,
		Channel: models.ChannelEmail,
		Subject: fmt.Sprintf("New Feedback: %s", event.Title),
		Content: event.Content,
		Data: map[string]interface{}{
			"feedback_id": event.FeedbackID.String(),
			"coach_id":    event.CoachID.String(),
			"type":        event.Type,
			"priority":    event.Priority,
		},
		Priority: getPriorityFromString(event.Priority),
	}

	return c.notificationSender.SendNotification(notification)
}

func (c *Consumer) handleProgramCreated(event models.ProgramCreatedEvent) error {
	notification := models.NotificationMessage{
		UserID:  event.AthleteID,
		Type:    models.NotificationNewProgram,
		Channel: models.ChannelEmail,
		Subject: fmt.Sprintf("New Training Program: %s", event.Name),
		Content: fmt.Sprintf("Your new %s training program is ready! It's a %d-week program starting %s.", 
			event.Phase, event.WeeksTotal, event.StartDate.Format("January 2, 2006")),
		Data: map[string]interface{}{
			"program_id":   event.ProgramID.String(),
			"name":         event.Name,
			"phase":        event.Phase,
			"weeks_total":  event.WeeksTotal,
			"ai_generated": event.AIGenerated,
		},
		Priority: 4,
	}

	return c.notificationSender.SendNotification(notification)
}

func (c *Consumer) handleSessionMissed(event models.SessionMissedEvent) error {
	notification := models.NotificationMessage{
		UserID:  event.AthleteID,
		Type:    models.NotificationMissedSession,
		Channel: models.ChannelEmail,
		Subject: "Missed Training Session",
		Content: fmt.Sprintf("You missed your scheduled training session '%s' that was planned for %s.", 
			event.SessionName, event.ScheduledFor.Format("January 2, 2006")),
		Data: map[string]interface{}{
			"session_id":    event.SessionID.String(),
			"program_id":    event.ProgramID.String(),
			"session_name":  event.SessionName,
			"scheduled_for": event.ScheduledFor,
		},
		Priority: 3,
	}

	return c.notificationSender.SendNotification(notification)
}

func (c *Consumer) handleUserRegistered(event models.UserRegisteredEvent) error {
	notification := models.NotificationMessage{
		UserID:  event.UserID,
		Type:    models.NotificationWelcome,
		Channel: models.ChannelEmail,
		Subject: "Welcome to Powerlifting Coach!",
		Content: fmt.Sprintf("Welcome %s! We're excited to help you on your powerlifting journey.", event.Name),
		Data: map[string]interface{}{
			"name":      event.Name,
			"user_type": event.UserType,
		},
		Priority: 2,
	}

	return c.notificationSender.SendNotification(notification)
}

func (c *Consumer) handleAccessGranted(event models.AccessGrantedEvent) error {
	// Notify both coach and athlete
	coachNotification := models.NotificationMessage{
		UserID:  event.CoachID,
		Type:    models.NotificationAccessGranted,
		Channel: models.ChannelEmail,
		Subject: "New Athlete Access Granted",
		Content: "You now have access to a new athlete's data and can provide coaching feedback.",
		Data: map[string]interface{}{
			"athlete_id":  event.AthleteID.String(),
			"access_code": event.AccessCode,
		},
		Priority: 3,
	}

	athleteNotification := models.NotificationMessage{
		UserID:  event.AthleteID,
		Type:    models.NotificationAccessGranted,
		Channel: models.ChannelEmail,
		Subject: "Coach Access Granted",
		Content: "Your coach now has access to your training data and can provide personalized feedback.",
		Data: map[string]interface{}{
			"coach_id":    event.CoachID.String(),
			"access_code": event.AccessCode,
		},
		Priority: 3,
	}

	if err := c.notificationSender.SendNotification(coachNotification); err != nil {
		return err
	}

	return c.notificationSender.SendNotification(athleteNotification)
}

func (c *Consumer) handleFormAnalyzed(event models.FormAnalyzedEvent) error {
	notification := models.NotificationMessage{
		UserID:  event.AthleteID,
		Type:    models.NotificationFormAnalysis,
		Channel: models.ChannelEmail,
		Subject: "Form Analysis Complete",
		Content: fmt.Sprintf("Your %s form analysis is complete! Score: %.1f/10. %s", 
			event.Exercise, event.Score, event.Feedback),
		Data: map[string]interface{}{
			"analysis_id": event.AnalysisID.String(),
			"video_id":    event.VideoID.String(),
			"exercise":    event.Exercise,
			"score":       event.Score,
			"feedback":    event.Feedback,
		},
		Priority: 4,
	}

	return c.notificationSender.SendNotification(notification)
}

// handleMessageFailureWithRetry handles message failure with retry logic
func (c *Consumer) handleMessageFailureWithRetry(msg amqp.Delivery, routingKey string) {
	retryCount := utils.GetRetryCount(msg)
	log.Error().Int("retry_count", retryCount).Msg("Handling message failure")

	// Handle failure with retry logic
	if handleErr := utils.HandleMessageFailure(c.channel, msg, c.exchangeName, routingKey); handleErr != nil {
		log.Error().Err(handleErr).Msg("Failed to handle message failure")
		// Fallback to simple nack without requeue to avoid infinite loops
		msg.Nack(false, false)
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func getPriorityFromString(priority string) int {
	switch priority {
	case "urgent":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	default:
		return 3
	}
}
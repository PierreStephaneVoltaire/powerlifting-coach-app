package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"github.com/rs/zerolog/log"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/metrics"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/utils"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

const (
	VideoProcessingQueue    = "video.processing"
	VideoMetadataQueue      = "video.metadata"
	VideoProcessingExchange = "video.exchange"
)

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &RabbitMQClient{
		connection: conn,
		channel:    ch,
	}

	if err := client.setupQueues(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to setup queues: %w", err)
	}

	return client, nil
}

func (r *RabbitMQClient) setupQueues() error {
	// Setup Dead Letter Queue
	if err := utils.SetupDLQ(r.channel); err != nil {
		return fmt.Errorf("failed to setup DLQ: %w", err)
	}

	// Declare exchange
	err := r.channel.ExchangeDeclare(
		VideoProcessingExchange,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare processing queue with single active consumer
	_, err = utils.DeclareQueueWithSingleActiveConsumer(r.channel, VideoProcessingQueue)
	if err != nil {
		return fmt.Errorf("failed to declare processing queue: %w", err)
	}

	// Bind processing queue to exchange
	err = r.channel.QueueBind(
		VideoProcessingQueue,
		"process",
		VideoProcessingExchange,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind processing queue: %w", err)
	}

	// Declare metadata queue with single active consumer
	_, err = utils.DeclareQueueWithSingleActiveConsumer(r.channel, VideoMetadataQueue)
	if err != nil {
		return fmt.Errorf("failed to declare metadata queue: %w", err)
	}

	// Bind metadata queue to exchange
	err = r.channel.QueueBind(
		VideoMetadataQueue,
		"metadata",
		VideoProcessingExchange,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind metadata queue: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) PublishVideoProcessing(message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		VideoProcessingExchange,
		"process",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Make message persistent
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info().Interface("message", message).Msg("Published video processing message")
	return nil
}

func (r *RabbitMQClient) ConsumeVideoProcessing(handler func([]byte) error) error {
	// Set QoS to process one message at a time
	err := r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := r.channel.Consume(
		VideoProcessingQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		metrics.SetActiveConsumers("video-service", VideoProcessingQueue, 1)
		for msg := range msgs {
			start := time.Now()
			metrics.RecordMessageConsumed("video-service", VideoProcessingQueue, "process")

			log.Info().Str("body", string(msg.Body)).Msg("Received video processing message")

			err := handler(msg.Body)
			duration := time.Since(start)

			if err != nil {
				retryCount := utils.GetRetryCount(msg)
				metrics.RecordMessageFailed("video-service", VideoProcessingQueue, "process", retryCount, duration)
				log.Error().Err(err).Int("retry_count", retryCount).Msg("Failed to process message")

				// Handle failure with retry logic
				if handleErr := utils.HandleMessageFailureWithMetrics(r.channel, msg, VideoProcessingExchange, "process", "video-service", VideoProcessingQueue); handleErr != nil {
					log.Error().Err(handleErr).Msg("Failed to handle message failure")
					// Fallback to simple nack without requeue to avoid infinite loops
					msg.Nack(false, false)
				}
			} else {
				metrics.RecordMessageProcessed("video-service", VideoProcessingQueue, "process", duration)
				msg.Ack(false) // Acknowledge message
			}
		}
	}()

	log.Info().Msg("Video processing consumer started")
	return nil
}

func (r *RabbitMQClient) ConsumeVideoMetadata(handler func([]byte) error) error {
	// Set QoS
	err := r.channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := r.channel.Consume(
		VideoMetadataQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register metadata consumer: %w", err)
	}

	go func() {
		metrics.SetActiveConsumers("video-service", VideoMetadataQueue, 1)
		for msg := range msgs {
			start := time.Now()
			metrics.RecordMessageConsumed("video-service", VideoMetadataQueue, "metadata")

			log.Info().Str("body", string(msg.Body)).Msg("Received video metadata message")

			err := handler(msg.Body)
			duration := time.Since(start)

			if err != nil {
				retryCount := utils.GetRetryCount(msg)
				metrics.RecordMessageFailed("video-service", VideoMetadataQueue, "metadata", retryCount, duration)
				log.Error().Err(err).Int("retry_count", retryCount).Msg("Failed to process metadata")

				// Handle failure with retry logic
				if handleErr := utils.HandleMessageFailureWithMetrics(r.channel, msg, VideoProcessingExchange, "metadata", "video-service", VideoMetadataQueue); handleErr != nil {
					log.Error().Err(handleErr).Msg("Failed to handle message failure")
					// Fallback to simple nack without requeue to avoid infinite loops
					msg.Nack(false, false)
				}
			} else {
				metrics.RecordMessageProcessed("video-service", VideoMetadataQueue, "metadata", duration)
				msg.Ack(false) // Acknowledge message
			}
		}
	}()

	log.Info().Msg("Video metadata consumer started")
	return nil
}

func (r *RabbitMQClient) PublishEvent(routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.channel.Publish(
		AppEventsExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Info().Str("routing_key", routingKey).Msg("Published event")
	return nil
}

func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
}
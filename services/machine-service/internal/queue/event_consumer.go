package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type EventConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	handlers   map[string]EventHandler
}

type EventHandler func(ctx context.Context, payload []byte) error

type BaseEvent struct {
	SchemaVersion     string                 `json:"schema_version"`
	EventType         string                 `json:"event_type"`
	ClientGeneratedID string                 `json:"client_generated_id"`
	UserID            string                 `json:"user_id"`
	Timestamp         string                 `json:"timestamp"`
	SourceService     string                 `json:"source_service"`
	Data              map[string]interface{} `json:"data"`
}

const AppEventsExchange = "app.events"

func NewEventConsumer(url string) (*EventConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &EventConsumer{
		connection: conn,
		channel:    ch,
		handlers:   make(map[string]EventHandler),
	}

	if err := consumer.setupExchange(); err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to setup exchange: %w", err)
	}

	return consumer, nil
}

func (c *EventConsumer) setupExchange() error {
	err := c.channel.ExchangeDeclare(
		AppEventsExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	return nil
}

func (c *EventConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
	log.Info().Str("event_type", eventType).Msg("Registered event handler")
}

func (c *EventConsumer) StartConsuming(queueName string, routingKeys []string) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, key := range routingKeys {
		err = c.channel.QueueBind(
			queueName,
			key,
			AppEventsExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to %s: %w", key, err)
		}
		log.Info().Str("routing_key", key).Str("queue", queueName).Msg("Bound queue to routing key")
	}

	err = c.channel.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for msg := range msgs {
			c.handleMessage(msg)
		}
	}()

	log.Info().Str("queue", queueName).Msg("Event consumer started")
	return nil
}

func (c *EventConsumer) handleMessage(msg amqp.Delivery) {
	var event BaseEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal event")
		msg.Nack(false, false)
		return
	}

	handler, exists := c.handlers[event.EventType]
	if !exists {
		log.Warn().Str("event_type", event.EventType).Msg("No handler registered for event type")
		msg.Ack(false)
		return
	}

	ctx := context.Background()
	err := handler(ctx, msg.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("event_type", event.EventType).
			Str("client_generated_id", event.ClientGeneratedID).
			Msg("Failed to handle event")
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Info().
		Str("event_type", event.EventType).
		Str("client_generated_id", event.ClientGeneratedID).
		Msg("Event processed successfully")
}

func (c *EventConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

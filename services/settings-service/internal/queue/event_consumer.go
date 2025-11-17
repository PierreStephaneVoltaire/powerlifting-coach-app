package queue

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/PierreStephaneVoltaire/powerlifting-coach-app/shared/utils"
)

type EventHandler func(ctx context.Context, payload []byte) error

type EventConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	handlers   map[string]EventHandler
}

func NewEventConsumer(url string) (*EventConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"app.events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Setup Dead Letter Queue
	if err := utils.SetupDLQ(ch); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	consumer := &EventConsumer{
		connection: conn,
		channel:    ch,
		handlers:   make(map[string]EventHandler),
	}

	return consumer, nil
}

func (ec *EventConsumer) RegisterHandler(eventType string, handler EventHandler) {
	ec.handlers[eventType] = handler
	log.Info().Str("event_type", eventType).Msg("Event handler registered")
}

func (ec *EventConsumer) StartConsuming(queueName string, routingKeys []string) error {
	// Declare queue with single active consumer
	q, err := utils.DeclareQueueWithSingleActiveConsumer(ec.channel, queueName)
	if err != nil {
		return err
	}

	for _, routingKey := range routingKeys {
		err = ec.channel.QueueBind(
			q.Name,
			routingKey,
			"app.events",
			false,
			nil,
		)
		if err != nil {
			return err
		}
		log.Info().Str("routing_key", routingKey).Str("queue", queueName).Msg("Queue bound to routing key")
	}

	err = ec.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return err
	}

	msgs, err := ec.channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Info().Str("queue", queueName).Msg("Started consuming events")

	go func() {
		for msg := range msgs {
			ec.processMessage(msg)
		}
	}()

	return nil
}

func (ec *EventConsumer) processMessage(msg amqp.Delivery) {
	ctx := context.Background()
	eventType := msg.RoutingKey

	log.Info().Str("event_type", eventType).Msg("Processing event")

	handler, exists := ec.handlers[eventType]
	if !exists {
		log.Warn().Str("event_type", eventType).Msg("No handler registered for event type")
		msg.Nack(false, false)
		return
	}

	err := handler(ctx, msg.Body)
	if err != nil {
		retryCount := utils.GetRetryCount(msg)
		log.Error().Err(err).Str("event_type", eventType).Int("retry_count", retryCount).Msg("Event handler failed")

		// Handle failure with retry logic
		if handleErr := utils.HandleMessageFailure(ec.channel, msg, "app.events", eventType); handleErr != nil {
			log.Error().Err(handleErr).Msg("Failed to handle message failure")
			// Fallback to simple nack without requeue to avoid infinite loops
			msg.Nack(false, false)
		}
		return
	}

	msg.Ack(false)
	log.Info().Str("event_type", eventType).Msg("Event processed successfully")
}

func (ec *EventConsumer) Close() {
	if ec.channel != nil {
		ec.channel.Close()
	}
	if ec.connection != nil {
		ec.connection.Close()
	}
}

package queue

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type Publisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
}

func NewPublisher(rabbitmqURL string) (*Publisher, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	publisher := &Publisher{
		conn:         conn,
		channel:      ch,
		exchangeName: "app.events",
	}

	if err := publisher.setupExchange(); err != nil {
		publisher.Close()
		return nil, fmt.Errorf("failed to setup exchange: %w", err)
	}

	return publisher, nil
}

func (p *Publisher) setupExchange() error {
	err := p.channel.ExchangeDeclare(
		p.exchangeName,
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

	return nil
}

func (p *Publisher) PublishEvent(eventType string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.Publish(
		p.exchangeName,
		eventType,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Info().
		Str("event_type", eventType).
		Msg("Event published to RabbitMQ")

	return nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

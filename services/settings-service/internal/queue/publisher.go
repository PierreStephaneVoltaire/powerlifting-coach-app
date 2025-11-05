package queue

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
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

	return &Publisher{
		connection: conn,
		channel:    ch,
	}, nil
}

func (p *Publisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = p.channel.Publish(
		"app.events",
		eventType,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Error().Err(err).Str("event_type", eventType).Msg("Failed to publish event")
		return err
	}

	log.Info().Str("event_type", eventType).Msg("Event published")
	return nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
}

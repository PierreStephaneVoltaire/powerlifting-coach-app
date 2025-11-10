package queue

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

const (
	VideoProcessingQueue    = "video.processing"
	VideoProcessingExchange = "video.exchange"
	VideoMetadataQueue      = "video.metadata"
	VideoMetadataExchange   = "video.exchange"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

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
		return nil, err
	}

	return client, nil
}

func (c *RabbitMQClient) setupQueues() error {
	// Declare exchange
	err := c.channel.ExchangeDeclare(
		VideoProcessingExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare processing queue
	_, err = c.channel.QueueDeclare(
		VideoProcessingQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare processing queue: %w", err)
	}

	err = c.channel.QueueBind(
		VideoProcessingQueue,
		"process",
		VideoProcessingExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind processing queue: %w", err)
	}

	// Declare metadata queue
	_, err = c.channel.QueueDeclare(
		VideoMetadataQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare metadata queue: %w", err)
	}

	err = c.channel.QueueBind(
		VideoMetadataQueue,
		"metadata",
		VideoProcessingExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind metadata queue: %w", err)
	}

	return nil
}

func (c *RabbitMQClient) ConsumeVideoProcessing() (<-chan amqp.Delivery, error) {
	err := c.channel.Qos(1, 0, false) // Process one at a time
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		VideoProcessingQueue,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume: %w", err)
	}

	return msgs, nil
}

func (c *RabbitMQClient) PublishVideoMetadata(metadata interface{}) error {
	body, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = c.channel.Publish(
		VideoProcessingExchange,
		"metadata",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish metadata: %w", err)
	}

	return nil
}

func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

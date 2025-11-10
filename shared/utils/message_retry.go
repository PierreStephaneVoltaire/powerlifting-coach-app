package utils

import (
	"fmt"
	"strconv"

	"github.com/streadway/amqp"
)

const (
	// MaxRetries is the maximum number of times a message will be retried
	MaxRetries = 5

	// RetryCountHeader is the header key used to track retry count
	RetryCountHeader = "x-retry-count"

	// DLQExchange is the exchange name for dead letter queue
	DLQExchange = "app.dlq"

	// DLQQueue is the queue name for dead letter queue
	DLQQueue = "app.dlq.queue"
)

// GetRetryCount extracts the retry count from message headers
func GetRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}

	if count, ok := msg.Headers[RetryCountHeader]; ok {
		switch v := count.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}

	return 0
}

// ShouldRequeue determines if a message should be requeued or sent to DLQ
func ShouldRequeue(msg amqp.Delivery) bool {
	retryCount := GetRetryCount(msg)
	return retryCount < MaxRetries
}

// IncrementRetryCount creates a new Publishing with incremented retry count
func IncrementRetryCount(msg amqp.Delivery) amqp.Publishing {
	retryCount := GetRetryCount(msg)

	headers := make(amqp.Table)
	if msg.Headers != nil {
		for k, v := range msg.Headers {
			headers[k] = v
		}
	}

	headers[RetryCountHeader] = retryCount + 1

	return amqp.Publishing{
		ContentType:  msg.ContentType,
		Body:         msg.Body,
		DeliveryMode: msg.DeliveryMode,
		Priority:     msg.Priority,
		Headers:      headers,
		Timestamp:    msg.Timestamp,
		Type:         msg.Type,
		AppId:        msg.AppId,
		UserId:       msg.UserId,
	}
}

// HandleMessageFailure handles message failure with retry logic
// Returns true if message was handled (requeued or sent to DLQ), false if caller should nack
func HandleMessageFailure(channel *amqp.Channel, msg amqp.Delivery, originalExchange, originalRoutingKey string) error {
	retryCount := GetRetryCount(msg)

	if retryCount >= MaxRetries {
		// Send to DLQ
		dlqMsg := amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			DeliveryMode: amqp.Persistent,
			Headers:      msg.Headers,
			Timestamp:    msg.Timestamp,
			Type:         msg.Type,
			AppId:        msg.AppId,
			UserId:       msg.UserId,
		}

		// Add metadata about failure
		if dlqMsg.Headers == nil {
			dlqMsg.Headers = make(amqp.Table)
		}
		dlqMsg.Headers["x-original-exchange"] = originalExchange
		dlqMsg.Headers["x-original-routing-key"] = originalRoutingKey
		dlqMsg.Headers["x-death-reason"] = "max-retries-exceeded"

		err := channel.Publish(
			DLQExchange,
			DLQQueue,
			false,
			false,
			dlqMsg,
		)

		if err != nil {
			return fmt.Errorf("failed to publish to DLQ: %w", err)
		}

		// Ack the original message since we've moved it to DLQ
		return msg.Ack(false)
	}

	// Requeue with incremented retry count
	republishMsg := IncrementRetryCount(msg)

	err := channel.Publish(
		originalExchange,
		originalRoutingKey,
		false,
		false,
		republishMsg,
	)

	if err != nil {
		return fmt.Errorf("failed to republish message: %w", err)
	}

	// Ack the original message since we've republished it
	return msg.Ack(false)
}

// SetupDLQ creates the dead letter queue and exchange
func SetupDLQ(channel *amqp.Channel) error {
	// Declare DLQ exchange
	err := channel.ExchangeDeclare(
		DLQExchange,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ exchange: %w", err)
	}

	// Declare DLQ queue
	_, err = channel.QueueDeclare(
		DLQQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ queue: %w", err)
	}

	// Bind DLQ queue to DLQ exchange
	err = channel.QueueBind(
		DLQQueue,
		DLQQueue,
		DLQExchange,
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ queue: %w", err)
	}

	return nil
}

// DeclareQueueWithSingleActiveConsumer declares a queue with single active consumer enabled
// This ensures only one consumer processes messages at a time, even with multiple replicas
func DeclareQueueWithSingleActiveConsumer(channel *amqp.Channel, queueName string) (amqp.Queue, error) {
	args := amqp.Table{
		"x-single-active-consumer": true,
	}

	return channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,  // args
	)
}

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MessagesConsumed tracks total messages consumed from queues
	MessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_consumed_total",
			Help: "Total number of messages consumed from RabbitMQ queues",
		},
		[]string{"service", "queue", "routing_key"},
	)

	// MessagesProcessed tracks successfully processed messages
	MessagesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_processed_total",
			Help: "Total number of messages successfully processed",
		},
		[]string{"service", "queue", "routing_key"},
	)

	// MessagesFailed tracks failed message processing attempts
	MessagesFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_failed_total",
			Help: "Total number of failed message processing attempts",
		},
		[]string{"service", "queue", "routing_key", "retry_count"},
	)

	// MessagesSentToDLQ tracks messages sent to dead letter queue
	MessagesSentToDLQ = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_dlq_total",
			Help: "Total number of messages sent to dead letter queue",
		},
		[]string{"service", "queue", "routing_key"},
	)

	// MessageRetries tracks retry attempts
	MessageRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_message_retries_total",
			Help: "Total number of message retry attempts",
		},
		[]string{"service", "queue", "routing_key", "retry_count"},
	)

	// MessageProcessingDuration tracks how long it takes to process messages
	MessageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rabbitmq_message_processing_duration_seconds",
			Help:    "Time taken to process messages in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "queue", "routing_key", "status"},
	)

	// ActiveConsumers tracks the number of active consumers
	ActiveConsumers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_active_consumers",
			Help: "Number of active RabbitMQ consumers",
		},
		[]string{"service", "queue"},
	)

	// MessagesPublished tracks messages published to exchanges
	MessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_published_total",
			Help: "Total number of messages published to RabbitMQ exchanges",
		},
		[]string{"service", "exchange", "routing_key"},
	)

	// PublishErrors tracks errors when publishing messages
	PublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_publish_errors_total",
			Help: "Total number of errors when publishing messages",
		},
		[]string{"service", "exchange", "routing_key"},
	)
)

// RecordMessageConsumed records that a message was consumed
func RecordMessageConsumed(service, queue, routingKey string) {
	MessagesConsumed.WithLabelValues(service, queue, routingKey).Inc()
}

// RecordMessageProcessed records successful message processing
func RecordMessageProcessed(service, queue, routingKey string, duration time.Duration) {
	MessagesProcessed.WithLabelValues(service, queue, routingKey).Inc()
	MessageProcessingDuration.WithLabelValues(service, queue, routingKey, "success").Observe(duration.Seconds())
}

// RecordMessageFailed records failed message processing
func RecordMessageFailed(service, queue, routingKey string, retryCount int, duration time.Duration) {
	MessagesFailed.WithLabelValues(service, queue, routingKey, formatRetryCount(retryCount)).Inc()
	MessageProcessingDuration.WithLabelValues(service, queue, routingKey, "failed").Observe(duration.Seconds())
}

// RecordMessageRetry records a retry attempt
func RecordMessageRetry(service, queue, routingKey string, retryCount int) {
	MessageRetries.WithLabelValues(service, queue, routingKey, formatRetryCount(retryCount)).Inc()
}

// RecordMessageToDLQ records a message sent to DLQ
func RecordMessageToDLQ(service, queue, routingKey string) {
	MessagesSentToDLQ.WithLabelValues(service, queue, routingKey).Inc()
}

// RecordMessagePublished records a published message
func RecordMessagePublished(service, exchange, routingKey string) {
	MessagesPublished.WithLabelValues(service, exchange, routingKey).Inc()
}

// RecordPublishError records a publish error
func RecordPublishError(service, exchange, routingKey string) {
	PublishErrors.WithLabelValues(service, exchange, routingKey).Inc()
}

// SetActiveConsumers sets the number of active consumers
func SetActiveConsumers(service, queue string, count int) {
	ActiveConsumers.WithLabelValues(service, queue).Set(float64(count))
}

// formatRetryCount formats retry count for metric labels
func formatRetryCount(count int) string {
	if count == 0 {
		return "0"
	} else if count == 1 {
		return "1"
	} else if count == 2 {
		return "2"
	} else if count == 3 {
		return "3"
	} else if count == 4 {
		return "4"
	} else if count >= 5 {
		return "5+"
	}
	return "unknown"
}

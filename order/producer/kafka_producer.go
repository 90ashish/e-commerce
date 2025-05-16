package producer

import (
	"context"
	"encoding/json"

	"e-commerce/common/models"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps a kafka-go Writer to implement OrderPublisher.
// It publishes messages keyed by UserID for partition affinity.
type KafkaProducer struct {
	Writer *kafka.Writer
}

// Ensure KafkaProducer satisfies OrderPublisher
var _ OrderPublisher = (*KafkaProducer)(nil)

// NewKafkaProducer constructs a new KafkaProducer for the given brokers and topic.
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.Hash{}, // Distribute by key hash
	})
	return &KafkaProducer{Writer: writer}
}

// Publish serializes the OrderCreated event to JSON and writes to Kafka.
func (kp *KafkaProducer) Publish(event models.OrderCreated) error {
	// Marshal the Go struct into JSON bytes
	data, err := json.Marshal(event)
	if err != nil {
		return err // JSON encoding failed
	}
	// Construct a Kafka message with key affinity
	msg := kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
	}
	// Write synchronously; context.Background used for demo
	return kp.Writer.WriteMessages(context.Background(), msg)
}

// Close flushes pending messages and closes network connections.
func (kp *KafkaProducer) Close() error {
	return kp.Writer.Close()
}

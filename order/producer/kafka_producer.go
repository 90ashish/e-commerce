package producer

import (
	"context"
	"e-commerce/common/models"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps a kafka writer
type KafkaProducer struct {
	Writer *kafka.Writer
}

// NewKafkaProducer constructs a new KafkaProducer for the given brokers and topic.
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{}, // key-based partitioning
		},
	}
}

// Publish serializes the OrderCreated event to JSON and writes to Kafka.
func (kp *KafkaProducer) Publish(event models.OrderCreated) error {
	// Marshal the Go struct into JSON bytes
	data, err := json.Marshal(event)
	if err != nil {
		return err
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

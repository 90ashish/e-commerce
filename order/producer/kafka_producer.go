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

// NewKafkaProducer constructs a producer for a topic
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{}, // key-based partitioning
		},
	}
}

// Publish sends an OrderCreated event
func (kp *KafkaProducer) Publish(event models.OrderCreated) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(event.UserID),
		Value: data,
	}
	return kp.Writer.WriteMessages(context.Background(), msg)
}

// Close flushes and closes the writer
func (kp *KafkaProducer) Close() error {
	return kp.Writer.Close()
}

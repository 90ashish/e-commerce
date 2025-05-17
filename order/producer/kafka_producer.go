package producer

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"e-commerce/common/models"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaProducer wraps a kafka writer with retry and idempotency.
type KafkaProducer struct {
	writer   *kafka.Writer
	logger   *zap.Logger
	seenKeys sync.Map // for deduping OrderID
}

// NewKafkaProducer constructs a producer with backoff retry.
func NewKafkaProducer(brokers []string, topic string, log *zap.Logger) *KafkaProducer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})
	return &KafkaProducer{writer: w, logger: log}
}

// Publish sends an OrderCreated event, retrying transient errors.
// It also dedupes on OrderID: only first publish is allowed.
func (kp *KafkaProducer) Publish(evt models.OrderCreated) error {
	// Idempotency: skip if already seen
	if _, loaded := kp.seenKeys.LoadOrStore(evt.OrderID, true); loaded {
		kp.logger.Warn("Duplicate OrderID, skipping publish", zap.String("orderID", evt.OrderID))
		return nil
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	msg := kafka.Message{Key: []byte(evt.OrderID), Value: data}

	// Retry with exponential backoff
	backoff := 100 * time.Millisecond
	for i := 0; i < 5; i++ {
		err = kp.writer.WriteMessages(context.Background(), msg)
		if err == nil {
			kp.logger.Info("Published order event", zap.String("orderID", evt.OrderID))
			return nil
		}
		kp.logger.Warn("Publish failed, retrying", zap.Error(err), zap.Int("attempt", i+1))
		time.Sleep(backoff)
		backoff *= 2
	}
	return errors.New("failed to publish after retries: " + err.Error())
}

// Close flushes and closes the writer
func (kp *KafkaProducer) Close() error {
	kp.logger.Info("Closing Kafka producer, flushing messages")
	return kp.writer.Close()
}

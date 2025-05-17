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

// InventoryProducer adds retry and dedupe logic.
type InventoryProducer struct {
	reservedWriter *kafka.Writer
	failedWriter   *kafka.Writer
	logger         *zap.Logger
	seenKeys       sync.Map // dedupe by orderID
}

func NewInventoryProducer(brokers []string, log *zap.Logger) *InventoryProducer {
	return &InventoryProducer{
		reservedWriter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    "inventory.reserved",
			Balancer: &kafka.Hash{},
		}),
		failedWriter: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    "inventory.failed",
			Balancer: &kafka.Hash{},
		}),
		logger: log,
	}
}

// publishWithRetry writes the message, retrying on transient errors,
// and dedupes by the provided key (orderID), not by topic.
func (p *InventoryProducer) publishWithRetry(
	writer *kafka.Writer,
	orderID string,
	value []byte,
) error {
	// Deduplication: skip if we've already published this orderID
	if _, loaded := p.seenKeys.LoadOrStore(orderID, true); loaded {
		p.logger.Warn("Duplicate publish skipped", zap.String("orderID", orderID))
		return nil
	}

	msg := kafka.Message{Key: []byte(orderID), Value: value}
	backoff := 100 * time.Millisecond
	for attempt := 1; attempt <= 5; attempt++ {
		if err := writer.WriteMessages(context.Background(), msg); err != nil {
			p.logger.Warn("Publish failed, retrying",
				zap.String("topic", writer.Topic),
				zap.String("orderID", orderID),
				zap.Error(err),
				zap.Int("attempt", attempt),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		p.logger.Info("Published event",
			zap.String("topic", writer.Topic),
			zap.String("orderID", orderID),
		)
		return nil
	}
	return errors.New("failed to publish after retries")
}

// EmitReserved publishes a Reservation event.
func (p *InventoryProducer) EmitReserved(evt models.InventoryReserved) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.publishWithRetry(p.reservedWriter, evt.OrderID, data)
}

// EmitFailed publishes a Failure event.
func (p *InventoryProducer) EmitFailed(evt models.InventoryFailed) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.publishWithRetry(p.failedWriter, evt.OrderID, data)
}

// Close flushes both writers.
func (p *InventoryProducer) Close() error {
	p.logger.Info("Closing InventoryProducer")
	if err := p.reservedWriter.Close(); err != nil {
		return err
	}
	return p.failedWriter.Close()
}

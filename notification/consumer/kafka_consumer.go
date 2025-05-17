package consumer

import (
	"context"
	"encoding/json"

	"e-commerce/common/models"
	"e-commerce/notification/sink"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// NotificationConsumer reads and notifies with structured logging.
type NotificationConsumer struct {
	reservedReader *kafka.Reader
	failedReader   *kafka.Reader
	sink           sink.NotificationSink
	logger         *zap.Logger
}

// NewNotificationConsumer creates two readers (one per topic) sharing the same group.
func NewNotificationConsumer(
	brokers []string,
	groupID string,
	notifSink sink.NotificationSink,
	log *zap.Logger,
) *NotificationConsumer {
	return &NotificationConsumer{
		reservedReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          "inventory.reserved",
			GroupID:        groupID,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: 0,
		}),
		failedReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          "inventory.failed",
			GroupID:        groupID,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: 0,
		}),
		sink:   notifSink,
		logger: log,
	}
}

// Run starts two goroutines, one for each topic reader.
// It returns when the context is canceled.
func (c *NotificationConsumer) Run(ctx context.Context) {
	c.logger.Info("🔔 Notification consumer started")
	// Helper to process one reader
	process := func(reader *kafka.Reader, handle func([]byte) error) {
		for {
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Warn("FetchMessage error", zap.Error(err))
				return
			}
			if err := handle(m.Value); err != nil {
				c.logger.Error("Handle message error", zap.Error(err))
			}
			if err := reader.CommitMessages(ctx, m); err != nil {
				c.logger.Warn("Commit offset failed", zap.Error(err))
			}
		}
	}

	// Reserved
	go process(c.reservedReader, func(val []byte) error {
		var evt models.InventoryReserved
		if err := json.Unmarshal(val, &evt); err != nil {
			return err
		}
		return c.sink.NotifyReserved(evt)
	})
	// Failed
	go process(c.failedReader, func(val []byte) error {
		var evt models.InventoryFailed
		if err := json.Unmarshal(val, &evt); err != nil {
			return err
		}
		return c.sink.NotifyFailed(evt)
	})

	<-ctx.Done()
}

// Close shuts down both readers.
func (c *NotificationConsumer) Close() error {
	c.logger.Info("Closing NotificationConsumer")
	if err := c.reservedReader.Close(); err != nil {
		return err
	}
	return c.failedReader.Close()
}

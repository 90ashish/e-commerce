// notification/consumer/kafka_consumer.go

package consumer

import (
	"context"
	"encoding/json"
	"log"

	"e-commerce/common/models"
	"e-commerce/notification/sink"

	"github.com/segmentio/kafka-go"
)

// NotificationConsumer reads both reserved & failed topics.
type NotificationConsumer struct {
	reservedReader *kafka.Reader
	failedReader   *kafka.Reader
	sink           sink.NotificationSink
}

// NewNotificationConsumer creates two readers (one per topic) sharing the same group.
func NewNotificationConsumer(brokers []string, groupID string, notifSink sink.NotificationSink) *NotificationConsumer {
	// Reader for "inventory.reserved"
	reservedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "inventory.reserved",
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // manual commits
	})

	// Reader for "inventory.failed"
	failedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "inventory.failed",
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &NotificationConsumer{
		reservedReader: reservedReader,
		failedReader:   failedReader,
		sink:           notifSink,
	}
}

// Run starts two goroutines, one for each topic reader.
// It returns when the context is canceled.
func (c *NotificationConsumer) Run(ctx context.Context) {
	log.Println("🔔 Notification consumer started")
	// Process reserved events
	go func() {
		for {
			m, err := c.reservedReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("⚠️ reserved fetch error: %v", err)
				return
			}
			var evt models.InventoryReserved
			if err := json.Unmarshal(m.Value, &evt); err != nil {
				log.Printf("⚠️ invalid reserved payload: %v", err)
				_ = c.reservedReader.CommitMessages(ctx, m)
				continue
			}
			if err := c.sink.NotifyReserved(evt); err != nil {
				log.Printf("⚠️ notify reserved failed: %v", err)
			}
			if err := c.reservedReader.CommitMessages(ctx, m); err != nil {
				log.Printf("⚠️ commit reserved offset error: %v", err)
			}
		}
	}()

	// Process failed events
	go func() {
		for {
			m, err := c.failedReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("⚠️ failed fetch error: %v", err)
				return
			}
			var evt models.InventoryFailed
			if err := json.Unmarshal(m.Value, &evt); err != nil {
				log.Printf("⚠️ invalid failed payload: %v", err)
				_ = c.failedReader.CommitMessages(ctx, m)
				continue
			}
			if err := c.sink.NotifyFailed(evt); err != nil {
				log.Printf("⚠️ notify failed failed: %v", err)
			}
			if err := c.failedReader.CommitMessages(ctx, m); err != nil {
				log.Printf("⚠️ commit failed offset error: %v", err)
			}
		}
	}()

	// Block until context is canceled
	<-ctx.Done()
}

// Close shuts down both readers.
func (c *NotificationConsumer) Close() error {
	if err := c.reservedReader.Close(); err != nil {
		return err
	}
	return c.failedReader.Close()
}

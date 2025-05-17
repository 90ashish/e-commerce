package consumer

import (
	"context"
	"encoding/json"
	"log"

	"e-commerce/common/models"
	"e-commerce/inventory/producer"
	"e-commerce/inventory/service"

	"github.com/segmentio/kafka-go"
)

// InventoryConsumer coordinates reads, business logic, and emits.
type InventoryConsumer struct {
	reader   *kafka.Reader
	producer *producer.InventoryProducer
	stockSvc *service.StockService
}

// NewInventoryConsumer sets up a consumer group on "orders.created".
func NewInventoryConsumer(brokers []string, groupID string, stockSvc *service.StockService, prod *producer.InventoryProducer) *InventoryConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "orders.created",
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		// Disable auto-commit: we’ll commit only after produce
		CommitInterval: 0,
	})
	return &InventoryConsumer{
		reader:   r,
		producer: prod,
		stockSvc: stockSvc,
	}
}

// Run starts the infinite consume→process loop.
func (c *InventoryConsumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("consumer fetch error: %v", err)
			return
		}

		// Deserialize incoming OrderCreated
		var order models.OrderCreated
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid payload, skipping offset %d: %v", m.Offset, err)
			// commit to skip bad message
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		// Business logic: attempt to reserve stock
		_, err = c.stockSvc.Reserve(order.Items)
		if err != nil {
			// Emit failure event
			failEvt := models.InventoryFailed{
				OrderID: order.OrderID,
				Items:   order.Items,
				Reason:  err.Error(),
			}
			if err := c.producer.EmitFailed(failEvt); err != nil {
				log.Printf("failed to emit failure for order %s: %v", order.OrderID, err)
				continue
			}
		} else {
			// Emit reserved event
			resEvt := models.InventoryReserved{
				OrderID: order.OrderID,
				Items:   order.Items,
			}
			if err := c.producer.EmitReserved(resEvt); err != nil {
				log.Printf("failed to emit reserved for order %s: %v", order.OrderID, err)
				continue
			}
		}

		// Only after emit succeeds do we commit the offset
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("failed to commit offset %d: %v", m.Offset, err)
		}
	}
}

// Close shuts down the reader.
func (c *InventoryConsumer) Close() error {
	return c.reader.Close()
}

package consumer

import (
	"context"
	"encoding/json"

	"e-commerce/common/models"
	"e-commerce/inventory/producer"
	"e-commerce/inventory/service"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// InventoryConsumer processes orders.created with structured logging.
type InventoryConsumer struct {
	reader   *kafka.Reader
	producer *producer.InventoryProducer
	stockSvc *service.StockService
	logger   *zap.Logger
}

// NewInventoryConsumer configures a manual-commit reader.
func NewInventoryConsumer(
	brokers []string,
	groupID string,
	stockSvc *service.StockService,
	prod *producer.InventoryProducer,
	log *zap.Logger,
) *InventoryConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "orders.created",
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	return &InventoryConsumer{
		reader:   r,
		producer: prod,
		stockSvc: stockSvc,
		logger:   log,
	}
}

// Run consumes, processes, and acknowledges messages.
func (c *InventoryConsumer) Run(ctx context.Context) {
	c.logger.Info("Inventory consumer started")
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.logger.Warn("FetchMessage error, stopping consumer", zap.Error(err))
			return
		}

		var order models.OrderCreated
		if err := json.Unmarshal(m.Value, &order); err != nil {
			c.logger.Error("Invalid OrderCreated payload", zap.Error(err), zap.Int64("offset", m.Offset))
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		// Reserve stock
		if _, err := c.stockSvc.Reserve(order.Items); err != nil {
			c.logger.Info("Stock reserve failed", zap.String("orderID", order.OrderID), zap.Error(err))
			failEvt := models.InventoryFailed{OrderID: order.OrderID, Items: order.Items, Reason: err.Error()}
			if err := c.producer.EmitFailed(failEvt); err != nil {
				c.logger.Error("EmitFailed error", zap.Error(err), zap.String("orderID", order.OrderID))
				continue
			}
		} else {
			c.logger.Info("Stock reserved", zap.String("orderID", order.OrderID))
			resEvt := models.InventoryReserved{OrderID: order.OrderID, Items: order.Items}
			if err := c.producer.EmitReserved(resEvt); err != nil {
				c.logger.Error("EmitReserved error", zap.Error(err), zap.String("orderID", order.OrderID))
				continue
			}
		}

		// Commit after successful emit
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("CommitMessages error", zap.Error(err), zap.Int64("offset", m.Offset))
		}
	}
}

// Close shuts down the reader.
func (c *InventoryConsumer) Close() error {
	c.logger.Info("Closing InventoryConsumer")
	return c.reader.Close()
}

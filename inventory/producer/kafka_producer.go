package producer

import (
	"context"
	"encoding/json"

	"e-commerce/common/models"

	"github.com/segmentio/kafka-go"
)

// InventoryProducer emits reserved/failed events.
type InventoryProducer struct {
	reservedWriter *kafka.Writer
	failedWriter   *kafka.Writer
}

// NewInventoryProducer constructs writers for both topics.
func NewInventoryProducer(brokers []string) *InventoryProducer {
	return &InventoryProducer{
		reservedWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "inventory.reserved",
			Balancer: &kafka.Hash{},
		},
		failedWriter: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "inventory.failed",
			Balancer: &kafka.Hash{},
		},
	}
}

// EmitReserved publishes an InventoryReserved event.
func (p *InventoryProducer) EmitReserved(evt models.InventoryReserved) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(evt.OrderID),
		Value: data,
	}
	return p.reservedWriter.WriteMessages(context.Background(), msg)
}

// EmitFailed publishes an InventoryFailed event.
func (p *InventoryProducer) EmitFailed(evt models.InventoryFailed) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte(evt.OrderID),
		Value: data,
	}
	return p.failedWriter.WriteMessages(context.Background(), msg)
}

// Close flushes both writers.
func (p *InventoryProducer) Close() error {
	if err := p.reservedWriter.Close(); err != nil {
		return err
	}
	return p.failedWriter.Close()
}

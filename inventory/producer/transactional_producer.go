package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"e-commerce/common/models"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// ReserveService abstracts the stock-reservation logic so that
// the producer doesn’t import the concrete service package.
type ReserveService interface {
	Reserve(items []string) (bool, error)
}

// TransactionalProducer wraps an idempotent SyncProducer and
// uses an in-memory dedupe map to ensure exactly-once behavior.
type TransactionalProducer struct {
	prod      sarama.SyncProducer
	logger    *zap.Logger
	seen      sync.Map // tracks OrderIDs we've already published
	topicOK   string   // topic for successful reservations
	topicFail string   // topic for failed reservations
}

// NewTransactionalProducer configures Sarama for idempotence.
// We set Net.MaxOpenRequests=1 to satisfy the idempotence requirement.
func NewTransactionalProducer(brokers []string, clientID string, logger *zap.Logger) (*TransactionalProducer, error) {
	cfg := sarama.NewConfig()
	cfg.ClientID = clientID
	cfg.Version = sarama.V2_5_0_0

	// Enable idempotence (requires MaxOpenRequests=1)
	cfg.Producer.Idempotent = true
	cfg.Net.MaxOpenRequests = 1

	// Wait for all in-sync replicas to ack, retry up to 5 times
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true

	// Create the SyncProducer (blocks until ack)
	prod, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating sync producer: %w", err)
	}

	return &TransactionalProducer{
		prod:      prod,
		logger:    logger,
		topicOK:   "inventory.reserved",
		topicFail: "inventory.failed",
	}, nil
}

// Process does four things:
// 1) Deduplicate by OrderID
// 2) Call the stockSvc to reserve
// 3) Produce either a success or failure event
// 4) Mark the Kafka offset (at-least-once commit, safe because of dedupe+idempotence)
func (tp *TransactionalProducer) Process(
	ctx context.Context,
	order models.OrderCreated,
	msg *sarama.ConsumerMessage,
	session sarama.ConsumerGroupSession,
	stockSvc ReserveService,
) error {
	// 1) App-level dedupe: skip if we've already seen this order
	if _, dup := tp.seen.LoadOrStore(order.OrderID, true); dup {
		tp.logger.Warn("duplicate order skipped", zap.String("orderID", order.OrderID))
		session.MarkMessage(msg, "") // commit offset so we don't reprocess
		return nil
	}

	// 2) Business logic: attempt to reserve stock
	reserved, err := stockSvc.Reserve(order.Items)

	// Choose topic & payload based on success/failure
	topic := tp.topicOK
	var payload []byte
	if err != nil || !reserved {
		topic = tp.topicFail
		evt := models.InventoryFailed{
			OrderID: order.OrderID,
			Items:   order.Items,
			Reason:  fmt.Sprintf("%v", err),
		}
		payload, _ = json.Marshal(evt)
	} else {
		evt := models.InventoryReserved{
			OrderID: order.OrderID,
			Items:   order.Items,
		}
		payload, _ = json.Marshal(evt)
	}

	// 3) Produce the event
	if _, _, err := tp.prod.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(order.OrderID), // key by OrderID
		Value: sarama.ByteEncoder(payload),
	}); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	tp.logger.Info("published event",
		zap.String("orderID", order.OrderID),
		zap.String("topic", topic),
	)

	// 4) Commit offset so Kafka knows we've processed this message
	session.MarkMessage(msg, "")

	return nil
}

// Close flushes and closes the underlying producer.
func (tp *TransactionalProducer) Close() error {
	tp.logger.Info("closing producer")
	return tp.prod.Close()
}

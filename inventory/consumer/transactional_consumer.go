package consumer

import (
	"context"
	"encoding/json"

	"e-commerce/common/models"
	"e-commerce/inventory/producer"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// TxConsumer reads orders.created, hands them off to the producer,
// and marks offsets only after Process returns (durable at-least-once).
type TxConsumer struct {
	group    sarama.ConsumerGroup
	producer *producer.TransactionalProducer
	stockSvc producer.ReserveService
	logger   *zap.Logger
}

// NewTxConsumer builds a Kafka consumer group instance.
func NewTxConsumer(
	brokers []string,
	groupID string,
	stockSvc producer.ReserveService,
	prod *producer.TransactionalProducer,
	logger *zap.Logger,
) (*TxConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_5_0_0
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	grp, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &TxConsumer{
		group:    grp,
		producer: prod,
		stockSvc: stockSvc,
		logger:   logger,
	}, nil
}

// Setup is invoked when a new session starts. We don’t need to do anything.
func (c *TxConsumer) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup is invoked at the end of a session. No action needed.
func (c *TxConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim is where all the message handling happens.
func (c *TxConsumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	// Loop over messages
	for msg := range claim.Messages() {
		// Deserialize the OrderCreated event
		var order models.OrderCreated
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			c.logger.Warn("invalid payload", zap.Error(err))
			session.MarkMessage(msg, "")
			continue
		}

		// Delegate to our transactional producer
		if err := c.producer.Process(
			session.Context(), order, msg, session, c.stockSvc,
		); err != nil {
			c.logger.Error("processing failed", zap.Error(err), zap.String("orderID", order.OrderID))
		}
	}
	return nil
}

// Close shuts down the consumer group.
func (c *TxConsumer) Close() error {
	return c.group.Close()
}

// Run kicks off the consume loop against the "orders.created" topic.
// It handles rebalance and will exit when ctx is canceled.
func (c *TxConsumer) Run(ctx context.Context) {
	topics := []string{"orders.created"}
	for {
		if err := c.group.Consume(ctx, topics, c); err != nil {
			c.logger.Error("consume error", zap.Error(err))
		}
		if ctx.Err() != nil {
			return
		}
	}
}

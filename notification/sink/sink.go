package sink

import (
	"errors"
	"sync"
	"time"

	"e-commerce/common/models"

	"go.uber.org/zap"
)

// NotificationSink defines delivery methods.
type NotificationSink interface {
	NotifyReserved(models.InventoryReserved) error
	NotifyFailed(models.InventoryFailed) error
}

// ConsoleSink logs to stdout.
type ConsoleSink struct {
	logger *zap.Logger
}

func NewConsoleSink(log *zap.Logger) *ConsoleSink {
	return &ConsoleSink{logger: log}
}

func (s *ConsoleSink) NotifyReserved(evt models.InventoryReserved) error {
	s.logger.Info("✔️ Reservation notification",
		zap.String("orderID", evt.OrderID),
		zap.Any("items", evt.Items),
	)
	return nil
}

func (s *ConsoleSink) NotifyFailed(evt models.InventoryFailed) error {
	s.logger.Info("❌ Failure notification",
		zap.String("orderID", evt.OrderID),
		zap.Any("items", evt.Items),
		zap.String("reason", evt.Reason),
	)
	return nil
}

// RetryDedupeSink wraps another sink to add retry & dedupe.
type RetryDedupeSink struct {
	inner    NotificationSink
	logger   *zap.Logger
	seenKeys sync.Map
}

func NewRetryDedupeSink(inner NotificationSink, log *zap.Logger) *RetryDedupeSink {
	return &RetryDedupeSink{inner: inner, logger: log}
}

// NotifyReserved with retry & idempotency
func (r *RetryDedupeSink) NotifyReserved(evt models.InventoryReserved) error {
	if _, loaded := r.seenKeys.LoadOrStore(evt.OrderID, true); loaded {
		r.logger.Debug("Duplicate notification skipped", zap.String("orderID", evt.OrderID))
		return nil
	}
	return r.retryNotify(func() error {
		return r.inner.NotifyReserved(evt)
	}, "reserved", evt.OrderID)
}

// NotifyFailed with retry & idempotency
func (r *RetryDedupeSink) NotifyFailed(evt models.InventoryFailed) error {
	if _, loaded := r.seenKeys.LoadOrStore(evt.OrderID, true); loaded {
		r.logger.Debug("Duplicate notification skipped", zap.String("orderID", evt.OrderID))
		return nil
	}
	return r.retryNotify(func() error {
		return r.inner.NotifyFailed(evt)
	}, "failed", evt.OrderID)
}

// retryNotify runs the callback up to 3 times with backoff.
func (r *RetryDedupeSink) retryNotify(fn func() error, typ, orderID string) error {
	backoff := 100 * time.Millisecond
	for i := 1; i <= 3; i++ {
		if err := fn(); err != nil {
			r.logger.Warn("Notification failed, retrying",
				zap.String("type", typ),
				zap.String("orderID", orderID),
				zap.Error(err),
				zap.Int("attempt", i),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		r.logger.Info("Notification delivered",
			zap.String("type", typ),
			zap.String("orderID", orderID),
		)
		return nil
	}
	return errors.New("notification failed after retries: " + orderID)
}

package main

import (
	"context"
	"encoding/json"
	"time"

	"e-commerce/common/config"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// OrderCreated is a superset of both V1 & V2.
// We only care about counting messages here.
type OrderCreated struct {
	OrderID   string   `json:"orderID"`
	UserID    string   `json:"userID"`
	Items     []string `json:"items"`
	Total     float64  `json:"total"`
	PromoCode *string  `json:"promoCode,omitempty"`
}

// Metric is emitted once per-minute.
type Metric struct {
	WindowStart time.Time `json:"window_start"`
	Count       int       `json:"count"`
}

func main() {
	// 1. Load shared config (common/config)
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 2. Init Zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("Starting Aggregator", "env", cfg.Env, "brokers", cfg.KafkaBrokers)

	// 3. Kafka reader on orders.created
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   "orders.created",
		GroupID: "aggregator-group",
	})
	defer reader.Close()

	// 4. Kafka writer for metrics.order.rate
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    "metrics.order.rate",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// 5. Counting loop
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	count := 0
	windowStart := time.Now().UTC().Truncate(time.Minute)
	ctx := context.Background()

	// 5a. Consume in background
	go func() {
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				sugar.Warnw("read order failed", "error", err)
				continue
			}
			var oc OrderCreated
			if err := json.Unmarshal(m.Value, &oc); err != nil {
				sugar.Warnw("invalid order payload", "error", err)
				continue
			}
			count++
		}
	}()

	// 5b. Publish metrics each minute
	for range ticker.C {
		metric := Metric{WindowStart: windowStart, Count: count}
		data, _ := json.Marshal(metric)
		err := writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(windowStart.Format(time.RFC3339)),
				Value: data,
			},
		)
		if err != nil {
			sugar.Errorw("failed to write metric", "error", err)
		} else {
			sugar.Infow("published metric", "window_start", windowStart, "count", count)
		}
		// reset
		windowStart = time.Now().UTC().Truncate(time.Minute)
		count = 0
	}
}

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/common/logger"
	"e-commerce/inventory/consumer"
	"e-commerce/inventory/producer"
	"e-commerce/inventory/service"

	"go.uber.org/zap"
)

func main() {
	// 1) Load configuration (profiles: dev/prod) & init structured logger
	cfg, _ := config.Load()
	log, _ := logger.NewLogger(cfg.Env, cfg.LogLevel)
	defer log.Sync()
	log.Info("Starting Inventory Service", zap.String("env", cfg.Env))

	// 2) Initialize in-memory stock levels
	stockSvc := service.NewStockService(map[string]int{"foo": 10, "bar": 5})

	// 3) Create our idempotent, deduping producer
	prod, err := producer.NewTransactionalProducer(cfg.KafkaBrokers, "inv-client-1", log)
	if err != nil {
		log.Fatal("producer init failed", zap.Error(err))
	}
	defer prod.Close()

	// 4) Create the consumer group
	cons, err := consumer.NewTxConsumer(cfg.KafkaBrokers, "inventory-group", stockSvc, prod, log)
	if err != nil {
		log.Fatal("consumer init failed", zap.Error(err))
	}
	defer cons.Close()

	// 5) Run consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	go cons.Run(ctx)

	// 6) Wait for OS signal (Ctrl+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 7) Begin shutdown
	log.Info("Shutdown signal received")
	cancel()
	time.Sleep(5 * time.Second)
	log.Info("Exited cleanly")
}

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
	// 1. Load config & logger
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log, err := logger.NewLogger(cfg.Env, cfg.LogLevel)
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer log.Sync()
	log.Info("Starting Inventory Service", zap.String("env", cfg.Env))

	// 2. Stock service
	initialStock := map[string]int{"foo": 10, "bar": 5}
	stockSvc := service.NewStockService(initialStock)

	// 3. Producer
	invProd := producer.NewInventoryProducer(cfg.KafkaBrokers, log)
	defer invProd.Close()

	// 4. Consumer
	invCons := consumer.NewInventoryConsumer(cfg.KafkaBrokers, "inventory-group", stockSvc, invProd, log)
	defer invCons.Close()

	// 5. Run consumer in background with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	go invCons.Run(ctx)

	// 6. Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Info("Shutdown signal received, closing...")

	// 7. Cancel & allow in-flight work up to 5s
	cancel()
	time.Sleep(5 * time.Second)
	log.Info("Inventory Service exited cleanly")
}

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/inventory/consumer"
	"e-commerce/inventory/producer"
	"e-commerce/inventory/service"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Initialize stock levels (e.g., 10 units of "foo" and "bar")
	initialStock := map[string]int{"foo": 10, "bar": 5}
	stockSvc := service.NewStockService(initialStock)

	// 3. Set up producer
	invProd := producer.NewInventoryProducer(cfg.KafkaBrokers)
	defer invProd.Close()

	// 4. Set up consumer group
	invCons := consumer.NewInventoryConsumer(cfg.KafkaBrokers, "inventory-group", stockSvc, invProd)
	defer invCons.Close()

	// 5. Run consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	go invCons.Run(ctx)

	// 6. Wait for OS signal to shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutdown signal received…")

	// 7. Give in-flight work up to 5s to finish
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("Inventory service exiting")
}

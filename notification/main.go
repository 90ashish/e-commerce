package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/notification/consumer"
	"e-commerce/notification/sink"
)

func main() {
	// 1. Load shared configuration (Kafka brokers, etc.)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Choose a sink implementation
	notifSink := sink.NewConsoleSink()
	// If you want email in future:
	// notifSink := sink.NewEmailSink()

	// 3. Initialize the notification consumer
	notifCons := consumer.NewNotificationConsumer(cfg.KafkaBrokers, "notification-group", notifSink)
	defer notifCons.Close()

	// 4. Run the consumer loop in the background
	ctx, cancel := context.WithCancel(context.Background())
	go notifCons.Run(ctx)

	// 5. Wait for OS interrupt (Ctrl+C) to shut down
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("🛑 Shutdown signal received for Notification service")

	// 6. Cancel context, give 5s for cleanup, then exit
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("👋 Notification service exiting")
}

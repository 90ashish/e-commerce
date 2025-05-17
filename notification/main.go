package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/common/logger"
	"e-commerce/notification/consumer"
	"e-commerce/notification/sink"

	"go.uber.org/zap"
)

func main() {
	// 1. Load config + logger
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log, err := logger.NewLogger(cfg.Env, cfg.LogLevel)
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer log.Sync()
	log.Info("Starting Notification Service", zap.String("env", cfg.Env))

	// 2. Choose sink (console by default)
	baseSink := sink.NewConsoleSink(log)
	// Wrap with retry and dedupe
	notifSink := sink.NewRetryDedupeSink(baseSink, log)

	// 3. Initialize consumer
	notifCons := consumer.NewNotificationConsumer(cfg.KafkaBrokers, "notification-group", notifSink, log)
	defer notifCons.Close()

	// 4. Run consumer
	ctx, cancel := context.WithCancel(context.Background())
	go notifCons.Run(ctx)

	// 5. Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Info("Shutdown signal received, exiting...")

	// 6. Cancel and wait for in-flight work
	cancel()
	time.Sleep(5 * time.Second)
	log.Info("Notification Service shut down cleanly")
}

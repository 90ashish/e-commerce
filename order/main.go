package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/common/logger"
	"e-commerce/order/handler"
	"e-commerce/order/producer"

	"github.com/gin-gonic/gin"
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
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()
	log.Info("Starting Order Service", zap.String("env", cfg.Env))

	// 2. Prepare Gin with zap middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(logger.GinZapMiddleware(log), gin.Recovery())

	// 3. Kafka producer with retry & idempotency
	kp := producer.NewKafkaProducer(cfg.KafkaBrokers, "orders.created", log)
	defer kp.Close()

	// 4. Register handler
	h := handler.NewOrderHandler(kp, log)
	router.POST("/orders", h.CreateOrder)

	// 5. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":8090",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server listen failed", zap.Error(err))
		}
	}()
	log.Info("HTTP server started on :8090")

	// 6. Wait for SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown signal received, shutting down...")

	// 7. Shutdown with 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}
	log.Info("Order Service exited cleanly")
}

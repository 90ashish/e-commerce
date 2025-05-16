package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-commerce/common/config"
	"e-commerce/order/handler"
	"e-commerce/order/producer"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration (Kafka brokers, etc.)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize the Kafka producer
	kp := producer.NewKafkaProducer(cfg.KafkaBrokers, "orders.created")
	defer kp.Close()

	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Register the order handler
	orderHandler := handler.NewOrderHandler(kp)
	router.POST("/orders", orderHandler.CreateOrder)

	// Create HTTP server in a goroutine
	srv := &http.Server{
		Addr:    ":8090",
		Handler: router,
	}
	go func() {
		log.Println("Order service listening on :8090")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("Server exited cleanly")
}

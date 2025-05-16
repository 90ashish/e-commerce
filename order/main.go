package main

import (
	"e-commerce/common/config"
	"e-commerce/order/handler"
	"e-commerce/order/producer"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load shared config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize Kafka producer
	kp := producer.NewKafkaProducer(cfg.KafkaBrokers, "orders.created")
	defer kp.Close()

	// Initialize Gin router
	router := gin.Default()
	orderHandler := handler.NewOrderHandler(kp)
	router.POST("/orders", orderHandler.CreateOrder)

	// Start HTTP server
	log.Println("Order service running on :8090")
	if err := router.Run(":8090"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

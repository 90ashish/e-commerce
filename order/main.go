package main

import (
	"context"
	"e-commerce/common/config"
	"e-commerce/common/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// Set up Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   "orders.created",
	})
	defer writer.Close()

	// HTTP handler for creating orders
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var event models.OrderCreated
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "bad JSON format", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "serialization error", http.StatusInternalServerError)
			return
		}

		// Publish to Kafka
		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(event.UserID),
				Value: data,
			},
		)
		if err != nil {
			log.Printf("failed to publish event: %v", err)
			http.Error(w, "failed to publish event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Order received")
	})

	log.Println("Order service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

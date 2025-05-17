# Makefile for e-commerce Kafka + Go microservices

# Docker Compose command
DC := docker-compose

# Kafka container name (as per your compose file)
KAFKA := kafka

# Kafka broker address
BROKER := localhost:9092

# Go module directories
ORDER_DIR := order
INV_DIR   := inventory

.PHONY: help up down topics order inventory services

help:  ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

up:  ## Start Zookeeper, Kafka & Schema Registry
	$(DC) up -d

down:  ## Stop & remove all containers
	$(DC) down

topics:  ## Create the three required topics
	@echo "→ Creating topics on $(BROKER)..."
	@docker exec -it $(KAFKA) kafka-topics.sh \
		--bootstrap-server $(BROKER) \
		--create --topic orders.created \
		--partitions 3 --replication-factor 1 || true
	@docker exec -it $(KAFKA) kafka-topics.sh \
		--bootstrap-server $(BROKER) \
		--create --topic inventory.reserved \
		--partitions 3 --replication-factor 1 || true
	@docker exec -it $(KAFKA) kafka-topics.sh \
		--bootstrap-server $(BROKER) \
		--create --topic inventory.failed \
		--partitions 3 --replication-factor 1 || true
	@echo "→ Topics:"
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) --list

show-topics: ## Shows topics created in kafka
	@echo "→ Topics created inside kafka:"
	@docker exec -it kafka kafka-topics.sh --list --bootstrap-server localhost:9092

order:  ## Run the Order service (blocking)
	@echo "→ Starting Order service on :8090"
	@cd $(ORDER_DIR) && go run main.go

inventory:  ## Run the Inventory service (blocking)
	@echo "→ Starting Inventory service"
	@cd $(INV_DIR) && go run main.go

services: up topics  ## Start everything you need (then run services separately)
	@echo "→ Kafka stack and topics ready. Now run 'make order' and 'make inventory' in separate shells."


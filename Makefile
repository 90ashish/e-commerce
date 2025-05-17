# Makefile for e-commerce Kafka + Go microservices

# Docker Compose command
DC := docker-compose

# Kafka container name (as per your compose file)
KAFKA := kafka

# Kafka broker address (used by kafka-topics.sh)
BROKER := localhost:9092

.PHONY: help up down build-services topics show-topics \
        show-inventory-reservations show-inventory-failures

help:  ## Show this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

up: ## Bring up Kafka stack, create topics, build & run Go services
	@echo "→ Starting Zookeeper, Kafka & Schema Registry..."
	$(DC) up -d zookeeper kafka schema-registry

	@echo "→ Creating required Kafka topics..."
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

	@echo "→ Building & launching Order, Inventory & Notification services..."
	$(DC) build order-service inventory-service notification-service
	$(DC) up -d order-service inventory-service notification-service

	@echo "→ All services are now up and running."

down: ## Stop & remove all containers
	@echo "→ Tearing down all services..."
	$(DC) down

show-topics: ## List all Kafka topics
	@echo "→ Topics in Kafka:"
	@docker exec -it $(KAFKA) kafka-topics.sh \
		--bootstrap-server $(BROKER) --list

show-inventory-reservations: ## Listen for successful inventory reservations
	@echo "→ Listening on inventory.reserved:"
	@docker exec -it $(KAFKA) kafka-console-consumer.sh \
		--bootstrap-server $(BROKER) \
		--topic inventory.reserved \
		--from-beginning

show-inventory-failures: ## Listen for inventory failures
	@echo "→ Listening on inventory.failed:"
	@docker exec -it $(KAFKA) kafka-console-consumer.sh \
		--bootstrap-server $(BROKER) \
		--topic inventory.failed \
		--from-beginning
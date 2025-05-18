# Makefile for e-commerce Kafka + Go microservices

DC := docker-compose
KAFKA := kafka
BROKER := localhost:9092

.PHONY: help up down build-services topics show-topics \
        show-inventory-reservations show-inventory-failures scale-inventory

help:  ## Show this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN{FS=":.*?## "}{printf "\033[36m%-30s\033[0m %s\n",$$1,$$2}'

up: ## Bring up all infra and services, scale inventory to 2
	@echo "→ Starting infra..."
	$(DC) up -d zookeeper kafka schema-registry

	@echo "→ Recreating topics with 4 partitions..."
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--delete --topic orders.created || true
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--create --topic orders.created --partitions 4 --replication-factor 1
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--delete --topic inventory.reserved || true
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--create --topic inventory.reserved --partitions 4 --replication-factor 1
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--delete --topic inventory.failed || true
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
		--create --topic inventory.failed --partitions 4 --replication-factor 1

	@echo "→ Building service images..."
	$(DC) build order-service inventory-service notification-service

	@echo "→ Launching services..."
	$(DC) up -d order-service notification-service

	@echo "→ Scaling inventory-service to 2 instances..."
	$(DC) up -d --scale inventory-service=2

	@echo "→ All services running (inventory-service x2)."

down:  ## Stop & remove all containers
	@echo "→ Tearing down all services..."
	$(DC) down

show-topics:  ## Describe Kafka topics
	@echo "→ Kafka topics and partitions:"
	@docker exec -it $(KAFKA) kafka-topics.sh \
		--bootstrap-server $(BROKER) --describe

show-inventory-reservations:  ## Listen on inventory.reserved
	@echo "→ Listening on inventory.reserved:"
	@docker exec -it $(KAFKA) kafka-console-consumer.sh \
		--bootstrap-server $(BROKER) \
		--topic inventory.reserved --from-beginning

show-inventory-failures:  ## Listen on inventory.failed
	@echo "→ Listening on inventory.failed:"
	@docker exec -it $(KAFKA) kafka-console-consumer.sh \
		--bootstrap-server $(BROKER) \
		--topic inventory.failed --from-beginning

## Load Testing :=
run-load-test: ## Runs the load test script, Install k6 on your machine before running it
	@echo "-> running load testing..."
	@k6 run load-test.js

measure-consumer-lag: ## While load-testing, watch consumer lag with Kafka’s built-in tool
	@echo "-> Shows consumer lag:"
	@docker exec -it kafka kafka-consumer-groups.sh \
	--bootstrap-server kafka:9092 \
	--describe \
	--group inventory-group

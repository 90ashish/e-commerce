# Makefile for e-commerce Kafka + Go microservices

DC := docker-compose
KAFKA := kafka
BROKER := localhost:9092

.PHONY: help up down build-services topics show-topics \
        show-inventory-reservations show-inventory-failures scale-inventory \
		run-load-test measure-consumer-lag \
		register-schema-v1 register-schema-v2 get-schema-versions \
		gen-models-v1 gen-models-v2 gen-models \
		show-metrics

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

	@echo "→ Creating metrics.order.rate topic..."
	@docker exec -it $(KAFKA) kafka-topics.sh --bootstrap-server $(BROKER) \
	    --create --topic metrics.order.rate --partitions 4 --replication-factor 1 || true

	@echo "→ Building service images..."
	$(DC) build order-service inventory-service notification-service aggregator

	@echo "→ Launching services..."
	$(DC) up -d order-service notification-service aggregator

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

# ---------------------------------------------------------------------------
# Register Schemas in Schema Registry & Avro Code Generation
# ---------------------------------------------------------------------------

register-schema-v1: ## Register order_created V1 schema
	@echo "→ Registering V1 schema"
	@curl -X POST \
	  -H "Content-Type: application/vnd.schemaregistry.v1+json" \
	  --data "$$(jq -Rs '{schema:.}' schemas/order_created_v1.avsc)" \
	  http://localhost:8081/subjects/orders.created-value/versions

register-schema-v2: ## Register order_created V2 schema
	@echo "→ Registering V2 schema"
	@curl -X POST \
	  -H "Content-Type: application/vnd.schemaregistry.v1+json" \
	  --data "$$(jq -Rs '{schema:.}' schemas/order_created_v2.avsc)" \
	  http://localhost:8081/subjects/orders.created-value/versions

get-schema-versions: ## List all versions for orders.created-value
	@echo "→ Schema versions:"
	@curl -s http://localhost:8081/subjects/orders.created-value/versions

# Generate Go types for OrderCreated V1
gen-models-v1: ## Generate Go structs from order_created_v1.avsc
	@echo "→ Generating Go types for OrderCreated V1"
	@mkdir -p common/models/v1
	@gogen-avro \
	  -package models_v1 \
	  common/models/v1 \
	  schemas/order_created_v1.avsc

# Generate Go types for OrderCreated V2
gen-models-v2: ## Generate Go structs from order_created_v2.avsc
	@echo "→ Generating Go types for OrderCreated V2"
	@mkdir -p common/models/v2
	@gogen-avro \
	  -package models_v2 \
	  common/models/v2 \
	  schemas/order_created_v2.avsc

# Convenience: regenerate both V1 and V2 models
gen-models: gen-models-v1 gen-models-v2 ## Generate all Avro-based Go types

show-metrics: ## Listen to the metrics.order.rate topic from the beginning
	@echo "→ Listening on metrics.order.rate (print key)..."
	@docker exec -it $(KAFKA) \
	  kafka-console-consumer.sh \
	    --bootstrap-server $(BROKER) \
	    --topic metrics.order.rate \
	    --from-beginning \
	    --property print.key=true
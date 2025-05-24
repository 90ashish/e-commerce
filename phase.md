# Kafka-based Go Microservices Progress Tracker

## Phase 1 – Local Infrastructure & Boilerplate 【✔️ done】

**Goal:** Stand up a local Kafka ecosystem and blank Go-service scaffold.  
**What we did:**

- Wrote a `docker-compose.yml` with Zookeeper, Kafka broker, Schema Registry (optional).
- Created the three topics (`orders.created`, `inventory.reserved`, `inventory.failed`).
- Set up a mono-repo: `order/`, `inventory/`, `notification/` + `common/` for models & config.
- Initialized `go.mod` at repo root and loaded env via Viper.

---

## Phase 2 – Order Service (Producer) 【✔️ done】

**Goal:** Expose HTTP `POST /orders` and publish `orders.created`.  
**What we did:**

- Built a Gin-based API with request validation for `userID`, `items`, `total`.
- Integrated `github.com/segmentio/kafka-go` / Sarama to produce to `orders.created`, keyed by `userID`.
- Added error handling, structured logs, and graceful shutdown.
- *(Unit tests to be written later.)*

---

## Phase 3 – Inventory Service (Consumer → Producer) 【✔️ done】

**Goal:** Consume orders, reserve stock, emit success/failure events.  
**What we did:**

- Created an `InventoryConsumer` joining consumer-group on `orders.created`.
- Wrote `StockService` (in-memory map) to reserve items.
- On success, produced to `inventory.reserved`; on failure, to `inventory.failed`.
- Committed offsets only after produce succeeded.
- Wrote integration test (order → inventory round-trip).

---

## Phase 4 – Notification Service (Consumer) 【✔️ done】

**Goal:** Notify user of confirm/out-of-stock.  
**What we did:**

- Built `NotificationConsumer` with two readers (`inventory.reserved`, `inventory.failed`).
- Defined `NotificationSink` interface, with console & email stubs.
- Logging hooks for “Order confirmed” & “Out of stock.”
- End-to-end tests: submitted orders, asserted notifications.

---

## Phase 5 – Configuration & Reliability 【✔️ done】

**Goal:** Add production-grade cross-cutting concerns.  
**What we did:**

- Centralized all config via Viper—with dev/prod profiles.
- Replaced stdlib logging with Zap; added Gin-Zap middleware.
- Graceful shutdown (contexts + SIGINT).
- Producer/consumer retry & app-level idempotency (dedupe maps).
- Structured logs with trace IDs.

---

## Phase 6 – Scaling & Partitioning 【✔️ done】

**Goal:** Horizontally scale and partition correctly.  
**What we did:**

- Recreated topics with 4+ partitions.
- Scaled inventory-service to 2+ replicas in same consumer group.
- Verified exactly-once processing across restarts.
- Load-tested with k6, measured throughput and consumer lag via `kafka-consumer-groups.sh`.
- Confirmed ordering by `userID` keying.

---

## Phase 7 – Transactions & Exactly-Once 【✔️ done】

**Goal:** Guarantee no duplicates or data loss.  
**What we did (idempotent + dedupe fallback):**

- Swapped in IBM’s Sarama v1.45.1 with idempotent `SyncProducer` (`Net.MaxOpenRequests=1`).
- Built a `TransactionalProducer` wrapper with app-level dedupe for `orderIDs`.
- *(Full Kafka TXN APIs deferred until Sarama stabilizes.)*
- Ensured each order is published exactly once and offsets committed after produce.

---

## Phase 8 – Schema Evolution & Stream Processing [todo]

**Goal:** Support versioned schemas and aggregate streams.  
**To do:**

- Bring up Schema Registry, define Avro/JSON schema for `OrderCreatedV1 → V2` (add a field).
- Code-generate Go structs, wire into producer & consumers.
- Build a metrics “aggregator” service: subscribe to `orders.created`, compute QPS/minute, publish to `metrics.order.rate`.
- Enhance `NotificationConsumer` to handle mixed V1/V2 payloads.

---

## Phase 9 – Connectors & Monitoring [todo]

**Goal:** Integrate Kafka Connect and observability.  
**To do:**

- Stand up Kafka Connect with a Postgres/Mongo sink connector for audit logs.
- Expose broker & consumer metrics via Prometheus (JMX exporter or HTTP).
- Create Grafana dashboards for consumer lag, QPS, error rates.
- Define alerts for lag spikes and failure conditions.
# Kafka-based Go Microservices Progress Tracker

## Phase 1 – Local Infrastructure & Boilerplate 【✔️ done】

**Goal:** Stand up a local Kafka ecosystem and blank Go-service scaffold.  
**What we did:**

- Wrote a `docker-compose.yml` with Zookeeper, Kafka broker, Schema Registry (optional).
- Created the three topics (`orders.created`, `inventory.reserved`, `inventory.failed`).
- Set up a mono-repo: `order/`, `inventory/`, `notification/` + `common/` for models & config.
- Initialized `go.mod` at repo root and loaded env via Viper.

---

## Phase 2 – Order Service (Producer) 【✔️ done】

**Goal:** Expose HTTP `POST /orders` and publish `orders.created`.  
**What we did:**

- Built a Gin-based API with request validation for `userID`, `items`, `total`.
- Integrated `github.com/segmentio/kafka-go` / Sarama to produce to `orders.created`, keyed by `userID`.
- Added error handling, structured logs, and graceful shutdown.
- *(Unit tests to be written later.)*

---

## Phase 3 – Inventory Service (Consumer → Producer) 【✔️ done】

**Goal:** Consume orders, reserve stock, emit success/failure events.  
**What we did:**

- Created an `InventoryConsumer` joining consumer-group on `orders.created`.
- Wrote `StockService` (in-memory map) to reserve items.
- On success, produced to `inventory.reserved`; on failure, to `inventory.failed`.
- Committed offsets only after produce succeeded.
- Wrote integration test (order → inventory round-trip).

---

## Phase 4 – Notification Service (Consumer) 【✔️ done】

**Goal:** Notify user of confirm/out-of-stock.  
**What we did:**

- Built `NotificationConsumer` with two readers (`inventory.reserved`, `inventory.failed`).
- Defined `NotificationSink` interface, with console & email stubs.
- Logging hooks for “Order confirmed” & “Out of stock.”
- End-to-end tests: submitted orders, asserted notifications.

---

## Phase 5 – Configuration & Reliability 【✔️ done】

**Goal:** Add production-grade cross-cutting concerns.  
**What we did:**

- Centralized all config via Viper—with dev/prod profiles.
- Replaced stdlib logging with Zap; added Gin-Zap middleware.
- Graceful shutdown (contexts + SIGINT).
- Producer/consumer retry & app-level idempotency (dedupe maps).
- Structured logs with trace IDs.

---

## Phase 6 – Scaling & Partitioning 【✔️ done】

**Goal:** Horizontally scale and partition correctly.  
**What we did:**

- Recreated topics with 4+ partitions.
- Scaled inventory-service to 2+ replicas in same consumer group.
- Verified exactly-once processing across restarts.
- Load-tested with k6, measured throughput and consumer lag via `kafka-consumer-groups.sh`.
- Confirmed ordering by `userID` keying.

---

## Phase 7 – Transactions & Exactly-Once 【✔️ done】

**Goal:** Guarantee no duplicates or data loss.  
**What we did (idempotent + dedupe fallback):**

- Swapped in IBM’s Sarama v1.45.1 with idempotent `SyncProducer` (`Net.MaxOpenRequests=1`).
- Built a `TransactionalProducer` wrapper with app-level dedupe for `orderIDs`.
- *(Full Kafka TXN APIs deferred until Sarama stabilizes.)*
- Ensured each order is published exactly once and offsets committed after produce.

---

## Phase 8 – Schema Evolution & Stream Processing [todo]

**Goal:** Support versioned schemas and aggregate streams.  
**To do:**

- Bring up Schema Registry, define Avro/JSON schema for `OrderCreatedV1 → V2` (add a field).
- Code-generate Go structs, wire into producer & consumers.
- Build a metrics “aggregator” service: subscribe to `orders.created`, compute QPS/minute, publish to `metrics.order.rate`.
- Enhance `NotificationConsumer` to handle mixed V1/V2 payloads.

---

## Phase 9 – Connectors & Monitoring [todo]

**Goal:** Integrate Kafka Connect and observability.  
**To do:**

- Stand up Kafka Connect with a Postgres/Mongo sink connector for audit logs.
- Expose broker & consumer metrics via Prometheus (JMX exporter or HTTP).
- Create Grafana dashboards for consumer lag, QPS, error rates.
- Define alerts for lag spikes and failure conditions.

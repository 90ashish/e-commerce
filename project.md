# Event-Driven E-Commerce Order Processing (Kafka + Golang)

## 1. Project Idea: Amazon-Style Order Pipeline

Build a mini e-commerce backend with **three Go microservices**, communicating **exclusively via Kafka**.

### 🧱 Microservices Overview

- **Order Service (Producer)**
  - Exposes a REST endpoint to create orders.
  - Publishes an `orders.created` event with order payload.

- **Inventory Service (Consumer → Producer)**
  - Subscribes to `orders.created`.
  - Checks/updates stock (in-memory or simple DB).
  - Publishes either `inventory.reserved` or `inventory.failed`.

- **Notification Service (Consumer)**
  - Listens on `inventory.reserved` and `inventory.failed`.
  - Sends “order confirmed” or “out of stock” notifications (console log or email stub).

---

## 2. Kafka Features You’ll Master

- **Topic partitioning & keying**  
  Ensure orders from the same user always go to the same partition for ordering guarantees.

- **Consumer groups**  
  Horizontally scale inventory or notification services.

- **Offset management**  
  Precisely commit offsets only after successful processing.

- **Exactly-once semantics (optional)**  
  Use Kafka transactions to avoid double-reservation of stock.

- **Schema evolution**  
  Integrate with **Confluent Schema Registry** to version `OrderCreatedV1 → V2` payloads.

---

## 3. Golang Tooling

- **Client Libraries**
  - [`segmentio/kafka-go`](https://github.com/segmentio/kafka-go): Idiomatic and simple.
  - [`Shopify/sarama`](https://github.com/IBM/sarama): Rich features, supports transactions and interceptors.

- **Configuration**
  - Use `Viper` or environment variables to externalize brokers, topics, and group IDs.

- **Graceful Shutdown**
  - Implement `context.Context` + signal handling to cleanly close consumers/producers and flush offsets.

---

## 4. Extensions for Deeper Learning

- **Stream Processing**
  - Build a Go app that consumes `orders.created`, aggregates order count per minute, and publishes to a `metrics.order.rate` topic.

- **Kafka Connect**
  - Export Kafka events to MongoDB or PostgreSQL using Kafka Connect (no Go code needed).

- **Monitoring**
  - Integrate **Prometheus** + JMX/HTTP exporters for consumer lag, broker health, etc.

- **Idempotency & Retries**
  - Design retry-safe services using deduplication maps and at-least-once semantics.

---

## ✅ What You'll Learn

This hands-on project covers:

- Producers & consumers
- Partitions & keying
- Offset commit strategies
- Error handling patterns
- Schema evolution
- Idempotency
- Kafka transactions (optional)

> You'll come away with **real-world Kafka skills in a Go microservices architecture** that mimic production-grade pipelines.

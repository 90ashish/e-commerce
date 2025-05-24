Phase 1 – Local Infrastructure & Boilerplate [done]

Goal: Stand up a local Kafka ecosystem and a blank Go-service template.

    Docker Compose:

        Kafka broker + Zookeeper

        (Optional) Confluent Schema Registry

    Topics:

        orders.created

        inventory.reserved

        inventory.failed

    Go Service Scaffold:

        Mono-repo layout with three service folders: order/, inventory/, notification/

        Shared common/ for models (e.g. Order struct) and config

        go.mod in repo root, viper/envconfig for env vars

Phase 2 – Order Service (Producer) [done]

Goal: Expose a REST endpoint and publish orders.created events.

    HTTP API:

        POST /orders → accepts JSON payload (userID, items, total)

        Basic request validation (e.g. non-empty items)

    Kafka Producer:

        Use kafka-go or sarama to publish to orders.created

        Key by userID for partition affinity

    Unit Tests:

        Mock producer to assert message format & topic

Phase 3 – Inventory Service (Consumer → Producer) [done]

Goal: Consume new orders, check stock, emit success/failure.

    Kafka Consumer:

        Join orders-created consumer group

        Deserialize payload

    Stock Logic:

        In-memory map or simple Redis/Postgres table

        Reserve if available ≥ quantity, else mark fail

    Emit Events:

        On success → inventory.reserved (include orderID)

        On failure → inventory.failed

    Offset Handling:

        Commit only after produce succeeds

    Integration Test:

        Bring up Order + Inventory → verify round-trip

Phase 4 – Notification Service (Consumer) [done]

Goal: Send out confirmation or out-of-stock notifications.

    Dual Consumers:

        One group subscribes both result topics

    Notification Logic:

        On reserved → log/email “Order confirmed”

        On failed → log/email “Out of stock”

    Pluggable Sink:

        Write interface so you can swap from console to SMTP or webhook

    End-to-End Test:

        Fire an order that succeeds and one that fails; assert logs

Phase 5 – Configuration & Reliability [done]

Goal: Bolt on production-grade cross-cutting concerns.

    Graceful Shutdown:

        context.Context + SIGINT handler → drain, commit, close

    Retry & Idempotency:

        Retry transient Kafka errors

        Dedupe on orderID if replayed

    Structured Logging:

        zap or logrus with request trace IDs

    Configuration:

        Centralize via Viper or env vars; support profiles (dev/prod)

Phase 6 – Scaling & Partitioning [done]

Goal: Make it horizontally scalable and correctly partitioned.

    Consumer Groups:

        Simulate 2+ inventory instances

        Verify each message processed exactly once

    Topic Partitions:

        Increase to 4+ and test keying strategy

    Load Testing:

        Use k6 or wrk to slam POST /orders

        Measure consumer lag under load

Phase 7 – Transactions & Exactly-Once [done]

Goal: Ensure no double-reservations or lost events.

    Kafka Transactions (if using Sarama):

        Wrap consume–produce cycle in a transaction

        Commit offsets transactionally

    Error Scenarios:

        Test failures mid-transaction → assert rollback

Phase 8 – Schema Evolution & Stream Processing [todo]

Goal: Add versioned schemas and a simple stream aggregator.

    Schema Registry:

        Define Avro/JSON Schema for OrderCreatedV1 → V2 (add a field)

        Code-gen Go structs from schema

    Metrics Service:

        New Go “aggregator” app subscribes orders.created

        Computes QPS per minute, writes to metrics.order.rate

    Consumer in Notification:

        Auto-adapt to both V1 & V2 schemas

Phase 9 – Connectors & Monitoring

Goal: Integrate Kafka Connect and observability.

    Kafka Connect:

        Sink connector to Postgres or MongoDB for audit logs

    Prometheus Exporters:

        Expose consumer-lag, broker‐health via JMX or HTTP

    Alerting:

        Grafana dashboard for lag spikes, error rates
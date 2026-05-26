# ADR 001: Why Kafka (MSK) instead of RabbitMQ for the Event Bus

## Status
Accepted

## Context
The Saudi Distributed Commerce Fabric requires a reliable, high-throughput message backbone for:
- Saga orchestration across 9 microservices
- Event sourcing for audit trails
- Real-time fraud detection stream processing

We evaluated RabbitMQ, NATS, AWS SQS/SNS, and Apache Kafka.

## Decision
We chose **Amazon MSK (Apache Kafka)** as the primary event bus.

## Rationale

| Criteria | Kafka | RabbitMQ | NATS | SQS/SNS |
|----------|-------|----------|------|---------|
| Throughput | 1M+ msg/sec | 50K msg/sec | 3M+ msg/sec | 10K msg/sec |
| Event Sourcing | Native log retention | Poor | Poor | None |
| Replay capability | Yes (seek to offset) | No | JetStream limited | No |
| Schema Evolution | Schema Registry (Avro) | No native | No | No |
| Exactly-Once | Supported | Supported | At-least-once | At-least-once |
| KSA Residency | MSK available in me-south-1 | Self-hosted only | Self-hosted only | Yes |
| Operational Maturity | Battle-tested at LinkedIn/Netflix | Mature | Emerging | AWS-managed |

### Key Deciding Factors

1. **Event Sourcing Requirement**: Our Audit Service requires immutable, append-only event logs with 7-year retention. Kafka's log-based storage model is a natural fit.

2. **SAMA Compliance**: MSK provides encryption at rest (KMS) and in transit (TLS 1.3) with CloudTrail integration for audit logging.

3. **Replay for Bug Recovery**: If the Settlement Service has a bug processing a day's worth of transactions, we can replay the `order-events` topic from offset T0 rather than requesting data dumps.

4. **Schema Registry**: Prevents consumer breakage when the Order Service adds a new field to `OrderCreated`. Avro backward compatibility is enforced at the registry level.

## Consequences

### Positive
- Exactly-once semantics for payment events (critical for MADA reconciliation)
- Horizontal scalability during Ramadan flash sales
- Native stream processing with Kafka Streams for real-time fraud scoring

### Negative
- Operational complexity: requires understanding of partitions, consumer groups, and rebalancing
- Higher latency (ms) compared to RabbitMQ for simple RPC patterns
- Need for careful topic design to avoid hot partitions

## Alternatives Considered
- **RabbitMQ**: Rejected due to lack of native event sourcing and replay capabilities
- **NATS JetStream**: Promising but too new for SAMA-grade financial workloads
- **SQS/SNS**: Rejected due to 14-day max retention and no replay capability

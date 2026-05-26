# ADR 003: Why the Outbox Pattern for Event Publishing

## Status
Accepted

## Context
In distributed systems, services must atomically update their database AND publish an event to Kafka. The naive approach (dual-write) risks inconsistency:

1. DB commit succeeds, Kafka publish fails → Order exists but no one knows
2. Kafka publish succeeds, DB rollback → Ghost event

This is unacceptable for payment processing where MADA reconciliation depends on event completeness.

## Decision
We implemented the **Outbox Pattern** with Debezium CDC as the relay mechanism.

## How It Works

```
┌─────────────┐     Transaction     ┌──────────────┐
│   Order     │ ──────────────────> │   outbox     │
│   Service   │                     │   table      │
└─────────────┘                     └──────────────┘
                                           │
                                    ┌──────▼──────┐
                                    │  Debezium   │
                                    │  CDC        │
                                    └──────┬──────┘
                                           │
                                    ┌──────▼──────┐
                                    │   Kafka     │
                                    └─────────────┘
```

1. Order Service writes to `orders` table AND `outbox_events` table in a single PostgreSQL transaction
2. Debezium monitors PostgreSQL WAL (Write-Ahead Log) and captures the `outbox_events` insert
3. Debezium publishes the event to Kafka and marks the row as processed

## Rationale

| Approach | Consistency Guarantee | Complexity | SAMA Suitability |
|----------|----------------------|------------|------------------|
| Dual Write | None | Low | ❌ Rejected |
| Outbox + CDC | Strong (ACID tx) | Medium | ✅ Accepted |
| Transactional Outbox + Polling | Strong (ACID tx) | Medium | ✅ Alternative |
| Saga Compensation | Eventual | High | ⚠️ For recovery only |

### Why Not Debezium Directly on `orders` Table?
- **Event schema stability**: The `orders` table schema changes frequently (new columns for business requirements). The outbox table has a stable schema (`id`, `topic`, `key`, `payload`, `headers`).
- **Domain event abstraction**: We want to publish `OrderPaymentRequested`, not `INSERT INTO orders`.
- **Idempotency**: Outbox events carry correlation IDs and idempotency keys that Debezium doesn't natively handle.

## Consequences

### Positive
- Exactly-once event delivery (within the bounds of PostgreSQL + Kafka transactions)
- No distributed transactions (2PC) required
- Works with our Flyway migration strategy

### Negative
- Additional table to maintain and monitor
- Slight latency increase (~100-500ms) between DB commit and Kafka delivery
- Need to ensure outbox table doesn't grow unbounded (archival strategy required)

## Implementation Details
- Outbox table created via Flyway migration in each service
- Background Go routine polls unprocessed events every 2 seconds (local dev)
- Production uses Debezium PostgreSQL connector running as a Kafka Connect worker

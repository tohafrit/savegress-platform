# Introduction to Savegress

## What is Savegress?

Savegress is a high-performance Change Data Capture (CDC) platform that captures and streams database changes in real-time. It enables you to:

- **Stream changes instantly** - Capture INSERT, UPDATE, DELETE operations as they happen
- **Keep systems in sync** - Replicate data across databases, caches, and search engines
- **Build event-driven architectures** - Power microservices with real-time events
- **Enable analytics** - Feed data warehouses and lakes with fresh data

## How CDC Works

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   Application                                                            │
│       │                                                                  │
│       ▼                                                                  │
│   ┌─────────┐     ┌───────────────┐     ┌──────────────────────────┐   │
│   │   DB    │────▶│   Savegress   │────▶│    Downstream Systems    │   │
│   │ (Source)│     │    Engine     │     │  • Kafka                 │   │
│   └─────────┘     └───────────────┘     │  • Elasticsearch         │   │
│       │                   │              │  • Redis                 │   │
│       │                   │              │  • Data Warehouse        │   │
│   Transaction         CDC Event          │  • Microservices         │   │
│     Log              Streaming           └──────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

Unlike polling or application-level triggers, CDC reads directly from the database's transaction log:

| Approach | Latency | Load on DB | Data Consistency |
|----------|---------|------------|------------------|
| Polling | Seconds to minutes | High | May miss changes |
| Triggers | Low | High | Possible deadlocks |
| **CDC** | Milliseconds | **Minimal** | **Guaranteed** |

## Key Benefits

### 1. Zero Application Changes
CDC captures changes at the database level - no code modifications needed.

```sql
-- Your existing application code stays the same
INSERT INTO orders (customer_id, amount) VALUES (123, 99.99);

-- Savegress automatically captures this as an event
```

### 2. Complete Change History
Every change includes before and after states:

```json
{
  "operation": "UPDATE",
  "table": "orders",
  "before": { "status": "pending", "amount": 99.99 },
  "after": { "status": "shipped", "amount": 99.99 }
}
```

### 3. Transaction Ordering
Changes are streamed in exact transaction commit order, preserving data integrity.

### 4. Minimal Database Impact
CDC reads from the transaction log, not the live tables - typically <1% overhead.

## Architecture

Savegress consists of two main components:

### Engine
Connects to your source database and captures changes:
- Reads transaction logs (WAL, binlog, etc.)
- Parses and transforms events
- Handles schema changes
- Manages checkpoints

### Broker
Receives events from engines and delivers to consumers:
- High-throughput message storage
- Reliable delivery guarantees
- Consumer group management
- Dead letter queue handling

```
┌──────────────────────────────────────────────────────────────────────┐
│                           SAVEGRESS                                   │
│                                                                       │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────────────┐  │
│  │   Engine    │      │   Broker    │      │     Consumers       │  │
│  │             │─────▶│             │─────▶│                     │  │
│  │  PostgreSQL │      │  Storage    │      │  Webhook            │  │
│  │  MySQL      │      │  DLQ        │      │  Kafka              │  │
│  │  MongoDB    │      │  Metrics    │      │  gRPC               │  │
│  │  ...        │      │             │      │  Custom             │  │
│  └─────────────┘      └─────────────┘      └─────────────────────┘  │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

## Use Cases

### Data Replication
Keep read replicas, caches, and search indexes up-to-date:
```
Primary DB → Savegress → Elasticsearch, Redis, Reporting DB
```

### Event-Driven Microservices
Convert database changes to domain events:
```
Orders DB → Savegress → Order Service, Inventory Service, Notification Service
```

### Real-time Analytics
Feed data warehouses without batch ETL:
```
Production DB → Savegress → Snowflake, BigQuery, Redshift
```

### Audit & Compliance
Capture complete change history for regulatory requirements:
```
All DBs → Savegress → Audit Log Storage (S3, GCS)
```

### Cache Invalidation
Automatically invalidate caches when data changes:
```
DB → Savegress → Redis (invalidate) → Application
```

## Savegress vs Alternatives

| Feature | Savegress | Debezium | AWS DMS | Fivetran |
|---------|-----------|----------|---------|----------|
| Latency | < 100ms | < 100ms | Seconds | Minutes |
| Self-hosted | ✅ | ✅ | ❌ | ❌ |
| Built-in compression | ✅ | ❌ | ❌ | ❌ |
| DLQ included | ✅ | ❌ | ❌ | ❌ |
| Single binary | ✅ | ❌ | N/A | N/A |
| Resource usage | Low | Medium | N/A | N/A |
| Free tier | Unlimited | Unlimited | Limited | Limited |

## Editions

### Community (Free)
- PostgreSQL, MySQL, MariaDB
- 1 source, 10 tables, 1K events/sec
- Basic rate limiting, circuit breaker
- Perfect for development and small projects

### Pro
- All Community features
- MongoDB, SQL Server, Cassandra, DynamoDB
- Webhook, Kafka, gRPC outputs
- Compression, DLQ, Backpressure
- 10 sources, 100 tables, 50K events/sec

### Enterprise
- All Pro features
- Oracle support
- PITR, Cloud Storage (S3/GCS/Azure)
- HA Clustering, Multi-region
- mTLS, RBAC, Vault, Audit logging
- Unlimited scale

## Next Steps

1. [Quick Start](quickstart.md) - Get running in 5 minutes
2. [Installation](installation.md) - Production deployment
3. [First Pipeline](first-pipeline.md) - Create your first CDC pipeline

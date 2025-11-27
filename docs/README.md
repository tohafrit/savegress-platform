# Savegress Documentation

Welcome to the official Savegress documentation. Savegress is a high-performance Change Data Capture (CDC) platform that streams database changes in real-time.

## Quick Navigation

### Getting Started
- [Introduction](getting-started/introduction.md) - What is Savegress and why use it
- [Quick Start](getting-started/quickstart.md) - Get running in 5 minutes
- [Installation](getting-started/installation.md) - Detailed installation guide
- [First Pipeline](getting-started/first-pipeline.md) - Create your first CDC pipeline

### Configuration
- [Configuration Reference](configuration/reference.md) - Complete configuration options
- [Optimization Guide](configuration/optimization.md) - Performance tuning
- [Environment Variables](configuration/environment.md) - Environment configuration
- [Examples](configuration/examples.md) - Common configuration patterns

### Database Connectors (Sources)
- [PostgreSQL](connectors/sources/postgresql.md) - Logical replication
- [MySQL](connectors/sources/mysql.md) - Binlog capture
- [MariaDB](connectors/sources/mariadb.md) - GTID support
- [MongoDB](connectors/sources/mongodb.md) - Change Streams (Pro)
- [SQL Server](connectors/sources/sqlserver.md) - CDC with LSN (Pro)
- [Cassandra](connectors/sources/cassandra.md) - Commit log (Pro)
- [DynamoDB](connectors/sources/dynamodb.md) - DynamoDB Streams (Pro)
- [Oracle](connectors/sources/oracle.md) - LogMiner (Enterprise)

### Output Connectors (Sinks)
- [Stdout/File](connectors/sinks/stdout.md) - Basic output
- [HTTP Webhook](connectors/sinks/webhook.md) - HTTP delivery (Pro)
- [Kafka](connectors/sinks/kafka.md) - Kafka producer (Pro)
- [gRPC](connectors/sinks/grpc.md) - gRPC streaming (Pro)
- [Custom SDK](connectors/sinks/custom.md) - Build your own (Enterprise)

### Features
- [Compression](features/compression.md) - Hybrid compression (Pro)
- [Dead Letter Queue](features/dlq.md) - Failed message handling (Pro)
- [Backpressure](features/backpressure.md) - Flow control (Pro)
- [Rate Limiting](features/rate-limiting.md) - Traffic control
- [Schema Evolution](features/schema-evolution.md) - Auto-detect changes (Pro)
- [PITR](features/pitr.md) - Point-in-time recovery (Enterprise)
- [High Availability](features/ha.md) - Clustering (Enterprise)
- [Security](features/security.md) - mTLS, RBAC, Vault (Enterprise)

### API Reference
- [REST API](api/rest.md) - Platform API
- [gRPC API](api/grpc.md) - Streaming API
- [Event Format](api/events.md) - CDC event schema
- [Metrics](api/metrics.md) - Prometheus metrics

### Operations
- [Monitoring](operations/monitoring.md) - Observability setup
- [Alerting](operations/alerting.md) - Alert configuration
- [Backup & Recovery](operations/backup.md) - Data protection
- [Troubleshooting](troubleshooting/README.md) - Common issues

### Licensing
- [Plans & Features](licensing/plans.md) - Community, Pro, Enterprise
- [Limits](licensing/limits.md) - Usage limits by plan
- [Activation](licensing/activation.md) - License management

---

## Feature Matrix

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| PostgreSQL, MySQL, MariaDB | ✅ | ✅ | ✅ |
| MongoDB, SQL Server, Cassandra, DynamoDB | - | ✅ | ✅ |
| Oracle | - | - | ✅ |
| Webhook, Kafka, gRPC outputs | - | ✅ | ✅ |
| Compression (4-10x savings) | - | ✅ | ✅ |
| Dead Letter Queue | - | ✅ | ✅ |
| Schema Evolution | - | ✅ | ✅ |
| Prometheus Metrics | - | ✅ | ✅ |
| PITR, Cloud Storage | - | - | ✅ |
| HA Clustering | - | - | ✅ |
| mTLS, RBAC, Audit | - | - | ✅ |
| **Max Sources** | 1 | 10 | Unlimited |
| **Max Tables** | 10 | 100 | Unlimited |
| **Throughput** | 1K/sec | 50K/sec | Unlimited |

---

## Support

- **Community**: [GitHub Issues](https://github.com/savegress/savegress/issues)
- **Pro**: Email support (24h response)
- **Enterprise**: Dedicated support with SLA

---

*Last updated: 2025-11-27*

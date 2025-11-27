# Plans & Features

Savegress offers three editions designed for different use cases and team sizes.

## Edition Comparison

| | Community | Pro | Enterprise |
|---|:---------:|:---:|:----------:|
| **Price** | Free Forever | Contact Sales | Contact Sales |
| **Target** | Startups, POC, Dev | Scale-ups, Production | Large Orgs, Regulated |
| **Support** | Community | Email (24h) | Dedicated + SLA |

---

## Community Edition (Free)

**Perfect for:** Startups, proof-of-concept, development, small projects.

### Included Features

#### Database Connectors
- PostgreSQL (logical replication)
- MySQL (binlog capture)
- MariaDB (GTID support)

#### Output
- Stdout/File output

#### Safety Features (Always Free)
- Token bucket rate limiting
- Circuit breaker pattern
- Health check endpoints
- Basic internal metrics
- Basic schema detection

#### Limits
| Resource | Limit |
|----------|-------|
| Sources | 1 |
| Tables | 10 |
| Throughput | 1,000 events/sec |
| Data Retention | 1 day |

### What You Can Build

- Development environment with real CDC
- Small production workloads
- Single-database sync scenarios
- Learning and experimentation

---

## Pro Edition

**Perfect for:** Production workloads, growing teams, DevOps-focused organizations.

### Everything in Community, plus:

#### Additional Database Connectors
- MongoDB (Change Streams)
- SQL Server (CDC with LSN)
- Cassandra (Commit log CDC)
- DynamoDB (DynamoDB Streams)

#### Output Connectors
- HTTP Webhook
- Kafka Producer
- gRPC Streaming
- Point-in-time Snapshots

#### Performance
- Hybrid Compression (4-10x storage savings)
  - LZ4 for speed
  - ZSTD for ratio
  - Auto-selection based on data

#### Reliability
- Advanced Rate Limiting (adaptive, sliding window)
- Automatic Backpressure Control
- Dead Letter Queue (DLQ)
- Event Replay for Recovery

#### Schema Management
- Auto Schema Evolution
- Safe migration detection

#### Observability
- Prometheus Metrics Export
- SLA Monitoring (Bronze/Silver/Gold)

#### Limits
| Resource | Limit |
|----------|-------|
| Sources | 10 |
| Tables | 100 |
| Throughput | 50,000 events/sec |
| Data Retention | 30 days |

### ROI Examples

1. **Compression:** Save 4-10x on storage and bandwidth
   - 1TB/day → 100-250GB/day
   - At $0.023/GB (S3), save $500-700/month

2. **DLQ:** Never lose an event
   - Failed webhook? → DLQ → Replay later
   - Zero data loss guarantee

3. **Prometheus:** Integrate with existing monitoring
   - No custom instrumentation needed
   - Out-of-the-box Grafana dashboards

---

## Enterprise Edition

**Perfect for:** Large organizations, regulated industries, mission-critical workloads.

### Everything in Pro, plus:

#### Database Connectors
- Oracle (LogMiner with SCN)

#### Output Connectors
- Custom Output SDK (build your own connectors)

#### Performance
- SIMD-optimized Compression (AVX2/512, NEON)
  - 2-3x faster compression on modern CPUs

#### Reliability
- Exactly-Once Delivery Semantics
- Guaranteed no duplicates

#### Disaster Recovery
- Point-in-Time Recovery (PITR)
- Cloud Storage Backends (S3, GCS, Azure Blob)

#### Schema Management
- Migration Approval Workflow
- Change management with approvals

#### Observability
- OpenTelemetry Distributed Tracing
- Full end-to-end visibility

#### High Availability
- Active-Passive HA Mode
- Raft Consensus Clustering
- Multi-Region Deployment

#### Security & Compliance
- End-to-End Encryption
- Mutual TLS (mTLS) Authentication
- Role-Based Access Control (RBAC)
- HashiCorp Vault Integration
- Comprehensive Audit Logging
- SSO/SAML Integration
- LDAP Directory Integration
- Multi-Tenant Isolation

#### Limits
| Resource | Limit |
|----------|-------|
| Sources | **Unlimited** |
| Tables | **Unlimited** |
| Throughput | **Unlimited** |
| Data Retention | **Unlimited** |

### Compliance Features

| Requirement | Feature |
|-------------|---------|
| SOC 2 | RBAC, Audit Logging, Encryption |
| HIPAA | Encryption, Access Control, Audit |
| GDPR | Data isolation, Audit trails |
| PCI-DSS | Encryption, mTLS, RBAC |

---

## Feature Matrix

### Database Connectors

| Connector | Community | Pro | Enterprise |
|-----------|:---------:|:---:|:----------:|
| PostgreSQL | ✅ | ✅ | ✅ |
| MySQL | ✅ | ✅ | ✅ |
| MariaDB | ✅ | ✅ | ✅ |
| MongoDB | - | ✅ | ✅ |
| SQL Server | - | ✅ | ✅ |
| Cassandra | - | ✅ | ✅ |
| DynamoDB | - | ✅ | ✅ |
| Oracle | - | - | ✅ |

### Output Connectors

| Output | Community | Pro | Enterprise |
|--------|:---------:|:---:|:----------:|
| Stdout/File | ✅ | ✅ | ✅ |
| HTTP Webhook | - | ✅ | ✅ |
| Kafka | - | ✅ | ✅ |
| gRPC Streaming | - | ✅ | ✅ |
| Custom SDK | - | - | ✅ |

### Performance

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| Basic Processing | ✅ | ✅ | ✅ |
| Hybrid Compression | - | ✅ | ✅ |
| SIMD Optimization | - | - | ✅ |

### Reliability

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| Basic Rate Limiting | ✅ | ✅ | ✅ |
| Circuit Breaker | ✅ | ✅ | ✅ |
| Advanced Rate Limiting | - | ✅ | ✅ |
| Backpressure Control | - | ✅ | ✅ |
| Dead Letter Queue | - | ✅ | ✅ |
| Event Replay | - | ✅ | ✅ |
| Exactly-Once | - | - | ✅ |

### Disaster Recovery

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| Checkpoints | ✅ | ✅ | ✅ |
| Snapshots | - | ✅ | ✅ |
| PITR | - | - | ✅ |
| Cloud Storage | - | - | ✅ |

### Observability

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| Internal Metrics | ✅ | ✅ | ✅ |
| Health Checks | ✅ | ✅ | ✅ |
| Prometheus | - | ✅ | ✅ |
| SLA Monitoring | - | ✅ | ✅ |
| OpenTelemetry | - | - | ✅ |

### High Availability

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| Single Instance | ✅ | ✅ | ✅ |
| HA Mode | - | - | ✅ |
| Raft Clustering | - | - | ✅ |
| Multi-Region | - | - | ✅ |

### Security

| Feature | Community | Pro | Enterprise |
|---------|:---------:|:---:|:----------:|
| TLS Encryption | ✅ | ✅ | ✅ |
| mTLS | - | - | ✅ |
| RBAC | - | - | ✅ |
| Vault Integration | - | - | ✅ |
| Audit Logging | - | - | ✅ |
| SSO/LDAP | - | - | ✅ |
| Multi-Tenant | - | - | ✅ |

---

## Upgrade Path

### Community → Pro

When you need:
- More database connectors (MongoDB, SQL Server, etc.)
- Output to Kafka or webhooks
- Compression to reduce costs
- DLQ for reliability
- More scale (10 sources, 100 tables, 50K/sec)

### Pro → Enterprise

When you need:
- Oracle database support
- Exactly-once delivery guarantees
- PITR for disaster recovery
- HA clustering for uptime
- Compliance features (RBAC, audit, encryption)
- Unlimited scale

---

## FAQ

### What happens when I hit a limit?

You'll see a clear message with current usage and upgrade options. The system continues working within your limits - we never stop your production workload.

### Can I try Enterprise features?

Yes! Contact sales for a 14-day Enterprise trial. No credit card required.

### Is there a self-hosted option?

Yes! All plans can be self-hosted. Enterprise includes deployment support.

### Can I downgrade?

Yes, but you'll lose access to higher-tier features. We recommend testing in a non-production environment first.

### What's the license duration?

- Community: Forever free
- Pro: Annual subscription
- Enterprise: Annual or multi-year

---

## Contact

- **Sales:** sales@savegress.io
- **Support:** support@savegress.io
- **Website:** https://savegress.io/pricing

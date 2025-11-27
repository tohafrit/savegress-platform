# Performance Optimization Guide

This guide helps you choose the optimal configuration for your workload.

## Quick Start: Choose Your Profile

Start by identifying your primary workload type:

| Workload | Primary Goal | Recommended Profile |
|----------|--------------|---------------------|
| **Real-time Analytics** | Minimal latency | [Low Latency](#low-latency-profile) |
| **Event Streaming** | Reliability, ordering | [Streaming](#streaming-profile) |
| **Data Replication** | Consistency, recovery | [Replication](#replication-profile) |
| **Batch Processing** | Throughput, efficiency | [Batch](#batch-profile) |

---

## Decision Tree

Use this flowchart to find your optimal configuration:

```
START
  │
  ├─ What's your latency requirement?
  │   │
  │   ├─ < 10ms ──────────────────────► Ultra-Low Latency
  │   │                                  • No compression
  │   │                                  • Single-event batches
  │   │                                  • LZ4 only if must compress
  │   │
  │   ├─ 10-100ms ────────────────────► Low Latency
  │   │                                  • LZ4 compression
  │   │                                  • Small batches (10-50)
  │   │                                  • Token bucket rate limiting
  │   │
  │   └─ > 100ms ─────────────────────► Standard Latency
  │                                      │
  │                                      ├─ What's your delivery guarantee?
  │                                      │   │
  │                                      │   ├─ At-least-once ──► Streaming Profile
  │                                      │   │                    • DLQ enabled
  │                                      │   │                    • Leader ACK
  │                                      │   │
  │                                      │   ├─ Exactly-once ───► Enterprise Streaming
  │                                      │   │                    • All ISR ACK
  │                                      │   │                    • Transaction mode
  │                                      │   │
  │                                      │   └─ Best-effort ────► Batch Profile
  │                                      │                        • Maximum throughput
  │                                      │
  │                                      └─ What's your recovery requirement?
  │                                          │
  │                                          ├─ RPO < 1 min ────► PITR Profile
  │                                          │                    (Enterprise)
  │                                          │
  │                                          └─ RPO > 1 min ────► Standard Profile
```

---

## Low Latency Profile

For real-time dashboards, live metrics, instant notifications.

```yaml
# Target: < 100ms end-to-end latency

compression:
  enabled: true
  algorithm: lz4
  lz4:
    level: 1          # Fastest
  min_size: 1024      # Skip small messages

batching:
  max_size: 10
  max_wait: 10ms
  adaptive: false     # Fixed small batches

buffer:
  type: ring
  size: 4096
  overflow_policy: drop_oldest

rate_limiting:
  algorithm: token_bucket
  tokens_per_second: 50000
  burst_size: 500

backpressure:
  strategy: pause
  high_watermark: 0.7

replication:
  ack_mode: leader

checkpoint:
  interval: 30s
  sync_mode: async
```

### When to Use

- Dashboard updates
- Live monitoring
- Instant alerts
- User-facing real-time features

### Trade-offs

| Pro | Con |
|-----|-----|
| Minimal latency | Higher network overhead |
| Fast recovery | Less throughput efficiency |
| Simple config | May drop events under load |

---

## Streaming Profile

For event-driven architectures, microservices, reliable delivery.

```yaml
# Target: Reliable delivery with good throughput

compression:
  enabled: true
  algorithm: hybrid
  hybrid:
    threshold: 4096
    small_algo: lz4
    large_algo: zstd
    large_level: 3

batching:
  max_size: 200
  max_wait: 100ms
  adaptive: true
  adaptive_config:
    min_size: 50
    max_size: 500
    target_latency: 50ms

buffer:
  size: 8192
  pool:
    enabled: true
    max_size: 64
  overflow:
    enabled: true
    policy: buffer
    path: /var/lib/savegress/overflow

rate_limiting:
  algorithm: sliding_window   # Pro
  window_size: 1s
  max_requests: 25000

backpressure:
  enabled: true
  strategy: adaptive_throttle
  high_watermark: 0.8
  low_watermark: 0.4

dlq:
  enabled: true
  max_retries: 5
  retry_delay: 1s
  exponential_backoff: true

replication:
  ack_mode: leader

retry:
  max_attempts: 5
  initial_delay: 100ms
  max_delay: 30s
  multiplier: 2.0

checkpoint:
  interval: 10s
```

### When to Use

- Microservices event bus
- Order processing
- Inventory updates
- Any system requiring guaranteed delivery

### Trade-offs

| Pro | Con |
|-----|-----|
| Guaranteed delivery | Higher latency |
| DLQ for failures | More complex |
| Automatic retry | More resource usage |

---

## Replication Profile

For database sync, DR, maintaining consistency.

```yaml
# Target: Data consistency and recovery

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 5

batching:
  max_size: 500
  max_wait: 200ms
  adaptive: true

buffer:
  size: 16384
  overflow:
    enabled: true
    max_size: 5GB

backpressure:
  strategy: buffer
  high_watermark: 0.9

replication:
  ack_mode: all           # Wait for all replicas
  min_isr: 2

checkpoint:
  interval: 5s
  sync_mode: sync         # fsync after each

snapshot:                 # Pro
  enabled: true
  interval: 1h
  retention_count: 24

schema:
  evolution:
    enabled: true
    compatible_changes: auto
    breaking_changes: warn

# Enterprise
pitr:
  enabled: true
  retention: 7d
  granularity: 1m

storage:
  backend: s3
  s3:
    bucket: savegress-backup
    sync_interval: 5m
```

### When to Use

- Database replication
- Disaster recovery
- Data warehouse loading
- Compliance requirements

### Trade-offs

| Pro | Con |
|-----|-----|
| Zero data loss | Higher latency |
| Point-in-time recovery | More storage |
| Full consistency | Higher costs |

---

## Batch Profile

For ETL, periodic sync, bulk operations.

```yaml
# Target: Maximum throughput, minimum resource usage

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 9              # Maximum compression
  simd:
    enabled: true         # Enterprise

batching:
  mode: time              # Time-based batching
  interval: 5m            # Batch every 5 minutes
  max_size: 10000

buffer:
  size: 32768
  pool:
    enabled: true
    max_size: 128

parallel:
  table_parallelism: 16
  transaction_parallelism: 8

rate_limiting:
  enabled: false          # No limit for batch

backpressure:
  strategy: buffer

replication:
  ack_mode: leader

checkpoint:
  interval: 5m
```

### When to Use

- ETL pipelines
- Data warehouse loading
- Periodic batch sync
- Report generation

### Trade-offs

| Pro | Con |
|-----|-----|
| Maximum efficiency | High latency |
| Best compression | Delayed visibility |
| Lowest cost | Not for real-time |

---

## High Volume Profile (Enterprise)

For 50K+ events/second, enterprise scale.

```yaml
# Target: > 50,000 events/sec

compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 3
  simd:
    enabled: true
    instruction_set: auto  # AVX2/512 or NEON

batching:
  max_size: 1000
  max_wait: 50ms
  adaptive: true
  target_throughput: 100000

buffer:
  size: 65536
  pool:
    enabled: true
    initial_size: 64
    max_size: 256
    prealloc_size: 64KB

parallel:
  table_parallelism: 32
  transaction_parallelism: 16
  priority:
    enabled: true
    high_priority_tables:
      - orders
      - payments

rate_limiting:
  algorithm: adaptive
  adaptive:
    min_rate: 10000
    max_rate: 200000
    target_latency: 100ms

backpressure:
  strategy: weighted_fair
  high_watermark: 0.85

dlq:
  enabled: true
  preserve_order: true

exactly_once:
  enabled: true

replication:
  ack_mode: all
  min_isr: 2

ha:
  enabled: true
  cluster:
    consensus: raft
    nodes: 3
```

---

## Tuning Individual Components

### Compression Tuning

| Goal | Algorithm | Level | SIMD |
|------|-----------|-------|------|
| Lowest latency | None | - | - |
| Low latency | LZ4 | 1-3 | No |
| Balanced | Hybrid | Auto | Yes |
| Best ratio | ZSTD | 9-15 | Yes |
| Maximum | ZSTD | 19-22 | Yes |

### Batching Tuning

| Goal | Max Size | Max Wait | Adaptive |
|------|----------|----------|----------|
| Ultra-low latency | 1 | 1ms | No |
| Low latency | 10-50 | 10ms | No |
| Balanced | 100-500 | 100ms | Yes |
| High throughput | 500-2000 | 200ms | Yes |
| Batch processing | 10000+ | 5m+ | No |

### Rate Limiting Tuning

| Algorithm | Overhead | Best For |
|-----------|----------|----------|
| Token Bucket | Lowest | Simple rate limiting |
| Sliding Window | Low | Smooth traffic |
| Adaptive | Medium | Variable workloads |

### Backpressure Strategies

| Strategy | Behavior | Best For |
|----------|----------|----------|
| `pause` | Stop reading | Simple, fast |
| `drop_oldest` | Discard old events | Monitoring, metrics |
| `drop_newest` | Discard new events | Backfill scenarios |
| `buffer` | Spill to disk | Reliability |
| `adaptive_throttle` | Auto-adjust rate | Variable load |
| `weighted_fair` | Priority-based | Multi-tenant |

---

## Monitoring Performance

### Key Metrics

```prometheus
# Throughput
savegress_events_processed_total
rate(savegress_events_processed_total[1m])

# Latency
savegress_event_latency_seconds{quantile="0.99"}

# Compression
savegress_compression_ratio
savegress_compression_duration_seconds

# Backpressure
savegress_backpressure_events_total
savegress_buffer_utilization

# Replication lag
savegress_replication_lag_seconds
```

### Alerting Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Event latency P99 | > 500ms | > 2s |
| Replication lag | > 10s | > 60s |
| Buffer utilization | > 70% | > 90% |
| DLQ size | > 1000 | > 10000 |
| Error rate | > 1% | > 5% |

---

## See Also

- [Configuration Reference](reference.md) - All options
- [Compression](../features/compression.md) - Compression details
- [DLQ](../features/dlq.md) - Dead letter queue
- [Metrics](../api/metrics.md) - Prometheus metrics

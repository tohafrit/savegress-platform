# Configuration Reference

Complete reference for all Savegress configuration options.

## Configuration Sources

Configuration is loaded in order of precedence (later overrides earlier):

1. Default values
2. Configuration file (`--config` flag)
3. Environment variables (`SAVEGRESS_*`)
4. Command-line flags

## Environment Variable Interpolation

Use `${VAR}` or `${VAR:-default}` in YAML files:

```yaml
source:
  password: ${DB_PASSWORD}
  host: ${DB_HOST:-localhost}
```

---

## Engine Configuration

### Source (Database Connection)

```yaml
source:
  # Database type (required)
  # Values: postgres, mysql, mariadb, mongodb, sqlserver, cassandra, dynamodb, oracle
  type: postgres

  # Connection settings
  host: localhost
  port: 5432
  database: mydb
  user: cdc_user
  password: ${DB_PASSWORD}

  # Connection pool
  max_connections: 5
  connection_timeout: 30s
  idle_timeout: 5m

  # TLS settings
  tls:
    enabled: false
    mode: verify-full  # disable, require, verify-ca, verify-full
    ca_file: /path/to/ca.crt
    cert_file: /path/to/client.crt
    key_file: /path/to/client.key
    skip_verify: false
```

### PostgreSQL-specific

```yaml
source:
  type: postgres
  # ... connection settings ...

  # Replication settings
  slot_name: savegress_slot        # Replication slot name
  publication: savegress_pub       # Publication name
  create_slot: true                # Auto-create slot
  drop_slot_on_close: false        # Keep slot on shutdown

  # WAL settings
  standby_message_interval: 10s    # Status message interval
  wal_receiver_status_interval: 10s
```

### MySQL/MariaDB-specific

```yaml
source:
  type: mysql  # or mariadb
  # ... connection settings ...

  server_id: 12345                 # Unique server ID for replication
  binlog_position: ""              # Start position (empty = latest)
  gtid_mode: true                  # Use GTID (MariaDB)
```

### MongoDB-specific (Pro)

```yaml
source:
  type: mongodb
  # ... connection settings ...

  replica_set: rs0
  read_preference: primaryPreferred
  change_stream_options:
    full_document: updateLookup
    batch_size: 1000
```

### Table Filtering

```yaml
source:
  # Include only these tables
  tables:
    - public.users
    - public.orders
    - inventory.*           # Wildcard

  # Exclude these tables
  exclude_tables:
    - public.sessions
    - public.logs
    - "*.audit_*"           # Wildcard

  # Column filtering
  columns:
    public.users:
      - id
      - name
      - email
      # Excludes: password_hash, internal_notes

    public.orders:
      - "*"                  # All columns
      - "!internal_*"        # Except internal_*
```

### Row Filtering (Pro)

```yaml
source:
  row_filters:
    public.users:
      condition: "status = 'active'"
    public.orders:
      condition: "amount > 0 AND deleted_at IS NULL"
```

---

## Output Configuration

### Stdout (Default)

```yaml
output:
  type: stdout
  format: json              # json, text
  pretty: false             # Pretty print JSON
```

### File

```yaml
output:
  type: file
  path: /var/log/savegress/events.jsonl
  rotate:
    enabled: true
    max_size: 100MB
    max_files: 10
    compress: true
```

### HTTP Webhook (Pro)

```yaml
output:
  type: webhook
  url: https://your-api.com/events
  method: POST

  # Headers
  headers:
    Authorization: "Bearer ${WEBHOOK_TOKEN}"
    Content-Type: application/json

  # Batching
  batch_size: 100           # Events per request
  batch_timeout: 1s         # Max wait time

  # Retry
  retry:
    enabled: true
    max_attempts: 5
    initial_delay: 100ms
    max_delay: 30s
    multiplier: 2.0

  # Timeout
  timeout: 30s

  # TLS
  tls:
    skip_verify: false
    ca_file: /path/to/ca.crt
```

### Kafka (Pro)

```yaml
output:
  type: kafka

  brokers:
    - kafka1:9092
    - kafka2:9092
    - kafka3:9092

  topic: cdc-events
  # Or dynamic topic per table
  topic_template: "cdc.${schema}.${table}"

  # Producer settings
  acks: all                 # none, leader, all
  compression: snappy       # none, gzip, snappy, lz4, zstd
  batch_size: 16384
  linger_ms: 5
  max_request_size: 1048576

  # Partitioning
  partitioner: hash         # hash, round-robin, manual
  partition_key: "${table}" # Key for hash partitioner

  # Authentication
  sasl:
    enabled: true
    mechanism: SCRAM-SHA-512  # PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
    username: ${KAFKA_USER}
    password: ${KAFKA_PASSWORD}

  # TLS
  tls:
    enabled: true
    ca_file: /path/to/ca.crt
```

### gRPC (Pro)

```yaml
output:
  type: grpc
  broker_address: localhost:9092

  # Connection settings
  max_message_size: 16777216  # 16MB
  keepalive_interval: 30s
  keepalive_timeout: 10s

  # TLS
  tls:
    enabled: false
    ca_file: /path/to/ca.crt
    cert_file: /path/to/client.crt
    key_file: /path/to/client.key
```

---

## Performance Configuration

### Batching

```yaml
batching:
  enabled: true
  max_size: 100             # Max events per batch
  max_wait: 100ms           # Max time to wait for batch
  adaptive: true            # Auto-adjust based on load

  # Adaptive settings (Pro)
  adaptive_config:
    min_size: 10
    max_size: 1000
    target_latency: 50ms
    adjustment_interval: 1s
```

### Compression (Pro)

```yaml
compression:
  enabled: true
  algorithm: hybrid         # none, lz4, zstd, hybrid

  # LZ4 settings
  lz4:
    level: 3                # 1-12, higher = slower, smaller

  # ZSTD settings
  zstd:
    level: 3                # 1-22, higher = slower, smaller

  # Hybrid settings (auto-select best)
  hybrid:
    threshold: 4096         # Bytes: below=LZ4, above=ZSTD
    small_algo: lz4
    large_algo: zstd
    large_level: 5

  # SIMD optimization (Enterprise)
  simd:
    enabled: true
    instruction_set: auto   # auto, avx2, avx512, neon

  # Skip compression for small messages
  min_size: 256             # Don't compress below this
```

### Buffer & Memory

```yaml
buffer:
  type: ring                # ring, channel
  size: 8192                # Buffer capacity

  # Buffer pool
  pool:
    enabled: true
    initial_size: 16
    max_size: 64
    prealloc_size: 64KB

  # Overflow handling
  overflow:
    enabled: true
    policy: buffer          # drop_oldest, drop_newest, buffer, block
    path: /var/lib/savegress/overflow
    max_size: 1GB
    compression: true
```

### Parallel Processing

```yaml
parallel:
  # Table-level parallelism
  table_parallelism: 8

  # Transaction parallelism
  transaction_parallelism: 4

  # Priority queues (Enterprise)
  priority:
    enabled: true
    high_priority_tables:
      - orders
      - payments
    medium_priority_tables:
      - users
      - products
```

---

## Reliability Configuration

### Rate Limiting

```yaml
rate_limiting:
  enabled: true

  # Algorithm: token_bucket (Community), sliding_window/adaptive (Pro)
  algorithm: token_bucket

  # Token bucket settings
  tokens_per_second: 10000
  burst_size: 1000

  # Sliding window settings (Pro)
  window_size: 1s
  max_requests: 10000

  # Adaptive settings (Pro)
  adaptive:
    enabled: true
    min_rate: 1000
    max_rate: 100000
    target_latency: 100ms
    adjustment_interval: 1s
```

### Backpressure (Pro)

```yaml
backpressure:
  enabled: true

  # Strategy
  strategy: adaptive_throttle
  # Options: pause, drop_oldest, drop_newest, buffer, adaptive_throttle,
  #          weighted_fair (Enterprise), deadline (Enterprise)

  # Watermarks
  high_watermark: 0.8       # Start backpressure at 80%
  low_watermark: 0.4        # Stop backpressure at 40%

  # Throttle settings
  throttle:
    min_rate: 100
    max_rate: 100000
    adjustment_interval: 100ms
```

### Dead Letter Queue (Pro)

```yaml
dlq:
  enabled: true
  base_dir: /var/lib/savegress/dlq

  # Retention
  retention_days: 14
  max_messages: 5000000

  # Retry before DLQ
  max_retries: 5
  retry_delay: 1s
  exponential_backoff: true
  max_retry_delay: 5m

  # Storage settings
  segment_size: 1GB
```

### Circuit Breaker

```yaml
circuit_breaker:
  enabled: true

  # Thresholds
  failure_threshold: 5      # Failures to open
  success_threshold: 3      # Successes to close
  timeout: 30s              # Time in open state

  # Half-open settings
  half_open_requests: 3

  # Adaptive (Pro)
  adaptive:
    enabled: true
    min_samples: 100
    error_rate_threshold: 0.5
```

### Retry

```yaml
retry:
  enabled: true
  max_attempts: 5
  initial_delay: 100ms
  max_delay: 30s
  multiplier: 2.0
  jitter: 0.1               # Add randomness to prevent thundering herd
```

---

## Checkpoint & Recovery

### Checkpointing

```yaml
checkpoint:
  enabled: true
  dir: /var/lib/savegress/checkpoints
  interval: 10s
  sync_mode: async          # async, sync (fsync after each)
```

### Snapshots (Pro)

```yaml
snapshot:
  enabled: true
  interval: 1h
  retention_count: 24
  dir: /var/lib/savegress/snapshots
```

### PITR - Point-in-Time Recovery (Enterprise)

```yaml
pitr:
  enabled: true
  retention: 7d
  granularity: 1m
  storage_dir: /var/lib/savegress/pitr
```

### Cloud Storage (Enterprise)

```yaml
storage:
  backend: s3               # local, s3, gcs, azure

  s3:
    bucket: savegress-backup
    region: us-east-1
    access_key: ${AWS_ACCESS_KEY}
    secret_key: ${AWS_SECRET_KEY}
    prefix: cdc/
    sync_interval: 5m

  gcs:
    bucket: savegress-backup
    project: my-project
    credentials_file: /path/to/credentials.json

  azure:
    container: savegress-backup
    account_name: ${AZURE_ACCOUNT}
    account_key: ${AZURE_KEY}
```

---

## Schema Management

### Schema Detection

```yaml
schema:
  detection: auto           # none, basic, auto
  cache_ttl: 5m
```

### Schema Evolution (Pro)

```yaml
schema:
  evolution:
    enabled: true
    compatible_changes: auto  # auto, warn, error
    breaking_changes: warn    # warn, error, block

    # Notification
    notify:
      enabled: true
      channels:
        - slack
        - email
```

### Schema Approval Workflow (Enterprise)

```yaml
schema:
  evolution:
    approval_workflow: true
    approval_timeout: 24h
    approvers:
      - admin@example.com
    notify_channels:
      - slack
```

---

## Observability

### Logging

```yaml
logging:
  level: info               # debug, info, warn, error
  format: json              # json, text
  output: stdout            # stdout, stderr, /path/to/file

  # File rotation
  rotation:
    enabled: true
    max_size: 100MB
    max_files: 10
    compress: true
```

### Metrics

```yaml
metrics:
  enabled: true
  address: :8081
  path: /metrics

  # Prometheus labels
  labels:
    environment: production
    service: cdc-engine
```

### Prometheus Export (Pro)

```yaml
metrics:
  prometheus:
    enabled: true
    path: /metrics
    namespace: savegress
```

### OpenTelemetry (Enterprise)

```yaml
tracing:
  enabled: true
  provider: otlp            # otlp, jaeger, zipkin

  otlp:
    endpoint: otel-collector:4317
    insecure: false
    headers:
      Authorization: "Bearer ${OTEL_TOKEN}"

  sampling:
    ratio: 0.1              # Sample 10% of traces
```

### SLA Monitoring (Pro)

```yaml
sla:
  enabled: true
  tier: gold                # bronze, silver, gold

  # SLA targets
  targets:
    bronze:
      latency_p99: 1000ms
      availability: 99.0
    silver:
      latency_p99: 500ms
      availability: 99.9
    gold:
      latency_p99: 100ms
      availability: 99.99
```

---

## High Availability (Enterprise)

### HA Mode

```yaml
ha:
  enabled: true
  mode: active-passive      # active-passive, active-active
```

### Raft Clustering

```yaml
cluster:
  enabled: true
  consensus: raft

  node_id: node-1
  bind_address: :7000
  advertise_address: node1.example.com:7000

  peers:
    - node1.example.com:7000
    - node2.example.com:7000
    - node3.example.com:7000

  election_timeout: 1s
  heartbeat_interval: 100ms
  snapshot_interval: 5m
```

### Multi-Region

```yaml
multi_region:
  enabled: true
  local_region: us-east

  regions:
    - name: us-east
      primary: true
      endpoints:
        - us-east-1.example.com:9092
    - name: eu-west
      replica: true
      endpoints:
        - eu-west-1.example.com:9092

  sync_mode: async          # sync, async
  conflict_resolution: last-write-wins
```

---

## Security (Enterprise)

### Mutual TLS

```yaml
tls:
  enabled: true
  mode: mtls                # tls, mtls

  cert_file: /etc/savegress/certs/server.crt
  key_file: /etc/savegress/certs/server.key
  ca_file: /etc/savegress/certs/ca.crt

  client_auth: require      # none, request, require
  min_version: "1.2"
```

### RBAC

```yaml
rbac:
  enabled: true
  provider: local           # local, ldap, oidc

  roles:
    admin:
      permissions: ["*"]
    operator:
      permissions: ["read", "write", "manage:pipeline"]
    viewer:
      permissions: ["read"]
```

### HashiCorp Vault

```yaml
vault:
  enabled: true
  address: https://vault.example.com:8200
  auth_method: kubernetes   # token, kubernetes, aws

  secrets:
    db_password: secret/data/savegress/db#password
    api_key: secret/data/savegress/api#key
```

### Audit Logging

```yaml
audit:
  enabled: true
  output: /var/log/savegress/audit.log
  events:
    - authentication
    - authorization
    - configuration_change
    - pipeline_management
```

---

## License Configuration

```yaml
license:
  # License key (required for Pro/Enterprise)
  key: ${SAVEGRESS_LICENSE_KEY}

  # Or license file
  file: /etc/savegress/license.key

  # Offline mode (no validation server)
  offline: false

  # Grace period after expiry
  grace_period: 7d
```

---

## Complete Example

```yaml
# Production configuration example

source:
  type: postgres
  host: ${DB_HOST}
  port: 5432
  database: production
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  slot_name: savegress_prod
  publication: savegress_pub
  tables:
    - public.users
    - public.orders
    - public.products
  tls:
    enabled: true
    mode: verify-full
    ca_file: /etc/savegress/certs/db-ca.crt

output:
  type: kafka
  brokers:
    - kafka1:9092
    - kafka2:9092
    - kafka3:9092
  topic_template: "cdc.${schema}.${table}"
  acks: all
  compression: zstd
  sasl:
    enabled: true
    mechanism: SCRAM-SHA-512
    username: ${KAFKA_USER}
    password: ${KAFKA_PASSWORD}

compression:
  enabled: true
  algorithm: hybrid

batching:
  max_size: 500
  max_wait: 50ms
  adaptive: true

rate_limiting:
  algorithm: adaptive
  adaptive:
    min_rate: 5000
    max_rate: 50000
    target_latency: 100ms

backpressure:
  enabled: true
  strategy: adaptive_throttle
  high_watermark: 0.8

dlq:
  enabled: true
  retention_days: 14
  max_retries: 5

checkpoint:
  interval: 10s
  sync_mode: async

schema:
  evolution:
    enabled: true
    compatible_changes: auto

logging:
  level: info
  format: json

metrics:
  prometheus:
    enabled: true

license:
  key: ${SAVEGRESS_LICENSE_KEY}
```

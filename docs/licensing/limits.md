# Usage Limits

Understanding and managing Savegress usage limits.

## Limits by Plan

| Limit | Community | Pro | Enterprise |
|-------|:---------:|:---:|:----------:|
| **Sources** | 1 | 10 | Unlimited |
| **Tables** | 10 | 100 | Unlimited |
| **Throughput** | 1,000/sec | 50,000/sec | Unlimited |
| **Retention** | 1 day | 30 days | Unlimited |

---

## Understanding Each Limit

### Sources

A **source** is a database connection that Savegress captures changes from.

```yaml
# This counts as 1 source
source:
  type: postgres
  host: db1.example.com
  database: orders

# This counts as another source (2 total)
source:
  type: postgres
  host: db2.example.com
  database: inventory
```

**Community:** 1 source
- One database connection
- Perfect for single-database setups

**Pro:** 10 sources
- Connect multiple databases
- Support for microservices architecture

**Enterprise:** Unlimited
- Connect as many databases as needed
- Multi-region, multi-tenant deployments

### Tables

A **table** is a single database table being tracked for changes.

```yaml
source:
  tables:
    - public.users      # 1 table
    - public.orders     # 2 tables
    - public.products   # 3 tables
```

**Community:** 10 tables
- Core tables for a small application

**Pro:** 100 tables
- Full application database tracking
- Multiple microservice databases

**Enterprise:** Unlimited
- Enterprise-wide CDC
- Data warehouse ingestion

### Throughput

**Throughput** is measured in events per second (events/sec).

An event is any INSERT, UPDATE, or DELETE operation captured.

```
100 inserts/sec + 50 updates/sec + 10 deletes/sec = 160 events/sec
```

**Community:** 1,000 events/sec
- ~86 million events/day
- Suitable for development and small workloads

**Pro:** 50,000 events/sec
- ~4.3 billion events/day
- Production-grade throughput

**Enterprise:** Unlimited
- No artificial limits
- Limited only by hardware

### Data Retention

**Retention** is how long Savegress stores events in the broker before deletion.

**Community:** 1 day
- Recent events only
- Real-time use cases

**Pro:** 30 days
- Replay capability
- Recovery window

**Enterprise:** Unlimited
- Compliance requirements
- Historical analysis

---

## Checking Current Usage

### Via CLI

```bash
savegress-engine --license-info

# Output:
# License: Pro
# Customer: Acme Corp
#
# Usage:
#   Sources:    3 / 10 (30%)
#   Tables:    45 / 100 (45%)
#   Throughput: 12,345 / 50,000 events/sec (25%)
```

### Via API

```bash
curl http://localhost:8080/api/v1/license/usage

# Response:
{
  "tier": "pro",
  "usage": {
    "sources": { "current": 3, "max": 10, "percentage": 30 },
    "tables": { "current": 45, "max": 100, "percentage": 45 },
    "throughput": { "current": 12345, "max": 50000, "percentage": 25 }
  }
}
```

### Via Prometheus Metrics

```prometheus
# Current source count
savegress_sources_active{tier="pro"} 3

# Current table count
savegress_tables_tracked{tier="pro"} 45

# Current throughput
savegress_events_per_second{tier="pro"} 12345

# Limit utilization percentage
savegress_limit_utilization{resource="sources",tier="pro"} 0.30
savegress_limit_utilization{resource="tables",tier="pro"} 0.45
savegress_limit_utilization{resource="throughput",tier="pro"} 0.25
```

---

## Limit Behavior

### Approaching Limits

When you reach 80% of a limit, you'll see warnings:

```
WARN: Approaching source limit (8/10, 80%). Consider upgrading to Enterprise for unlimited sources.
```

### At Limit

When you reach a limit:

**Sources/Tables:**
- New sources/tables are rejected
- Existing pipelines continue working
- Clear error message with upgrade path

```
ERROR: Source limit reached (10/10). Cannot add new source.
Upgrade to Enterprise for unlimited sources: https://savegress.io/upgrade
```

**Throughput:**
- Events are rate-limited (not dropped)
- Backpressure applied upstream
- Clear warning in logs

```
WARN: Throughput limit reached (50,000/sec). Applying backpressure.
Events are being queued, not lost. Upgrade for higher throughput.
```

### Grace Period

Community and Pro have a 10% grace buffer:
- Pro with 10 sources can briefly use 11
- Pro with 50K/sec can burst to 55K/sec
- Grace period: 1 hour before enforcement

Enterprise has no limits, so no grace period needed.

---

## Limit Enforcement Points

### Source Limit

Enforced when:
- Adding a new source via configuration
- Creating a pipeline via API
- Starting the engine with a new config

```yaml
# Rejected if at source limit
sources:
  - type: postgres
    host: new-db.example.com  # ERROR: Source limit reached
```

### Table Limit

Enforced when:
- Adding tables to configuration
- Publication includes more tables
- Dynamic table discovery

```yaml
source:
  tables:
    - public.*  # Expands to 150 tables
    # ERROR: Table limit exceeded (150 > 100)
```

### Throughput Limit

Enforced in real-time:
- Token bucket rate limiting
- Backpressure to source
- Events queued, not dropped

```
Events arriving: 60,000/sec
Limit: 50,000/sec
Result: 10,000/sec queued with backpressure
```

---

## Optimizing Within Limits

### Reduce Source Count

Use connection pooling and shared connections:

```yaml
# Instead of 3 separate sources:
sources:
  - host: db.example.com
    database: app1
  - host: db.example.com
    database: app2
  - host: db.example.com
    database: app3

# Use 1 source with multiple databases (PostgreSQL):
source:
  host: db.example.com
  databases:
    - app1
    - app2
    - app3
# Counts as 1 source
```

### Reduce Table Count

Filter to essential tables:

```yaml
source:
  # Don't use wildcards carelessly
  # tables:
  #   - public.*  # Might include 200+ tables

  # Be specific
  tables:
    - public.users
    - public.orders
    - public.products

  exclude_tables:
    - public.sessions
    - public.logs
    - public.migrations
```

### Optimize Throughput

1. **Enable compression** (Pro+): Reduce network overhead

```yaml
compression:
  enabled: true
  algorithm: hybrid
```

2. **Use batching**: More efficient processing

```yaml
batching:
  max_size: 500
  max_wait: 100ms
```

3. **Filter unnecessary events**:

```yaml
source:
  exclude_tables:
    - public.health_checks
    - public.sessions

  row_filters:
    public.logs:
      condition: "level = 'ERROR'"  # Only capture errors
```

---

## Upgrading Limits

### Temporary Increase

Contact support for temporary limit increases:
- Planned migrations
- Peak traffic events
- Testing higher loads

### Permanent Upgrade

1. **Community → Pro:**
   - Purchase Pro license
   - Apply license key
   - Limits increase immediately

2. **Pro → Enterprise:**
   - Contact sales
   - Enterprise license
   - All limits removed

### License Application

```bash
# Via environment variable
export SAVEGRESS_LICENSE_KEY="eyJhbGciOiJFZDI1NTE5..."
savegress-engine --config config.yaml

# Via config file
license:
  key: "eyJhbGciOiJFZDI1NTE5..."

# Verify new limits
savegress-engine --license-info
```

---

## Monitoring & Alerts

### Prometheus Alerts

```yaml
groups:
  - name: savegress-limits
    rules:
      - alert: ApproachingSourceLimit
        expr: savegress_limit_utilization{resource="sources"} > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Approaching source limit ({{ $value | humanizePercentage }})"

      - alert: AtSourceLimit
        expr: savegress_limit_utilization{resource="sources"} >= 1.0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Source limit reached"

      - alert: ThroughputLimitExceeded
        expr: savegress_events_per_second > savegress_throughput_limit * 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Throughput approaching limit"
```

### Grafana Dashboard

```json
{
  "title": "Savegress Limits",
  "panels": [
    {
      "title": "Source Utilization",
      "type": "gauge",
      "targets": [
        {
          "expr": "savegress_limit_utilization{resource=\"sources\"} * 100"
        }
      ],
      "thresholds": {
        "steps": [
          { "value": 0, "color": "green" },
          { "value": 80, "color": "yellow" },
          { "value": 95, "color": "red" }
        ]
      }
    }
  ]
}
```

---

## FAQ

### What happens to events when throughput limit is hit?

Events are **queued**, not dropped. Backpressure is applied to slow down the source. Once throughput drops, queued events are processed.

### Can I exceed limits temporarily?

Yes, there's a 10% grace buffer for 1 hour. For planned spikes, contact support for temporary increases.

### How is throughput measured?

Throughput is a 1-minute rolling average of events per second. Brief spikes above the limit are allowed.

### Do DELETE operations count as events?

Yes. INSERT, UPDATE, and DELETE all count as events toward the throughput limit.

### Does schema change (DDL) count as an event?

DDL events count as 1 event per schema change, regardless of tables affected.

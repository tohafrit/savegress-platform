# DataWatch — Real-time Analytics Add-on

> "Real-time insights, zero ETL"

## Overview

DataWatch превращает CDC события Savegress в actionable analytics без традиционного ETL pipeline. Автоматические метрики, аномалии, и data quality monitoring из коробки.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           DATAWATCH                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   CDC Events Stream                                                      │
│         │                                                                │
│         ▼                                                                │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│   │   Metrics   │───▶│  Anomaly    │───▶│   Alerts    │                │
│   │   Engine    │    │  Detection  │    │   Router    │                │
│   └─────────────┘    └─────────────┘    └─────────────┘                │
│         │                  │                   │                        │
│         ▼                  ▼                   ▼                        │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                │
│   │  Time-Series│    │   ML Models │    │  Slack/PD   │                │
│   │   Storage   │    │   (Anomaly) │    │  Webhooks   │                │
│   └─────────────┘    └─────────────┘    └─────────────┘                │
│         │                                                               │
│         ▼                                                                │
│   ┌─────────────────────────────────────────────────────┐              │
│   │              Dashboard Builder                       │              │
│   │   ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐     │              │
│   │   │Chart│  │Table│  │Gauge│  │Map  │  │Alert│     │              │
│   │   └─────┘  └─────┘  └─────┘  └─────┘  └─────┘     │              │
│   └─────────────────────────────────────────────────────┘              │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Target Audience

| Segment | Use Case | Pain Point Solved |
|---------|----------|-------------------|
| **SaaS Startups** | Product analytics | No need for Amplitude/Mixpanel |
| **Data Teams** | Operational dashboards | Real-time without Kafka setup |
| **DevOps** | System monitoring | DB-level observability |
| **Business Analysts** | KPI tracking | Self-service analytics |

---

## Core Features

### 1. Auto-Metrics Engine

Автоматическое создание метрик из CDC событий без конфигурации.

```yaml
# Автоматически генерируемые метрики для таблицы "orders"
auto_metrics:
  orders:
    # Count metrics
    - orders_total              # Total count
    - orders_created_rate       # INSERTs per minute
    - orders_updated_rate       # UPDATEs per minute
    - orders_deleted_rate       # DELETEs per minute

    # Field-based metrics (auto-detected numeric fields)
    - orders_amount_sum         # SUM of amount field
    - orders_amount_avg         # AVG of amount field
    - orders_amount_p99         # P99 of amount field

    # Status distribution (auto-detected enum/status fields)
    - orders_by_status          # Count by status field

    # Time-based
    - orders_lag_seconds        # CDC lag for this table
```

**Умное определение типов:**
```go
type MetricInference struct {
    // Определяем тип поля
    NumericFields   []string  // amount, price, quantity → sum/avg/p99
    StatusFields    []string  // status, state, type → distribution
    TimestampFields []string  // created_at, updated_at → freshness
    IDFields        []string  // user_id, order_id → cardinality
}
```

---

### 2. Dashboard Builder

Drag & drop интерфейс для создания дашбордов.

```
┌─────────────────────────────────────────────────────────────────┐
│  DataWatch Dashboard: E-commerce Overview            [Edit] [⋮] │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Orders/min    │  │   Revenue/hr    │  │   Avg Order     │ │
│  │                 │  │                 │  │                 │ │
│  │      127        │  │    $12,450      │  │     $98.03      │ │
│  │    ▲ +15%       │  │    ▲ +8%        │  │    ▼ -2%        │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Orders Over Time                                    [1h]│  │
│  │  150│    ╭──╮                                            │  │
│  │     │   ╭╯  ╰╮  ╭─╮                                      │  │
│  │  100│──╯      ╰─╯  ╰──╮                                  │  │
│  │     │                  ╰──                                │  │
│  │   50│                                                     │  │
│  │     └────────────────────────────────────────────────    │  │
│  │      10:00    10:15    10:30    10:45    11:00           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────────────────┐  ┌────────────────────────────┐   │
│  │  Orders by Status      │  │  Top Products (1h)         │   │
│  │  ┌────────────────┐    │  │                            │   │
│  │  │████████░░│ paid 72% │  │  1. Widget Pro    $4,200   │   │
│  │  │███░░░░░░░│ pend 18% │  │  2. Gadget X      $3,100   │   │
│  │  │█░░░░░░░░░│ canc 10% │  │  3. Tool Basic    $2,800   │   │
│  │  └────────────────┘    │  │  4. Service Plus  $1,900   │   │
│  └────────────────────────┘  └────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Widget Types:**
| Widget | Description | Use Case |
|--------|-------------|----------|
| **Counter** | Single number with trend | KPIs, totals |
| **Line Chart** | Time-series data | Trends over time |
| **Bar Chart** | Categorical comparison | Distribution |
| **Pie/Donut** | Proportions | Status breakdown |
| **Table** | Raw data view | Top N, details |
| **Gauge** | Progress to goal | Quota tracking |
| **Heatmap** | 2D distribution | Time patterns |
| **Alert List** | Active alerts | Incident awareness |

---

### 3. Anomaly Detection

ML-based detection без конфигурации.

```
┌─────────────────────────────────────────────────────────────────┐
│  Anomaly Detection Engine                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Historical Data                Algorithms                       │
│  ┌─────────────┐               ┌─────────────────────────────┐  │
│  │ 7 days of   │──────────────▶│  1. Statistical (Z-score)   │  │
│  │ metric data │               │  2. Seasonal (STL decomp)   │  │
│  └─────────────┘               │  3. ML (Isolation Forest)   │  │
│                                └─────────────────────────────┘  │
│         │                                    │                   │
│         ▼                                    ▼                   │
│  ┌─────────────┐               ┌─────────────────────────────┐  │
│  │  Real-time  │               │     Anomaly Detected!       │  │
│  │  metric     │──────────────▶│                             │  │
│  │  stream     │               │  orders_rate: 450/min       │  │
│  └─────────────┘               │  Expected: 100-150/min      │  │
│                                │  Severity: HIGH             │  │
│                                └─────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Anomaly Types:**
```go
type AnomalyType string

const (
    AnomalySpike     AnomalyType = "spike"      // Внезапный рост
    AnomalyDrop      AnomalyType = "drop"       // Внезапное падение
    AnomalyTrend     AnomalyType = "trend"      // Длительное изменение тренда
    AnomalySeasons   AnomalyType = "seasonal"   // Нарушение сезонности
    AnomalyMissing   AnomalyType = "missing"    // Отсутствие данных
)

type Anomaly struct {
    ID          string      `json:"id"`
    MetricName  string      `json:"metric_name"`
    Type        AnomalyType `json:"type"`
    Severity    string      `json:"severity"`  // low, medium, high, critical
    Value       float64     `json:"value"`
    Expected    Range       `json:"expected"`
    DetectedAt  time.Time   `json:"detected_at"`
    Description string      `json:"description"`
}
```

---

### 4. Data Quality Monitor

Отслеживание качества данных в реальном времени.

```yaml
data_quality_rules:
  orders:
    completeness:
      - field: customer_id
        rule: not_null
        threshold: 99.9%

      - field: email
        rule: not_empty
        threshold: 95%

    validity:
      - field: amount
        rule: positive
        alert_on: any_violation

      - field: status
        rule: in_set
        values: [pending, paid, shipped, cancelled]

    freshness:
      - rule: max_age
        threshold: 5m
        description: "No orders older than 5 minutes"

    consistency:
      - rule: referential_integrity
        field: customer_id
        references: customers.id
```

**Quality Dashboard:**
```
┌─────────────────────────────────────────────────────────────────┐
│  Data Quality Score: 94.2%                         [Last 24h]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Table          Completeness   Validity   Freshness   Score     │
│  ─────────────────────────────────────────────────────────────  │
│  orders         99.8%          98.5%      100%        99.4%  ✓  │
│  customers      97.2%          99.1%      100%        98.8%  ✓  │
│  products       100%           95.0%      100%        98.3%  ✓  │
│  payments       99.5%          82.3%      98%         93.3%  ⚠  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  Recent Issues:                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ ⚠ payments.status: 17.7% invalid values (unknown)       │  │
│  │ ⚠ payments.processed_at: 2% records stale (>5min)       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 5. Schema Change Tracker

Отслеживание изменений схемы БД.

```
┌─────────────────────────────────────────────────────────────────┐
│  Schema Changes                                    [Last 30d]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Nov 28, 14:32  │  orders                                       │
│  ──────────────────────────────────────────────────────────     │
│  + ADD COLUMN   │  discount_code VARCHAR(50)                    │
│  + ADD INDEX    │  idx_orders_discount_code                     │
│                 │                                                │
│  Nov 25, 09:15  │  customers                                    │
│  ──────────────────────────────────────────────────────────     │
│  ~ ALTER COLUMN │  phone: VARCHAR(20) → VARCHAR(30)             │
│                 │                                                │
│  Nov 22, 16:45  │  products                                     │
│  ──────────────────────────────────────────────────────────     │
│  - DROP COLUMN  │  legacy_sku (was VARCHAR(50))                 │
│  ⚠ BREAKING     │  May affect downstream consumers              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 6. Alerts & Notifications

```yaml
alert_rules:
  - name: "Order Spike"
    metric: orders_created_rate
    condition: "> 200% of baseline"
    for: 5m
    severity: warning
    channels: [slack, email]

  - name: "Revenue Drop"
    metric: orders_amount_sum
    condition: "< 50% of same_hour_last_week"
    for: 15m
    severity: critical
    channels: [pagerduty, slack]

  - name: "Data Quality Issue"
    metric: dq_score_payments
    condition: "< 90%"
    for: 10m
    severity: warning
    channels: [slack]

notification_channels:
  slack:
    webhook_url: ${SLACK_WEBHOOK}
    channel: "#data-alerts"

  pagerduty:
    service_key: ${PAGERDUTY_KEY}

  email:
    recipients: [data-team@company.com]
```

---

## Technical Architecture

### Backend Components

```
datawatch/
├── cmd/
│   └── datawatch/
│       └── main.go              # Standalone binary or plugin
├── internal/
│   ├── metrics/
│   │   ├── engine.go            # Metrics computation
│   │   ├── auto_discover.go     # Auto-metric generation
│   │   ├── aggregator.go        # Time-window aggregations
│   │   └── storage.go           # Time-series storage interface
│   ├── anomaly/
│   │   ├── detector.go          # Anomaly detection coordinator
│   │   ├── statistical.go       # Z-score, MAD
│   │   ├── seasonal.go          # STL decomposition
│   │   └── ml.go                # Isolation Forest
│   ├── quality/
│   │   ├── monitor.go           # Data quality monitoring
│   │   ├── rules.go             # Rule definitions
│   │   └── scorer.go            # Quality score calculation
│   ├── schema/
│   │   ├── tracker.go           # Schema change detection
│   │   └── diff.go              # Schema comparison
│   ├── alerts/
│   │   ├── engine.go            # Alert evaluation
│   │   ├── router.go            # Notification routing
│   │   └── channels/            # Slack, PagerDuty, etc.
│   └── api/
│       ├── handlers.go          # HTTP handlers
│       ├── websocket.go         # Real-time updates
│       └── routes.go            # Route registration
├── pkg/
│   └── sdk/                     # Client SDK
└── ui/
    └── dashboard/               # React dashboard components
```

### Data Flow

```go
// CDC Event → DataWatch Pipeline
type DataWatchPipeline struct {
    metricsEngine  *metrics.Engine
    anomalyEngine  *anomaly.Detector
    qualityMonitor *quality.Monitor
    schemaTracker  *schema.Tracker
    alertEngine    *alerts.Engine
}

func (p *DataWatchPipeline) ProcessEvent(event *cdc.Event) error {
    // 1. Update metrics
    p.metricsEngine.Record(event)

    // 2. Check for anomalies
    if anomaly := p.anomalyEngine.Check(event); anomaly != nil {
        p.alertEngine.Trigger(anomaly)
    }

    // 3. Validate data quality
    if violations := p.qualityMonitor.Validate(event); len(violations) > 0 {
        p.alertEngine.TriggerQualityIssue(violations)
    }

    // 4. Track schema changes
    if event.Type == cdc.EventTypeDDL {
        p.schemaTracker.RecordChange(event)
    }

    return nil
}
```

### Storage

```go
// Time-series storage interface
type MetricStorage interface {
    // Write
    Record(metric string, value float64, labels map[string]string, ts time.Time) error

    // Query
    Query(metric string, from, to time.Time, aggregation string) ([]DataPoint, error)
    QueryRange(metric string, from, to time.Time, step time.Duration) ([]DataPoint, error)

    // Metadata
    ListMetrics() ([]MetricMeta, error)
    GetMetricMeta(metric string) (*MetricMeta, error)
}

// Supported backends
type StorageBackend string

const (
    StorageEmbedded   StorageBackend = "embedded"   // Built-in (SQLite + custom)
    StoragePrometheus StorageBackend = "prometheus" // Remote write
    StorageInfluxDB   StorageBackend = "influxdb"   // InfluxDB
    StorageTimescale  StorageBackend = "timescale"  // TimescaleDB
    StorageClickHouse StorageBackend = "clickhouse" // ClickHouse
)
```

---

## Frontend Components

### Dashboard Page

```typescript
// app/(portal)/datawatch/page.tsx
export default function DataWatchPage() {
  return (
    <div className="space-y-6">
      {/* Overview Cards */}
      <div className="grid grid-cols-4 gap-4">
        <MetricCard metric="events_total" title="Events Today" />
        <MetricCard metric="anomalies_active" title="Active Anomalies" />
        <MetricCard metric="dq_score" title="Data Quality" format="percent" />
        <MetricCard metric="alerts_triggered" title="Alerts (24h)" />
      </div>

      {/* Main Dashboard */}
      <DashboardGrid dashboardId="default" />

      {/* Recent Anomalies */}
      <AnomalyList limit={5} />

      {/* Schema Changes */}
      <SchemaChangeLog limit={5} />
    </div>
  );
}
```

### Widget Components

```typescript
// components/datawatch/widgets/
├── MetricCard.tsx       // Single metric display
├── LineChart.tsx        // Time-series chart
├── BarChart.tsx         // Bar chart
├── PieChart.tsx         // Pie/donut chart
├── DataTable.tsx        // Tabular data
├── Gauge.tsx            // Progress gauge
├── Heatmap.tsx          // Time heatmap
├── AlertList.tsx        // Alert feed
└── AnomalyBadge.tsx     // Anomaly indicator
```

---

## API Endpoints

```yaml
# DataWatch API
/api/v1/datawatch:
  # Metrics
  GET  /metrics                    # List all metrics
  GET  /metrics/{name}             # Get metric details
  GET  /metrics/{name}/query       # Query metric data

  # Dashboards
  GET  /dashboards                 # List dashboards
  POST /dashboards                 # Create dashboard
  GET  /dashboards/{id}            # Get dashboard
  PUT  /dashboards/{id}            # Update dashboard
  DELETE /dashboards/{id}          # Delete dashboard

  # Widgets
  POST /dashboards/{id}/widgets    # Add widget
  PUT  /widgets/{id}               # Update widget
  DELETE /widgets/{id}             # Delete widget

  # Anomalies
  GET  /anomalies                  # List anomalies
  GET  /anomalies/{id}             # Get anomaly details
  POST /anomalies/{id}/acknowledge # Acknowledge anomaly

  # Data Quality
  GET  /quality/score              # Overall DQ score
  GET  /quality/tables/{table}     # Table DQ details
  GET  /quality/rules              # List DQ rules
  POST /quality/rules              # Create DQ rule

  # Schema
  GET  /schema/changes             # List schema changes
  GET  /schema/tables/{table}      # Table schema history

  # Alerts
  GET  /alerts                     # List alert rules
  POST /alerts                     # Create alert rule
  GET  /alerts/history             # Alert history
```

---

## Configuration

```yaml
# datawatch.yaml
datawatch:
  enabled: true

  metrics:
    auto_discover: true
    retention: 30d
    storage: embedded  # embedded, prometheus, influxdb

  anomaly:
    enabled: true
    algorithms:
      - statistical   # Z-score based
      - seasonal      # STL decomposition
    sensitivity: medium  # low, medium, high
    baseline_window: 7d

  quality:
    enabled: true
    default_rules: true  # Apply default rules
    score_threshold: 90  # Alert below this

  schema:
    track_changes: true
    alert_on_breaking: true

  alerts:
    evaluation_interval: 1m
    channels:
      slack:
        webhook_url: ${SLACK_WEBHOOK}
      email:
        smtp_host: smtp.example.com
        from: alerts@company.com

  storage:
    # Embedded storage (default)
    embedded:
      path: /var/lib/datawatch/data

    # Or external Prometheus
    prometheus:
      url: http://prometheus:9090
      remote_write: true
```

---

## Integration with Savegress Portal

### Sidebar Addition

```typescript
// In sidebar.tsx - add to CDC group or create Analytics group
{
  title: 'Analytics',
  icon: BarChart3,
  items: [
    { name: 'DataWatch', href: '/datawatch', icon: Activity },
    { name: 'Dashboards', href: '/datawatch/dashboards', icon: LayoutDashboard },
    { name: 'Alerts', href: '/datawatch/alerts', icon: Bell },
  ],
}
```

### License Integration

```go
// DataWatch requires Pro or Enterprise license
func (d *DataWatch) CheckLicense(license *license.License) error {
    if license.Tier == "community" {
        return ErrDataWatchRequiresPro
    }

    // Check feature flags
    if !license.HasFeature("datawatch") {
        return ErrDataWatchNotIncluded
    }

    return nil
}
```

---

## Pricing & Packaging

| Feature | Community | Pro | Enterprise |
|---------|-----------|-----|------------|
| Auto-Metrics | 5 tables | Unlimited | Unlimited |
| Dashboards | 1 | 10 | Unlimited |
| Anomaly Detection | ❌ | ✅ | ✅ |
| Data Quality | Basic | Full | Full + Custom |
| Schema Tracking | ❌ | ✅ | ✅ |
| Alerts | ❌ | 10 rules | Unlimited |
| Retention | 24h | 30d | Custom |
| Export | ❌ | CSV | CSV + API |

---

## Development Roadmap

### Phase 1: MVP (4 weeks)
- [ ] Auto-metrics engine
- [ ] Basic dashboard with pre-built widgets
- [ ] Simple threshold alerts
- [ ] Integration with Savegress Portal

### Phase 2: Intelligence (4 weeks)
- [ ] Anomaly detection (statistical)
- [ ] Data quality monitoring
- [ ] Schema change tracking
- [ ] Dashboard builder UI

### Phase 3: Advanced (4 weeks)
- [ ] ML-based anomaly detection
- [ ] Custom metrics & calculations
- [ ] External storage backends
- [ ] Grafana/Metabase export

### Phase 4: Enterprise (ongoing)
- [ ] Custom dashboards sharing
- [ ] Role-based access
- [ ] Audit logging
- [ ] SLA reporting

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Time to first dashboard | < 5 minutes |
| Auto-discovered metrics | 90%+ accuracy |
| Anomaly detection precision | > 85% |
| False positive rate | < 15% |
| User activation (create dashboard) | 60% of Pro users |

---

*Document version: 1.0*
*Last updated: November 2024*

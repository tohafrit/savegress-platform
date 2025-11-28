# IoTSense - IoT & Industry 4.0 Add-on

## Overview

**Tagline:** Industrial CDC for connected devices and smart manufacturing

**Target Market:** Manufacturing, Energy & Utilities, Smart Cities, Logistics, Agriculture, Automotive

IoTSense enables real-time processing of IoT sensor data through CDC, providing predictive maintenance, anomaly detection, and digital twin capabilities for industrial applications.

## Core Features

### 1. Sensor Data Ingestion

High-throughput ingestion of IoT sensor data with protocol support.

**Supported Protocols:**
- MQTT (v3.1.1, v5.0)
- AMQP
- CoAP
- OPC-UA
- Modbus TCP/RTU
- BACnet
- Kafka
- HTTP/REST

**Ingestion Configuration:**
```yaml
sensor_ingestion:
  protocols:
    - name: "mqtt"
      broker: "mqtt://broker.factory.local:1883"
      topics:
        - pattern: "sensors/+/temperature"
          parser: "json"
          schema:
            device_id: "$.deviceId"
            value: "$.temp"
            unit: "$.unit"
            timestamp: "$.ts"

        - pattern: "machines/+/vibration"
          parser: "protobuf"
          schema_file: "vibration.proto"

    - name: "opcua"
      endpoint: "opc.tcp://plc.factory.local:4840"
      nodes:
        - id: "ns=2;s=Line1.Motor1.Speed"
          name: "motor1_speed"
          sampling_interval_ms: 100
        - id: "ns=2;s=Line1.Motor1.Current"
          name: "motor1_current"
          sampling_interval_ms: 100

    - name: "modbus"
      host: "192.168.1.100"
      port: 502
      registers:
        - address: 0
          count: 2
          type: "float32"
          name: "tank_level"
        - address: 2
          count: 2
          type: "float32"
          name: "flow_rate"
```

**Data Normalization:**
```yaml
normalization:
  rules:
    - source: "temperature_sensors"
      transforms:
        - type: "unit_conversion"
          from_field: "value"
          from_unit: "fahrenheit"
          to_unit: "celsius"

        - type: "timestamp_normalize"
          from_field: "ts"
          timezone: "UTC"

        - type: "add_metadata"
          fields:
            plant_id: "${env.PLANT_ID}"
            line_id: "${topic.split('/')[1]}"
```

### 2. Predictive Maintenance

ML-powered failure prediction for equipment and machinery.

**Capabilities:**
- Remaining Useful Life (RUL) prediction
- Failure probability scoring
- Maintenance scheduling optimization
- Spare parts inventory prediction
- Work order generation

**Predictive Models:**
```yaml
predictive_maintenance:
  models:
    - name: "motor_failure"
      type: "remaining_useful_life"
      equipment_class: "electric_motor"
      features:
        - name: "vibration_rms"
          window: "1h"
          aggregation: "mean"
        - name: "temperature"
          window: "1h"
          aggregation: "max"
        - name: "current_draw"
          window: "1h"
          aggregation: "mean"
        - name: "operating_hours"
          type: "cumulative"
      thresholds:
        critical: 7  # days
        warning: 30
        normal: 90

    - name: "pump_seal_degradation"
      type: "anomaly_detection"
      algorithm: "isolation_forest"
      features:
        - "pressure_differential"
        - "flow_rate"
        - "vibration_spectrum"
      contamination: 0.01

    - name: "bearing_wear"
      type: "classification"
      algorithm: "gradient_boosting"
      features:
        - name: "vibration_fft"
          type: "spectrum"
          bands: [100, 500, 1000, 2000, 5000]
        - name: "temperature_trend"
          window: "24h"
          type: "slope"
      classes:
        - "healthy"
        - "early_wear"
        - "advanced_wear"
        - "critical"
```

**Maintenance Actions:**
```yaml
maintenance_actions:
  triggers:
    - condition: "rul_days < 7"
      actions:
        - type: "work_order"
          priority: "emergency"
          template: "motor_replacement"
          assign_to: "maintenance_team_a"

        - type: "alert"
          channels: ["pagerduty", "sms"]
          recipients: ["plant_manager", "maintenance_lead"]

        - type: "inventory_check"
          parts: ["motor_${equipment.model}", "bearings", "coupling"]

    - condition: "rul_days < 30"
      actions:
        - type: "schedule_maintenance"
          window: "next_weekend"
          duration: "4h"

        - type: "order_parts"
          if_stock_below: 1
```

### 3. Real-time Anomaly Detection

Detect equipment anomalies and process deviations in real-time.

**Detection Methods:**
- Statistical process control (SPC)
- Machine learning (Isolation Forest, Autoencoders)
- Rule-based thresholds
- Pattern matching
- Correlation analysis

**Anomaly Rules:**
```yaml
anomaly_detection:
  rules:
    - name: "temperature_spike"
      type: "threshold"
      sensor: "motor_temperature"
      conditions:
        - metric: "value"
          operator: ">"
          threshold: 80
          severity: "warning"
        - metric: "value"
          operator: ">"
          threshold: 95
          severity: "critical"
        - metric: "rate_of_change"
          window: "5m"
          operator: ">"
          threshold: 10
          severity: "warning"

    - name: "vibration_anomaly"
      type: "ml"
      sensor: "motor_vibration"
      model: "autoencoder"
      features:
        - "rms"
        - "peak"
        - "crest_factor"
        - "kurtosis"
      reconstruction_threshold: 0.15

    - name: "process_correlation"
      type: "correlation_break"
      sensors: ["flow_rate", "pressure"]
      expected_correlation: 0.85
      min_correlation: 0.5
      window: "1h"

    - name: "spc_control"
      type: "control_chart"
      sensor: "product_dimension"
      chart_type: "xbar_r"
      subgroup_size: 5
      rules:
        - "beyond_3sigma"
        - "2_of_3_beyond_2sigma"
        - "4_of_5_beyond_1sigma"
        - "8_consecutive_same_side"
```

### 4. Digital Twin Engine

Create and maintain digital replicas of physical assets.

**Features:**
- Real-time state synchronization
- Physics-based simulation
- What-if scenario analysis
- Historical state replay
- 3D visualization integration

**Digital Twin Definition:**
```yaml
digital_twins:
  - name: "production_line_1"
    type: "assembly_line"
    components:
      - id: "conveyor_1"
        type: "conveyor"
        sensors:
          - id: "conv1_speed"
            type: "encoder"
            unit: "m/s"
          - id: "conv1_motor_temp"
            type: "temperature"
            unit: "celsius"
        parameters:
          length: 50  # meters
          max_speed: 2.5

      - id: "robot_arm_1"
        type: "robot_arm"
        model: "fanuc_m20"
        sensors:
          - id: "arm1_joint_positions"
            type: "joint_encoder"
            joints: 6
          - id: "arm1_torque"
            type: "torque"
            joints: 6
        parameters:
          reach: 1.8
          payload: 20

    relationships:
      - from: "conveyor_1"
        to: "robot_arm_1"
        type: "feeds_into"
        transfer_time: 2.5  # seconds

    simulation:
      physics_engine: "gazebo"
      model_file: "line1.urdf"
      update_rate_hz: 100
```

**Twin Operations:**
```yaml
twin_operations:
  sync:
    mode: "realtime"
    latency_target_ms: 100

  simulation:
    - name: "throughput_optimization"
      type: "what_if"
      variables:
        - parameter: "conveyor_1.speed"
          range: [1.0, 2.5]
          step: 0.1
      objective: "maximize:throughput"
      constraints:
        - "robot_arm_1.cycle_time >= 3.0"
        - "energy_consumption <= 150kW"

  replay:
    enabled: true
    retention: "90d"
    granularity: "100ms"
```

### 5. Edge Computing Support

Distributed processing at the edge for low-latency applications.

**Edge Capabilities:**
- Local ML inference
- Data filtering and aggregation
- Store-and-forward
- Offline operation
- Edge-to-cloud sync

**Edge Configuration:**
```yaml
edge_computing:
  nodes:
    - id: "edge_node_1"
      location: "plant_floor_1"
      hardware:
        type: "nvidia_jetson"
        model: "orin_nano"

      local_processing:
        - name: "vibration_fft"
          type: "transform"
          input: "raw_vibration"
          output: "vibration_spectrum"
          sampling_rate: 10000

        - name: "anomaly_detection"
          type: "ml_inference"
          model: "motor_anomaly_v2"
          input: "vibration_spectrum"
          output: "anomaly_score"
          latency_target_ms: 10

      data_routing:
        - condition: "anomaly_score > 0.7"
          destination: "cloud"
          priority: "high"

        - condition: "true"
          destination: "cloud"
          aggregation: "1m"
          method: "mean"

      offline_buffer:
        enabled: true
        max_size_gb: 10
        sync_on_reconnect: true
```

### 6. Asset Performance Management

Track and optimize asset performance across the organization.

**Metrics:**
- Overall Equipment Effectiveness (OEE)
- Mean Time Between Failures (MTBF)
- Mean Time To Repair (MTTR)
- Asset utilization
- Energy efficiency
- Production quality

**OEE Configuration:**
```yaml
oee_tracking:
  equipment:
    - id: "cnc_machine_1"
      shifts:
        - name: "day"
          start: "06:00"
          end: "14:00"
        - name: "evening"
          start: "14:00"
          end: "22:00"
        - name: "night"
          start: "22:00"
          end: "06:00"

      availability:
        running_signal: "machine_state"
        running_value: "running"
        planned_downtime:
          - type: "scheduled_maintenance"
            schedule: "sunday 06:00-10:00"
          - type: "changeover"
            signal: "changeover_active"

      performance:
        ideal_cycle_time: 45  # seconds
        actual_cycle_signal: "cycle_complete"

      quality:
        good_parts_signal: "parts_passed_qc"
        reject_signal: "parts_rejected"

  dashboards:
    - name: "plant_oee"
      widgets:
        - type: "gauge"
          title: "Overall OEE"
          aggregation: "plant"

        - type: "waterfall"
          title: "OEE Breakdown"
          components: ["availability", "performance", "quality"]

        - type: "pareto"
          title: "Top Downtime Reasons"
          metric: "downtime_minutes"
          group_by: "reason"

        - type: "trend"
          title: "OEE Trend"
          period: "30d"
          compare_to: "previous_period"
```

## Technical Architecture

### Data Flow

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│     Sensors     │───▶│  Edge Node   │───▶│    IoTSense     │
│   (MQTT/OPC)    │    │  Processing  │    │     Engine      │
└─────────────────┘    └──────────────┘    └────────┬────────┘
                                                     │
┌─────────────────┐    ┌──────────────┐    ┌────────▼────────┐
│      PLCs       │───▶│   Savegress  │───▶│   Time Series   │
│    (Modbus)     │    │     CDC      │    │    Database     │
└─────────────────┘    └──────────────┘    └────────┬────────┘
                                                     │
┌─────────────────┐    ┌──────────────┐    ┌────────▼────────┐
│     SCADA       │───▶│   Savegress  │───▶│    ML Engine    │
│    Systems      │    │     CDC      │    │  (Predictions)  │
└─────────────────┘    └──────────────┘    └────────┬────────┘
                                                     │
                    ┌────────────────────────────────┼────────────────────────────────┐
                    │                                │                                │
           ┌────────▼────────┐            ┌─────────▼────────┐            ┌──────────▼──────────┐
           │  Operations     │            │   Maintenance    │            │    Digital Twin     │
           │   Dashboard     │            │   Management     │            │   Visualization     │
           └─────────────────┘            └──────────────────┘            └─────────────────────┘
```

### Edge Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Edge Node                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Protocol  │  │    Data     │  │      ML Inference       │ │
│  │   Adapters  │  │  Pipeline   │  │        Engine           │ │
│  │  MQTT/OPC   │  │             │  │   (TensorRT/ONNX)       │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                │                      │               │
│  ┌──────▼────────────────▼──────────────────────▼──────────────┐ │
│  │                   Local Message Bus                         │ │
│  └────────────────────────┬───────────────────────────────────┘ │
│                           │                                     │
│  ┌────────────────────────▼───────────────────────────────────┐ │
│  │              Store & Forward Buffer                        │ │
│  │              (SQLite/RocksDB)                              │ │
│  └────────────────────────┬───────────────────────────────────┘ │
│                           │                                     │
│  ┌────────────────────────▼───────────────────────────────────┐ │
│  │              Cloud Sync Agent                              │ │
│  │              (MQTT/HTTPS)                                   │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  IoTSense Cloud │
                    └─────────────────┘
```

## API Reference

### Sensor Data API

```http
POST /api/v1/iotsense/telemetry
Content-Type: application/json

{
  "device_id": "sensor_001",
  "timestamp": "2024-01-15T10:30:00.123Z",
  "measurements": [
    {
      "name": "temperature",
      "value": 72.5,
      "unit": "celsius",
      "quality": "good"
    },
    {
      "name": "vibration_rms",
      "value": 2.3,
      "unit": "mm/s",
      "quality": "good"
    }
  ],
  "metadata": {
    "plant": "factory_1",
    "line": "assembly_1",
    "equipment": "motor_001"
  }
}
```

### Predictive Maintenance API

```http
GET /api/v1/iotsense/maintenance/predictions
  ?equipment_id=motor_001
  &horizon_days=30
```

**Response:**
```json
{
  "equipment_id": "motor_001",
  "predictions": [
    {
      "failure_mode": "bearing_failure",
      "probability": 0.73,
      "estimated_rul_days": 12,
      "confidence_interval": [8, 18],
      "recommended_action": "schedule_bearing_replacement",
      "factors": [
        {
          "name": "vibration_increase",
          "contribution": 0.45,
          "trend": "increasing"
        },
        {
          "name": "temperature_anomaly",
          "contribution": 0.28,
          "trend": "stable"
        }
      ]
    }
  ],
  "health_score": 62,
  "next_scheduled_maintenance": "2024-01-28"
}
```

### Digital Twin API

```http
GET /api/v1/iotsense/twins/{twin_id}/state
```

**Response:**
```json
{
  "twin_id": "production_line_1",
  "timestamp": "2024-01-15T10:30:00.123Z",
  "components": [
    {
      "id": "conveyor_1",
      "state": {
        "speed": 1.8,
        "motor_temperature": 45.2,
        "status": "running"
      },
      "health": {
        "score": 95,
        "status": "healthy"
      }
    },
    {
      "id": "robot_arm_1",
      "state": {
        "joint_positions": [0, 45, -30, 0, 90, 0],
        "gripper_state": "closed",
        "cycle_count_today": 1523
      },
      "health": {
        "score": 88,
        "status": "healthy"
      }
    }
  ],
  "kpis": {
    "oee": 82.5,
    "throughput": 450,
    "quality_rate": 99.2
  }
}
```

```http
POST /api/v1/iotsense/twins/{twin_id}/simulate
Content-Type: application/json

{
  "scenario": "speed_optimization",
  "parameters": {
    "conveyor_1.speed": 2.2,
    "robot_arm_1.cycle_time": 3.5
  },
  "duration_minutes": 60,
  "metrics": ["throughput", "energy_consumption", "quality_rate"]
}
```

**Response:**
```json
{
  "simulation_id": "sim_abc123",
  "scenario": "speed_optimization",
  "results": {
    "throughput": {
      "baseline": 450,
      "simulated": 528,
      "change_percent": 17.3
    },
    "energy_consumption": {
      "baseline": 125,
      "simulated": 142,
      "change_percent": 13.6,
      "unit": "kWh"
    },
    "quality_rate": {
      "baseline": 99.2,
      "simulated": 98.8,
      "change_percent": -0.4
    }
  },
  "recommendation": "increase_conveyor_speed",
  "confidence": 0.87
}
```

### OEE API

```http
GET /api/v1/iotsense/oee
  ?equipment_id=cnc_machine_1
  &from=2024-01-01
  &to=2024-01-31
  &granularity=shift
```

**Response:**
```json
{
  "equipment_id": "cnc_machine_1",
  "period": {
    "from": "2024-01-01",
    "to": "2024-01-31"
  },
  "summary": {
    "oee": 78.5,
    "availability": 92.3,
    "performance": 88.7,
    "quality": 96.1
  },
  "breakdown": [
    {
      "date": "2024-01-01",
      "shift": "day",
      "oee": 82.1,
      "availability": 95.0,
      "performance": 90.2,
      "quality": 95.8,
      "downtime_events": [
        {
          "reason": "material_shortage",
          "duration_minutes": 15
        }
      ]
    }
  ],
  "top_losses": [
    {
      "category": "availability",
      "reason": "unplanned_maintenance",
      "total_minutes": 450,
      "impact_oee": 3.2
    },
    {
      "category": "performance",
      "reason": "minor_stoppages",
      "total_minutes": 320,
      "impact_oee": 2.8
    }
  ]
}
```

## Configuration

### IoTSense Configuration File

```yaml
# iotsense.yaml
iotsense:
  # Data Ingestion
  ingestion:
    mqtt:
      brokers:
        - url: "mqtt://broker.factory.local:1883"
          client_id: "iotsense_${HOSTNAME}"
          qos: 1
      topics:
        - "sensors/#"
        - "machines/#"

    opcua:
      endpoints:
        - url: "opc.tcp://plc1.factory.local:4840"
          security_policy: "Basic256Sha256"
          certificate: "/etc/iotsense/certs/client.crt"

    modbus:
      devices:
        - host: "192.168.1.100"
          port: 502
          polling_interval_ms: 1000

  # Time Series Storage
  storage:
    type: "timescaledb"
    host: "timeseries.factory.local"
    database: "iotsense"
    retention:
      raw: "7d"
      1m_aggregates: "90d"
      1h_aggregates: "2y"
      1d_aggregates: "10y"

  # Predictive Maintenance
  predictive_maintenance:
    enabled: true
    models_path: "/var/lib/iotsense/models"
    prediction_interval: "1h"
    training:
      schedule: "weekly"
      min_samples: 10000

  # Digital Twins
  digital_twins:
    enabled: true
    sync_interval_ms: 100
    simulation_engine: "gazebo"

  # Edge Computing
  edge:
    enabled: true
    nodes:
      - id: "edge_1"
        endpoint: "https://edge1.factory.local:8443"
    sync_interval_s: 60
    offline_buffer_gb: 10

  # Alerting
  alerts:
    channels:
      opcenter:
        url: "https://opcenter.factory.local/api"
        api_key: "${OPCENTER_API_KEY}"
      email:
        smtp: "smtp.factory.local"
      sms:
        provider: "twilio"
        account_sid: "${TWILIO_SID}"

  # OEE Tracking
  oee:
    enabled: true
    shift_calendar: "/etc/iotsense/shifts.yaml"
    downtime_reasons: "/etc/iotsense/downtime_codes.yaml"
```

## Integration Examples

### MQTT Sensor Integration

```python
import paho.mqtt.client as mqtt
from iotsense import IoTSense

iot = IoTSense(api_key=os.environ['IOTSENSE_API_KEY'])

def on_message(client, userdata, msg):
    # Parse sensor data
    data = json.loads(msg.payload)

    # Send to IoTSense
    iot.telemetry.send({
        'device_id': data['device_id'],
        'timestamp': data['timestamp'],
        'measurements': [
            {
                'name': 'temperature',
                'value': data['temperature'],
                'unit': 'celsius'
            },
            {
                'name': 'humidity',
                'value': data['humidity'],
                'unit': 'percent'
            }
        ]
    })

client = mqtt.Client()
client.on_message = on_message
client.connect("broker.factory.local", 1883)
client.subscribe("sensors/#")
client.loop_forever()
```

### OPC-UA PLC Integration

```go
package main

import (
    "github.com/gopcua/opcua"
    "github.com/savegress/iotsense-go"
)

func main() {
    // Connect to PLC
    c := opcua.NewClient("opc.tcp://plc.factory.local:4840")
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Initialize IoTSense
    iot := iotsense.NewClient(
        iotsense.WithAPIKey(os.Getenv("IOTSENSE_API_KEY")),
    )

    // Subscribe to nodes
    nodes := []string{
        "ns=2;s=Motor1.Speed",
        "ns=2;s=Motor1.Temperature",
        "ns=2;s=Motor1.Current",
    }

    sub, err := c.Subscribe(&opcua.SubscriptionParameters{
        Interval: 100 * time.Millisecond,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, nodeID := range nodes {
        sub.Monitor(opcua.NodeID(nodeID), func(v *opcua.DataValue) {
            iot.Telemetry.Send(ctx, &iotsense.Telemetry{
                DeviceID:  "motor_001",
                Timestamp: time.Now(),
                Measurements: []iotsense.Measurement{
                    {
                        Name:  nodeID,
                        Value: v.Value.Float(),
                    },
                },
            })
        })
    }

    select {}
}
```

### Predictive Maintenance Webhook

```javascript
const express = require('express');
const app = express();

app.post('/webhooks/iotsense', (req, res) => {
  const { type, data } = req.body;

  switch (type) {
    case 'prediction.critical':
      handleCriticalPrediction(data);
      break;
    case 'anomaly.detected':
      handleAnomaly(data);
      break;
    case 'oee.threshold_breached':
      handleOEEAlert(data);
      break;
  }

  res.status(200).send('OK');
});

async function handleCriticalPrediction(data) {
  const { equipment_id, failure_mode, rul_days, probability } = data;

  // Create work order in CMMS
  await cmms.createWorkOrder({
    equipment: equipment_id,
    type: 'preventive',
    priority: rul_days < 7 ? 'emergency' : 'high',
    description: `Predicted ${failure_mode} - RUL: ${rul_days} days (${Math.round(probability * 100)}% confidence)`,
    due_date: new Date(Date.now() + rul_days * 0.5 * 24 * 60 * 60 * 1000),
  });

  // Check spare parts
  const parts = await inventory.checkStock(equipment_id, failure_mode);
  if (parts.some(p => p.quantity < p.min_quantity)) {
    await procurement.createPurchaseRequest(
      parts.filter(p => p.quantity < p.min_quantity)
    );
  }

  // Notify maintenance team
  await slack.send('#maintenance', {
    text: `⚠️ Predictive maintenance alert: ${equipment_id}`,
    attachments: [{
      color: rul_days < 7 ? 'danger' : 'warning',
      fields: [
        { title: 'Failure Mode', value: failure_mode },
        { title: 'RUL', value: `${rul_days} days` },
        { title: 'Probability', value: `${Math.round(probability * 100)}%` },
      ]
    }]
  });
}
```

## Pricing

| Tier | Devices | Data Points/Day | Features | Price |
|------|---------|-----------------|----------|-------|
| **Starter** | Up to 100 | 10M | Basic Telemetry, Dashboards | $499/mo |
| **Professional** | Up to 1,000 | 100M | + Predictive Maintenance, OEE | $1,999/mo |
| **Enterprise** | Unlimited | Unlimited | + Digital Twins, Edge, Custom ML | Custom |

**Add-ons:**
- Edge Computing Runtime: +$99/node/mo
- Custom ML Model Training: $10,000 one-time
- Digital Twin Development: Custom
- OPC-UA Server License: +$199/mo
- 24/7 Operations Support: +$999/mo

## Industry Certifications

- ISO 27001 (Information Security)
- IEC 62443 (Industrial Cybersecurity)
- ISA-95 Compliant
- OPC Foundation Certified

## Roadmap

### Q1 2024
- [x] MQTT/OPC-UA ingestion
- [x] Basic anomaly detection
- [x] OEE dashboards

### Q2 2024
- [ ] Edge computing runtime
- [ ] Digital twin framework
- [ ] Predictive maintenance v2

### Q3 2024
- [ ] AR/VR visualization integration
- [ ] Autonomous maintenance scheduling
- [ ] Energy optimization module

### Q4 2024
- [ ] Fleet-wide asset management
- [ ] Supply chain integration
- [ ] Carbon footprint tracking

# Installation Guide

This guide covers production installation of Savegress.

## System Requirements

### Minimum Requirements

| Component | Requirement |
|-----------|-------------|
| CPU | 2 cores |
| RAM | 2 GB |
| Disk | 10 GB SSD |
| OS | Linux (amd64, arm64), macOS, Windows |

### Recommended (Production)

| Component | Requirement |
|-----------|-------------|
| CPU | 4+ cores |
| RAM | 8+ GB |
| Disk | 100+ GB SSD (NVMe preferred) |
| Network | Low latency to database |

### Scaling Guidelines

| Throughput | CPU | RAM | Disk |
|------------|-----|-----|------|
| 1K events/sec | 2 cores | 2 GB | 10 GB |
| 10K events/sec | 4 cores | 4 GB | 50 GB |
| 50K events/sec | 8 cores | 8 GB | 100 GB |
| 100K+ events/sec | 16 cores | 16 GB | 500 GB |

## Installation Methods

### 1. Docker (Recommended)

```bash
# Pull the latest images
docker pull savegress/engine:latest
docker pull savegress/broker:latest

# Run engine
docker run -d \
  --name savegress-engine \
  --restart unless-stopped \
  -v /etc/savegress:/etc/savegress \
  -v /var/lib/savegress:/var/lib/savegress \
  savegress/engine:latest

# Run broker
docker run -d \
  --name savegress-broker \
  --restart unless-stopped \
  -p 9092:9092 \
  -p 8080:8080 \
  -v /var/lib/savegress/broker:/data \
  savegress/broker:latest
```

### 2. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  engine:
    image: savegress/engine:latest
    restart: unless-stopped
    volumes:
      - ./config/engine.yaml:/etc/savegress/config.yaml
      - engine-data:/var/lib/savegress
    environment:
      - SAVEGRESS_LOG_LEVEL=info
      - SAVEGRESS_LICENSE_KEY=${LICENSE_KEY}
    depends_on:
      - broker

  broker:
    image: savegress/broker:latest
    restart: unless-stopped
    ports:
      - "9092:9092"   # gRPC
      - "8080:8080"   # HTTP/Metrics
    volumes:
      - ./config/broker.yaml:/etc/savegress/config.yaml
      - broker-data:/data
    environment:
      - SAVEGRESS_LOG_LEVEL=info

volumes:
  engine-data:
  broker-data:
```

```bash
docker-compose up -d
```

### 3. Binary Installation

```bash
# Download latest release
VERSION=$(curl -s https://api.github.com/repos/savegress/savegress/releases/latest | grep tag_name | cut -d '"' -f 4)

# Linux amd64
curl -LO https://github.com/savegress/savegress/releases/download/${VERSION}/savegress-linux-amd64.tar.gz
tar -xzf savegress-linux-amd64.tar.gz

# Linux arm64
curl -LO https://github.com/savegress/savegress/releases/download/${VERSION}/savegress-linux-arm64.tar.gz
tar -xzf savegress-linux-arm64.tar.gz

# macOS
curl -LO https://github.com/savegress/savegress/releases/download/${VERSION}/savegress-darwin-amd64.tar.gz
tar -xzf savegress-darwin-amd64.tar.gz

# Install
sudo mv savegress-engine savegress-broker /usr/local/bin/
sudo chmod +x /usr/local/bin/savegress-*
```

### 4. Systemd Service

```ini
# /etc/systemd/system/savegress-engine.service
[Unit]
Description=Savegress CDC Engine
After=network.target

[Service]
Type=simple
User=savegress
Group=savegress
ExecStart=/usr/local/bin/savegress-engine --config /etc/savegress/engine.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/savegress /var/log/savegress

[Install]
WantedBy=multi-user.target
```

```ini
# /etc/systemd/system/savegress-broker.service
[Unit]
Description=Savegress CDC Broker
After=network.target

[Service]
Type=simple
User=savegress
Group=savegress
ExecStart=/usr/local/bin/savegress-broker --config /etc/savegress/broker.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/savegress /var/log/savegress

[Install]
WantedBy=multi-user.target
```

```bash
# Create user and directories
sudo useradd -r -s /bin/false savegress
sudo mkdir -p /etc/savegress /var/lib/savegress /var/log/savegress
sudo chown savegress:savegress /var/lib/savegress /var/log/savegress

# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable savegress-engine savegress-broker
sudo systemctl start savegress-engine savegress-broker

# Check status
sudo systemctl status savegress-engine
sudo systemctl status savegress-broker
```

### 5. Kubernetes

```yaml
# savegress-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: savegress-engine
spec:
  replicas: 1
  selector:
    matchLabels:
      app: savegress-engine
  template:
    metadata:
      labels:
        app: savegress-engine
    spec:
      containers:
      - name: engine
        image: savegress/engine:latest
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        volumeMounts:
        - name: config
          mountPath: /etc/savegress
        - name: data
          mountPath: /var/lib/savegress
        env:
        - name: SAVEGRESS_LICENSE_KEY
          valueFrom:
            secretKeyRef:
              name: savegress-secrets
              key: license-key
      volumes:
      - name: config
        configMap:
          name: savegress-config
      - name: data
        persistentVolumeClaim:
          claimName: savegress-data
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: savegress-config
data:
  config.yaml: |
    source:
      type: postgres
      host: postgres-service
      port: 5432
      database: mydb
      user: cdc_user
      password: ${DB_PASSWORD}
    output:
      type: grpc
      broker_address: savegress-broker:9092
---
apiVersion: v1
kind: Secret
metadata:
  name: savegress-secrets
type: Opaque
data:
  license-key: <base64-encoded-license>
  db-password: <base64-encoded-password>
```

See [Kubernetes Guide](../operations/kubernetes.md) for full Helm charts and operators.

## Directory Structure

```
/etc/savegress/              # Configuration
├── engine.yaml              # Engine config
├── broker.yaml              # Broker config
├── license.key              # License file (Pro/Enterprise)
└── certs/                   # TLS certificates
    ├── ca.crt
    ├── server.crt
    └── server.key

/var/lib/savegress/          # Data
├── engine/
│   ├── checkpoints/         # Position tracking
│   └── state/               # Engine state
└── broker/
    ├── data/                # Message storage
    ├── dlq/                 # Dead letter queue
    └── wal/                 # Write-ahead log

/var/log/savegress/          # Logs
├── engine.log
└── broker.log
```

## Configuration Files

### Engine Configuration

```yaml
# /etc/savegress/engine.yaml
source:
  type: postgres
  host: db.example.com
  port: 5432
  database: production
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  slot_name: savegress_slot
  publication: savegress_pub

output:
  type: grpc
  broker_address: localhost:9092

checkpoint:
  dir: /var/lib/savegress/engine/checkpoints
  interval: 10s

logging:
  level: info
  format: json
  output: /var/log/savegress/engine.log

metrics:
  enabled: true
  address: :8081
```

### Broker Configuration

```yaml
# /etc/savegress/broker.yaml
server:
  grpc_address: :9092
  http_address: :8080

storage:
  data_dir: /var/lib/savegress/broker/data
  segment_size: 1GB
  retention: 7d

dlq:
  enabled: true
  retention_days: 14
  max_messages: 1000000

logging:
  level: info
  format: json
  output: /var/log/savegress/broker.log

metrics:
  enabled: true
  prometheus_path: /metrics
```

## License Activation

### Community Edition

No license required. Community features work out of the box.

### Pro/Enterprise Edition

```bash
# Set license via environment variable
export SAVEGRESS_LICENSE_KEY="eyJhbGciOiJFZDI1NTE5..."

# Or via config file
echo "eyJhbGciOiJFZDI1NTE5..." > /etc/savegress/license.key

# Or via command line
savegress-engine --license "eyJhbGciOiJFZDI1NTE5..."
```

Verify license:

```bash
savegress-engine --license-info

# Output:
# License: Pro
# Customer: Acme Corp
# Expires: 2026-01-15
# Features: compression, dlq, prometheus, webhook, kafka...
# Limits: 10 sources, 100 tables, 50000 events/sec
```

## Health Checks

```bash
# Engine health
curl http://localhost:8081/health

# Broker health
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics
```

## Verification

```bash
# Check services are running
systemctl status savegress-engine savegress-broker

# Check logs
journalctl -u savegress-engine -f
journalctl -u savegress-broker -f

# Test connection
savegress-engine --test-connection

# View metrics
curl -s http://localhost:8080/metrics | grep savegress
```

## Next Steps

- [First Pipeline](first-pipeline.md) - Create your first CDC pipeline
- [Configuration Reference](../configuration/reference.md) - All configuration options
- [Security Setup](../features/security.md) - TLS, authentication
- [Monitoring](../operations/monitoring.md) - Prometheus & Grafana

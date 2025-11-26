package services

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/models"
)

// ConfigGeneratorService generates deployment configurations
type ConfigGeneratorService struct {
	connectionService *ConnectionService
	pipelineService   *PipelineService
}

// NewConfigGeneratorService creates a new config generator service
func NewConfigGeneratorService(connService *ConnectionService, pipelineService *PipelineService) *ConfigGeneratorService {
	return &ConfigGeneratorService{
		connectionService: connService,
		pipelineService:   pipelineService,
	}
}

// GenerateConfig generates deployment configuration
func (s *ConfigGeneratorService) GenerateConfig(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
	var pipeline *models.Pipeline
	var sourceConn *models.Connection

	if pipelineID != nil {
		var err error
		pipeline, err = s.pipelineService.GetPipeline(ctx, userID, *pipelineID)
		if err != nil {
			return "", fmt.Errorf("pipeline not found: %w", err)
		}

		sourceConn, err = s.connectionService.GetConnectionWithPassword(ctx, pipeline.SourceConnID)
		if err != nil {
			return "", fmt.Errorf("source connection not found: %w", err)
		}
	}

	switch format {
	case "docker-compose", "docker":
		return s.generateDockerCompose(pipeline, sourceConn, licenseKey)
	case "helm", "kubernetes", "k8s":
		return s.generateHelmValues(pipeline, sourceConn, licenseKey)
	case "env", "dotenv":
		return s.generateEnvFile(pipeline, sourceConn, licenseKey)
	case "systemd":
		return s.generateSystemdUnit(pipeline, sourceConn, licenseKey)
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func (s *ConfigGeneratorService) generateDockerCompose(pipeline *models.Pipeline, sourceConn *models.Connection, licenseKey string) (string, error) {
	tmpl := `# Savegress CDC Engine - Docker Compose Configuration
# Generated for your pipeline
# Documentation: https://docs.savegress.io/installation/docker

version: '3.8'

services:
  savegress-engine:
    image: savegress/cdc-engine:latest
    container_name: savegress-engine
    restart: unless-stopped
    environment:
      # License
      - SAVEGRESS_LICENSE_KEY={{.LicenseKey}}
      - SAVEGRESS_PLATFORM_URL=https://api.savegress.com

      # Source Database
      - CDC_SOURCE_TYPE={{.SourceType}}
      - CDC_SOURCE_HOST={{.SourceHost}}
      - CDC_SOURCE_PORT={{.SourcePort}}
      - CDC_SOURCE_DATABASE={{.SourceDatabase}}
      - CDC_SOURCE_USER={{.SourceUser}}
      - CDC_SOURCE_PASSWORD=${SOURCE_DB_PASSWORD}
      {{if .SourceSSL}}- CDC_SOURCE_SSL_MODE={{.SourceSSL}}{{end}}

      # Output Configuration
      - CDC_OUTPUT_TYPE={{.OutputType}}
      {{if .OutputURL}}- CDC_OUTPUT_URL={{.OutputURL}}{{end}}

      # Tables to replicate (comma-separated)
      {{if .Tables}}- CDC_TABLES={{.Tables}}{{end}}

      # Performance
      - CDC_BATCH_SIZE=100
      - CDC_BATCH_TIMEOUT_MS=50

    volumes:
      - savegress-data:/var/lib/savegress
    ports:
      - "9090:9090"  # Metrics endpoint
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  savegress-data:

# To run:
# 1. Set environment variable: export SOURCE_DB_PASSWORD=your_password
# 2. Run: docker-compose up -d
# 3. Check logs: docker-compose logs -f savegress-engine
`

	data := map[string]interface{}{
		"LicenseKey":     licenseKey,
		"SourceType":     "postgres",
		"SourceHost":     "your-database-host",
		"SourcePort":     "5432",
		"SourceDatabase": "your_database",
		"SourceUser":     "replication_user",
		"SourceSSL":      "prefer",
		"OutputType":     "http",
		"OutputURL":      "",
		"Tables":         "",
	}

	if sourceConn != nil {
		data["SourceType"] = sourceConn.Type
		data["SourceHost"] = sourceConn.Host
		data["SourcePort"] = fmt.Sprintf("%d", sourceConn.Port)
		data["SourceDatabase"] = sourceConn.Database
		data["SourceUser"] = sourceConn.Username
		data["SourceSSL"] = sourceConn.SSLMode
	}

	if pipeline != nil {
		data["OutputType"] = pipeline.TargetType
		if url, ok := pipeline.TargetConfig["url"]; ok {
			data["OutputURL"] = url
		}
		if len(pipeline.Tables) > 0 {
			data["Tables"] = strings.Join(pipeline.Tables, ",")
		}
	}

	t, err := template.New("docker-compose").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *ConfigGeneratorService) generateHelmValues(pipeline *models.Pipeline, sourceConn *models.Connection, licenseKey string) (string, error) {
	tmpl := `# Savegress CDC Engine - Helm Values
# Generated for your pipeline
# Documentation: https://docs.savegress.io/installation/kubernetes
#
# Install with:
#   helm repo add savegress https://charts.savegress.io
#   helm install cdc-engine savegress/cdc-engine -f values.yaml

# License configuration
license:
  key: "{{.LicenseKey}}"
  # Or use existing secret:
  # existingSecret: "savegress-license"
  # secretKey: "license-key"

# Source database configuration
source:
  type: {{.SourceType}}
  host: {{.SourceHost}}
  port: {{.SourcePort}}
  database: {{.SourceDatabase}}
  username: {{.SourceUser}}
  # Password from secret:
  existingSecret: "source-db-credentials"
  secretKey: "password"
  sslMode: {{.SourceSSL}}

# Output configuration
output:
  type: {{.OutputType}}
  {{if .OutputURL}}url: {{.OutputURL}}{{end}}

# Tables to replicate
tables:
{{range .TablesList}}  - {{.}}
{{end}}

# Resource limits
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "1Gi"
    cpu: "1000m"

# Metrics
metrics:
  enabled: true
  port: 9090
  serviceMonitor:
    enabled: false  # Enable if using Prometheus Operator

# Persistence for checkpoints
persistence:
  enabled: true
  size: 1Gi
  storageClass: ""  # Use default

# Pod configuration
replicaCount: 1
nodeSelector: {}
tolerations: []
affinity: {}
`

	data := map[string]interface{}{
		"LicenseKey":     licenseKey,
		"SourceType":     "postgres",
		"SourceHost":     "your-database-host",
		"SourcePort":     5432,
		"SourceDatabase": "your_database",
		"SourceUser":     "replication_user",
		"SourceSSL":      "prefer",
		"OutputType":     "http",
		"OutputURL":      "",
		"TablesList":     []string{"public.*"},
	}

	if sourceConn != nil {
		data["SourceType"] = sourceConn.Type
		data["SourceHost"] = sourceConn.Host
		data["SourcePort"] = sourceConn.Port
		data["SourceDatabase"] = sourceConn.Database
		data["SourceUser"] = sourceConn.Username
		data["SourceSSL"] = sourceConn.SSLMode
	}

	if pipeline != nil {
		data["OutputType"] = pipeline.TargetType
		if url, ok := pipeline.TargetConfig["url"]; ok {
			data["OutputURL"] = url
		}
		if len(pipeline.Tables) > 0 {
			data["TablesList"] = pipeline.Tables
		}
	}

	t, err := template.New("helm").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *ConfigGeneratorService) generateEnvFile(pipeline *models.Pipeline, sourceConn *models.Connection, licenseKey string) (string, error) {
	tmpl := `# Savegress CDC Engine - Environment Configuration
# Generated for your pipeline
# Documentation: https://docs.savegress.io/configuration

# License
SAVEGRESS_LICENSE_KEY={{.LicenseKey}}
SAVEGRESS_PLATFORM_URL=https://api.savegress.com

# Source Database
CDC_SOURCE_TYPE={{.SourceType}}
CDC_SOURCE_HOST={{.SourceHost}}
CDC_SOURCE_PORT={{.SourcePort}}
CDC_SOURCE_DATABASE={{.SourceDatabase}}
CDC_SOURCE_USER={{.SourceUser}}
CDC_SOURCE_PASSWORD=your_password_here
CDC_SOURCE_SSL_MODE={{.SourceSSL}}

# Output Configuration
CDC_OUTPUT_TYPE={{.OutputType}}
{{if .OutputURL}}CDC_OUTPUT_URL={{.OutputURL}}{{end}}

# Tables (comma-separated)
{{if .Tables}}CDC_TABLES={{.Tables}}{{end}}

# Performance Tuning
CDC_BATCH_SIZE=100
CDC_BATCH_TIMEOUT_MS=50

# Checkpoint Configuration
CDC_CHECKPOINT_TYPE=file
CDC_CHECKPOINT_PATH=/var/lib/savegress/checkpoint

# Logging
CDC_LOG_LEVEL=info
CDC_LOG_FORMAT=json
`

	data := map[string]interface{}{
		"LicenseKey":     licenseKey,
		"SourceType":     "postgres",
		"SourceHost":     "your-database-host",
		"SourcePort":     5432,
		"SourceDatabase": "your_database",
		"SourceUser":     "replication_user",
		"SourceSSL":      "prefer",
		"OutputType":     "http",
		"OutputURL":      "",
		"Tables":         "",
	}

	if sourceConn != nil {
		data["SourceType"] = sourceConn.Type
		data["SourceHost"] = sourceConn.Host
		data["SourcePort"] = sourceConn.Port
		data["SourceDatabase"] = sourceConn.Database
		data["SourceUser"] = sourceConn.Username
		data["SourceSSL"] = sourceConn.SSLMode
	}

	if pipeline != nil {
		data["OutputType"] = pipeline.TargetType
		if url, ok := pipeline.TargetConfig["url"]; ok {
			data["OutputURL"] = url
		}
		if len(pipeline.Tables) > 0 {
			data["Tables"] = strings.Join(pipeline.Tables, ",")
		}
	}

	t, err := template.New("env").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *ConfigGeneratorService) generateSystemdUnit(pipeline *models.Pipeline, sourceConn *models.Connection, licenseKey string) (string, error) {
	tmpl := `# Savegress CDC Engine - Systemd Service Unit
# Generated for your pipeline
# Documentation: https://docs.savegress.io/installation/binary
#
# Installation:
# 1. Download binary: curl -L https://releases.savegress.io/cdc-engine/latest/cdc-engine-linux-amd64 -o /usr/local/bin/cdc-engine
# 2. Make executable: chmod +x /usr/local/bin/cdc-engine
# 3. Create config: cp savegress.env /etc/savegress/savegress.env
# 4. Install service: cp savegress.service /etc/systemd/system/
# 5. Enable service: systemctl enable savegress && systemctl start savegress

[Unit]
Description=Savegress CDC Engine
Documentation=https://docs.savegress.io
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=savegress
Group=savegress
EnvironmentFile=/etc/savegress/savegress.env
ExecStart=/usr/local/bin/cdc-engine
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=savegress

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/savegress

# Resource limits
LimitNOFILE=65535
MemoryMax=1G

[Install]
WantedBy=multi-user.target
`

	return tmpl, nil
}

// GenerateQuickStart generates a quick start guide
func (s *ConfigGeneratorService) GenerateQuickStart(licenseKey string, sourceType string) string {
	return fmt.Sprintf(`# Savegress Quick Start Guide

## 1. Pull the Docker image
docker pull savegress/cdc-engine:latest

## 2. Run with your configuration
docker run -d \
  --name savegress-engine \
  -e SAVEGRESS_LICENSE_KEY=%s \
  -e CDC_SOURCE_TYPE=%s \
  -e CDC_SOURCE_HOST=your-database-host \
  -e CDC_SOURCE_PORT=5432 \
  -e CDC_SOURCE_DATABASE=your_database \
  -e CDC_SOURCE_USER=replication_user \
  -e CDC_SOURCE_PASSWORD=your_password \
  -e CDC_OUTPUT_TYPE=http \
  -e CDC_OUTPUT_URL=https://your-endpoint.com/events \
  -p 9090:9090 \
  savegress/cdc-engine:latest

## 3. Check the status
docker logs -f savegress-engine

## 4. View metrics
curl http://localhost:9090/metrics

## Need help?
- Documentation: https://docs.savegress.io
- Dashboard: https://savegress.com/dashboard
- Support: support@savegress.io
`, licenseKey, sourceType)
}

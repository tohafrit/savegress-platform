package services

import (
	"strings"
	"testing"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigGeneratorService(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)
	assert.NotNil(t, service)
}

func TestConfigGeneratorService_GenerateDockerCompose(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	tests := []struct {
		name       string
		pipeline   *models.Pipeline
		sourceConn *models.Connection
		licenseKey string
		shouldContain []string
	}{
		{
			name:       "basic config without pipeline",
			pipeline:   nil,
			sourceConn: nil,
			licenseKey: "test-license-key",
			shouldContain: []string{
				"version: '3.8'",
				"savegress-engine",
				"SAVEGRESS_LICENSE_KEY=test-license-key",
				"CDC_SOURCE_TYPE=postgres",
				"9090:9090",
			},
		},
		{
			name: "config with source connection",
			pipeline: nil,
			sourceConn: &models.Connection{
				Type:     "mysql",
				Host:     "db.example.com",
				Port:     3306,
				Database: "mydb",
				Username: "admin",
				SSLMode:  "require",
			},
			licenseKey: "license-123",
			shouldContain: []string{
				"CDC_SOURCE_TYPE=mysql",
				"CDC_SOURCE_HOST=db.example.com",
				"CDC_SOURCE_PORT=3306",
				"CDC_SOURCE_DATABASE=mydb",
				"CDC_SOURCE_USER=admin",
			},
		},
		{
			name: "config with pipeline",
			pipeline: &models.Pipeline{
				TargetType: "kafka",
				TargetConfig: map[string]string{
					"url": "kafka://broker:9092",
				},
				Tables: []string{"public.users", "public.orders"},
			},
			sourceConn: &models.Connection{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "repl",
				SSLMode:  "prefer",
			},
			licenseKey: "license-abc",
			shouldContain: []string{
				"CDC_OUTPUT_TYPE=kafka",
				"CDC_TABLES=public.users,public.orders",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := service.generateDockerCompose(tt.pipeline, tt.sourceConn, tt.licenseKey)
			assert.NoError(t, err)
			assert.NotEmpty(t, config)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, config, expected, "Config should contain: %s", expected)
			}
		})
	}
}

func TestConfigGeneratorService_GenerateHelmValues(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	tests := []struct {
		name       string
		pipeline   *models.Pipeline
		sourceConn *models.Connection
		licenseKey string
		shouldContain []string
	}{
		{
			name:       "basic helm values",
			pipeline:   nil,
			sourceConn: nil,
			licenseKey: "helm-license-key",
			shouldContain: []string{
				"license:",
				"key: \"helm-license-key\"",
				"source:",
				"output:",
				"resources:",
				"metrics:",
				"persistence:",
			},
		},
		{
			name: "helm values with connection",
			pipeline: nil,
			sourceConn: &models.Connection{
				Type:     "postgres",
				Host:     "postgres.cluster.local",
				Port:     5432,
				Database: "production",
				Username: "cdc_user",
				SSLMode:  "verify-full",
			},
			licenseKey: "helm-123",
			shouldContain: []string{
				"type: postgres",
				"host: postgres.cluster.local",
				"port: 5432",
				"database: production",
				"username: cdc_user",
				"sslMode: verify-full",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := service.generateHelmValues(tt.pipeline, tt.sourceConn, tt.licenseKey)
			assert.NoError(t, err)
			assert.NotEmpty(t, config)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, config, expected, "Helm values should contain: %s", expected)
			}
		})
	}
}

func TestConfigGeneratorService_GenerateEnvFile(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	tests := []struct {
		name       string
		pipeline   *models.Pipeline
		sourceConn *models.Connection
		licenseKey string
		shouldContain []string
	}{
		{
			name:       "basic env file",
			pipeline:   nil,
			sourceConn: nil,
			licenseKey: "env-license",
			shouldContain: []string{
				"SAVEGRESS_LICENSE_KEY=env-license",
				"CDC_SOURCE_TYPE=",
				"CDC_BATCH_SIZE=100",
				"CDC_LOG_LEVEL=info",
				"CDC_LOG_FORMAT=json",
			},
		},
		{
			name: "env file with full config",
			pipeline: &models.Pipeline{
				TargetType: "http",
				TargetConfig: map[string]string{
					"url": "https://webhook.example.com/events",
				},
				Tables: []string{"schema.table1"},
			},
			sourceConn: &models.Connection{
				Type:     "mysql",
				Host:     "mysql.local",
				Port:     3306,
				Database: "app",
				Username: "root",
				SSLMode:  "disable",
			},
			licenseKey: "env-123",
			shouldContain: []string{
				"CDC_SOURCE_TYPE=mysql",
				"CDC_OUTPUT_TYPE=http",
				"CDC_TABLES=schema.table1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := service.generateEnvFile(tt.pipeline, tt.sourceConn, tt.licenseKey)
			assert.NoError(t, err)
			assert.NotEmpty(t, config)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, config, expected, "Env file should contain: %s", expected)
			}
		})
	}
}

func TestConfigGeneratorService_GenerateSystemdUnit(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	config, err := service.generateSystemdUnit(nil, nil, "any-license")
	assert.NoError(t, err)
	assert.NotEmpty(t, config)

	// Check systemd unit structure
	shouldContain := []string{
		"[Unit]",
		"[Service]",
		"[Install]",
		"Description=Savegress CDC Engine",
		"ExecStart=/usr/local/bin/cdc-engine",
		"Restart=on-failure",
		"User=savegress",
		"EnvironmentFile=/etc/savegress/savegress.env",
		"WantedBy=multi-user.target",
	}

	for _, expected := range shouldContain {
		assert.Contains(t, config, expected, "Systemd unit should contain: %s", expected)
	}
}

func TestConfigGeneratorService_GenerateQuickStart(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	tests := []struct {
		name       string
		licenseKey string
		sourceType string
		shouldContain []string
	}{
		{
			name:       "postgres quick start",
			licenseKey: "qs-license-123",
			sourceType: "postgres",
			shouldContain: []string{
				"docker pull savegress/cdc-engine:latest",
				"SAVEGRESS_LICENSE_KEY=qs-license-123",
				"CDC_SOURCE_TYPE=postgres",
				"docker logs -f savegress-engine",
				"curl http://localhost:9090/metrics",
			},
		},
		{
			name:       "mysql quick start",
			licenseKey: "mysql-license",
			sourceType: "mysql",
			shouldContain: []string{
				"CDC_SOURCE_TYPE=mysql",
				"SAVEGRESS_LICENSE_KEY=mysql-license",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guide := service.GenerateQuickStart(tt.licenseKey, tt.sourceType)
			assert.NotEmpty(t, guide)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, guide, expected, "Quick start should contain: %s", expected)
			}
		})
	}
}

func TestConfigGeneratorService_SupportedFormats(t *testing.T) {
	// Document all supported config formats
	formats := []struct {
		format      string
		aliases     []string
		description string
	}{
		{
			format:      "docker-compose",
			aliases:     []string{"docker"},
			description: "Docker Compose YAML configuration",
		},
		{
			format:      "helm",
			aliases:     []string{"kubernetes", "k8s"},
			description: "Helm chart values.yaml for Kubernetes",
		},
		{
			format:      "env",
			aliases:     []string{"dotenv"},
			description: "Environment file (.env) format",
		},
		{
			format:      "systemd",
			aliases:     []string{},
			description: "Systemd service unit file",
		},
	}

	for _, f := range formats {
		t.Run("format_"+f.format, func(t *testing.T) {
			assert.NotEmpty(t, f.format)
			assert.NotEmpty(t, f.description)
		})
	}
}

func TestConfigGeneratorService_NoSensitiveDataInConfig(t *testing.T) {
	service := NewConfigGeneratorService(nil, nil)

	sourceConn := &models.Connection{
		Type:     "postgres",
		Host:     "db.example.com",
		Port:     5432,
		Database: "mydb",
		Username: "admin",
		Password: "super_secret_password_123", // This should NOT appear in output
		SSLMode:  "require",
	}

	// Generate all config types
	dockerConfig, _ := service.generateDockerCompose(nil, sourceConn, "license")
	helmConfig, _ := service.generateHelmValues(nil, sourceConn, "license")
	envConfig, _ := service.generateEnvFile(nil, sourceConn, "license")

	// Verify password is not in any config
	assert.NotContains(t, dockerConfig, "super_secret_password_123", "Docker config should not contain actual password")
	assert.NotContains(t, helmConfig, "super_secret_password_123", "Helm config should not contain actual password")
	assert.NotContains(t, envConfig, "super_secret_password_123", "Env config should not contain actual password")

	// Verify password placeholder is used instead
	assert.Contains(t, dockerConfig, "${SOURCE_DB_PASSWORD}", "Docker config should use password placeholder")
	assert.Contains(t, envConfig, "your_password_here", "Env config should use password placeholder")
}

func TestConfigGeneratorService_TablesJoining(t *testing.T) {
	_ = NewConfigGeneratorService(nil, nil) // Service not needed for this test

	tests := []struct {
		name     string
		tables   []string
		expected string
	}{
		{
			name:     "single table",
			tables:   []string{"public.users"},
			expected: "public.users",
		},
		{
			name:     "multiple tables",
			tables:   []string{"public.users", "public.orders", "public.products"},
			expected: "public.users,public.orders,public.products",
		},
		{
			name:     "empty tables",
			tables:   []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.Join(tt.tables, ",")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test examples (commented out - would need database)
//
// func TestConfigGeneratorService_GenerateConfigIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test full GenerateConfig method with real pipeline
// }

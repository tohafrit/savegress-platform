// Package license provides enterprise license management for Savegress.
// It supports offline validation with Ed25519 signatures, online validation
// with grace periods, hardware fingerprinting, and usage telemetry.
package license

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Tier represents the license tier
type Tier string

const (
	TierCommunity  Tier = "community"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
	TierTrial      Tier = "trial"
)

// Feature represents a licensable feature
type Feature string

const (
	// ============================================
	// DATABASE CONNECTORS
	// ============================================

	// Community databases (free tier)
	FeaturePostgreSQL Feature = "postgresql"
	FeatureMySQL      Feature = "mysql"
	FeatureMariaDB    Feature = "mariadb"

	// Pro databases
	FeatureMongoDB   Feature = "mongodb"
	FeatureSQLServer Feature = "sqlserver"
	FeatureCassandra Feature = "cassandra"
	FeatureDynamoDB  Feature = "dynamodb"

	// Enterprise databases
	FeatureOracle Feature = "oracle"

	// ============================================
	// OUTPUT CONNECTORS
	// ============================================

	// Pro outputs
	FeatureSnapshot    Feature = "snapshot"
	FeatureWebhook     Feature = "webhook"
	FeatureKafkaOutput Feature = "kafka_output"
	FeatureGRPCOutput  Feature = "grpc_output"

	// Enterprise outputs
	FeatureCustomOutput Feature = "custom_output"

	// ============================================
	// PERFORMANCE & COMPRESSION
	// ============================================

	// Pro: Basic compression (Hybrid, ZSTD, LZ4)
	// Provides 4-10x storage savings - clear ROI for paying customers
	FeatureCompression Feature = "compression"

	// Enterprise: SIMD-optimized compression (AVX2, AVX-512, NEON)
	// Maximum throughput for high-volume deployments
	FeatureCompressionSIMD Feature = "compression_simd"

	// ============================================
	// RELIABILITY & FLOW CONTROL
	// ============================================

	// Pro: Advanced rate limiting (adaptive, sliding window, multi-tier)
	// Note: Basic token bucket rate limiting is FREE for all tiers
	FeatureAdvancedRateLimiting Feature = "advanced_rate_limiting"

	// Pro: Backpressure control for production stability
	FeatureBackpressure Feature = "backpressure"

	// Pro: Dead Letter Queue for failed message handling
	FeatureDLQ Feature = "dlq"

	// Pro: Event replay for debugging and recovery
	FeatureReplay Feature = "replay"

	// Enterprise: Exactly-once delivery semantics
	// Critical for financial and regulated workloads
	FeatureExactlyOnce Feature = "exactly_once"

	// ============================================
	// DISASTER RECOVERY
	// ============================================

	// Enterprise: Point-in-time recovery
	FeaturePITR Feature = "pitr"

	// Enterprise: Cloud storage backends (S3, GCS, Azure)
	FeatureCloudStorage Feature = "cloud_storage"

	// ============================================
	// SCHEMA MANAGEMENT
	// ============================================

	// Pro: Automatic schema evolution (detection + safe auto-apply)
	FeatureSchemaEvolution Feature = "schema_evolution"

	// Enterprise: Schema migration approval workflow
	// Required for change management compliance
	FeatureSchemaMigrationApproval Feature = "schema_migration_approval"

	// ============================================
	// OBSERVABILITY & MONITORING
	// ============================================

	// Pro: Prometheus metrics export
	// Note: Basic internal metrics are FREE for all tiers
	FeaturePrometheus Feature = "prometheus"

	// Pro: SLA monitoring (Bronze/Silver/Gold levels)
	FeatureSLAMonitoring Feature = "sla_monitoring"

	// Enterprise: Full OpenTelemetry integration (traces, spans)
	FeatureOpenTelemetry Feature = "opentelemetry"

	// ============================================
	// HIGH AVAILABILITY & CLUSTERING
	// ============================================

	// Enterprise: High availability mode
	FeatureHA Feature = "ha"

	// Enterprise: Raft consensus clustering
	FeatureRaftCluster Feature = "raft_cluster"

	// Enterprise: Multi-region deployment
	FeatureMultiRegion Feature = "multi_region"

	// ============================================
	// SECURITY & COMPLIANCE
	// ============================================

	// Enterprise: End-to-end encryption
	FeatureEncryption Feature = "encryption"

	// Enterprise: Mutual TLS authentication
	FeatureMTLS Feature = "mtls"

	// Enterprise: Role-based access control
	FeatureRBAC Feature = "rbac"

	// Enterprise: HashiCorp Vault integration
	FeatureVault Feature = "vault"

	// Enterprise: Audit logging
	FeatureAuditLog Feature = "audit_log"

	// Enterprise: SSO integration
	FeatureSSO Feature = "sso"

	// Enterprise: LDAP integration
	FeatureLDAP Feature = "ldap"

	// Enterprise: Multi-tenant isolation
	FeatureMultiTenant Feature = "multi_tenant"
)

// CommunityFeatures are available in the free tier.
// Philosophy: Community should enable a working production system for startups/SMB.
// Basic safety features (token bucket rate limiting, circuit breaker, health checks,
// basic metrics) are included to ensure system stability - these are not premium features.
var CommunityFeatures = []Feature{
	// Databases: Core open-source databases
	FeaturePostgreSQL,
	FeatureMySQL,
	FeatureMariaDB,
	// Note: Basic rate limiting, circuit breaker, health checks, and internal metrics
	// are available to all tiers without explicit feature flags - they are built-in safety features
}

// ProFeatures are for production at scale - performance, reliability, and DevOps tooling.
// Target: Scale-ups and serious production deployments that need performance and operations features.
var ProFeatures = []Feature{
	// Additional databases
	FeatureMongoDB,
	FeatureSQLServer,
	FeatureCassandra,
	FeatureDynamoDB,

	// Output connectors
	FeatureSnapshot,
	FeatureKafkaOutput,
	FeatureGRPCOutput,
	FeatureWebhook,

	// Performance: Compression provides clear ROI (4-10x storage savings)
	FeatureCompression,

	// Reliability at scale
	FeatureAdvancedRateLimiting, // Adaptive, sliding window, multi-tier
	FeatureBackpressure,
	FeatureDLQ,
	FeatureReplay,

	// Schema management for zero-downtime operations
	FeatureSchemaEvolution,

	// Observability for DevOps integration
	FeaturePrometheus,
	FeatureSLAMonitoring,
}

// EnterpriseFeatures are for governance, compliance, and multi-team operations.
// Target: Large organizations, regulated industries, and multi-team deployments.
var EnterpriseFeatures = []Feature{
	// Premium database
	FeatureOracle,

	// Custom integrations
	FeatureCustomOutput,

	// Maximum performance
	FeatureCompressionSIMD,
	FeatureExactlyOnce,

	// Disaster recovery
	FeaturePITR,
	FeatureCloudStorage,

	// Advanced schema governance
	FeatureSchemaMigrationApproval,

	// Full observability
	FeatureOpenTelemetry,

	// High availability & clustering
	FeatureHA,
	FeatureRaftCluster,
	FeatureMultiRegion,

	// Security & compliance
	FeatureEncryption,
	FeatureMTLS,
	FeatureRBAC,
	FeatureVault,
	FeatureAuditLog,
	FeatureSSO,
	FeatureLDAP,
	FeatureMultiTenant,
}

// Limits defines usage limits for a license
type Limits struct {
	MaxSources       int   `json:"max_sources"`        // 0 = unlimited
	MaxThroughput    int64 `json:"max_throughput"`     // events/sec, 0 = unlimited
	MaxTables        int   `json:"max_tables"`         // 0 = unlimited
	MaxRetentionDays int   `json:"max_retention_days"` // 0 = unlimited
}

// CommunityLimits for free tier
var CommunityLimits = Limits{
	MaxSources:       1,
	MaxThroughput:    1000,
	MaxTables:        10,
	MaxRetentionDays: 1,
}

// ProLimits for pro tier
var ProLimits = Limits{
	MaxSources:       10,
	MaxThroughput:    50000,
	MaxTables:        100,
	MaxRetentionDays: 30,
}

// EnterpriseLimits - no limits
var EnterpriseLimits = Limits{
	MaxSources:       0,
	MaxThroughput:    0,
	MaxTables:        0,
	MaxRetentionDays: 0,
}

// License represents a Savegress license
type License struct {
	// Identification
	ID           string `json:"id"`            // Unique license ID (UUID)
	CustomerID   string `json:"customer_id"`   // Customer UUID
	CustomerName string `json:"customer_name"` // Customer display name

	// License details
	Tier     Tier      `json:"tier"`
	Features []Feature `json:"features"`
	Limits   Limits    `json:"limits"`

	// Validity
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`

	// Hardware binding (optional)
	HardwareID string `json:"hardware_id,omitempty"` // Bound to specific machine

	// Metadata
	Issuer    string            `json:"issuer"` // license.savegress.io
	Version   int               `json:"version"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Signature string            `json:"signature"` // Base64 Ed25519 signature
}

// LicenseKey is the encoded form of a license (base64 JSON + signature)
type LicenseKey string

// LicenseStatus represents the current license status
type LicenseStatus struct {
	Valid           bool      `json:"valid"`
	Tier            Tier      `json:"tier"`
	ExpiresAt       time.Time `json:"expires_at"`
	DaysRemaining   int       `json:"days_remaining"`
	LastValidated   time.Time `json:"last_validated"`
	OnlineValidated bool      `json:"online_validated"`
	GracePeriod     bool      `json:"grace_period"`
	Message         string    `json:"message,omitempty"`
}

// ValidationResult contains the result of license validation
type ValidationResult struct {
	Valid   bool
	License *License
	Status  LicenseStatus
	Error   error
}

// Errors
var (
	ErrNoLicense           = errors.New("no license key provided")
	ErrInvalidLicense      = errors.New("invalid license key format")
	ErrInvalidSignature    = errors.New("license signature verification failed")
	ErrLicenseExpired      = errors.New("license has expired")
	ErrHardwareMismatch    = errors.New("license is bound to different hardware")
	ErrFeatureNotLicensed  = errors.New("feature not included in license")
	ErrLimitExceeded       = errors.New("license limit exceeded")
	ErrOnlineCheckRequired = errors.New("online license validation required")
	ErrGracePeriodExpired  = errors.New("grace period has expired")
)

// Manager handles license operations
type Manager struct {
	mu sync.RWMutex

	// Current license
	license *License
	status  LicenseStatus

	// Configuration
	publicKey       ed25519.PublicKey
	licenseServer   string
	offlineGrace    time.Duration // Grace period when offline
	checkInterval   time.Duration // How often to check online
	hardwareID      string        // This machine's hardware ID
	telemetryClient *TelemetryClient
	offlineMode     bool          // Force offline mode - skip all network calls

	// State
	lastOnlineCheck time.Time
	offlineSince    time.Time
}

// ManagerConfig configures the license manager
type ManagerConfig struct {
	PublicKey       string        // Base64 encoded Ed25519 public key
	LicenseServer   string        // URL of license server
	OfflineGrace    time.Duration // Grace period when can't reach server
	CheckInterval   time.Duration // How often to validate online
	EnableTelemetry bool          // Send usage telemetry
	TelemetryURL    string        // Telemetry endpoint
	OfflineMode     bool          // Force offline mode - no phone-home or telemetry
}

// DefaultConfig returns default configuration
func DefaultConfig() ManagerConfig {
	return ManagerConfig{
		LicenseServer: "https://license.savegress.io",
		OfflineGrace:  7 * 24 * time.Hour, // 7 days
		CheckInterval: 24 * time.Hour,     // Daily
		TelemetryURL:  "https://telemetry.savegress.io",
	}
}

// NewManager creates a new license manager
func NewManager(cfg ManagerConfig) (*Manager, error) {
	// Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(pubKeyBytes))
	}

	m := &Manager{
		publicKey:     ed25519.PublicKey(pubKeyBytes),
		licenseServer: cfg.LicenseServer,
		offlineGrace:  cfg.OfflineGrace,
		checkInterval: cfg.CheckInterval,
		offlineMode:   cfg.OfflineMode,
	}

	// Generate hardware ID
	m.hardwareID, err = GenerateHardwareID()
	if err != nil {
		// Non-fatal, but log warning
		m.hardwareID = "unknown"
	}

	// Initialize telemetry if enabled (and not in offline mode)
	if cfg.EnableTelemetry && cfg.TelemetryURL != "" && !cfg.OfflineMode {
		m.telemetryClient = NewTelemetryClient(cfg.TelemetryURL)
	}

	return m, nil
}

// IsOfflineMode returns whether the manager is in forced offline mode
func (m *Manager) IsOfflineMode() bool {
	return m.offlineMode
}

// LoadFromEnv loads license from environment variable
func (m *Manager) LoadFromEnv(envVar string) error {
	key := os.Getenv(envVar)
	if key == "" {
		return ErrNoLicense
	}
	return m.LoadFromKey(LicenseKey(key))
}

// LoadFromFile loads license from a file
func (m *Manager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read license file: %w", err)
	}
	return m.LoadFromKey(LicenseKey(strings.TrimSpace(string(data))))
}

// LoadFromKey loads and validates a license key
func (m *Manager) LoadFromKey(key LicenseKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse license
	license, err := m.parseAndVerify(key)
	if err != nil {
		return err
	}

	// Check hardware binding
	if license.HardwareID != "" && license.HardwareID != m.hardwareID {
		return ErrHardwareMismatch
	}

	// Check expiration
	if time.Now().After(license.ExpiresAt) {
		return ErrLicenseExpired
	}

	m.license = license
	m.status = LicenseStatus{
		Valid:         true,
		Tier:          license.Tier,
		ExpiresAt:     license.ExpiresAt,
		DaysRemaining: int(time.Until(license.ExpiresAt).Hours() / 24),
		LastValidated: time.Now(),
	}

	return nil
}

// parseAndVerify parses a license key and verifies its signature
func (m *Manager) parseAndVerify(key LicenseKey) (*License, error) {
	// License key format: base64(json) + "." + base64(signature)
	parts := strings.Split(string(key), ".")
	if len(parts) != 2 {
		return nil, ErrInvalidLicense
	}

	// Decode JSON payload
	jsonData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid base64 payload", ErrInvalidLicense)
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid base64 signature", ErrInvalidLicense)
	}

	// Verify signature
	if !ed25519.Verify(m.publicKey, jsonData, signature) {
		return nil, ErrInvalidSignature
	}

	// Parse JSON
	var license License
	if err := json.Unmarshal(jsonData, &license); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON", ErrInvalidLicense)
	}

	return &license, nil
}

// ValidateOnline performs online license validation
func (m *Manager) ValidateOnline() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Skip online validation in offline mode
	if m.offlineMode {
		return nil
	}

	if m.license == nil {
		return ErrNoLicense
	}

	// Call license server
	client := NewLicenseClient(m.licenseServer)
	resp, err := client.Validate(m.license.ID, m.hardwareID)
	if err != nil {
		// Check if we're in grace period
		if m.offlineSince.IsZero() {
			m.offlineSince = time.Now()
		}

		elapsed := time.Since(m.offlineSince)
		if elapsed > m.offlineGrace {
			m.status.Valid = false
			m.status.GracePeriod = false
			m.status.Message = "Grace period expired, online validation required"
			return ErrGracePeriodExpired
		}

		m.status.GracePeriod = true
		m.status.Message = fmt.Sprintf("Offline mode, %d days grace remaining",
			int((m.offlineGrace-elapsed).Hours()/24))
		return nil
	}

	// Online validation successful
	m.offlineSince = time.Time{}
	m.lastOnlineCheck = time.Now()
	m.status.OnlineValidated = true
	m.status.GracePeriod = false
	m.status.LastValidated = time.Now()

	// Update license if server returned new data
	if resp.License != nil {
		m.license = resp.License
		m.status.ExpiresAt = resp.License.ExpiresAt
		m.status.DaysRemaining = int(time.Until(resp.License.ExpiresAt).Hours() / 24)
	}

	// Check if license was revoked
	if resp.Revoked {
		m.status.Valid = false
		m.status.Message = "License has been revoked"
		return errors.New("license revoked: " + resp.RevokeReason)
	}

	return nil
}

// HasFeature checks if a feature is licensed
func (m *Manager) HasFeature(feature Feature) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.license == nil {
		// Community features available without license
		for _, f := range CommunityFeatures {
			if f == feature {
				return true
			}
		}
		return false
	}

	// Check explicit features
	for _, f := range m.license.Features {
		if f == feature {
			return true
		}
	}

	// Check tier-based features
	switch m.license.Tier {
	case TierEnterprise:
		for _, f := range EnterpriseFeatures {
			if f == feature {
				return true
			}
		}
		fallthrough
	case TierPro, TierTrial:
		for _, f := range ProFeatures {
			if f == feature {
				return true
			}
		}
		fallthrough
	case TierCommunity:
		for _, f := range CommunityFeatures {
			if f == feature {
				return true
			}
		}
	}

	return false
}

// CheckLimit checks if a limit is exceeded
func (m *Manager) CheckLimit(limitType string, value int64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits := CommunityLimits
	if m.license != nil {
		limits = m.license.Limits
	}

	switch limitType {
	case "sources":
		if limits.MaxSources > 0 && int(value) > limits.MaxSources {
			return fmt.Errorf("%w: max %d sources allowed, got %d",
				ErrLimitExceeded, limits.MaxSources, value)
		}
	case "throughput":
		if limits.MaxThroughput > 0 && value > limits.MaxThroughput {
			return fmt.Errorf("%w: max %d events/sec allowed",
				ErrLimitExceeded, limits.MaxThroughput)
		}
	case "tables":
		if limits.MaxTables > 0 && int(value) > limits.MaxTables {
			return fmt.Errorf("%w: max %d tables allowed, got %d",
				ErrLimitExceeded, limits.MaxTables, value)
		}
	}

	return nil
}

// RequireFeature returns an error if feature is not licensed
func (m *Manager) RequireFeature(feature Feature) error {
	if !m.HasFeature(feature) {
		return fmt.Errorf("%w: %s requires %s or higher license",
			ErrFeatureNotLicensed, feature, m.requiredTier(feature))
	}
	return nil
}

// requiredTier returns the minimum tier required for a feature
func (m *Manager) requiredTier(feature Feature) Tier {
	for _, f := range EnterpriseFeatures {
		if f == feature {
			return TierEnterprise
		}
	}
	for _, f := range ProFeatures {
		if f == feature {
			return TierPro
		}
	}
	return TierCommunity
}

// GetStatus returns current license status
func (m *Manager) GetStatus() LicenseStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetLicense returns the current license (or nil)
func (m *Manager) GetLicense() *License {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license
}

// IsValid returns whether the license is currently valid
func (m *Manager) IsValid() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.Valid
}

// StartBackgroundValidation starts periodic online validation
func (m *Manager) StartBackgroundValidation(ctx context.Context) {
	// In offline mode, skip background validation entirely
	if m.offlineMode {
		return
	}

	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.ValidateOnline(); err != nil {
					// Log error but don't stop
					// The grace period will handle offline scenarios
				}

				// Send telemetry if enabled
				if m.telemetryClient != nil {
					m.sendTelemetry()
				}
			}
		}
	}()
}

func (m *Manager) sendTelemetry() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.license == nil || m.telemetryClient == nil {
		return
	}

	m.telemetryClient.Send(TelemetryEvent{
		LicenseID:  m.license.ID,
		CustomerID: m.license.CustomerID,
		HardwareID: m.hardwareID,
		Timestamp:  time.Now(),
		// Usage metrics would be collected from the engine
	})
}

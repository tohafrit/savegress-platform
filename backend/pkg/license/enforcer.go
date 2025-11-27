package license

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Enforcer performs runtime license enforcement
type Enforcer struct {
	mu sync.RWMutex

	manager   *Manager
	collector *UsageCollector

	// Runtime tracking
	currentSources    int
	currentTables     int
	currentThroughput int64 // events/sec averaged over last minute

	// Throughput calculation
	eventCount    int64
	lastCheck     time.Time
	checkInterval time.Duration

	// Callbacks for enforcement actions
	onLimitExceeded func(limitType string, current, max int64)
	onLicenseExpiry func(daysRemaining int)
}

// EnforcerConfig configures the enforcer
type EnforcerConfig struct {
	Manager             *Manager
	Collector           *UsageCollector
	CheckInterval       time.Duration
	OnLimitExceeded     func(limitType string, current, max int64)
	OnLicenseExpiry     func(daysRemaining int)
}

// NewEnforcer creates a new license enforcer
func NewEnforcer(cfg EnforcerConfig) *Enforcer {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Minute
	}

	return &Enforcer{
		manager:       cfg.Manager,
		collector:     cfg.Collector,
		checkInterval: cfg.CheckInterval,
		lastCheck:     time.Now(),
		onLimitExceeded: cfg.OnLimitExceeded,
		onLicenseExpiry: cfg.OnLicenseExpiry,
	}
}

// StartEnforcement begins background license enforcement
func (e *Enforcer) StartEnforcement(ctx context.Context) {
	go e.enforcementLoop(ctx)
}

func (e *Enforcer) enforcementLoop(ctx context.Context) {
	ticker := time.NewTicker(e.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.performChecks()
		}
	}
}

func (e *Enforcer) performChecks() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Calculate throughput
	elapsed := time.Since(e.lastCheck).Seconds()
	if elapsed > 0 {
		e.currentThroughput = int64(float64(e.eventCount) / elapsed)
	}
	e.eventCount = 0
	e.lastCheck = time.Now()

	// Check limits
	e.checkLimits()

	// Check expiry
	e.checkExpiry()
}

func (e *Enforcer) checkLimits() {
	license := e.manager.GetLicense()
	if license == nil {
		// Use community limits
		e.checkAgainstLimits(CommunityLimits)
		return
	}

	e.checkAgainstLimits(license.Limits)
}

func (e *Enforcer) checkAgainstLimits(limits Limits) {
	// Check sources
	if limits.MaxSources > 0 && e.currentSources > limits.MaxSources {
		if e.onLimitExceeded != nil {
			e.onLimitExceeded("sources", int64(e.currentSources), int64(limits.MaxSources))
		}
	}

	// Check tables
	if limits.MaxTables > 0 && e.currentTables > limits.MaxTables {
		if e.onLimitExceeded != nil {
			e.onLimitExceeded("tables", int64(e.currentTables), int64(limits.MaxTables))
		}
	}

	// Check throughput
	if limits.MaxThroughput > 0 && e.currentThroughput > limits.MaxThroughput {
		if e.onLimitExceeded != nil {
			e.onLimitExceeded("throughput", e.currentThroughput, limits.MaxThroughput)
		}
	}
}

func (e *Enforcer) checkExpiry() {
	status := e.manager.GetStatus()

	// Warn at 30, 14, 7, 3, 1 days
	warnDays := []int{30, 14, 7, 3, 1}

	for _, days := range warnDays {
		if status.DaysRemaining == days {
			if e.onLicenseExpiry != nil {
				e.onLicenseExpiry(days)
			}
			break
		}
	}
}

// RecordEvent records an event for throughput tracking
func (e *Enforcer) RecordEvent() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eventCount++
}

// SetSources updates the current source count
func (e *Enforcer) SetSources(count int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.currentSources = count
}

// SetTables updates the current table count
func (e *Enforcer) SetTables(count int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.currentTables = count
}

// CheckSourceAllowed checks if adding a source is allowed
func (e *Enforcer) CheckSourceAllowed(sourceType string) error {
	// Check feature
	feature := sourceTypeToFeature(sourceType)
	if err := e.manager.RequireFeature(feature); err != nil {
		return err
	}

	// Check source limit
	e.mu.RLock()
	current := e.currentSources
	e.mu.RUnlock()

	return e.manager.CheckLimit("sources", int64(current+1))
}

// CheckTableAllowed checks if adding a table is allowed
func (e *Enforcer) CheckTableAllowed() error {
	e.mu.RLock()
	current := e.currentTables
	e.mu.RUnlock()

	return e.manager.CheckLimit("tables", int64(current+1))
}

// sourceTypeToFeature maps source type to feature
func sourceTypeToFeature(sourceType string) Feature {
	switch sourceType {
	case "postgres", "postgresql":
		return FeaturePostgreSQL
	case "mysql":
		return FeatureMySQL
	case "mariadb":
		return FeatureMariaDB
	case "mongodb":
		return FeatureMongoDB
	case "sqlserver":
		return FeatureSQLServer
	case "oracle":
		return FeatureOracle
	case "cassandra":
		return FeatureCassandra
	case "dynamodb":
		return FeatureDynamoDB
	default:
		return Feature(sourceType)
	}
}

// FeatureGate provides a simple way to gate features
type FeatureGate struct {
	manager *Manager
}

// NewFeatureGate creates a new feature gate
func NewFeatureGate(manager *Manager) *FeatureGate {
	return &FeatureGate{manager: manager}
}

// Require returns an error if feature is not available
func (g *FeatureGate) Require(feature Feature) error {
	return g.manager.RequireFeature(feature)
}

// IsEnabled returns true if feature is available
func (g *FeatureGate) IsEnabled(feature Feature) bool {
	return g.manager.HasFeature(feature)
}

// RequireAny returns an error if none of the features are available
func (g *FeatureGate) RequireAny(features ...Feature) error {
	for _, f := range features {
		if g.manager.HasFeature(f) {
			return nil
		}
	}
	return fmt.Errorf("%w: one of %v required", ErrFeatureNotLicensed, features)
}

// RequireAll returns an error if any of the features are not available
func (g *FeatureGate) RequireAll(features ...Feature) error {
	for _, f := range features {
		if !g.manager.HasFeature(f) {
			return g.manager.RequireFeature(f)
		}
	}
	return nil
}

// GetTier returns the current license tier
func (g *FeatureGate) GetTier() Tier {
	license := g.manager.GetLicense()
	if license == nil {
		return TierCommunity
	}
	return license.Tier
}

// ============================================
// FEATURE-SPECIFIC GATE FUNCTIONS
// ============================================
// These functions provide clear, documented checks for each licensed feature.
// They return descriptive errors that help users understand licensing requirements.

// RequireCompression checks if compression feature is licensed (Pro+)
func (g *FeatureGate) RequireCompression() error {
	if !g.manager.HasFeature(FeatureCompression) {
		return fmt.Errorf("%w: compression requires Pro license or higher", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireCompressionSIMD checks if SIMD compression is licensed (Enterprise)
func (g *FeatureGate) RequireCompressionSIMD() error {
	if !g.manager.HasFeature(FeatureCompressionSIMD) {
		return fmt.Errorf("%w: SIMD compression optimization requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireAdvancedRateLimiting checks if advanced rate limiting is licensed (Pro+)
// Note: Basic token bucket rate limiting is always available
func (g *FeatureGate) RequireAdvancedRateLimiting() error {
	if !g.manager.HasFeature(FeatureAdvancedRateLimiting) {
		return fmt.Errorf("%w: advanced rate limiting (adaptive, sliding window, multi-tier) requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireBackpressure checks if backpressure control is licensed (Pro+)
func (g *FeatureGate) RequireBackpressure() error {
	if !g.manager.HasFeature(FeatureBackpressure) {
		return fmt.Errorf("%w: backpressure control requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireDLQ checks if Dead Letter Queue is licensed (Pro+)
func (g *FeatureGate) RequireDLQ() error {
	if !g.manager.HasFeature(FeatureDLQ) {
		return fmt.Errorf("%w: Dead Letter Queue requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireReplay checks if event replay is licensed (Pro+)
func (g *FeatureGate) RequireReplay() error {
	if !g.manager.HasFeature(FeatureReplay) {
		return fmt.Errorf("%w: event replay requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireSchemaEvolution checks if schema evolution is licensed (Pro+)
func (g *FeatureGate) RequireSchemaEvolution() error {
	if !g.manager.HasFeature(FeatureSchemaEvolution) {
		return fmt.Errorf("%w: schema evolution requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireSchemaMigrationApproval checks if schema migration approval workflow is licensed (Enterprise)
func (g *FeatureGate) RequireSchemaMigrationApproval() error {
	if !g.manager.HasFeature(FeatureSchemaMigrationApproval) {
		return fmt.Errorf("%w: schema migration approval workflow requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequirePrometheus checks if Prometheus metrics export is licensed (Pro+)
func (g *FeatureGate) RequirePrometheus() error {
	if !g.manager.HasFeature(FeaturePrometheus) {
		return fmt.Errorf("%w: Prometheus metrics export requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireSLAMonitoring checks if SLA monitoring is licensed (Pro+)
func (g *FeatureGate) RequireSLAMonitoring() error {
	if !g.manager.HasFeature(FeatureSLAMonitoring) {
		return fmt.Errorf("%w: SLA monitoring requires Pro license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireOpenTelemetry checks if OpenTelemetry tracing is licensed (Enterprise)
func (g *FeatureGate) RequireOpenTelemetry() error {
	if !g.manager.HasFeature(FeatureOpenTelemetry) {
		return fmt.Errorf("%w: OpenTelemetry tracing requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequirePITR checks if Point-in-Time Recovery is licensed (Enterprise)
func (g *FeatureGate) RequirePITR() error {
	if !g.manager.HasFeature(FeaturePITR) {
		return fmt.Errorf("%w: Point-in-Time Recovery requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireCloudStorage checks if cloud storage backends are licensed (Enterprise)
func (g *FeatureGate) RequireCloudStorage() error {
	if !g.manager.HasFeature(FeatureCloudStorage) {
		return fmt.Errorf("%w: cloud storage backends (S3, GCS, Azure) require Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireExactlyOnce checks if exactly-once delivery is licensed (Enterprise)
func (g *FeatureGate) RequireExactlyOnce() error {
	if !g.manager.HasFeature(FeatureExactlyOnce) {
		return fmt.Errorf("%w: exactly-once delivery semantics require Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireMTLS checks if mutual TLS is licensed (Enterprise)
func (g *FeatureGate) RequireMTLS() error {
	if !g.manager.HasFeature(FeatureMTLS) {
		return fmt.Errorf("%w: mutual TLS authentication requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireRBAC checks if RBAC is licensed (Enterprise)
func (g *FeatureGate) RequireRBAC() error {
	if !g.manager.HasFeature(FeatureRBAC) {
		return fmt.Errorf("%w: role-based access control requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireVault checks if HashiCorp Vault integration is licensed (Enterprise)
func (g *FeatureGate) RequireVault() error {
	if !g.manager.HasFeature(FeatureVault) {
		return fmt.Errorf("%w: HashiCorp Vault integration requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireHA checks if high availability is licensed (Enterprise)
func (g *FeatureGate) RequireHA() error {
	if !g.manager.HasFeature(FeatureHA) {
		return fmt.Errorf("%w: high availability mode requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireRaftCluster checks if Raft clustering is licensed (Enterprise)
func (g *FeatureGate) RequireRaftCluster() error {
	if !g.manager.HasFeature(FeatureRaftCluster) {
		return fmt.Errorf("%w: Raft consensus clustering requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// RequireMultiRegion checks if multi-region deployment is licensed (Enterprise)
func (g *FeatureGate) RequireMultiRegion() error {
	if !g.manager.HasFeature(FeatureMultiRegion) {
		return fmt.Errorf("%w: multi-region deployment requires Enterprise license", ErrFeatureNotLicensed)
	}
	return nil
}

// ============================================
// UPGRADE PROMPTS
// ============================================
// These functions return user-friendly upgrade messages for sales conversion.

// GetUpgradePrompt returns a user-friendly upgrade message for a feature
func GetUpgradePrompt(feature Feature) string {
	prompts := map[Feature]string{
		// Pro features
		FeatureCompression:         "Upgrade to Pro to enable compression and save 4-10x on storage costs. Visit https://savegress.io/pricing",
		FeatureDLQ:                 "Upgrade to Pro to enable Dead Letter Queue and prevent data loss. Visit https://savegress.io/pricing",
		FeatureSchemaEvolution:     "Upgrade to Pro to enable automatic schema evolution. Visit https://savegress.io/pricing",
		FeaturePrometheus:          "Upgrade to Pro to export Prometheus metrics for your monitoring stack. Visit https://savegress.io/pricing",
		FeatureAdvancedRateLimiting: "Upgrade to Pro for adaptive rate limiting and better flow control. Visit https://savegress.io/pricing",
		FeatureBackpressure:        "Upgrade to Pro for backpressure control at high throughput. Visit https://savegress.io/pricing",
		FeatureReplay:              "Upgrade to Pro to replay events for debugging and recovery. Visit https://savegress.io/pricing",
		FeatureSLAMonitoring:       "Upgrade to Pro for SLA monitoring and alerting. Visit https://savegress.io/pricing",

		// Enterprise features
		FeaturePITR:                    "Upgrade to Enterprise for Point-in-Time Recovery. Visit https://savegress.io/pricing",
		FeatureCloudStorage:            "Upgrade to Enterprise to use S3, GCS, or Azure storage backends. Visit https://savegress.io/pricing",
		FeatureOpenTelemetry:           "Upgrade to Enterprise for full OpenTelemetry distributed tracing. Visit https://savegress.io/pricing",
		FeatureCompressionSIMD:         "Upgrade to Enterprise for SIMD-optimized compression. Visit https://savegress.io/pricing",
		FeatureExactlyOnce:             "Upgrade to Enterprise for exactly-once delivery semantics. Visit https://savegress.io/pricing",
		FeatureSchemaMigrationApproval: "Upgrade to Enterprise for schema migration approval workflows. Visit https://savegress.io/pricing",
		FeatureMTLS:                    "Upgrade to Enterprise for mutual TLS authentication. Visit https://savegress.io/pricing",
		FeatureRBAC:                    "Upgrade to Enterprise for role-based access control. Visit https://savegress.io/pricing",
		FeatureVault:                   "Upgrade to Enterprise for HashiCorp Vault integration. Visit https://savegress.io/pricing",
		FeatureHA:                      "Upgrade to Enterprise for high availability mode. Visit https://savegress.io/pricing",
		FeatureRaftCluster:             "Upgrade to Enterprise for Raft consensus clustering. Visit https://savegress.io/pricing",
		FeatureMultiRegion:             "Upgrade to Enterprise for multi-region deployment. Visit https://savegress.io/pricing",
		FeatureOracle:                  "Upgrade to Enterprise to use Oracle as a source. Visit https://savegress.io/pricing",
	}

	if prompt, ok := prompts[feature]; ok {
		return prompt
	}
	return "Upgrade your license to access this feature. Visit https://savegress.io/pricing"
}

// GetLimitExceededPrompt returns a user-friendly message when limits are exceeded
func GetLimitExceededPrompt(limitType string, current, max int64) string {
	switch limitType {
	case "sources":
		return fmt.Sprintf(
			"Source limit reached (%d/%d). Upgrade to Pro for up to 10 sources, or Enterprise for unlimited. Visit https://savegress.io/pricing",
			current, max)
	case "tables":
		return fmt.Sprintf(
			"Table limit reached (%d/%d). Upgrade to Pro for up to 100 tables, or Enterprise for unlimited. Visit https://savegress.io/pricing",
			current, max)
	case "throughput":
		return fmt.Sprintf(
			"Throughput limit reached (%d/%d events/sec). Upgrade to Pro for 50K events/sec, or Enterprise for unlimited. Visit https://savegress.io/pricing",
			current, max)
	default:
		return fmt.Sprintf(
			"License limit reached (%d/%d). Upgrade your license for higher limits. Visit https://savegress.io/pricing",
			current, max)
	}
}

// GetExpiryWarning returns a user-friendly expiry warning message
func GetExpiryWarning(daysRemaining int) string {
	if daysRemaining <= 0 {
		return "Your license has expired. Please renew at https://savegress.io/account to continue using premium features."
	}
	if daysRemaining == 1 {
		return "Your license expires tomorrow. Renew at https://savegress.io/account to avoid service interruption."
	}
	if daysRemaining <= 7 {
		return fmt.Sprintf("Your license expires in %d days. Renew at https://savegress.io/account", daysRemaining)
	}
	return fmt.Sprintf("License expires in %d days. Renew at https://savegress.io/account", daysRemaining)
}

// IsCommunity returns true if on community tier
func (g *FeatureGate) IsCommunity() bool {
	return g.GetTier() == TierCommunity
}

// IsPro returns true if on pro tier or higher
func (g *FeatureGate) IsPro() bool {
	tier := g.GetTier()
	return tier == TierPro || tier == TierEnterprise || tier == TierTrial
}

// IsEnterprise returns true if on enterprise tier
func (g *FeatureGate) IsEnterprise() bool {
	return g.GetTier() == TierEnterprise
}

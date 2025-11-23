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

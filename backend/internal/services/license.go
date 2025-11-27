package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
	"github.com/savegress/platform/backend/pkg/license"
)

var (
	ErrLicenseNotFound        = errors.New("license not found")
	ErrLicenseExpired         = errors.New("license has expired")
	ErrLicenseRevoked         = errors.New("license has been revoked")
	ErrInvalidLicense         = errors.New("invalid license format")
	ErrInvalidSignature       = errors.New("license signature verification failed")
	ErrHardwareMismatch       = errors.New("license is bound to different hardware")
	ErrActivationLimitReached = errors.New("activation limit reached")
)

// LicenseService handles license operations
type LicenseService struct {
	db        *repository.PostgresDB
	generator *license.LicenseGenerator
	issuer    string
}

// NewLicenseService creates a new license service
func NewLicenseService(db *repository.PostgresDB, privateKeyBase64 string) *LicenseService {
	svc := &LicenseService{
		db:     db,
		issuer: "license.savegress.io",
	}

	// Load private key and create generator
	if privateKeyBase64 != "" {
		generator, err := license.NewLicenseGeneratorFromBase64(privateKeyBase64)
		if err == nil {
			svc.generator = generator
		}
	}

	return svc
}

// CreateLicense creates a new license for a user
func (s *LicenseService) CreateLicense(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
	// Get user
	var userName, company string
	err := s.db.Pool().QueryRow(ctx, "SELECT name, COALESCE(company, '') FROM users WHERE id = $1", userID).Scan(&userName, &company)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Determine limits based on tier
	lim := s.getLimitsForTier(tier)
	features := s.getFeaturesForTier(tier)

	// Create license record
	lic := &models.License{
		ID:            uuid.New(),
		UserID:        userID,
		Tier:          tier,
		Status:        "active",
		MaxSources:    lim.MaxSources,
		MaxTables:     lim.MaxTables,
		MaxThroughput: lim.MaxThroughput,
		Features:      features,
		HardwareID:    hardwareID,
		IssuedAt:      time.Now().UTC(),
		ExpiresAt:     time.Now().UTC().AddDate(0, 0, validDays),
		CreatedAt:     time.Now().UTC(),
	}

	// Generate signed license key using shared library
	licenseKey, err := s.generateLicenseKey(lic, userID.String(), fmt.Sprintf("%s (%s)", userName, company))
	if err != nil {
		return nil, fmt.Errorf("failed to generate license key: %w", err)
	}
	lic.LicenseKey = licenseKey

	// Store in database
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO licenses (id, user_id, license_key, tier, status, max_sources, max_tables, max_throughput, features, hardware_id, issued_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, lic.ID, lic.UserID, lic.LicenseKey, lic.Tier, lic.Status,
		lic.MaxSources, lic.MaxTables, lic.MaxThroughput, lic.Features,
		lic.HardwareID, lic.IssuedAt, lic.ExpiresAt, lic.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store license: %w", err)
	}

	return lic, nil
}

// ValidateLicense validates a license (called by CDC engines)
func (s *LicenseService) ValidateLicense(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
	var license models.License
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, tier, status, max_sources, max_tables, max_throughput, features, hardware_id, issued_at, expires_at, revoked_at
		FROM licenses WHERE id = $1
	`, licenseID).Scan(&license.ID, &license.UserID, &license.Tier, &license.Status,
		&license.MaxSources, &license.MaxTables, &license.MaxThroughput, &license.Features,
		&license.HardwareID, &license.IssuedAt, &license.ExpiresAt, &license.RevokedAt)
	if err != nil {
		return nil, ErrLicenseNotFound
	}

	// Check status
	if license.Status == "revoked" || license.RevokedAt != nil {
		return nil, ErrLicenseRevoked
	}

	// Check expiration
	if time.Now().After(license.ExpiresAt) {
		return nil, ErrLicenseExpired
	}

	// Check hardware binding
	if license.HardwareID != "" && license.HardwareID != hardwareID {
		return nil, ErrHardwareMismatch
	}

	return &license, nil
}

// ActivateLicense records a license activation
func (s *LicenseService) ActivateLicense(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error) {
	// Validate license first
	license, err := s.ValidateLicense(ctx, licenseID.String(), hardwareID)
	if err != nil {
		return nil, err
	}

	// Check if already activated on this hardware
	var existingID uuid.UUID
	err = s.db.Pool().QueryRow(ctx, `
		SELECT id FROM license_activations
		WHERE license_id = $1 AND hardware_id = $2 AND deactivated_at IS NULL
	`, licenseID, hardwareID).Scan(&existingID)

	if err == nil {
		// Update existing activation
		now := time.Now().UTC()
		_, err = s.db.Pool().Exec(ctx, `
			UPDATE license_activations
			SET last_seen_at = $1, hostname = $2, platform = $3, version = $4, ip_address = $5
			WHERE id = $6
		`, now, hostname, platform, version, ipAddress, existingID)
		if err != nil {
			return nil, fmt.Errorf("failed to update activation: %w", err)
		}
		return &models.LicenseActivation{ID: existingID, LicenseID: licenseID, HardwareID: hardwareID, LastSeenAt: now}, nil
	}

	// Check activation limit (enterprise = unlimited, pro = 5, trial = 1)
	var activationCount int
	_ = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM license_activations WHERE license_id = $1 AND deactivated_at IS NULL
	`, licenseID).Scan(&activationCount)

	maxActivations := s.getMaxActivations(license.Tier)
	if maxActivations > 0 && activationCount >= maxActivations {
		return nil, ErrActivationLimitReached
	}

	// Create new activation
	activation := &models.LicenseActivation{
		ID:          uuid.New(),
		LicenseID:   licenseID,
		HardwareID:  hardwareID,
		Hostname:    hostname,
		Platform:    platform,
		Version:     version,
		IPAddress:   ipAddress,
		ActivatedAt: time.Now().UTC(),
		LastSeenAt:  time.Now().UTC(),
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO license_activations (id, license_id, hardware_id, hostname, platform, version, ip_address, activated_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, activation.ID, activation.LicenseID, activation.HardwareID, activation.Hostname,
		activation.Platform, activation.Version, activation.IPAddress, activation.ActivatedAt, activation.LastSeenAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create activation: %w", err)
	}

	return activation, nil
}

// DeactivateLicense removes an activation
func (s *LicenseService) DeactivateLicense(ctx context.Context, licenseID uuid.UUID, hardwareID string) error {
	now := time.Now().UTC()
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE license_activations SET deactivated_at = $1
		WHERE license_id = $2 AND hardware_id = $3 AND deactivated_at IS NULL
	`, now, licenseID, hardwareID)
	return err
}

// RevokeLicense revokes a license
func (s *LicenseService) RevokeLicense(ctx context.Context, licenseID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE licenses SET status = 'revoked', revoked_at = $1 WHERE id = $2
	`, now, licenseID)
	return err
}

// GetUserLicenses returns all licenses for a user
func (s *LicenseService) GetUserLicenses(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, license_key, tier, status, max_sources, max_tables, max_throughput, features, hardware_id, issued_at, expires_at, revoked_at, created_at
		FROM licenses WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	licenses := make([]models.License, 0)
	for rows.Next() {
		var l models.License
		err := rows.Scan(&l.ID, &l.UserID, &l.LicenseKey, &l.Tier, &l.Status,
			&l.MaxSources, &l.MaxTables, &l.MaxThroughput, &l.Features,
			&l.HardwareID, &l.IssuedAt, &l.ExpiresAt, &l.RevokedAt, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		licenses = append(licenses, l)
	}
	return licenses, nil
}

// ListAllLicenses returns all licenses with pagination (admin only)
func (s *LicenseService) ListAllLicenses(ctx context.Context, limit, offset int) ([]models.License, int, error) {
	// Get total count
	var total int
	err := s.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM licenses`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated licenses
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, license_key, tier, status, max_sources, max_tables, max_throughput, features, hardware_id, issued_at, expires_at, revoked_at, created_at
		FROM licenses ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	licenses := make([]models.License, 0)
	for rows.Next() {
		var l models.License
		err := rows.Scan(&l.ID, &l.UserID, &l.LicenseKey, &l.Tier, &l.Status,
			&l.MaxSources, &l.MaxTables, &l.MaxThroughput, &l.Features,
			&l.HardwareID, &l.IssuedAt, &l.ExpiresAt, &l.RevokedAt, &l.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		licenses = append(licenses, l)
	}
	return licenses, total, nil
}

// GetLicenseActivations returns all activations for a license
func (s *LicenseService) GetLicenseActivations(ctx context.Context, licenseID uuid.UUID) ([]models.LicenseActivation, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, license_id, hardware_id, hostname, platform, version, ip_address, activated_at, last_seen_at, deactivated_at
		FROM license_activations WHERE license_id = $1 ORDER BY activated_at DESC
	`, licenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activations := make([]models.LicenseActivation, 0)
	for rows.Next() {
		var a models.LicenseActivation
		err := rows.Scan(&a.ID, &a.LicenseID, &a.HardwareID, &a.Hostname, &a.Platform,
			&a.Version, &a.IPAddress, &a.ActivatedAt, &a.LastSeenAt, &a.DeactivatedAt)
		if err != nil {
			return nil, err
		}
		activations = append(activations, a)
	}
	return activations, nil
}

// generateLicenseKey creates a signed license key using the shared license package
func (s *LicenseService) generateLicenseKey(lic *models.License, customerID, customerName string) (string, error) {
	if s.generator == nil {
		return "", errors.New("license generator not configured")
	}

	// Map tier string to license.Tier
	tier := license.TierCommunity
	switch lic.Tier {
	case "enterprise":
		tier = license.TierEnterprise
	case "pro":
		tier = license.TierPro
	case "trial":
		tier = license.TierTrial
	}

	req := license.GenerateRequest{
		CustomerID:   customerID,
		CustomerName: customerName,
		Tier:         tier,
		ValidDays:    int(lic.ExpiresAt.Sub(lic.IssuedAt).Hours() / 24),
		HardwareID:   lic.HardwareID,
		Metadata: map[string]string{
			"license_id": lic.ID.String(),
		},
	}

	key, err := s.generator.Generate(req)
	if err != nil {
		return "", err
	}

	return string(key), nil
}

type limits struct {
	MaxSources    int
	MaxTables     int
	MaxThroughput int64
}

func (s *LicenseService) getLimitsForTier(tier string) limits {
	switch tier {
	case "enterprise":
		return limits{MaxSources: 0, MaxTables: 0, MaxThroughput: 0} // Unlimited
	case "pro":
		return limits{MaxSources: 10, MaxTables: 100, MaxThroughput: 50000}
	case "trial":
		return limits{MaxSources: 5, MaxTables: 50, MaxThroughput: 10000}
	default: // community
		return limits{MaxSources: 1, MaxTables: 10, MaxThroughput: 1000}
	}
}

func (s *LicenseService) getFeaturesForTier(tier string) []string {
	// Community features - core databases only
	community := []string{
		"postgresql", "mysql", "mariadb",
	}

	// Pro features - adds scale, performance, and DevOps tooling
	pro := append(community,
		// Additional databases
		"mongodb", "sqlserver", "cassandra", "dynamodb",
		// Output connectors
		"snapshot", "kafka_output", "grpc_output", "webhook",
		// Performance
		"compression",
		// Reliability
		"advanced_rate_limiting", "backpressure", "dlq", "replay",
		// Schema management
		"schema_evolution",
		// Observability
		"prometheus", "sla_monitoring",
	)

	// Enterprise features - adds compliance, security, and HA
	enterprise := append(pro,
		// Premium database
		"oracle",
		// Custom integrations
		"custom_output",
		// Maximum performance
		"compression_simd", "exactly_once",
		// Disaster recovery
		"pitr", "cloud_storage",
		// Advanced schema governance
		"schema_migration_approval",
		// Full observability
		"opentelemetry",
		// High availability & clustering
		"ha", "raft_cluster", "multi_region",
		// Security & compliance
		"encryption", "mtls", "rbac", "vault",
		"audit_log", "sso", "ldap", "multi_tenant",
	)

	switch tier {
	case "enterprise":
		return enterprise
	case "pro", "trial":
		return pro
	default:
		return community
	}
}

func (s *LicenseService) getMaxActivations(tier string) int {
	switch tier {
	case "enterprise":
		return 0 // Unlimited
	case "pro":
		return 10
	case "trial":
		return 2
	default:
		return 1
	}
}

// GetAllLicensesPaginated returns all licenses with pagination and optional filters
func (s *LicenseService) GetAllLicensesPaginated(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
	offset := (page - 1) * limit

	// Build query with optional filters
	query := `
		SELECT id, user_id, license_key, tier, status, max_sources, max_tables, max_throughput,
			   features, hardware_id, issued_at, expires_at, revoked_at, created_at
		FROM licenses WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM licenses WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if tier != "" {
		argCount++
		query += fmt.Sprintf(" AND tier = $%d", argCount)
		countQuery += fmt.Sprintf(" AND tier = $%d", argCount)
		args = append(args, tier)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	// Get total count
	var total int
	err := s.db.Pool().QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count licenses: %w", err)
	}

	// Add pagination
	argCount++
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argCount)
	args = append(args, limit)
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	// Execute query
	rows, err := s.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query licenses: %w", err)
	}
	defer rows.Close()

	licenses := make([]models.License, 0)
	for rows.Next() {
		var l models.License
		err := rows.Scan(&l.ID, &l.UserID, &l.LicenseKey, &l.Tier, &l.Status,
			&l.MaxSources, &l.MaxTables, &l.MaxThroughput, &l.Features,
			&l.HardwareID, &l.IssuedAt, &l.ExpiresAt, &l.RevokedAt, &l.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan license: %w", err)
		}
		licenses = append(licenses, l)
	}

	return licenses, total, nil
}

// CreateLicenseForSubscription creates a license for a new subscription (used by billing webhook)
func (s *LicenseService) CreateLicenseForSubscription(ctx context.Context, userID uuid.UUID, plan string) (*models.License, error) {
	// Map plan to tier
	tier := plan
	if tier == "" {
		tier = "pro"
	}

	// Default validity: 365 days for subscriptions (renewed on payment)
	validDays := 365

	// No hardware binding for subscription licenses (bound on first activation)
	return s.CreateLicense(ctx, userID, tier, validDays, "")
}

// UpdateLicenseTier updates the tier for a user's active license
func (s *LicenseService) UpdateLicenseTier(ctx context.Context, userID uuid.UUID, newTier string) error {
	// Get limits and features for new tier
	lim := s.getLimitsForTier(newTier)
	features := s.getFeaturesForTier(newTier)

	_, err := s.db.Pool().Exec(ctx, `
		UPDATE licenses
		SET tier = $1, max_sources = $2, max_tables = $3, max_throughput = $4, features = $5, updated_at = NOW()
		WHERE user_id = $6 AND status = 'active'
	`, newTier, lim.MaxSources, lim.MaxTables, lim.MaxThroughput, features, userID)

	return err
}

// ExtendLicense extends a user's active license by the specified number of days
func (s *LicenseService) ExtendLicense(ctx context.Context, userID uuid.UUID, days int) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE licenses
		SET expires_at = expires_at + INTERVAL '1 day' * $1, updated_at = NOW()
		WHERE user_id = $2 AND status = 'active'
	`, days, userID)

	return err
}

// RevokeUserLicenses revokes all active licenses for a user (used when subscription is canceled)
func (s *LicenseService) RevokeUserLicenses(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE licenses SET status = 'revoked', revoked_at = $1
		WHERE user_id = $2 AND status = 'active'
	`, now, userID)
	return err
}

// ============================================
// ANALYTICS & STATISTICS
// ============================================

// LicenseStats contains license statistics for dashboards
type LicenseStats struct {
	TotalLicenses     int            `json:"total_licenses"`
	ActiveLicenses    int            `json:"active_licenses"`
	ExpiredLicenses   int            `json:"expired_licenses"`
	RevokedLicenses   int            `json:"revoked_licenses"`
	ExpiringIn30Days  int            `json:"expiring_in_30_days"`
	LicensesByTier    map[string]int `json:"licenses_by_tier"`
	ActiveActivations int            `json:"active_activations"`
	RevenueMetrics    RevenueMetrics `json:"revenue_metrics"`
}

// RevenueMetrics contains revenue-related statistics
type RevenueMetrics struct {
	ProLicenses        int `json:"pro_licenses"`
	EnterpriseLicenses int `json:"enterprise_licenses"`
	TrialLicenses      int `json:"trial_licenses"`
	ConversionRate     float64 `json:"conversion_rate"` // Trial -> Paid
}

// GetLicenseStats returns comprehensive license statistics (admin only)
func (s *LicenseService) GetLicenseStats(ctx context.Context) (*LicenseStats, error) {
	stats := &LicenseStats{
		LicensesByTier: make(map[string]int),
	}

	// Total licenses
	err := s.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM licenses`).Scan(&stats.TotalLicenses)
	if err != nil {
		return nil, fmt.Errorf("failed to count total licenses: %w", err)
	}

	// Active licenses
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM licenses
		WHERE status = 'active' AND expires_at > NOW()
	`).Scan(&stats.ActiveLicenses)
	if err != nil {
		return nil, fmt.Errorf("failed to count active licenses: %w", err)
	}

	// Expired licenses
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM licenses
		WHERE expires_at <= NOW() AND status != 'revoked'
	`).Scan(&stats.ExpiredLicenses)
	if err != nil {
		return nil, fmt.Errorf("failed to count expired licenses: %w", err)
	}

	// Revoked licenses
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM licenses WHERE status = 'revoked'
	`).Scan(&stats.RevokedLicenses)
	if err != nil {
		return nil, fmt.Errorf("failed to count revoked licenses: %w", err)
	}

	// Expiring in 30 days
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM licenses
		WHERE status = 'active'
		AND expires_at > NOW()
		AND expires_at <= NOW() + INTERVAL '30 days'
	`).Scan(&stats.ExpiringIn30Days)
	if err != nil {
		return nil, fmt.Errorf("failed to count expiring licenses: %w", err)
	}

	// Licenses by tier
	rows, err := s.db.Pool().Query(ctx, `
		SELECT tier, COUNT(*) FROM licenses
		WHERE status = 'active' AND expires_at > NOW()
		GROUP BY tier
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get licenses by tier: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tier string
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			return nil, err
		}
		stats.LicensesByTier[tier] = count
	}

	// Active activations
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM license_activations WHERE deactivated_at IS NULL
	`).Scan(&stats.ActiveActivations)
	if err != nil {
		return nil, fmt.Errorf("failed to count active activations: %w", err)
	}

	// Revenue metrics
	stats.RevenueMetrics.ProLicenses = stats.LicensesByTier["pro"]
	stats.RevenueMetrics.EnterpriseLicenses = stats.LicensesByTier["enterprise"]
	stats.RevenueMetrics.TrialLicenses = stats.LicensesByTier["trial"]

	// Conversion rate (trials that converted to pro/enterprise)
	var totalTrials, convertedTrials int
	err = s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM licenses WHERE tier = 'trial'
	`).Scan(&totalTrials)
	if err == nil && totalTrials > 0 {
		// Count users who had trial and now have pro/enterprise
		err = s.db.Pool().QueryRow(ctx, `
			SELECT COUNT(DISTINCT l1.user_id) FROM licenses l1
			WHERE l1.tier = 'trial'
			AND EXISTS (
				SELECT 1 FROM licenses l2
				WHERE l2.user_id = l1.user_id
				AND l2.tier IN ('pro', 'enterprise')
				AND l2.issued_at > l1.issued_at
			)
		`).Scan(&convertedTrials)
		if err == nil {
			stats.RevenueMetrics.ConversionRate = float64(convertedTrials) / float64(totalTrials) * 100
		}
	}

	return stats, nil
}

// UsageRecord represents usage data from an engine
type UsageRecord struct {
	LicenseID     uuid.UUID `json:"license_id"`
	HardwareID    string    `json:"hardware_id"`
	EventsTotal   int64     `json:"events_total"`
	BytesTotal    int64     `json:"bytes_total"`
	ErrorCount    int64     `json:"error_count"`
	AvgLatencyMs  float64   `json:"avg_latency_ms"`
	SourcesActive int       `json:"sources_active"`
	TablesTracked int       `json:"tables_tracked"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// RecordUsage stores usage telemetry from CDC engines
func (s *LicenseService) RecordUsage(ctx context.Context, record UsageRecord) error {
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO license_usage (
			license_id, hardware_id, events_total, bytes_total,
			error_count, avg_latency_ms, sources_active, tables_tracked, recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, record.LicenseID, record.HardwareID, record.EventsTotal, record.BytesTotal,
		record.ErrorCount, record.AvgLatencyMs, record.SourcesActive, record.TablesTracked, record.RecordedAt)
	return err
}

// GetUsageStats returns usage statistics for a license
func (s *LicenseService) GetUsageStats(ctx context.Context, licenseID uuid.UUID, days int) ([]UsageRecord, error) {
	if days == 0 {
		days = 30
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT license_id, hardware_id, events_total, bytes_total,
			   error_count, avg_latency_ms, sources_active, tables_tracked, recorded_at
		FROM license_usage
		WHERE license_id = $1 AND recorded_at > NOW() - INTERVAL '1 day' * $2
		ORDER BY recorded_at DESC
	`, licenseID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]UsageRecord, 0)
	for rows.Next() {
		var r UsageRecord
		err := rows.Scan(&r.LicenseID, &r.HardwareID, &r.EventsTotal, &r.BytesTotal,
			&r.ErrorCount, &r.AvgLatencyMs, &r.SourcesActive, &r.TablesTracked, &r.RecordedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// GetAggregatedUsage returns aggregated usage for billing/reporting
func (s *LicenseService) GetAggregatedUsage(ctx context.Context, licenseID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var totalEvents, totalBytes, totalErrors int64
	var avgLatency float64
	var maxSources, maxTables int

	err := s.db.Pool().QueryRow(ctx, `
		SELECT
			COALESCE(SUM(events_total), 0),
			COALESCE(SUM(bytes_total), 0),
			COALESCE(SUM(error_count), 0),
			COALESCE(AVG(avg_latency_ms), 0),
			COALESCE(MAX(sources_active), 0),
			COALESCE(MAX(tables_tracked), 0)
		FROM license_usage
		WHERE license_id = $1 AND recorded_at BETWEEN $2 AND $3
	`, licenseID, startDate, endDate).Scan(&totalEvents, &totalBytes, &totalErrors, &avgLatency, &maxSources, &maxTables)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_events":      totalEvents,
		"total_bytes":       totalBytes,
		"total_errors":      totalErrors,
		"avg_latency_ms":    avgLatency,
		"max_sources_used":  maxSources,
		"max_tables_tracked": maxTables,
		"period_start":      startDate,
		"period_end":        endDate,
	}, nil
}

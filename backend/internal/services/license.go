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
	community := []string{"postgresql", "mysql", "mariadb"}
	pro := append(community, "mongodb", "sqlserver", "cassandra", "dynamodb", "snapshot", "kafka_output", "grpc_output")
	enterprise := append(pro, "oracle", "ha", "raft_cluster", "sso", "ldap", "audit_log")

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

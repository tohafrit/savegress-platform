package license

import (
	"testing"
	"time"
)

func TestGenerateAndVerifyLicense(t *testing.T) {
	// Generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	// Create generator
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")

	// Generate a license
	licenseKey, err := gen.Generate(GenerateRequest{
		CustomerID:   "cust-123",
		CustomerName: "Test Company",
		Tier:         TierPro,
		ValidDays:    365,
	})
	if err != nil {
		t.Fatalf("failed to generate license: %v", err)
	}

	t.Logf("Generated license key: %s", licenseKey)

	// Verify the license
	license, err := VerifyLicense(licenseKey, keyPair.PublicKey)
	if err != nil {
		t.Fatalf("failed to verify license: %v", err)
	}

	// Check fields
	if license.CustomerID != "cust-123" {
		t.Errorf("expected customer ID 'cust-123', got '%s'", license.CustomerID)
	}
	if license.CustomerName != "Test Company" {
		t.Errorf("expected customer name 'Test Company', got '%s'", license.CustomerName)
	}
	if license.Tier != TierPro {
		t.Errorf("expected tier 'pro', got '%s'", license.Tier)
	}
	if license.Issuer != "test.savegress.io" {
		t.Errorf("expected issuer 'test.savegress.io', got '%s'", license.Issuer)
	}

	// Check expiration
	expectedExpiry := time.Now().AddDate(0, 0, 365)
	if license.ExpiresAt.Before(expectedExpiry.Add(-time.Hour)) || license.ExpiresAt.After(expectedExpiry.Add(time.Hour)) {
		t.Errorf("expiry date not within expected range")
	}
}

func TestLicenseManager(t *testing.T) {
	// Generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	// Create manager
	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Generate a license
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")
	licenseKey, err := gen.Generate(GenerateRequest{
		CustomerID:   "cust-456",
		CustomerName: "Another Company",
		Tier:         TierEnterprise,
		ValidDays:    30,
	})
	if err != nil {
		t.Fatalf("failed to generate license: %v", err)
	}

	// Load license
	err = manager.LoadFromKey(licenseKey)
	if err != nil {
		t.Fatalf("failed to load license: %v", err)
	}

	// Check status
	status := manager.GetStatus()
	if !status.Valid {
		t.Error("expected license to be valid")
	}
	if status.Tier != TierEnterprise {
		t.Errorf("expected tier 'enterprise', got '%s'", status.Tier)
	}
}

func TestFeatureChecking(t *testing.T) {
	// Generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Without license, only community features available
	if !manager.HasFeature(FeaturePostgreSQL) {
		t.Error("PostgreSQL should be available in community")
	}
	if !manager.HasFeature(FeatureMySQL) {
		t.Error("MySQL should be available in community")
	}
	if manager.HasFeature(FeatureOracle) {
		t.Error("Oracle should NOT be available in community")
	}
	if manager.HasFeature(FeatureSQLServer) {
		t.Error("SQL Server should NOT be available in community")
	}

	// Load pro license
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")
	licenseKey, _ := gen.Generate(GenerateRequest{
		CustomerID:   "test",
		CustomerName: "Test",
		Tier:         TierPro,
		ValidDays:    30,
	})
	manager.LoadFromKey(licenseKey)

	// Pro features should be available
	if !manager.HasFeature(FeatureMongoDB) {
		t.Error("MongoDB should be available in pro")
	}
	if !manager.HasFeature(FeatureSQLServer) {
		t.Error("SQL Server should be available in pro")
	}
	// Oracle is enterprise only
	if manager.HasFeature(FeatureOracle) {
		t.Error("Oracle should NOT be available in pro")
	}
}

func TestLimitChecking(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Community limits
	err = manager.CheckLimit("sources", 1)
	if err != nil {
		t.Error("1 source should be allowed in community")
	}

	err = manager.CheckLimit("sources", 2)
	if err == nil {
		t.Error("2 sources should NOT be allowed in community")
	}

	// Load pro license with custom limits
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")
	licenseKey, _ := gen.Generate(GenerateRequest{
		CustomerID:   "test",
		CustomerName: "Test",
		Tier:         TierPro,
		Limits: &Limits{
			MaxSources:    5,
			MaxThroughput: 10000,
		},
		ValidDays: 30,
	})
	manager.LoadFromKey(licenseKey)

	err = manager.CheckLimit("sources", 5)
	if err != nil {
		t.Error("5 sources should be allowed")
	}

	err = manager.CheckLimit("sources", 6)
	if err == nil {
		t.Error("6 sources should NOT be allowed")
	}
}

func TestExpiredLicense(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Generate expired license (negative days)
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")
	licenseKey, _ := gen.Generate(GenerateRequest{
		CustomerID:   "test",
		CustomerName: "Test",
		Tier:         TierPro,
		ValidDays:    -1, // Expired yesterday
	})

	err = manager.LoadFromKey(licenseKey)
	if err != ErrLicenseExpired {
		t.Errorf("expected ErrLicenseExpired, got %v", err)
	}
}

func TestInvalidSignature(t *testing.T) {
	// Generate two different key pairs
	keyPair1, _ := GenerateKeyPair()
	keyPair2, _ := GenerateKeyPair()

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair2.PublicKeyBase64() // Use different key

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Generate license with first key
	gen := NewLicenseGenerator(keyPair1, "test.savegress.io")
	licenseKey, _ := gen.Generate(GenerateRequest{
		CustomerID:   "test",
		CustomerName: "Test",
		Tier:         TierPro,
		ValidDays:    30,
	})

	// Try to load with manager using second key
	err = manager.LoadFromKey(licenseKey)
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestHardwareBinding(t *testing.T) {
	keyPair, _ := GenerateKeyPair()

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Generate license bound to different hardware
	gen := NewLicenseGenerator(keyPair, "test.savegress.io")
	licenseKey, _ := gen.Generate(GenerateRequest{
		CustomerID:   "test",
		CustomerName: "Test",
		Tier:         TierPro,
		ValidDays:    30,
		HardwareID:   "different-hardware-id",
	})

	err = manager.LoadFromKey(licenseKey)
	if err != ErrHardwareMismatch {
		t.Errorf("expected ErrHardwareMismatch, got %v", err)
	}
}

func TestHardwareID(t *testing.T) {
	hwID, err := GenerateHardwareID()
	if err != nil {
		t.Logf("Hardware ID generation failed (may be expected in some environments): %v", err)
		return
	}

	if hwID == "" {
		t.Error("hardware ID should not be empty")
	}

	t.Logf("Generated hardware ID: %s", hwID)

	// Generate again - should be stable
	hwID2, _ := GenerateHardwareID()
	if hwID != hwID2 {
		t.Error("hardware ID should be stable")
	}
}

func TestFeatureGate(t *testing.T) {
	keyPair, _ := GenerateKeyPair()

	cfg := DefaultConfig()
	cfg.PublicKey = keyPair.PublicKeyBase64()

	manager, _ := NewManager(cfg)
	gate := NewFeatureGate(manager)

	// Community tier by default
	if !gate.IsCommunity() {
		t.Error("should be community tier without license")
	}

	err := gate.Require(FeaturePostgreSQL)
	if err != nil {
		t.Error("PostgreSQL should be available")
	}

	err = gate.Require(FeatureOracle)
	if err == nil {
		t.Error("Oracle should require license")
	}

	// RequireAny
	err = gate.RequireAny(FeatureOracle, FeaturePostgreSQL)
	if err != nil {
		t.Error("RequireAny should pass when one feature is available")
	}

	err = gate.RequireAny(FeatureOracle, FeatureSSO)
	if err == nil {
		t.Error("RequireAny should fail when no features available")
	}
}

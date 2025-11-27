package services

import (
	"testing"
)

func TestLicenseService_getLimitsForTier(t *testing.T) {
	service := &LicenseService{}

	tests := []struct {
		name               string
		tier               string
		expectedMaxSources int
		expectedMaxTables  int
		expectedMaxTP      int64
	}{
		{
			name:               "enterprise tier - unlimited",
			tier:               "enterprise",
			expectedMaxSources: 0,
			expectedMaxTables:  0,
			expectedMaxTP:      0,
		},
		{
			name:               "pro tier",
			tier:               "pro",
			expectedMaxSources: 10,
			expectedMaxTables:  100,
			expectedMaxTP:      50000,
		},
		{
			name:               "trial tier",
			tier:               "trial",
			expectedMaxSources: 5,
			expectedMaxTables:  50,
			expectedMaxTP:      10000,
		},
		{
			name:               "community tier",
			tier:               "community",
			expectedMaxSources: 1,
			expectedMaxTables:  10,
			expectedMaxTP:      1000,
		},
		{
			name:               "unknown tier defaults to community",
			tier:               "unknown",
			expectedMaxSources: 1,
			expectedMaxTables:  10,
			expectedMaxTP:      1000,
		},
		{
			name:               "empty tier defaults to community",
			tier:               "",
			expectedMaxSources: 1,
			expectedMaxTables:  10,
			expectedMaxTP:      1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := service.getLimitsForTier(tt.tier)

			if limits.MaxSources != tt.expectedMaxSources {
				t.Errorf("MaxSources = %d, want %d", limits.MaxSources, tt.expectedMaxSources)
			}
			if limits.MaxTables != tt.expectedMaxTables {
				t.Errorf("MaxTables = %d, want %d", limits.MaxTables, tt.expectedMaxTables)
			}
			if limits.MaxThroughput != tt.expectedMaxTP {
				t.Errorf("MaxThroughput = %d, want %d", limits.MaxThroughput, tt.expectedMaxTP)
			}
		})
	}
}

func TestLicenseService_getFeaturesForTier(t *testing.T) {
	service := &LicenseService{}

	tests := []struct {
		name             string
		tier             string
		expectedFeatures []string
		notExpected      []string
	}{
		{
			name: "enterprise tier has all features",
			tier: "enterprise",
			expectedFeatures: []string{
				"postgresql", "mysql", "mariadb",
				"mongodb", "sqlserver", "cassandra", "dynamodb",
				"oracle", "ha", "raft_cluster", "sso", "ldap", "audit_log",
			},
		},
		{
			name: "pro tier has mid-level features",
			tier: "pro",
			expectedFeatures: []string{
				"postgresql", "mysql", "mariadb",
				"mongodb", "sqlserver", "cassandra", "dynamodb",
				"snapshot", "kafka_output", "grpc_output",
			},
			notExpected: []string{"oracle", "ha", "raft_cluster", "sso", "ldap"},
		},
		{
			name: "trial tier same as pro",
			tier: "trial",
			expectedFeatures: []string{
				"postgresql", "mysql", "mariadb",
				"mongodb", "sqlserver",
			},
			notExpected: []string{"oracle", "ha", "raft_cluster"},
		},
		{
			name:             "community tier has basic features",
			tier:             "community",
			expectedFeatures: []string{"postgresql", "mysql", "mariadb"},
			notExpected:      []string{"mongodb", "oracle", "ha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := service.getFeaturesForTier(tt.tier)
			featureSet := make(map[string]bool)
			for _, f := range features {
				featureSet[f] = true
			}

			for _, expected := range tt.expectedFeatures {
				if !featureSet[expected] {
					t.Errorf("tier %q should have feature %q", tt.tier, expected)
				}
			}

			for _, notExpected := range tt.notExpected {
				if featureSet[notExpected] {
					t.Errorf("tier %q should NOT have feature %q", tt.tier, notExpected)
				}
			}
		})
	}
}

func TestLicenseService_getMaxActivations(t *testing.T) {
	service := &LicenseService{}

	tests := []struct {
		name     string
		tier     string
		expected int
	}{
		{
			name:     "enterprise unlimited activations",
			tier:     "enterprise",
			expected: 0, // 0 means unlimited
		},
		{
			name:     "pro 10 activations",
			tier:     "pro",
			expected: 10,
		},
		{
			name:     "trial 2 activations",
			tier:     "trial",
			expected: 2,
		},
		{
			name:     "community 1 activation",
			tier:     "community",
			expected: 1,
		},
		{
			name:     "unknown defaults to 1",
			tier:     "unknown",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getMaxActivations(tt.tier)
			if result != tt.expected {
				t.Errorf("getMaxActivations(%q) = %d, want %d", tt.tier, result, tt.expected)
			}
		})
	}
}

func TestLicenseService_Errors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "license not found",
			err:      ErrLicenseNotFound,
			expected: "license not found",
		},
		{
			name:     "license expired",
			err:      ErrLicenseExpired,
			expected: "license has expired",
		},
		{
			name:     "license revoked",
			err:      ErrLicenseRevoked,
			expected: "license has been revoked",
		},
		{
			name:     "invalid license",
			err:      ErrInvalidLicense,
			expected: "invalid license format",
		},
		{
			name:     "invalid signature",
			err:      ErrInvalidSignature,
			expected: "license signature verification failed",
		},
		{
			name:     "hardware mismatch",
			err:      ErrHardwareMismatch,
			expected: "license is bound to different hardware",
		},
		{
			name:     "activation limit reached",
			err:      ErrActivationLimitReached,
			expected: "activation limit reached",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("error = %q, want %q", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestNewLicenseService(t *testing.T) {
	// Test creating service without private key
	service := NewLicenseService(nil, "")

	if service == nil {
		t.Fatal("service should not be nil")
	}

	if service.issuer != "license.savegress.io" {
		t.Errorf("issuer = %q, want %q", service.issuer, "license.savegress.io")
	}

	if service.generator != nil {
		t.Error("generator should be nil when no private key is provided")
	}
}

func TestLimits_Structure(t *testing.T) {
	l := limits{
		MaxSources:    10,
		MaxTables:     100,
		MaxThroughput: 50000,
	}

	if l.MaxSources != 10 {
		t.Errorf("MaxSources = %d, want 10", l.MaxSources)
	}
	if l.MaxTables != 100 {
		t.Errorf("MaxTables = %d, want 100", l.MaxTables)
	}
	if l.MaxThroughput != 50000 {
		t.Errorf("MaxThroughput = %d, want 50000", l.MaxThroughput)
	}
}

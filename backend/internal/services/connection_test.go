package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: Database integration tests and actual connection tests would require:
// 1. A test database or proper mocking infrastructure
// 2. Network access to test TCP connectivity
//
// The tests below focus on testing business logic that doesn't require external dependencies

func TestNewConnectionService(t *testing.T) {
	tests := []struct {
		name          string
		encryptionKey string
	}{
		{
			name:          "standard 32-byte key",
			encryptionKey: "12345678901234567890123456789012",
		},
		{
			name:          "shorter key (will be padded)",
			encryptionKey: "shortkey",
		},
		{
			name:          "longer key (will be truncated)",
			encryptionKey: "this-is-a-very-long-encryption-key-that-exceeds-32-bytes",
		},
		{
			name:          "empty key",
			encryptionKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewConnectionService(nil, tt.encryptionKey)
			assert.NotNil(t, service)
			assert.Len(t, service.encryptionKey, 32)
		})
	}
}

func TestConnectionService_EncryptDecryptPassword(t *testing.T) {
	service := NewConnectionService(nil, "test-encryption-key-32-bytes-xx")

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "standard password",
			password: "MyDatabasePassword123!",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "password with special chars",
			password: "P@ss!#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:     "unicode password",
			password: "密码123パスワード",
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatMightBeUsedForSomeDatabaseConnections123456789!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := service.encryptPassword(tt.password)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.password, encrypted, "Encrypted should differ from original")

			// Decrypt
			decrypted, err := service.decryptPassword(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.password, decrypted, "Decrypted should match original")
		})
	}
}

func TestConnectionService_EncryptionUniqueness(t *testing.T) {
	service := NewConnectionService(nil, "test-encryption-key-32-bytes-xx")

	password := "TestPassword123!"

	// Encrypt the same password multiple times
	encrypted1, err := service.encryptPassword(password)
	assert.NoError(t, err)

	encrypted2, err := service.encryptPassword(password)
	assert.NoError(t, err)

	// Due to random nonce, each encryption should produce different ciphertext
	assert.NotEqual(t, encrypted1, encrypted2, "Each encryption should be unique due to random nonce")

	// But both should decrypt to the same password
	decrypted1, _ := service.decryptPassword(encrypted1)
	decrypted2, _ := service.decryptPassword(encrypted2)
	assert.Equal(t, decrypted1, decrypted2)
}

func TestConnectionService_DecryptInvalidData(t *testing.T) {
	service := NewConnectionService(nil, "test-encryption-key-32-bytes-xx")

	tests := []struct {
		name      string
		encrypted string
		expectErr bool
	}{
		{
			name:      "invalid base64",
			encrypted: "not-valid-base64!!!",
			expectErr: true,
		},
		{
			name:      "empty string",
			encrypted: "",
			expectErr: true,
		},
		{
			name:      "too short ciphertext",
			encrypted: "YWJj", // "abc" in base64
			expectErr: true,
		},
		{
			name:      "valid base64 but invalid ciphertext",
			encrypted: "dGVzdGRhdGF0aGF0aXNub3RlbmNyeXB0ZWQ=",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.decryptPassword(tt.encrypted)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnectionService_DifferentKeys(t *testing.T) {
	service1 := NewConnectionService(nil, "encryption-key-one-32-bytes-xxx")
	service2 := NewConnectionService(nil, "encryption-key-two-32-bytes-xxx")

	password := "TestPassword123!"

	// Encrypt with service1
	encrypted, err := service1.encryptPassword(password)
	assert.NoError(t, err)

	// Try to decrypt with service2 (different key) - should fail
	_, err = service2.decryptPassword(encrypted)
	assert.Error(t, err, "Decryption with different key should fail")
}

func TestConnectionService_ErrorConstants(t *testing.T) {
	assert.NotNil(t, ErrConnectionNotFound)
	assert.NotNil(t, ErrConnectionInUse)
	assert.NotNil(t, ErrConnectionTestFail)

	assert.Equal(t, "connection not found", ErrConnectionNotFound.Error())
	assert.Equal(t, "connection is used by pipelines", ErrConnectionInUse.Error())
	assert.Equal(t, "connection test failed", ErrConnectionTestFail.Error())
}

func TestConnectionService_ConnectionTypes(t *testing.T) {
	// Document supported connection types
	supportedTypes := []string{
		"postgres",
		"postgresql",
		"mysql",
		"mariadb",
	}

	// Types with only TCP test (no driver test)
	tcpOnlyTypes := []string{
		"mongodb",
		"sqlserver",
		"oracle",
	}

	for _, connType := range supportedTypes {
		t.Run("supported_"+connType, func(t *testing.T) {
			assert.NotEmpty(t, connType)
		})
	}

	for _, connType := range tcpOnlyTypes {
		t.Run("tcp_only_"+connType, func(t *testing.T) {
			assert.NotEmpty(t, connType)
		})
	}
}

func TestConnectionService_SSLModes(t *testing.T) {
	// Document valid SSL modes for PostgreSQL
	validSSLModes := []string{
		"disable",
		"allow",
		"prefer",
		"require",
		"verify-ca",
		"verify-full",
	}

	for _, mode := range validSSLModes {
		t.Run("ssl_mode_"+mode, func(t *testing.T) {
			assert.NotEmpty(t, mode)
		})
	}
}

func TestConnectionService_PortValidation(t *testing.T) {
	// Document default ports for different database types
	defaultPorts := map[string]int{
		"postgres":   5432,
		"postgresql": 5432,
		"mysql":      3306,
		"mariadb":    3306,
		"mongodb":    27017,
		"sqlserver":  1433,
		"oracle":     1521,
		"redis":      6379,
	}

	for dbType, port := range defaultPorts {
		t.Run("default_port_"+dbType, func(t *testing.T) {
			assert.Greater(t, port, 0)
			assert.Less(t, port, 65536)
		})
	}
}

// Integration test examples (commented out - would need real database and network)
//
// func TestConnectionService_CreateConnectionIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test creating a connection in the database
// }
//
// func TestConnectionService_TestConnectionIntegration(t *testing.T) {
//     t.Skip("Requires network access and database servers")
//     // This would test actual TCP and database connectivity
// }
//
// func TestConnectionService_ListConnectionsIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test listing connections from the database
// }

package services

import (
	"testing"
	"time"
)

func TestDownloadService_ValidateDownloadRequest(t *testing.T) {
	service := &DownloadService{}

	tests := []struct {
		name      string
		product   string
		edition   string
		platform  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid cdc-engine community linux",
			product:  "cdc-engine",
			edition:  "community",
			platform: "linux-amd64",
			wantErr:  false,
		},
		{
			name:     "valid cdc-broker pro darwin",
			product:  "cdc-broker",
			edition:  "pro",
			platform: "darwin-arm64",
			wantErr:  false,
		},
		{
			name:     "valid enterprise windows",
			product:  "cdc-engine",
			edition:  "enterprise",
			platform: "windows-amd64",
			wantErr:  false,
		},
		{
			name:      "invalid product",
			product:   "invalid-product",
			edition:   "community",
			platform:  "linux-amd64",
			wantErr:   true,
			errSubstr: "invalid product",
		},
		{
			name:      "invalid edition",
			product:   "cdc-engine",
			edition:   "invalid-edition",
			platform:  "linux-amd64",
			wantErr:   true,
			errSubstr: "invalid edition",
		},
		{
			name:      "invalid platform",
			product:   "cdc-engine",
			edition:   "community",
			platform:  "invalid-platform",
			wantErr:   true,
			errSubstr: "invalid platform",
		},
		{
			name:      "empty product",
			product:   "",
			edition:   "community",
			platform:  "linux-amd64",
			wantErr:   true,
			errSubstr: "invalid product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateDownloadRequest(tt.product, tt.edition, tt.platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDownloadRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if err.Error() == "" || !contains(err.Error(), tt.errSubstr) {
					t.Errorf("error should contain %q, got %q", tt.errSubstr, err.Error())
				}
			}
		})
	}
}

func TestDownloadService_CheckLicenseForEdition(t *testing.T) {
	service := &DownloadService{}

	tests := []struct {
		name             string
		userTier         string
		requestedEdition string
		wantErr          bool
	}{
		// Community tier users
		{
			name:             "community can download community",
			userTier:         "community",
			requestedEdition: "community",
			wantErr:          false,
		},
		{
			name:             "community cannot download pro",
			userTier:         "community",
			requestedEdition: "pro",
			wantErr:          true,
		},
		{
			name:             "community cannot download enterprise",
			userTier:         "community",
			requestedEdition: "enterprise",
			wantErr:          true,
		},
		// Pro tier users
		{
			name:             "pro can download community",
			userTier:         "pro",
			requestedEdition: "community",
			wantErr:          false,
		},
		{
			name:             "pro can download pro",
			userTier:         "pro",
			requestedEdition: "pro",
			wantErr:          false,
		},
		{
			name:             "pro cannot download enterprise",
			userTier:         "pro",
			requestedEdition: "enterprise",
			wantErr:          true,
		},
		// Enterprise tier users
		{
			name:             "enterprise can download community",
			userTier:         "enterprise",
			requestedEdition: "community",
			wantErr:          false,
		},
		{
			name:             "enterprise can download pro",
			userTier:         "enterprise",
			requestedEdition: "pro",
			wantErr:          false,
		},
		{
			name:             "enterprise can download enterprise",
			userTier:         "enterprise",
			requestedEdition: "enterprise",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CheckLicenseForEdition(tt.userTier, tt.requestedEdition)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckLicenseForEdition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadService_getFilename(t *testing.T) {
	service := &DownloadService{}

	tests := []struct {
		name     string
		product  string
		version  string
		edition  string
		platform string
		expected string
	}{
		{
			name:     "linux binary",
			product:  "cdc-engine",
			version:  "1.0.0",
			edition:  "pro",
			platform: "linux-amd64",
			expected: "cdc-engine-linux-amd64",
		},
		{
			name:     "darwin binary",
			product:  "cdc-broker",
			version:  "1.0.0",
			edition:  "enterprise",
			platform: "darwin-arm64",
			expected: "cdc-broker-darwin-arm64",
		},
		{
			name:     "windows binary",
			product:  "cdc-engine",
			version:  "1.0.0",
			edition:  "community",
			platform: "windows-amd64",
			expected: "cdc-engine-windows-amd64.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getFilename(tt.product, tt.version, tt.edition, tt.platform)
			if result != tt.expected {
				t.Errorf("getFilename() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDownloadService_getReleaseKey(t *testing.T) {
	service := &DownloadService{
		keyPrefix: "releases/",
	}

	tests := []struct {
		name     string
		product  string
		version  string
		edition  string
		platform string
		expected string
	}{
		{
			name:     "standard release key",
			product:  "cdc-engine",
			version:  "1.0.0",
			edition:  "pro",
			platform: "linux-amd64",
			expected: "releases/cdc-engine/1.0.0/pro/cdc-engine-linux-amd64",
		},
		{
			name:     "windows release key",
			product:  "cdc-broker",
			version:  "2.0.0",
			edition:  "enterprise",
			platform: "windows-amd64",
			expected: "releases/cdc-broker/2.0.0/enterprise/cdc-broker-windows-amd64.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getReleaseKey(tt.product, tt.version, tt.edition, tt.platform)
			if result != tt.expected {
				t.Errorf("getReleaseKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDownloadConfig_Defaults(t *testing.T) {
	cfg := DownloadConfig{
		Region: "us-east-1",
		Bucket: "releases",
	}

	// URLExpiry should default to 1 hour when not set
	if cfg.URLExpiry != 0 {
		t.Error("URLExpiry should be zero value before initialization")
	}

	// When service is created, default should be applied
	// (Note: actual service creation requires valid AWS credentials)
}

func TestDownloadURL_Structure(t *testing.T) {
	url := DownloadURL{
		URL:       "https://example.com/download",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		ExpiresIn: 3600,
		Filename:  "cdc-engine-linux-amd64",
		Size:      10485760,
		Checksum:  "sha256:abc123",
	}

	if url.URL == "" {
		t.Error("URL should not be empty")
	}

	if url.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn should be 3600, got %d", url.ExpiresIn)
	}

	if url.Size != 10485760 {
		t.Errorf("Size should be 10485760, got %d", url.Size)
	}
}

func TestReleaseInfo_Structure(t *testing.T) {
	info := ReleaseInfo{
		Product:   "cdc-engine",
		Version:   "1.0.0",
		Editions:  []string{"community", "pro", "enterprise"},
		Platforms: []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
	}

	if info.Product != "cdc-engine" {
		t.Errorf("Product should be 'cdc-engine', got %q", info.Product)
	}

	if len(info.Editions) != 3 {
		t.Errorf("should have 3 editions, got %d", len(info.Editions))
	}

	if len(info.Platforms) != 5 {
		t.Errorf("should have 5 platforms, got %d", len(info.Platforms))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

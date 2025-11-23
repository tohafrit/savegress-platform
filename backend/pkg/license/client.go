package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LicenseClient communicates with the license server
type LicenseClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLicenseClient creates a new license client
func NewLicenseClient(baseURL string) *LicenseClient {
	return &LicenseClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateRequest is sent to the license server
type ValidateRequest struct {
	LicenseID  string `json:"license_id"`
	HardwareID string `json:"hardware_id"`
	Version    string `json:"version"`    // Software version
	Timestamp  int64  `json:"timestamp"`  // Current Unix timestamp
	Checksum   string `json:"checksum"`   // HMAC of request for integrity
}

// ValidateResponse from the license server
type ValidateResponse struct {
	Valid        bool     `json:"valid"`
	License      *License `json:"license,omitempty"`      // Updated license if changed
	Revoked      bool     `json:"revoked"`
	RevokeReason string   `json:"revoke_reason,omitempty"`
	Message      string   `json:"message,omitempty"`
	ServerTime   int64    `json:"server_time"`            // For clock sync check
	NextCheck    int64    `json:"next_check"`             // When to check next (Unix)
}

// Validate checks a license with the server
func (c *LicenseClient) Validate(licenseID, hardwareID string) (*ValidateResponse, error) {
	req := ValidateRequest{
		LicenseID:  licenseID,
		HardwareID: hardwareID,
		Timestamp:  time.Now().Unix(),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/validate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Savegress-Engine/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("license server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("license server error: %s", resp.Status)
	}

	var validateResp ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &validateResp, nil
}

// ActivateRequest for initial license activation
type ActivateRequest struct {
	LicenseKey string `json:"license_key"`
	HardwareID string `json:"hardware_id"`
	Hostname   string `json:"hostname"`
	Platform   string `json:"platform"`
}

// ActivateResponse from activation
type ActivateResponse struct {
	Success    bool     `json:"success"`
	License    *License `json:"license,omitempty"`
	InstanceID string   `json:"instance_id"` // Unique ID for this installation
	Message    string   `json:"message,omitempty"`
}

// Activate activates a license for this machine
func (c *LicenseClient) Activate(key LicenseKey, hardwareID, hostname, platform string) (*ActivateResponse, error) {
	req := ActivateRequest{
		LicenseKey: string(key),
		HardwareID: hardwareID,
		Hostname:   hostname,
		Platform:   platform,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/activate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Savegress-Engine/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("license server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("activation failed: %s", resp.Status)
	}

	var activateResp ActivateResponse
	if err := json.NewDecoder(resp.Body).Decode(&activateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &activateResp, nil
}

// DeactivateRequest for license deactivation
type DeactivateRequest struct {
	LicenseID  string `json:"license_id"`
	InstanceID string `json:"instance_id"`
	HardwareID string `json:"hardware_id"`
}

// Deactivate removes a license activation
func (c *LicenseClient) Deactivate(licenseID, instanceID, hardwareID string) error {
	req := DeactivateRequest{
		LicenseID:  licenseID,
		InstanceID: instanceID,
		HardwareID: hardwareID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/deactivate", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Savegress-Engine/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("license server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deactivation failed: %s", resp.Status)
	}

	return nil
}

// GetLicenseInfo retrieves license information from server
func (c *LicenseClient) GetLicenseInfo(licenseID string) (*License, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/license/"+licenseID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "Savegress-Engine/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("license server unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoLicense
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", resp.Status)
	}

	var license License
	if err := json.NewDecoder(resp.Body).Decode(&license); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &license, nil
}

package license

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
)

// CLI provides command-line interface for license management
type CLI struct {
	manager       *Manager
	licenseServer string
	outputFormat  string // "text" or "json"
}

// NewCLI creates a new license CLI
func NewCLI(manager *Manager, licenseServer string) *CLI {
	return &CLI{
		manager:       manager,
		licenseServer: licenseServer,
		outputFormat:  "text",
	}
}

// SetOutputFormat sets output format ("text" or "json")
func (c *CLI) SetOutputFormat(format string) {
	c.outputFormat = format
}

// Activate activates a license key
func (c *CLI) Activate(licenseKey string) error {
	key := LicenseKey(licenseKey)

	// First validate locally
	if err := c.manager.LoadFromKey(key); err != nil {
		return fmt.Errorf("invalid license key: %w", err)
	}

	// Then activate with server
	client := NewLicenseClient(c.licenseServer)
	hardwareID := GetHardwareIDWithFallback()
	hostname, _ := os.Hostname()

	resp, err := client.Activate(key, hardwareID, hostname, runtime.GOOS)
	if err != nil {
		// Offline activation - just use local validation
		c.printSuccess("License activated (offline mode)")
		return nil
	}

	if !resp.Success {
		return fmt.Errorf("activation failed: %s", resp.Message)
	}

	c.printSuccess(fmt.Sprintf("License activated successfully. Instance ID: %s", resp.InstanceID))
	return nil
}

// Deactivate deactivates the current license
func (c *CLI) Deactivate() error {
	license := c.manager.GetLicense()
	if license == nil {
		return fmt.Errorf("no license is currently active")
	}

	client := NewLicenseClient(c.licenseServer)
	hardwareID := GetHardwareIDWithFallback()

	err := client.Deactivate(license.ID, "", hardwareID)
	if err != nil {
		return fmt.Errorf("deactivation failed: %w", err)
	}

	c.printSuccess("License deactivated successfully")
	return nil
}

// Status shows current license status
func (c *CLI) Status() error {
	status := c.manager.GetStatus()
	license := c.manager.GetLicense()

	if c.outputFormat == "json" {
		return c.outputJSON(map[string]interface{}{
			"status":  status,
			"license": license,
			"edition": Edition,
			"hardware_id": GetHardwareIDWithFallback(),
		})
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "┌─────────────────────────────────────────────────────────┐")
	fmt.Fprintln(w, "│                  SAVEGRESS LICENSE STATUS               │")
	fmt.Fprintln(w, "└─────────────────────────────────────────────────────────┘")
	fmt.Fprintln(w, "")

	// Edition info
	fmt.Fprintf(w, "  Edition:\t%s\n", EditionFull)
	fmt.Fprintf(w, "  Build:\t%s\n", Edition)

	fmt.Fprintln(w, "")

	if license == nil {
		fmt.Fprintln(w, "  Status:\t⚠️  No license (Community mode)")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  Available features:")
		for _, f := range CommunityFeatures {
			fmt.Fprintf(w, "    • %s\n", f)
		}
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  Limits:")
		fmt.Fprintf(w, "    • Max sources:\t%d\n", CommunityLimits.MaxSources)
		fmt.Fprintf(w, "    • Max tables:\t%d\n", CommunityLimits.MaxTables)
		fmt.Fprintf(w, "    • Max throughput:\t%d events/sec\n", CommunityLimits.MaxThroughput)
	} else {
		// License info
		statusIcon := "✅"
		if !status.Valid {
			statusIcon = "❌"
		} else if status.GracePeriod {
			statusIcon = "⚠️"
		}

		fmt.Fprintf(w, "  Status:\t%s %s\n", statusIcon, c.statusText(status))
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  License ID:\t%s\n", license.ID)
		fmt.Fprintf(w, "  Customer:\t%s\n", license.CustomerName)
		fmt.Fprintf(w, "  Tier:\t\t%s\n", strings.ToUpper(string(license.Tier)))
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  Issued:\t%s\n", license.IssuedAt.Format("2006-01-02"))
		fmt.Fprintf(w, "  Expires:\t%s\n", license.ExpiresAt.Format("2006-01-02"))
		fmt.Fprintf(w, "  Days remaining:\t%d\n", status.DaysRemaining)

		if license.HardwareID != "" {
			fmt.Fprintln(w, "")
			fmt.Fprintf(w, "  Hardware bound:\tYes\n")
		}

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  Licensed features:")
		for _, f := range license.Features {
			fmt.Fprintf(w, "    • %s\n", f)
		}

		if license.Limits.MaxSources > 0 || license.Limits.MaxTables > 0 || license.Limits.MaxThroughput > 0 {
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "  Limits:")
			if license.Limits.MaxSources > 0 {
				fmt.Fprintf(w, "    • Max sources:\t%d\n", license.Limits.MaxSources)
			} else {
				fmt.Fprintf(w, "    • Max sources:\tunlimited\n")
			}
			if license.Limits.MaxTables > 0 {
				fmt.Fprintf(w, "    • Max tables:\t%d\n", license.Limits.MaxTables)
			} else {
				fmt.Fprintf(w, "    • Max tables:\tunlimited\n")
			}
			if license.Limits.MaxThroughput > 0 {
				fmt.Fprintf(w, "    • Max throughput:\t%d events/sec\n", license.Limits.MaxThroughput)
			} else {
				fmt.Fprintf(w, "    • Max throughput:\tunlimited\n")
			}
		}

		if status.GracePeriod {
			fmt.Fprintln(w, "")
			fmt.Fprintf(w, "  ⚠️  Running in grace period: %s\n", status.Message)
		}
	}

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Hardware ID:\t%s\n", GetHardwareIDWithFallback())
	fmt.Fprintln(w, "")

	return nil
}

func (c *CLI) statusText(status LicenseStatus) string {
	if !status.Valid {
		if status.Message != "" {
			return status.Message
		}
		return "Invalid"
	}

	if status.GracePeriod {
		return "Valid (grace period)"
	}

	if status.DaysRemaining <= 7 {
		return fmt.Sprintf("Valid (expires in %d days)", status.DaysRemaining)
	}

	return "Valid"
}

// Info shows detailed license information
func (c *CLI) Info(licenseKey string) error {
	// Parse without validating (to show info even if expired)
	parts := strings.Split(licenseKey, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid license key format")
	}

	// Try to load and verify
	err := c.manager.LoadFromKey(LicenseKey(licenseKey))
	if err != nil && err != ErrLicenseExpired {
		return fmt.Errorf("invalid license: %w", err)
	}

	license := c.manager.GetLicense()
	if license == nil {
		return fmt.Errorf("could not parse license")
	}

	if c.outputFormat == "json" {
		return c.outputJSON(license)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "License Information:")
	fmt.Fprintln(w, strings.Repeat("-", 50))
	fmt.Fprintf(w, "  ID:\t\t%s\n", license.ID)
	fmt.Fprintf(w, "  Customer ID:\t%s\n", license.CustomerID)
	fmt.Fprintf(w, "  Customer:\t%s\n", license.CustomerName)
	fmt.Fprintf(w, "  Tier:\t\t%s\n", license.Tier)
	fmt.Fprintf(w, "  Issued:\t%s\n", license.IssuedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "  Expires:\t%s\n", license.ExpiresAt.Format(time.RFC3339))
	fmt.Fprintf(w, "  Issuer:\t%s\n", license.Issuer)
	fmt.Fprintf(w, "  Version:\t%d\n", license.Version)

	if license.HardwareID != "" {
		fmt.Fprintf(w, "  Hardware:\t%s\n", license.HardwareID)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  Features:")
	for _, f := range license.Features {
		fmt.Fprintf(w, "    - %s\n", f)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  Limits:")
	fmt.Fprintf(w, "    Max Sources:\t%d (0=unlimited)\n", license.Limits.MaxSources)
	fmt.Fprintf(w, "    Max Tables:\t\t%d (0=unlimited)\n", license.Limits.MaxTables)
	fmt.Fprintf(w, "    Max Throughput:\t%d events/sec (0=unlimited)\n", license.Limits.MaxThroughput)

	// Expiry status
	if time.Now().After(license.ExpiresAt) {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  ❌ LICENSE EXPIRED")
	}

	fmt.Fprintln(w, "")

	return nil
}

// Verify checks if the current license is valid
func (c *CLI) Verify() error {
	status := c.manager.GetStatus()

	if c.outputFormat == "json" {
		return c.outputJSON(status)
	}

	if !status.Valid {
		return fmt.Errorf("license is not valid: %s", status.Message)
	}

	c.printSuccess("License is valid")
	return nil
}

// Features lists all available features
func (c *CLI) Features() error {
	license := c.manager.GetLicense()

	type featureInfo struct {
		Name      Feature `json:"name"`
		Available bool    `json:"available"`
		Compiled  bool    `json:"compiled"`
		Tier      string  `json:"required_tier"`
	}

	allFeatures := append(append(
		CommunityFeatures,
		ProFeatures...),
		EnterpriseFeatures...,
	)

	features := make([]featureInfo, 0, len(allFeatures))

	for _, f := range allFeatures {
		info := featureInfo{
			Name:     f,
			Compiled: true, // In dev build, all compiled
		}

		// Determine required tier
		for _, ef := range EnterpriseFeatures {
			if ef == f {
				info.Tier = "enterprise"
				break
			}
		}
		if info.Tier == "" {
			for _, pf := range ProFeatures {
				if pf == f {
					info.Tier = "pro"
					break
				}
			}
		}
		if info.Tier == "" {
			info.Tier = "community"
		}

		// Check if available with current license
		info.Available = c.manager.HasFeature(f)

		features = append(features, info)
	}

	if c.outputFormat == "json" {
		return c.outputJSON(features)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Feature\tRequired Tier\tAvailable")
	fmt.Fprintln(w, strings.Repeat("-", 50))

	for _, f := range features {
		available := "❌"
		if f.Available {
			available = "✅"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", f.Name, f.Tier, available)
	}

	fmt.Fprintln(w, "")

	if license != nil {
		fmt.Fprintf(w, "Current tier: %s\n", strings.ToUpper(string(license.Tier)))
	} else {
		fmt.Fprintln(w, "Current tier: COMMUNITY (no license)")
	}
	fmt.Fprintln(w, "")

	return nil
}

func (c *CLI) printSuccess(msg string) {
	if c.outputFormat == "json" {
		c.outputJSON(map[string]string{"status": "success", "message": msg})
		return
	}
	fmt.Println("✅", msg)
}

func (c *CLI) outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// GenerateHardwareIDCommand prints the hardware ID
func (c *CLI) GenerateHardwareIDCommand() error {
	hwID := GetHardwareIDWithFallback()

	if c.outputFormat == "json" {
		return c.outputJSON(map[string]string{"hardware_id": hwID})
	}

	fmt.Println("Hardware ID:", hwID)
	return nil
}

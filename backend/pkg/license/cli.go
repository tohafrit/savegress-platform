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

// ============================================
// AUTO-LOGIN FLOW
// ============================================
// These methods allow users to login with email/password
// and automatically retrieve their license key.

// Login authenticates with email/password and retrieves license
func (c *CLI) Login(email, password string) error {
	client := NewLicenseClient(c.licenseServer)

	fmt.Printf("Logging in as %s...\n", email)

	// Authenticate and get license
	resp, err := client.Login(email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if resp.LicenseKey == "" {
		return fmt.Errorf("no active license found for this account. Please subscribe at https://savegress.io/billing")
	}

	// Save license key to config file
	if err := c.saveLicenseKey(resp.LicenseKey); err != nil {
		return fmt.Errorf("failed to save license: %w", err)
	}

	// Load the license
	if err := c.manager.LoadFromKey(LicenseKey(resp.LicenseKey)); err != nil {
		return fmt.Errorf("invalid license received: %w", err)
	}

	license := c.manager.GetLicense()
	c.printSuccess(fmt.Sprintf("Logged in successfully! License: %s (%s)", license.Tier, license.CustomerName))

	return nil
}

// LoginInteractive prompts for email and password interactively
func (c *CLI) LoginInteractive() error {
	fmt.Println("")
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│              SAVEGRESS LICENSE LOGIN                    │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println("")
	fmt.Println("No license key found. Please login with your Savegress account")
	fmt.Println("to retrieve your license automatically.")
	fmt.Println("")
	fmt.Println("Don't have an account? Sign up at https://savegress.io/register")
	fmt.Println("")

	// Read email
	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)

	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Read password (Note: In production, use terminal.ReadPassword for hidden input)
	fmt.Print("Password: ")
	var password string
	fmt.Scanln(&password)

	if password == "" {
		return fmt.Errorf("password is required")
	}

	return c.Login(email, password)
}

// Logout removes the saved license key
func (c *CLI) Logout() error {
	configPath := c.getLicenseConfigPath()

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove license file: %w", err)
	}

	c.printSuccess("Logged out successfully. License key removed.")
	return nil
}

// CheckOrPromptLogin checks if license exists, prompts login if not
func (c *CLI) CheckOrPromptLogin() error {
	// Try to load from embedded key first
	embeddedKey := c.getEmbeddedLicenseKey()
	if embeddedKey != "" {
		if err := c.manager.LoadFromKey(LicenseKey(embeddedKey)); err == nil {
			return nil // Embedded license is valid
		}
	}

	// Try to load from config file
	configPath := c.getLicenseConfigPath()
	if err := c.manager.LoadFromFile(configPath); err == nil {
		return nil // Saved license is valid
	}

	// Try to load from environment variable
	if err := c.manager.LoadFromEnv("SAVEGRESS_LICENSE_KEY"); err == nil {
		return nil // ENV license is valid
	}

	// No license found - prompt for login
	fmt.Println("")
	fmt.Println("⚠️  No license key found.")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  1. Run 'cdc-engine license login' to login with your account")
	fmt.Println("  2. Set SAVEGRESS_LICENSE_KEY environment variable")
	fmt.Println("  3. Download a personalized binary from https://savegress.io/downloads")
	fmt.Println("")
	fmt.Println("Running in Community mode with limited features...")
	fmt.Println("")

	return nil // Allow running in community mode
}

// saveLicenseKey saves the license key to config file
func (c *CLI) saveLicenseKey(key string) error {
	configPath := c.getLicenseConfigPath()

	// Create config directory if it doesn't exist
	configDir := c.getConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write license key with restricted permissions
	if err := os.WriteFile(configPath, []byte(key), 0600); err != nil {
		return fmt.Errorf("failed to write license file: %w", err)
	}

	return nil
}

// getLicenseConfigPath returns the path to the license config file
func (c *CLI) getLicenseConfigPath() string {
	return c.getConfigDir() + "/license.key"
}

// getConfigDir returns the config directory path
func (c *CLI) getConfigDir() string {
	// Use XDG config dir on Linux, or home dir on other platforms
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		return configDir + "/savegress"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/savegress"
	}

	switch runtime.GOOS {
	case "darwin":
		return homeDir + "/Library/Application Support/Savegress"
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData + "\\Savegress"
		}
		return homeDir + "\\.savegress"
	default:
		return homeDir + "/.config/savegress"
	}
}

// getEmbeddedLicenseKey returns the embedded license key (replaced at download time)
func (c *CLI) getEmbeddedLicenseKey() string {
	// This placeholder is replaced with the actual license key when downloading
	// from the portal. The key is exactly 512 bytes to allow binary replacement.
	const embeddedKey = "SAVEGRESS_LICENSE_PLACEHOLDER_" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000"

	// Check if placeholder was replaced with actual key
	if strings.HasPrefix(embeddedKey, "SAVEGRESS_LICENSE_PLACEHOLDER_") {
		return "" // Still placeholder, no embedded key
	}

	// Trim null padding from the key
	return strings.TrimRight(embeddedKey, "\x00")
}

package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// HardwareInfo contains machine identification data
type HardwareInfo struct {
	MachineID   string   `json:"machine_id"`
	Hostname    string   `json:"hostname"`
	Platform    string   `json:"platform"`
	Arch        string   `json:"arch"`
	CPUCount    int      `json:"cpu_count"`
	MACAddresses []string `json:"mac_addresses"`
}

// GenerateHardwareID creates a unique hardware fingerprint
// This is used for license binding to prevent license sharing
func GenerateHardwareID() (string, error) {
	info, err := CollectHardwareInfo()
	if err != nil {
		return "", err
	}

	// Create a stable hash from hardware info
	// We use multiple factors for resilience against hardware changes
	data := fmt.Sprintf("%s|%s|%s|%v",
		info.MachineID,
		info.Platform,
		info.Arch,
		info.MACAddresses,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]), nil // First 16 bytes = 32 hex chars
}

// CollectHardwareInfo gathers hardware identification data
func CollectHardwareInfo() (*HardwareInfo, error) {
	info := &HardwareInfo{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		info.Hostname = hostname
	}

	// Get machine ID (platform specific)
	info.MachineID = getMachineID()

	// Get MAC addresses
	info.MACAddresses = getMACAddresses()

	return info, nil
}

// getMachineID returns a platform-specific machine identifier
func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxMachineID()
	case "darwin":
		return getDarwinMachineID()
	case "windows":
		return getWindowsMachineID()
	default:
		return ""
	}
}

func getLinuxMachineID() string {
	// Try /etc/machine-id first (systemd)
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Try /var/lib/dbus/machine-id
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Try DMI product UUID
	if data, err := os.ReadFile("/sys/class/dmi/id/product_uuid"); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}

func getDarwinMachineID() string {
	// Get hardware UUID on macOS
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse output for IOPlatformUUID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				uuid := strings.TrimSpace(parts[1])
				uuid = strings.Trim(uuid, "\"")
				return uuid
			}
		}
	}

	return ""
}

func getWindowsMachineID() string {
	// Get MachineGuid from registry
	cmd := exec.Command("reg", "query",
		`HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`,
		"/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "MachineGuid") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2]
			}
		}
	}

	return ""
}

// getMACAddresses returns sorted list of MAC addresses
func getMACAddresses() []string {
	var macs []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return macs
	}

	for _, iface := range interfaces {
		// Skip loopback and interfaces without MAC
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		mac := iface.HardwareAddr.String()
		// Skip virtual/docker interfaces (common prefixes)
		if strings.HasPrefix(mac, "02:42:") || // Docker
			strings.HasPrefix(mac, "00:00:00:") ||
			strings.HasPrefix(mac, "fe:") {
			continue
		}

		macs = append(macs, mac)
	}

	// Sort for consistency
	sort.Strings(macs)
	return macs
}

// KubernetesHardwareID generates a hardware ID suitable for Kubernetes
// where traditional hardware IDs may not be stable
func KubernetesHardwareID() string {
	var components []string

	// Use pod name if available
	if podName := os.Getenv("POD_NAME"); podName != "" {
		components = append(components, podName)
	}

	// Use namespace
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		components = append(components, ns)
	}

	// Use node name
	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		components = append(components, nodeName)
	}

	// Use service account
	if saData, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		components = append(components, string(saData))
	}

	if len(components) == 0 {
		return ""
	}

	data := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(data))
	return "k8s-" + hex.EncodeToString(hash[:12])
}

// IsKubernetes detects if running in Kubernetes
func IsKubernetes() bool {
	// Check for Kubernetes service account
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}

	// Check for Kubernetes env vars
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	return false
}

// GetHardwareIDWithFallback tries multiple methods to get a hardware ID
func GetHardwareIDWithFallback() string {
	// Try Kubernetes first if applicable
	if IsKubernetes() {
		if id := KubernetesHardwareID(); id != "" {
			return id
		}
	}

	// Try standard hardware ID
	if id, err := GenerateHardwareID(); err == nil && id != "" {
		return id
	}

	// Fallback to hostname-based ID
	hostname, _ := os.Hostname()
	if hostname != "" {
		hash := sha256.Sum256([]byte(hostname))
		return "host-" + hex.EncodeToString(hash[:12])
	}

	// Last resort: random ID (will change on restart)
	return ""
}

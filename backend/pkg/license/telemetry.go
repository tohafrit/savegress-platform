package license

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// TelemetryEvent contains usage data
type TelemetryEvent struct {
	// Identification
	LicenseID  string `json:"license_id"`
	CustomerID string `json:"customer_id"`
	HardwareID string `json:"hardware_id"`
	InstanceID string `json:"instance_id,omitempty"`

	// Timing
	Timestamp   time.Time `json:"timestamp"`
	UptimeHours float64   `json:"uptime_hours"`

	// Usage metrics
	EventsProcessed int64 `json:"events_processed"`
	BytesProcessed  int64 `json:"bytes_processed"`
	TablesTracked   int   `json:"tables_tracked"`
	SourcesActive   int   `json:"sources_active"`

	// Performance
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	MaxLatencyMs    float64 `json:"max_latency_ms"`
	ErrorCount      int64   `json:"error_count"`
	RestartCount    int     `json:"restart_count"`

	// Environment
	Version    string `json:"version"`
	Platform   string `json:"platform"`
	GoVersion  string `json:"go_version"`
	SourceType string `json:"source_type"`

	// Features used
	FeaturesUsed []string `json:"features_used,omitempty"`
}

// TelemetryClient sends usage data to telemetry server
type TelemetryClient struct {
	mu         sync.Mutex
	baseURL    string
	httpClient *http.Client
	buffer     []TelemetryEvent
	maxBuffer  int
	enabled    bool
}

// NewTelemetryClient creates a new telemetry client
func NewTelemetryClient(baseURL string) *TelemetryClient {
	return &TelemetryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		buffer:    make([]TelemetryEvent, 0, 100),
		maxBuffer: 100,
		enabled:   true,
	}
}

// SetEnabled enables or disables telemetry
func (c *TelemetryClient) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// Send queues a telemetry event for sending
func (c *TelemetryClient) Send(event TelemetryEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return
	}

	c.buffer = append(c.buffer, event)

	// Flush if buffer is full
	if len(c.buffer) >= c.maxBuffer {
		go c.flush()
	}
}

// Flush sends all buffered events
func (c *TelemetryClient) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.flush()
}

func (c *TelemetryClient) flush() error {
	if len(c.buffer) == 0 {
		return nil
	}

	// Copy and clear buffer
	events := make([]TelemetryEvent, len(c.buffer))
	copy(events, c.buffer)
	c.buffer = c.buffer[:0]

	// Send events
	body, err := json.Marshal(events)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/telemetry", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Savegress-Engine/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Silently fail - telemetry should not impact operation
		return nil
	}
	defer resp.Body.Close()

	return nil
}

// UsageCollector collects usage metrics for telemetry
type UsageCollector struct {
	mu sync.RWMutex

	startTime       time.Time
	eventsProcessed int64
	bytesProcessed  int64
	tablesTracked   int
	sourcesActive   int
	errorCount      int64
	restartCount    int
	featuresUsed    map[string]bool

	// Latency tracking
	latencySum   float64
	latencyCount int64
	maxLatency   float64
}

// NewUsageCollector creates a new usage collector
func NewUsageCollector() *UsageCollector {
	return &UsageCollector{
		startTime:    time.Now(),
		featuresUsed: make(map[string]bool),
	}
}

// RecordEvent records a processed event
func (c *UsageCollector) RecordEvent(bytes int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventsProcessed++
	c.bytesProcessed += bytes
}

// RecordLatency records processing latency
func (c *UsageCollector) RecordLatency(latencyMs float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latencySum += latencyMs
	c.latencyCount++
	if latencyMs > c.maxLatency {
		c.maxLatency = latencyMs
	}
}

// RecordError records an error
func (c *UsageCollector) RecordError() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorCount++
}

// RecordFeatureUsed records that a feature was used
func (c *UsageCollector) RecordFeatureUsed(feature string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.featuresUsed[feature] = true
}

// SetSourcesActive sets the number of active sources
func (c *UsageCollector) SetSourcesActive(count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sourcesActive = count
}

// SetTablesTracked sets the number of tracked tables
func (c *UsageCollector) SetTablesTracked(count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tablesTracked = count
}

// IncrementRestarts increments the restart counter
func (c *UsageCollector) IncrementRestarts() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.restartCount++
}

// GetMetrics returns current metrics as a telemetry event
func (c *UsageCollector) GetMetrics() TelemetryEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var avgLatency float64
	if c.latencyCount > 0 {
		avgLatency = c.latencySum / float64(c.latencyCount)
	}

	features := make([]string, 0, len(c.featuresUsed))
	for f := range c.featuresUsed {
		features = append(features, f)
	}

	return TelemetryEvent{
		Timestamp:       time.Now(),
		UptimeHours:     time.Since(c.startTime).Hours(),
		EventsProcessed: c.eventsProcessed,
		BytesProcessed:  c.bytesProcessed,
		TablesTracked:   c.tablesTracked,
		SourcesActive:   c.sourcesActive,
		AvgLatencyMs:    avgLatency,
		MaxLatencyMs:    c.maxLatency,
		ErrorCount:      c.errorCount,
		RestartCount:    c.restartCount,
		FeaturesUsed:    features,
	}
}

// Reset resets the collector (typically called after sending telemetry)
func (c *UsageCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventsProcessed = 0
	c.bytesProcessed = 0
	c.errorCount = 0
	c.latencySum = 0
	c.latencyCount = 0
	c.maxLatency = 0
	// Keep featuresUsed, sourcesActive, tablesTracked
}

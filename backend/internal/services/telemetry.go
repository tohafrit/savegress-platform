package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

// TelemetryService handles usage telemetry from CDC engines
type TelemetryService struct {
	db    *repository.PostgresDB
	redis *repository.RedisClient
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(db *repository.PostgresDB, redis *repository.RedisClient) *TelemetryService {
	return &TelemetryService{db: db, redis: redis}
}

// TelemetryInput represents incoming telemetry data
type TelemetryInput struct {
	LicenseID       string  `json:"license_id"`
	HardwareID      string  `json:"hardware_id"`
	Timestamp       int64   `json:"timestamp"`
	EventsProcessed int64   `json:"events_processed"`
	BytesProcessed  int64   `json:"bytes_processed"`
	TablesTracked   int     `json:"tables_tracked"`
	SourcesActive   int     `json:"sources_active"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	ErrorCount      int64   `json:"error_count"`
	UptimeHours     float64 `json:"uptime_hours"`
	Version         string  `json:"version"`
	SourceType      string  `json:"source_type"`
}

// RecordTelemetry stores telemetry data
func (s *TelemetryService) RecordTelemetry(ctx context.Context, input TelemetryInput) error {
	licenseID, err := uuid.Parse(input.LicenseID)
	if err != nil {
		return fmt.Errorf("invalid license ID: %w", err)
	}

	record := &models.TelemetryRecord{
		ID:              uuid.New(),
		LicenseID:       licenseID,
		HardwareID:      input.HardwareID,
		Timestamp:       time.Unix(input.Timestamp, 0).UTC(),
		EventsProcessed: input.EventsProcessed,
		BytesProcessed:  input.BytesProcessed,
		TablesTracked:   input.TablesTracked,
		SourcesActive:   input.SourcesActive,
		AvgLatencyMs:    input.AvgLatencyMs,
		ErrorCount:      input.ErrorCount,
		UptimeHours:     input.UptimeHours,
		Version:         input.Version,
		SourceType:      input.SourceType,
	}

	// Store in database (sampled - keep 1 per hour)
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO telemetry (id, license_id, hardware_id, timestamp, events_processed, bytes_processed,
			tables_tracked, sources_active, avg_latency_ms, error_count, uptime_hours, version, source_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (license_id, hardware_id, date_trunc('hour', timestamp)) DO UPDATE SET
			events_processed = EXCLUDED.events_processed,
			bytes_processed = EXCLUDED.bytes_processed,
			avg_latency_ms = EXCLUDED.avg_latency_ms,
			error_count = EXCLUDED.error_count,
			uptime_hours = EXCLUDED.uptime_hours
	`, record.ID, record.LicenseID, record.HardwareID, record.Timestamp,
		record.EventsProcessed, record.BytesProcessed, record.TablesTracked,
		record.SourcesActive, record.AvgLatencyMs, record.ErrorCount,
		record.UptimeHours, record.Version, record.SourceType)
	if err != nil {
		return fmt.Errorf("failed to store telemetry: %w", err)
	}

	// Cache latest state in Redis for real-time dashboard
	state, _ := json.Marshal(input)
	key := fmt.Sprintf("telemetry:%s:%s", input.LicenseID, input.HardwareID)
	s.redis.Client().Set(ctx, key, state, 5*time.Minute)

	return nil
}

// DashboardStats holds aggregated stats for dashboard
type DashboardStats struct {
	TotalEventsProcessed int64   `json:"total_events_processed"`
	TotalBytesProcessed  int64   `json:"total_bytes_processed"`
	ActiveInstances      int     `json:"active_instances"`
	ActiveLicenses       int     `json:"active_licenses"`
	AvgLatencyMs         float64 `json:"avg_latency_ms"`
	TotalErrors          int64   `json:"total_errors"`
	TotalUptimeHours     float64 `json:"total_uptime_hours"`
}

// GetDashboardStats returns aggregated stats for a user
func (s *TelemetryService) GetDashboardStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	var stats DashboardStats

	// Get aggregated stats from last 24 hours
	err := s.db.Pool().QueryRow(ctx, `
		SELECT
			COALESCE(SUM(t.events_processed), 0),
			COALESCE(SUM(t.bytes_processed), 0),
			COUNT(DISTINCT t.hardware_id),
			COUNT(DISTINCT t.license_id),
			COALESCE(AVG(t.avg_latency_ms), 0),
			COALESCE(SUM(t.error_count), 0),
			COALESCE(SUM(t.uptime_hours), 0)
		FROM telemetry t
		JOIN licenses l ON t.license_id = l.id
		WHERE l.user_id = $1 AND t.timestamp > NOW() - INTERVAL '24 hours'
	`, userID).Scan(&stats.TotalEventsProcessed, &stats.TotalBytesProcessed,
		&stats.ActiveInstances, &stats.ActiveLicenses, &stats.AvgLatencyMs,
		&stats.TotalErrors, &stats.TotalUptimeHours)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// UsageDataPoint represents a time-series data point
type UsageDataPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	EventsProcessed int64     `json:"events_processed"`
	BytesProcessed  int64     `json:"bytes_processed"`
	AvgLatencyMs    float64   `json:"avg_latency_ms"`
	ErrorCount      int64     `json:"error_count"`
}

// GetUsageHistory returns usage time series for a user
func (s *TelemetryService) GetUsageHistory(ctx context.Context, userID uuid.UUID, days int) ([]UsageDataPoint, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT
			date_trunc('hour', t.timestamp) as hour,
			SUM(t.events_processed),
			SUM(t.bytes_processed),
			AVG(t.avg_latency_ms),
			SUM(t.error_count)
		FROM telemetry t
		JOIN licenses l ON t.license_id = l.id
		WHERE l.user_id = $1 AND t.timestamp > NOW() - ($2 || ' days')::INTERVAL
		GROUP BY hour
		ORDER BY hour
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []UsageDataPoint
	for rows.Next() {
		var p UsageDataPoint
		if err := rows.Scan(&p.Timestamp, &p.EventsProcessed, &p.BytesProcessed, &p.AvgLatencyMs, &p.ErrorCount); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// Instance represents an active CDC engine instance
type Instance struct {
	HardwareID      string    `json:"hardware_id"`
	Hostname        string    `json:"hostname"`
	LicenseID       string    `json:"license_id"`
	LicenseTier     string    `json:"license_tier"`
	Version         string    `json:"version"`
	SourceType      string    `json:"source_type"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	EventsProcessed int64     `json:"events_processed"`
	Status          string    `json:"status"` // online, offline
}

// GetActiveInstances returns active instances for a user
func (s *TelemetryService) GetActiveInstances(ctx context.Context, userID uuid.UUID) ([]Instance, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT DISTINCT ON (a.hardware_id)
			a.hardware_id,
			a.hostname,
			a.license_id,
			l.tier,
			a.version,
			t.source_type,
			a.last_seen_at,
			COALESCE(t.events_processed, 0)
		FROM license_activations a
		JOIN licenses l ON a.license_id = l.id
		LEFT JOIN LATERAL (
			SELECT source_type, events_processed
			FROM telemetry
			WHERE license_id = a.license_id AND hardware_id = a.hardware_id
			ORDER BY timestamp DESC LIMIT 1
		) t ON true
		WHERE l.user_id = $1 AND a.deactivated_at IS NULL
		ORDER BY a.hardware_id, a.last_seen_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]Instance, 0)
	now := time.Now()
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(&inst.HardwareID, &inst.Hostname, &inst.LicenseID,
			&inst.LicenseTier, &inst.Version, &inst.SourceType,
			&inst.LastSeenAt, &inst.EventsProcessed); err != nil {
			return nil, err
		}
		// Consider online if seen in last 5 minutes
		if now.Sub(inst.LastSeenAt) < 5*time.Minute {
			inst.Status = "online"
		} else {
			inst.Status = "offline"
		}
		instances = append(instances, inst)
	}
	return instances, nil
}

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Live(t *testing.T) {
	// Create handler with nil dependencies for liveness check
	handler := &HealthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", response.Status)
	}

	if response.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestHealthStatus_JSON(t *testing.T) {
	status := HealthStatus{
		Status:  "ok",
		Version: "1.0.0",
		Services: map[string]string{
			"postgres": "healthy",
			"redis":    "healthy",
		},
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded HealthStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != status.Status {
		t.Errorf("expected status %q, got %q", status.Status, decoded.Status)
	}

	if decoded.Version != status.Version {
		t.Errorf("expected version %q, got %q", status.Version, decoded.Version)
	}

	if len(decoded.Services) != len(status.Services) {
		t.Errorf("expected %d services, got %d", len(status.Services), len(decoded.Services))
	}
}

func TestDatabaseHealth_JSON(t *testing.T) {
	health := DatabaseHealth{
		Healthy:        true,
		ResponseTimeMs: 5,
		MaxConns:       100,
		IdleConns:      10,
		TotalConns:     25,
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DatabaseHealth
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Healthy != health.Healthy {
		t.Errorf("expected healthy %v, got %v", health.Healthy, decoded.Healthy)
	}

	if decoded.ResponseTimeMs != health.ResponseTimeMs {
		t.Errorf("expected response time %d, got %d", health.ResponseTimeMs, decoded.ResponseTimeMs)
	}
}

func TestCacheHealth_JSON(t *testing.T) {
	health := CacheHealth{
		Healthy:        true,
		ResponseTimeMs: 2,
		Connected:      true,
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CacheHealth
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Healthy != health.Healthy {
		t.Errorf("expected healthy %v, got %v", health.Healthy, decoded.Healthy)
	}

	if decoded.Connected != health.Connected {
		t.Errorf("expected connected %v, got %v", health.Connected, decoded.Connected)
	}
}

func TestCacheHealth_NilRedis(t *testing.T) {
	// Handler with nil redis
	handler := &HealthHandler{
		db:    nil,
		redis: nil,
	}

	// checkCacheHealth should handle nil redis gracefully
	health := handler.checkCacheHealth(nil)

	if health.Connected {
		t.Error("expected connected to be false when redis is nil")
	}

	if !health.Healthy {
		// With nil redis, healthy should still be true (cache is optional)
		t.Error("expected healthy to be true when redis is nil")
	}
}

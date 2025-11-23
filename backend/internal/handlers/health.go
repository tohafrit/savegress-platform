package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/savegress/platform/backend/internal/repository"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *repository.PostgresDB
	redis *repository.RedisClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *repository.PostgresDB, redis *repository.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HealthStatus represents health check response
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services,omitempty"`
}

// Live handles kubernetes liveness probe
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Version:   Version,
	})
}

// Ready handles kubernetes readiness probe
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	// Check PostgreSQL
	if err := h.db.Ping(ctx); err != nil {
		services["postgres"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		services["postgres"] = "healthy"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx); err != nil {
			services["redis"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			services["redis"] = "healthy"
		}
	}

	status := "ok"
	if !allHealthy {
		status = "degraded"
	}

	response := HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Version:   Version,
		Services:  services,
	}

	if allHealthy {
		respondSuccess(w, response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, response)
	}
}

// Detailed returns detailed health information (internal use only)
func (h *HealthHandler) Detailed(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	type DetailedHealth struct {
		HealthStatus
		Database DatabaseHealth `json:"database"`
		Cache    CacheHealth    `json:"cache"`
	}

	dbHealth := h.checkDatabaseHealth(ctx)
	cacheHealth := h.checkCacheHealth(ctx)

	allHealthy := dbHealth.Healthy && cacheHealth.Healthy

	status := "ok"
	if !allHealthy {
		status = "degraded"
	}

	respondSuccess(w, DetailedHealth{
		HealthStatus: HealthStatus{
			Status:    status,
			Timestamp: time.Now().UTC(),
			Version:   Version,
		},
		Database: dbHealth,
		Cache:    cacheHealth,
	})
}

// DatabaseHealth represents database health details
type DatabaseHealth struct {
	Healthy        bool          `json:"healthy"`
	ResponseTimeMs int64         `json:"response_time_ms"`
	MaxConns       int32         `json:"max_connections"`
	IdleConns      int32         `json:"idle_connections"`
	TotalConns     int32         `json:"total_connections"`
	Error          string        `json:"error,omitempty"`
}

func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) DatabaseHealth {
	start := time.Now()

	health := DatabaseHealth{
		Healthy: true,
	}

	if err := h.db.Ping(ctx); err != nil {
		health.Healthy = false
		health.Error = err.Error()
	}

	health.ResponseTimeMs = time.Since(start).Milliseconds()

	stats := h.db.Stat()
	health.MaxConns = stats.MaxConns()
	health.IdleConns = stats.IdleConns()
	health.TotalConns = stats.TotalConns()

	return health
}

// CacheHealth represents cache health details
type CacheHealth struct {
	Healthy        bool   `json:"healthy"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	Connected      bool   `json:"connected"`
	Error          string `json:"error,omitempty"`
}

func (h *HealthHandler) checkCacheHealth(ctx context.Context) CacheHealth {
	health := CacheHealth{
		Healthy: true,
	}

	if h.redis == nil {
		health.Connected = false
		return health
	}

	start := time.Now()

	if err := h.redis.Ping(ctx); err != nil {
		health.Healthy = false
		health.Connected = false
		health.Error = err.Error()
	} else {
		health.Connected = true
	}

	health.ResponseTimeMs = time.Since(start).Milliseconds()

	return health
}

// Version is set at build time
var Version = "dev"

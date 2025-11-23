package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// TelemetryHandler handles telemetry endpoints
type TelemetryHandler struct {
	telemetryService *services.TelemetryService
	licenseService   *services.LicenseService
}

// NewTelemetryHandler creates a new telemetry handler
func NewTelemetryHandler(telemetryService *services.TelemetryService, licenseService *services.LicenseService) *TelemetryHandler {
	return &TelemetryHandler{
		telemetryService: telemetryService,
		licenseService:   licenseService,
	}
}

// Receive accepts telemetry data from CDC engines
func (h *TelemetryHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var input services.TelemetryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate license exists
	_, err := h.licenseService.ValidateLicense(r.Context(), input.LicenseID, input.HardwareID)
	if err != nil {
		// Still accept telemetry but don't fail
		// This allows collecting data even for expired licenses
	}

	if err := h.telemetryService.RecordTelemetry(r.Context(), input); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to record telemetry")
		return
	}

	respondSuccess(w, map[string]string{"status": "recorded"})
}

// GetStats returns dashboard stats for user
func (h *TelemetryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	stats, err := h.telemetryService.GetDashboardStats(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	respondSuccess(w, stats)
}

// GetUsage returns usage history for user
func (h *TelemetryHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 7
	}

	usage, err := h.telemetryService.GetUsageHistory(r.Context(), userID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get usage")
		return
	}

	respondSuccess(w, map[string]interface{}{"usage": usage})
}

// GetInstances returns active instances for user
func (h *TelemetryHandler) GetInstances(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	instances, err := h.telemetryService.GetActiveInstances(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get instances")
		return
	}

	respondSuccess(w, map[string]interface{}{"instances": instances})
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// LicenseHandler handles license endpoints
type LicenseHandler struct {
	licenseService *services.LicenseService
	authService    *services.AuthService
}

// NewLicenseHandler creates a new license handler
func NewLicenseHandler(licenseService *services.LicenseService, authService *services.AuthService) *LicenseHandler {
	return &LicenseHandler{
		licenseService: licenseService,
		authService:    authService,
	}
}

// Validate handles license validation (called by CDC engines)
func (h *LicenseHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseID  string `json:"license_id"`
		HardwareID string `json:"hardware_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	license, err := h.licenseService.ValidateLicense(r.Context(), req.LicenseID, req.HardwareID)
	if err != nil {
		switch err {
		case services.ErrLicenseNotFound:
			respondError(w, http.StatusNotFound, "license not found")
		case services.ErrLicenseExpired:
			respondError(w, http.StatusForbidden, "license has expired")
		case services.ErrLicenseRevoked:
			respondError(w, http.StatusForbidden, "license has been revoked")
		case services.ErrHardwareMismatch:
			respondError(w, http.StatusForbidden, "license is bound to different hardware")
		default:
			respondError(w, http.StatusInternalServerError, "validation failed")
		}
		return
	}

	respondSuccess(w, map[string]interface{}{
		"valid":   true,
		"license": license,
	})
}

// Activate handles license activation
func (h *LicenseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseKey string `json:"license_key"`
		HardwareID string `json:"hardware_id"`
		Hostname   string `json:"hostname"`
		Platform   string `json:"platform"`
		Version    string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse license ID from key (simplified - in production parse JWT)
	// For now, assume license_key contains license ID
	licenseID, err := uuid.Parse(req.LicenseKey)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid license key")
		return
	}

	ipAddress := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ipAddress = xff
	}

	activation, err := h.licenseService.ActivateLicense(r.Context(), licenseID, req.HardwareID, req.Hostname, req.Platform, req.Version, ipAddress)
	if err != nil {
		switch err {
		case services.ErrActivationLimitReached:
			respondError(w, http.StatusForbidden, "activation limit reached")
		default:
			respondError(w, http.StatusInternalServerError, "activation failed: "+err.Error())
		}
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success":    true,
		"activation": activation,
	})
}

// Deactivate handles license deactivation
func (h *LicenseHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseID  string `json:"license_id"`
		HardwareID string `json:"hardware_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	licenseID, err := uuid.Parse(req.LicenseID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid license ID")
		return
	}

	if err := h.licenseService.DeactivateLicense(r.Context(), licenseID, req.HardwareID); err != nil {
		respondError(w, http.StatusInternalServerError, "deactivation failed")
		return
	}

	respondSuccess(w, map[string]string{"message": "license deactivated"})
}

// List returns user's licenses
func (h *LicenseHandler) List(w http.ResponseWriter, r *http.Request) {
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

	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get licenses")
		return
	}

	respondSuccess(w, map[string]interface{}{"licenses": licenses})
}

// Get returns a specific license
func (h *LicenseHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	licenseID := chi.URLParam(r, "id")
	license, err := h.licenseService.ValidateLicense(r.Context(), licenseID, "")
	if err != nil {
		respondError(w, http.StatusNotFound, "license not found")
		return
	}

	// Verify ownership
	if license.UserID != userID && claims.Role != "admin" {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	respondSuccess(w, license)
}

// Create creates a new license (typically called after payment)
func (h *LicenseHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Tier       string `json:"tier"`
		ValidDays  int    `json:"valid_days"`
		HardwareID string `json:"hardware_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ValidDays == 0 {
		req.ValidDays = 365
	}

	license, err := h.licenseService.CreateLicense(r.Context(), userID, req.Tier, req.ValidDays, req.HardwareID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create license: "+err.Error())
		return
	}

	respondCreated(w, license)
}

// Revoke revokes a license
func (h *LicenseHandler) Revoke(w http.ResponseWriter, r *http.Request) {
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

	licenseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid license ID")
		return
	}

	// Verify ownership (or admin)
	license, err := h.licenseService.ValidateLicense(r.Context(), licenseID.String(), "")
	if err != nil {
		respondError(w, http.StatusNotFound, "license not found")
		return
	}
	if license.UserID != userID && claims.Role != "admin" {
		respondError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.licenseService.RevokeLicense(r.Context(), licenseID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to revoke license")
		return
	}

	respondSuccess(w, map[string]string{"message": "license revoked"})
}

// GetActivations returns activations for a license
func (h *LicenseHandler) GetActivations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	licenseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid license ID")
		return
	}

	activations, err := h.licenseService.GetLicenseActivations(r.Context(), licenseID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get activations")
		return
	}

	respondSuccess(w, map[string]interface{}{"activations": activations})
}

// ListAll returns all licenses (admin only) with pagination
func (h *LicenseHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Admin only
	if claims.Role != "admin" {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := parseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse optional filters
	tier := r.URL.Query().Get("tier")
	status := r.URL.Query().Get("status")

	// Get paginated licenses
	licenses, total, err := h.licenseService.GetAllLicensesPaginated(r.Context(), page, limit, tier, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get licenses")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"licenses": licenses,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// parseInt is a helper to parse string to int
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// AdminGenerate generates a license (admin only)
func (h *LicenseHandler) AdminGenerate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     string `json:"user_id"`
		Tier       string `json:"tier"`
		ValidDays  int    `json:"valid_days"`
		HardwareID string `json:"hardware_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if req.ValidDays == 0 {
		req.ValidDays = 365
	}

	license, err := h.licenseService.CreateLicense(r.Context(), userID, req.Tier, req.ValidDays, req.HardwareID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create license: "+err.Error())
		return
	}

	respondCreated(w, license)
}

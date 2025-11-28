package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// PersonalizedDownloadHandler handles personalized download requests
type PersonalizedDownloadHandler struct {
	downloadService *services.DownloadService
	licenseService  *services.LicenseService
}

// NewPersonalizedDownloadHandler creates a new handler
func NewPersonalizedDownloadHandler(downloadService *services.DownloadService, licenseService *services.LicenseService) *PersonalizedDownloadHandler {
	return &PersonalizedDownloadHandler{
		downloadService: downloadService,
		licenseService:  licenseService,
	}
}

// DownloadPersonalized streams a binary with embedded license key
// This endpoint streams the binary directly to the user with their license embedded
func (h *PersonalizedDownloadHandler) DownloadPersonalized(w http.ResponseWriter, r *http.Request) {
	// Get user from context
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

	// Get parameters
	product := chi.URLParam(r, "product")
	version := chi.URLParam(r, "version")
	platform := chi.URLParam(r, "platform")
	edition := chi.URLParam(r, "edition")

	// Validate parameters
	if product == "" || platform == "" {
		respondError(w, http.StatusBadRequest, "product and platform are required")
		return
	}

	if version == "" || version == "latest" {
		version = "1.0.0" // TODO: get latest version from service
	}

	if edition == "" {
		edition = "community"
	}

	// Validate request
	if err := h.downloadService.ValidateDownloadRequest(product, edition, platform); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user's active license
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err != nil || len(licenses) == 0 {
		respondError(w, http.StatusForbidden, "no active license found - please subscribe first")
		return
	}

	// Find active license
	var activeLicense *models.License
	for i, lic := range licenses {
		if lic.Status == "active" {
			activeLicense = &licenses[i]
			break
		}
	}

	if activeLicense == nil {
		respondError(w, http.StatusForbidden, "no active license found")
		return
	}

	// Check if license tier allows this edition
	if err := h.downloadService.CheckLicenseForEdition(activeLicense.Tier, edition); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}

	// Get personalized binary with embedded license
	binary, filename, err := h.downloadService.GetPersonalizedBinary(
		r.Context(),
		product,
		version,
		edition,
		platform,
		activeLicense.LicenseKey,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to prepare download: %v", err))
		return
	}

	// Set headers for binary download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binary)))
	w.Header().Set("X-License-Embedded", "true")
	w.Header().Set("X-License-Tier", activeLicense.Tier)

	// Stream the binary
	w.WriteHeader(http.StatusOK)
	w.Write(binary)
}

// GetDownloadInfo returns download information without streaming the binary
// This is useful for showing download options with proper license checks
func (h *PersonalizedDownloadHandler) GetDownloadInfo(w http.ResponseWriter, r *http.Request) {
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

	// Get user's license tier
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get license info")
		return
	}

	var activeLicense *models.License
	for i, lic := range licenses {
		if lic.Status == "active" {
			activeLicense = &licenses[i]
			break
		}
	}

	// Return available downloads based on license
	type DownloadOption struct {
		Product   string   `json:"product"`
		Version   string   `json:"version"`
		Editions  []string `json:"available_editions"`
		Platforms []string `json:"platforms"`
		Tier      string   `json:"license_tier"`
	}

	availableEditions := []string{"community"}
	tier := "community"

	if activeLicense != nil {
		tier = activeLicense.Tier
		switch tier {
		case "enterprise":
			availableEditions = []string{"community", "pro", "enterprise"}
		case "pro":
			availableEditions = []string{"community", "pro"}
		}
	}

	downloads := []DownloadOption{
		{
			Product:   "cdc-engine",
			Version:   "1.0.0",
			Editions:  availableEditions,
			Platforms: []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
			Tier:      tier,
		},
	}

	respondSuccess(w, map[string]interface{}{
		"downloads":       downloads,
		"license_tier":    tier,
		"license_embedded": true,
		"message":         "Downloads include your license key - no manual configuration needed",
	})
}

// GenerateInstallScript generates a personalized install script for the user
func (h *PersonalizedDownloadHandler) GenerateInstallScript(w http.ResponseWriter, r *http.Request) {
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

	// Get user's license
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err != nil || len(licenses) == 0 {
		respondError(w, http.StatusForbidden, "no active license found")
		return
	}

	var activeLicense *models.License
	for i, lic := range licenses {
		if lic.Status == "active" {
			activeLicense = &licenses[i]
			break
		}
	}

	if activeLicense == nil {
		respondError(w, http.StatusForbidden, "no active license found")
		return
	}

	// Generate one-time download token
	downloadToken := uuid.New().String()
	// TODO: Store token in Redis with expiry and link to userID

	// Truncate license key for display
	keyDisplay := activeLicense.LicenseKey
	if len(keyDisplay) > 20 {
		keyDisplay = keyDisplay[:20] + "..."
	}

	// Generate script
	script := fmt.Sprintf(`#!/bin/bash
# Savegress CDC Engine - One-line installer
# License: %s (%s)
# Generated for: %s

set -e

VERSION="1.0.0"
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
DOWNLOAD_TOKEN="%s"

echo "Installing Savegress CDC Engine..."

# Download personalized binary (includes your license)
curl -fsSL "https://api.savegress.io/api/v1/downloads/install/$DOWNLOAD_TOKEN/$PLATFORM" -o cdc-engine

# Make executable
chmod +x cdc-engine

# Move to path (requires sudo)
if [ -w /usr/local/bin ]; then
    mv cdc-engine /usr/local/bin/
else
    sudo mv cdc-engine /usr/local/bin/
fi

echo "✓ Savegress CDC Engine installed successfully!"
echo "  Your license is already embedded - no configuration needed."
echo ""
echo "Get started:"
echo "  cdc-engine --help"
echo "  cdc-engine init"
`,
		activeLicense.Tier,
		keyDisplay,
		claims.Email,
		downloadToken,
	)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=install-savegress.sh")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(script))
}

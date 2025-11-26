package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// ConfigHandler handles configuration generation endpoints
type ConfigHandler struct {
	configService  *services.ConfigGeneratorService
	licenseService *services.LicenseService
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configService *services.ConfigGeneratorService, licenseService *services.LicenseService) *ConfigHandler {
	return &ConfigHandler{
		configService:  configService,
		licenseService: licenseService,
	}
}

// Generate generates a deployment configuration
func (h *ConfigHandler) Generate(w http.ResponseWriter, r *http.Request) {
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

	// Get query parameters
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "docker-compose"
	}

	pipelineIDStr := r.URL.Query().Get("pipeline_id")
	var pipelineID *uuid.UUID
	if pipelineIDStr != "" {
		id, err := uuid.Parse(pipelineIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pipeline_id")
			return
		}
		pipelineID = &id
	}

	// Get user's license key
	licenseKey := "YOUR_LICENSE_KEY"
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err == nil && len(licenses) > 0 {
		// Use the first active license
		for _, lic := range licenses {
			if lic.Status == "active" {
				licenseKey = lic.LicenseKey
				break
			}
		}
	}

	config, err := h.configService.GenerateConfig(r.Context(), userID, format, pipelineID, licenseKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate config: "+err.Error())
		return
	}

	// Set content type based on format
	contentType := "text/plain"
	filename := "savegress-config"
	switch format {
	case "docker-compose", "docker":
		contentType = "text/yaml"
		filename = "docker-compose.yml"
	case "helm", "kubernetes", "k8s":
		contentType = "text/yaml"
		filename = "values.yaml"
	case "env", "dotenv":
		contentType = "text/plain"
		filename = "savegress.env"
	case "systemd":
		contentType = "text/plain"
		filename = "savegress.service"
	}

	// Check if download is requested
	if r.URL.Query().Get("download") == "true" {
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(config))
}

// GetQuickStart returns a quick start guide
func (h *ConfigHandler) GetQuickStart(w http.ResponseWriter, r *http.Request) {
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

	sourceType := r.URL.Query().Get("source_type")
	if sourceType == "" {
		sourceType = "postgres"
	}

	// Get user's license key
	licenseKey := "YOUR_LICENSE_KEY"
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
	if err == nil && len(licenses) > 0 {
		for _, lic := range licenses {
			if lic.Status == "active" {
				licenseKey = lic.LicenseKey
				break
			}
		}
	}

	guide := h.configService.GenerateQuickStart(licenseKey, sourceType)

	w.Header().Set("Content-Type", "text/markdown")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(guide))
}

// GetFormats returns available configuration formats
func (h *ConfigHandler) GetFormats(w http.ResponseWriter, r *http.Request) {
	formats := []map[string]interface{}{
		{
			"id":          "docker-compose",
			"name":        "Docker Compose",
			"description": "Docker Compose configuration for local or server deployment",
			"filename":    "docker-compose.yml",
			"icon":        "docker",
		},
		{
			"id":          "helm",
			"name":        "Kubernetes (Helm)",
			"description": "Helm values file for Kubernetes deployment",
			"filename":    "values.yaml",
			"icon":        "kubernetes",
		},
		{
			"id":          "env",
			"name":        "Environment File",
			"description": "Environment variables file for binary or container",
			"filename":    "savegress.env",
			"icon":        "file-text",
		},
		{
			"id":          "systemd",
			"name":        "Systemd Service",
			"description": "Systemd unit file for Linux servers",
			"filename":    "savegress.service",
			"icon":        "server",
		},
	}

	respondSuccess(w, map[string]interface{}{"formats": formats})
}

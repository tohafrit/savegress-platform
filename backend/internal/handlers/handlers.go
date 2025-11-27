package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/savegress/platform/backend/internal/services"
)

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondSuccess(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusOK, data)
}

func respondCreated(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusCreated, data)
}

// DownloadHandler handles download endpoints
type DownloadHandler struct {
	storageService *services.StorageService
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler(storageService *services.StorageService) *DownloadHandler {
	return &DownloadHandler{storageService: storageService}
}

// Available releases configuration
var availableReleases = []ReleaseInfo{
	{
		Product:   "cdc-engine",
		Version:   "1.0.0",
		Editions:  []string{"community", "pro", "enterprise"},
		Platforms: []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
	},
}

// ReleaseInfo contains information about an available release
type ReleaseInfo struct {
	Product   string   `json:"product"`
	Version   string   `json:"version"`
	Editions  []string `json:"editions"`
	Platforms []string `json:"platforms"`
}

// ListDownloads returns available downloads
func (h *DownloadHandler) ListDownloads(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, map[string]interface{}{"downloads": availableReleases})
}

// GetDownloadURL generates a pre-signed URL for downloading a release
func (h *DownloadHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	product := chi.URLParam(r, "product")
	version := chi.URLParam(r, "version")
	platform := chi.URLParam(r, "platform")
	edition := chi.URLParam(r, "edition")

	// Default values
	if product == "" {
		product = r.URL.Query().Get("product")
	}
	if version == "" {
		version = r.URL.Query().Get("version")
	}
	if platform == "" {
		platform = r.URL.Query().Get("platform")
	}
	if edition == "" {
		edition = r.URL.Query().Get("edition")
	}

	// Validate required parameters
	if product == "" || version == "" || platform == "" {
		respondError(w, http.StatusBadRequest, "product, version, and platform are required")
		return
	}

	// Default edition to community
	if edition == "" {
		edition = "community"
	}

	// Validate edition
	validEditions := map[string]bool{"community": true, "pro": true, "enterprise": true}
	if !validEditions[edition] {
		respondError(w, http.StatusBadRequest, "invalid edition: must be community, pro, or enterprise")
		return
	}

	// Validate platform
	validPlatforms := map[string]bool{
		"linux-amd64":   true,
		"linux-arm64":   true,
		"darwin-amd64":  true,
		"darwin-arm64":  true,
		"windows-amd64": true,
	}
	if !validPlatforms[platform] {
		respondError(w, http.StatusBadRequest, "invalid platform")
		return
	}

	// Check if storage service is configured
	if h.storageService == nil {
		// Fallback for development - return a placeholder URL
		respondSuccess(w, map[string]interface{}{
			"url":        "https://releases.savegress.io/" + services.GetReleaseKey(product, version, platform, edition),
			"expires_in": 3600,
			"message":    "S3 storage not configured - using placeholder URL",
		})
		return
	}

	// Generate the S3 key for this release
	key := services.GetReleaseKey(product, version, platform, edition)

	// Generate pre-signed URL
	downloadInfo, err := h.storageService.GenerateDownloadURL(r.Context(), key)
	if err != nil {
		respondError(w, http.StatusNotFound, "release not found or not available")
		return
	}

	respondSuccess(w, downloadInfo)
}

// Legacy handlers for backwards compatibility
func ListDownloads(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, map[string]interface{}{"downloads": availableReleases})
}

func GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	// Legacy endpoint - redirect to use the handler
	respondError(w, http.StatusBadRequest, "please use /api/v1/downloads/{product}/{version}/{platform}/{edition} endpoint")
}

package handlers

import (
	"encoding/json"
	"net/http"
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

// Placeholder handlers for downloads
func ListDownloads(w http.ResponseWriter, r *http.Request) {
	downloads := []map[string]interface{}{
		{
			"product": "cdc-engine",
			"version": "1.0.0",
			"editions": []string{"community", "pro", "enterprise"},
			"platforms": []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64", "windows-amd64"},
		},
	}
	respondSuccess(w, map[string]interface{}{"downloads": downloads})
}

func GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	// TODO: Generate signed S3 URL for download
	respondSuccess(w, map[string]string{
		"url": "https://releases.savegress.io/cdc-engine/1.0.0/cdc-engine-linux-amd64",
		"expires_in": "3600",
	})
}

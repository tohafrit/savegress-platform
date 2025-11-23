package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/savegress/platform/backend/internal/services"
)

// EarlyAccessHandler handles early access requests
type EarlyAccessHandler struct {
	earlyAccessService *services.EarlyAccessService
	turnstileSecret    string
}

// NewEarlyAccessHandler creates a new early access handler
func NewEarlyAccessHandler(earlyAccessService *services.EarlyAccessService, turnstileSecret string) *EarlyAccessHandler {
	return &EarlyAccessHandler{
		earlyAccessService: earlyAccessService,
		turnstileSecret:    turnstileSecret,
	}
}

// EarlyAccessRequest represents the early access form data
type EarlyAccessRequest struct {
	Email           string `json:"email"`
	Company         string `json:"company"`
	CurrentSolution string `json:"currentSolution,omitempty"`
	DataVolume      string `json:"dataVolume,omitempty"`
	Message         string `json:"message,omitempty"`
	TurnstileToken  string `json:"turnstileToken"`
}

// Submit handles early access form submissions
func (h *EarlyAccessHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req EarlyAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Company == "" {
		respondError(w, http.StatusBadRequest, "email and company are required")
		return
	}

	if req.TurnstileToken == "" {
		respondError(w, http.StatusBadRequest, "captcha verification required")
		return
	}

	// Verify Turnstile token
	if !h.verifyTurnstile(req.TurnstileToken, getClientIP(r)) {
		respondError(w, http.StatusBadRequest, "captcha verification failed")
		return
	}

	// Get client info
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Save to database
	err := h.earlyAccessService.Submit(r.Context(), services.EarlyAccessInput{
		Email:           req.Email,
		Company:         req.Company,
		CurrentSolution: req.CurrentSolution,
		DataVolume:      req.DataVolume,
		Message:         req.Message,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to submit request")
		return
	}

	respondCreated(w, map[string]interface{}{
		"success": true,
		"message": "Request submitted successfully",
	})
}

func (h *EarlyAccessHandler) verifyTurnstile(token, ip string) bool {
	// Skip verification for test keys
	if len(h.turnstileSecret) > 6 && h.turnstileSecret[:6] == "1x0000" {
		return true
	}

	data := url.Values{}
	data.Set("secret", h.turnstileSecret)
	data.Set("response", token)
	if ip != "" {
		data.Set("remoteip", ip)
	}

	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", data)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Success
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

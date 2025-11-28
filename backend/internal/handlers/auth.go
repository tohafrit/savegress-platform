package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/savegress/platform/backend/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService    *services.AuthService
	emailService   *services.EmailService
	licenseService *services.LicenseService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, emailService *services.EmailService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
	}
}

// SetLicenseService sets the license service for CLI login
func (h *AuthHandler) SetLicenseService(licenseService *services.LicenseService) {
	h.licenseService = licenseService
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Company  string `json:"company"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validation
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, tokens, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name, req.Company)
	if err != nil {
		if err == services.ErrUserExists {
			respondError(w, http.StatusConflict, "user with this email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	respondCreated(w, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "login failed")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"tokens": tokens,
	})
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Create reset token (this returns empty string if user doesn't exist, for security)
	token, err := h.authService.CreatePasswordResetToken(r.Context(), req.Email)
	if err != nil {
		// Log error but don't expose to user
		log.Printf("Error creating password reset token: %v", err)
	}

	// Send email if token was created and email service is configured
	if token != "" && h.emailService != nil {
		if err := h.emailService.SendPasswordResetEmail(r.Context(), req.Email, token); err != nil {
			log.Printf("Error sending password reset email: %v", err)
		}
	}

	// Always return success (don't reveal if email exists)
	respondSuccess(w, map[string]string{
		"message": "if an account exists with this email, a reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	if req.Password == "" || len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Reset the password
	err := h.authService.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			respondError(w, http.StatusBadRequest, "invalid or expired reset token")
		case services.ErrResetTokenExpired:
			respondError(w, http.StatusBadRequest, "reset token has expired")
		case services.ErrResetTokenUsed:
			respondError(w, http.StatusBadRequest, "reset token has already been used")
		default:
			respondError(w, http.StatusInternalServerError, "failed to reset password")
		}
		return
	}

	respondSuccess(w, map[string]string{
		"message": "password has been reset successfully",
	})
}

// CLILogin handles login from CLI and returns the user's license key
func (h *AuthHandler) CLILogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "login failed")
		return
	}

	// Get user's active license
	var licenseKey string
	var edition string

	if h.licenseService != nil {
		licenses, err := h.licenseService.GetUserLicenses(r.Context(), user.ID)
		if err == nil {
			for _, lic := range licenses {
				if lic.Status == "active" {
					licenseKey = lic.LicenseKey
					edition = lic.Tier
					break
				}
			}
		}
	}

	if licenseKey == "" {
		respondError(w, http.StatusForbidden, "no active license found - please subscribe at https://savegress.io")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success":     true,
		"license_key": licenseKey,
		"edition":     edition,
		"token":       tokens.AccessToken,
		"message":     "Login successful",
	})
}

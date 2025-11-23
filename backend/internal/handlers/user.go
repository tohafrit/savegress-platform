package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// UserHandler handles user endpoints
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile returns current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondSuccess(w, user)
}

// UpdateProfile updates current user's profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name    string `json:"name"`
		Company string `json:"company"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.UpdateProfile(r.Context(), claims.UserID, req.Name, req.Company); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	respondSuccess(w, map[string]string{"message": "profile updated"})
}

// ChangePassword changes current user's password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	if err := h.userService.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == services.ErrInvalidCredentials {
			respondError(w, http.StatusBadRequest, "current password is incorrect")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	respondSuccess(w, map[string]string{"message": "password changed"})
}

// ListUsers returns list of users (admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 20
	}

	users, total, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"users": users,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// GetUser returns a specific user (admin only)
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondSuccess(w, user)
}

// UpdateUser updates a user (admin only)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondSuccess(w, map[string]string{"message": "user updated"})
}

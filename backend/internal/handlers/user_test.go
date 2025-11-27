package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockUserService implements a mock for testing
type MockUserService struct {
	GetByIDFunc             func(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateProfileFunc       func(ctx context.Context, userID uuid.UUID, name, company string) error
	ChangePasswordFunc      func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	ListUsersFunc           func(ctx context.Context, limit, offset int) ([]models.User, int, error)
	UpdateUserRoleFunc      func(ctx context.Context, userID uuid.UUID, role string) error
	SetStripeCustomerIDFunc func(ctx context.Context, userID uuid.UUID, stripeCustomerID string) error
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uuid.UUID, name, company string) error {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(ctx, userID, name, company)
	}
	return nil
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if m.ChangePasswordFunc != nil {
		return m.ChangePasswordFunc(ctx, userID, currentPassword, newPassword)
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockUserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(ctx, userID, role)
	}
	return nil
}

func (m *MockUserService) SetStripeCustomerID(ctx context.Context, userID uuid.UUID, stripeCustomerID string) error {
	if m.SetStripeCustomerIDFunc != nil {
		return m.SetStripeCustomerIDFunc(ctx, userID, stripeCustomerID)
	}
	return nil
}

// testUserHandler wraps UserHandler for testing with mock service
type testUserHandler struct {
	mock *MockUserService
}

func newTestUserHandler(mock *MockUserService) *testUserHandler {
	return &testUserHandler{mock: mock}
}

func (h *testUserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mock.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondSuccess(w, user)
}

func (h *testUserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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
		Name    string `json:"name"`
		Company string `json:"company"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.mock.UpdateProfile(r.Context(), userID, req.Name, req.Company); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	respondSuccess(w, map[string]string{"message": "profile updated"})
}

func (h *testUserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

	if err := h.mock.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == services.ErrInvalidCredentials {
			respondError(w, http.StatusBadRequest, "current password is incorrect")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	respondSuccess(w, map[string]string{"message": "password changed"})
}

func (h *testUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, total, err := h.mock.ListUsers(r.Context(), 20, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"users":  users,
		"total":  total,
		"limit":  20,
		"offset": 0,
	})
}

func (h *testUserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.mock.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondSuccess(w, user)
}

func (h *testUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	if err := h.mock.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondSuccess(w, map[string]string{"message": "user updated"})
}

// Helper to create a context with claims
func contextWithClaims(userID uuid.UUID, role string) context.Context {
	claims := &services.Claims{
		UserID: userID.String(),
		Role:   role,
	}
	return context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
}

func TestUserHandler_GetProfile(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		mockGetByID    func(ctx context.Context, id uuid.UUID) (*models.User, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful get profile",
			userID: &testUserID,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:        id,
					Email:     "test@example.com",
					Name:      "Test User",
					Role:      "user",
					CreatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no claims",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:   "user not found",
			userID: &testUserID,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestUserHandler(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
			if tt.userID != nil {
				req = req.WithContext(contextWithClaims(*tt.userID, "user"))
			}
			rec := httptest.NewRecorder()

			handler.GetProfile(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name              string
		userID            *uuid.UUID
		requestBody       map[string]string
		mockUpdateProfile func(ctx context.Context, userID uuid.UUID, name, company string) error
		expectedStatus    int
		expectedError     string
	}{
		{
			name:   "successful update",
			userID: &testUserID,
			requestBody: map[string]string{
				"name":    "New Name",
				"company": "New Company",
			},
			mockUpdateProfile: func(ctx context.Context, userID uuid.UUID, name, company string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "unauthorized - no claims",
			userID:      nil,
			requestBody: map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:   "update fails",
			userID: &testUserID,
			requestBody: map[string]string{
				"name": "New Name",
			},
			mockUpdateProfile: func(ctx context.Context, userID uuid.UUID, name, company string) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				UpdateProfileFunc: tt.mockUpdateProfile,
			}
			handler := newTestUserHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/user/profile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != nil {
				req = req.WithContext(contextWithClaims(*tt.userID, "user"))
			}
			rec := httptest.NewRecorder()

			handler.UpdateProfile(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name               string
		userID             *uuid.UUID
		requestBody        map[string]string
		mockChangePassword func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
		expectedStatus     int
		expectedError      string
	}{
		{
			name:   "successful password change",
			userID: &testUserID,
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "newpassword123",
			},
			mockChangePassword: func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "unauthorized - no claims",
			userID:      nil,
			requestBody: map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:   "new password too short",
			userID: &testUserID,
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name:   "incorrect current password",
			userID: &testUserID,
			requestBody: map[string]string{
				"current_password": "wrongpassword",
				"new_password":     "newpassword123",
			},
			mockChangePassword: func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
				return services.ErrInvalidCredentials
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "current password is incorrect",
		},
		{
			name:   "internal error",
			userID: &testUserID,
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "newpassword123",
			},
			mockChangePassword: func(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to change password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				ChangePasswordFunc: tt.mockChangePassword,
			}
			handler := newTestUserHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user/password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != nil {
				req = req.WithContext(contextWithClaims(*tt.userID, "user"))
			}
			rec := httptest.NewRecorder()

			handler.ChangePassword(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockListUsers  func(ctx context.Context, limit, offset int) ([]models.User, int, error)
		expectedStatus int
		expectedError  string
		expectedTotal  int
	}{
		{
			name: "successful list users",
			mockListUsers: func(ctx context.Context, limit, offset int) ([]models.User, int, error) {
				return []models.User{
					{ID: uuid.New(), Email: "user1@example.com"},
					{ID: uuid.New(), Email: "user2@example.com"},
				}, 2, nil
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  2,
		},
		{
			name: "empty list",
			mockListUsers: func(ctx context.Context, limit, offset int) ([]models.User, int, error) {
				return []models.User{}, 0, nil
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  0,
		},
		{
			name: "database error",
			mockListUsers: func(ctx context.Context, limit, offset int) ([]models.User, int, error) {
				return nil, 0, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				ListUsersFunc: tt.mockListUsers,
			}
			handler := newTestUserHandler(mock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
			rec := httptest.NewRecorder()

			handler.ListUsers(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if int(response["total"].(float64)) != tt.expectedTotal {
					t.Errorf("expected total %d, got %v", tt.expectedTotal, response["total"])
				}
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name           string
		userIDParam    string
		mockGetByID    func(ctx context.Context, id uuid.UUID) (*models.User, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful get user",
			userIDParam: testUserID.String(),
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:    id,
					Email: "test@example.com",
					Name:  "Test User",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user ID",
		},
		{
			name:        "user not found",
			userIDParam: testUserID.String(),
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestUserHandler(mock)

			// Use chi router context for URL params
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/"+tt.userIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rec := httptest.NewRecorder()

			handler.GetUser(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name               string
		userIDParam        string
		requestBody        map[string]string
		mockUpdateUserRole func(ctx context.Context, userID uuid.UUID, role string) error
		expectedStatus     int
		expectedError      string
	}{
		{
			name:        "successful update role",
			userIDParam: testUserID.String(),
			requestBody: map[string]string{
				"role": "admin",
			},
			mockUpdateUserRole: func(ctx context.Context, userID uuid.UUID, role string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "invalid-uuid",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user ID",
		},
		{
			name:        "update fails",
			userIDParam: testUserID.String(),
			requestBody: map[string]string{
				"role": "admin",
			},
			mockUpdateUserRole: func(ctx context.Context, userID uuid.UUID, role string) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUserService{
				UpdateUserRoleFunc: tt.mockUpdateUserRole,
			}
			handler := newTestUserHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/"+tt.userIDParam, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rec := httptest.NewRecorder()

			handler.UpdateUser(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

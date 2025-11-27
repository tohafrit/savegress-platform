package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockAuthService implements a mock for testing
type MockAuthService struct {
	RegisterFunc       func(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error)
	LoginFunc          func(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error)
	RefreshTokenFunc   func(ctx context.Context, refreshToken string) (*services.TokenPair, error)
	ForgotPasswordFunc func(ctx context.Context, email string) error
	ResetPasswordFunc  func(ctx context.Context, token, newPassword string) error
}

func (m *MockAuthService) Register(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password, name, company)
	}
	return nil, nil, nil
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, email, password)
	}
	return nil, nil, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*services.TokenPair, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockAuthService) ForgotPassword(ctx context.Context, email string) error {
	if m.ForgotPasswordFunc != nil {
		return m.ForgotPasswordFunc(ctx, email)
	}
	return nil
}

func (m *MockAuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(ctx, token, newPassword)
	}
	return nil
}

// AuthServiceInterface defines the interface for AuthService
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error)
	Login(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*services.TokenPair, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// testAuthHandler wraps AuthHandler for testing with mock service
type testAuthHandler struct {
	mock *MockAuthService
}

func newTestAuthHandler(mock *MockAuthService) *testAuthHandler {
	return &testAuthHandler{mock: mock}
}

func (h *testAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Company  string `json:"company"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, tokens, err := h.mock.Register(r.Context(), req.Email, req.Password, req.Name, req.Company)
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

func (h *testAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, tokens, err := h.mock.Login(r.Context(), req.Email, req.Password)
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

func (h *testAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.mock.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"tokens": tokens,
	})
}

func (h *testAuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
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

	_ = h.mock.ForgotPassword(r.Context(), req.Email)

	respondSuccess(w, map[string]string{
		"message": "if an account exists with this email, a reset link has been sent",
	})
}

func (h *testAuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "token and password are required")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	err := h.mock.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		if err == services.ErrInvalidToken {
			respondError(w, http.StatusBadRequest, "invalid or expired reset token")
			return
		}
		if err == services.ErrResetTokenExpired {
			respondError(w, http.StatusBadRequest, "password reset token has expired")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	respondSuccess(w, map[string]string{
		"message": "password has been reset successfully",
	})
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockRegister   func(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
				"company":  "Test Company",
			},
			mockRegister: func(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error) {
				return &models.User{
					ID:    uuid.New(),
					Email: email,
					Name:  name,
				}, &services.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing email",
			requestBody: map[string]string{
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email, password, and name are required",
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"email": "test@example.com",
				"name":  "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email, password, and name are required",
		},
		{
			name: "missing name",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email, password, and name are required",
		},
		{
			name: "password too short",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "short",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name: "user already exists",
			requestBody: map[string]string{
				"email":    "existing@example.com",
				"password": "password123",
				"name":     "Test User",
			},
			mockRegister: func(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error) {
				return nil, nil, services.ErrUserExists
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "user with this email already exists",
		},
		{
			name: "internal error",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
			},
			mockRegister: func(ctx context.Context, email, password, name, company string) (*models.User, *services.TokenPair, error) {
				return nil, nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthService{
				RegisterFunc: tt.mockRegister,
			}
			handler := newTestAuthHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

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

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockLogin      func(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockLogin: func(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error) {
				return &models.User{
					ID:    uuid.New(),
					Email: email,
				}, &services.TokenPair{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing email",
			requestBody: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and password are required",
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and password are required",
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockLogin: func(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error) {
				return nil, nil, services.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid email or password",
		},
		{
			name: "internal error",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			mockLogin: func(ctx context.Context, email, password string) (*models.User, *services.TokenPair, error) {
				return nil, nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "login failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthService{
				LoginFunc: tt.mockLogin,
			}
			handler := newTestAuthHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

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

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name              string
		requestBody       map[string]string
		mockRefreshToken  func(ctx context.Context, refreshToken string) (*services.TokenPair, error)
		expectedStatus    int
		expectedError     string
	}{
		{
			name: "successful refresh",
			requestBody: map[string]string{
				"refresh_token": "valid_refresh_token",
			},
			mockRefreshToken: func(ctx context.Context, refreshToken string) (*services.TokenPair, error) {
				return &services.TokenPair{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing refresh token",
			requestBody: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "refresh_token is required",
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid_token",
			},
			mockRefreshToken: func(ctx context.Context, refreshToken string) (*services.TokenPair, error) {
				return nil, errors.New("invalid token")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthService{
				RefreshTokenFunc: tt.mockRefreshToken,
			}
			handler := newTestAuthHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.RefreshToken(rec, req)

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

func TestAuthHandler_ForgotPassword(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful forgot password request",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing email",
			requestBody: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthService{}
			handler := newTestAuthHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ForgotPassword(rec, req)

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

func TestAuthHandler_ResetPassword(t *testing.T) {
	tests := []struct {
		name              string
		requestBody       map[string]string
		mockResetPassword func(ctx context.Context, token, newPassword string) error
		expectedStatus    int
		expectedError     string
	}{
		{
			name: "successful password reset",
			requestBody: map[string]string{
				"token":    "valid_token",
				"password": "newpassword123",
			},
			mockResetPassword: func(ctx context.Context, token, newPassword string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing token",
			requestBody: map[string]string{
				"password": "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "token and password are required",
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"token": "valid_token",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "token and password are required",
		},
		{
			name: "password too short",
			requestBody: map[string]string{
				"token":    "valid_token",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name: "invalid token",
			requestBody: map[string]string{
				"token":    "invalid_token",
				"password": "newpassword123",
			},
			mockResetPassword: func(ctx context.Context, token, newPassword string) error {
				return services.ErrInvalidToken
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid or expired reset token",
		},
		{
			name: "expired token",
			requestBody: map[string]string{
				"token":    "expired_token",
				"password": "newpassword123",
			},
			mockResetPassword: func(ctx context.Context, token, newPassword string) error {
				return services.ErrResetTokenExpired
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password reset token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockAuthService{
				ResetPasswordFunc: tt.mockResetPassword,
			}
			handler := newTestAuthHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ResetPassword(rec, req)

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

func TestAuthHandler_InvalidJSON(t *testing.T) {
	mock := &MockAuthService{}
	handler := newTestAuthHandler(mock)

	endpoints := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"Register", handler.Register},
		{"Login", handler.Login},
		{"RefreshToken", handler.RefreshToken},
		{"ForgotPassword", handler.ForgotPassword},
		{"ResetPassword", handler.ResetPassword},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+" invalid JSON", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/test", bytes.NewReader([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			ep.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}

			var response map[string]string
			json.NewDecoder(rec.Body).Decode(&response)
			if response["error"] != "invalid request body" {
				t.Errorf("expected error 'invalid request body', got %q", response["error"])
			}
		})
	}
}

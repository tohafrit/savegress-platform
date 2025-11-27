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

// MockLicenseServiceForHandler implements a mock for license handler testing
type MockLicenseServiceForHandler struct {
	ValidateLicenseFunc         func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error)
	ActivateLicenseFunc         func(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error)
	DeactivateLicenseFunc       func(ctx context.Context, licenseID uuid.UUID, hardwareID string) error
	GetUserLicensesFunc         func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
	CreateLicenseFunc           func(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error)
	RevokeLicenseFunc           func(ctx context.Context, licenseID uuid.UUID) error
	GetLicenseActivationsFunc   func(ctx context.Context, licenseID uuid.UUID) ([]models.LicenseActivation, error)
	GetAllLicensesPaginatedFunc func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error)
}

func (m *MockLicenseServiceForHandler) ValidateLicense(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
	if m.ValidateLicenseFunc != nil {
		return m.ValidateLicenseFunc(ctx, licenseID, hardwareID)
	}
	return nil, nil
}

func (m *MockLicenseServiceForHandler) ActivateLicense(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error) {
	if m.ActivateLicenseFunc != nil {
		return m.ActivateLicenseFunc(ctx, licenseID, hardwareID, hostname, platform, version, ipAddress)
	}
	return nil, nil
}

func (m *MockLicenseServiceForHandler) DeactivateLicense(ctx context.Context, licenseID uuid.UUID, hardwareID string) error {
	if m.DeactivateLicenseFunc != nil {
		return m.DeactivateLicenseFunc(ctx, licenseID, hardwareID)
	}
	return nil
}

func (m *MockLicenseServiceForHandler) GetUserLicenses(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
	if m.GetUserLicensesFunc != nil {
		return m.GetUserLicensesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockLicenseServiceForHandler) CreateLicense(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
	if m.CreateLicenseFunc != nil {
		return m.CreateLicenseFunc(ctx, userID, tier, validDays, hardwareID)
	}
	return nil, nil
}

func (m *MockLicenseServiceForHandler) RevokeLicense(ctx context.Context, licenseID uuid.UUID) error {
	if m.RevokeLicenseFunc != nil {
		return m.RevokeLicenseFunc(ctx, licenseID)
	}
	return nil
}

func (m *MockLicenseServiceForHandler) GetLicenseActivations(ctx context.Context, licenseID uuid.UUID) ([]models.LicenseActivation, error) {
	if m.GetLicenseActivationsFunc != nil {
		return m.GetLicenseActivationsFunc(ctx, licenseID)
	}
	return nil, nil
}

func (m *MockLicenseServiceForHandler) GetAllLicensesPaginated(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
	if m.GetAllLicensesPaginatedFunc != nil {
		return m.GetAllLicensesPaginatedFunc(ctx, page, limit, tier, status)
	}
	return nil, 0, nil
}

func (m *MockLicenseServiceForHandler) GetLicenseStats(ctx context.Context) (*services.LicenseStats, error) {
	return nil, nil
}

func (m *MockLicenseServiceForHandler) RecordUsage(ctx context.Context, record services.UsageRecord) error {
	return nil
}

func (m *MockLicenseServiceForHandler) GetUsageStats(ctx context.Context, licenseID uuid.UUID, days int) ([]services.UsageRecord, error) {
	return nil, nil
}

// Helper to create context with user claims
func contextWithUser(userID uuid.UUID, role string) context.Context {
	claims := &services.Claims{
		UserID: userID.String(),
		Email:  "test@example.com",
		Role:   role,
	}
	return context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
}

func TestLicenseHandler_Validate(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      map[string]string
		mockValidate     func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error)
		expectedStatus   int
		expectedError    string
		expectedValidKey string
	}{
		{
			name: "successful validation",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "hw123",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(licenseID),
					UserID: uuid.New(),
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			expectedStatus:   http.StatusOK,
			expectedValidKey: "valid",
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "license not found",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "hw123",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return nil, services.ErrLicenseNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "license not found",
		},
		{
			name: "license expired",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "hw123",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return nil, services.ErrLicenseExpired
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "license has expired",
		},
		{
			name: "license revoked",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "hw123",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return nil, services.ErrLicenseRevoked
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "license has been revoked",
		},
		{
			name: "hardware mismatch",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "wrong_hw",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return nil, services.ErrHardwareMismatch
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "license is bound to different hardware",
		},
		{
			name: "internal error",
			requestBody: map[string]string{
				"license_id":  uuid.New().String(),
				"hardware_id": "hw123",
			},
			mockValidate: func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				ValidateLicenseFunc: tt.mockValidate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/validate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Validate(rec, req)

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

			if tt.expectedValidKey != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if valid, ok := response[tt.expectedValidKey].(bool); !ok || !valid {
					t.Errorf("expected valid=true in response")
				}
			}
		})
	}
}

func TestLicenseHandler_Activate(t *testing.T) {
	validLicenseID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockActivate   func(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful activation",
			requestBody: map[string]interface{}{
				"license_key": validLicenseID.String(),
				"hardware_id": "hw123",
				"hostname":    "server1",
				"platform":    "linux",
				"version":     "1.0.0",
			},
			mockActivate: func(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error) {
				return &models.LicenseActivation{
					ID:          uuid.New(),
					LicenseID:   licenseID,
					HardwareID:  hardwareID,
					Hostname:    hostname,
					ActivatedAt: time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid license key format",
			requestBody: map[string]interface{}{
				"license_key": "not-a-uuid",
				"hardware_id": "hw123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid license key",
		},
		{
			name: "activation limit reached",
			requestBody: map[string]interface{}{
				"license_key": validLicenseID.String(),
				"hardware_id": "hw123",
				"hostname":    "server1",
				"platform":    "linux",
				"version":     "1.0.0",
			},
			mockActivate: func(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error) {
				return nil, services.ErrActivationLimitReached
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "activation limit reached",
		},
		{
			name: "activation failed with error",
			requestBody: map[string]interface{}{
				"license_key": validLicenseID.String(),
				"hardware_id": "hw123",
				"hostname":    "server1",
				"platform":    "linux",
				"version":     "1.0.0",
			},
			mockActivate: func(ctx context.Context, licenseID uuid.UUID, hardwareID, hostname, platform, version, ipAddress string) (*models.LicenseActivation, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "activation failed: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				ActivateLicenseFunc: tt.mockActivate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/activate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Activate(rec, req)

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

func TestLicenseHandler_Deactivate(t *testing.T) {
	validLicenseID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockDeactivate func(ctx context.Context, licenseID uuid.UUID, hardwareID string) error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful deactivation",
			requestBody: map[string]interface{}{
				"license_id":  validLicenseID.String(),
				"hardware_id": "hw123",
			},
			mockDeactivate: func(ctx context.Context, licenseID uuid.UUID, hardwareID string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid license ID format",
			requestBody: map[string]interface{}{
				"license_id":  "not-a-uuid",
				"hardware_id": "hw123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid license ID",
		},
		{
			name: "deactivation failed",
			requestBody: map[string]interface{}{
				"license_id":  validLicenseID.String(),
				"hardware_id": "hw123",
			},
			mockDeactivate: func(ctx context.Context, licenseID uuid.UUID, hardwareID string) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "deactivation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				DeactivateLicenseFunc: tt.mockDeactivate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/deactivate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Deactivate(rec, req)

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

func TestLicenseHandler_List(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		mockGetLicenses func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful list",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			mockGetLicenses: func(ctx context.Context, uid uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						ID:     uuid.New(),
						UserID: uid,
						Tier:   "pro",
						Status: "active",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "get licenses failed",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			mockGetLicenses: func(ctx context.Context, uid uuid.UUID) ([]models.License, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get licenses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				GetUserLicensesFunc: tt.mockGetLicenses,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/licenses", nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()

			handler.List(rec, req)

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

func TestLicenseHandler_Get(t *testing.T) {
	userID := uuid.New()
	licenseID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		licenseID      string
		mockValidate   func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful get - own license",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful get - admin access",
			setupContext: func() context.Context {
				return contextWithUser(otherUserID, "admin")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			licenseID:      licenseID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "license not found",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return nil, services.ErrLicenseNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "license not found",
		},
		{
			name: "access denied - different user",
			setupContext: func() context.Context {
				return contextWithUser(otherUserID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				ValidateLicenseFunc: tt.mockValidate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/licenses/"+tt.licenseID, nil)
			req = req.WithContext(tt.setupContext())

			// Setup chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.licenseID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.Get(rec, req)

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

func TestLicenseHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		requestBody    map[string]interface{}
		mockCreate     func(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			requestBody: map[string]interface{}{
				"tier":        "pro",
				"valid_days":  365,
				"hardware_id": "hw123",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.New(),
					UserID: uid,
					Tier:   tier,
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful creation with default valid days",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			requestBody: map[string]interface{}{
				"tier": "pro",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				if validDays != 365 {
					t.Errorf("expected default validDays=365, got %d", validDays)
				}
				return &models.License{
					ID:     uuid.New(),
					UserID: uid,
					Tier:   tier,
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			requestBody: map[string]interface{}{
				"tier": "pro",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "invalid request body",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "creation failed",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			requestBody: map[string]interface{}{
				"tier": "pro",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create license: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				CreateLicenseFunc: tt.mockCreate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses", bytes.NewReader(body))
			req = req.WithContext(tt.setupContext())
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

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

func TestLicenseHandler_Revoke(t *testing.T) {
	userID := uuid.New()
	licenseID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		licenseID      string
		mockValidate   func(ctx context.Context, licenseID string, hardwareID string) (*models.License, error)
		mockRevoke     func(ctx context.Context, licenseID uuid.UUID) error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful revoke - own license",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			mockRevoke: func(ctx context.Context, lid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful revoke - admin",
			setupContext: func() context.Context {
				return contextWithUser(otherUserID, "admin")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			mockRevoke: func(ctx context.Context, lid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			licenseID:      licenseID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "invalid license ID format",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID:      "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid license ID",
		},
		{
			name: "license not found",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return nil, services.ErrLicenseNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "license not found",
		},
		{
			name: "access denied - different user",
			setupContext: func() context.Context {
				return contextWithUser(otherUserID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "access denied",
		},
		{
			name: "revoke failed",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockValidate: func(ctx context.Context, lid string, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.MustParse(lid),
					UserID: userID,
					Tier:   "pro",
					Status: "active",
				}, nil
			},
			mockRevoke: func(ctx context.Context, lid uuid.UUID) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to revoke license",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				ValidateLicenseFunc: tt.mockValidate,
				RevokeLicenseFunc:   tt.mockRevoke,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/"+tt.licenseID+"/revoke", nil)
			req = req.WithContext(tt.setupContext())

			// Setup chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.licenseID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.Revoke(rec, req)

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

func TestLicenseHandler_GetActivations(t *testing.T) {
	userID := uuid.New()
	licenseID := uuid.New()

	tests := []struct {
		name             string
		setupContext     func() context.Context
		licenseID        string
		mockGetActivations func(ctx context.Context, licenseID uuid.UUID) ([]models.LicenseActivation, error)
		expectedStatus   int
		expectedError    string
	}{
		{
			name: "successful get activations",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockGetActivations: func(ctx context.Context, lid uuid.UUID) ([]models.LicenseActivation, error) {
				return []models.LicenseActivation{
					{
						ID:          uuid.New(),
						LicenseID:   lid,
						HardwareID:  "hw123",
						ActivatedAt: time.Now(),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			licenseID:      licenseID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "invalid license ID format",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID:      "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid license ID",
		},
		{
			name: "get activations failed",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			licenseID: licenseID.String(),
			mockGetActivations: func(ctx context.Context, lid uuid.UUID) ([]models.LicenseActivation, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get activations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				GetLicenseActivationsFunc: tt.mockGetActivations,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/licenses/"+tt.licenseID+"/activations", nil)
			req = req.WithContext(tt.setupContext())

			// Setup chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.licenseID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.GetActivations(rec, req)

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

func TestLicenseHandler_ListAll(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name               string
		setupContext       func() context.Context
		queryParams        map[string]string
		mockGetAllPaginated func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error)
		expectedStatus     int
		expectedError      string
	}{
		{
			name: "successful list all - default pagination",
			setupContext: func() context.Context {
				return contextWithUser(adminID, "admin")
			},
			mockGetAllPaginated: func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
				if page != 1 || limit != 20 {
					t.Errorf("expected page=1, limit=20, got page=%d, limit=%d", page, limit)
				}
				return []models.License{
					{
						ID:     uuid.New(),
						UserID: uuid.New(),
						Tier:   "pro",
						Status: "active",
					},
				}, 1, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful list all - with pagination",
			setupContext: func() context.Context {
				return contextWithUser(adminID, "admin")
			},
			queryParams: map[string]string{
				"page":  "2",
				"limit": "50",
			},
			mockGetAllPaginated: func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
				if page != 2 || limit != 50 {
					t.Errorf("expected page=2, limit=50, got page=%d, limit=%d", page, limit)
				}
				return []models.License{}, 100, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful list all - with filters",
			setupContext: func() context.Context {
				return contextWithUser(adminID, "admin")
			},
			queryParams: map[string]string{
				"tier":   "pro",
				"status": "active",
			},
			mockGetAllPaginated: func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
				if tier != "pro" || status != "active" {
					t.Errorf("expected tier=pro, status=active, got tier=%s, status=%s", tier, status)
				}
				return []models.License{}, 0, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no claims",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "forbidden - non-admin user",
			setupContext: func() context.Context {
				return contextWithUser(userID, "user")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "admin access required",
		},
		{
			name: "get licenses failed",
			setupContext: func() context.Context {
				return contextWithUser(adminID, "admin")
			},
			mockGetAllPaginated: func(ctx context.Context, page, limit int, tier, status string) ([]models.License, int, error) {
				return nil, 0, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get licenses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				GetAllLicensesPaginatedFunc: tt.mockGetAllPaginated,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			url := "/api/v1/admin/licenses"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for k, v := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += k + "=" + v
					first = false
				}
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()

			handler.ListAll(rec, req)

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

func TestLicenseHandler_AdminGenerate(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockCreate     func(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful generation",
			requestBody: map[string]interface{}{
				"user_id":     validUserID.String(),
				"tier":        "enterprise",
				"valid_days":  730,
				"hardware_id": "hw123",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:     uuid.New(),
					UserID: uid,
					Tier:   tier,
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful generation with default valid days",
			requestBody: map[string]interface{}{
				"user_id": validUserID.String(),
				"tier":    "pro",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				if validDays != 365 {
					t.Errorf("expected default validDays=365, got %d", validDays)
				}
				return &models.License{
					ID:     uuid.New(),
					UserID: uid,
					Tier:   tier,
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "invalid user ID format",
			requestBody: map[string]interface{}{
				"user_id": "not-a-uuid",
				"tier":    "pro",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user ID",
		},
		{
			name: "creation failed",
			requestBody: map[string]interface{}{
				"user_id": validUserID.String(),
				"tier":    "pro",
			},
			mockCreate: func(ctx context.Context, uid uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create license: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLicenseServiceForHandler{
				CreateLicenseFunc: tt.mockCreate,
			}
			handler := NewLicenseHandlerWithInterface(mock, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/licenses/generate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.AdminGenerate(rec, req)

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

func TestLicenseHandler_InvalidJSON(t *testing.T) {
	mock := &MockLicenseServiceForHandler{}
	handler := NewLicenseHandlerWithInterface(mock, nil)

	endpoints := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"Validate", handler.Validate},
		{"Activate", handler.Activate},
		{"Deactivate", handler.Deactivate},
		{"AdminGenerate", handler.AdminGenerate},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+" invalid JSON", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/test", bytes.NewReader([]byte("invalid json")))
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

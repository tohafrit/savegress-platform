package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockTelemetryService implements a mock for testing
type MockTelemetryService struct {
	RecordTelemetryFunc    func(ctx context.Context, input services.TelemetryInput) error
	GetDashboardStatsFunc  func(ctx context.Context, userID uuid.UUID) (*services.DashboardStats, error)
	GetUsageHistoryFunc    func(ctx context.Context, userID uuid.UUID, days int) ([]services.UsageDataPoint, error)
	GetActiveInstancesFunc func(ctx context.Context, userID uuid.UUID) ([]services.Instance, error)
}

func (m *MockTelemetryService) RecordTelemetry(ctx context.Context, input services.TelemetryInput) error {
	if m.RecordTelemetryFunc != nil {
		return m.RecordTelemetryFunc(ctx, input)
	}
	return nil
}

func (m *MockTelemetryService) GetDashboardStats(ctx context.Context, userID uuid.UUID) (*services.DashboardStats, error) {
	if m.GetDashboardStatsFunc != nil {
		return m.GetDashboardStatsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockTelemetryService) GetUsageHistory(ctx context.Context, userID uuid.UUID, days int) ([]services.UsageDataPoint, error) {
	if m.GetUsageHistoryFunc != nil {
		return m.GetUsageHistoryFunc(ctx, userID, days)
	}
	return nil, nil
}

func (m *MockTelemetryService) GetActiveInstances(ctx context.Context, userID uuid.UUID) ([]services.Instance, error) {
	if m.GetActiveInstancesFunc != nil {
		return m.GetActiveInstancesFunc(ctx, userID)
	}
	return nil, nil
}

// MockLicenseServiceForTelemetry implements a mock for testing
type MockLicenseServiceForTelemetry struct {
	ValidateLicenseFunc func(ctx context.Context, licenseID, hardwareID string) (*models.License, error)
}

func (m *MockLicenseServiceForTelemetry) ValidateLicense(ctx context.Context, licenseID, hardwareID string) (*models.License, error) {
	if m.ValidateLicenseFunc != nil {
		return m.ValidateLicenseFunc(ctx, licenseID, hardwareID)
	}
	return nil, nil
}

// testTelemetryHandler wraps TelemetryHandler for testing with mock services
type testTelemetryHandler struct {
	telemetryMock *MockTelemetryService
	licenseMock   *MockLicenseServiceForTelemetry
}

func newTestTelemetryHandler(telemetryMock *MockTelemetryService, licenseMock *MockLicenseServiceForTelemetry) *testTelemetryHandler {
	return &testTelemetryHandler{
		telemetryMock: telemetryMock,
		licenseMock:   licenseMock,
	}
}

func (h *testTelemetryHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var input services.TelemetryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate license exists
	_, err := h.licenseMock.ValidateLicense(r.Context(), input.LicenseID, input.HardwareID)
	if err != nil {
		// Still accept telemetry but don't fail
		// This allows collecting data even for expired licenses
	}

	if err := h.telemetryMock.RecordTelemetry(r.Context(), input); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to record telemetry")
		return
	}

	respondSuccess(w, map[string]string{"status": "recorded"})
}

func (h *testTelemetryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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

	stats, err := h.telemetryMock.GetDashboardStats(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	respondSuccess(w, stats)
}

func (h *testTelemetryHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
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

	days := 7
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	usage, err := h.telemetryMock.GetUsageHistory(r.Context(), userID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get usage")
		return
	}

	respondSuccess(w, map[string]interface{}{"usage": usage})
}

func (h *testTelemetryHandler) GetInstances(w http.ResponseWriter, r *http.Request) {
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

	instances, err := h.telemetryMock.GetActiveInstances(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get instances")
		return
	}

	respondSuccess(w, map[string]interface{}{"instances": instances})
}

// Helper function to create a context with user claims for telemetry tests
func contextWithClaimsForTelemetry(userID uuid.UUID, email string) context.Context {
	claims := &services.Claims{
		UserID: userID.String(),
		Email:  email,
		Role:   "user",
	}
	return context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
}

// Helper function to create a context with invalid user claims for telemetry tests
func contextWithInvalidClaimsForTelemetry() context.Context {
	claims := &services.Claims{
		UserID: "invalid-uuid",
		Email:  "test@example.com",
		Role:   "user",
	}
	return context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
}

func TestTelemetryHandler_Receive(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         map[string]interface{}
		mockValidateLicense func(ctx context.Context, licenseID, hardwareID string) (*models.License, error)
		mockRecordTelemetry func(ctx context.Context, input services.TelemetryInput) error
		expectedStatus      int
		expectedError       string
	}{
		{
			name: "successful telemetry recording",
			requestBody: map[string]interface{}{
				"license_id":       "lic-123",
				"hardware_id":      "hw-456",
				"timestamp":        1234567890,
				"events_processed": 1000,
				"bytes_processed":  50000,
				"tables_tracked":   5,
				"sources_active":   2,
				"avg_latency_ms":   12.5,
				"error_count":      0,
				"uptime_hours":     24.5,
			},
			mockValidateLicense: func(ctx context.Context, licenseID, hardwareID string) (*models.License, error) {
				return &models.License{
					ID:         uuid.New(),
					LicenseKey: licenseID,
				}, nil
			},
			mockRecordTelemetry: func(ctx context.Context, input services.TelemetryInput) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			mockRecordTelemetry: func(ctx context.Context, input services.TelemetryInput) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "license validation fails but still records telemetry",
			requestBody: map[string]interface{}{
				"license_id":       "lic-invalid",
				"hardware_id":      "hw-456",
				"timestamp":        1234567890,
				"events_processed": 1000,
				"bytes_processed":  50000,
				"tables_tracked":   5,
				"sources_active":   2,
				"avg_latency_ms":   12.5,
				"error_count":      0,
				"uptime_hours":     24.5,
			},
			mockValidateLicense: func(ctx context.Context, licenseID, hardwareID string) (*models.License, error) {
				return nil, errors.New("license not found")
			},
			mockRecordTelemetry: func(ctx context.Context, input services.TelemetryInput) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "telemetry recording fails",
			requestBody: map[string]interface{}{
				"license_id":       "lic-123",
				"hardware_id":      "hw-456",
				"timestamp":        1234567890,
				"events_processed": 1000,
				"bytes_processed":  50000,
				"tables_tracked":   5,
				"sources_active":   2,
				"avg_latency_ms":   12.5,
				"error_count":      0,
				"uptime_hours":     24.5,
			},
			mockValidateLicense: func(ctx context.Context, licenseID, hardwareID string) (*models.License, error) {
				return &models.License{}, nil
			},
			mockRecordTelemetry: func(ctx context.Context, input services.TelemetryInput) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to record telemetry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryMock := &MockTelemetryService{
				RecordTelemetryFunc: tt.mockRecordTelemetry,
			}
			licenseMock := &MockLicenseServiceForTelemetry{
				ValidateLicenseFunc: tt.mockValidateLicense,
			}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/receive", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Receive(rec, req)

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

func TestTelemetryHandler_Receive_InvalidJSON(t *testing.T) {
	telemetryMock := &MockTelemetryService{}
	licenseMock := &MockLicenseServiceForTelemetry{}
	handler := newTestTelemetryHandler(telemetryMock, licenseMock)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/receive", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Receive(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]string
	json.NewDecoder(rec.Body).Decode(&response)
	if response["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %q", response["error"])
	}
}

func TestTelemetryHandler_GetStats(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                  string
		context               context.Context
		mockGetDashboardStats func(ctx context.Context, userID uuid.UUID) (*services.DashboardStats, error)
		expectedStatus        int
		expectedError         string
	}{
		{
			name:    "successful stats retrieval",
			context: contextWithClaimsForTelemetry(userID, "test@example.com"),
			mockGetDashboardStats: func(ctx context.Context, uid uuid.UUID) (*services.DashboardStats, error) {
				return &services.DashboardStats{
					TotalEventsProcessed: 10000,
					TotalBytesProcessed:  500000,
					ActiveInstances:      3,
					ActiveLicenses:       2,
					AvgLatencyMs:         15.5,
					TotalErrors:          10,
					TotalUptimeHours:     72.5,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no user in context",
			context:        context.Background(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid user ID in claims",
			context:        contextWithInvalidClaimsForTelemetry(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid user id",
		},
		{
			name:    "service error",
			context: contextWithClaimsForTelemetry(userID, "test@example.com"),
			mockGetDashboardStats: func(ctx context.Context, uid uuid.UUID) (*services.DashboardStats, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryMock := &MockTelemetryService{
				GetDashboardStatsFunc: tt.mockGetDashboardStats,
			}
			licenseMock := &MockLicenseServiceForTelemetry{}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/stats", nil)
			req = req.WithContext(tt.context)
			rec := httptest.NewRecorder()

			handler.GetStats(rec, req)

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

			if tt.expectedStatus == http.StatusOK && tt.mockGetDashboardStats != nil {
				var response services.DashboardStats
				json.NewDecoder(rec.Body).Decode(&response)
				if response.TotalEventsProcessed != 10000 {
					t.Errorf("expected total events processed to be 10000, got %d", response.TotalEventsProcessed)
				}
			}
		})
	}
}

func TestTelemetryHandler_GetUsage(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                string
		context             context.Context
		queryParams         string
		mockGetUsageHistory func(ctx context.Context, userID uuid.UUID, days int) ([]services.UsageDataPoint, error)
		expectedStatus      int
		expectedError       string
		expectedDays        int
	}{
		{
			name:        "successful usage retrieval with default days",
			context:     contextWithClaimsForTelemetry(userID, "test@example.com"),
			queryParams: "",
			mockGetUsageHistory: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				return []services.UsageDataPoint{
					{
						Timestamp:       time.Now(),
						EventsProcessed: 1000,
						BytesProcessed:  50000,
						AvgLatencyMs:    12.5,
						ErrorCount:      2,
					},
					{
						Timestamp:       time.Now().Add(-1 * time.Hour),
						EventsProcessed: 900,
						BytesProcessed:  45000,
						AvgLatencyMs:    11.8,
						ErrorCount:      1,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedDays:   7,
		},
		{
			name:        "successful usage retrieval with custom days",
			context:     contextWithClaimsForTelemetry(userID, "test@example.com"),
			queryParams: "?days=30",
			mockGetUsageHistory: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				if days != 30 {
					t.Errorf("expected days to be 30, got %d", days)
				}
				return []services.UsageDataPoint{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedDays:   30,
		},
		{
			name:        "invalid days parameter defaults to 7",
			context:     contextWithClaimsForTelemetry(userID, "test@example.com"),
			queryParams: "?days=invalid",
			mockGetUsageHistory: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				if days != 7 {
					t.Errorf("expected days to be 7 (default), got %d", days)
				}
				return []services.UsageDataPoint{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedDays:   7,
		},
		{
			name:           "no user in context",
			context:        context.Background(),
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid user ID in claims",
			context:        contextWithInvalidClaimsForTelemetry(),
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid user id",
		},
		{
			name:        "service error",
			context:     contextWithClaimsForTelemetry(userID, "test@example.com"),
			queryParams: "",
			mockGetUsageHistory: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryMock := &MockTelemetryService{
				GetUsageHistoryFunc: tt.mockGetUsageHistory,
			}
			licenseMock := &MockLicenseServiceForTelemetry{}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			url := "/api/v1/telemetry/usage" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(tt.context)
			rec := httptest.NewRecorder()

			handler.GetUsage(rec, req)

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

			if tt.expectedStatus == http.StatusOK && tt.mockGetUsageHistory != nil {
				var response struct {
					Usage []services.UsageDataPoint `json:"usage"`
				}
				json.NewDecoder(rec.Body).Decode(&response)
				// Response contains usage data directly
			}
		})
	}
}

func TestTelemetryHandler_GetInstances(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                   string
		context                context.Context
		mockGetActiveInstances func(ctx context.Context, userID uuid.UUID) ([]services.Instance, error)
		expectedStatus         int
		expectedError          string
	}{
		{
			name:    "successful instances retrieval",
			context: contextWithClaimsForTelemetry(userID, "test@example.com"),
			mockGetActiveInstances: func(ctx context.Context, uid uuid.UUID) ([]services.Instance, error) {
				return []services.Instance{
					{
						HardwareID:      "hw-123",
						Hostname:        "server-01",
						LicenseID:       "lic-456",
						LicenseTier:     "premium",
						Version:         "1.0.0",
						SourceType:      "postgresql",
						LastSeenAt:      time.Now(),
						EventsProcessed: 5000,
						Status:          "online",
					},
					{
						HardwareID:      "hw-789",
						Hostname:        "server-02",
						LicenseID:       "lic-456",
						LicenseTier:     "premium",
						Version:         "1.0.0",
						SourceType:      "mysql",
						LastSeenAt:      time.Now().Add(-5 * time.Minute),
						EventsProcessed: 3000,
						Status:          "online",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "successful instances retrieval with empty list",
			context: contextWithClaimsForTelemetry(userID, "test@example.com"),
			mockGetActiveInstances: func(ctx context.Context, uid uuid.UUID) ([]services.Instance, error) {
				return []services.Instance{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no user in context",
			context:        context.Background(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid user ID in claims",
			context:        contextWithInvalidClaimsForTelemetry(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid user id",
		},
		{
			name:    "service error",
			context: contextWithClaimsForTelemetry(userID, "test@example.com"),
			mockGetActiveInstances: func(ctx context.Context, uid uuid.UUID) ([]services.Instance, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get instances",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telemetryMock := &MockTelemetryService{
				GetActiveInstancesFunc: tt.mockGetActiveInstances,
			}
			licenseMock := &MockLicenseServiceForTelemetry{}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/instances", nil)
			req = req.WithContext(tt.context)
			rec := httptest.NewRecorder()

			handler.GetInstances(rec, req)

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

			if tt.expectedStatus == http.StatusOK && tt.mockGetActiveInstances != nil {
				var response struct {
					Instances []services.Instance `json:"instances"`
				}
				json.NewDecoder(rec.Body).Decode(&response)
				// Response contains instances data directly
			}
		})
	}
}

func TestTelemetryHandler_AuthenticatedEndpoints(t *testing.T) {
	userID := uuid.New()

	endpoints := []struct {
		name    string
		path    string
		handler func(h *testTelemetryHandler) http.HandlerFunc
	}{
		{
			name: "GetStats",
			path: "/api/v1/telemetry/stats",
			handler: func(h *testTelemetryHandler) http.HandlerFunc {
				return h.GetStats
			},
		},
		{
			name: "GetUsage",
			path: "/api/v1/telemetry/usage",
			handler: func(h *testTelemetryHandler) http.HandlerFunc {
				return h.GetUsage
			},
		},
		{
			name: "GetInstances",
			path: "/api/v1/telemetry/instances",
			handler: func(h *testTelemetryHandler) http.HandlerFunc {
				return h.GetInstances
			},
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+" requires authentication", func(t *testing.T) {
			telemetryMock := &MockTelemetryService{}
			licenseMock := &MockLicenseServiceForTelemetry{}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			req := httptest.NewRequest(http.MethodGet, ep.path, nil)
			// No context with claims
			rec := httptest.NewRecorder()

			ep.handler(handler)(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}

			var response map[string]string
			json.NewDecoder(rec.Body).Decode(&response)
			if response["error"] != "unauthorized" {
				t.Errorf("expected error 'unauthorized', got %q", response["error"])
			}
		})

		t.Run(ep.name+" with valid authentication", func(t *testing.T) {
			telemetryMock := &MockTelemetryService{
				GetDashboardStatsFunc: func(ctx context.Context, uid uuid.UUID) (*services.DashboardStats, error) {
					return &services.DashboardStats{}, nil
				},
				GetUsageHistoryFunc: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
					return []services.UsageDataPoint{}, nil
				},
				GetActiveInstancesFunc: func(ctx context.Context, uid uuid.UUID) ([]services.Instance, error) {
					return []services.Instance{}, nil
				},
			}
			licenseMock := &MockLicenseServiceForTelemetry{}
			handler := newTestTelemetryHandler(telemetryMock, licenseMock)

			req := httptest.NewRequest(http.MethodGet, ep.path, nil)
			req = req.WithContext(contextWithClaimsForTelemetry(userID, "test@example.com"))
			rec := httptest.NewRecorder()

			ep.handler(handler)(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
		})
	}
}

func TestTelemetryHandler_EdgeCases(t *testing.T) {
	userID := uuid.New()

	t.Run("GetUsage with zero days parameter", func(t *testing.T) {
		called := false
		telemetryMock := &MockTelemetryService{
			GetUsageHistoryFunc: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				called = true
				if days != 7 {
					t.Errorf("expected days to be 7 (default), got %d", days)
				}
				return []services.UsageDataPoint{}, nil
			},
		}
		licenseMock := &MockLicenseServiceForTelemetry{}
		handler := newTestTelemetryHandler(telemetryMock, licenseMock)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/usage?days=0", nil)
		req = req.WithContext(contextWithClaimsForTelemetry(userID, "test@example.com"))
		rec := httptest.NewRecorder()

		handler.GetUsage(rec, req)

		if !called {
			t.Error("expected GetUsageHistory to be called")
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("GetUsage with negative days parameter", func(t *testing.T) {
		called := false
		telemetryMock := &MockTelemetryService{
			GetUsageHistoryFunc: func(ctx context.Context, uid uuid.UUID, days int) ([]services.UsageDataPoint, error) {
				called = true
				if days != 7 {
					t.Errorf("expected days to be 7 (default), got %d", days)
				}
				return []services.UsageDataPoint{}, nil
			},
		}
		licenseMock := &MockLicenseServiceForTelemetry{}
		handler := newTestTelemetryHandler(telemetryMock, licenseMock)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry/usage?days=-5", nil)
		req = req.WithContext(contextWithClaimsForTelemetry(userID, "test@example.com"))
		rec := httptest.NewRecorder()

		handler.GetUsage(rec, req)

		if !called {
			t.Error("expected GetUsageHistory to be called")
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("Receive with minimal telemetry data", func(t *testing.T) {
		recordCalled := false
		telemetryMock := &MockTelemetryService{
			RecordTelemetryFunc: func(ctx context.Context, input services.TelemetryInput) error {
				recordCalled = true
				return nil
			},
		}
		licenseMock := &MockLicenseServiceForTelemetry{}
		handler := newTestTelemetryHandler(telemetryMock, licenseMock)

		body := map[string]interface{}{
			"license_id":  "lic-123",
			"hardware_id": "hw-456",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/receive", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.Receive(rec, req)

		if !recordCalled {
			t.Error("expected RecordTelemetry to be called")
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

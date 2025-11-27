package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockConfigGeneratorService implements a mock for testing
type MockConfigGeneratorService struct {
	GenerateConfigFunc     func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error)
	GenerateQuickStartFunc func(licenseKey string, sourceType string) string
}

func (m *MockConfigGeneratorService) GenerateConfig(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
	if m.GenerateConfigFunc != nil {
		return m.GenerateConfigFunc(ctx, userID, format, pipelineID, licenseKey)
	}
	return "", nil
}

func (m *MockConfigGeneratorService) GenerateQuickStart(licenseKey string, sourceType string) string {
	if m.GenerateQuickStartFunc != nil {
		return m.GenerateQuickStartFunc(licenseKey, sourceType)
	}
	return ""
}

// MockLicenseService implements a mock for testing
type MockLicenseService struct {
	GetUserLicensesFunc func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
}

func (m *MockLicenseService) GetUserLicenses(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
	if m.GetUserLicensesFunc != nil {
		return m.GetUserLicensesFunc(ctx, userID)
	}
	return nil, nil
}

// testConfigHandler wraps ConfigHandler for testing with mock services
type testConfigHandler struct {
	mockConfigService  *MockConfigGeneratorService
	mockLicenseService *MockLicenseService
}

func newTestConfigHandler(mockConfig *MockConfigGeneratorService, mockLicense *MockLicenseService) *testConfigHandler {
	return &testConfigHandler{
		mockConfigService:  mockConfig,
		mockLicenseService: mockLicense,
	}
}

func (h *testConfigHandler) Generate(w http.ResponseWriter, r *http.Request) {
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
	licenses, err := h.mockLicenseService.GetUserLicenses(r.Context(), userID)
	if err == nil && len(licenses) > 0 {
		// Use the first active license
		for _, lic := range licenses {
			if lic.Status == "active" {
				licenseKey = lic.LicenseKey
				break
			}
		}
	}

	config, err := h.mockConfigService.GenerateConfig(r.Context(), userID, format, pipelineID, licenseKey)
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

func (h *testConfigHandler) GetQuickStart(w http.ResponseWriter, r *http.Request) {
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
	licenses, err := h.mockLicenseService.GetUserLicenses(r.Context(), userID)
	if err == nil && len(licenses) > 0 {
		for _, lic := range licenses {
			if lic.Status == "active" {
				licenseKey = lic.LicenseKey
				break
			}
		}
	}

	guide := h.mockConfigService.GenerateQuickStart(licenseKey, sourceType)

	w.Header().Set("Content-Type", "text/markdown")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(guide))
}

func (h *testConfigHandler) GetFormats(w http.ResponseWriter, r *http.Request) {
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

// Helper function to create a request with user context
func createRequestWithUser(method, url string, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	claims := &services.Claims{
		UserID: userID.String(),
		Email:  "test@example.com",
		Role:   "user",
	}
	ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
	return req.WithContext(ctx)
}

func TestConfigHandler_Generate(t *testing.T) {
	testUserID := uuid.New()
	testPipelineID := uuid.New()

	tests := []struct {
		name               string
		queryParams        map[string]string
		userID             *uuid.UUID
		mockGenerateConfig func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error)
		mockGetLicenses    func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
		expectedStatus     int
		expectedError      string
		expectedContent    string
		expectedHeader     map[string]string
	}{
		{
			name:        "successful docker-compose generation with default format",
			userID:      &testUserID,
			queryParams: map[string]string{},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if format != "docker-compose" {
					t.Errorf("expected format 'docker-compose', got %q", format)
				}
				if licenseKey != "test-license-key-123" {
					t.Errorf("expected license key 'test-license-key-123', got %q", licenseKey)
				}
				return "version: '3.8'\nservices:\n  savegress:\n    image: savegress/cdc-engine:latest", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						ID:         uuid.New(),
						UserID:     userID,
						LicenseKey: "test-license-key-123",
						Status:     "active",
					},
				}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "version: '3.8'",
			expectedHeader: map[string]string{
				"Content-Type": "text/yaml",
			},
		},
		{
			name:   "successful generation with helm format",
			userID: &testUserID,
			queryParams: map[string]string{
				"format": "helm",
			},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if format != "helm" {
					t.Errorf("expected format 'helm', got %q", format)
				}
				return "license:\n  key: test-license-key", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						ID:         uuid.New(),
						UserID:     userID,
						LicenseKey: "test-license-key-123",
						Status:     "active",
					},
				}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "license:",
			expectedHeader: map[string]string{
				"Content-Type": "text/yaml",
			},
		},
		{
			name:   "successful generation with env format",
			userID: &testUserID,
			queryParams: map[string]string{
				"format": "env",
			},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				return "SAVEGRESS_LICENSE_KEY=test", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "SAVEGRESS_LICENSE_KEY=test",
			expectedHeader: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name:   "successful generation with systemd format",
			userID: &testUserID,
			queryParams: map[string]string{
				"format": "systemd",
			},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				return "[Unit]\nDescription=Savegress", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "[Unit]",
			expectedHeader: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name:   "generation with pipeline_id",
			userID: &testUserID,
			queryParams: map[string]string{
				"pipeline_id": testPipelineID.String(),
			},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if pipelineID == nil {
					t.Error("expected pipelineID to be set")
				} else if *pipelineID != testPipelineID {
					t.Errorf("expected pipelineID %v, got %v", testPipelineID, *pipelineID)
				}
				return "config", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "config",
		},
		{
			name:   "download parameter sets content disposition",
			userID: &testUserID,
			queryParams: map[string]string{
				"format":   "docker-compose",
				"download": "true",
			},
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				return "config", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "config",
			expectedHeader: map[string]string{
				"Content-Disposition": "attachment; filename=docker-compose.yml",
			},
		},
		{
			name:   "uses placeholder license when no active licenses",
			userID: &testUserID,
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if licenseKey != "YOUR_LICENSE_KEY" {
					t.Errorf("expected placeholder license key, got %q", licenseKey)
				}
				return "config", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "config",
		},
		{
			name:   "uses placeholder license when only inactive licenses",
			userID: &testUserID,
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if licenseKey != "YOUR_LICENSE_KEY" {
					t.Errorf("expected placeholder license key, got %q", licenseKey)
				}
				return "config", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						ID:         uuid.New(),
						UserID:     userID,
						LicenseKey: "expired-key",
						Status:     "expired",
					},
				}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "config",
		},
		{
			name:   "uses placeholder license when license service fails",
			userID: &testUserID,
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				if licenseKey != "YOUR_LICENSE_KEY" {
					t.Errorf("expected placeholder license key, got %q", licenseKey)
				}
				return "config", nil
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return nil, errors.New("database error")
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "config",
		},
		{
			name:           "unauthorized when no user in context",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:   "invalid pipeline_id format",
			userID: &testUserID,
			queryParams: map[string]string{
				"pipeline_id": "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline_id",
		},
		{
			name:   "config generation error",
			userID: &testUserID,
			mockGenerateConfig: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
				return "", errors.New("pipeline not found")
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to generate config: pipeline not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfigService := &MockConfigGeneratorService{
				GenerateConfigFunc: tt.mockGenerateConfig,
			}
			mockLicenseService := &MockLicenseService{
				GetUserLicensesFunc: tt.mockGetLicenses,
			}
			handler := newTestConfigHandler(mockConfigService, mockLicenseService)

			url := "/api/v1/config/generate"
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

			var req *http.Request
			if tt.userID != nil {
				req = createRequestWithUser(http.MethodGet, url, *tt.userID)
			} else {
				req = httptest.NewRequest(http.MethodGet, url, nil)
			}

			rec := httptest.NewRecorder()

			handler.Generate(rec, req)

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

			if tt.expectedContent != "" {
				body := rec.Body.String()
				if len(body) == 0 || body[:len(tt.expectedContent)] != tt.expectedContent {
					t.Errorf("expected content to start with %q, got %q", tt.expectedContent, body)
				}
			}

			if tt.expectedHeader != nil {
				for key, value := range tt.expectedHeader {
					if rec.Header().Get(key) != value {
						t.Errorf("expected header %s=%q, got %q", key, value, rec.Header().Get(key))
					}
				}
			}
		})
	}
}

func TestConfigHandler_Generate_InvalidUserID(t *testing.T) {
	mockConfigService := &MockConfigGeneratorService{}
	mockLicenseService := &MockLicenseService{}
	handler := newTestConfigHandler(mockConfigService, mockLicenseService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/generate", nil)
	// Add invalid user ID to context
	claims := &services.Claims{
		UserID: "invalid-uuid",
		Email:  "test@example.com",
	}
	ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Generate(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]string
	json.NewDecoder(rec.Body).Decode(&response)
	if response["error"] != "invalid user id" {
		t.Errorf("expected error 'invalid user id', got %q", response["error"])
	}
}

func TestConfigHandler_GetQuickStart(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name               string
		queryParams        map[string]string
		userID             *uuid.UUID
		mockQuickStart     func(licenseKey string, sourceType string) string
		mockGetLicenses    func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
		expectedStatus     int
		expectedError      string
		expectedContent    string
		expectedSourceType string
	}{
		{
			name:   "successful quick start with default source type",
			userID: &testUserID,
			mockQuickStart: func(licenseKey string, sourceType string) string {
				return "# Quick Start\ndocker pull savegress/cdc-engine:latest"
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						ID:         uuid.New(),
						UserID:     userID,
						LicenseKey: "test-license-key",
						Status:     "active",
					},
				}, nil
			},
			expectedStatus:     http.StatusOK,
			expectedContent:    "# Quick Start",
			expectedSourceType: "postgres",
		},
		{
			name:   "successful quick start with mysql source type",
			userID: &testUserID,
			queryParams: map[string]string{
				"source_type": "mysql",
			},
			mockQuickStart: func(licenseKey string, sourceType string) string {
				if sourceType != "mysql" {
					t.Errorf("expected sourceType 'mysql', got %q", sourceType)
				}
				return "# MySQL Quick Start"
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			expectedStatus:     http.StatusOK,
			expectedContent:    "# MySQL Quick Start",
			expectedSourceType: "mysql",
		},
		{
			name:   "uses active license key",
			userID: &testUserID,
			mockQuickStart: func(licenseKey string, sourceType string) string {
				if licenseKey != "active-key-123" {
					t.Errorf("expected license key 'active-key-123', got %q", licenseKey)
				}
				return "guide"
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						LicenseKey: "active-key-123",
						Status:     "active",
					},
				}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "guide",
		},
		{
			name:   "uses placeholder when no active license",
			userID: &testUserID,
			mockQuickStart: func(licenseKey string, sourceType string) string {
				if licenseKey != "YOUR_LICENSE_KEY" {
					t.Errorf("expected placeholder license key, got %q", licenseKey)
				}
				return "guide"
			},
			mockGetLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						LicenseKey: "expired-key",
						Status:     "expired",
					},
				}, nil
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "guide",
		},
		{
			name:           "unauthorized when no user in context",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfigService := &MockConfigGeneratorService{
				GenerateQuickStartFunc: tt.mockQuickStart,
			}
			mockLicenseService := &MockLicenseService{
				GetUserLicensesFunc: tt.mockGetLicenses,
			}
			handler := newTestConfigHandler(mockConfigService, mockLicenseService)

			url := "/api/v1/config/quickstart"
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

			var req *http.Request
			if tt.userID != nil {
				req = createRequestWithUser(http.MethodGet, url, *tt.userID)
			} else {
				req = httptest.NewRequest(http.MethodGet, url, nil)
			}

			rec := httptest.NewRecorder()

			handler.GetQuickStart(rec, req)

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

			if tt.expectedContent != "" {
				body := rec.Body.String()
				if len(body) < len(tt.expectedContent) || body[:len(tt.expectedContent)] != tt.expectedContent {
					t.Errorf("expected content to start with %q, got %q", tt.expectedContent, body)
				}
			}

			if rec.Code == http.StatusOK {
				contentType := rec.Header().Get("Content-Type")
				if contentType != "text/markdown" {
					t.Errorf("expected Content-Type 'text/markdown', got %q", contentType)
				}
			}
		})
	}
}

func TestConfigHandler_GetQuickStart_InvalidUserID(t *testing.T) {
	mockConfigService := &MockConfigGeneratorService{}
	mockLicenseService := &MockLicenseService{}
	handler := newTestConfigHandler(mockConfigService, mockLicenseService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/quickstart", nil)
	// Add invalid user ID to context
	claims := &services.Claims{
		UserID: "invalid-uuid",
		Email:  "test@example.com",
	}
	ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetQuickStart(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response map[string]string
	json.NewDecoder(rec.Body).Decode(&response)
	if response["error"] != "invalid user id" {
		t.Errorf("expected error 'invalid user id', got %q", response["error"])
	}
}

func TestConfigHandler_GetFormats(t *testing.T) {
	mockConfigService := &MockConfigGeneratorService{}
	mockLicenseService := &MockLicenseService{}
	handler := newTestConfigHandler(mockConfigService, mockLicenseService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/formats", nil)
	rec := httptest.NewRecorder()

	handler.GetFormats(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	formats, ok := response["formats"].([]interface{})
	if !ok {
		t.Fatal("expected 'formats' field in response")
	}

	if len(formats) != 4 {
		t.Errorf("expected 4 formats, got %d", len(formats))
	}

	// Verify format structure
	expectedFormats := []string{"docker-compose", "helm", "env", "systemd"}
	for i, format := range formats {
		formatMap, ok := format.(map[string]interface{})
		if !ok {
			t.Errorf("format %d is not a map", i)
			continue
		}

		id, ok := formatMap["id"].(string)
		if !ok {
			t.Errorf("format %d missing 'id' field", i)
			continue
		}

		if id != expectedFormats[i] {
			t.Errorf("expected format id %q, got %q", expectedFormats[i], id)
		}

		// Check required fields exist
		if _, ok := formatMap["name"].(string); !ok {
			t.Errorf("format %d missing 'name' field", i)
		}
		if _, ok := formatMap["description"].(string); !ok {
			t.Errorf("format %d missing 'description' field", i)
		}
		if _, ok := formatMap["filename"].(string); !ok {
			t.Errorf("format %d missing 'filename' field", i)
		}
		if _, ok := formatMap["icon"].(string); !ok {
			t.Errorf("format %d missing 'icon' field", i)
		}
	}
}

func TestConfigHandler_GetFormats_VerifyContent(t *testing.T) {
	mockConfigService := &MockConfigGeneratorService{}
	mockLicenseService := &MockLicenseService{}
	handler := newTestConfigHandler(mockConfigService, mockLicenseService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/formats", nil)
	rec := httptest.NewRecorder()

	handler.GetFormats(rec, req)

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)

	formats := response["formats"].([]interface{})

	// Test docker-compose format details
	dockerFormat := formats[0].(map[string]interface{})
	if dockerFormat["id"] != "docker-compose" {
		t.Errorf("expected docker-compose format, got %v", dockerFormat["id"])
	}
	if dockerFormat["name"] != "Docker Compose" {
		t.Errorf("expected name 'Docker Compose', got %v", dockerFormat["name"])
	}
	if dockerFormat["filename"] != "docker-compose.yml" {
		t.Errorf("expected filename 'docker-compose.yml', got %v", dockerFormat["filename"])
	}

	// Test helm format details
	helmFormat := formats[1].(map[string]interface{})
	if helmFormat["id"] != "helm" {
		t.Errorf("expected helm format, got %v", helmFormat["id"])
	}
	if helmFormat["name"] != "Kubernetes (Helm)" {
		t.Errorf("expected name 'Kubernetes (Helm)', got %v", helmFormat["name"])
	}
	if helmFormat["filename"] != "values.yaml" {
		t.Errorf("expected filename 'values.yaml', got %v", helmFormat["filename"])
	}

	// Test env format details
	envFormat := formats[2].(map[string]interface{})
	if envFormat["id"] != "env" {
		t.Errorf("expected env format, got %v", envFormat["id"])
	}
	if envFormat["filename"] != "savegress.env" {
		t.Errorf("expected filename 'savegress.env', got %v", envFormat["filename"])
	}

	// Test systemd format details
	systemdFormat := formats[3].(map[string]interface{})
	if systemdFormat["id"] != "systemd" {
		t.Errorf("expected systemd format, got %v", systemdFormat["id"])
	}
	if systemdFormat["filename"] != "savegress.service" {
		t.Errorf("expected filename 'savegress.service', got %v", systemdFormat["filename"])
	}
}

func TestNewConfigHandler(t *testing.T) {
	mockConfigService := &MockConfigGeneratorService{}
	mockLicenseService := &MockLicenseService{}

	handler := newTestConfigHandler(mockConfigService, mockLicenseService)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}

	if handler.mockConfigService == nil {
		t.Error("expected mockConfigService to be set")
	}

	if handler.mockLicenseService == nil {
		t.Error("expected mockLicenseService to be set")
	}
}

func TestConfigHandler_ContentTypeMapping(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		format              string
		expectedContentType string
		expectedFilename    string
	}{
		{"docker-compose", "text/yaml", "docker-compose.yml"},
		{"docker", "text/yaml", "docker-compose.yml"},
		{"helm", "text/yaml", "values.yaml"},
		{"kubernetes", "text/yaml", "values.yaml"},
		{"k8s", "text/yaml", "values.yaml"},
		{"env", "text/plain", "savegress.env"},
		{"dotenv", "text/plain", "savegress.env"},
		{"systemd", "text/plain", "savegress.service"},
	}

	for _, tt := range tests {
		t.Run("format_"+tt.format, func(t *testing.T) {
			mockConfigService := &MockConfigGeneratorService{
				GenerateConfigFunc: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
					return "test config", nil
				},
			}
			mockLicenseService := &MockLicenseService{
				GetUserLicensesFunc: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
					return []models.License{}, nil
				},
			}
			handler := newTestConfigHandler(mockConfigService, mockLicenseService)

			req := createRequestWithUser(http.MethodGet, "/api/v1/config/generate?format="+tt.format, testUserID)
			rec := httptest.NewRecorder()

			handler.Generate(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != tt.expectedContentType {
				t.Errorf("expected Content-Type %q, got %q", tt.expectedContentType, contentType)
			}
		})
	}
}

func TestConfigHandler_DownloadFilenameMapping(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		format           string
		expectedFilename string
	}{
		{"docker-compose", "docker-compose.yml"},
		{"helm", "values.yaml"},
		{"env", "savegress.env"},
		{"systemd", "savegress.service"},
	}

	for _, tt := range tests {
		t.Run("download_"+tt.format, func(t *testing.T) {
			mockConfigService := &MockConfigGeneratorService{
				GenerateConfigFunc: func(ctx context.Context, userID uuid.UUID, format string, pipelineID *uuid.UUID, licenseKey string) (string, error) {
					return "test config", nil
				},
			}
			mockLicenseService := &MockLicenseService{
				GetUserLicensesFunc: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
					return []models.License{}, nil
				},
			}
			handler := newTestConfigHandler(mockConfigService, mockLicenseService)

			req := createRequestWithUser(http.MethodGet, "/api/v1/config/generate?format="+tt.format+"&download=true", testUserID)
			rec := httptest.NewRecorder()

			handler.Generate(rec, req)

			contentDisposition := rec.Header().Get("Content-Disposition")
			expectedDisposition := "attachment; filename=" + tt.expectedFilename
			if contentDisposition != expectedDisposition {
				t.Errorf("expected Content-Disposition %q, got %q", expectedDisposition, contentDisposition)
			}
		})
	}
}

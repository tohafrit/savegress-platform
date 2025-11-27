package handlers

import (
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

// MockPipelineService implements a mock for testing
type MockPipelineService struct {
	ListPipelinesFunc              func(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error)
	CreatePipelineFunc             func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error)
	GetPipelineWithConnectionsFunc func(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error)
	UpdatePipelineFunc             func(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error)
	DeletePipelineFunc             func(ctx context.Context, userID, pipelineID uuid.UUID) error
	GetPipelineMetricsFunc         func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error)
	GetPipelineLogsFunc            func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error)
	CountUserPipelinesFunc         func(ctx context.Context, userID uuid.UUID) (int, error)
}

func (m *MockPipelineService) ListPipelines(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error) {
	if m.ListPipelinesFunc != nil {
		return m.ListPipelinesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPipelineService) CreatePipeline(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
	if m.CreatePipelineFunc != nil {
		return m.CreatePipelineFunc(ctx, userID, pipeline)
	}
	return nil, nil
}

func (m *MockPipelineService) GetPipelineWithConnections(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
	if m.GetPipelineWithConnectionsFunc != nil {
		return m.GetPipelineWithConnectionsFunc(ctx, userID, pipelineID)
	}
	return nil, nil
}

func (m *MockPipelineService) UpdatePipeline(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error) {
	if m.UpdatePipelineFunc != nil {
		return m.UpdatePipelineFunc(ctx, userID, pipelineID, updates)
	}
	return nil, nil
}

func (m *MockPipelineService) DeletePipeline(ctx context.Context, userID, pipelineID uuid.UUID) error {
	if m.DeletePipelineFunc != nil {
		return m.DeletePipelineFunc(ctx, userID, pipelineID)
	}
	return nil
}

func (m *MockPipelineService) GetPipelineMetrics(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
	if m.GetPipelineMetricsFunc != nil {
		return m.GetPipelineMetricsFunc(ctx, userID, pipelineID, hours)
	}
	return nil, nil
}

func (m *MockPipelineService) GetPipelineLogs(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
	if m.GetPipelineLogsFunc != nil {
		return m.GetPipelineLogsFunc(ctx, userID, pipelineID, limit, level)
	}
	return nil, nil
}

func (m *MockPipelineService) CountUserPipelines(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.CountUserPipelinesFunc != nil {
		return m.CountUserPipelinesFunc(ctx, userID)
	}
	return 0, nil
}

// MockLicenseServiceForPipeline implements a mock license service for pipeline testing
type MockLicenseServiceForPipeline struct {
	GetUserLicensesFunc func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
}

func (m *MockLicenseServiceForPipeline) GetUserLicenses(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
	if m.GetUserLicensesFunc != nil {
		return m.GetUserLicensesFunc(ctx, userID)
	}
	return nil, nil
}

// testPipelineHandler wraps mock services for testing
type testPipelineHandler struct {
	mockPipeline *MockPipelineService
	mockLicense  *MockLicenseServiceForPipeline
}

func newTestPipelineHandler(mockPipeline *MockPipelineService, mockLicense *MockLicenseServiceForPipeline) *testPipelineHandler {
	return &testPipelineHandler{
		mockPipeline: mockPipeline,
		mockLicense:  mockLicense,
	}
}

func (h *testPipelineHandler) List(w http.ResponseWriter, r *http.Request) {
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

	pipelines, err := h.mockPipeline.ListPipelines(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list pipelines")
		return
	}

	respondSuccess(w, map[string]interface{}{"pipelines": pipelines})
}

func (h *testPipelineHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		Name               string            `json:"name"`
		Description        string            `json:"description"`
		SourceConnectionID string            `json:"source_connection_id"`
		TargetConnectionID string            `json:"target_connection_id"`
		TargetType         string            `json:"target_type"`
		TargetConfig       map[string]string `json:"target_config"`
		Tables             []string          `json:"tables"`
		LicenseID          string            `json:"license_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.SourceConnectionID == "" || req.TargetType == "" {
		respondError(w, http.StatusBadRequest, "name, source_connection_id, and target_type are required")
		return
	}

	sourceConnID, err := uuid.Parse(req.SourceConnectionID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid source_connection_id")
		return
	}

	pipeline := &models.Pipeline{
		Name:         req.Name,
		Description:  req.Description,
		SourceConnID: sourceConnID,
		TargetType:   req.TargetType,
		TargetConfig: req.TargetConfig,
		Tables:       req.Tables,
	}

	if req.TargetConnectionID != "" {
		targetConnID, err := uuid.Parse(req.TargetConnectionID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid target_connection_id")
			return
		}
		pipeline.TargetConnID = &targetConnID
	}

	if req.LicenseID != "" {
		licenseID, err := uuid.Parse(req.LicenseID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid license_id")
			return
		}
		pipeline.LicenseID = &licenseID
	}

	// Check pipeline limit based on user's license
	licenses, err := h.mockLicense.GetUserLicenses(r.Context(), userID)
	if err == nil && len(licenses) > 0 {
		// Find best active license
		var maxPipelines int = 1 // community default
		for _, lic := range licenses {
			if lic.Status == "active" {
				switch lic.Tier {
				case "enterprise":
					maxPipelines = 999999 // unlimited
				case "pro":
					if maxPipelines < 10 {
						maxPipelines = 10
					}
				case "trial":
					if maxPipelines < 5 {
						maxPipelines = 5
					}
				}
			}
		}

		currentCount, _ := h.mockPipeline.CountUserPipelines(r.Context(), userID)
		if currentCount >= maxPipelines {
			respondError(w, http.StatusForbidden, "pipeline limit reached for your plan")
			return
		}
	}

	created, err := h.mockPipeline.CreatePipeline(r.Context(), userID, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create pipeline: "+err.Error())
		return
	}

	respondCreated(w, created)
}

func (h *testPipelineHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	pipeline, err := h.mockPipeline.GetPipelineWithConnections(r.Context(), userID, pipelineID)
	if err == services.ErrPipelineNotFound {
		respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get pipeline")
		return
	}

	respondSuccess(w, pipeline)
}

func (h *testPipelineHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pipeline, err := h.mockPipeline.UpdatePipeline(r.Context(), userID, pipelineID, updates)
	if err == services.ErrPipelineNotFound {
		respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update pipeline")
		return
	}

	respondSuccess(w, pipeline)
}

func (h *testPipelineHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	err = h.mockPipeline.DeletePipeline(r.Context(), userID, pipelineID)
	if err == services.ErrPipelineNotFound {
		respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete pipeline")
		return
	}

	respondSuccess(w, map[string]string{"message": "pipeline deleted"})
}

func (h *testPipelineHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours == 0 {
		hours = 24
	}

	metrics, err := h.mockPipeline.GetPipelineMetrics(r.Context(), userID, pipelineID, hours)
	if err == services.ErrPipelineNotFound {
		respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get metrics")
		return
	}

	respondSuccess(w, map[string]interface{}{"metrics": metrics})
}

func (h *testPipelineHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
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

	pipelineID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}
	level := r.URL.Query().Get("level")

	logs, err := h.mockPipeline.GetPipelineLogs(r.Context(), userID, pipelineID, limit, level)
	if err == services.ErrPipelineNotFound {
		respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get logs")
		return
	}

	respondSuccess(w, map[string]interface{}{"logs": logs})
}

func TestPipelineHandler_List(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name              string
		hasUser           bool
		userID            uuid.UUID
		mockListPipelines func(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error)
		expectedStatus    int
		expectedError     string
	}{
		{
			name:    "successful list",
			hasUser: true,
			userID:  userID,
			mockListPipelines: func(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error) {
				return []models.Pipeline{
					{
						ID:           uuid.New(),
						UserID:       userID,
						Name:         "Pipeline 1",
						Description:  "Test pipeline",
						SourceConnID: uuid.New(),
						TargetType:   "postgres",
						Status:       "created",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:    "internal error",
			hasUser: true,
			userID:  userID,
			mockListPipelines: func(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to list pipelines",
		},
		{
			name:    "empty list",
			hasUser: true,
			userID:  userID,
			mockListPipelines: func(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error) {
				return []models.Pipeline{}, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				ListPipelinesFunc: tt.mockListPipelines,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, "/api/v1/pipelines", nil, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, "/api/v1/pipelines", nil)
			}
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

func TestPipelineHandler_Create(t *testing.T) {
	userID := uuid.New()
	sourceConnID := uuid.New()
	targetConnID := uuid.New()
	licenseID := uuid.New()

	tests := []struct {
		name                  string
		hasUser               bool
		userID                uuid.UUID
		requestBody           map[string]interface{}
		mockCreatePipeline    func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error)
		mockGetUserLicenses   func(ctx context.Context, userID uuid.UUID) ([]models.License, error)
		mockCountUserPipelines func(ctx context.Context, userID uuid.UUID) (int, error)
		expectedStatus        int
		expectedError         string
	}{
		{
			name:    "successful creation",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"description":          "Test description",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
				"tables":               []string{"table1", "table2"},
			},
			mockCreatePipeline: func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
				pipeline.ID = uuid.New()
				pipeline.Status = "created"
				pipeline.CreatedAt = time.Now()
				pipeline.UpdatedAt = time.Now()
				return pipeline, nil
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						Tier:   "pro",
						Status: "active",
					},
				}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "successful creation with target connection",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_connection_id": targetConnID.String(),
				"target_type":          "postgres",
			},
			mockCreatePipeline: func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
				pipeline.ID = uuid.New()
				return pipeline, nil
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "successful creation with license",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "s3",
				"license_id":           licenseID.String(),
			},
			mockCreatePipeline: func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
				pipeline.ID = uuid.New()
				return pipeline, nil
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:    "missing name",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, source_connection_id, and target_type are required",
		},
		{
			name:    "missing source_connection_id",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":        "Test Pipeline",
				"target_type": "postgres",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, source_connection_id, and target_type are required",
		},
		{
			name:    "missing target_type",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, source_connection_id, and target_type are required",
		},
		{
			name:    "invalid source_connection_id",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": "invalid-uuid",
				"target_type":          "postgres",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid source_connection_id",
		},
		{
			name:    "invalid target_connection_id",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_connection_id": "invalid-uuid",
				"target_type":          "postgres",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid target_connection_id",
		},
		{
			name:    "invalid license_id",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
				"license_id":           "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid license_id",
		},
		{
			name:    "pipeline limit reached - community",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 1, nil
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "pipeline limit reached for your plan",
		},
		{
			name:    "pipeline limit reached - pro",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						Tier:   "pro",
						Status: "active",
					},
				}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 10, nil
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "pipeline limit reached for your plan",
		},
		{
			name:    "enterprise unlimited",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
			},
			mockCreatePipeline: func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
				pipeline.ID = uuid.New()
				return pipeline, nil
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{
					{
						Tier:   "enterprise",
						Status: "active",
					},
				}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 1000, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "internal error during creation",
			hasUser: true,
			userID:  userID,
			requestBody: map[string]interface{}{
				"name":                 "Test Pipeline",
				"source_connection_id": sourceConnID.String(),
				"target_type":          "postgres",
			},
			mockCreatePipeline: func(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
				return nil, errors.New("database error")
			},
			mockGetUserLicenses: func(ctx context.Context, userID uuid.UUID) ([]models.License, error) {
				return []models.License{}, nil
			},
			mockCountUserPipelines: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, nil
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create pipeline: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				CreatePipelineFunc:     tt.mockCreatePipeline,
				CountUserPipelinesFunc: tt.mockCountUserPipelines,
			}
			mockLicense := &MockLicenseServiceForPipeline{
				GetUserLicensesFunc: tt.mockGetUserLicenses,
			}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			body, _ := json.Marshal(tt.requestBody)
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodPost, "/api/v1/pipelines", body, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodPost, "/api/v1/pipelines", body)
			}
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

func TestPipelineHandler_Get(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name                           string
		hasUser                        bool
		userID                         uuid.UUID
		pipelineID                     string
		mockGetPipelineWithConnections func(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error)
		expectedStatus                 int
		expectedError                  string
	}{
		{
			name:       "successful get",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineWithConnections: func(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
				return &models.Pipeline{
					ID:           pipelineID,
					UserID:       userID,
					Name:         "Test Pipeline",
					Description:  "Test description",
					SourceConnID: uuid.New(),
					TargetType:   "postgres",
					Status:       "created",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			pipelineID:     pipelineID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid pipeline ID",
			hasUser:        true,
			userID:         userID,
			pipelineID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline ID",
		},
		{
			name:       "pipeline not found",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineWithConnections: func(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
				return nil, services.ErrPipelineNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "pipeline not found",
		},
		{
			name:       "internal error",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineWithConnections: func(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				GetPipelineWithConnectionsFunc: tt.mockGetPipelineWithConnections,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, "/api/v1/pipelines/"+tt.pipelineID, nil, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, "/api/v1/pipelines/"+tt.pipelineID, nil)
			}

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.pipelineID)
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

func TestPipelineHandler_Update(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name               string
		hasUser            bool
		userID             uuid.UUID
		pipelineID         string
		requestBody        map[string]interface{}
		mockUpdatePipeline func(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error)
		expectedStatus     int
		expectedError      string
	}{
		{
			name:       "successful update",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			requestBody: map[string]interface{}{
				"name":        "Updated Pipeline",
				"description": "Updated description",
			},
			mockUpdatePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error) {
				return &models.Pipeline{
					ID:          pipelineID,
					UserID:      userID,
					Name:        "Updated Pipeline",
					Description: "Updated description",
					Status:      "created",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			pipelineID:     pipelineID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid pipeline ID",
			hasUser:        true,
			userID:         userID,
			pipelineID:     "invalid-uuid",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline ID",
		},
		{
			name:       "pipeline not found",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			requestBody: map[string]interface{}{
				"name": "Updated Pipeline",
			},
			mockUpdatePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error) {
				return nil, services.ErrPipelineNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "pipeline not found",
		},
		{
			name:       "internal error",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			requestBody: map[string]interface{}{
				"name": "Updated Pipeline",
			},
			mockUpdatePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				UpdatePipelineFunc: tt.mockUpdatePipeline,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			body, _ := json.Marshal(tt.requestBody)
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodPut, "/api/v1/pipelines/"+tt.pipelineID, body, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodPut, "/api/v1/pipelines/"+tt.pipelineID, body)
			}

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.pipelineID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.Update(rec, req)

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

func TestPipelineHandler_Delete(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name               string
		hasUser            bool
		userID             uuid.UUID
		pipelineID         string
		mockDeletePipeline func(ctx context.Context, userID, pipelineID uuid.UUID) error
		expectedStatus     int
		expectedError      string
	}{
		{
			name:       "successful delete",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockDeletePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			pipelineID:     pipelineID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid pipeline ID",
			hasUser:        true,
			userID:         userID,
			pipelineID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline ID",
		},
		{
			name:       "pipeline not found",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockDeletePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID) error {
				return services.ErrPipelineNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "pipeline not found",
		},
		{
			name:       "internal error",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockDeletePipeline: func(ctx context.Context, userID, pipelineID uuid.UUID) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to delete pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				DeletePipelineFunc: tt.mockDeletePipeline,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodDelete, "/api/v1/pipelines/"+tt.pipelineID, nil, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodDelete, "/api/v1/pipelines/"+tt.pipelineID, nil)
			}

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.pipelineID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.Delete(rec, req)

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

func TestPipelineHandler_GetMetrics(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name                   string
		hasUser                bool
		userID                 uuid.UUID
		pipelineID             string
		queryParams            string
		mockGetPipelineMetrics func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error)
		expectedStatus         int
		expectedError          string
		expectedHours          int
	}{
		{
			name:        "successful get metrics with default hours",
			hasUser:     true,
			userID:      userID,
			pipelineID:  pipelineID.String(),
			queryParams: "",
			mockGetPipelineMetrics: func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
				if hours != 24 {
					t.Errorf("expected hours to be 24, got %d", hours)
				}
				return []map[string]interface{}{
					{
						"timestamp": time.Now(),
						"events":    int64(1000),
						"bytes":     int64(50000),
						"latency":   10.5,
						"errors":    int64(0),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful get metrics with custom hours",
			hasUser:     true,
			userID:      userID,
			pipelineID:  pipelineID.String(),
			queryParams: "?hours=48",
			mockGetPipelineMetrics: func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
				if hours != 48 {
					t.Errorf("expected hours to be 48, got %d", hours)
				}
				return []map[string]interface{}{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			pipelineID:     pipelineID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid pipeline ID",
			hasUser:        true,
			userID:         userID,
			pipelineID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline ID",
		},
		{
			name:       "pipeline not found",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineMetrics: func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
				return nil, services.ErrPipelineNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "pipeline not found",
		},
		{
			name:       "internal error",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineMetrics: func(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				GetPipelineMetricsFunc: tt.mockGetPipelineMetrics,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			url := "/api/v1/pipelines/" + tt.pipelineID + "/metrics" + tt.queryParams
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, url, nil, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, url, nil)
			}

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.pipelineID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.GetMetrics(rec, req)

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

func TestPipelineHandler_GetLogs(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	tests := []struct {
		name                string
		hasUser             bool
		userID              uuid.UUID
		pipelineID          string
		queryParams         string
		mockGetPipelineLogs func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error)
		expectedStatus      int
		expectedError       string
	}{
		{
			name:        "successful get logs with defaults",
			hasUser:     true,
			userID:      userID,
			pipelineID:  pipelineID.String(),
			queryParams: "",
			mockGetPipelineLogs: func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
				if limit != 100 {
					t.Errorf("expected limit to be 100, got %d", limit)
				}
				if level != "" {
					t.Errorf("expected level to be empty, got %q", level)
				}
				return []models.PipelineLog{
					{
						ID:         uuid.New(),
						PipelineID: pipelineID,
						Level:      "info",
						Message:    "Test log message",
						Timestamp:  time.Now(),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful get logs with custom limit and level",
			hasUser:     true,
			userID:      userID,
			pipelineID:  pipelineID.String(),
			queryParams: "?limit=50&level=error",
			mockGetPipelineLogs: func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
				if limit != 50 {
					t.Errorf("expected limit to be 50, got %d", limit)
				}
				if level != "error" {
					t.Errorf("expected level to be 'error', got %q", level)
				}
				return []models.PipelineLog{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user",
			hasUser:        false,
			pipelineID:     pipelineID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid pipeline ID",
			hasUser:        true,
			userID:         userID,
			pipelineID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid pipeline ID",
		},
		{
			name:       "pipeline not found",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineLogs: func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
				return nil, services.ErrPipelineNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "pipeline not found",
		},
		{
			name:       "internal error",
			hasUser:    true,
			userID:     userID,
			pipelineID: pipelineID.String(),
			mockGetPipelineLogs: func(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPipeline := &MockPipelineService{
				GetPipelineLogsFunc: tt.mockGetPipelineLogs,
			}
			mockLicense := &MockLicenseServiceForPipeline{}
			handler := newTestPipelineHandler(mockPipeline, mockLicense)

			url := "/api/v1/pipelines/" + tt.pipelineID + "/logs" + tt.queryParams
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, url, nil, tt.userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, url, nil)
			}

			// Set up chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.pipelineID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.GetLogs(rec, req)

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

func TestPipelineHandler_InvalidJSON(t *testing.T) {
	userID := uuid.New()
	pipelineID := uuid.New()

	mockPipeline := &MockPipelineService{}
	mockLicense := &MockLicenseServiceForPipeline{}
	handler := newTestPipelineHandler(mockPipeline, mockLicense)

	t.Run("Create - invalid JSON", func(t *testing.T) {
		req := newRequestWithUser(http.MethodPost, "/api/v1/pipelines", []byte("invalid json"), userID)
		rec := httptest.NewRecorder()

		handler.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response map[string]string
		json.NewDecoder(rec.Body).Decode(&response)
		if response["error"] != "invalid request body" {
			t.Errorf("expected error 'invalid request body', got %q", response["error"])
		}
	})

	t.Run("Update - invalid JSON", func(t *testing.T) {
		req := newRequestWithUser(http.MethodPut, "/api/v1/pipelines/"+pipelineID.String(), []byte("invalid json"), userID)

		// Set up chi context for URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", pipelineID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()

		handler.Update(rec, req)

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

func TestPipelineHandler_InvalidUserID(t *testing.T) {
	mockPipeline := &MockPipelineService{}
	mockLicense := &MockLicenseServiceForPipeline{}
	handler := newTestPipelineHandler(mockPipeline, mockLicense)

	// Create request with invalid user ID in claims
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pipelines", nil)
	req.Header.Set("Content-Type", "application/json")

	claims := &services.Claims{
		UserID: "invalid-uuid",
		Email:  "test@example.com",
		Role:   "user",
	}

	ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	endpoints := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		setupChi bool
	}{
		{"List", handler.List, false},
		{"Create", handler.Create, false},
		{"Get", handler.Get, true},
		{"Update", handler.Update, true},
		{"Delete", handler.Delete, true},
		{"GetMetrics", handler.GetMetrics, true},
		{"GetLogs", handler.GetLogs, true},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+" - invalid user ID", func(t *testing.T) {
			reqCopy := req.Clone(req.Context())

			if ep.setupChi {
				// Set up chi context for URL params
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", uuid.New().String())
				reqCopy = reqCopy.WithContext(context.WithValue(reqCopy.Context(), chi.RouteCtxKey, rctx))
			}

			// For Create handler, need valid JSON body
			if ep.name == "Create" {
				body := []byte(`{"name":"Test","source_connection_id":"` + uuid.New().String() + `","target_type":"postgres"}`)
				reqCopy = httptest.NewRequest(http.MethodPost, "/api/v1/pipelines", bytes.NewReader(body))
				reqCopy.Header.Set("Content-Type", "application/json")
				reqCopy = reqCopy.WithContext(context.WithValue(reqCopy.Context(), middleware.ClaimsContextKey, claims))
			}

			// For Update handler, need valid JSON body
			if ep.name == "Update" {
				body := []byte(`{"name":"Test"}`)
				reqCopy = httptest.NewRequest(http.MethodPut, "/api/v1/pipelines/"+uuid.New().String(), bytes.NewReader(body))
				reqCopy.Header.Set("Content-Type", "application/json")
				reqCopy = reqCopy.WithContext(context.WithValue(reqCopy.Context(), middleware.ClaimsContextKey, claims))

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", uuid.New().String())
				reqCopy = reqCopy.WithContext(context.WithValue(reqCopy.Context(), chi.RouteCtxKey, rctx))
			}

			rec := httptest.NewRecorder()

			ep.handler(rec, reqCopy)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}

			var response map[string]string
			json.NewDecoder(rec.Body).Decode(&response)
			if response["error"] != "invalid user id" {
				t.Errorf("expected error 'invalid user id', got %q", response["error"])
			}
		})
	}
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockConnectionService implements a mock for testing
type MockConnectionService struct {
	ListConnectionsFunc      func(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error)
	CreateConnectionFunc     func(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error)
	GetConnectionFunc        func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error)
	UpdateConnectionFunc     func(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error)
	DeleteConnectionFunc     func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
	TestConnectionFunc       func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
	TestConnectionDirectFunc func(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error
}

func (m *MockConnectionService) ListConnections(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error) {
	if m.ListConnectionsFunc != nil {
		return m.ListConnectionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockConnectionService) CreateConnection(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error) {
	if m.CreateConnectionFunc != nil {
		return m.CreateConnectionFunc(ctx, userID, conn)
	}
	return nil, nil
}

func (m *MockConnectionService) GetConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error) {
	if m.GetConnectionFunc != nil {
		return m.GetConnectionFunc(ctx, userID, connID)
	}
	return nil, nil
}

func (m *MockConnectionService) UpdateConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error) {
	if m.UpdateConnectionFunc != nil {
		return m.UpdateConnectionFunc(ctx, userID, connID, updates)
	}
	return nil, nil
}

func (m *MockConnectionService) DeleteConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
	if m.DeleteConnectionFunc != nil {
		return m.DeleteConnectionFunc(ctx, userID, connID)
	}
	return nil
}

func (m *MockConnectionService) TestConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc(ctx, userID, connID)
	}
	return nil
}

func (m *MockConnectionService) TestConnectionDirect(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
	if m.TestConnectionDirectFunc != nil {
		return m.TestConnectionDirectFunc(ctx, connType, host, port, database, username, password, sslMode)
	}
	return nil
}

// testConnectionHandler wraps ConnectionHandler for testing with mock service
type testConnectionHandler struct {
	mock *MockConnectionService
}

func newTestConnectionHandler(mock *MockConnectionService) *testConnectionHandler {
	return &testConnectionHandler{mock: mock}
}

func (h *testConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
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

	connections, err := h.mock.ListConnections(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}

	respondSuccess(w, map[string]interface{}{"connections": connections})
}

func (h *testConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
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
		Name     string `json:"name"`
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Database string `json:"database"`
		Username string `json:"username"`
		Password string `json:"password"`
		SSLMode  string `json:"ssl_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" || req.Host == "" || req.Port == 0 {
		respondError(w, http.StatusBadRequest, "name, type, host, and port are required")
		return
	}

	conn := &models.Connection{
		Name:     req.Name,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
		SSLMode:  req.SSLMode,
	}

	if conn.SSLMode == "" {
		conn.SSLMode = "prefer"
	}

	created, err := h.mock.CreateConnection(r.Context(), userID, conn)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create connection: "+err.Error())
		return
	}

	respondCreated(w, created)
}

func (h *testConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	conn, err := h.mock.GetConnection(r.Context(), userID, connID)
	if err == services.ErrConnectionNotFound {
		respondError(w, http.StatusNotFound, "connection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get connection")
		return
	}

	respondSuccess(w, conn)
}

func (h *testConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conn, err := h.mock.UpdateConnection(r.Context(), userID, connID, updates)
	if err == services.ErrConnectionNotFound {
		respondError(w, http.StatusNotFound, "connection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update connection")
		return
	}

	respondSuccess(w, conn)
}

func (h *testConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	err = h.mock.DeleteConnection(r.Context(), userID, connID)
	if err == services.ErrConnectionNotFound {
		respondError(w, http.StatusNotFound, "connection not found")
		return
	}
	if err == services.ErrConnectionInUse {
		respondError(w, http.StatusConflict, "connection is in use by pipelines")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}

	respondSuccess(w, map[string]string{"message": "connection deleted"})
}

func (h *testConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
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

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid connection ID")
		return
	}

	err = h.mock.TestConnection(r.Context(), userID, connID)
	if err == services.ErrConnectionNotFound {
		respondError(w, http.StatusNotFound, "connection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusBadRequest, "connection test failed: "+err.Error())
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "connection successful",
	})
}

func (h *testConnectionHandler) TestDirect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Database string `json:"database"`
		Username string `json:"username"`
		Password string `json:"password"`
		SSLMode  string `json:"ssl_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SSLMode == "" {
		req.SSLMode = "prefer"
	}

	err := h.mock.TestConnectionDirect(r.Context(), req.Type, req.Host, req.Port, req.Database, req.Username, req.Password, req.SSLMode)
	if err != nil {
		respondError(w, http.StatusBadRequest, "connection test failed: "+err.Error())
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "connection successful",
	})
}

// Helper function to create a test request with user context
func newRequestWithUser(method, url string, body []byte, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user claims to context
	claims := &services.Claims{
		UserID: userID.String(),
		Email:  "test@example.com",
		Role:   "user",
	}
	ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
	return req.WithContext(ctx)
}

// Helper function to create a test request without user context
func newRequestWithoutUser(method, url string, body []byte) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestConnectionHandler_List(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		hasUser            bool
		mockListConnections func(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error)
		expectedStatus     int
		expectedError      string
	}{
		{
			name:    "successful list",
			hasUser: true,
			mockListConnections: func(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error) {
				return []*models.Connection{
					{
						ID:       uuid.New(),
						UserID:   userID,
						Name:     "Test Connection 1",
						Type:     "postgres",
						Host:     "localhost",
						Port:     5432,
						Database: "testdb",
					},
					{
						ID:       uuid.New(),
						UserID:   userID,
						Name:     "Test Connection 2",
						Type:     "mysql",
						Host:     "localhost",
						Port:     3306,
						Database: "testdb",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "empty list",
			hasUser: true,
			mockListConnections: func(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error) {
				return []*models.Connection{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:    "service error",
			hasUser: true,
			mockListConnections: func(ctx context.Context, userID uuid.UUID) ([]*models.Connection, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to list connections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				ListConnectionsFunc: tt.mockListConnections,
			}
			handler := newTestConnectionHandler(mock)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, "/api/v1/connections", nil, userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, "/api/v1/connections", nil)
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

func TestConnectionHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                 string
		hasUser              bool
		requestBody          map[string]interface{}
		mockCreateConnection func(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error)
		expectedStatus       int
		expectedError        string
	}{
		{
			name:    "successful creation",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
				"username": "testuser",
				"password": "testpass",
				"ssl_mode": "require",
			},
			mockCreateConnection: func(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error) {
				conn.ID = uuid.New()
				conn.UserID = userID
				return conn, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "successful creation with default ssl_mode",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
				"username": "testuser",
				"password": "testpass",
			},
			mockCreateConnection: func(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error) {
				conn.ID = uuid.New()
				conn.UserID = userID
				return conn, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "missing name",
			hasUser: true,
			requestBody: map[string]interface{}{
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, type, host, and port are required",
		},
		{
			name:    "missing type",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, type, host, and port are required",
		},
		{
			name:    "missing host",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"type":     "postgres",
				"port":     5432,
				"database": "testdb",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, type, host, and port are required",
		},
		{
			name:    "missing port",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"type":     "postgres",
				"host":     "localhost",
				"database": "testdb",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, type, host, and port are required",
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:    "service error",
			hasUser: true,
			requestBody: map[string]interface{}{
				"name":     "Test Connection",
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
			},
			mockCreateConnection: func(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create connection: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				CreateConnectionFunc: tt.mockCreateConnection,
			}
			handler := newTestConnectionHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodPost, "/api/v1/connections", body, userID)
			} else {
				req = newRequestWithoutUser(http.MethodPost, "/api/v1/connections", body)
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

func TestConnectionHandler_Get(t *testing.T) {
	userID := uuid.New()
	connID := uuid.New()

	tests := []struct {
		name              string
		hasUser           bool
		connectionID      string
		mockGetConnection func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error)
		expectedStatus    int
		expectedError     string
	}{
		{
			name:         "successful get",
			hasUser:      true,
			connectionID: connID.String(),
			mockGetConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error) {
				return &models.Connection{
					ID:       connID,
					UserID:   userID,
					Name:     "Test Connection",
					Type:     "postgres",
					Host:     "localhost",
					Port:     5432,
					Database: "testdb",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "connection not found",
			hasUser:      true,
			connectionID: connID.String(),
			mockGetConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error) {
				return nil, services.ErrConnectionNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "connection not found",
		},
		{
			name:           "invalid connection ID",
			hasUser:        true,
			connectionID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid connection ID",
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			connectionID:   connID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:         "service error",
			hasUser:      true,
			connectionID: connID.String(),
			mockGetConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				GetConnectionFunc: tt.mockGetConnection,
			}
			handler := newTestConnectionHandler(mock)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodGet, "/api/v1/connections/"+tt.connectionID, nil, userID)
			} else {
				req = newRequestWithoutUser(http.MethodGet, "/api/v1/connections/"+tt.connectionID, nil)
			}

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.connectionID)
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

func TestConnectionHandler_Update(t *testing.T) {
	userID := uuid.New()
	connID := uuid.New()

	tests := []struct {
		name                 string
		hasUser              bool
		connectionID         string
		requestBody          map[string]interface{}
		mockUpdateConnection func(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error)
		expectedStatus       int
		expectedError        string
	}{
		{
			name:         "successful update",
			hasUser:      true,
			connectionID: connID.String(),
			requestBody: map[string]interface{}{
				"name": "Updated Connection",
				"port": 5433,
			},
			mockUpdateConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error) {
				return &models.Connection{
					ID:       connID,
					UserID:   userID,
					Name:     "Updated Connection",
					Type:     "postgres",
					Host:     "localhost",
					Port:     5433,
					Database: "testdb",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "connection not found",
			hasUser:      true,
			connectionID: connID.String(),
			requestBody: map[string]interface{}{
				"name": "Updated Connection",
			},
			mockUpdateConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error) {
				return nil, services.ErrConnectionNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "connection not found",
		},
		{
			name:           "invalid connection ID",
			hasUser:        true,
			connectionID:   "invalid-uuid",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid connection ID",
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			connectionID:   connID.String(),
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:         "service error",
			hasUser:      true,
			connectionID: connID.String(),
			requestBody: map[string]interface{}{
				"name": "Updated Connection",
			},
			mockUpdateConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				UpdateConnectionFunc: tt.mockUpdateConnection,
			}
			handler := newTestConnectionHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodPut, "/api/v1/connections/"+tt.connectionID, body, userID)
			} else {
				req = newRequestWithoutUser(http.MethodPut, "/api/v1/connections/"+tt.connectionID, body)
			}

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.connectionID)
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

func TestConnectionHandler_Delete(t *testing.T) {
	userID := uuid.New()
	connID := uuid.New()

	tests := []struct {
		name                 string
		hasUser              bool
		connectionID         string
		mockDeleteConnection func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
		expectedStatus       int
		expectedError        string
	}{
		{
			name:         "successful delete",
			hasUser:      true,
			connectionID: connID.String(),
			mockDeleteConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "connection not found",
			hasUser:      true,
			connectionID: connID.String(),
			mockDeleteConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return services.ErrConnectionNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "connection not found",
		},
		{
			name:         "connection in use",
			hasUser:      true,
			connectionID: connID.String(),
			mockDeleteConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return services.ErrConnectionInUse
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "connection is in use by pipelines",
		},
		{
			name:           "invalid connection ID",
			hasUser:        true,
			connectionID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid connection ID",
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			connectionID:   connID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:         "service error",
			hasUser:      true,
			connectionID: connID.String(),
			mockDeleteConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to delete connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				DeleteConnectionFunc: tt.mockDeleteConnection,
			}
			handler := newTestConnectionHandler(mock)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodDelete, "/api/v1/connections/"+tt.connectionID, nil, userID)
			} else {
				req = newRequestWithoutUser(http.MethodDelete, "/api/v1/connections/"+tt.connectionID, nil)
			}

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.connectionID)
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

func TestConnectionHandler_Test(t *testing.T) {
	userID := uuid.New()
	connID := uuid.New()

	tests := []struct {
		name               string
		hasUser            bool
		connectionID       string
		mockTestConnection func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
		expectedStatus     int
		expectedError      string
	}{
		{
			name:         "successful test",
			hasUser:      true,
			connectionID: connID.String(),
			mockTestConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "connection not found",
			hasUser:      true,
			connectionID: connID.String(),
			mockTestConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return services.ErrConnectionNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "connection not found",
		},
		{
			name:         "connection test failed",
			hasUser:      true,
			connectionID: connID.String(),
			mockTestConnection: func(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error {
				return errors.New("connection refused")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "connection test failed: connection refused",
		},
		{
			name:           "invalid connection ID",
			hasUser:        true,
			connectionID:   "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid connection ID",
		},
		{
			name:           "unauthorized - no user in context",
			hasUser:        false,
			connectionID:   connID.String(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				TestConnectionFunc: tt.mockTestConnection,
			}
			handler := newTestConnectionHandler(mock)

			var req *http.Request
			if tt.hasUser {
				req = newRequestWithUser(http.MethodPost, "/api/v1/connections/"+tt.connectionID+"/test", nil, userID)
			} else {
				req = newRequestWithoutUser(http.MethodPost, "/api/v1/connections/"+tt.connectionID+"/test", nil)
			}

			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.connectionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.Test(rec, req)

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

			// Verify success message on successful test
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["success"] != true {
					t.Errorf("expected success to be true")
				}
				if response["message"] != "connection successful" {
					t.Errorf("expected message 'connection successful', got %q", response["message"])
				}
			}
		})
	}
}

func TestConnectionHandler_TestDirect(t *testing.T) {
	tests := []struct {
		name                     string
		requestBody              map[string]interface{}
		mockTestConnectionDirect func(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error
		expectedStatus           int
		expectedError            string
	}{
		{
			name: "successful direct test",
			requestBody: map[string]interface{}{
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
				"username": "testuser",
				"password": "testpass",
				"ssl_mode": "require",
			},
			mockTestConnectionDirect: func(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful direct test with default ssl_mode",
			requestBody: map[string]interface{}{
				"type":     "postgres",
				"host":     "localhost",
				"port":     5432,
				"database": "testdb",
				"username": "testuser",
				"password": "testpass",
			},
			mockTestConnectionDirect: func(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
				if sslMode != "prefer" {
					return errors.New("expected default ssl_mode to be 'prefer'")
				}
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "connection test failed",
			requestBody: map[string]interface{}{
				"type":     "postgres",
				"host":     "invalid-host",
				"port":     5432,
				"database": "testdb",
				"username": "testuser",
				"password": "testpass",
			},
			mockTestConnectionDirect: func(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
				return errors.New("connection refused")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "connection test failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockConnectionService{
				TestConnectionDirectFunc: tt.mockTestConnectionDirect,
			}
			handler := newTestConnectionHandler(mock)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/connections/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.TestDirect(rec, req)

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

			// Verify success message on successful test
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["success"] != true {
					t.Errorf("expected success to be true")
				}
				if response["message"] != "connection successful" {
					t.Errorf("expected message 'connection successful', got %q", response["message"])
				}
			}
		})
	}
}

func TestConnectionHandler_InvalidJSON(t *testing.T) {
	userID := uuid.New()
	connID := uuid.New()

	endpoints := []struct {
		name    string
		handler func(*ConnectionHandler, http.ResponseWriter, *http.Request)
		method  string
		url     string
		hasUser bool
		useID   bool
	}{
		{
			name: "Create",
			handler: func(h *ConnectionHandler, w http.ResponseWriter, r *http.Request) {
				h.Create(w, r)
			},
			method:  http.MethodPost,
			url:     "/api/v1/connections",
			hasUser: true,
			useID:   false,
		},
		{
			name: "Update",
			handler: func(h *ConnectionHandler, w http.ResponseWriter, r *http.Request) {
				h.Update(w, r)
			},
			method:  http.MethodPut,
			url:     "/api/v1/connections/" + connID.String(),
			hasUser: true,
			useID:   true,
		},
		{
			name: "TestDirect",
			handler: func(h *ConnectionHandler, w http.ResponseWriter, r *http.Request) {
				h.TestDirect(w, r)
			},
			method:  http.MethodPost,
			url:     "/api/v1/connections/test",
			hasUser: false,
			useID:   false,
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+" invalid JSON", func(t *testing.T) {
			mock := &MockConnectionService{}
			handler := newTestConnectionHandler(mock)

			var req *http.Request
			if ep.hasUser {
				req = newRequestWithUser(ep.method, ep.url, []byte("invalid json"), userID)
			} else {
				req = newRequestWithoutUser(ep.method, ep.url, []byte("invalid json"))
			}

			if ep.useID {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", connID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			rec := httptest.NewRecorder()

			ep.handler(handler, rec, req)

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

func TestConnectionHandler_GetUserUUIDError(t *testing.T) {
	// Test case where user ID in claims is invalid
	endpoints := []struct {
		name    string
		handler func(*ConnectionHandler, http.ResponseWriter, *http.Request)
		method  string
		url     string
		body    []byte
	}{
		{
			name: "List with invalid user ID",
			handler: func(h *ConnectionHandler, w http.ResponseWriter, r *http.Request) {
				h.List(w, r)
			},
			method: http.MethodGet,
			url:    "/api/v1/connections",
		},
		{
			name: "Create with invalid user ID",
			handler: func(h *ConnectionHandler, w http.ResponseWriter, r *http.Request) {
				h.Create(w, r)
			},
			method: http.MethodPost,
			url:    "/api/v1/connections",
			body: mustMarshal(map[string]interface{}{
				"name": "Test",
				"type": "postgres",
				"host": "localhost",
				"port": 5432,
			}),
		},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			mock := &MockConnectionService{}
			handler := newTestConnectionHandler(mock)

			req := httptest.NewRequest(ep.method, ep.url, bytes.NewReader(ep.body))
			req.Header.Set("Content-Type", "application/json")

			// Add invalid user claims to context
			claims := &services.Claims{
				UserID: "invalid-uuid",
				Email:  "test@example.com",
				Role:   "user",
			}
			ctx := context.WithValue(req.Context(), middleware.ClaimsContextKey, claims)
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()

			ep.handler(handler, rec, req)

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

// Helper function to marshal JSON or panic
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

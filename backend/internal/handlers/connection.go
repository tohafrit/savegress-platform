package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// ConnectionServiceInterface defines the interface for connection service operations
type ConnectionServiceInterface interface {
	ListConnections(ctx context.Context, userID uuid.UUID) ([]models.Connection, error)
	CreateConnection(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error)
	GetConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) (*models.Connection, error)
	UpdateConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error)
	DeleteConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
	TestConnection(ctx context.Context, userID uuid.UUID, connID uuid.UUID) error
	TestConnectionDirect(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error
}

// ConnectionHandler handles connection endpoints
type ConnectionHandler struct {
	connectionService ConnectionServiceInterface
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(connectionService *services.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{connectionService: connectionService}
}

// NewConnectionHandlerWithInterface creates a new connection handler with interface (for testing)
func NewConnectionHandlerWithInterface(connectionService ConnectionServiceInterface) *ConnectionHandler {
	return &ConnectionHandler{connectionService: connectionService}
}

// List returns all connections for the user
func (h *ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
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

	connections, err := h.connectionService.ListConnections(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}

	respondSuccess(w, map[string]interface{}{"connections": connections})
}

// Create creates a new connection
func (h *ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	created, err := h.connectionService.CreateConnection(r.Context(), userID, conn)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create connection: "+err.Error())
		return
	}

	respondCreated(w, created)
}

// Get returns a specific connection
func (h *ConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	conn, err := h.connectionService.GetConnection(r.Context(), userID, connID)
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

// Update updates a connection
func (h *ConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	conn, err := h.connectionService.UpdateConnection(r.Context(), userID, connID, updates)
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

// Delete deletes a connection
func (h *ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err = h.connectionService.DeleteConnection(r.Context(), userID, connID)
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

// Test tests a connection
func (h *ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
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

	err = h.connectionService.TestConnection(r.Context(), userID, connID)
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

// TestDirect tests a connection without saving it
func (h *ConnectionHandler) TestDirect(w http.ResponseWriter, r *http.Request) {
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

	err := h.connectionService.TestConnectionDirect(r.Context(), req.Type, req.Host, req.Port, req.Database, req.Username, req.Password, req.SSLMode)
	if err != nil {
		respondError(w, http.StatusBadRequest, "connection test failed: "+err.Error())
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "connection successful",
	})
}

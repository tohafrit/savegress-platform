package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// PipelineHandler handles pipeline endpoints
type PipelineHandler struct {
	pipelineService *services.PipelineService
	licenseService  *services.LicenseService
}

// NewPipelineHandler creates a new pipeline handler
func NewPipelineHandler(pipelineService *services.PipelineService, licenseService *services.LicenseService) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
		licenseService:  licenseService,
	}
}

// List returns all pipelines for the user
func (h *PipelineHandler) List(w http.ResponseWriter, r *http.Request) {
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

	pipelines, err := h.pipelineService.ListPipelines(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list pipelines")
		return
	}

	respondSuccess(w, map[string]interface{}{"pipelines": pipelines})
}

// Create creates a new pipeline
func (h *PipelineHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	licenses, err := h.licenseService.GetUserLicenses(r.Context(), userID)
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

		currentCount, _ := h.pipelineService.CountUserPipelines(r.Context(), userID)
		if currentCount >= maxPipelines {
			respondError(w, http.StatusForbidden, "pipeline limit reached for your plan")
			return
		}
	}

	created, err := h.pipelineService.CreatePipeline(r.Context(), userID, pipeline)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create pipeline: "+err.Error())
		return
	}

	respondCreated(w, created)
}

// Get returns a specific pipeline
func (h *PipelineHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	pipeline, err := h.pipelineService.GetPipelineWithConnections(r.Context(), userID, pipelineID)
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

// Update updates a pipeline
func (h *PipelineHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	pipeline, err := h.pipelineService.UpdatePipeline(r.Context(), userID, pipelineID, updates)
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

// Delete deletes a pipeline
func (h *PipelineHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	err = h.pipelineService.DeletePipeline(r.Context(), userID, pipelineID)
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

// GetMetrics returns metrics for a pipeline
func (h *PipelineHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
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

	metrics, err := h.pipelineService.GetPipelineMetrics(r.Context(), userID, pipelineID, hours)
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

// GetLogs returns logs for a pipeline
func (h *PipelineHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
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

	logs, err := h.pipelineService.GetPipelineLogs(r.Context(), userID, pipelineID, limit, level)
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

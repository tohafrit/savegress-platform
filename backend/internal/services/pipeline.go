package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

var (
	ErrPipelineNotFound = errors.New("pipeline not found")
	ErrPipelineLimitReached = errors.New("pipeline limit reached for your plan")
)

// PipelineService handles pipeline management
type PipelineService struct {
	db *repository.PostgresDB
}

// NewPipelineService creates a new pipeline service
func NewPipelineService(db *repository.PostgresDB) *PipelineService {
	return &PipelineService{db: db}
}

// CreatePipeline creates a new pipeline
func (s *PipelineService) CreatePipeline(ctx context.Context, userID uuid.UUID, pipeline *models.Pipeline) (*models.Pipeline, error) {
	pipeline.ID = uuid.New()
	pipeline.UserID = userID
	pipeline.Status = "created"
	pipeline.CreatedAt = time.Now().UTC()
	pipeline.UpdatedAt = pipeline.CreatedAt

	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO pipelines (id, user_id, name, description, source_connection_id, target_connection_id, target_type, target_config, tables, status, license_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, pipeline.ID, pipeline.UserID, pipeline.Name, pipeline.Description, pipeline.SourceConnID,
		pipeline.TargetConnID, pipeline.TargetType, pipeline.TargetConfig, pipeline.Tables,
		pipeline.Status, pipeline.LicenseID, pipeline.CreatedAt, pipeline.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	return pipeline, nil
}

// GetPipeline retrieves a pipeline by ID
func (s *PipelineService) GetPipeline(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
	var p models.Pipeline
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, name, description, source_connection_id, target_connection_id, target_type, target_config, tables, status, license_id, hardware_id,
			events_processed, bytes_processed, current_lag_ms, last_event_at, error_message, created_at, updated_at
		FROM pipelines WHERE id = $1 AND user_id = $2
	`, pipelineID, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.SourceConnID, &p.TargetConnID, &p.TargetType,
		&p.TargetConfig, &p.Tables, &p.Status, &p.LicenseID, &p.HardwareID,
		&p.EventsProcessed, &p.BytesProcessed, &p.CurrentLag, &p.LastEventAt, &p.ErrorMessage,
		&p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrPipelineNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPipelineWithConnections retrieves a pipeline with its connections
func (s *PipelineService) GetPipelineWithConnections(ctx context.Context, userID, pipelineID uuid.UUID) (*models.Pipeline, error) {
	p, err := s.GetPipeline(ctx, userID, pipelineID)
	if err != nil {
		return nil, err
	}

	// Get source connection
	var source models.Connection
	err = s.db.Pool().QueryRow(ctx, `
		SELECT id, name, type, host, port, database, username, ssl_mode
		FROM connections WHERE id = $1
	`, p.SourceConnID).Scan(&source.ID, &source.Name, &source.Type, &source.Host, &source.Port, &source.Database, &source.Username, &source.SSLMode)
	if err == nil {
		p.SourceConnection = &source
	}

	// Get target connection if exists
	if p.TargetConnID != nil {
		var target models.Connection
		err = s.db.Pool().QueryRow(ctx, `
			SELECT id, name, type, host, port, database, username, ssl_mode
			FROM connections WHERE id = $1
		`, p.TargetConnID).Scan(&target.ID, &target.Name, &target.Type, &target.Host, &target.Port, &target.Database, &target.Username, &target.SSLMode)
		if err == nil {
			p.TargetConnection = &target
		}
	}

	return p, nil
}

// ListPipelines returns all pipelines for a user
func (s *PipelineService) ListPipelines(ctx context.Context, userID uuid.UUID) ([]models.Pipeline, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT p.id, p.user_id, p.name, p.description, p.source_connection_id, p.target_connection_id, p.target_type, p.tables, p.status,
			p.events_processed, p.bytes_processed, p.current_lag_ms, p.last_event_at, p.error_message, p.created_at, p.updated_at,
			sc.name as source_name, sc.type as source_type
		FROM pipelines p
		LEFT JOIN connections sc ON p.source_connection_id = sc.id
		WHERE p.user_id = $1 ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pipelines := make([]models.Pipeline, 0)
	for rows.Next() {
		var p models.Pipeline
		var sourceName, sourceType *string
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.SourceConnID, &p.TargetConnID, &p.TargetType, &p.Tables, &p.Status,
			&p.EventsProcessed, &p.BytesProcessed, &p.CurrentLag, &p.LastEventAt, &p.ErrorMessage, &p.CreatedAt, &p.UpdatedAt,
			&sourceName, &sourceType)
		if err != nil {
			return nil, err
		}
		if sourceName != nil && sourceType != nil {
			p.SourceConnection = &models.Connection{Name: *sourceName, Type: *sourceType}
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, nil
}

// UpdatePipeline updates a pipeline
func (s *PipelineService) UpdatePipeline(ctx context.Context, userID, pipelineID uuid.UUID, updates map[string]interface{}) (*models.Pipeline, error) {
	p, err := s.GetPipeline(ctx, userID, pipelineID)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		p.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		p.Description = desc
	}
	if tables, ok := updates["tables"].([]interface{}); ok {
		p.Tables = make([]string, len(tables))
		for i, t := range tables {
			p.Tables[i] = t.(string)
		}
	}
	if targetType, ok := updates["target_type"].(string); ok {
		p.TargetType = targetType
	}
	if targetConfig, ok := updates["target_config"].(map[string]interface{}); ok {
		p.TargetConfig = make(map[string]string)
		for k, v := range targetConfig {
			p.TargetConfig[k] = fmt.Sprintf("%v", v)
		}
	}

	p.UpdatedAt = time.Now().UTC()

	_, err = s.db.Pool().Exec(ctx, `
		UPDATE pipelines SET name = $1, description = $2, tables = $3, target_type = $4, target_config = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8
	`, p.Name, p.Description, p.Tables, p.TargetType, p.TargetConfig, p.UpdatedAt, pipelineID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update pipeline: %w", err)
	}

	return p, nil
}

// UpdatePipelineStatus updates only the status of a pipeline
func (s *PipelineService) UpdatePipelineStatus(ctx context.Context, pipelineID uuid.UUID, status string, errorMsg string) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE pipelines SET status = $1, error_message = $2, updated_at = $3 WHERE id = $4
	`, status, errorMsg, time.Now().UTC(), pipelineID)
	return err
}

// UpdatePipelineStats updates runtime statistics from telemetry
func (s *PipelineService) UpdatePipelineStats(ctx context.Context, pipelineID uuid.UUID, eventsProcessed, bytesProcessed, lagMs int64) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE pipelines SET events_processed = $1, bytes_processed = $2, current_lag_ms = $3, last_event_at = $4, updated_at = $5
		WHERE id = $6
	`, eventsProcessed, bytesProcessed, lagMs, time.Now().UTC(), time.Now().UTC(), pipelineID)
	return err
}

// DeletePipeline deletes a pipeline
func (s *PipelineService) DeletePipeline(ctx context.Context, userID, pipelineID uuid.UUID) error {
	result, err := s.db.Pool().Exec(ctx, `DELETE FROM pipelines WHERE id = $1 AND user_id = $2`, pipelineID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPipelineNotFound
	}
	return nil
}

// GetPipelineLogs retrieves logs for a pipeline
func (s *PipelineService) GetPipelineLogs(ctx context.Context, userID, pipelineID uuid.UUID, limit int, level string) ([]models.PipelineLog, error) {
	// Verify ownership
	_, err := s.GetPipeline(ctx, userID, pipelineID)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, pipeline_id, level, message, details, timestamp FROM pipeline_logs WHERE pipeline_id = $1`
	args := []interface{}{pipelineID}

	if level != "" {
		query += " AND level = $2"
		args = append(args, level)
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.PipelineLog, 0)
	for rows.Next() {
		var l models.PipelineLog
		err := rows.Scan(&l.ID, &l.PipelineID, &l.Level, &l.Message, &l.Details, &l.Timestamp)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// AddPipelineLog adds a log entry for a pipeline
func (s *PipelineService) AddPipelineLog(ctx context.Context, pipelineID uuid.UUID, level, message string, details map[string]interface{}) error {
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO pipeline_logs (id, pipeline_id, level, message, details, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), pipelineID, level, message, details, time.Now().UTC())
	return err
}

// GetPipelineMetrics retrieves metrics for a pipeline
func (s *PipelineService) GetPipelineMetrics(ctx context.Context, userID, pipelineID uuid.UUID, hours int) ([]map[string]interface{}, error) {
	// Verify ownership
	_, err := s.GetPipeline(ctx, userID, pipelineID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT
			date_trunc('hour', t.timestamp) as hour,
			SUM(t.events_processed) as events,
			SUM(t.bytes_processed) as bytes,
			AVG(t.avg_latency_ms) as latency,
			SUM(t.error_count) as errors
		FROM telemetry t
		WHERE t.pipeline_id = $1 AND t.timestamp > NOW() - ($2 || ' hours')::INTERVAL
		GROUP BY hour
		ORDER BY hour
	`, pipelineID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]map[string]interface{}, 0)
	for rows.Next() {
		var hour time.Time
		var events, bytes, errors int64
		var latency float64
		if err := rows.Scan(&hour, &events, &bytes, &latency, &errors); err != nil {
			return nil, err
		}
		metrics = append(metrics, map[string]interface{}{
			"timestamp": hour,
			"events":    events,
			"bytes":     bytes,
			"latency":   latency,
			"errors":    errors,
		})
	}
	return metrics, nil
}

// CountUserPipelines counts pipelines for a user
func (s *PipelineService) CountUserPipelines(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := s.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM pipelines WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

// BindPipelineToLicense associates a pipeline with a license
func (s *PipelineService) BindPipelineToLicense(ctx context.Context, pipelineID, licenseID uuid.UUID, hardwareID string) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE pipelines SET license_id = $1, hardware_id = $2 WHERE id = $3
	`, licenseID, hardwareID, pipelineID)
	return err
}

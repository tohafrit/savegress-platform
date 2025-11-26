package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
	ErrConnectionInUse    = errors.New("connection is used by pipelines")
	ErrConnectionTestFail = errors.New("connection test failed")
)

// ConnectionService handles database connection management
type ConnectionService struct {
	db            *repository.PostgresDB
	encryptionKey []byte
}

// NewConnectionService creates a new connection service
func NewConnectionService(db *repository.PostgresDB, encryptionKey string) *ConnectionService {
	key := make([]byte, 32)
	copy(key, []byte(encryptionKey))
	return &ConnectionService{
		db:            db,
		encryptionKey: key,
	}
}

// CreateConnection creates a new database connection
func (s *ConnectionService) CreateConnection(ctx context.Context, userID uuid.UUID, conn *models.Connection) (*models.Connection, error) {
	conn.ID = uuid.New()
	conn.UserID = userID
	conn.CreatedAt = time.Now().UTC()
	conn.UpdatedAt = conn.CreatedAt

	// Encrypt password
	encryptedPass, err := s.encryptPassword(conn.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO connections (id, user_id, name, type, host, port, database, username, password, ssl_mode, options, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, conn.ID, conn.UserID, conn.Name, conn.Type, conn.Host, conn.Port, conn.Database,
		conn.Username, encryptedPass, conn.SSLMode, conn.Options, conn.CreatedAt, conn.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	conn.Password = "" // Don't return password
	return conn, nil
}

// GetConnection retrieves a connection by ID
func (s *ConnectionService) GetConnection(ctx context.Context, userID, connID uuid.UUID) (*models.Connection, error) {
	var conn models.Connection
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, name, type, host, port, database, username, ssl_mode, options, last_tested_at, test_status, created_at, updated_at
		FROM connections WHERE id = $1 AND user_id = $2
	`, connID, userID).Scan(&conn.ID, &conn.UserID, &conn.Name, &conn.Type, &conn.Host, &conn.Port,
		&conn.Database, &conn.Username, &conn.SSLMode, &conn.Options, &conn.LastTestedAt, &conn.TestStatus,
		&conn.CreatedAt, &conn.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// ListConnections returns all connections for a user
func (s *ConnectionService) ListConnections(ctx context.Context, userID uuid.UUID) ([]models.Connection, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, name, type, host, port, database, username, ssl_mode, options, last_tested_at, test_status, created_at, updated_at
		FROM connections WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := make([]models.Connection, 0)
	for rows.Next() {
		var conn models.Connection
		err := rows.Scan(&conn.ID, &conn.UserID, &conn.Name, &conn.Type, &conn.Host, &conn.Port,
			&conn.Database, &conn.Username, &conn.SSLMode, &conn.Options, &conn.LastTestedAt, &conn.TestStatus,
			&conn.CreatedAt, &conn.UpdatedAt)
		if err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

// UpdateConnection updates a connection
func (s *ConnectionService) UpdateConnection(ctx context.Context, userID, connID uuid.UUID, updates map[string]interface{}) (*models.Connection, error) {
	// Get existing connection
	conn, err := s.GetConnection(ctx, userID, connID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		conn.Name = name
	}
	if host, ok := updates["host"].(string); ok {
		conn.Host = host
	}
	if port, ok := updates["port"].(float64); ok {
		conn.Port = int(port)
	}
	if database, ok := updates["database"].(string); ok {
		conn.Database = database
	}
	if username, ok := updates["username"].(string); ok {
		conn.Username = username
	}
	if sslMode, ok := updates["ssl_mode"].(string); ok {
		conn.SSLMode = sslMode
	}

	conn.UpdatedAt = time.Now().UTC()

	// Handle password update separately
	passwordUpdate := ""
	if password, ok := updates["password"].(string); ok && password != "" {
		encryptedPass, err := s.encryptPassword(password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		passwordUpdate = encryptedPass
	}

	if passwordUpdate != "" {
		_, err = s.db.Pool().Exec(ctx, `
			UPDATE connections SET name = $1, host = $2, port = $3, database = $4, username = $5, password = $6, ssl_mode = $7, updated_at = $8
			WHERE id = $9 AND user_id = $10
		`, conn.Name, conn.Host, conn.Port, conn.Database, conn.Username, passwordUpdate, conn.SSLMode, conn.UpdatedAt, connID, userID)
	} else {
		_, err = s.db.Pool().Exec(ctx, `
			UPDATE connections SET name = $1, host = $2, port = $3, database = $4, username = $5, ssl_mode = $6, updated_at = $7
			WHERE id = $8 AND user_id = $9
		`, conn.Name, conn.Host, conn.Port, conn.Database, conn.Username, conn.SSLMode, conn.UpdatedAt, connID, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update connection: %w", err)
	}

	return conn, nil
}

// DeleteConnection deletes a connection
func (s *ConnectionService) DeleteConnection(ctx context.Context, userID, connID uuid.UUID) error {
	// Check if connection is in use
	var count int
	err := s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM pipelines WHERE (source_connection_id = $1 OR target_connection_id = $1)
	`, connID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrConnectionInUse
	}

	_, err = s.db.Pool().Exec(ctx, `DELETE FROM connections WHERE id = $1 AND user_id = $2`, connID, userID)
	return err
}

// TestConnection tests a database connection
func (s *ConnectionService) TestConnection(ctx context.Context, userID, connID uuid.UUID) error {
	// Get connection with password
	var conn models.Connection
	var encryptedPass string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, type, host, port, database, username, password, ssl_mode
		FROM connections WHERE id = $1 AND user_id = $2
	`, connID, userID).Scan(&conn.ID, &conn.Type, &conn.Host, &conn.Port, &conn.Database, &conn.Username, &encryptedPass, &conn.SSLMode)
	if err == pgx.ErrNoRows {
		return ErrConnectionNotFound
	}
	if err != nil {
		return err
	}

	// Decrypt password
	password, err := s.decryptPassword(encryptedPass)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Test connection
	testErr := s.doConnectionTest(ctx, conn.Type, conn.Host, conn.Port, conn.Database, conn.Username, password, conn.SSLMode)

	// Update test status
	now := time.Now().UTC()
	status := "success"
	if testErr != nil {
		status = "failed"
	}

	_, _ = s.db.Pool().Exec(ctx, `
		UPDATE connections SET last_tested_at = $1, test_status = $2 WHERE id = $3
	`, now, status, connID)

	return testErr
}

// TestConnectionDirect tests a connection without saving
func (s *ConnectionService) TestConnectionDirect(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
	return s.doConnectionTest(ctx, connType, host, port, database, username, password, sslMode)
}

func (s *ConnectionService) doConnectionTest(ctx context.Context, connType, host string, port int, database, username, password, sslMode string) error {
	// First, test TCP connectivity
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", address, err)
	}
	conn.Close()

	// Then test database-specific connection
	switch connType {
	case "postgres", "postgresql":
		return s.testPostgres(ctx, host, port, database, username, password, sslMode)
	case "mysql", "mariadb":
		return s.testMySQL(ctx, host, port, database, username, password, sslMode)
	default:
		// For other types, just TCP test is enough for now
		return nil
	}
}

func (s *ConnectionService) testPostgres(ctx context.Context, host string, port int, database, username, password, sslMode string) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		username, password, host, port, database, sslMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

func (s *ConnectionService) testMySQL(ctx context.Context, host string, port int, database, username, password, sslMode string) error {
	// MySQL DSN format: username:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// Encryption helpers
func (s *ConnectionService) encryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *ConnectionService) decryptPassword(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetConnectionWithPassword retrieves connection with decrypted password (internal use)
func (s *ConnectionService) GetConnectionWithPassword(ctx context.Context, connID uuid.UUID) (*models.Connection, error) {
	var conn models.Connection
	var encryptedPass string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, name, type, host, port, database, username, password, ssl_mode, options
		FROM connections WHERE id = $1
	`, connID).Scan(&conn.ID, &conn.UserID, &conn.Name, &conn.Type, &conn.Host, &conn.Port,
		&conn.Database, &conn.Username, &encryptedPass, &conn.SSLMode, &conn.Options)
	if err == pgx.ErrNoRows {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, err
	}

	password, err := s.decryptPassword(encryptedPass)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}
	conn.Password = password

	return &conn, nil
}

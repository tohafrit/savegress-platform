package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserExists           = errors.New("user with this email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrResetTokenExpired    = errors.New("password reset token has expired")
	ErrResetTokenUsed       = errors.New("password reset token has already been used")
)

// AuthService handles authentication
type AuthService struct {
	db        *repository.PostgresDB
	redis     *repository.RedisClient
	jwtSecret []byte
}

// NewAuthService creates a new auth service
func NewAuthService(db *repository.PostgresDB, redis *repository.RedisClient, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		redis:     redis,
		jwtSecret: []byte(jwtSecret),
	}
}

// TokenPair holds access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims holds JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GetUserUUID returns UserID as uuid.UUID
func (c *Claims) GetUserUUID() (uuid.UUID, error) {
	return uuid.Parse(c.UserID)
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, password, name, company string) (*models.User, *TokenPair, error) {
	// Check if user exists
	var exists bool
	err := s.db.Pool().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, nil, ErrUserExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Company:      company,
		Role:         "user",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, company, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.PasswordHash, user.Name, user.Company, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *TokenPair, error) {
	var user models.User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, email, password_hash, name, COALESCE(company, ''), role, email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Company, &user.Role, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Update last login
	now := time.Now().UTC()
	user.LastLoginAt = &now
	_, _ = s.db.Pool().Exec(ctx, "UPDATE users SET last_login_at = $1 WHERE id = $2", now, user.ID)

	// Generate tokens
	tokens, err := s.generateTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify refresh token exists and is valid
	var userID uuid.UUID
	var expiresAt time.Time
	var revokedAt *time.Time

	err := s.db.Pool().QueryRow(ctx, `
		SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1
	`, refreshToken).Scan(&userID, &expiresAt, &revokedAt)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if revokedAt != nil || time.Now().After(expiresAt) {
		return nil, ErrInvalidToken
	}

	// Get user
	var user models.User
	err = s.db.Pool().QueryRow(ctx, `
		SELECT id, email, name, COALESCE(company, ''), role FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &user.Company, &user.Role)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Revoke old refresh token
	_, _ = s.db.Pool().Exec(ctx, "UPDATE refresh_tokens SET revoked_at = $1 WHERE token = $2", time.Now().UTC(), refreshToken)

	// Generate new tokens
	return s.generateTokenPair(&user)
}

// ValidateToken validates an access token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, email, name, COALESCE(company, ''), role, email_verified, COALESCE(stripe_customer_id, ''), created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &user.Company, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// CreatePasswordResetToken creates a password reset token for a user
func (s *AuthService) CreatePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Find user by email
	var userID uuid.UUID
	err := s.db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		// Don't reveal if user exists - silently return
		return "", nil
	}

	// Generate secure token using UUIDs
	token := uuid.New().String() + uuid.New().String()
	token = token[:64] // 64 character token

	// Hash the token for storage (we'll verify with bcrypt)
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Store reset token (expires in 1 hour)
	resetID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, resetID, userID, string(tokenHash), expiresAt, time.Now().UTC())
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	// Return the raw token (to be sent via email)
	return token, nil
}

// ValidatePasswordResetToken validates a password reset token
func (s *AuthService) ValidatePasswordResetToken(ctx context.Context, token string) (*models.PasswordReset, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_resets
		WHERE expires_at > NOW() AND used_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, ErrInvalidToken
	}
	defer rows.Close()

	// Check each recent token
	for rows.Next() {
		var reset models.PasswordReset
		if err := rows.Scan(&reset.ID, &reset.UserID, &reset.Token, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt); err != nil {
			continue
		}

		// Compare with bcrypt
		if err := bcrypt.CompareHashAndPassword([]byte(reset.Token), []byte(token)); err == nil {
			return &reset, nil
		}
	}

	return nil, ErrInvalidToken
}

// ResetPassword resets a user's password using a valid token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate the token
	reset, err := s.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Check if already used
	if reset.UsedAt != nil {
		return ErrResetTokenUsed
	}

	// Check expiration
	if time.Now().After(reset.ExpiresAt) {
		return ErrResetTokenExpired
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password and mark token as used (in a transaction)
	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update password
	_, err = tx.Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3
	`, string(hash), time.Now().UTC(), reset.UserID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE password_resets SET used_at = $1 WHERE id = $2
	`, now, reset.ID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Revoke all refresh tokens for this user (force re-login)
	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL
	`, now, reset.UserID)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	return tx.Commit(ctx)
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, email, name, COALESCE(company, ''), role, email_verified, COALESCE(stripe_customer_id, ''), created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.Company, &user.Role, &user.EmailVerified, &user.StripeCustomerID, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	// Access token (15 minutes)
	accessExpiry := time.Now().Add(15 * time.Minute)
	accessClaims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "savegress",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token (7 days)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenID := uuid.New()
	refreshTokenString := refreshTokenID.String()

	// Store refresh token in database
	ctx := context.Background()
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, refreshTokenID, user.ID, refreshTokenString, refreshExpiry, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}

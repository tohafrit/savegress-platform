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
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
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

package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// NOTE: Database integration tests (Register, Login, RefreshToken, GetUserByID, etc.)
// would require:
// 1. A test database or proper mocking infrastructure
// 2. Refactoring services to accept database interfaces instead of concrete types
// 3. More complex test setup with transaction rollback
//
// The tests below focus on testing business logic that doesn't require database access

func TestAuthService_ValidateToken(t *testing.T) {
	service := &AuthService{
		jwtSecret: []byte("test-secret-key"),
	}

	userID := uuid.New()

	tests := []struct {
		name          string
		setupToken    func() string
		expectedError error
		expectedEmail string
	}{
		{
			name: "valid token",
			setupToken: func() string {
				claims := &Claims{
					UserID: userID.String(),
					Email:  "user@example.com",
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Issuer:    "savegress",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(service.jwtSecret)
				return tokenString
			},
			expectedError: nil,
			expectedEmail: "user@example.com",
		},
		{
			name: "expired token",
			setupToken: func() string {
				claims := &Claims{
					UserID: userID.String(),
					Email:  "user@example.com",
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						Issuer:    "savegress",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(service.jwtSecret)
				return tokenString
			},
			expectedError: ErrInvalidToken,
		},
		{
			name: "invalid signature",
			setupToken: func() string {
				claims := &Claims{
					UserID: userID.String(),
					Email:  "user@example.com",
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						Issuer:    "savegress",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret"))
				return tokenString
			},
			expectedError: ErrInvalidToken,
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "not.a.valid.jwt.token"
			},
			expectedError: ErrInvalidToken,
		},
		{
			name: "empty token",
			setupToken: func() string {
				return ""
			},
			expectedError: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			claims, err := service.ValidateToken(tokenString)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedEmail, claims.Email)
				assert.Equal(t, userID.String(), claims.UserID)
			}
		})
	}
}

func TestClaims_GetUserUUID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectedError bool
	}{
		{
			name:          "valid UUID",
			userID:        uuid.New().String(),
			expectedError: false,
		},
		{
			name:          "invalid UUID",
			userID:        "not-a-uuid",
			expectedError: true,
		},
		{
			name:          "empty UUID",
			userID:        "",
			expectedError: true,
		},
		{
			name:          "partial UUID",
			userID:        "123e4567-e89b",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				UserID: tt.userID,
			}

			userUUID, err := claims.GetUserUUID()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, userUUID.String())
			}
		})
	}
}

func TestPasswordHashing(t *testing.T) {
	// Test bcrypt password hashing (used internally by AuthService)
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "standard password",
			password: "MySecurePassword123!",
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatShouldStillWorkProperly123456789",
		},
		{
			name:     "special characters",
			password: "P@ssw0rd!#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify correct password
			err = bcrypt.CompareHashAndPassword(hash, []byte(tt.password))
			assert.NoError(t, err)

			// Verify incorrect password
			err = bcrypt.CompareHashAndPassword(hash, []byte("WrongPassword"))
			assert.Error(t, err)
		})
	}
}

func TestTokenPair_Structure(t *testing.T) {
	// Test TokenPair structure
	tokens := &TokenPair{
		AccessToken:  "access_token_value",
		RefreshToken: "refresh_token_value",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.True(t, tokens.ExpiresAt.After(time.Now()))
}

func TestClaims_Structure(t *testing.T) {
	// Test Claims structure
	userID := uuid.New()
	claims := &Claims{
		UserID: userID.String(),
		Email:  "user@example.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "savegress",
		},
	}

	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "savegress", claims.Issuer)
}

func TestAuthService_ErrorConstants(t *testing.T) {
	// Test that error constants are defined correctly
	assert.NotNil(t, ErrInvalidCredentials)
	assert.NotNil(t, ErrUserExists)
	assert.NotNil(t, ErrUserNotFound)
	assert.NotNil(t, ErrInvalidToken)
	assert.NotNil(t, ErrResetTokenExpired)
	assert.NotNil(t, ErrResetTokenUsed)

	assert.Equal(t, "invalid email or password", ErrInvalidCredentials.Error())
	assert.Equal(t, "user with this email already exists", ErrUserExists.Error())
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "invalid or expired token", ErrInvalidToken.Error())
	assert.Equal(t, "password reset token has expired", ErrResetTokenExpired.Error())
	assert.Equal(t, "password reset token has already been used", ErrResetTokenUsed.Error())
}

func TestNewAuthService(t *testing.T) {
	// Test service creation
	jwtSecret := "test-jwt-secret"
	service := &AuthService{
		jwtSecret: []byte(jwtSecret),
	}

	assert.NotNil(t, service)
	assert.Equal(t, []byte(jwtSecret), service.jwtSecret)
}

func TestJWTSigningMethods(t *testing.T) {
	service := &AuthService{
		jwtSecret: []byte("test-secret"),
	}

	// Create a token with HS256
	claims := &Claims{
		UserID: uuid.New().String(),
		Email:  "test@example.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(service.jwtSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate with correct secret
	parsedClaims, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, claims.Email, parsedClaims.Email)
}

func TestAuthService_SetEmailService(t *testing.T) {
	service := &AuthService{
		jwtSecret: []byte("test-secret"),
	}

	// Create a mock email service
	emailService := &EmailService{
		fromAddress: "noreply@example.com",
		baseURL:     "https://example.com",
	}

	service.SetEmailService(emailService)
	assert.Equal(t, emailService, service.emailService)
}

// Integration test examples (commented out - would need real database)
//
// func TestAuthService_RegisterIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test the full registration flow with a real database
// }
//
// func TestAuthService_LoginIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test the full login flow with a real database
// }
//
// func TestAuthService_PasswordResetIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test the full password reset flow with a real database
// }

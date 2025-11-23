package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/savegress/platform/backend/internal/services"
)

type contextKey string

const (
	UserContextKey   contextKey = "user"
	ClaimsContextKey contextKey = "claims"
)

// Auth middleware validates JWT tokens
func Auth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Extract Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Get user from database
			userUUID, err := claims.GetUserUUID()
			if err != nil {
				http.Error(w, `{"error": "invalid user id in token"}`, http.StatusUnauthorized)
				return
			}
			user, err := authService.GetUserByID(r.Context(), userUUID)
			if err != nil {
				http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
				return
			}

			// Add user and claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, ClaimsContextKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware ensures user is an admin
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*services.Claims)
		if !ok || claims.Role != "admin" {
			http.Error(w, `{"error": "admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext returns the user from context
func GetUserFromContext(ctx context.Context) *services.Claims {
	claims, _ := ctx.Value(ClaimsContextKey).(*services.Claims)
	return claims
}

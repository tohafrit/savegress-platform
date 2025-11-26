package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

// UserService handles user operations
type UserService struct {
	db *repository.PostgresDB
}

// NewUserService creates a new user service
func NewUserService(db *repository.PostgresDB) *UserService {
	return &UserService{db: db}
}

// GetByID returns a user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, email, name, COALESCE(company, ''), role, email_verified, COALESCE(stripe_customer_id, ''), created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.Company, &user.Role,
		&user.EmailVerified, &user.StripeCustomerID, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, name, company string) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE users SET name = $1, company = $2, updated_at = $3 WHERE id = $4
	`, name, company, time.Now().UTC(), userID)
	return err
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get current password hash
	var currentHash string
	err := s.db.Pool().QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3
	`, string(newHash), time.Now().UTC(), userID)
	return err
}

// SetStripeCustomerID sets the Stripe customer ID for a user
func (s *UserService) SetStripeCustomerID(ctx context.Context, userID uuid.UUID, customerID string) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE users SET stripe_customer_id = $1, updated_at = $2 WHERE id = $3
	`, customerID, time.Now().UTC(), userID)
	return err
}

// ListUsers returns paginated list of users (admin only)
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	var total int
	err := s.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, email, name, COALESCE(company, ''), role, email_verified, created_at, updated_at, last_login_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Company, &u.Role,
			&u.EmailVerified, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

// UpdateUserRole updates a user's role (admin only)
func (s *UserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	if role != "user" && role != "admin" {
		return fmt.Errorf("invalid role: %s", role)
	}
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE users SET role = $1, updated_at = $2 WHERE id = $3
	`, role, time.Now().UTC(), userID)
	return err
}

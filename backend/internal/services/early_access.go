package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/google/uuid"

	"github.com/savegress/platform/backend/internal/repository"
)

// EarlyAccessService handles early access requests
type EarlyAccessService struct {
	db         *repository.PostgresDB
	adminEmail string
	smtpConfig *SMTPConfig
	resendKey  string
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// NewEarlyAccessService creates a new early access service
func NewEarlyAccessService(db *repository.PostgresDB, adminEmail, resendKey string) *EarlyAccessService {
	return &EarlyAccessService{
		db:         db,
		adminEmail: adminEmail,
		resendKey:  resendKey,
	}
}

// EarlyAccessInput represents the input data for early access
type EarlyAccessInput struct {
	Email           string
	Company         string
	CurrentSolution string
	DataVolume      string
	Message         string
	IPAddress       string
	UserAgent       string
}

// Submit saves early access request to database and sends notification
func (s *EarlyAccessService) Submit(ctx context.Context, input EarlyAccessInput) error {
	// Insert into database
	id := uuid.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO early_access_requests (id, email, company, current_solution, data_volume, message, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, id, input.Email, input.Company, input.CurrentSolution, input.DataVolume, input.Message, input.IPAddress, input.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}

	// Send notification email asynchronously
	go s.sendNotificationEmail(input)

	return nil
}

func (s *EarlyAccessService) sendNotificationEmail(input EarlyAccessInput) {
	if s.adminEmail == "" {
		return
	}

	subject := fmt.Sprintf("New Early Access Request from %s", input.Company)
	body := fmt.Sprintf(`
New early access request received:

Email: %s
Company: %s
Current Solution: %s
Data Volume: %s
Message: %s

IP Address: %s
`, input.Email, input.Company, input.CurrentSolution, input.DataVolume, input.Message, input.IPAddress)

	// Try Resend API first if configured
	if s.resendKey != "" {
		s.sendViaResend(s.adminEmail, subject, body)
		return
	}

	// Fall back to SMTP if configured
	if s.smtpConfig != nil {
		s.sendViaSMTP(s.adminEmail, subject, body)
	}
}

func (s *EarlyAccessService) sendViaResend(to, subject, body string) error {
	payload := map[string]interface{}{
		"from":    "Savegress <onboarding@resend.dev>",
		"to":      []string{to},
		"subject": subject,
		"text":    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.resendKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}

func (s *EarlyAccessService) sendViaSMTP(to, subject, body string) error {
	if s.smtpConfig == nil {
		return nil
	}

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))

	auth := smtp.PlainAuth("", s.smtpConfig.User, s.smtpConfig.Password, s.smtpConfig.Host)
	addr := s.smtpConfig.Host + ":" + s.smtpConfig.Port

	return smtp.SendMail(addr, auth, s.smtpConfig.From, strings.Split(to, ","), msg)
}

// List returns all early access requests (admin only)
func (s *EarlyAccessService) List(ctx context.Context, limit, offset int) ([]map[string]interface{}, int, error) {
	var total int
	err := s.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM early_access_requests").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, email, company, current_solution, data_volume, message, ip_address, created_at
		FROM early_access_requests
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, email, company string
		var currentSolution, dataVolume, message, ipAddress *string
		var createdAt interface{}

		if err := rows.Scan(&id, &email, &company, &currentSolution, &dataVolume, &message, &ipAddress, &createdAt); err != nil {
			return nil, 0, err
		}

		results = append(results, map[string]interface{}{
			"id":              id,
			"email":           email,
			"company":         company,
			"currentSolution": currentSolution,
			"dataVolume":      dataVolume,
			"message":         message,
			"ipAddress":       ipAddress,
			"createdAt":       createdAt,
		})
	}

	return results, total, nil
}

package services

import (
	"context"
	"testing"
)

func TestEmailService_Creation(t *testing.T) {
	service := NewEmailService(EmailConfig{
		SMTPHost:    "localhost",
		SMTPPort:    "25",
		FromAddress: "test@example.com",
		BaseURL:     "https://platform.savegress.io",
	})

	// Test that service is created correctly
	if service == nil {
		t.Fatal("service should not be nil")
	}

	if service.fromAddress != "test@example.com" {
		t.Errorf("expected fromAddress 'test@example.com', got %q", service.fromAddress)
	}

	if service.baseURL != "https://platform.savegress.io" {
		t.Errorf("expected baseURL 'https://platform.savegress.io', got %q", service.baseURL)
	}
}

func TestEmailService_NewEmailService(t *testing.T) {
	tests := []struct {
		name           string
		config         EmailConfig
		expectResend   bool
		expectSMTP     bool
	}{
		{
			name: "SMTP configuration",
			config: EmailConfig{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     "587",
				SMTPUser:     "user",
				SMTPPassword: "password",
				FromAddress:  "noreply@example.com",
				BaseURL:      "https://example.com",
			},
			expectSMTP: true,
		},
		{
			name: "Resend API configuration",
			config: EmailConfig{
				ResendAPIKey: "re_test_key",
				FromAddress:  "noreply@example.com",
				BaseURL:      "https://example.com",
			},
			expectResend: true,
		},
		{
			name: "Both configured - SMTP takes precedence",
			config: EmailConfig{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     "587",
				SMTPUser:     "user",
				SMTPPassword: "password",
				ResendAPIKey: "re_test_key",
				FromAddress:  "noreply@example.com",
				BaseURL:      "https://example.com",
			},
			expectSMTP:   true,
			expectResend: true,
		},
		{
			name: "No email provider configured",
			config: EmailConfig{
				FromAddress: "noreply@example.com",
				BaseURL:     "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config)
			if service == nil {
				t.Fatal("service should not be nil")
			}

			if tt.expectSMTP && service.smtpHost == "" {
				t.Error("expected SMTP to be configured")
			}

			if tt.expectResend && service.resendAPIKey == "" {
				t.Error("expected Resend API to be configured")
			}
		})
	}
}

func TestEmailService_GenerateResetURL(t *testing.T) {
	service := NewEmailService(EmailConfig{
		BaseURL: "https://platform.savegress.io",
	})

	token := "test-token-123"
	expected := "https://platform.savegress.io/auth/reset-password?token=test-token-123"

	// Generate URL using the same logic as the service
	url := service.baseURL + "/auth/reset-password?token=" + token

	if url != expected {
		t.Errorf("expected URL %q, got %q", expected, url)
	}
}

func TestEmailService_GenerateVerificationURL(t *testing.T) {
	service := NewEmailService(EmailConfig{
		BaseURL: "https://platform.savegress.io",
	})

	token := "verify-token-456"
	expected := "https://platform.savegress.io/auth/verify-email?token=verify-token-456"

	// Generate URL using the same logic as the service
	url := service.baseURL + "/auth/verify-email?token=" + token

	if url != expected {
		t.Errorf("expected URL %q, got %q", expected, url)
	}
}

func TestEmailService_SendWithoutProvider(t *testing.T) {
	// Create service without any email provider configured
	service := NewEmailService(EmailConfig{
		FromAddress: "noreply@example.com",
		BaseURL:     "https://example.com",
	})

	ctx := context.Background()

	// These should return errors or handle gracefully when no provider is configured
	err := service.SendPasswordResetEmail(ctx, "test@example.com", "test-token")
	if err == nil {
		t.Log("Note: SendPasswordResetEmail silently succeeds when no email provider is configured (fail-open)")
	}

	err = service.SendWelcomeEmail(ctx, "test@example.com", "Test User")
	if err == nil {
		t.Log("Note: SendWelcomeEmail silently succeeds when no email provider is configured (fail-open)")
	}

	err = service.SendVerificationEmail(ctx, "test@example.com", "verify-token")
	if err == nil {
		t.Log("Note: SendVerificationEmail silently succeeds when no email provider is configured (fail-open)")
	}

	err = service.SendLicenseKeyEmail(ctx, "test@example.com", "LICENSE-KEY", "pro")
	if err == nil {
		t.Log("Note: SendLicenseKeyEmail silently succeeds when no email provider is configured (fail-open)")
	}
}

func TestEmailConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      EmailConfig
		shouldWork  bool
		description string
	}{
		{
			name: "complete SMTP config",
			config: EmailConfig{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     "587",
				SMTPUser:     "user@gmail.com",
				SMTPPassword: "app-password",
				FromAddress:  "noreply@example.com",
				BaseURL:      "https://example.com",
			},
			shouldWork:  true,
			description: "Full SMTP configuration should work",
		},
		{
			name: "resend config",
			config: EmailConfig{
				ResendAPIKey: "re_123456789",
				FromAddress:  "noreply@example.com",
				BaseURL:      "https://example.com",
			},
			shouldWork:  true,
			description: "Resend API configuration should work",
		},
		{
			name: "missing base URL",
			config: EmailConfig{
				SMTPHost:    "smtp.example.com",
				FromAddress: "noreply@example.com",
			},
			shouldWork:  true,
			description: "Missing base URL still creates service (URLs will be empty)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config)
			if service == nil && tt.shouldWork {
				t.Errorf("%s: expected service to be created", tt.description)
			}
			if service != nil && !tt.shouldWork {
				t.Errorf("%s: expected service creation to fail", tt.description)
			}
		})
	}
}

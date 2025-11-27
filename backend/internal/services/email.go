package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"time"
)

// EmailService handles sending emails
type EmailService struct {
	provider    EmailProvider
	fromAddress string
	fromName    string
	baseURL     string // For constructing links in emails
}

// EmailProvider abstracts different email providers
type EmailProvider interface {
	Send(ctx context.Context, to, subject, htmlBody, textBody string) error
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider    string // "smtp", "resend", "sendgrid"
	FromAddress string
	FromName    string
	BaseURL     string // e.g., "https://app.savegress.io"

	// SMTP settings
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string

	// Resend settings
	ResendAPIKey string

	// SendGrid settings
	SendGridAPIKey string
}

// NewEmailService creates a new email service
func NewEmailService(cfg EmailConfig) (*EmailService, error) {
	var provider EmailProvider

	switch cfg.Provider {
	case "resend":
		if cfg.ResendAPIKey == "" {
			return nil, fmt.Errorf("resend API key is required")
		}
		provider = &ResendProvider{apiKey: cfg.ResendAPIKey, fromAddress: cfg.FromAddress}
	case "sendgrid":
		if cfg.SendGridAPIKey == "" {
			return nil, fmt.Errorf("sendgrid API key is required")
		}
		provider = &SendGridProvider{apiKey: cfg.SendGridAPIKey, fromAddress: cfg.FromAddress}
	case "smtp":
		if cfg.SMTPHost == "" {
			return nil, fmt.Errorf("SMTP host is required")
		}
		provider = &SMTPProvider{
			host:     cfg.SMTPHost,
			port:     cfg.SMTPPort,
			user:     cfg.SMTPUser,
			password: cfg.SMTPPassword,
			from:     cfg.FromAddress,
		}
	default:
		// No-op provider for development
		provider = &NoOpProvider{}
	}

	return &EmailService{
		provider:    provider,
		fromAddress: cfg.FromAddress,
		fromName:    cfg.FromName,
		baseURL:     cfg.BaseURL,
	}, nil
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	subject := "Reset your Savegress password"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #0066cc; margin: 0;">Savegress</h1>
        </div>

        <h2 style="color: #333;">Reset Your Password</h2>

        <p>We received a request to reset the password for your Savegress account. Click the button below to create a new password:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #0066cc; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: 500;">Reset Password</a>
        </div>

        <p style="color: #666; font-size: 14px;">This link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email.</p>

        <p style="color: #666; font-size: 14px;">If the button doesn't work, copy and paste this URL into your browser:</p>
        <p style="color: #0066cc; font-size: 14px; word-break: break-all;">%s</p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; text-align: center;">
            © Savegress CDC Platform<br>
            This is an automated message, please do not reply.
        </p>
    </div>
</body>
</html>
`, resetURL, resetURL)

	textBody := fmt.Sprintf(`Reset Your Savegress Password

We received a request to reset the password for your Savegress account.

Click the following link to reset your password:
%s

This link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email.

---
Savegress CDC Platform
This is an automated message, please do not reply.
`, resetURL)

	return s.provider.Send(ctx, to, subject, htmlBody, textBody)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, to, name string) error {
	subject := "Welcome to Savegress!"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #0066cc; margin: 0;">Savegress</h1>
        </div>

        <h2 style="color: #333;">Welcome, %s!</h2>

        <p>Thank you for joining Savegress. We're excited to have you on board!</p>

        <p>With Savegress, you can:</p>
        <ul>
            <li>Capture real-time changes from 9+ databases</li>
            <li>Stream data with exactly-once delivery guarantees</li>
            <li>Use SIMD-accelerated compression for high throughput</li>
            <li>Deploy highly available clusters with Raft consensus</li>
        </ul>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s/dashboard" style="background-color: #0066cc; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: 500;">Go to Dashboard</a>
        </div>

        <p>Need help getting started? Check out our <a href="%s/docs">documentation</a> or reach out to our support team.</p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; text-align: center;">
            © Savegress CDC Platform
        </p>
    </div>
</body>
</html>
`, template.HTMLEscapeString(name), s.baseURL, s.baseURL)

	textBody := fmt.Sprintf(`Welcome to Savegress, %s!

Thank you for joining Savegress. We're excited to have you on board!

With Savegress, you can:
- Capture real-time changes from 9+ databases
- Stream data with exactly-once delivery guarantees
- Use SIMD-accelerated compression for high throughput
- Deploy highly available clusters with Raft consensus

Get started: %s/dashboard

Need help? Check out our documentation: %s/docs

---
Savegress CDC Platform
`, name, s.baseURL, s.baseURL)

	return s.provider.Send(ctx, to, subject, htmlBody, textBody)
}

// SendPaymentFailedEmail notifies user of payment failure
func (s *EmailService) SendPaymentFailedEmail(ctx context.Context, to, name string) error {
	subject := "Payment Failed - Action Required"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #0066cc; margin: 0;">Savegress</h1>
        </div>

        <h2 style="color: #cc0000;">Payment Failed</h2>

        <p>Hi %s,</p>

        <p>We were unable to process your recent payment for your Savegress subscription. Please update your payment method to avoid any interruption to your service.</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s/billing" style="background-color: #0066cc; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: 500;">Update Payment Method</a>
        </div>

        <p style="color: #666; font-size: 14px;">If you have any questions, please contact our support team.</p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; text-align: center;">
            © Savegress CDC Platform
        </p>
    </div>
</body>
</html>
`, template.HTMLEscapeString(name), s.baseURL)

	textBody := fmt.Sprintf(`Payment Failed - Action Required

Hi %s,

We were unable to process your recent payment for your Savegress subscription. Please update your payment method to avoid any interruption to your service.

Update your payment method: %s/billing

If you have any questions, please contact our support team.

---
Savegress CDC Platform
`, name, s.baseURL)

	return s.provider.Send(ctx, to, subject, htmlBody, textBody)
}

// SendSubscriptionCanceledEmail notifies user of subscription cancellation
func (s *EmailService) SendSubscriptionCanceledEmail(ctx context.Context, to, name string, endDate time.Time) error {
	subject := "Your Savegress Subscription Has Been Canceled"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="text-align: center; margin-bottom: 30px;">
            <h1 style="color: #0066cc; margin: 0;">Savegress</h1>
        </div>

        <h2>Subscription Canceled</h2>

        <p>Hi %s,</p>

        <p>Your Savegress subscription has been canceled. You will continue to have access to your current plan until <strong>%s</strong>.</p>

        <p>After this date, your account will be downgraded to the Community tier.</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s/billing" style="background-color: #0066cc; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: 500;">Reactivate Subscription</a>
        </div>

        <p style="color: #666; font-size: 14px;">We'd love to have you back! If you have any feedback on how we can improve, please let us know.</p>

        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

        <p style="color: #999; font-size: 12px; text-align: center;">
            © Savegress CDC Platform
        </p>
    </div>
</body>
</html>
`, template.HTMLEscapeString(name), endDate.Format("January 2, 2006"), s.baseURL)

	textBody := fmt.Sprintf(`Subscription Canceled

Hi %s,

Your Savegress subscription has been canceled. You will continue to have access to your current plan until %s.

After this date, your account will be downgraded to the Community tier.

Reactivate your subscription: %s/billing

We'd love to have you back! If you have any feedback on how we can improve, please let us know.

---
Savegress CDC Platform
`, name, endDate.Format("January 2, 2006"), s.baseURL)

	return s.provider.Send(ctx, to, subject, htmlBody, textBody)
}

// --- Email Providers ---

// ResendProvider sends emails via Resend API
type ResendProvider struct {
	apiKey      string
	fromAddress string
}

func (p *ResendProvider) Send(ctx context.Context, to, subject, htmlBody, textBody string) error {
	payload := map[string]interface{}{
		"from":    p.fromAddress,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
		"text":    textBody,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendGridProvider sends emails via SendGrid API
type SendGridProvider struct {
	apiKey      string
	fromAddress string
}

func (p *SendGridProvider) Send(ctx context.Context, to, subject, htmlBody, textBody string) error {
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": to}}},
		},
		"from":    map[string]string{"email": p.fromAddress},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": textBody},
			{"type": "text/html", "value": htmlBody},
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email API returned status %d", resp.StatusCode)
	}

	return nil
}

// SMTPProvider sends emails via SMTP
type SMTPProvider struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, htmlBody, textBody string) error {
	auth := smtp.PlainAuth("", p.user, p.password, p.host)

	// Build MIME message
	boundary := "boundary-savegress-email"
	message := fmt.Sprintf("From: %s\r\n", p.from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary)

	// Text part
	message += fmt.Sprintf("--%s\r\n", boundary)
	message += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
	message += textBody + "\r\n\r\n"

	// HTML part
	message += fmt.Sprintf("--%s\r\n", boundary)
	message += "Content-Type: text/html; charset=utf-8\r\n\r\n"
	message += htmlBody + "\r\n\r\n"

	message += fmt.Sprintf("--%s--", boundary)

	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	return smtp.SendMail(addr, auth, p.from, []string{to}, []byte(message))
}

// NoOpProvider is a no-op email provider for development
type NoOpProvider struct{}

func (p *NoOpProvider) Send(ctx context.Context, to, subject, htmlBody, textBody string) error {
	// In development, just log the email
	fmt.Printf("[EMAIL] To: %s, Subject: %s\n", to, subject)
	return nil
}

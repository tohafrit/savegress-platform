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

// LicensePurchaseInfo contains information about a purchased license
type LicensePurchaseInfo struct {
	UserName       string
	Email          string
	Plan           string    // "Pro" or "Enterprise"
	LicenseKey     string
	Amount         string    // e.g., "$99.00"
	BillingPeriod  string    // "monthly" or "yearly"
	NextBillingDate time.Time
	InvoiceURL     string
}

// SendLicensePurchaseEmail sends a confirmation email after successful license purchase
func (s *EmailService) SendLicensePurchaseEmail(ctx context.Context, info LicensePurchaseInfo) error {
	subject := fmt.Sprintf("Your Savegress %s License is Active", info.Plan)

	// Mask the license key for display (show first 8 and last 4 chars)
	maskedKey := info.LicenseKey
	if len(maskedKey) > 16 {
		maskedKey = maskedKey[:8] + "..." + maskedKey[len(maskedKey)-4:]
	}

	// Features based on plan
	var features string
	if info.Plan == "Enterprise" {
		features = `
            <li>✓ All Pro features included</li>
            <li>✓ Oracle CDC connector</li>
            <li>✓ High Availability cluster mode</li>
            <li>✓ SSO / SAML authentication</li>
            <li>✓ Audit logging & compliance</li>
            <li>✓ Priority support with SLA</li>
            <li>✓ Unlimited sources & tables</li>`
	} else {
		features = `
            <li>✓ All Community features included</li>
            <li>✓ MongoDB CDC connector</li>
            <li>✓ SQL Server CDC connector</li>
            <li>✓ Apache Kafka output</li>
            <li>✓ Advanced compression (Zstd, LZ4)</li>
            <li>✓ Up to 10 sources, 100 tables</li>
            <li>✓ Email support</li>`
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f5f5f5; margin: 0; padding: 0;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #0066cc 0%%, #004499 100%%); padding: 30px; border-radius: 10px 10px 0 0; text-align: center;">
            <h1 style="color: white; margin: 0; font-size: 28px;">Savegress</h1>
            <p style="color: rgba(255,255,255,0.9); margin: 10px 0 0 0;">CDC Platform</p>
        </div>

        <!-- Main Content -->
        <div style="background: white; padding: 30px; border-radius: 0 0 10px 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
            <!-- Success Badge -->
            <div style="text-align: center; margin-bottom: 25px;">
                <div style="display: inline-block; background-color: #e8f5e9; border-radius: 50px; padding: 10px 25px;">
                    <span style="color: #2e7d32; font-weight: 600;">✓ Payment Successful</span>
                </div>
            </div>

            <h2 style="color: #333; margin-top: 0;">Thank you for your purchase, %s!</h2>

            <p>Your <strong>Savegress %s</strong> license is now active. Here are your license details:</p>

            <!-- License Details Box -->
            <div style="background-color: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 25px 0;">
                <table style="width: 100%%; border-collapse: collapse;">
                    <tr>
                        <td style="padding: 8px 0; color: #666;">Plan:</td>
                        <td style="padding: 8px 0; font-weight: 600; text-align: right;">%s</td>
                    </tr>
                    <tr>
                        <td style="padding: 8px 0; color: #666;">Amount:</td>
                        <td style="padding: 8px 0; font-weight: 600; text-align: right;">%s / %s</td>
                    </tr>
                    <tr>
                        <td style="padding: 8px 0; color: #666;">License Key:</td>
                        <td style="padding: 8px 0; font-family: monospace; font-size: 13px; text-align: right;">%s</td>
                    </tr>
                    <tr>
                        <td style="padding: 8px 0; color: #666;">Next Billing:</td>
                        <td style="padding: 8px 0; text-align: right;">%s</td>
                    </tr>
                </table>
            </div>

            <!-- Features -->
            <h3 style="color: #333; margin-bottom: 10px;">What's Included:</h3>
            <ul style="color: #555; padding-left: 20px;">
                %s
            </ul>

            <!-- CTA Buttons -->
            <div style="text-align: center; margin: 30px 0;">
                <a href="%s/downloads" style="background-color: #0066cc; color: white; padding: 14px 35px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: 600; margin: 5px;">Download Software</a>
            </div>

            <div style="text-align: center; margin-bottom: 20px;">
                <a href="%s/docs" style="color: #0066cc; text-decoration: none; margin: 0 15px;">Documentation</a>
                <span style="color: #ccc;">|</span>
                <a href="%s/licenses" style="color: #0066cc; text-decoration: none; margin: 0 15px;">Manage License</a>
                <span style="color: #ccc;">|</span>
                <a href="%s/billing" style="color: #0066cc; text-decoration: none; margin: 0 15px;">Billing</a>
            </div>

            <!-- Quick Start -->
            <div style="background-color: #fff3e0; border-left: 4px solid #ff9800; padding: 15px; margin: 25px 0; border-radius: 0 8px 8px 0;">
                <strong style="color: #e65100;">Quick Start:</strong>
                <p style="margin: 10px 0 0 0; color: #666;">Your license is automatically embedded when you download from the portal. Just download and run!</p>
            </div>

            <!-- Invoice Link -->
            %s
        </div>

        <!-- Footer -->
        <div style="text-align: center; padding: 20px; color: #999; font-size: 12px;">
            <p>© %d Savegress. All rights reserved.</p>
            <p>This is an official receipt for your records.</p>
            <p style="margin-top: 15px;">
                <a href="mailto:support@savegress.io" style="color: #666;">support@savegress.io</a>
            </p>
        </div>
    </div>
</body>
</html>
`,
		template.HTMLEscapeString(info.UserName),
		info.Plan,
		info.Plan,
		info.Amount,
		info.BillingPeriod,
		maskedKey,
		info.NextBillingDate.Format("January 2, 2006"),
		features,
		s.baseURL,
		s.baseURL,
		s.baseURL,
		s.baseURL,
		s.invoiceLinkHTML(info.InvoiceURL),
		time.Now().Year(),
	)

	textBody := fmt.Sprintf(`Thank you for your purchase, %s!

Your Savegress %s license is now active.

LICENSE DETAILS
===============
Plan: %s
Amount: %s / %s
License Key: %s
Next Billing: %s

QUICK START
===========
Your license is automatically embedded when you download from the portal.
Just download and run - no manual configuration needed!

Download: %s/downloads
Documentation: %s/docs
Manage License: %s/licenses

---
© %d Savegress. All rights reserved.
This is an official receipt for your records.

Questions? Contact support@savegress.io
`,
		info.UserName,
		info.Plan,
		info.Plan,
		info.Amount,
		info.BillingPeriod,
		maskedKey,
		info.NextBillingDate.Format("January 2, 2006"),
		s.baseURL,
		s.baseURL,
		s.baseURL,
		time.Now().Year(),
	)

	return s.provider.Send(ctx, info.Email, subject, htmlBody, textBody)
}

// invoiceLinkHTML returns HTML for invoice link if URL is provided
func (s *EmailService) invoiceLinkHTML(invoiceURL string) string {
	if invoiceURL == "" {
		return ""
	}
	return fmt.Sprintf(`
            <p style="text-align: center; margin-top: 20px;">
                <a href="%s" style="color: #666; text-decoration: none; font-size: 14px;">📄 View/Download Invoice</a>
            </p>`, invoiceURL)
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

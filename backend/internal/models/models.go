package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a customer account
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	Name           string     `json:"name" db:"name"`
	Company        string     `json:"company,omitempty" db:"company"`
	Role           string     `json:"role" db:"role"` // user, admin
	EmailVerified  bool       `json:"email_verified" db:"email_verified"`
	StripeCustomerID string   `json:"-" db:"stripe_customer_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// License represents a software license
type License struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	LicenseKey   string     `json:"license_key" db:"license_key"`
	Tier         string     `json:"tier" db:"tier"` // community, pro, enterprise, trial
	Status       string     `json:"status" db:"status"` // active, expired, revoked
	MaxSources   int        `json:"max_sources" db:"max_sources"`
	MaxTables    int        `json:"max_tables" db:"max_tables"`
	MaxThroughput int64     `json:"max_throughput" db:"max_throughput"`
	Features     []string   `json:"features" db:"features"`
	HardwareID   string     `json:"hardware_id,omitempty" db:"hardware_id"`
	IssuedAt     time.Time  `json:"issued_at" db:"issued_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// LicenseActivation records where a license is used
type LicenseActivation struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	LicenseID   uuid.UUID  `json:"license_id" db:"license_id"`
	HardwareID  string     `json:"hardware_id" db:"hardware_id"`
	Hostname    string     `json:"hostname" db:"hostname"`
	Platform    string     `json:"platform" db:"platform"`
	Version     string     `json:"version" db:"version"`
	IPAddress   string     `json:"ip_address" db:"ip_address"`
	ActivatedAt time.Time  `json:"activated_at" db:"activated_at"`
	LastSeenAt  time.Time  `json:"last_seen_at" db:"last_seen_at"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty" db:"deactivated_at"`
}

// Subscription represents a Stripe subscription
type Subscription struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              uuid.UUID  `json:"user_id" db:"user_id"`
	StripeSubscriptionID string    `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	StripePriceID       string     `json:"stripe_price_id" db:"stripe_price_id"`
	Status              string     `json:"status" db:"status"` // active, past_due, canceled, trialing
	Plan                string     `json:"plan" db:"plan"` // pro, enterprise
	CurrentPeriodStart  time.Time  `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd    time.Time  `json:"current_period_end" db:"current_period_end"`
	CancelAtPeriodEnd   bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// TelemetryRecord stores usage data from CDC engines
type TelemetryRecord struct {
	ID              uuid.UUID `json:"id" db:"id"`
	LicenseID       uuid.UUID `json:"license_id" db:"license_id"`
	HardwareID      string    `json:"hardware_id" db:"hardware_id"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
	EventsProcessed int64     `json:"events_processed" db:"events_processed"`
	BytesProcessed  int64     `json:"bytes_processed" db:"bytes_processed"`
	TablesTracked   int       `json:"tables_tracked" db:"tables_tracked"`
	SourcesActive   int       `json:"sources_active" db:"sources_active"`
	AvgLatencyMs    float64   `json:"avg_latency_ms" db:"avg_latency_ms"`
	ErrorCount      int64     `json:"error_count" db:"error_count"`
	UptimeHours     float64   `json:"uptime_hours" db:"uptime_hours"`
	Version         string    `json:"version" db:"version"`
	SourceType      string    `json:"source_type" db:"source_type"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	StripeInvoiceID string     `json:"stripe_invoice_id" db:"stripe_invoice_id"`
	Amount          int64      `json:"amount" db:"amount"` // in cents
	Currency        string     `json:"currency" db:"currency"`
	Status          string     `json:"status" db:"status"` // draft, open, paid, void, uncollectible
	InvoiceURL      string     `json:"invoice_url" db:"invoice_url"`
	InvoicePDF      string     `json:"invoice_pdf" db:"invoice_pdf"`
	PeriodStart     time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd       time.Time  `json:"period_end" db:"period_end"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// PasswordReset stores password reset tokens
type PasswordReset struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RefreshToken stores JWT refresh tokens
type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

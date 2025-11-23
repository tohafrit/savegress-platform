package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

var (
	ErrNoSubscription = errors.New("no active subscription")
	ErrInvalidWebhook = errors.New("invalid webhook signature")
)

// BillingService handles Stripe billing
type BillingService struct {
	db              *repository.PostgresDB
	webhookSecret   string
	proPriceID      string
	enterprisePriceID string
}

// NewBillingService creates a new billing service
func NewBillingService(secretKey, webhookSecret string) *BillingService {
	stripe.Key = secretKey
	return &BillingService{
		webhookSecret: webhookSecret,
	}
}

// SetDB sets the database connection (needed for circular dependency)
func (s *BillingService) SetDB(db *repository.PostgresDB) {
	s.db = db
}

// CreateCustomer creates a Stripe customer for a user
func (s *BillingService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(user.Email),
		Name:  stripe.String(user.Name),
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return c.ID, nil
}

// CreateCheckoutSession creates a Stripe checkout session for subscription
func (s *BillingService) CreateCheckoutSession(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error) {
	priceID := s.getPriceID(plan)
	if priceID == "" {
		return "", errors.New("invalid plan")
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(user.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_id": user.ID.String(),
			"plan":    plan,
		},
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// GetSubscription returns the user's current subscription
func (s *BillingService) GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, stripe_subscription_id, stripe_price_id, status, plan,
			   current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at
		FROM subscriptions WHERE user_id = $1 AND status IN ('active', 'trialing', 'past_due')
	`, userID).Scan(&sub.ID, &sub.UserID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Status, &sub.Plan, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, ErrNoSubscription
	}
	return &sub, nil
}

// CancelSubscription cancels a subscription at period end
func (s *BillingService) CancelSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return err
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update local record
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE subscriptions SET cancel_at_period_end = true, updated_at = NOW()
		WHERE id = $1
	`, sub.ID)

	return err
}

// ReactivateSubscription reactivates a canceled subscription
func (s *BillingService) ReactivateSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return err
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	_, err = s.db.Pool().Exec(ctx, `
		UPDATE subscriptions SET cancel_at_period_end = false, updated_at = NOW()
		WHERE id = $1
	`, sub.ID)

	return err
}

// ListInvoices returns invoices for a user
func (s *BillingService) ListInvoices(ctx context.Context, userID uuid.UUID, limit int) ([]models.Invoice, error) {
	rows, err := s.db.Pool().Query(ctx, `
		SELECT id, user_id, stripe_invoice_id, amount, currency, status, invoice_url, invoice_pdf, period_start, period_end, created_at
		FROM invoices WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]models.Invoice, 0)
	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.UserID, &inv.StripeInvoiceID, &inv.Amount,
			&inv.Currency, &inv.Status, &inv.InvoiceURL, &inv.InvoicePDF,
			&inv.PeriodStart, &inv.PeriodEnd, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}

// ListPaymentMethods returns payment methods for a customer
func (s *BillingService) ListPaymentMethods(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(stripeCustomerID),
		Type:     stripe.String("card"),
	}

	methods := make([]*stripe.PaymentMethod, 0)
	iter := paymentmethod.List(params)
	for iter.Next() {
		methods = append(methods, iter.PaymentMethod())
	}

	return methods, iter.Err()
}

// CreatePortalSession creates a Stripe billing portal session
func (s *BillingService) CreatePortalSession(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create portal session: %w", err)
	}

	return sess.URL, nil
}

// HandleWebhook processes Stripe webhooks
func (s *BillingService) HandleWebhook(payload []byte, signature string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, ErrInvalidWebhook
	}
	return &event, nil
}

func (s *BillingService) getPriceID(plan string) string {
	switch plan {
	case "pro":
		return s.proPriceID
	case "enterprise":
		return s.enterprisePriceID
	default:
		return ""
	}
}

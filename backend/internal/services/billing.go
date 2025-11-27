package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/setupintent"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/repository"
)

var (
	ErrNoSubscription        = errors.New("no active subscription")
	ErrInvalidWebhook        = errors.New("invalid webhook signature")
	ErrInvalidPlan           = errors.New("invalid plan")
	ErrSamePlan              = errors.New("already on this plan")
	ErrPaymentMethodNotFound = errors.New("payment method not found")
)

// BillingService handles Stripe billing
type BillingService struct {
	db                *repository.PostgresDB
	webhookSecret     string
	proPriceID        string
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

// UpdateSubscription upgrades or downgrades a subscription to a new plan
func (s *BillingService) UpdateSubscription(ctx context.Context, userID uuid.UUID, newPlan string) error {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return err
	}

	newPriceID := s.getPriceID(newPlan)
	if newPriceID == "" {
		return errors.New("invalid plan")
	}

	// Get the current subscription from Stripe
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if len(stripeSub.Items.Data) == 0 {
		return errors.New("subscription has no items")
	}

	// Update the subscription item with new price
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(stripeSub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String(string(stripe.SubscriptionSchedulePhaseProrationBehaviorCreateProrations)),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update local record
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE subscriptions SET plan = $1, stripe_price_id = $2, updated_at = NOW()
		WHERE id = $3
	`, newPlan, newPriceID, sub.ID)

	return err
}

// RemovePaymentMethod removes a payment method from a customer
func (s *BillingService) RemovePaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}
	return nil
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

// SetPriceIDs sets the Stripe price IDs for subscription plans
func (s *BillingService) SetPriceIDs(proPriceID, enterprisePriceID string) {
	s.proPriceID = proPriceID
	s.enterprisePriceID = enterprisePriceID
}

// UpdateSubscriptionPlan changes a subscription to a different plan (upgrade/downgrade)
func (s *BillingService) UpdateSubscriptionPlan(ctx context.Context, userID uuid.UUID, newPlan string) (*models.Subscription, error) {
	// Get current subscription
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Validate new plan
	newPriceID := s.getPriceID(newPlan)
	if newPriceID == "" {
		return nil, ErrInvalidPlan
	}

	// Check if already on this plan
	if sub.Plan == newPlan {
		return nil, ErrSamePlan
	}

	// Get subscription items
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe subscription: %w", err)
	}

	if len(stripeSub.Items.Data) == 0 {
		return nil, fmt.Errorf("subscription has no items")
	}

	// Update the subscription item with new price
	itemID := stripeSub.Items.Data[0].ID
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String(string(stripe.SubscriptionSchedulePhaseProrationBehaviorCreateProrations)),
	}

	updatedSub, err := subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update local record
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET plan = $1, stripe_price_id = $2, updated_at = NOW()
		WHERE id = $3
	`, newPlan, newPriceID, sub.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update local subscription: %w", err)
	}

	sub.Plan = newPlan
	sub.StripePriceID = newPriceID
	sub.CurrentPeriodEnd = time.Unix(updatedSub.CurrentPeriodEnd, 0)

	return sub, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (s *BillingService) AttachPaymentMethod(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(stripeCustomerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return fmt.Errorf("failed to attach payment method: %w", err)
	}

	return nil
}

// DetachPaymentMethod removes a payment method from a customer
func (s *BillingService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}
	return nil
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (s *BillingService) SetDefaultPaymentMethod(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	_, err := customer.Update(stripeCustomerID, params)
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}
	return nil
}

// CreateSetupIntent creates a SetupIntent for adding a new payment method via Stripe.js
func (s *BillingService) CreateSetupIntent(ctx context.Context, stripeCustomerID string) (string, error) {
	params := &stripe.SetupIntentParams{
		Customer:           stripe.String(stripeCustomerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	}

	intent, err := setupintent.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create setup intent: %w", err)
	}

	return intent.ClientSecret, nil
}

// CreateOrUpdateSubscription creates a subscription from a checkout session or updates existing
func (s *BillingService) CreateOrUpdateSubscription(ctx context.Context, userID uuid.UUID, stripeSubID, stripePriceID, plan, status string, periodStart, periodEnd int64) error {
	// Check if subscription exists
	var existingID uuid.UUID
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id FROM subscriptions WHERE user_id = $1
	`, userID).Scan(&existingID)

	if err == nil {
		// Update existing
		_, err = s.db.Pool().Exec(ctx, `
			UPDATE subscriptions
			SET stripe_subscription_id = $1, stripe_price_id = $2, plan = $3, status = $4,
				current_period_start = to_timestamp($5), current_period_end = to_timestamp($6), updated_at = NOW()
			WHERE id = $7
		`, stripeSubID, stripePriceID, plan, status,
			periodStart, periodEnd, existingID)
		return err
	}

	// Create new
	subID := uuid.New()
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO subscriptions (id, user_id, stripe_subscription_id, stripe_price_id, plan, status,
			current_period_start, current_period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, to_timestamp($7), to_timestamp($8), NOW(), NOW())
	`, subID, userID, stripeSubID, stripePriceID, plan, status, periodStart, periodEnd)
	return err
}

// UpdateSubscriptionStatus updates subscription status in the database
func (s *BillingService) UpdateSubscriptionStatus(ctx context.Context, stripeSubID, status string, cancelAtPeriodEnd bool) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = $1, cancel_at_period_end = $2, updated_at = NOW()
		WHERE stripe_subscription_id = $3
	`, status, cancelAtPeriodEnd, stripeSubID)
	return err
}

// DeleteSubscription marks a subscription as canceled
func (s *BillingService) DeleteSubscription(ctx context.Context, stripeSubID string) error {
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = 'canceled', updated_at = NOW()
		WHERE stripe_subscription_id = $1
	`, stripeSubID)
	return err
}

// RecordInvoice stores invoice information from Stripe webhook
func (s *BillingService) RecordInvoice(ctx context.Context, userID uuid.UUID, stripeInvoiceID string, amount int64, currency, status, invoiceURL, invoicePDF string, periodStart, periodEnd int64) error {
	invoiceID := uuid.New()
	_, err := s.db.Pool().Exec(ctx, `
		INSERT INTO invoices (id, user_id, stripe_invoice_id, amount, currency, status, invoice_url, invoice_pdf, period_start, period_end, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, to_timestamp($9), to_timestamp($10), NOW())
		ON CONFLICT (stripe_invoice_id) DO UPDATE SET status = $6, invoice_url = $7, invoice_pdf = $8
	`, invoiceID, userID, stripeInvoiceID, amount, currency, status, invoiceURL, invoicePDF, periodStart, periodEnd)
	return err
}

// GetUserByStripeCustomerID finds a user by their Stripe customer ID
func (s *BillingService) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error) {
	var user models.User
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, email, name, COALESCE(company, '') FROM users WHERE stripe_customer_id = $1
	`, stripeCustomerID).Scan(&user.ID, &user.Email, &user.Name, &user.Company)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetPlanFromPriceID returns the plan name for a Stripe price ID
func (s *BillingService) GetPlanFromPriceID(priceID string) string {
	switch priceID {
	case s.proPriceID:
		return "pro"
	case s.enterprisePriceID:
		return "enterprise"
	default:
		return "unknown"
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// BillingHandler handles billing endpoints
type BillingHandler struct {
	billingService *services.BillingService
	licenseService *services.LicenseService
	userService    *services.UserService
	emailService   *services.EmailService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *services.BillingService, licenseService *services.LicenseService, userService *services.UserService, emailService *services.EmailService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		licenseService: licenseService,
		userService:    userService,
		emailService:   emailService,
	}
}

// GetSubscription returns user's subscription
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	sub, err := h.billingService.GetSubscription(r.Context(), userID)
	if err == services.ErrNoSubscription {
		respondSuccess(w, map[string]interface{}{"subscription": nil})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	respondSuccess(w, map[string]interface{}{"subscription": sub})
}

// CreateSubscription creates a checkout session
func (h *BillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req struct {
		Plan       string `json:"plan"` // pro, enterprise
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "user not found")
		return
	}

	// Create Stripe customer if needed
	if user.StripeCustomerID == "" {
		customerID, err := h.billingService.CreateCustomer(r.Context(), user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create customer")
			return
		}
		user.StripeCustomerID = customerID
		h.userService.SetStripeCustomerID(r.Context(), user.ID, customerID)
	}

	checkoutURL, err := h.billingService.CreateCheckoutSession(r.Context(), user, req.Plan, req.SuccessURL, req.CancelURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	respondSuccess(w, map[string]string{"checkout_url": checkoutURL})
}

// UpdateSubscription updates subscription plan (upgrade/downgrade)
func (h *BillingHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req struct {
		Plan string `json:"plan"` // pro, enterprise
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Plan != "pro" && req.Plan != "enterprise" {
		respondError(w, http.StatusBadRequest, "invalid plan: must be 'pro' or 'enterprise'")
		return
	}

	// Update subscription
	sub, err := h.billingService.UpdateSubscriptionPlan(r.Context(), userID, req.Plan)
	if err != nil {
		switch err {
		case services.ErrNoSubscription:
			respondError(w, http.StatusNotFound, "no active subscription found")
		case services.ErrInvalidPlan:
			respondError(w, http.StatusBadRequest, "invalid plan")
		case services.ErrSamePlan:
			respondError(w, http.StatusBadRequest, "already on this plan")
		default:
			respondError(w, http.StatusInternalServerError, "failed to update subscription")
		}
		return
	}

	respondSuccess(w, map[string]interface{}{
		"message":      "subscription updated successfully",
		"subscription": sub,
	})
}

// CancelSubscription cancels subscription
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	if err := h.billingService.CancelSubscription(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	respondSuccess(w, map[string]string{"message": "subscription will be canceled at period end"})
}

// ListInvoices returns user's invoices
func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	invoices, err := h.billingService.ListInvoices(r.Context(), userID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get invoices")
		return
	}

	respondSuccess(w, map[string]interface{}{"invoices": invoices})
}

// ListPaymentMethods returns user's payment methods
func (h *BillingHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondSuccess(w, map[string]interface{}{"payment_methods": []interface{}{}})
		return
	}

	methods, err := h.billingService.ListPaymentMethods(r.Context(), user.StripeCustomerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get payment methods")
		return
	}

	respondSuccess(w, map[string]interface{}{"payment_methods": methods})
}

// AddPaymentMethod creates a SetupIntent for adding a payment method via Stripe.js
func (h *BillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found - please create a subscription first")
		return
	}

	// Create a SetupIntent for the frontend to use with Stripe.js
	clientSecret, err := h.billingService.CreateSetupIntent(r.Context(), user.StripeCustomerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create setup intent")
		return
	}

	respondSuccess(w, map[string]string{
		"client_secret": clientSecret,
		"message":       "use this client_secret with Stripe.js confirmCardSetup()",
	})
}

// AttachPaymentMethod attaches a payment method created via Stripe.js
func (h *BillingHandler) AttachPaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id"`
		SetAsDefault    bool   `json:"set_as_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PaymentMethodID == "" {
		respondError(w, http.StatusBadRequest, "payment_method_id is required")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	// Attach payment method to customer
	if err := h.billingService.AttachPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to attach payment method")
		return
	}

	// Set as default if requested
	if req.SetAsDefault {
		if err := h.billingService.SetDefaultPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
			log.Printf("Failed to set default payment method: %v", err)
		}
	}

	respondSuccess(w, map[string]string{"message": "payment method attached successfully"})
}

// RemovePaymentMethod removes a payment method
func (h *BillingHandler) RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	paymentMethodID := chi.URLParam(r, "id")
	if paymentMethodID == "" {
		respondError(w, http.StatusBadRequest, "payment method ID is required")
		return
	}

	// Detach the payment method
	if err := h.billingService.DetachPaymentMethod(r.Context(), paymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to remove payment method")
		return
	}

	respondSuccess(w, map[string]string{"message": "payment method removed successfully"})
}

// SetDefaultPaymentMethod sets a payment method as the default
func (h *BillingHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	if err := h.billingService.SetDefaultPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to set default payment method")
		return
	}

	respondSuccess(w, map[string]string{"message": "default payment method updated"})
}

// CreatePortalSession creates Stripe billing portal session
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := claims.GetUserUUID()
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req struct {
		ReturnURL string `json:"return_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	portalURL, err := h.billingService.CreatePortalSession(r.Context(), user.StripeCustomerID, req.ReturnURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create portal session")
		return
	}

	respondSuccess(w, map[string]string{"portal_url": portalURL})
}

// HandleWebhook processes Stripe webhooks
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := h.billingService.HandleWebhook(payload, signature)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid webhook signature")
		return
	}

	ctx := r.Context()

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(ctx, event)
	case "invoice.paid":
		h.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		h.handlePaymentFailed(ctx, event)
	}

	respondSuccess(w, map[string]string{"received": "true"})
}

func (h *BillingHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) {
	// Parse the checkout session from the event
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

	// Extract user ID from metadata
	userIDStr, ok := session.Metadata["user_id"]
	if !ok {
		log.Printf("No user_id in checkout session metadata")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user_id in metadata: %v", err)
		return
	}

	plan, _ := session.Metadata["plan"]
	if plan == "" {
		plan = "pro" // Default
	}

	// Get subscription details from Stripe
	if session.Subscription == nil {
		log.Printf("No subscription in checkout session")
		return
	}

	subID := session.Subscription.ID

	// Create or update subscription in database
	var priceID string
	if len(session.LineItems.Data) > 0 && session.LineItems.Data[0].Price != nil {
		priceID = session.LineItems.Data[0].Price.ID
	}

	err = h.billingService.CreateOrUpdateSubscription(
		ctx,
		userID,
		subID,
		priceID,
		plan,
		"active",
		session.Subscription.CurrentPeriodStart,
		session.Subscription.CurrentPeriodEnd,
	)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		return
	}

	// Upgrade existing license or create new one
	// First, try to revoke any existing community license
	existingLicenses, err := h.licenseService.GetUserLicenses(ctx, userID)
	if err == nil {
		for _, lic := range existingLicenses {
			if lic.Status == "active" && lic.Tier == "community" {
				// Revoke community license when upgrading
				_ = h.licenseService.RevokeLicense(ctx, lic.ID)
				log.Printf("Revoked community license %s for user %s (upgrading to %s)", lic.ID, userID, plan)
			}
		}
	}

	// Create new license for this subscription
	newLicense, err := h.licenseService.CreateLicense(ctx, userID, plan, 365, "")
	if err != nil {
		log.Printf("Error creating license: %v", err)
	}

	log.Printf("Checkout completed for user %s, plan: %s", userID, plan)

	// Send purchase confirmation email
	if h.emailService != nil && newLicense != nil {
		user, err := h.userService.GetByID(ctx, userID)
		if err == nil {
			// Determine plan display name and amount
			planName := "Pro"
			amount := "$99.00"
			if plan == "enterprise" {
				planName = "Enterprise"
				amount = "$499.00"
			}

			// Get invoice URL if available
			var invoiceURL string
			if session.Invoice != nil {
				invoiceURL = session.Invoice.HostedInvoiceURL
			}

			purchaseInfo := services.LicensePurchaseInfo{
				UserName:        user.Name,
				Email:           user.Email,
				Plan:            planName,
				LicenseKey:      newLicense.LicenseKey,
				Amount:          amount,
				BillingPeriod:   "month",
				NextBillingDate: time.Unix(session.Subscription.CurrentPeriodEnd, 0),
				InvoiceURL:      invoiceURL,
			}

			if err := h.emailService.SendLicensePurchaseEmail(ctx, purchaseInfo); err != nil {
				log.Printf("Error sending license purchase email: %v", err)
			} else {
				log.Printf("Sent license purchase confirmation email to %s", user.Email)
			}
		}
	}
}

func (h *BillingHandler) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	// Update subscription status in database
	err := h.billingService.UpdateSubscriptionStatus(
		ctx,
		sub.ID,
		string(sub.Status),
		sub.CancelAtPeriodEnd,
	)
	if err != nil {
		log.Printf("Error updating subscription status: %v", err)
	}

	log.Printf("Subscription %s updated: status=%s, cancel_at_period_end=%v", sub.ID, sub.Status, sub.CancelAtPeriodEnd)
}

func (h *BillingHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	// Mark subscription as canceled
	err := h.billingService.DeleteSubscription(ctx, sub.ID)
	if err != nil {
		log.Printf("Error deleting subscription: %v", err)
	}

	// Find user and downgrade to community
	if sub.Customer != nil {
		user, err := h.billingService.GetUserByStripeCustomerID(ctx, sub.Customer.ID)
		if err == nil {
			// Revoke paid license and create community license
			existingLicenses, err := h.licenseService.GetUserLicenses(ctx, user.ID)
			if err == nil {
				for _, lic := range existingLicenses {
					if lic.Status == "active" && (lic.Tier == "pro" || lic.Tier == "enterprise") {
						_ = h.licenseService.RevokeLicense(ctx, lic.ID)
						log.Printf("Revoked %s license %s for user %s (subscription canceled)", lic.Tier, lic.ID, user.ID)
					}
				}
			}

			// Create new community license
			_, err = h.licenseService.CreateLicense(ctx, user.ID, "community", 365, "")
			if err != nil {
				log.Printf("Error creating community license after cancellation: %v", err)
			} else {
				log.Printf("Created community license for user %s after subscription cancellation", user.ID)
			}

			// Send notification
			if h.emailService != nil {
				periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)
				_ = h.emailService.SendSubscriptionCanceledEmail(ctx, user.Email, user.Name, periodEnd)
			}
		}
	}

	log.Printf("Subscription %s deleted", sub.ID)
}

func (h *BillingHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice: %v", err)
		return
	}

	// Find user by customer ID
	if invoice.Customer == nil {
		log.Printf("No customer in invoice")
		return
	}

	user, err := h.billingService.GetUserByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		log.Printf("User not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Record invoice
	var invoiceURL, invoicePDF string
	if invoice.HostedInvoiceURL != "" {
		invoiceURL = invoice.HostedInvoiceURL
	}
	if invoice.InvoicePDF != "" {
		invoicePDF = invoice.InvoicePDF
	}

	err = h.billingService.RecordInvoice(
		ctx,
		user.ID,
		invoice.ID,
		invoice.AmountPaid,
		string(invoice.Currency),
		string(invoice.Status),
		invoiceURL,
		invoicePDF,
		invoice.PeriodStart,
		invoice.PeriodEnd,
	)
	if err != nil {
		log.Printf("Error recording invoice: %v", err)
	}

	log.Printf("Invoice %s paid for user %s", invoice.ID, user.ID)
}

func (h *BillingHandler) handlePaymentFailed(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice: %v", err)
		return
	}

	// Find user and send notification
	if invoice.Customer == nil {
		return
	}

	user, err := h.billingService.GetUserByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		log.Printf("User not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Send payment failed email
	if h.emailService != nil {
		if err := h.emailService.SendPaymentFailedEmail(ctx, user.Email, user.Name); err != nil {
			log.Printf("Error sending payment failed email: %v", err)
		}
	}

	log.Printf("Payment failed for user %s", user.ID)
}

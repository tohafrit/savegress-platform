package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/services"
)

// BillingHandler handles billing endpoints
type BillingHandler struct {
	billingService *services.BillingService
	licenseService *services.LicenseService
	userService    *services.UserService
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *services.BillingService, licenseService *services.LicenseService, userService *services.UserService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		licenseService: licenseService,
		userService:    userService,
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

// UpdateSubscription updates subscription plan
func (h *BillingHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement plan upgrade/downgrade
	respondError(w, http.StatusNotImplemented, "not implemented")
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

// AddPaymentMethod adds a payment method
func (h *BillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement - typically handled by Stripe.js
	respondError(w, http.StatusNotImplemented, "use Stripe.js to add payment methods")
}

// RemovePaymentMethod removes a payment method
func (h *BillingHandler) RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	// TODO: Implement
	respondError(w, http.StatusNotImplemented, "not implemented")
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

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.paid":
		h.handleInvoicePaid(event)
	case "invoice.payment_failed":
		h.handlePaymentFailed(event)
	}

	respondSuccess(w, map[string]string{"received": "true"})
}

func (h *BillingHandler) handleCheckoutCompleted(event *stripe.Event) {
	// Create license for new subscription
	// TODO: Extract user ID and plan from session metadata
}

func (h *BillingHandler) handleSubscriptionUpdated(event *stripe.Event) {
	// Update subscription status in database
}

func (h *BillingHandler) handleSubscriptionDeleted(event *stripe.Event) {
	// Revoke license when subscription ends
}

func (h *BillingHandler) handleInvoicePaid(event *stripe.Event) {
	// Record invoice in database
}

func (h *BillingHandler) handlePaymentFailed(event *stripe.Event) {
	// Send notification to user
}

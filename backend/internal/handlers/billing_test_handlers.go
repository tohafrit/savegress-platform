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

// testBillingHandler wraps mock services for billing handler testing
type testBillingHandler struct {
	mockBilling *MockBillingService
	mockUser    *MockUserService
	mockLicense *MockLicenseServiceForHandler
	mockEmail   *MockEmailService
}

func newTestBillingHandler(mockBilling *MockBillingService, mockUser *MockUserService, mockLicense *MockLicenseServiceForHandler, mockEmail *MockEmailService) *testBillingHandler {
	return &testBillingHandler{
		mockBilling: mockBilling,
		mockUser:    mockUser,
		mockLicense: mockLicense,
		mockEmail:   mockEmail,
	}
}

func (h *testBillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
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

	sub, err := h.mockBilling.GetSubscription(r.Context(), userID)
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

func (h *testBillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
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
		Plan       string `json:"plan"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "user not found")
		return
	}

	if user.StripeCustomerID == "" {
		customerID, err := h.mockBilling.CreateCustomer(r.Context(), user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create customer")
			return
		}
		user.StripeCustomerID = customerID
		h.mockUser.SetStripeCustomerID(r.Context(), user.ID, customerID)
	}

	checkoutURL, err := h.mockBilling.CreateCheckoutSession(r.Context(), user, req.Plan, req.SuccessURL, req.CancelURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	respondSuccess(w, map[string]string{"checkout_url": checkoutURL})
}

func (h *testBillingHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
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
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Plan != "pro" && req.Plan != "enterprise" {
		respondError(w, http.StatusBadRequest, "invalid plan: must be 'pro' or 'enterprise'")
		return
	}

	sub, err := h.mockBilling.UpdateSubscriptionPlan(r.Context(), userID, req.Plan)
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

func (h *testBillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
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

	if err := h.mockBilling.CancelSubscription(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	respondSuccess(w, map[string]string{"message": "subscription will be canceled at period end"})
}

func (h *testBillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
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

	invoices, err := h.mockBilling.ListInvoices(r.Context(), userID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get invoices")
		return
	}

	respondSuccess(w, map[string]interface{}{"invoices": invoices})
}

func (h *testBillingHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondSuccess(w, map[string]interface{}{"payment_methods": []interface{}{}})
		return
	}

	methods, err := h.mockBilling.ListPaymentMethods(r.Context(), user.StripeCustomerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get payment methods")
		return
	}

	respondSuccess(w, map[string]interface{}{"payment_methods": methods})
}

func (h *testBillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found - please create a subscription first")
		return
	}

	clientSecret, err := h.mockBilling.CreateSetupIntent(r.Context(), user.StripeCustomerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create setup intent")
		return
	}

	respondSuccess(w, map[string]string{
		"client_secret": clientSecret,
		"message":       "use this client_secret with Stripe.js confirmCardSetup()",
	})
}

func (h *testBillingHandler) AttachPaymentMethod(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	if err := h.mockBilling.AttachPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to attach payment method")
		return
	}

	if req.SetAsDefault {
		if err := h.mockBilling.SetDefaultPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
			log.Printf("Failed to set default payment method: %v", err)
		}
	}

	respondSuccess(w, map[string]string{"message": "payment method attached successfully"})
}

func (h *testBillingHandler) RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
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

	if err := h.mockBilling.DetachPaymentMethod(r.Context(), paymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to remove payment method")
		return
	}

	respondSuccess(w, map[string]string{"message": "payment method removed successfully"})
}

func (h *testBillingHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	if err := h.mockBilling.SetDefaultPaymentMethod(r.Context(), user.StripeCustomerID, req.PaymentMethodID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to set default payment method")
		return
	}

	respondSuccess(w, map[string]string{"message": "default payment method updated"})
}

func (h *testBillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.mockUser.GetByID(r.Context(), userID)
	if err != nil || user.StripeCustomerID == "" {
		respondError(w, http.StatusBadRequest, "no billing account found")
		return
	}

	portalURL, err := h.mockBilling.CreatePortalSession(r.Context(), user.StripeCustomerID, req.ReturnURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create portal session")
		return
	}

	respondSuccess(w, map[string]string{"portal_url": portalURL})
}

func (h *testBillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := h.mockBilling.HandleWebhook(payload, signature)
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

func (h *testBillingHandler) handleCheckoutCompleted(ctx context.Context, event *stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session: %v", err)
		return
	}

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
		plan = "pro"
	}

	if session.Subscription == nil {
		log.Printf("No subscription in checkout session")
		return
	}

	subID := session.Subscription.ID

	var priceID string
	if len(session.LineItems.Data) > 0 && session.LineItems.Data[0].Price != nil {
		priceID = session.LineItems.Data[0].Price.ID
	}

	err = h.mockBilling.CreateOrUpdateSubscription(
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

	_, err = h.mockLicense.CreateLicense(ctx, userID, plan, 365, "")
	if err != nil {
		log.Printf("Error creating license: %v", err)
	}

	log.Printf("Checkout completed for user %s, plan: %s", userID, plan)
}

func (h *testBillingHandler) handleSubscriptionUpdated(ctx context.Context, event *stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	err := h.mockBilling.UpdateSubscriptionStatus(
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

func (h *testBillingHandler) handleSubscriptionDeleted(ctx context.Context, event *stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error parsing subscription: %v", err)
		return
	}

	err := h.mockBilling.DeleteSubscription(ctx, sub.ID)
	if err != nil {
		log.Printf("Error deleting subscription: %v", err)
	}

	if sub.Customer != nil && h.mockEmail != nil {
		user, err := h.mockBilling.GetUserByStripeCustomerID(ctx, sub.Customer.ID)
		if err == nil {
			periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)
			_ = h.mockEmail.SendSubscriptionCanceledEmail(ctx, user.Email, user.Name, periodEnd)
		}
	}

	log.Printf("Subscription %s deleted", sub.ID)
}

func (h *testBillingHandler) handleInvoicePaid(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice: %v", err)
		return
	}

	if invoice.Customer == nil {
		log.Printf("No customer in invoice")
		return
	}

	user, err := h.mockBilling.GetUserByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		log.Printf("User not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	var invoiceURL, invoicePDF string
	if invoice.HostedInvoiceURL != "" {
		invoiceURL = invoice.HostedInvoiceURL
	}
	if invoice.InvoicePDF != "" {
		invoicePDF = invoice.InvoicePDF
	}

	err = h.mockBilling.RecordInvoice(
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

func (h *testBillingHandler) handlePaymentFailed(ctx context.Context, event *stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice: %v", err)
		return
	}

	if invoice.Customer == nil {
		return
	}

	user, err := h.mockBilling.GetUserByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		log.Printf("User not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	if h.mockEmail != nil {
		if err := h.mockEmail.SendPaymentFailedEmail(ctx, user.Email, user.Name); err != nil {
			log.Printf("Error sending payment failed email: %v", err)
		}
	}

	log.Printf("Payment failed for user %s", user.ID)
}

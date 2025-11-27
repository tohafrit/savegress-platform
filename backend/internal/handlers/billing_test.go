package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"

	"github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/models"
	"github.com/savegress/platform/backend/internal/services"
)

// MockBillingService implements a mock for testing
type MockBillingService struct {
	GetSubscriptionFunc           func(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	CreateCustomerFunc            func(ctx context.Context, user *models.User) (string, error)
	CreateCheckoutSessionFunc     func(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error)
	UpdateSubscriptionPlanFunc    func(ctx context.Context, userID uuid.UUID, plan string) (*models.Subscription, error)
	CancelSubscriptionFunc        func(ctx context.Context, userID uuid.UUID) error
	ListInvoicesFunc              func(ctx context.Context, userID uuid.UUID, limit int) ([]models.Invoice, error)
	ListPaymentMethodsFunc        func(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error)
	CreateSetupIntentFunc         func(ctx context.Context, stripeCustomerID string) (string, error)
	AttachPaymentMethodFunc       func(ctx context.Context, stripeCustomerID, paymentMethodID string) error
	SetDefaultPaymentMethodFunc   func(ctx context.Context, stripeCustomerID, paymentMethodID string) error
	DetachPaymentMethodFunc       func(ctx context.Context, paymentMethodID string) error
	CreatePortalSessionFunc       func(ctx context.Context, stripeCustomerID, returnURL string) (string, error)
	HandleWebhookFunc             func(payload []byte, signature string) (*stripe.Event, error)
	CreateOrUpdateSubscriptionFunc func(ctx context.Context, userID uuid.UUID, stripeSubID, stripePriceID, plan, status string, periodStart, periodEnd int64) error
	UpdateSubscriptionStatusFunc  func(ctx context.Context, stripeSubID, status string, cancelAtPeriodEnd bool) error
	DeleteSubscriptionFunc        func(ctx context.Context, stripeSubID string) error
	RecordInvoiceFunc             func(ctx context.Context, userID uuid.UUID, stripeInvoiceID string, amount int64, currency, status, invoiceURL, invoicePDF string, periodStart, periodEnd int64) error
	GetUserByStripeCustomerIDFunc func(ctx context.Context, stripeCustomerID string) (*models.User, error)
}

func (m *MockBillingService) GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	if m.GetSubscriptionFunc != nil {
		return m.GetSubscriptionFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockBillingService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	if m.CreateCustomerFunc != nil {
		return m.CreateCustomerFunc(ctx, user)
	}
	return "", nil
}

func (m *MockBillingService) CreateCheckoutSession(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error) {
	if m.CreateCheckoutSessionFunc != nil {
		return m.CreateCheckoutSessionFunc(ctx, user, plan, successURL, cancelURL)
	}
	return "", nil
}

func (m *MockBillingService) UpdateSubscriptionPlan(ctx context.Context, userID uuid.UUID, plan string) (*models.Subscription, error) {
	if m.UpdateSubscriptionPlanFunc != nil {
		return m.UpdateSubscriptionPlanFunc(ctx, userID, plan)
	}
	return nil, nil
}

func (m *MockBillingService) CancelSubscription(ctx context.Context, userID uuid.UUID) error {
	if m.CancelSubscriptionFunc != nil {
		return m.CancelSubscriptionFunc(ctx, userID)
	}
	return nil
}

func (m *MockBillingService) ListInvoices(ctx context.Context, userID uuid.UUID, limit int) ([]models.Invoice, error) {
	if m.ListInvoicesFunc != nil {
		return m.ListInvoicesFunc(ctx, userID, limit)
	}
	return nil, nil
}

func (m *MockBillingService) ListPaymentMethods(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error) {
	if m.ListPaymentMethodsFunc != nil {
		return m.ListPaymentMethodsFunc(ctx, stripeCustomerID)
	}
	return nil, nil
}

func (m *MockBillingService) CreateSetupIntent(ctx context.Context, stripeCustomerID string) (string, error) {
	if m.CreateSetupIntentFunc != nil {
		return m.CreateSetupIntentFunc(ctx, stripeCustomerID)
	}
	return "", nil
}

func (m *MockBillingService) AttachPaymentMethod(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
	if m.AttachPaymentMethodFunc != nil {
		return m.AttachPaymentMethodFunc(ctx, stripeCustomerID, paymentMethodID)
	}
	return nil
}

func (m *MockBillingService) SetDefaultPaymentMethod(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
	if m.SetDefaultPaymentMethodFunc != nil {
		return m.SetDefaultPaymentMethodFunc(ctx, stripeCustomerID, paymentMethodID)
	}
	return nil
}

func (m *MockBillingService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	if m.DetachPaymentMethodFunc != nil {
		return m.DetachPaymentMethodFunc(ctx, paymentMethodID)
	}
	return nil
}

func (m *MockBillingService) CreatePortalSession(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
	if m.CreatePortalSessionFunc != nil {
		return m.CreatePortalSessionFunc(ctx, stripeCustomerID, returnURL)
	}
	return "", nil
}

func (m *MockBillingService) HandleWebhook(payload []byte, signature string) (*stripe.Event, error) {
	if m.HandleWebhookFunc != nil {
		return m.HandleWebhookFunc(payload, signature)
	}
	return nil, nil
}

func (m *MockBillingService) CreateOrUpdateSubscription(ctx context.Context, userID uuid.UUID, stripeSubID, stripePriceID, plan, status string, periodStart, periodEnd int64) error {
	if m.CreateOrUpdateSubscriptionFunc != nil {
		return m.CreateOrUpdateSubscriptionFunc(ctx, userID, stripeSubID, stripePriceID, plan, status, periodStart, periodEnd)
	}
	return nil
}

func (m *MockBillingService) UpdateSubscriptionStatus(ctx context.Context, stripeSubID, status string, cancelAtPeriodEnd bool) error {
	if m.UpdateSubscriptionStatusFunc != nil {
		return m.UpdateSubscriptionStatusFunc(ctx, stripeSubID, status, cancelAtPeriodEnd)
	}
	return nil
}

func (m *MockBillingService) DeleteSubscription(ctx context.Context, stripeSubID string) error {
	if m.DeleteSubscriptionFunc != nil {
		return m.DeleteSubscriptionFunc(ctx, stripeSubID)
	}
	return nil
}

func (m *MockBillingService) RecordInvoice(ctx context.Context, userID uuid.UUID, stripeInvoiceID string, amount int64, currency, status, invoiceURL, invoicePDF string, periodStart, periodEnd int64) error {
	if m.RecordInvoiceFunc != nil {
		return m.RecordInvoiceFunc(ctx, userID, stripeInvoiceID, amount, currency, status, invoiceURL, invoicePDF, periodStart, periodEnd)
	}
	return nil
}

func (m *MockBillingService) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error) {
	if m.GetUserByStripeCustomerIDFunc != nil {
		return m.GetUserByStripeCustomerIDFunc(ctx, stripeCustomerID)
	}
	return nil, nil
}

// MockEmailService implements a mock for testing
type MockEmailService struct {
	SendSubscriptionCanceledEmailFunc func(ctx context.Context, email, name string, periodEnd time.Time) error
	SendPaymentFailedEmailFunc        func(ctx context.Context, email, name string) error
}

func (m *MockEmailService) SendSubscriptionCanceledEmail(ctx context.Context, email, name string, periodEnd time.Time) error {
	if m.SendSubscriptionCanceledEmailFunc != nil {
		return m.SendSubscriptionCanceledEmailFunc(ctx, email, name, periodEnd)
	}
	return nil
}

func (m *MockEmailService) SendPaymentFailedEmail(ctx context.Context, email, name string) error {
	if m.SendPaymentFailedEmailFunc != nil {
		return m.SendPaymentFailedEmailFunc(ctx, email, name)
	}
	return nil
}

// Helper to create authenticated context
func createAuthContext(userID uuid.UUID) context.Context {
	claims := &services.Claims{
		UserID: userID.String(),
		Email:  "test@example.com",
		Role:   "user",
	}
	ctx := context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
	return ctx
}

func TestBillingHandler_GetSubscription(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name               string
		withAuth           bool
		mockGetSubscription func(ctx context.Context, uid uuid.UUID) (*models.Subscription, error)
		expectedStatus     int
		expectedError      string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "no subscription",
			withAuth: true,
			mockGetSubscription: func(ctx context.Context, uid uuid.UUID) (*models.Subscription, error) {
				return nil, services.ErrNoSubscription
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "successful retrieval",
			withAuth: true,
			mockGetSubscription: func(ctx context.Context, uid uuid.UUID) (*models.Subscription, error) {
				return &models.Subscription{
					ID:                   uuid.New(),
					UserID:               uid,
					StripeSubscriptionID: "sub_123",
					Plan:                 "pro",
					Status:               "active",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "internal error",
			withAuth: true,
			mockGetSubscription: func(ctx context.Context, uid uuid.UUID) (*models.Subscription, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get subscription",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				GetSubscriptionFunc: tt.mockGetSubscription,
			}
			handler := newTestBillingHandler(mockBilling, nil, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/subscription", nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.GetSubscription(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_CreateSubscription(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                   string
		withAuth               bool
		requestBody            map[string]string
		mockGetByID            func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockCreateCustomer     func(ctx context.Context, user *models.User) (string, error)
		mockCreateCheckout     func(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error)
		mockSetStripeCustomerID func(ctx context.Context, userID uuid.UUID, stripeCustomerID string) error
		expectedStatus         int
		expectedError          string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid request body",
			withAuth:       true,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:     "successful checkout session creation - existing customer",
			withAuth: true,
			requestBody: map[string]string{
				"plan":        "pro",
				"success_url": "https://example.com/success",
				"cancel_url":  "https://example.com/cancel",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					Email:            "test@example.com",
					Name:             "Test User",
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreateCheckout: func(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error) {
				return "https://checkout.stripe.com/session_123", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "successful checkout session creation - new customer",
			withAuth: true,
			requestBody: map[string]string{
				"plan":        "pro",
				"success_url": "https://example.com/success",
				"cancel_url":  "https://example.com/cancel",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					Email:            "test@example.com",
					Name:             "Test User",
					StripeCustomerID: "",
				}, nil
			},
			mockCreateCustomer: func(ctx context.Context, user *models.User) (string, error) {
				return "cus_new123", nil
			},
			mockSetStripeCustomerID: func(ctx context.Context, userID uuid.UUID, stripeCustomerID string) error {
				return nil
			},
			mockCreateCheckout: func(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error) {
				return "https://checkout.stripe.com/session_123", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "user not found",
			withAuth: true,
			requestBody: map[string]string{
				"plan":        "pro",
				"success_url": "https://example.com/success",
				"cancel_url":  "https://example.com/cancel",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "user not found",
		},
		{
			name:     "failed to create customer",
			withAuth: true,
			requestBody: map[string]string{
				"plan":        "pro",
				"success_url": "https://example.com/success",
				"cancel_url":  "https://example.com/cancel",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					Email:            "test@example.com",
					Name:             "Test User",
					StripeCustomerID: "",
				}, nil
			},
			mockCreateCustomer: func(ctx context.Context, user *models.User) (string, error) {
				return "", errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create customer",
		},
		{
			name:     "failed to create checkout session",
			withAuth: true,
			requestBody: map[string]string{
				"plan":        "pro",
				"success_url": "https://example.com/success",
				"cancel_url":  "https://example.com/cancel",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					Email:            "test@example.com",
					Name:             "Test User",
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreateCheckout: func(ctx context.Context, user *models.User, plan, successURL, cancelURL string) (string, error) {
				return "", errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create checkout session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				CreateCustomerFunc:        tt.mockCreateCustomer,
				CreateCheckoutSessionFunc: tt.mockCreateCheckout,
			}
			mockUser := &MockUserService{
				GetByIDFunc:             tt.mockGetByID,
				SetStripeCustomerIDFunc: tt.mockSetStripeCustomerID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/subscription", bytes.NewReader(body))
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.CreateSubscription(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_UpdateSubscription(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                    string
		withAuth                bool
		requestBody             map[string]string
		mockUpdateSubscription  func(ctx context.Context, userID uuid.UUID, plan string) (*models.Subscription, error)
		expectedStatus          int
		expectedError           string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid request body",
			withAuth:       true,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:     "invalid plan",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan: must be 'pro' or 'enterprise'",
		},
		{
			name:     "same plan error",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "pro",
			},
			mockUpdateSubscription: func(ctx context.Context, uid uuid.UUID, plan string) (*models.Subscription, error) {
				return nil, services.ErrSamePlan
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "already on this plan",
		},
		{
			name:     "no subscription error",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "pro",
			},
			mockUpdateSubscription: func(ctx context.Context, uid uuid.UUID, plan string) (*models.Subscription, error) {
				return nil, services.ErrNoSubscription
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "no active subscription found",
		},
		{
			name:     "invalid plan error from service",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "pro",
			},
			mockUpdateSubscription: func(ctx context.Context, uid uuid.UUID, plan string) (*models.Subscription, error) {
				return nil, services.ErrInvalidPlan
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan",
		},
		{
			name:     "successful update",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "enterprise",
			},
			mockUpdateSubscription: func(ctx context.Context, uid uuid.UUID, plan string) (*models.Subscription, error) {
				return &models.Subscription{
					ID:     uuid.New(),
					UserID: uid,
					Plan:   plan,
					Status: "active",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "internal error",
			withAuth: true,
			requestBody: map[string]string{
				"plan": "pro",
			},
			mockUpdateSubscription: func(ctx context.Context, uid uuid.UUID, plan string) (*models.Subscription, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update subscription",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				UpdateSubscriptionPlanFunc: tt.mockUpdateSubscription,
			}
			handler := newTestBillingHandler(mockBilling, nil, nil, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/billing/subscription", bytes.NewReader(body))
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.UpdateSubscription(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_CancelSubscription(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                   string
		withAuth               bool
		mockCancelSubscription func(ctx context.Context, userID uuid.UUID) error
		expectedStatus         int
		expectedError          string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "successful cancellation",
			withAuth: true,
			mockCancelSubscription: func(ctx context.Context, uid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "error cancelling",
			withAuth: true,
			mockCancelSubscription: func(ctx context.Context, uid uuid.UUID) error {
				return errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to cancel subscription",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				CancelSubscriptionFunc: tt.mockCancelSubscription,
			}
			handler := newTestBillingHandler(mockBilling, nil, nil, nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/billing/subscription", nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.CancelSubscription(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_ListInvoices(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name             string
		withAuth         bool
		mockListInvoices func(ctx context.Context, userID uuid.UUID, limit int) ([]models.Invoice, error)
		expectedStatus   int
		expectedError    string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "successful retrieval",
			withAuth: true,
			mockListInvoices: func(ctx context.Context, uid uuid.UUID, limit int) ([]models.Invoice, error) {
				return []models.Invoice{
					{
						ID:              uuid.New(),
						UserID:          uid,
						StripeInvoiceID: "inv_123",
						Amount:          9900,
						Currency:        "usd",
						Status:          "paid",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "error retrieving invoices",
			withAuth: true,
			mockListInvoices: func(ctx context.Context, uid uuid.UUID, limit int) ([]models.Invoice, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get invoices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				ListInvoicesFunc: tt.mockListInvoices,
			}
			handler := newTestBillingHandler(mockBilling, nil, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/invoices", nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.ListInvoices(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_ListPaymentMethods(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                   string
		withAuth               bool
		mockGetByID            func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockListPaymentMethods func(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error)
		expectedStatus         int
		expectedError          string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "no customer - returns empty list",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "user not found - returns empty list",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "successful retrieval",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockListPaymentMethods: func(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error) {
				return []*stripe.PaymentMethod{
					{
						ID:   "pm_123",
						Type: "card",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "error retrieving payment methods",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockListPaymentMethods: func(ctx context.Context, stripeCustomerID string) ([]*stripe.PaymentMethod, error) {
				return nil, errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get payment methods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				ListPaymentMethodsFunc: tt.mockListPaymentMethods,
			}
			mockUser := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/payment-methods", nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.ListPaymentMethods(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_AddPaymentMethod(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                  string
		withAuth              bool
		mockGetByID           func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockCreateSetupIntent func(ctx context.Context, stripeCustomerID string) (string, error)
		expectedStatus        int
		expectedError         string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "no billing account",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "",
				}, nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no billing account found - please create a subscription first",
		},
		{
			name:     "user not found",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no billing account found - please create a subscription first",
		},
		{
			name:     "successful setup intent creation",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreateSetupIntent: func(ctx context.Context, stripeCustomerID string) (string, error) {
				return "seti_123_secret_456", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "error creating setup intent",
			withAuth: true,
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreateSetupIntent: func(ctx context.Context, stripeCustomerID string) (string, error) {
				return "", errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create setup intent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				CreateSetupIntentFunc: tt.mockCreateSetupIntent,
			}
			mockUser := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/payment-methods/setup", nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.AddPaymentMethod(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_AttachPaymentMethod(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                     string
		withAuth                 bool
		requestBody              map[string]interface{}
		mockGetByID              func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockAttachPaymentMethod  func(ctx context.Context, stripeCustomerID, paymentMethodID string) error
		mockSetDefaultPayment    func(ctx context.Context, stripeCustomerID, paymentMethodID string) error
		expectedStatus           int
		expectedError            string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid request body",
			withAuth:       true,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:     "missing payment_method_id",
			withAuth: true,
			requestBody: map[string]interface{}{
				"payment_method_id": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "payment_method_id is required",
		},
		{
			name:     "successful attachment without default",
			withAuth: true,
			requestBody: map[string]interface{}{
				"payment_method_id": "pm_123",
				"set_as_default":    false,
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockAttachPaymentMethod: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "successful attachment with default",
			withAuth: true,
			requestBody: map[string]interface{}{
				"payment_method_id": "pm_123",
				"set_as_default":    true,
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockAttachPaymentMethod: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return nil
			},
			mockSetDefaultPayment: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "no billing account",
			withAuth: true,
			requestBody: map[string]interface{}{
				"payment_method_id": "pm_123",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "",
				}, nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no billing account found",
		},
		{
			name:     "error attaching payment method",
			withAuth: true,
			requestBody: map[string]interface{}{
				"payment_method_id": "pm_123",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockAttachPaymentMethod: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to attach payment method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				AttachPaymentMethodFunc:     tt.mockAttachPaymentMethod,
				SetDefaultPaymentMethodFunc: tt.mockSetDefaultPayment,
			}
			mockUser := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/payment-methods", bytes.NewReader(body))
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.AttachPaymentMethod(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_RemovePaymentMethod(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                    string
		withAuth                bool
		paymentMethodID         string
		mockDetachPaymentMethod func(ctx context.Context, paymentMethodID string) error
		expectedStatus          int
		expectedError           string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:            "missing payment method id",
			withAuth:        true,
			paymentMethodID: "",
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "payment method ID is required",
		},
		{
			name:            "successful removal",
			withAuth:        true,
			paymentMethodID: "pm_123",
			mockDetachPaymentMethod: func(ctx context.Context, paymentMethodID string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "error removing payment method",
			withAuth:        true,
			paymentMethodID: "pm_123",
			mockDetachPaymentMethod: func(ctx context.Context, paymentMethodID string) error {
				return errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to remove payment method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				DetachPaymentMethodFunc: tt.mockDetachPaymentMethod,
			}
			handler := newTestBillingHandler(mockBilling, nil, nil, nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/billing/payment-methods/"+tt.paymentMethodID, nil)
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}

			// Add URL params using chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.paymentMethodID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			handler.RemovePaymentMethod(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_SetDefaultPaymentMethod(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                        string
		withAuth                    bool
		requestBody                 map[string]string
		mockGetByID                 func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockSetDefaultPaymentMethod func(ctx context.Context, stripeCustomerID, paymentMethodID string) error
		expectedStatus              int
		expectedError               string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:     "successful update",
			withAuth: true,
			requestBody: map[string]string{
				"payment_method_id": "pm_123",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockSetDefaultPaymentMethod: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "no billing account",
			withAuth: true,
			requestBody: map[string]string{
				"payment_method_id": "pm_123",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "",
				}, nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no billing account found",
		},
		{
			name:     "error setting default",
			withAuth: true,
			requestBody: map[string]string{
				"payment_method_id": "pm_123",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockSetDefaultPaymentMethod: func(ctx context.Context, stripeCustomerID, paymentMethodID string) error {
				return errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to set default payment method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				SetDefaultPaymentMethodFunc: tt.mockSetDefaultPaymentMethod,
			}
			mockUser := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/billing/payment-methods/default", bytes.NewReader(body))
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.SetDefaultPaymentMethod(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_CreatePortalSession(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                    string
		withAuth                bool
		requestBody             map[string]string
		mockGetByID             func(ctx context.Context, id uuid.UUID) (*models.User, error)
		mockCreatePortalSession func(ctx context.Context, stripeCustomerID, returnURL string) (string, error)
		expectedStatus          int
		expectedError           string
	}{
		{
			name:           "unauthorized - no auth",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid request body",
			withAuth:       true,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:     "no billing account",
			withAuth: true,
			requestBody: map[string]string{
				"return_url": "https://example.com/billing",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "",
				}, nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "no billing account found",
		},
		{
			name:     "successful portal session creation",
			withAuth: true,
			requestBody: map[string]string{
				"return_url": "https://example.com/billing",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreatePortalSession: func(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
				return "https://billing.stripe.com/session/portal_123", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "error creating portal session",
			withAuth: true,
			requestBody: map[string]string{
				"return_url": "https://example.com/billing",
			},
			mockGetByID: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return &models.User{
					ID:               id,
					StripeCustomerID: "cus_123",
				}, nil
			},
			mockCreatePortalSession: func(ctx context.Context, stripeCustomerID, returnURL string) (string, error) {
				return "", errors.New("stripe error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create portal session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				CreatePortalSessionFunc: tt.mockCreatePortalSession,
			}
			mockUser := &MockUserService{
				GetByIDFunc: tt.mockGetByID,
			}
			handler := newTestBillingHandler(mockBilling, mockUser, nil, nil)

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			} else {
				body = []byte("invalid json")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/portal", bytes.NewReader(body))
			if tt.withAuth {
				req = req.WithContext(createAuthContext(userID))
			}
			rec := httptest.NewRecorder()

			handler.CreatePortalSession(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestBillingHandler_HandleWebhook(t *testing.T) {
	tests := []struct {
		name              string
		payload           string
		signature         string
		mockHandleWebhook func(payload []byte, signature string) (*stripe.Event, error)
		eventType         string
		expectedStatus    int
		expectedError     string
	}{
		{
			name:      "invalid signature",
			payload:   `{"type":"checkout.session.completed"}`,
			signature: "invalid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return nil, services.ErrInvalidWebhook
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid webhook signature",
		},
		{
			name:      "successful checkout.session.completed",
			payload:   `{"type":"checkout.session.completed"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "checkout.session.completed",
					Data: &stripe.EventData{
						Raw: []byte(`{
							"id": "cs_123",
							"subscription": {"id": "sub_123", "current_period_start": 1234567890, "current_period_end": 1234567890},
							"metadata": {"user_id": "` + uuid.New().String() + `", "plan": "pro"},
							"line_items": {"data": [{"price": {"id": "price_123"}}]}
						}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "successful customer.subscription.updated",
			payload:   `{"type":"customer.subscription.updated"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "customer.subscription.updated",
					Data: &stripe.EventData{
						Raw: []byte(`{
							"id": "sub_123",
							"status": "active",
							"cancel_at_period_end": false
						}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "successful customer.subscription.deleted",
			payload:   `{"type":"customer.subscription.deleted"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "customer.subscription.deleted",
					Data: &stripe.EventData{
						Raw: []byte(`{
							"id": "sub_123",
							"customer": {"id": "cus_123"},
							"current_period_end": 1234567890
						}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "successful invoice.paid",
			payload:   `{"type":"invoice.paid"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "invoice.paid",
					Data: &stripe.EventData{
						Raw: []byte(`{
							"id": "in_123",
							"customer": {"id": "cus_123"},
							"amount_paid": 9900,
							"currency": "usd",
							"status": "paid",
							"hosted_invoice_url": "https://invoice.stripe.com/inv_123",
							"invoice_pdf": "https://invoice.stripe.com/inv_123.pdf",
							"period_start": 1234567890,
							"period_end": 1234567890
						}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "successful invoice.payment_failed",
			payload:   `{"type":"invoice.payment_failed"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "invoice.payment_failed",
					Data: &stripe.EventData{
						Raw: []byte(`{
							"id": "in_123",
							"customer": {"id": "cus_123"}
						}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "unknown event type",
			payload:   `{"type":"unknown.event"}`,
			signature: "valid_signature",
			mockHandleWebhook: func(payload []byte, signature string) (*stripe.Event, error) {
				return &stripe.Event{
					Type: "unknown.event",
					Data: &stripe.EventData{
						Raw: []byte(`{}`),
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBilling := &MockBillingService{
				HandleWebhookFunc:             tt.mockHandleWebhook,
				CreateOrUpdateSubscriptionFunc: func(ctx context.Context, userID uuid.UUID, stripeSubID, stripePriceID, plan, status string, periodStart, periodEnd int64) error {
					return nil
				},
				UpdateSubscriptionStatusFunc: func(ctx context.Context, stripeSubID, status string, cancelAtPeriodEnd bool) error {
					return nil
				},
				DeleteSubscriptionFunc: func(ctx context.Context, stripeSubID string) error {
					return nil
				},
				RecordInvoiceFunc: func(ctx context.Context, userID uuid.UUID, stripeInvoiceID string, amount int64, currency, status, invoiceURL, invoicePDF string, periodStart, periodEnd int64) error {
					return nil
				},
				GetUserByStripeCustomerIDFunc: func(ctx context.Context, stripeCustomerID string) (*models.User, error) {
					return &models.User{
						ID:    uuid.New(),
						Email: "test@example.com",
						Name:  "Test User",
					}, nil
				},
			}
			mockLicense := &MockLicenseServiceForHandler{
				CreateLicenseFunc: func(ctx context.Context, userID uuid.UUID, tier string, validDays int, hardwareID string) (*models.License, error) {
					return &models.License{
						ID:     uuid.New(),
						UserID: userID,
						Tier:   tier,
					}, nil
				},
			}
			mockEmail := &MockEmailService{
				SendSubscriptionCanceledEmailFunc: func(ctx context.Context, email, name string, periodEnd time.Time) error {
					return nil
				},
				SendPaymentFailedEmailFunc: func(ctx context.Context, email, name string) error {
					return nil
				},
			}
			handler := newTestBillingHandler(mockBilling, nil, mockLicense, mockEmail)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", bytes.NewReader([]byte(tt.payload)))
			req.Header.Set("Stripe-Signature", tt.signature)
			rec := httptest.NewRecorder()

			handler.HandleWebhook(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

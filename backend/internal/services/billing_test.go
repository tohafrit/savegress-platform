package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v76"
)

// NOTE: Database integration tests and Stripe API tests would require:
// 1. A test database or proper mocking infrastructure
// 2. Stripe test mode configuration and webhook testing
// 3. More complex test setup with transaction rollback
//
// The tests below focus on testing business logic that doesn't require external dependencies

func TestBillingService_GetPriceID(t *testing.T) {
	service := NewBillingService("test_secret_key", "test_webhook_secret")
	service.SetPriceIDs("price_pro_123", "price_enterprise_456")

	tests := []struct {
		name            string
		plan            string
		expectedPriceID string
	}{
		{
			name:            "pro plan",
			plan:            "pro",
			expectedPriceID: "price_pro_123",
		},
		{
			name:            "enterprise plan",
			plan:            "enterprise",
			expectedPriceID: "price_enterprise_456",
		},
		{
			name:            "invalid plan",
			plan:            "invalid",
			expectedPriceID: "",
		},
		{
			name:            "empty plan",
			plan:            "",
			expectedPriceID: "",
		},
		{
			name:            "basic plan (not configured)",
			plan:            "basic",
			expectedPriceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priceID := service.getPriceID(tt.plan)
			assert.Equal(t, tt.expectedPriceID, priceID)
		})
	}
}

func TestBillingService_GetPlanFromPriceID(t *testing.T) {
	service := NewBillingService("test_secret_key", "test_webhook_secret")
	service.SetPriceIDs("price_pro_123", "price_enterprise_456")

	tests := []struct {
		name         string
		priceID      string
		expectedPlan string
	}{
		{
			name:         "pro price ID",
			priceID:      "price_pro_123",
			expectedPlan: "pro",
		},
		{
			name:         "enterprise price ID",
			priceID:      "price_enterprise_456",
			expectedPlan: "enterprise",
		},
		{
			name:         "unknown price ID",
			priceID:      "price_unknown_789",
			expectedPlan: "unknown",
		},
		{
			name:         "empty price ID",
			priceID:      "",
			expectedPlan: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := service.GetPlanFromPriceID(tt.priceID)
			assert.Equal(t, tt.expectedPlan, plan)
		})
	}
}

func TestBillingService_SetPriceIDs(t *testing.T) {
	service := NewBillingService("test_key", "test_webhook")

	proPriceID := "price_pro_new"
	enterprisePriceID := "price_enterprise_new"

	service.SetPriceIDs(proPriceID, enterprisePriceID)

	assert.Equal(t, proPriceID, service.getPriceID("pro"))
	assert.Equal(t, enterprisePriceID, service.getPriceID("enterprise"))
}

func TestNewBillingService(t *testing.T) {
	secretKey := "sk_test_123"
	webhookSecret := "whsec_test_456"

	service := NewBillingService(secretKey, webhookSecret)

	assert.NotNil(t, service)
	assert.Equal(t, webhookSecret, service.webhookSecret)
	// Verify Stripe key was set globally
	assert.Equal(t, secretKey, stripe.Key)
}

func TestBillingService_ErrorConstants(t *testing.T) {
	// Test that error constants are defined correctly
	assert.NotNil(t, ErrNoSubscription)
	assert.NotNil(t, ErrInvalidWebhook)
	assert.NotNil(t, ErrInvalidPlan)
	assert.NotNil(t, ErrSamePlan)
	assert.NotNil(t, ErrPaymentMethodNotFound)

	assert.Equal(t, "no active subscription", ErrNoSubscription.Error())
	assert.Equal(t, "invalid webhook signature", ErrInvalidWebhook.Error())
	assert.Equal(t, "invalid plan", ErrInvalidPlan.Error())
	assert.Equal(t, "already on this plan", ErrSamePlan.Error())
	assert.Equal(t, "payment method not found", ErrPaymentMethodNotFound.Error())
}

func TestBillingService_HandleWebhook_InvalidSignature(t *testing.T) {
	service := NewBillingService("test_key", "test_webhook_secret")

	payload := []byte(`{"type": "customer.created"}`)
	signature := "invalid_signature"

	event, err := service.HandleWebhook(payload, signature)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidWebhook, err)
	assert.Nil(t, event)
}

func TestBillingService_SetDB(t *testing.T) {
	service := NewBillingService("test_key", "test_webhook")

	// SetDB should not panic with nil
	service.SetDB(nil)
	assert.NotNil(t, service)
}

func TestBillingService_PlanValidation(t *testing.T) {
	service := NewBillingService("test_key", "test_webhook")
	service.SetPriceIDs("price_pro", "price_ent")

	tests := []struct {
		name     string
		plan     string
		isValid  bool
	}{
		{
			name:    "valid pro plan",
			plan:    "pro",
			isValid: true,
		},
		{
			name:    "valid enterprise plan",
			plan:    "enterprise",
			isValid: true,
		},
		{
			name:    "invalid basic plan",
			plan:    "basic",
			isValid: false,
		},
		{
			name:    "invalid free plan",
			plan:    "free",
			isValid: false,
		},
		{
			name:    "empty plan",
			plan:    "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priceID := service.getPriceID(tt.plan)
			if tt.isValid {
				assert.NotEmpty(t, priceID, "Expected valid plan to have a price ID")
			} else {
				assert.Empty(t, priceID, "Expected invalid plan to have empty price ID")
			}
		})
	}
}

func TestBillingService_PlanMapping(t *testing.T) {
	// Test bidirectional mapping between plans and price IDs
	service := NewBillingService("test_key", "test_webhook")

	proPriceID := "price_1234_pro"
	entPriceID := "price_5678_ent"

	service.SetPriceIDs(proPriceID, entPriceID)

	// Test plan -> price ID
	assert.Equal(t, proPriceID, service.getPriceID("pro"))
	assert.Equal(t, entPriceID, service.getPriceID("enterprise"))

	// Test price ID -> plan
	assert.Equal(t, "pro", service.GetPlanFromPriceID(proPriceID))
	assert.Equal(t, "enterprise", service.GetPlanFromPriceID(entPriceID))

	// Test round-trip
	planName := "pro"
	priceID := service.getPriceID(planName)
	recoveredPlan := service.GetPlanFromPriceID(priceID)
	assert.Equal(t, planName, recoveredPlan)
}

func TestBillingService_SubscriptionStatuses(t *testing.T) {
	// Document valid subscription statuses
	validStatuses := []string{
		"active",
		"past_due",
		"canceled",
		"trialing",
		"incomplete",
		"incomplete_expired",
		"unpaid",
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, status, "Status should not be empty")
	}
}

func TestBillingService_InvoiceStatuses(t *testing.T) {
	// Document valid invoice statuses
	validStatuses := []string{
		"draft",
		"open",
		"paid",
		"void",
		"uncollectible",
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, status, "Status should not be empty")
	}
}

// Integration test examples (commented out - would need real database and Stripe test mode)
//
// func TestBillingService_CreateCustomerIntegration(t *testing.T) {
//     t.Skip("Requires Stripe test mode and database connection")
//     // This would test creating a real Stripe customer
// }
//
// func TestBillingService_CreateCheckoutSessionIntegration(t *testing.T) {
//     t.Skip("Requires Stripe test mode")
//     // This would test creating a real checkout session
// }
//
// func TestBillingService_GetSubscriptionIntegration(t *testing.T) {
//     t.Skip("Requires database connection")
//     // This would test retrieving subscription from database
// }
//
// func TestBillingService_WebhookHandlingIntegration(t *testing.T) {
//     t.Skip("Requires Stripe webhook testing setup")
//     // This would test handling real Stripe webhooks
// }
//
// func TestBillingService_SubscriptionLifecycleIntegration(t *testing.T) {
//     t.Skip("Requires Stripe test mode and database")
//     // This would test: create -> update -> cancel -> reactivate
// }

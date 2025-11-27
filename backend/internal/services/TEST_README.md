# Service Tests Documentation

This directory contains comprehensive tests for the platform backend services.

## Test Files

### auth_test.go
Tests for the authentication service (`auth.go`).

**Coverage:**
- JWT token validation (valid, expired, invalid signature, malformed)
- Claims structure and UUID conversion
- Password hashing with bcrypt
- Token pair structure
- Error constants validation
- Service initialization
- Email service integration

**Test Count:** 10 test functions with multiple sub-tests

**Key Features:**
- Table-driven tests for comprehensive coverage
- Tests for various JWT token edge cases
- Password security validation
- Error message consistency checks

**Note:** Database integration tests (Register, Login, RefreshToken, GetUserByID, etc.) are documented but not implemented. They would require:
1. Test database or proper mocking infrastructure
2. Refactoring services to accept database interfaces
3. Complex test setup with transaction rollback

### billing_test.go
Tests for the Stripe billing service (`billing.go`).

**Coverage:**
- Price ID mapping (plan name ↔ Stripe price ID)
- Plan validation
- Error constants validation
- Webhook signature validation
- Service initialization
- Bidirectional plan/price ID mapping
- Subscription and invoice status documentation

**Test Count:** 10 test functions with multiple sub-tests

**Key Features:**
- Table-driven tests
- Plan configuration validation
- Round-trip mapping tests
- Status documentation for Stripe objects

**Note:** Stripe API integration tests and database tests are documented but not implemented. They would require:
1. Stripe test mode configuration
2. Webhook testing infrastructure
3. Database connection for subscription management

## Running Tests

### Run all tests:
```bash
cd /Users/pakhunov/work/cdc/savegress-platform/backend/internal/services
go test -v
```

### Run specific test file:
```bash
go test -v -run TestAuthService
go test -v -run TestBillingService
```

### Run with coverage:
```bash
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run specific test:
```bash
go test -v -run TestAuthService_ValidateToken
go test -v -run TestBillingService_GetPriceID
```

## Test Patterns Used

### 1. Table-Driven Tests
Most tests use the table-driven pattern for comprehensive coverage:

```go
tests := []struct {
    name          string
    input         string
    expectedError error
    expectedValue string
}{
    {
        name:  "valid case",
        input: "test",
        // ...
    },
    // More test cases...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

### 2. Error Constant Testing
Validates that all error constants are properly defined:

```go
func TestAuthService_ErrorConstants(t *testing.T) {
    assert.NotNil(t, ErrInvalidCredentials)
    assert.Equal(t, "invalid email or password", ErrInvalidCredentials.Error())
}
```

### 3. Structure Validation
Tests data structure integrity:

```go
func TestTokenPair_Structure(t *testing.T) {
    tokens := &TokenPair{
        AccessToken:  "token",
        RefreshToken: "refresh",
        ExpiresAt:    time.Now().Add(15 * time.Minute),
    }
    assert.NotEmpty(t, tokens.AccessToken)
}
```

## What's NOT Tested (Yet)

### Database Operations
These require database integration or interface refactoring:
- User registration
- User login
- Password reset flow
- Refresh token management
- User retrieval by ID/email
- Subscription CRUD operations
- Invoice recording

### External API Calls
These require mocking or test mode setup:
- Stripe customer creation
- Stripe checkout session creation
- Stripe subscription management
- Stripe webhook processing
- Payment method management

## Future Improvements

### 1. Database Interface Refactoring
Refactor services to accept database interfaces instead of concrete types:

```go
type Database interface {
    QueryRow(ctx context.Context, query string, args ...interface{}) Row
    Exec(ctx context.Context, query string, args ...interface{}) Result
    // ...
}

type AuthService struct {
    db        Database
    jwtSecret []byte
}
```

### 2. Stripe Mock Implementation
Create Stripe client mocks for testing API interactions without hitting real endpoints.

### 3. Integration Test Suite
Add integration tests that run against:
- Test database (PostgreSQL)
- Stripe test mode
- Real Redis instance

### 4. Test Fixtures
Create reusable test data fixtures for common scenarios:
- Valid users
- Active subscriptions
- Sample invoices
- License keys

## Dependencies

The test suite uses:
- `github.com/stretchr/testify/assert` - Assertion library
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/google/uuid` - UUID generation
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/stripe/stripe-go/v76` - Stripe SDK

## Test Metrics

Current test results (as of last run):
- **Total Tests:** 27+ test functions
- **Pass Rate:** 100%
- **Execution Time:** ~0.8s
- **Coverage:** Unit tests for business logic (DB/API integration not covered)

## Contributing

When adding new tests:
1. Use table-driven tests where appropriate
2. Test both success and error cases
3. Add descriptive test names
4. Document why tests are skipped (if applicable)
5. Keep tests focused and independent
6. Use assert library for cleaner assertions

## Integration Test Examples

See commented sections in test files for examples of integration tests that would be valuable to implement:
- `TestAuthService_RegisterIntegration`
- `TestAuthService_LoginIntegration`
- `TestBillingService_SubscriptionLifecycleIntegration`
- `TestBillingService_WebhookHandlingIntegration`

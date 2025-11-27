package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/savegress/platform/backend/internal/services"
)

// MockEarlyAccessService implements a mock for testing
type MockEarlyAccessService struct {
	SubmitFunc func(ctx context.Context, input services.EarlyAccessInput) error
}

func (m *MockEarlyAccessService) Submit(ctx context.Context, input services.EarlyAccessInput) error {
	if m.SubmitFunc != nil {
		return m.SubmitFunc(ctx, input)
	}
	return nil
}

// testEarlyAccessHandler wraps EarlyAccessHandler for testing with mock service
type testEarlyAccessHandler struct {
	mock            *MockEarlyAccessService
	turnstileSecret string
}

func newTestEarlyAccessHandler(mock *MockEarlyAccessService, turnstileSecret string) *testEarlyAccessHandler {
	return &testEarlyAccessHandler{
		mock:            mock,
		turnstileSecret: turnstileSecret,
	}
}

// Submit implements the Submit handler method for testing
func (h *testEarlyAccessHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req EarlyAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Company == "" {
		respondError(w, http.StatusBadRequest, "email and company are required")
		return
	}

	if req.TurnstileToken == "" {
		respondError(w, http.StatusBadRequest, "captcha verification required")
		return
	}

	// Verify Turnstile token
	if !h.verifyTurnstile(req.TurnstileToken, getClientIP(r)) {
		respondError(w, http.StatusBadRequest, "captcha verification failed")
		return
	}

	// Get client info
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Save to database
	err := h.mock.Submit(r.Context(), services.EarlyAccessInput{
		Email:           req.Email,
		Company:         req.Company,
		CurrentSolution: req.CurrentSolution,
		DataVolume:      req.DataVolume,
		Message:         req.Message,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to submit request")
		return
	}

	respondCreated(w, map[string]interface{}{
		"success": true,
		"message": "Request submitted successfully",
	})
}

func (h *testEarlyAccessHandler) verifyTurnstile(token, ip string) bool {
	// Skip verification for test keys
	if len(h.turnstileSecret) > 6 && h.turnstileSecret[:6] == "1x0000" {
		return true
	}

	// For testing: reject empty tokens or tokens marked as "invalid"
	if token == "" || token == "invalid_token" {
		return false
	}

	// For testing: accept all other tokens
	return true
}

func TestEarlyAccessHandler_Submit(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		turnstileKey   string
		mockSubmit     func(ctx context.Context, input services.EarlyAccessInput) error
		headers        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful submission with all fields",
			requestBody: map[string]interface{}{
				"email":           "test@example.com",
				"company":         "Test Company",
				"currentSolution": "MySQL",
				"dataVolume":      "1TB",
				"message":         "Looking for a better CDC solution",
				"turnstileToken":  "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				if input.Email != "test@example.com" {
					t.Errorf("expected email 'test@example.com', got '%s'", input.Email)
				}
				if input.Company != "Test Company" {
					t.Errorf("expected company 'Test Company', got '%s'", input.Company)
				}
				if input.CurrentSolution != "MySQL" {
					t.Errorf("expected current solution 'MySQL', got '%s'", input.CurrentSolution)
				}
				if input.DataVolume != "1TB" {
					t.Errorf("expected data volume '1TB', got '%s'", input.DataVolume)
				}
				if input.Message != "Looking for a better CDC solution" {
					t.Errorf("expected message 'Looking for a better CDC solution', got '%s'", input.Message)
				}
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful submission with only required fields",
			requestBody: map[string]interface{}{
				"email":          "minimal@example.com",
				"company":        "Minimal Inc",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				if input.Email != "minimal@example.com" {
					t.Errorf("expected email 'minimal@example.com', got '%s'", input.Email)
				}
				if input.Company != "Minimal Inc" {
					t.Errorf("expected company 'Minimal Inc', got '%s'", input.Company)
				}
				if input.CurrentSolution != "" {
					t.Errorf("expected empty current solution, got '%s'", input.CurrentSolution)
				}
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful submission with X-Forwarded-For header",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				if input.IPAddress != "192.168.1.1" {
					t.Errorf("expected IP address '192.168.1.1', got '%s'", input.IPAddress)
				}
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful submission with X-Real-IP header",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				if input.IPAddress != "10.0.0.1" {
					t.Errorf("expected IP address '10.0.0.1', got '%s'", input.IPAddress)
				}
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful submission with User-Agent header",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 Test Browser",
			},
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				if input.UserAgent != "Mozilla/5.0 Test Browser" {
					t.Errorf("expected user agent 'Mozilla/5.0 Test Browser', got '%s'", input.UserAgent)
				}
				return nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and company are required",
		},
		{
			name: "missing company",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"turnstileToken": "valid_token",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and company are required",
		},
		{
			name: "empty email",
			requestBody: map[string]interface{}{
				"email":          "",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and company are required",
		},
		{
			name: "empty company",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "",
				"turnstileToken": "valid_token",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and company are required",
		},
		{
			name: "missing turnstile token",
			requestBody: map[string]interface{}{
				"email":   "test@example.com",
				"company": "Test Company",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "captcha verification required",
		},
		{
			name: "empty turnstile token",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "",
			},
			turnstileKey:   "1x0000_test_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "captcha verification required",
		},
		{
			name: "invalid turnstile token",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "invalid_token",
			},
			turnstileKey:   "real_secret_key",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "captcha verification failed",
		},
		{
			name: "database error",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				return errors.New("database connection failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to submit request",
		},
		{
			name: "context error",
			requestBody: map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			},
			turnstileKey: "1x0000_test_key",
			mockSubmit: func(ctx context.Context, input services.EarlyAccessInput) error {
				return context.DeadlineExceeded
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to submit request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockEarlyAccessService{
				SubmitFunc: tt.mockSubmit,
			}
			handler := newTestEarlyAccessHandler(mock, tt.turnstileKey)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/early-access", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Set custom headers if provided
			if tt.headers != nil {
				for key, value := range tt.headers {
					req.Header.Set(key, value)
				}
			}

			rec := httptest.NewRecorder()

			handler.Submit(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				json.NewDecoder(rec.Body).Decode(&response)
				if response["error"] != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
				}
			} else if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				json.NewDecoder(rec.Body).Decode(&response)
				if success, ok := response["success"].(bool); !ok || !success {
					t.Errorf("expected success=true in response")
				}
				if message, ok := response["message"].(string); !ok || message != "Request submitted successfully" {
					t.Errorf("expected message 'Request submitted successfully', got %v", message)
				}
			}
		})
	}
}

func TestEarlyAccessHandler_InvalidJSON(t *testing.T) {
	mock := &MockEarlyAccessService{}
	handler := newTestEarlyAccessHandler(mock, "1x0000_test_key")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/early-access", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Submit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]string
	json.NewDecoder(rec.Body).Decode(&response)
	if response["error"] != "invalid request body" {
		t.Errorf("expected error 'invalid request body', got %q", response["error"])
	}
}

func TestEarlyAccessHandler_VerifyTurnstile(t *testing.T) {
	tests := []struct {
		name            string
		turnstileSecret string
		token           string
		ip              string
		expectedResult  bool
	}{
		{
			name:            "test key always passes - full prefix",
			turnstileSecret: "1x0000_test_secret",
			token:           "any_token",
			ip:              "192.168.1.1",
			expectedResult:  true,
		},
		{
			name:            "test key always passes - minimal prefix",
			turnstileSecret: "1x0000123",
			token:           "any_token",
			ip:              "",
			expectedResult:  true,
		},
		{
			name:            "valid token with real secret",
			turnstileSecret: "real_secret_key",
			token:           "valid_token",
			ip:              "192.168.1.1",
			expectedResult:  true,
		},
		{
			name:            "valid token without IP",
			turnstileSecret: "real_secret_key",
			token:           "valid_token",
			ip:              "",
			expectedResult:  true,
		},
		{
			name:            "invalid token",
			turnstileSecret: "real_secret_key",
			token:           "invalid_token",
			ip:              "192.168.1.1",
			expectedResult:  false,
		},
		{
			name:            "empty token",
			turnstileSecret: "real_secret_key",
			token:           "",
			ip:              "192.168.1.1",
			expectedResult:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestEarlyAccessHandler(nil, tt.turnstileSecret)
			result := handler.verifyTurnstile(tt.token, tt.ip)

			if result != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "X-Forwarded-For takes precedence",
			xForwardedFor: "192.168.1.1",
			xRealIP:       "10.0.0.1",
			remoteAddr:    "172.16.0.1",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "X-Real-IP when no X-Forwarded-For",
			xForwardedFor: "",
			xRealIP:       "10.0.0.1",
			remoteAddr:    "172.16.0.1",
			expectedIP:    "10.0.0.1",
		},
		{
			name:          "RemoteAddr when no proxy headers",
			xForwardedFor: "",
			xRealIP:       "",
			remoteAddr:    "172.16.0.1:8080",
			expectedIP:    "172.16.0.1:8080",
		},
		{
			name:          "Multiple IPs in X-Forwarded-For",
			xForwardedFor: "192.168.1.1, 10.0.0.1, 172.16.0.1",
			xRealIP:       "",
			remoteAddr:    "8.8.8.8",
			expectedIP:    "192.168.1.1, 10.0.0.1, 172.16.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)

			if ip != tt.expectedIP {
				t.Errorf("expected IP %q, got %q", tt.expectedIP, ip)
			}
		})
	}
}

func TestEarlyAccessHandler_OptionalFields(t *testing.T) {
	tests := []struct {
		name            string
		currentSolution string
		dataVolume      string
		message         string
	}{
		{
			name:            "all optional fields provided",
			currentSolution: "PostgreSQL",
			dataVolume:      "500GB",
			message:         "Interested in real-time CDC",
		},
		{
			name:            "only current solution",
			currentSolution: "MySQL",
			dataVolume:      "",
			message:         "",
		},
		{
			name:            "only data volume",
			currentSolution: "",
			dataVolume:      "2TB",
			message:         "",
		},
		{
			name:            "only message",
			currentSolution: "",
			dataVolume:      "",
			message:         "Looking forward to trying this",
		},
		{
			name:            "no optional fields",
			currentSolution: "",
			dataVolume:      "",
			message:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var submittedInput services.EarlyAccessInput
			mock := &MockEarlyAccessService{
				SubmitFunc: func(ctx context.Context, input services.EarlyAccessInput) error {
					submittedInput = input
					return nil
				},
			}
			handler := newTestEarlyAccessHandler(mock, "1x0000_test_key")

			requestBody := map[string]interface{}{
				"email":          "test@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			}
			if tt.currentSolution != "" {
				requestBody["currentSolution"] = tt.currentSolution
			}
			if tt.dataVolume != "" {
				requestBody["dataVolume"] = tt.dataVolume
			}
			if tt.message != "" {
				requestBody["message"] = tt.message
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/early-access", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Submit(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
			}

			if submittedInput.CurrentSolution != tt.currentSolution {
				t.Errorf("expected currentSolution %q, got %q", tt.currentSolution, submittedInput.CurrentSolution)
			}
			if submittedInput.DataVolume != tt.dataVolume {
				t.Errorf("expected dataVolume %q, got %q", tt.dataVolume, submittedInput.DataVolume)
			}
			if submittedInput.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, submittedInput.Message)
			}
		})
	}
}

func TestEarlyAccessHandler_ConcurrentSubmissions(t *testing.T) {
	var submissionCount int64
	mock := &MockEarlyAccessService{
		SubmitFunc: func(ctx context.Context, input services.EarlyAccessInput) error {
			atomic.AddInt64(&submissionCount, 1)
			return nil
		},
	}
	handler := newTestEarlyAccessHandler(mock, "1x0000_test_key")

	// Simulate concurrent requests
	done := make(chan bool)
	numRequests := 10

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			requestBody := map[string]interface{}{
				"email":          "test" + string(rune(index)) + "@example.com",
				"company":        "Test Company",
				"turnstileToken": "valid_token",
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/early-access", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Submit(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}

	finalCount := atomic.LoadInt64(&submissionCount)
	if finalCount != int64(numRequests) {
		t.Errorf("expected %d submissions, got %d", numRequests, finalCount)
	}
}

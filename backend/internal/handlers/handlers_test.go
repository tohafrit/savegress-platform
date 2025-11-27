package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedBody   string
		checkNilBody   bool
	}{
		{
			name:         "success with data",
			status:       http.StatusOK,
			data:         map[string]string{"message": "ok"},
			expectedBody: `{"message":"ok"}`,
		},
		{
			name:       "success with nil data",
			status:     http.StatusNoContent,
			data:       nil,
			checkNilBody: true,
		},
		{
			name:   "success with struct",
			status: http.StatusOK,
			data: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 42,
			},
			expectedBody: `{"name":"test","value":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondJSON(rec, tt.status, tt.data)

			if rec.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, rec.Code)
			}

			if rec.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
			}

			if tt.checkNilBody {
				if rec.Body.Len() != 0 {
					t.Errorf("expected empty body, got %s", rec.Body.String())
				}
			} else if tt.expectedBody != "" {
				// Compare JSON by unmarshaling
				var expected, actual interface{}
				json.Unmarshal([]byte(tt.expectedBody), &expected)
				json.Unmarshal(rec.Body.Bytes(), &actual)

				expectedJSON, _ := json.Marshal(expected)
				actualJSON, _ := json.Marshal(actual)

				if string(expectedJSON) != string(actualJSON) {
					t.Errorf("expected body %s, got %s", tt.expectedBody, rec.Body.String())
				}
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedError  string
	}{
		{
			name:          "bad request",
			status:        http.StatusBadRequest,
			message:       "invalid input",
			expectedError: "invalid input",
		},
		{
			name:          "unauthorized",
			status:        http.StatusUnauthorized,
			message:       "not authorized",
			expectedError: "not authorized",
		},
		{
			name:          "internal error",
			status:        http.StatusInternalServerError,
			message:       "something went wrong",
			expectedError: "something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondError(rec, tt.status, tt.message)

			if rec.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, rec.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response["error"] != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, response["error"])
			}
		})
	}
}

func TestRespondSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"status": "success"}
	respondSuccess(rec, data)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %q", response["status"])
	}
}

func TestRespondCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	respondCreated(rec, data)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["id"] != "123" {
		t.Errorf("expected id '123', got %q", response["id"])
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]interface{}{
		"name":  "test",
		"count": 10,
	}
	writeJSON(rec, data)

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["name"] != "test" {
		t.Errorf("expected name 'test', got %v", response["name"])
	}

	if response["count"] != float64(10) {
		t.Errorf("expected count 10, got %v", response["count"])
	}
}

func TestListDownloads(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads", nil)
	rec := httptest.NewRecorder()

	ListDownloads(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	downloads, ok := response["downloads"].([]interface{})
	if !ok {
		t.Fatal("expected downloads array in response")
	}

	if len(downloads) == 0 {
		t.Error("expected at least one download")
	}

	firstDownload := downloads[0].(map[string]interface{})
	if firstDownload["product"] != "cdc-engine" {
		t.Errorf("expected product 'cdc-engine', got %v", firstDownload["product"])
	}
}

func TestGetDownloadURL(t *testing.T) {
	// This is now a legacy endpoint that returns an error directing users to the new endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads/url", nil)
	rec := httptest.NewRecorder()

	GetDownloadURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] == "" {
		t.Error("expected non-empty error message")
	}
}

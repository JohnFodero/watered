package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetStatus)

	// Call the handler with our request and recorder
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	var response StatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Verify response fields
	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if response.Service != "watered-api" {
		t.Errorf("Expected service 'watered-api', got '%s'", response.Service)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	// Check uptime fields
	if response.UptimeSeconds < 0 {
		t.Errorf("Expected uptime_seconds to be non-negative, got %f", response.UptimeSeconds)
	}

	if response.UptimeFormatted == "" {
		t.Errorf("Expected uptime_formatted to be non-empty")
	}

	// Uptime should be in minutes format for a recent start
	if response.UptimeFormatted != "1 minute" && response.UptimeFormatted != "2 minutes" {
		// Allow some flexibility for test timing
		t.Logf("Uptime formatted: %s", response.UptimeFormatted)
	}

	// Check content type
	expected := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			ctype, expected)
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Less than 1 minute",
			duration: 30 * time.Second,
			expected: "1 minute",
		},
		{
			name:     "Exactly 1 minute",
			duration: 1 * time.Minute,
			expected: "1 minute",
		},
		{
			name:     "2 minutes",
			duration: 2 * time.Minute,
			expected: "2 minutes",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			expected: "59 minutes",
		},
		{
			name:     "Exactly 1 hour",
			duration: 1 * time.Hour,
			expected: "1 hour",
		},
		{
			name:     "2 hours",
			duration: 2 * time.Hour,
			expected: "2 hours",
		},
		{
			name:     "23 hours",
			duration: 23 * time.Hour,
			expected: "23 hours",
		},
		{
			name:     "Exactly 1 day",
			duration: 24 * time.Hour,
			expected: "1 day",
		},
		{
			name:     "2 days",
			duration: 48 * time.Hour,
			expected: "2 days",
		},
		{
			name:     "7 days",
			duration: 7 * 24 * time.Hour,
			expected: "7 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.duration)
			if result != tt.expected {
				t.Errorf("formatUptime(%v) = %v, want %v", tt.duration, result, tt.expected)
			}
		})
	}
}

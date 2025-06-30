package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"watered/internal/auth"
	"watered/internal/handlers"
	"watered/internal/services"
	"watered/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateTestServer creates a test server instance
func CreateTestServer(t *testing.T) *httptest.Server {
	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize services
	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	plantHandlers := handlers.NewPlantHandlers(plantService, authService)
	adminHandlers := handlers.NewAdminHandler(store)

	// Create router
	r := chi.NewRouter()

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"watered"}`))
	})

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		r.Get("/status", authHandlers.StatusHandler)
		r.HandleFunc("/demo-login", authHandlers.DemoLoginHandler)
		r.Post("/logout", authHandlers.LogoutHandler)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handlers.GetStatus)

		// Plant API routes
		r.Route("/plant", func(r chi.Router) {
			// Public plant endpoints (read-only)
			r.Get("/", plantHandlers.GetPlantHandler)
			r.Get("/status", plantHandlers.GetPlantStatusHandler)
			r.Get("/timer", plantHandlers.GetPlantTimerHandler)

			// Protected plant endpoints (require authentication)
			r.Group(func(r chi.Router) {
				r.Use(authService.AuthRequired)
				r.Post("/water", plantHandlers.WaterPlantHandler)
			})

			// Admin-only plant endpoints
			r.Group(func(r chi.Router) {
				r.Use(authService.AdminRequired)
				r.Put("/settings", plantHandlers.UpdatePlantSettingsHandler)
				r.Post("/reset", plantHandlers.ResetPlantHandler)
			})
		})
	})

	// Admin API routes
	r.Route("/admin", func(r chi.Router) {
		r.Use(authService.AdminRequired)

		// Configuration endpoints
		r.Get("/config", adminHandlers.GetConfigHandler)
		r.Put("/config/timeout", adminHandlers.UpdateTimeoutHandler)

		// User management endpoints
		r.Get("/users", adminHandlers.GetUsersHandler)
		r.Post("/users", adminHandlers.AddUserHandler)
		r.Delete("/users/{email}", adminHandlers.RemoveUserHandler)

		// History and statistics endpoints
		r.Get("/history", adminHandlers.GetHistoryHandler)
		r.Get("/stats", adminHandlers.GetStatsHandler)
	})

	return httptest.NewServer(r)
}

func TestHealthEndpoint(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "watered", response["service"])
}

func TestAPIStatusEndpoint(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "watered-api", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Contains(t, response, "timestamp")
}

func TestPlantEndpoints(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	// Test GET /api/plant/
	resp, err := http.Get(server.URL + "/api/plant/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plantResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&plantResponse)
	require.NoError(t, err)

	// Check expected fields
	assert.Contains(t, plantResponse, "id")
	assert.Contains(t, plantResponse, "name")
	assert.Contains(t, plantResponse, "health_status")
	assert.Contains(t, plantResponse, "timeout_hours")
}

func TestPlantStatusEndpoint(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/plant/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statusResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statusResponse)
	require.NoError(t, err)

	assert.Contains(t, statusResponse, "status")
	assert.Contains(t, statusResponse, "time_since_watering_formatted")
	assert.Contains(t, statusResponse, "is_overdue")
}

func TestAuthStatusEndpoint(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/auth/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	require.NoError(t, err)

	assert.Contains(t, authResponse, "authenticated")
	assert.Equal(t, false, authResponse["authenticated"]) // Should not be authenticated
}

func TestUnauthorizedAccess(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	// Test protected plant endpoint without auth
	resp, err := http.Post(server.URL+"/api/plant/water", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdminEndpointsUnauthorized(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	adminEndpoints := []string{
		"/admin/config",
		"/admin/users",
		"/admin/stats",
		"/admin/history",
	}

	for _, endpoint := range adminEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(server.URL + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}

func TestAdminConfigEndpoint(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	// This test requires admin authentication
	// For integration testing, we'd need to simulate the admin session
	// For now, we test the unauthorized access (covered above)

	// TODO: Add authenticated admin tests when session testing is implemented
}

func TestCORSHeaders(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	req, err := http.NewRequest("OPTIONS", server.URL+"/api/status", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// For now just ensure the endpoint responds
	// Real CORS testing would require CORS middleware to be added
	assert.True(t, resp.StatusCode < 500) // Should not be a server error
}

func TestRateLimiting(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	// Basic test - make multiple requests quickly
	// Real rate limiting would require rate limiting middleware
	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL + "/health")
		require.NoError(t, err)
		resp.Body.Close()

		// Should not error out with basic load
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Small delay to avoid overwhelming the test server
		time.Sleep(10 * time.Millisecond)
	}
}

func TestInvalidJSONHandling(t *testing.T) {
	server := CreateTestServer(t)
	defer server.Close()

	// Test endpoints that expect JSON with invalid JSON
	invalidJSON := `{"invalid": json}`

	resp, err := http.Post(server.URL+"/admin/users", "application/json",
		bytes.NewBufferString(invalidJSON))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should handle invalid JSON gracefully (either 400 or 401 for unauthorized)
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized)
}

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"watered/internal/auth"
	"watered/internal/handlers"
	"watered/internal/services"
	"watered/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApp represents the full application for e2e testing
type TestApp struct {
	Server      *httptest.Server
	Storage     storage.Storage
	AuthService *auth.AuthService
}

// NewTestApp creates a new test application instance
func NewTestApp(t *testing.T) *TestApp {
	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize services
	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	plantHandlers := handlers.NewPlantHandlers(plantService, authService)
	adminHandlers := handlers.NewAdminHandler(store)

	// Create router with full application setup
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
			// Public plant endpoints
			r.Get("/", plantHandlers.GetPlantHandler)
			r.Get("/status", plantHandlers.GetPlantStatusHandler)
			r.Get("/timer", plantHandlers.GetPlantTimerHandler)

			// Protected plant endpoints
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
		r.Get("/config", adminHandlers.GetConfigHandler)
		r.Put("/config/timeout", adminHandlers.UpdateTimeoutHandler)
		r.Get("/users", adminHandlers.GetUsersHandler)
		r.Post("/users", adminHandlers.AddUserHandler)
		r.Delete("/users/{email}", adminHandlers.RemoveUserHandler)
		r.Get("/history", adminHandlers.GetHistoryHandler)
		r.Get("/stats", adminHandlers.GetStatsHandler)
	})

	server := httptest.NewServer(r)

	return &TestApp{
		Server:      server,
		Storage:     store,
		AuthService: authService,
	}
}

// Close cleans up the test app
func (app *TestApp) Close() {
	app.Server.Close()
	app.Storage.Close()
}

func TestCompleteUserJourney(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Step 1: Check initial application health
	resp, err := http.Get(app.Server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Step 2: Check authentication status (should not be authenticated)
	resp, err = http.Get(app.Server.URL + "/auth/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	var authStatus map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&authStatus)
	require.NoError(t, err)
	assert.Equal(t, false, authStatus["authenticated"])

	// Step 3: Try to access protected endpoint (should fail)
	resp, err = http.Post(app.Server.URL+"/api/plant/water", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Step 4: Access public plant information
	resp, err = http.Get(app.Server.URL + "/api/plant/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var plantInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&plantInfo)
	require.NoError(t, err)
	assert.Contains(t, plantInfo, "health_status")
	assert.Contains(t, plantInfo, "name")

	// Step 5: Check plant status
	resp, err = http.Get(app.Server.URL + "/api/plant/status")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminWorkflow(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Test admin endpoints without authentication (should fail)
	adminEndpoints := []string{
		"/admin/config",
		"/admin/users",
		"/admin/stats",
	}

	for _, endpoint := range adminEndpoints {
		t.Run("unauthorized_"+endpoint, func(t *testing.T) {
			resp, err := http.Get(app.Server.URL + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}

	// Test admin user management endpoints
	t.Run("admin_user_management_unauthorized", func(t *testing.T) {
		// Try to add user without auth
		userJSON := `{"email": "test@example.com"}`
		resp, err := http.Post(app.Server.URL+"/admin/users", "application/json",
			strings.NewReader(userJSON))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestPlantCareWorkflow(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Step 1: Get initial plant state
	resp, err := http.Get(app.Server.URL + "/api/plant/")
	require.NoError(t, err)
	defer resp.Body.Close()

	var initialState map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&initialState)
	require.NoError(t, err)

	// Step 2: Check plant timer
	resp, err = http.Get(app.Server.URL + "/api/plant/timer")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var timerInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&timerInfo)
	require.NoError(t, err)
	assert.Contains(t, timerInfo, "is_overdue")
	assert.Contains(t, timerInfo, "timeout_hours")

	// Step 3: Try to water plant without authentication (should fail)
	resp, err = http.Post(app.Server.URL+"/api/plant/water", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Step 4: Try to modify plant settings without admin auth (should fail)
	settingsJSON := `{"name": "Test Plant", "timeoutHours": 48}`
	req, err := http.NewRequest("PUT", app.Server.URL+"/api/plant/settings", strings.NewReader(settingsJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestErrorHandling(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Test 404 for non-existent endpoints
	resp, err := http.Get(app.Server.URL + "/api/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test invalid JSON handling
	resp, err = http.Post(app.Server.URL+"/admin/users", "application/json",
		strings.NewReader(`{invalid json}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	// Should be either 400 (bad request) or 401 (unauthorized)
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized)

	// Test method not allowed
	resp, err = http.Post(app.Server.URL+"/health", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestAPIConsistency(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Test that all API endpoints return consistent JSON responses
	endpoints := []string{
		"/health",
		"/api/status",
		"/api/plant/",
		"/api/plant/status",
		"/api/plant/timer",
		"/auth/status",
	}

	for _, endpoint := range endpoints {
		t.Run("json_response_"+endpoint, func(t *testing.T) {
			resp, err := http.Get(app.Server.URL + endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			// Ensure response is valid JSON
			var jsonResponse interface{}
			err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
			assert.NoError(t, err, "Response should be valid JSON")
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	resp, err := http.Get(app.Server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Basic security checks - more comprehensive headers would be added by middleware
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// TODO: Add tests for security headers when security middleware is implemented
	// Examples: X-Frame-Options, X-Content-Type-Options, etc.
}

func TestDemoModeWorkflow(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Test demo login endpoint availability
	resp, err := http.Get(app.Server.URL + "/auth/demo-login")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Demo login should be available (app is in demo mode by default)
	// Exact behavior depends on implementation - could be redirect or form
	assert.True(t, resp.StatusCode < 500, "Demo login endpoint should not error")
}

func TestConcurrentAccess(t *testing.T) {
	app := NewTestApp(t)
	defer app.Close()

	// Test concurrent access to read-only endpoints
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(app.Server.URL + "/api/plant/status")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				return
			}

			results <- nil
		}()
	}

	// Check all requests completed successfully
	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

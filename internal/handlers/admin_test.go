package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"watered/internal/models"
	"watered/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminHandler_GetConfigHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupStorage   func(*storage.MemoryStorage)
		expectedStatus int
		expectedConfig *models.AdminConfig
	}{
		{
			name: "should return existing config",
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  48,
					AllowedEmails: []string{"test@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus: http.StatusOK,
			expectedConfig: &models.AdminConfig{
				TimeoutHours:  48,
				AllowedEmails: []string{"test@example.com"},
				AdminEmails:   []string{"admin@example.com"},
			},
		},
		{
			name:           "should create default config when none exists (demo mode)",
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusOK,
			expectedConfig: &models.AdminConfig{
				TimeoutHours:  24,
				AllowedEmails: []string{"demo@example.com", "user1@example.com", "user2@example.com", "test@example.com", "admin@example.com"},
				AdminEmails:   []string{"admin@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := storage.NewMemoryStorage()
			tt.setupStorage(store)
			handler := NewAdminHandler(store)

			// Create request
			req := httptest.NewRequest("GET", "/admin/config", nil)
			rr := httptest.NewRecorder()

			// Execute
			handler.GetConfigHandler(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var config models.AdminConfig
				err := json.Unmarshal(rr.Body.Bytes(), &config)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedConfig.TimeoutHours, config.TimeoutHours)
				assert.Equal(t, tt.expectedConfig.AllowedEmails, config.AllowedEmails)
				assert.Equal(t, tt.expectedConfig.AdminEmails, config.AdminEmails)
			}
		})
	}
}

func TestAdminHandler_UpdateTimeoutHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupStorage   func(*storage.MemoryStorage)
		expectedStatus int
		expectedHours  int
	}{
		{
			name:           "should update timeout successfully",
			requestBody:    map[string]interface{}{"timeoutHours": 72},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusOK,
			expectedHours:  72,
		},
		{
			name:           "should reject timeout below minimum",
			requestBody:    map[string]interface{}{"timeoutHours": 0},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should reject timeout above maximum",
			requestBody:    map[string]interface{}{"timeoutHours": 200},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "should update existing config",
			requestBody: map[string]interface{}{"timeoutHours": 36},
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  24,
					AllowedEmails: []string{"existing@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus: http.StatusOK,
			expectedHours:  36,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := storage.NewMemoryStorage()
			tt.setupStorage(store)
			handler := NewAdminHandler(store)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/admin/config/timeout", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Execute
			handler.UpdateTimeoutHandler(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Equal(t, float64(tt.expectedHours), response["timeoutHours"].(float64))

				// Verify config was actually updated
				config, err := store.GetAdminConfig()
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHours, config.TimeoutHours)
			}
		})
	}
}

func TestAdminHandler_AddUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupStorage   func(*storage.MemoryStorage)
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "should add new user successfully",
			requestBody:    map[string]interface{}{"email": "newuser@example.com"},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusCreated,
			shouldContain:  "newuser@example.com",
		},
		{
			name:        "should reject duplicate email",
			requestBody: map[string]interface{}{"email": "existing@example.com"},
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  24,
					AllowedEmails: []string{"existing@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "should reject invalid email format",
			requestBody:    map[string]interface{}{"email": "invalid-email"},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should reject empty email",
			requestBody:    map[string]interface{}{"email": ""},
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := storage.NewMemoryStorage()
			tt.setupStorage(store)
			handler := NewAdminHandler(store)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Execute
			handler.AddUserHandler(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				// Verify user was added to config
				config, err := store.GetAdminConfig()
				require.NoError(t, err)
				assert.Contains(t, config.AllowedEmails, tt.shouldContain)
			}
		})
	}
}

func TestAdminHandler_RemoveUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		emailParam     string
		setupStorage   func(*storage.MemoryStorage)
		expectedStatus int
	}{
		{
			name:       "should remove existing user successfully",
			emailParam: "remove@example.com",
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  24,
					AllowedEmails: []string{"keep@example.com", "remove@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "should return 404 for non-existent user",
			emailParam: "notfound@example.com",
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  24,
					AllowedEmails: []string{"existing@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "should return 404 when no config exists",
			emailParam:     "any@example.com",
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := storage.NewMemoryStorage()
			tt.setupStorage(store)
			handler := NewAdminHandler(store)

			// Create request with URL parameter
			req := httptest.NewRequest("DELETE", "/admin/users/"+tt.emailParam, nil)
			rr := httptest.NewRecorder()

			// Setup chi context for URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("email", tt.emailParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute
			handler.RemoveUserHandler(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				// Verify user was removed from config
				config, err := store.GetAdminConfig()
				require.NoError(t, err)
				assert.NotContains(t, config.AllowedEmails, tt.emailParam)
			}
		})
	}
}

func TestAdminHandler_GetUsersHandler(t *testing.T) {
	tests := []struct {
		name            string
		setupStorage    func(*storage.MemoryStorage)
		expectedStatus  int
		expectedAllowed []string
		expectedAdmins  []string
	}{
		{
			name: "should return configured users",
			setupStorage: func(s *storage.MemoryStorage) {
				config := &models.AdminConfig{
					TimeoutHours:  24,
					AllowedEmails: []string{"user1@example.com", "user2@example.com"},
					AdminEmails:   []string{"admin@example.com"},
				}
				s.UpdateAdminConfig(config)
			},
			expectedStatus:  http.StatusOK,
			expectedAllowed: []string{"user1@example.com", "user2@example.com"},
			expectedAdmins:  []string{"admin@example.com"},
		},
		{
			name:            "should return default config when none exists",
			setupStorage:    func(s *storage.MemoryStorage) {},
			expectedStatus:  http.StatusOK,
			expectedAllowed: []string{"admin@example.com"},
			expectedAdmins:  []string{"admin@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			store := storage.NewMemoryStorage()
			tt.setupStorage(store)
			handler := NewAdminHandler(store)

			// Create request
			req := httptest.NewRequest("GET", "/admin/users", nil)
			rr := httptest.NewRecorder()

			// Execute
			handler.GetUsersHandler(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				allowedEmails := response["allowedEmails"].([]interface{})
				adminEmails := response["adminEmails"].([]interface{})

				assert.Len(t, allowedEmails, len(tt.expectedAllowed))
				assert.Len(t, adminEmails, len(tt.expectedAdmins))
			}
		})
	}
}

func TestAdminHandler_GetConfigWithEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name             string
		allowedEmailsEnv string
		adminEmailsEnv   string
		expectedAllowed  []string
		expectedAdmins   []string
	}{
		{
			name:             "should use ADMIN_EMAILS env var only",
			allowedEmailsEnv: "",
			adminEmailsEnv:   "admin@company.com,manager@company.com",
			expectedAllowed:  []string{"admin@company.com", "manager@company.com"},
			expectedAdmins:   []string{"admin@company.com", "manager@company.com"},
		},
		{
			name:             "should use ALLOWED_EMAILS env var only",
			allowedEmailsEnv: "user1@company.com,user2@company.com",
			adminEmailsEnv:   "",
			expectedAllowed:  []string{"user1@company.com", "user2@company.com"},
			expectedAdmins:   []string{},
		},
		{
			name:             "should combine both env vars",
			allowedEmailsEnv: "user1@company.com,user2@company.com",
			adminEmailsEnv:   "admin@company.com",
			expectedAllowed:  []string{"user1@company.com", "user2@company.com", "admin@company.com"},
			expectedAdmins:   []string{"admin@company.com"},
		},
		{
			name:             "should handle empty env vars (demo mode)",
			allowedEmailsEnv: "",
			adminEmailsEnv:   "",
			expectedAllowed:  []string{"demo@example.com", "user1@example.com", "user2@example.com", "test@example.com", "admin@example.com"},
			expectedAdmins:   []string{"admin@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.allowedEmailsEnv != "" {
				os.Setenv("ALLOWED_EMAILS", tt.allowedEmailsEnv)
			} else {
				os.Unsetenv("ALLOWED_EMAILS")
			}
			if tt.adminEmailsEnv != "" {
				os.Setenv("ADMIN_EMAILS", tt.adminEmailsEnv)
			} else {
				os.Unsetenv("ADMIN_EMAILS")
			}

			// Ensure cleanup
			defer func() {
				os.Unsetenv("ALLOWED_EMAILS")
				os.Unsetenv("ADMIN_EMAILS")
			}()

			// Setup
			store := storage.NewMemoryStorage()
			handler := NewAdminHandler(store)

			// Create request
			req := httptest.NewRequest("GET", "/admin/config", nil)
			rr := httptest.NewRecorder()

			// Execute
			handler.GetConfigHandler(rr, req)

			// Assert
			assert.Equal(t, http.StatusOK, rr.Code)

			var config models.AdminConfig
			err := json.Unmarshal(rr.Body.Bytes(), &config)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAllowed, config.AllowedEmails)
			assert.Equal(t, tt.expectedAdmins, config.AdminEmails)
		})
	}
}

func TestAdminHandler_TimeoutSynchronization(t *testing.T) {
	t.Run("admin config should reflect plant timeout", func(t *testing.T) {
		// Setup
		store := storage.NewMemoryStorage()

		// Create plant with custom timeout
		plant := &models.PlantState{
			ID:           1,
			Name:         "Test Plant",
			TimeoutHours: 48, // Different from default 24
		}
		store.UpdatePlantState(plant)

		handler := NewAdminHandler(store)

		// Create request
		req := httptest.NewRequest("GET", "/admin/config", nil)
		rr := httptest.NewRecorder()

		// Execute
		handler.GetConfigHandler(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var config models.AdminConfig
		err := json.Unmarshal(rr.Body.Bytes(), &config)
		require.NoError(t, err)

		// Admin config timeout should match plant timeout
		assert.Equal(t, 48, config.TimeoutHours)
	})

	t.Run("updating admin timeout should update plant timeout", func(t *testing.T) {
		// Setup
		store := storage.NewMemoryStorage()

		// Create plant with initial timeout
		plant := &models.PlantState{
			ID:           1,
			Name:         "Test Plant",
			TimeoutHours: 24,
		}
		store.UpdatePlantState(plant)

		handler := NewAdminHandler(store)

		// Update timeout via admin endpoint
		requestBody := `{"timeoutHours": 72}`
		req := httptest.NewRequest("PUT", "/admin/config/timeout", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateTimeoutHandler(rr, req)

		// Assert update was successful
		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify plant timeout was updated
		updatedPlant, err := store.GetPlantState()
		require.NoError(t, err)
		assert.Equal(t, 72, updatedPlant.TimeoutHours)

		// Verify admin config now returns the updated timeout
		req2 := httptest.NewRequest("GET", "/admin/config", nil)
		rr2 := httptest.NewRecorder()
		handler.GetConfigHandler(rr2, req2)

		assert.Equal(t, http.StatusOK, rr2.Code)
		var config models.AdminConfig
		err = json.Unmarshal(rr2.Body.Bytes(), &config)
		require.NoError(t, err)
		assert.Equal(t, 72, config.TimeoutHours)
	})
}

func TestAdminHandler_GetStatsHandler(t *testing.T) {
	store := storage.NewMemoryStorage()

	// Setup test data
	config := &models.AdminConfig{
		TimeoutHours:  48,
		AllowedEmails: []string{"user1@example.com", "user2@example.com"},
		AdminEmails:   []string{"admin@example.com"},
	}
	store.UpdateAdminConfig(config)

	handler := NewAdminHandler(store)

	// Create request
	req := httptest.NewRequest("GET", "/admin/stats", nil)
	rr := httptest.NewRecorder()

	// Execute
	handler.GetStatsHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["totalUsers"].(float64))
	assert.Equal(t, float64(1), response["adminUsers"].(float64))
	assert.Equal(t, float64(48), response["timeoutHours"].(float64))
	assert.Equal(t, "healthy", response["systemStatus"].(string))
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			name:           "should create default config when none exists",
			setupStorage:   func(s *storage.MemoryStorage) {},
			expectedStatus: http.StatusOK,
			expectedConfig: &models.AdminConfig{
				TimeoutHours:  24,
				AllowedEmails: []string{"admin@example.com"},
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

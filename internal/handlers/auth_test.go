package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"watered/internal/auth"
	"watered/internal/storage"
)

func TestAuthHandlers_LoginHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := auth.NewAuthService(store)
	authHandlers := NewAuthHandlers(authService)

	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()

	authHandlers.LoginHandler(w, req)

	// Should redirect to Google OAuth
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected redirect location to be set")
	}

	// Check that location contains Google OAuth URL
	if !contains(location, "accounts.google.com") {
		t.Error("Expected redirect to Google OAuth endpoint")
	}
}

func TestAuthHandlers_StatusHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := auth.NewAuthService(store)
	authHandlers := NewAuthHandlers(authService)

	// Test unauthenticated status
	req := httptest.NewRequest("GET", "/auth/status", nil)
	w := httptest.NewRecorder()

	authHandlers.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	authenticated, ok := response["authenticated"].(bool)
	if !ok || authenticated {
		t.Error("Expected authenticated to be false for unauthenticated user")
	}

	if response["user"] != nil {
		t.Error("Expected user to be nil for unauthenticated user")
	}
}

func TestAuthHandlers_StatusHandler_Authenticated(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := auth.NewAuthService(store)
	authHandlers := NewAuthHandlers(authService)

	// Create authenticated session
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	userInfo := &auth.GoogleUserInfo{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService.SetAllowedEmails(map[string]bool{"test@example.com": true})
	err := authService.CreateSession(w, req, userInfo)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create new request with session cookie
	cookies := w.Result().Cookies()
	req2 := httptest.NewRequest("GET", "/auth/status", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	// Test authenticated status
	w2 := httptest.NewRecorder()
	authHandlers.StatusHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	authenticated, ok := response["authenticated"].(bool)
	if !ok || !authenticated {
		t.Error("Expected authenticated to be true for authenticated user")
	}

	user, ok := response["user"].(map[string]interface{})
	if !ok || user == nil {
		t.Error("Expected user data for authenticated user")
	}

	if user["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", user["email"])
	}
}

func TestAuthHandlers_LogoutHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := auth.NewAuthService(store)
	authHandlers := NewAuthHandlers(authService)

	// Create authenticated session first
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	userInfo := &auth.GoogleUserInfo{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService.SetAllowedEmails(map[string]bool{"test@example.com": true})
	err := authService.CreateSession(w, req, userInfo)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create logout request with session cookie
	cookies := w.Result().Cookies()
	req2 := httptest.NewRequest("POST", "/auth/logout", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	authHandlers.LogoutHandler(w2, req2)

	// Should redirect to login
	if w2.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w2.Code)
	}

	location := w2.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to '/login', got '%s'", location)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (substr == "" || 
		   func() bool {
			   for i := 0; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }())
}
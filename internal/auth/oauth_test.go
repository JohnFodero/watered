package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"watered/internal/storage"
)

func TestNewAuthService(t *testing.T) {
	// Save original env vars
	originalClientID := os.Getenv("GOOGLE_CLIENT_ID")
	originalClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	originalSessionSecret := os.Getenv("SESSION_SECRET")
	originalAllowedEmails := os.Getenv("ALLOWED_EMAILS")
	originalAdminEmails := os.Getenv("ADMIN_EMAILS")

	// Clean up after test
	defer func() {
		os.Setenv("GOOGLE_CLIENT_ID", originalClientID)
		os.Setenv("GOOGLE_CLIENT_SECRET", originalClientSecret)
		os.Setenv("SESSION_SECRET", originalSessionSecret)
		os.Setenv("ALLOWED_EMAILS", originalAllowedEmails)
		os.Setenv("ADMIN_EMAILS", originalAdminEmails)
	}()

	// Test with environment variables
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("SESSION_SECRET", "test-session-secret")
	os.Setenv("ALLOWED_EMAILS", "user1@example.com,user2@example.com")
	os.Setenv("ADMIN_EMAILS", "admin@example.com")

	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := NewAuthService(store)

	if authService == nil {
		t.Fatal("Expected auth service to be created")
	}

	if authService.oauth2Config == nil {
		t.Fatal("Expected OAuth2 config to be created")
	}

	if authService.oauth2Config.ClientID != "test-client-id" {
		t.Errorf("Expected client ID 'test-client-id', got '%s'", authService.oauth2Config.ClientID)
	}

	// Test email whitelist
	if !authService.IsUserAllowed("user1@example.com") {
		t.Error("Expected user1@example.com to be allowed")
	}

	if !authService.IsUserAllowed("admin@example.com") {
		t.Error("Expected admin@example.com to be allowed")
	}

	if authService.IsUserAllowed("unknown@example.com") {
		t.Error("Expected unknown@example.com to be denied")
	}

	// Test admin check
	if !authService.IsUserAdmin("admin@example.com") {
		t.Error("Expected admin@example.com to be admin")
	}

	if authService.IsUserAdmin("user1@example.com") {
		t.Error("Expected user1@example.com to not be admin")
	}
}

func TestGenerateStateToken(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	token1, err := authService.GenerateStateToken()
	if err != nil {
		t.Fatalf("Failed to generate state token: %v", err)
	}

	token2, err := authService.GenerateStateToken()
	if err != nil {
		t.Fatalf("Failed to generate second state token: %v", err)
	}

	if token1 == token2 {
		t.Error("Expected different state tokens")
	}

	if len(token1) == 0 {
		t.Error("Expected non-empty state token")
	}
}

func TestGetLoginURL(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	state := "test-state"
	url := authService.GetLoginURL(state)

	if url == "" {
		t.Error("Expected non-empty login URL")
	}

	// Check that URL contains expected components
	if !contains(url, "accounts.google.com") {
		t.Error("Expected login URL to contain Google OAuth endpoint")
	}

	if !contains(url, state) {
		t.Error("Expected login URL to contain state parameter")
	}
}

func TestCreateAndGetSession(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	// Create test request and response
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test user info
	userInfo := &GoogleUserInfo{
		ID:            "123",
		Email:         "test@example.com",
		VerifiedEmail: true,
		Name:          "Test User",
		Picture:       "https://example.com/picture.jpg",
	}

	// Set up allowed email
	authService.allowedEmails["test@example.com"] = true

	// Create session
	err := authService.CreateSession(w, req, userInfo)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting current user (need to create new request with session cookie)
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected session cookie to be set")
	}

	// Create new request with session cookie
	req2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	// Get current user
	user, err := authService.GetCurrentUser(req2)
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}

	// Test authentication check
	if !authService.IsAuthenticated(req2) {
		t.Error("Expected user to be authenticated")
	}
}

func TestClearSession(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	// Create and authenticate user
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	userInfo := &GoogleUserInfo{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService.allowedEmails["test@example.com"] = true
	err := authService.CreateSession(w, req, userInfo)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create new request with session cookie
	cookies := w.Result().Cookies()
	req2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	// Verify user is authenticated
	if !authService.IsAuthenticated(req2) {
		t.Fatal("Expected user to be authenticated before logout")
	}

	// Clear session
	w2 := httptest.NewRecorder()
	err = authService.ClearSession(w2, req2)
	if err != nil {
		t.Fatalf("Failed to clear session: %v", err)
	}

	// Create new request with cleared session cookie
	clearedCookies := w2.Result().Cookies()
	req3 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range clearedCookies {
		req3.AddCookie(cookie)
	}

	// Verify user is no longer authenticated
	if authService.IsAuthenticated(req3) {
		t.Error("Expected user to not be authenticated after logout")
	}
}

func TestAuthRequiredMiddleware(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	// Test handler that should only be called if authenticated
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := authService.AuthRequired(testHandler)

	// Test unauthenticated request
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect status %d, got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to '/login', got '%s'", location)
	}
}

func TestAdminRequiredMiddleware(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	// Test handler that should only be called if admin
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin success"))
	})

	middleware := authService.AdminRequired(testHandler)

	// Test unauthenticated request
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected forbidden status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCreateDemoSession(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	// Test demo mode detection
	if !authService.IsDemoMode() {
		t.Error("Expected service to be in demo mode when no credentials are set")
	}

	// Create test request and response
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test demo session creation
	err := authService.CreateDemoSession(w, req, "demo@example.com", "Demo User", false)
	if err != nil {
		t.Fatalf("Failed to create demo session: %v", err)
	}

	// Test getting current user from demo session
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected session cookie to be set")
	}

	// Create new request with session cookie
	req2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	user, err := authService.GetCurrentUser(req2)
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Email != "demo@example.com" {
		t.Errorf("Expected email 'demo@example.com', got '%s'", user.Email)
	}

	if user.Name != "Demo User" {
		t.Errorf("Expected name 'Demo User', got '%s'", user.Name)
	}

	// Test authentication check
	if !authService.IsAuthenticated(req2) {
		t.Error("Expected user to be authenticated")
	}
}

func TestDemoSessionWithUnauthorizedEmail(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	authService := NewAuthService(store)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Try to create demo session with unauthorized email
	err := authService.CreateDemoSession(w, req, "unauthorized@example.com", "Unauthorized User", false)
	if err == nil {
		t.Error("Expected error for unauthorized email")
	}

	if err.Error() != "user not in allowlist" {
		t.Errorf("Expected 'user not in allowlist' error, got '%s'", err.Error())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (substr == "" || s[len(s)-len(substr):] == substr || 
		   len(s) >= len(substr) && (s[:len(substr)] == substr || 
		   func() bool {
			   for i := 0; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }()))
}
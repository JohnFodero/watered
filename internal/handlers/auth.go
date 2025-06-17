package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"watered/internal/auth"
)

// AuthHandlers contains all authentication-related HTTP handlers
type AuthHandlers struct {
	authService *auth.AuthService
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService *auth.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// LoginHandler redirects users to Google OAuth2
func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate state token for CSRF protection
	state, err := h.authService.GenerateStateToken()
	if err != nil {
		log.Printf("Failed to generate state token: %v", err)
		http.Error(w, "Failed to initiate login", http.StatusInternalServerError)
		return
	}

	// Store state in session for validation
	session, err := h.authService.GetSession(r)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Redirect to Google OAuth2
	url := h.authService.GetLoginURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CallbackHandler handles OAuth2 callback from Google
func (h *AuthHandlers) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Validate state parameter
	state := r.FormValue("state")
	session, err := h.authService.GetSession(r)
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || state != expectedState {
		log.Printf("Invalid state parameter: expected %s, got %s", expectedState, state)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token and get user info
	userInfo, err := h.authService.HandleCallback(r.Context(), code)
	if err != nil {
		log.Printf("OAuth callback failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Check if user is allowed
	if !h.authService.IsUserAllowed(userInfo.Email) {
		log.Printf("User %s not in allowlist", userInfo.Email)
		http.Error(w, "Access denied: User not authorized", http.StatusForbidden)
		return
	}

	// Create session for user
	if err := h.authService.CreateSession(w, r, userInfo); err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s (%s) logged in successfully", userInfo.Name, userInfo.Email)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LogoutHandler clears the user session
func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user for logging
	user, _ := h.authService.GetCurrentUser(r)
	
	// Clear session
	if err := h.authService.ClearSession(w, r); err != nil {
		log.Printf("Failed to clear session: %v", err)
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	if user != nil {
		log.Printf("User %s logged out", user.Email)
	}

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// StatusHandler returns the current authentication status
func (h *AuthHandlers) StatusHandler(w http.ResponseWriter, r *http.Request) {
	type UserResponse struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		IsAdmin bool   `json:"is_admin"`
	}

	type AuthStatus struct {
		Authenticated bool          `json:"authenticated"`
		User          *UserResponse `json:"user,omitempty"`
	}

	user, err := h.authService.GetCurrentUser(r)
	status := AuthStatus{
		Authenticated: err == nil && user != nil,
	}

	if status.Authenticated {
		status.User = &UserResponse{
			Email:   user.Email,
			Name:    user.Name,
			IsAdmin: user.IsAdmin,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
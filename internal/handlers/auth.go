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
		log.Printf("LoginHandler: Failed to get session - %v", err)
		log.Printf("LoginHandler: User-Agent: %s", r.Header.Get("User-Agent"))
		log.Printf("LoginHandler: Request from: %s", r.RemoteAddr)
		http.Error(w, "Session initialization failed. Please clear your browser cookies and try again.", http.StatusInternalServerError)
		return
	}

	session.Values["oauth_state"] = state
	if err := session.Save(r, w); err != nil {
		log.Printf("LoginHandler: Failed to save session state - %v", err)
		http.Error(w, "Session storage failed. Please clear your browser cookies and try again.", http.StatusInternalServerError)
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
		log.Printf("CallbackHandler: Failed to get session - %v", err)
		log.Printf("CallbackHandler: User-Agent: %s", r.Header.Get("User-Agent"))
		log.Printf("CallbackHandler: Request from: %s", r.RemoteAddr)
		log.Printf("CallbackHandler: State parameter: %s", state)
		http.Error(w, "Session validation failed. Please clear your browser cookies and try logging in again.", http.StatusInternalServerError)
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

// DemoLoginHandler provides demo authentication for testing (only in demo mode)
func (h *AuthHandlers) DemoLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !h.authService.IsDemoMode() {
		http.Error(w, "Demo login only available in demo mode", http.StatusNotFound)
		return
	}

	// Handle GET for simple demo login (auto-login test user)
	if r.Method == "GET" {
		// Check if user is requesting a simple login
		if r.URL.Query().Get("simple") == "true" {
			// Create demo session with default test user
			if err := h.authService.CreateDemoSession(w, r, "test@example.com", "Demo User", false); err != nil {
				log.Printf("Failed to create demo session: %v", err)
				http.Error(w, "Failed to create demo session: "+err.Error(), http.StatusBadRequest)
				return
			}

			log.Printf("Demo user %s (%s) logged in successfully", "Demo User", "test@example.com")

			// Return JSON response for API users
			if r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"success": true, "message": "Demo login successful", "user": {"email": "test@example.com", "name": "Demo User"}}`))
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Handle POST for demo login
	if r.Method == "POST" {
		email := r.FormValue("email")
		name := r.FormValue("name")
		isAdmin := r.FormValue("admin") == "true"

		if email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		if name == "" {
			name = "Demo User"
		}

		// Create demo session
		if err := h.authService.CreateDemoSession(w, r, email, name, isAdmin); err != nil {
			log.Printf("Failed to create demo session: %v", err)
			http.Error(w, "Failed to create demo session: "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Demo user %s (%s) logged in successfully", name, email)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Show demo login form
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Demo Login - Watered</title>
    <link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header class="header">
        <div class="header-content">
            <a href="/" class="logo">🌱 Watered</a>
            <nav>
                <ul class="nav-links">
                    <li><a href="/">Home</a></li>
                    <li><a href="/login">Login</a></li>
                </ul>
            </nav>
        </div>
    </header>

    <div class="container">
        <main class="login-container">
            <h1 class="login-title">🧪 Demo Login</h1>
            <p style="text-align: center; margin-bottom: 2rem; color: var(--muted-text);">
                Test authentication without Google OAuth
            </p>

            <form method="post" style="margin-bottom: 2rem;">
                <div class="form-group">
                    <label for="email">Email:</label>
                    <select id="email" name="email" required>
                        <option value="">Select a demo user...</option>
                        <option value="demo@example.com">demo@example.com (Regular User)</option>
                        <option value="user1@example.com">user1@example.com (Regular User)</option>
                        <option value="user2@example.com">user2@example.com (Regular User)</option>
                        <option value="admin@example.com">admin@example.com (Admin)</option>
                    </select>
                </div>
                
                <div class="form-group">
                    <label for="name">Display Name:</label>
                    <input type="text" id="name" name="name" placeholder="Demo User" />
                </div>

                <div class="form-group">
                    <label>
                        <input type="checkbox" name="admin" value="true" /> 
                        Login as Admin (only works for admin@example.com)
                    </label>
                </div>

                <button type="submit" class="btn" style="width: 100%;">🚀 Demo Login</button>
            </form>

            <div style="background-color: var(--secondary-bg); padding: 1rem; border-radius: var(--border-radius); margin-top: 1rem;">
                <h4 style="margin: 0 0 0.5rem 0; color: var(--accent-color);">Demo Mode Instructions:</h4>
                <ul style="margin: 0; padding-left: 1.5rem; font-size: 0.9rem; color: var(--muted-text);">
                    <li>Choose any of the pre-configured demo users</li>
                    <li>Only admin@example.com can access admin features</li>
                    <li>Sessions work exactly like real Google OAuth</li>
                    <li>You can logout and test different users</li>
                </ul>
            </div>

            <div style="text-align: center; margin-top: 1rem;">
                <a href="/login" class="btn btn-secondary">← Back to Real Login</a>
            </div>
        </main>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
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

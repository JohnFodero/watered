package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"watered/internal/models"
	"watered/internal/storage"
)

// GoogleUserInfo represents user info from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// AuthService handles authentication operations
type AuthService struct {
	oauth2Config  *oauth2.Config
	store         *sessions.CookieStore
	storage       storage.Storage
	allowedEmails map[string]bool
	adminEmails   map[string]bool
}

// NewAuthService creates a new authentication service
func NewAuthService(storage storage.Storage) *AuthService {
	// Get OAuth2 credentials from environment
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	sessionSecret := os.Getenv("SESSION_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Printf("Warning: Google OAuth2 credentials not set. Demo mode enabled.")
		log.Printf("Visit /auth/demo-login to test authentication without Google OAuth.")
		clientID = "demo-client-id"
		clientSecret = "demo-client-secret"
	}

	if sessionSecret == "" {
		sessionSecret = "development-secret-change-in-production"
		log.Printf("Warning: SESSION_SECRET not set. Using development secret.")
	} else {
		log.Printf("SESSION_SECRET loaded successfully (length: %d characters)", len(sessionSecret))
	}

	// Determine redirect URL based on environment
	redirectURL := os.Getenv("REDIRECT_URL")
	if redirectURL == "" {
		// Default to localhost for development
		redirectURL = "http://localhost:8080/auth/callback"
	}

	log.Printf("OAuth redirect URL: %s", redirectURL)

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Determine if we should use secure cookies (HTTPS)
	secureCookies := os.Getenv("SECURE_COOKIES") == "true"
	environment := os.Getenv("ENVIRONMENT")

	// Auto-detect secure cookies for production environments
	if environment == "production" || environment == "prod" {
		secureCookies = true
	}

	log.Printf("Cookie configuration: secure=%v, environment=%s", secureCookies, environment)

	// Create secure cookie store
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
	}

	// Parse allowed emails
	allowedEmails := make(map[string]bool)
	adminEmails := make(map[string]bool)

	if allowedEmailsStr := os.Getenv("ALLOWED_EMAILS"); allowedEmailsStr != "" {
		for _, email := range strings.Split(allowedEmailsStr, ",") {
			allowedEmails[strings.TrimSpace(email)] = true
		}
	} else {
		// Demo allowed emails
		allowedEmails["demo@example.com"] = true
		allowedEmails["user1@example.com"] = true
		allowedEmails["user2@example.com"] = true
		allowedEmails["test@example.com"] = true
	}

	if adminEmailsStr := os.Getenv("ADMIN_EMAILS"); adminEmailsStr != "" {
		for _, email := range strings.Split(adminEmailsStr, ",") {
			email = strings.TrimSpace(email)
			adminEmails[email] = true
			allowedEmails[email] = true // Admins are also allowed users
		}
	} else {
		// Demo admin email
		adminEmails["admin@example.com"] = true
		allowedEmails["admin@example.com"] = true
	}

	return &AuthService{
		oauth2Config:  oauth2Config,
		store:         store,
		storage:       storage,
		allowedEmails: allowedEmails,
		adminEmails:   adminEmails,
	}
}

// GenerateStateToken creates a random state token for OAuth2 CSRF protection
func (a *AuthService) GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetLoginURL returns the Google OAuth2 login URL
func (a *AuthService) GetLoginURL(state string) string {
	return a.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleCallback processes the OAuth2 callback
func (a *AuthService) HandleCallback(ctx context.Context, code string) (*GoogleUserInfo, error) {
	token, err := a.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from Google
	client := a.oauth2Config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// IsUserAllowed checks if a user email is in the whitelist
func (a *AuthService) IsUserAllowed(email string) bool {
	// First check static configuration (for fallback/demo mode)
	if a.allowedEmails[email] {
		return true
	}

	// Then check dynamic admin configuration from storage
	config, err := a.storage.GetAdminConfig()
	if err != nil || config == nil {
		log.Printf("Warning: Failed to get admin config, using static allowlist: %v", err)
		return a.allowedEmails[email]
	}

	// Check if email is in the dynamic allowlist
	for _, allowedEmail := range config.AllowedEmails {
		if allowedEmail == email {
			return true
		}
	}

	return false
}

// IsUserAdmin checks if a user email is in the admin list
func (a *AuthService) IsUserAdmin(email string) bool {
	// First check static configuration (for fallback/demo mode)
	if a.adminEmails[email] {
		return true
	}

	// Then check dynamic admin configuration from storage
	config, err := a.storage.GetAdminConfig()
	if err != nil || config == nil {
		log.Printf("Warning: Failed to get admin config, using static admin list: %v", err)
		return a.adminEmails[email]
	}

	// Check if email is in the dynamic admin list
	for _, adminEmail := range config.AdminEmails {
		if adminEmail == email {
			return true
		}
	}

	return false
}

// CreateSession creates a new user session
func (a *AuthService) CreateSession(w http.ResponseWriter, r *http.Request, userInfo *GoogleUserInfo) error {
	session, err := a.store.Get(r, "watered-session")
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Store user info in session
	session.Values["user_id"] = userInfo.ID
	session.Values["user_email"] = userInfo.Email
	session.Values["user_name"] = userInfo.Name
	session.Values["user_picture"] = userInfo.Picture
	session.Values["is_admin"] = a.IsUserAdmin(userInfo.Email)
	session.Values["authenticated"] = true
	session.Values["login_time"] = time.Now().Unix()

	// Save session
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Create or update user in storage
	user := &models.User{
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		IsAdmin:  a.IsUserAdmin(userInfo.Email),
		JoinedAt: time.Now(),
	}

	if existingUser, err := a.storage.GetUser(userInfo.Email); err == nil && existingUser != nil {
		// Update existing user
		user.JoinedAt = existingUser.JoinedAt
	}

	if err := a.storage.CreateUser(user); err != nil {
		log.Printf("Warning: Failed to store user in database: %v", err)
	}

	return nil
}

// GetCurrentUser returns the current authenticated user
func (a *AuthService) GetCurrentUser(r *http.Request) (*models.User, error) {
	session, err := a.store.Get(r, "watered-session")
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
		return nil, nil
	}

	email, ok := session.Values["user_email"].(string)
	if !ok {
		return nil, fmt.Errorf("no email in session")
	}

	name, _ := session.Values["user_name"].(string)
	isAdmin, _ := session.Values["is_admin"].(bool)

	return &models.User{
		Email:   email,
		Name:    name,
		IsAdmin: isAdmin,
	}, nil
}

// IsAuthenticated checks if the current request is authenticated
func (a *AuthService) IsAuthenticated(r *http.Request) bool {
	user, err := a.GetCurrentUser(r)
	return err == nil && user != nil
}

// GetSession returns the current session
func (a *AuthService) GetSession(r *http.Request) (*sessions.Session, error) {
	session, err := a.store.Get(r, "watered-session")
	if err != nil {
		log.Printf("GetSession error: %v", err)
		log.Printf("GetSession: Cookie count: %d", len(r.Cookies()))
		for _, cookie := range r.Cookies() {
			if cookie.Name == "watered-session" {
				log.Printf("GetSession: Found session cookie (length: %d)", len(cookie.Value))
				break
			}
		}
	}
	return session, err
}

// ClearSession logs out the user by clearing their session
func (a *AuthService) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := a.store.Get(r, "watered-session")
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Clear session values
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1

	return session.Save(r, w)
}

// AuthRequired middleware that requires authentication
func (a *AuthService) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminRequired middleware that requires admin privileges
func (a *AuthService) AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.GetCurrentUser(r)
		if err != nil || user == nil || !user.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetAllowedEmails sets the allowed emails (for testing)
func (a *AuthService) SetAllowedEmails(emails map[string]bool) {
	a.allowedEmails = emails
}

// IsDemoMode checks if we're running in demo mode (no real Google credentials)
func (a *AuthService) IsDemoMode() bool {
	return a.oauth2Config.ClientID == "demo-client-id"
}

// CreateDemoSession creates a demo session for testing (bypasses Google OAuth)
func (a *AuthService) CreateDemoSession(w http.ResponseWriter, r *http.Request, email string, name string, isAdmin bool) error {
	if !a.IsDemoMode() {
		return fmt.Errorf("demo sessions only available in demo mode")
	}

	// Check if user is allowed
	if !a.IsUserAllowed(email) {
		return fmt.Errorf("user not in allowlist")
	}

	// Create demo user info
	userInfo := &GoogleUserInfo{
		ID:            "demo-" + email,
		Email:         email,
		VerifiedEmail: true,
		Name:          name,
		Picture:       "https://via.placeholder.com/150?text=" + name[0:1],
	}

	return a.CreateSession(w, r, userInfo)
}

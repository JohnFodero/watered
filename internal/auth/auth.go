package auth

import (
	"watered/internal/models"
)

// AuthService handles authentication operations
type AuthService struct {
	// Will be implemented in future tasks
}

// NewAuthService creates a new authentication service
func NewAuthService() *AuthService {
	return &AuthService{}
}

// ValidateUser validates a user's authentication (placeholder)
func (a *AuthService) ValidateUser(email string) (*models.User, error) {
	// Placeholder implementation
	// Will be replaced with Google OAuth2 in Task 3
	return nil, nil
}

// IsUserAllowed checks if a user email is in the whitelist (placeholder)
func (a *AuthService) IsUserAllowed(email string) bool {
	// Placeholder - will implement proper whitelist checking
	return false
}
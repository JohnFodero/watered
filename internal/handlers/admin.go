package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"watered/internal/models"
	"watered/internal/storage"

	"github.com/go-chi/chi/v5"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	storage storage.Storage
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(storage storage.Storage) *AdminHandler {
	return &AdminHandler{
		storage: storage,
	}
}

// getEmailsFromEnv parses comma-separated emails from environment variable
func getEmailsFromEnv(envVar string, fallback []string) []string {
	if envValue := os.Getenv(envVar); envValue != "" {
		var emails []string
		for _, email := range strings.Split(envValue, ",") {
			if trimmed := strings.TrimSpace(email); trimmed != "" {
				emails = append(emails, trimmed)
			}
		}
		return emails
	}
	return fallback
}

// GetConfigHandler returns the current admin configuration
func (h *AdminHandler) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	// If no config exists, create default
	if config == nil {
		// Get emails from environment variables, with empty fallback for production
		allowedEmails := getEmailsFromEnv("ALLOWED_EMAILS", []string{})
		adminEmails := getEmailsFromEnv("ADMIN_EMAILS", []string{})
		
		// In demo mode (no env vars set), provide demo defaults
		if len(allowedEmails) == 0 && len(adminEmails) == 0 {
			allowedEmails = []string{"demo@example.com", "user1@example.com", "user2@example.com", "test@example.com"}
			adminEmails = []string{"admin@example.com"}
		}
		
		// Ensure admin emails are also in allowed emails
		allowedEmailsMap := make(map[string]bool)
		for _, email := range allowedEmails {
			allowedEmailsMap[email] = true
		}
		for _, email := range adminEmails {
			if !allowedEmailsMap[email] {
				allowedEmails = append(allowedEmails, email)
			}
		}
		
		config = &models.AdminConfig{
			TimeoutHours:  24,
			AllowedEmails: allowedEmails,
			AdminEmails:   adminEmails,
		}
		if err := h.storage.UpdateAdminConfig(config); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create default config: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode config: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateTimeoutHandler updates the watering timeout configuration
func (h *AdminHandler) UpdateTimeoutHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TimeoutHours int `json:"timeoutHours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate timeout range
	if request.TimeoutHours < 1 || request.TimeoutHours > 168 {
		http.Error(w, "Timeout must be between 1 and 168 hours", http.StatusBadRequest)
		return
	}

	// Get current config
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	if config == nil {
		// Use environment variable logic for initial config
		allowedEmails := getEmailsFromEnv("ALLOWED_EMAILS", []string{})
		adminEmails := getEmailsFromEnv("ADMIN_EMAILS", []string{})
		
		// In demo mode (no env vars set), provide demo defaults
		if len(allowedEmails) == 0 && len(adminEmails) == 0 {
			allowedEmails = []string{"demo@example.com", "user1@example.com", "user2@example.com", "test@example.com"}
			adminEmails = []string{"admin@example.com"}
		}
		
		// Ensure admin emails are also in allowed emails
		allowedEmailsMap := make(map[string]bool)
		for _, email := range allowedEmails {
			allowedEmailsMap[email] = true
		}
		for _, email := range adminEmails {
			if !allowedEmailsMap[email] {
				allowedEmails = append(allowedEmails, email)
			}
		}
		
		config = &models.AdminConfig{
			TimeoutHours:  request.TimeoutHours,
			AllowedEmails: allowedEmails,
			AdminEmails:   adminEmails,
		}
	} else {
		config.TimeoutHours = request.TimeoutHours
	}

	// Update config
	if err := h.storage.UpdateAdminConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	// Also update the plant timeout to keep them synchronized
	plant, err := h.storage.GetPlantState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get plant state: %v", err), http.StatusInternalServerError)
		return
	}
	
	if plant != nil {
		plant.TimeoutHours = request.TimeoutHours
		if err := h.storage.UpdatePlantState(plant); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update plant timeout: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	response := map[string]interface{}{
		"success":      true,
		"timeoutHours": request.TimeoutHours,
		"message":      fmt.Sprintf("Timeout updated to %d hours", request.TimeoutHours),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUsersHandler returns the list of whitelisted users
func (h *AdminHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	if config == nil {
		config = &models.AdminConfig{
			TimeoutHours:  24,
			AllowedEmails: []string{"admin@example.com"},
			AdminEmails:   []string{"admin@example.com"},
		}
	}

	response := map[string]interface{}{
		"allowedEmails": config.AllowedEmails,
		"adminEmails":   config.AdminEmails,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddUserHandler adds a user to the whitelist
func (h *AdminHandler) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(request.Email))
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Get current config
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	if config == nil {
		config = &models.AdminConfig{
			TimeoutHours:  24,
			AllowedEmails: []string{},
			AdminEmails:   []string{"admin@example.com"},
		}
	}

	// Check if email already exists
	for _, existingEmail := range config.AllowedEmails {
		if existingEmail == email {
			http.Error(w, "Email already exists in whitelist", http.StatusConflict)
			return
		}
	}

	// Add email to allowed list
	config.AllowedEmails = append(config.AllowedEmails, email)

	// Update config
	if err := h.storage.UpdateAdminConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Added %s to allowed users", email),
		"email":   email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RemoveUserHandler removes a user from the whitelist
func (h *AdminHandler) RemoveUserHandler(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	email = strings.TrimSpace(strings.ToLower(email))

	// Get current config
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "No configuration found", http.StatusNotFound)
		return
	}

	// Check if email exists and remove it
	found := false
	newAllowedEmails := make([]string, 0, len(config.AllowedEmails))
	for _, existingEmail := range config.AllowedEmails {
		if existingEmail != email {
			newAllowedEmails = append(newAllowedEmails, existingEmail)
		} else {
			found = true
		}
	}

	if !found {
		http.Error(w, "Email not found in whitelist", http.StatusNotFound)
		return
	}

	config.AllowedEmails = newAllowedEmails

	// Update config
	if err := h.storage.UpdateAdminConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Removed %s from allowed users", email),
		"email":   email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHistoryHandler returns plant watering history
func (h *AdminHandler) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get current plant state
	plant, err := h.storage.GetPlantState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get plant state: %v", err), http.StatusInternalServerError)
		return
	}

	// For now, return current state as history
	// In a real implementation, this would return historical watering events
	history := map[string]interface{}{
		"currentState": plant,
		"events":       []interface{}{}, // TODO: Implement watering history storage
		"message":      "Plant history feature will be enhanced in future versions",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetStatsHandler returns usage statistics
func (h *AdminHandler) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	config, err := h.storage.GetAdminConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get admin config: %v", err), http.StatusInternalServerError)
		return
	}

	plant, err := h.storage.GetPlantState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get plant state: %v", err), http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"totalUsers":   len(config.AllowedEmails),
		"adminUsers":   len(config.AdminEmails),
		"timeoutHours": config.TimeoutHours,
		"plantWatered": plant != nil && plant.LastWatered != nil,
		"systemStatus": "healthy",
	}

	if plant != nil && plant.LastWatered != nil {
		stats["lastWatered"] = plant.LastWatered.Format("2006-01-02 15:04:05")
		stats["wateredBy"] = plant.WateredBy
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

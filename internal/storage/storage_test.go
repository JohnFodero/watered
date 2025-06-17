package storage

import (
	"testing"
	"time"

	"watered/internal/models"
)

func TestMemoryStorage_PlantOperations(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test getting plant state when none exists
	state, err := storage.GetPlantState()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if state != nil {
		t.Errorf("Expected nil state, got %v", state)
	}

	// Test updating plant state
	now := time.Now()
	plantState := &models.PlantState{
		ID:           1,
		Name:         "Test Plant",
		LastWatered:  now,
		TimeoutHours: 24,
		WateredBy:    "test@example.com",
	}

	err = storage.UpdatePlantState(plantState)
	if err != nil {
		t.Errorf("Expected no error updating plant state, got %v", err)
	}

	// Test getting updated plant state
	retrievedState, err := storage.GetPlantState()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedState == nil {
		t.Errorf("Expected plant state, got nil")
	}
	if retrievedState.Name != "Test Plant" {
		t.Errorf("Expected name 'Test Plant', got '%s'", retrievedState.Name)
	}
}

func TestMemoryStorage_UserOperations(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	email := "test@example.com"

	// Test getting user when none exists
	user, err := storage.GetUser(email)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}

	// Test creating user
	newUser := &models.User{
		Email:    email,
		Name:     "Test User",
		IsAdmin:  false,
		JoinedAt: time.Now(),
	}

	err = storage.CreateUser(newUser)
	if err != nil {
		t.Errorf("Expected no error creating user, got %v", err)
	}

	// Test getting created user
	retrievedUser, err := storage.GetUser(email)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedUser == nil {
		t.Errorf("Expected user, got nil")
	}
	if retrievedUser.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, retrievedUser.Email)
	}
	if retrievedUser.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", retrievedUser.Name)
	}
}

func TestMemoryStorage_AdminConfig(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test getting config when none exists
	config, err := storage.GetAdminConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config != nil {
		t.Errorf("Expected nil config, got %v", config)
	}

	// Test updating admin config
	adminConfig := &models.AdminConfig{
		TimeoutHours:  48,
		AllowedEmails: []string{"user1@example.com", "user2@example.com"},
		AdminEmails:   []string{"admin@example.com"},
		LastModified:  time.Now(),
		ModifiedBy:    "admin@example.com",
	}

	err = storage.UpdateAdminConfig(adminConfig)
	if err != nil {
		t.Errorf("Expected no error updating config, got %v", err)
	}

	// Test getting updated config
	retrievedConfig, err := storage.GetAdminConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedConfig == nil {
		t.Errorf("Expected config, got nil")
	}
	if retrievedConfig.TimeoutHours != 48 {
		t.Errorf("Expected timeout 48, got %d", retrievedConfig.TimeoutHours)
	}
	if len(retrievedConfig.AllowedEmails) != 2 {
		t.Errorf("Expected 2 allowed emails, got %d", len(retrievedConfig.AllowedEmails))
	}
}
package storage

import (
	"watered/internal/models"
)

// Storage defines the interface for data persistence
type Storage interface {
	// Plant operations
	GetPlantState() (*models.PlantState, error)
	UpdatePlantState(state *models.PlantState) error

	// User operations
	GetUser(email string) (*models.User, error)
	CreateUser(user *models.User) error

	// Admin operations
	GetAdminConfig() (*models.AdminConfig, error)
	UpdateAdminConfig(config *models.AdminConfig) error

	// Close the storage connection
	Close() error
}

// MemoryStorage provides in-memory storage for development
type MemoryStorage struct {
	plant  *models.PlantState
	users  map[string]*models.User
	config *models.AdminConfig
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users: make(map[string]*models.User),
	}
}

// GetPlantState returns the current plant state
func (m *MemoryStorage) GetPlantState() (*models.PlantState, error) {
	return m.plant, nil
}

// UpdatePlantState updates the plant state
func (m *MemoryStorage) UpdatePlantState(state *models.PlantState) error {
	m.plant = state
	return nil
}

// GetUser retrieves a user by email
func (m *MemoryStorage) GetUser(email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// CreateUser creates a new user
func (m *MemoryStorage) CreateUser(user *models.User) error {
	m.users[user.Email] = user
	return nil
}

// GetAdminConfig returns the admin configuration
func (m *MemoryStorage) GetAdminConfig() (*models.AdminConfig, error) {
	return m.config, nil
}

// UpdateAdminConfig updates the admin configuration
func (m *MemoryStorage) UpdateAdminConfig(config *models.AdminConfig) error {
	m.config = config
	return nil
}

// Close closes the storage connection (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	return nil
}

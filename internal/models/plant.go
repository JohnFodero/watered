package models

import "time"

// PlantState represents the current state of the plant
type PlantState struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LastWatered  time.Time `json:"last_watered"`
	TimeoutHours int       `json:"timeout_hours"`
	WateredBy    string    `json:"watered_by"`
}

// User represents a user in the system
type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsAdmin  bool   `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

// AdminConfig represents system configuration
type AdminConfig struct {
	TimeoutHours    int      `json:"timeout_hours"`
	AllowedEmails   []string `json:"allowed_emails"`
	AdminEmails     []string `json:"admin_emails"`
	LastModified    time.Time `json:"last_modified"`
	ModifiedBy      string   `json:"modified_by"`
}
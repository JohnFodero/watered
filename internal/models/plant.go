package models

import (
	"fmt"
	"time"
)

// PlantHealthStatus represents the health status of a plant
type PlantHealthStatus string

const (
	HealthStatusHealthy    PlantHealthStatus = "healthy"
	HealthStatusNeedsWater PlantHealthStatus = "needs_water"
	HealthStatusCritical   PlantHealthStatus = "critical"
	HealthStatusUnknown    PlantHealthStatus = "unknown"
)

// PlantState represents the current state of the plant
type PlantState struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	LastWatered  *time.Time `json:"last_watered"` // Pointer to handle null case
	TimeoutHours int        `json:"timeout_hours"`
	WateredBy    string     `json:"watered_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PlantWateringEvent represents a single watering event
type PlantWateringEvent struct {
	ID        int       `json:"id"`
	PlantID   int       `json:"plant_id"`
	WateredAt time.Time `json:"watered_at"`
	WateredBy string    `json:"watered_by"`
}

// GetHealthStatus calculates the current health status based on last watering time
func (p *PlantState) GetHealthStatus() PlantHealthStatus {
	if p.LastWatered == nil {
		return HealthStatusCritical
	}

	hoursSinceWatering := time.Since(*p.LastWatered).Hours()

	// Healthy: less than 50% of timeout
	if hoursSinceWatering < float64(p.TimeoutHours)*0.5 {
		return HealthStatusHealthy
	}

	// Needs water: between 50% and 100% of timeout
	if hoursSinceWatering < float64(p.TimeoutHours) {
		return HealthStatusNeedsWater
	}

	// Critical: past timeout
	return HealthStatusCritical
}

// GetTimeSinceWatering returns the duration since last watering
func (p *PlantState) GetTimeSinceWatering() *time.Duration {
	if p.LastWatered == nil {
		return nil
	}

	duration := time.Since(*p.LastWatered)
	return &duration
}

// GetHoursSinceWatering returns hours since last watering as a float
func (p *PlantState) GetHoursSinceWatering() *float64 {
	if p.LastWatered == nil {
		return nil
	}

	hours := time.Since(*p.LastWatered).Hours()
	return &hours
}

// IsOverdue returns true if the plant is past its watering timeout
func (p *PlantState) IsOverdue() bool {
	if p.LastWatered == nil {
		return true
	}

	return time.Since(*p.LastWatered).Hours() > float64(p.TimeoutHours)
}

// GetTimeUntilDue returns duration until watering is due (negative if overdue)
func (p *PlantState) GetTimeUntilDue() *time.Duration {
	if p.LastWatered == nil {
		return nil
	}

	nextWateringTime := p.LastWatered.Add(time.Duration(p.TimeoutHours) * time.Hour)
	timeUntilDue := time.Until(nextWateringTime)
	return &timeUntilDue
}

// GetFormattedTimeSinceWatering returns a human-readable string of time since watering
func (p *PlantState) GetFormattedTimeSinceWatering() string {
	if p.LastWatered == nil {
		return "Never watered"
	}

	duration := time.Since(*p.LastWatered)

	if duration.Hours() < 1 {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	hours := int(duration.Hours())
	if hours == 1 {
		return "1 hour ago"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// Validate checks if the plant state is valid
func (p *PlantState) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("plant name cannot be empty")
	}

	if p.TimeoutHours <= 0 {
		return fmt.Errorf("timeout hours must be positive")
	}

	if p.TimeoutHours > 8760 { // More than a year
		return fmt.Errorf("timeout hours cannot exceed 8760 (1 year)")
	}

	return nil
}

// User represents a user in the system
type User struct {
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

// AdminConfig represents system configuration
type AdminConfig struct {
	TimeoutHours  int       `json:"timeout_hours"`
	AllowedEmails []string  `json:"allowed_emails"`
	AdminEmails   []string  `json:"admin_emails"`
	LastModified  time.Time `json:"last_modified"`
	ModifiedBy    string    `json:"modified_by"`
}

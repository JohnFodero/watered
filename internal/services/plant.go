package services

import (
	"fmt"
	"log"
	"time"

	"watered/internal/models"
	"watered/internal/storage"
)

// PlantService handles plant-related business logic
type PlantService struct {
	storage storage.Storage
}

// NewPlantService creates a new plant service
func NewPlantService(storage storage.Storage) *PlantService {
	return &PlantService{
		storage: storage,
	}
}

// GetPlant returns the current plant state, creating a default one if none exists
func (s *PlantService) GetPlant() (*models.PlantState, error) {
	plant, err := s.storage.GetPlantState()
	if err != nil {
		return nil, fmt.Errorf("failed to get plant state: %w", err)
	}

	// Create default plant if none exists
	if plant == nil {
		plant = s.createDefaultPlant()
		if err := s.storage.UpdatePlantState(plant); err != nil {
			log.Printf("Warning: failed to save default plant: %v", err)
		}
	}

	return plant, nil
}

// WaterPlant records a watering event for the plant
func (s *PlantService) WaterPlant(wateredBy string) (*models.PlantState, error) {
	if wateredBy == "" {
		return nil, fmt.Errorf("watered_by field is required")
	}

	plant, err := s.GetPlant()
	if err != nil {
		return nil, fmt.Errorf("failed to get plant for watering: %w", err)
	}

	// Update watering information
	now := time.Now()
	plant.LastWatered = &now
	plant.WateredBy = wateredBy
	plant.UpdatedAt = now

	// Save the updated plant state
	if err := s.storage.UpdatePlantState(plant); err != nil {
		return nil, fmt.Errorf("failed to save watered plant: %w", err)
	}

	log.Printf("Plant watered by %s at %s", wateredBy, now.Format(time.RFC3339))
	return plant, nil
}

// GetPlantStatus returns just the health status information
func (s *PlantService) GetPlantStatus() (*PlantStatusResponse, error) {
	plant, err := s.GetPlant()
	if err != nil {
		return nil, err
	}

	return &PlantStatusResponse{
		Status:                     plant.GetHealthStatus(),
		TimeSinceWateringFormatted: plant.GetFormattedTimeSinceWatering(),
		HoursSinceWatering:         plant.GetHoursSinceWatering(),
		IsOverdue:                  plant.IsOverdue(),
		TimeUntilDue:               plant.GetTimeUntilDue(),
	}, nil
}

// GetPlantTimer returns timer-specific information
func (s *PlantService) GetPlantTimer() (*PlantTimerResponse, error) {
	plant, err := s.GetPlant()
	if err != nil {
		return nil, err
	}

	var nextWateringTime *time.Time
	if plant.LastWatered != nil {
		next := plant.LastWatered.Add(time.Duration(plant.TimeoutHours) * time.Hour)
		nextWateringTime = &next
	}

	return &PlantTimerResponse{
		LastWatered:                plant.LastWatered,
		TimeSinceWatering:          plant.GetTimeSinceWatering(),
		TimeSinceWateringFormatted: plant.GetFormattedTimeSinceWatering(),
		HoursSinceWatering:         plant.GetHoursSinceWatering(),
		TimeoutHours:               plant.TimeoutHours,
		NextWateringTime:           nextWateringTime,
		TimeUntilDue:               plant.GetTimeUntilDue(),
		IsOverdue:                  plant.IsOverdue(),
	}, nil
}

// UpdatePlantSettings updates plant configuration (timeout, name, etc.)
func (s *PlantService) UpdatePlantSettings(name string, timeoutHours int) (*models.PlantState, error) {
	plant, err := s.GetPlant()
	if err != nil {
		return nil, err
	}

	// Update settings
	if name != "" {
		plant.Name = name
	}

	if timeoutHours != 0 {
		if timeoutHours < 0 {
			return nil, fmt.Errorf("timeout hours cannot be negative")
		}
		plant.TimeoutHours = timeoutHours
	}

	plant.UpdatedAt = time.Now()

	// Validate the updated plant
	if err := plant.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plant settings: %w", err)
	}

	// Save the updated plant
	if err := s.storage.UpdatePlantState(plant); err != nil {
		return nil, fmt.Errorf("failed to save plant settings: %w", err)
	}

	log.Printf("Plant settings updated: name=%s, timeout=%d hours", plant.Name, plant.TimeoutHours)
	return plant, nil
}

// ResetPlant resets the plant to unwatered state (admin function)
func (s *PlantService) ResetPlant() (*models.PlantState, error) {
	plant, err := s.GetPlant()
	if err != nil {
		return nil, err
	}

	// Reset watering state
	plant.LastWatered = nil
	plant.WateredBy = ""
	plant.UpdatedAt = time.Now()

	if err := s.storage.UpdatePlantState(plant); err != nil {
		return nil, fmt.Errorf("failed to reset plant: %w", err)
	}

	log.Printf("Plant reset to unwatered state")
	return plant, nil
}

// createDefaultPlant creates a default plant configuration
func (s *PlantService) createDefaultPlant() *models.PlantState {
	now := time.Now()
	return &models.PlantState{
		ID:           1,
		Name:         "Our Plant",
		LastWatered:  nil, // Never watered initially
		TimeoutHours: 24,  // Default to 24 hours
		WateredBy:    "",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// PlantStatusResponse represents the response for plant status endpoint
type PlantStatusResponse struct {
	Status                     models.PlantHealthStatus `json:"status"`
	TimeSinceWateringFormatted string                   `json:"time_since_watering_formatted"`
	HoursSinceWatering         *float64                 `json:"hours_since_watering"`
	IsOverdue                  bool                     `json:"is_overdue"`
	TimeUntilDue               *time.Duration           `json:"time_until_due"`
}

// PlantTimerResponse represents the response for plant timer endpoint
type PlantTimerResponse struct {
	LastWatered                *time.Time     `json:"last_watered"`
	TimeSinceWatering          *time.Duration `json:"time_since_watering"`
	TimeSinceWateringFormatted string         `json:"time_since_watering_formatted"`
	HoursSinceWatering         *float64       `json:"hours_since_watering"`
	TimeoutHours               int            `json:"timeout_hours"`
	NextWateringTime           *time.Time     `json:"next_watering_time"`
	TimeUntilDue               *time.Duration `json:"time_until_due"`
	IsOverdue                  bool           `json:"is_overdue"`
}

package services

import (
	"testing"
	"time"

	"watered/internal/models"
	"watered/internal/storage"
)

func TestPlantService_GetPlant(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// First call should create default plant
	plant, err := service.GetPlant()
	if err != nil {
		t.Fatalf("Failed to get plant: %v", err)
	}

	if plant == nil {
		t.Fatal("Expected plant, got nil")
	}

	if plant.Name != "Our Plant" {
		t.Errorf("Expected default name 'Our Plant', got '%s'", plant.Name)
	}

	if plant.TimeoutHours != 24 {
		t.Errorf("Expected default timeout 24 hours, got %d", plant.TimeoutHours)
	}

	if plant.LastWatered != nil {
		t.Error("Expected nil last watered for new plant")
	}

	// Second call should return same plant
	plant2, err := service.GetPlant()
	if err != nil {
		t.Fatalf("Failed to get plant on second call: %v", err)
	}

	if plant2.ID != plant.ID {
		t.Error("Expected same plant ID on subsequent calls")
	}
}

func TestPlantService_WaterPlant(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// Test watering with valid user
	userEmail := "test@example.com"
	plant, err := service.WaterPlant(userEmail)
	if err != nil {
		t.Fatalf("Failed to water plant: %v", err)
	}

	if plant.LastWatered == nil {
		t.Error("Expected last watered to be set")
	}

	if plant.WateredBy != userEmail {
		t.Errorf("Expected watered by '%s', got '%s'", userEmail, plant.WateredBy)
	}

	// Check that watering time is recent (within last minute)
	timeSince := time.Since(*plant.LastWatered)
	if timeSince > time.Minute {
		t.Errorf("Expected recent watering time, got %v ago", timeSince)
	}

	// Test watering without user email
	_, err = service.WaterPlant("")
	if err == nil {
		t.Error("Expected error when watering without user email")
	}
}

func TestPlantService_GetPlantStatus(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// Get status for new plant (should be critical)
	status, err := service.GetPlantStatus()
	if err != nil {
		t.Fatalf("Failed to get plant status: %v", err)
	}

	if status.Status != models.HealthStatusCritical {
		t.Errorf("Expected critical status for new plant, got %s", status.Status)
	}

	if !status.IsOverdue {
		t.Error("Expected plant to be overdue when never watered")
	}

	// Water the plant and check status again
	_, err = service.WaterPlant("test@example.com")
	if err != nil {
		t.Fatalf("Failed to water plant: %v", err)
	}

	status, err = service.GetPlantStatus()
	if err != nil {
		t.Fatalf("Failed to get plant status after watering: %v", err)
	}

	if status.Status != models.HealthStatusHealthy {
		t.Errorf("Expected healthy status after watering, got %s", status.Status)
	}

	if status.IsOverdue {
		t.Error("Expected plant to not be overdue after watering")
	}
}

func TestPlantService_GetPlantTimer(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// Get timer for new plant
	timer, err := service.GetPlantTimer()
	if err != nil {
		t.Fatalf("Failed to get plant timer: %v", err)
	}

	if timer.LastWatered != nil {
		t.Error("Expected nil last watered for new plant")
	}

	if timer.TimeSinceWatering != nil {
		t.Error("Expected nil time since watering for new plant")
	}

	if timer.NextWateringTime != nil {
		t.Error("Expected nil next watering time for new plant")
	}

	if !timer.IsOverdue {
		t.Error("Expected plant to be overdue when never watered")
	}

	// Water the plant and check timer again
	_, err = service.WaterPlant("test@example.com")
	if err != nil {
		t.Fatalf("Failed to water plant: %v", err)
	}

	timer, err = service.GetPlantTimer()
	if err != nil {
		t.Fatalf("Failed to get plant timer after watering: %v", err)
	}

	if timer.LastWatered == nil {
		t.Error("Expected last watered to be set after watering")
	}

	if timer.TimeSinceWatering == nil {
		t.Error("Expected time since watering to be set after watering")
	}

	if timer.NextWateringTime == nil {
		t.Error("Expected next watering time to be set after watering")
	}

	if timer.IsOverdue {
		t.Error("Expected plant to not be overdue immediately after watering")
	}

	if timer.TimeoutHours != 24 {
		t.Errorf("Expected timeout 24 hours, got %d", timer.TimeoutHours)
	}
}

func TestPlantService_UpdatePlantSettings(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// Update plant name
	plant, err := service.UpdatePlantSettings("My Special Plant", 0)
	if err != nil {
		t.Fatalf("Failed to update plant name: %v", err)
	}

	if plant.Name != "My Special Plant" {
		t.Errorf("Expected name 'My Special Plant', got '%s'", plant.Name)
	}

	// Timeout should remain unchanged when 0 is passed
	if plant.TimeoutHours != 24 {
		t.Errorf("Expected timeout to remain 24, got %d", plant.TimeoutHours)
	}

	// Update timeout
	plant, err = service.UpdatePlantSettings("", 48)
	if err != nil {
		t.Fatalf("Failed to update plant timeout: %v", err)
	}

	if plant.TimeoutHours != 48 {
		t.Errorf("Expected timeout 48 hours, got %d", plant.TimeoutHours)
	}

	// Name should remain unchanged when empty string is passed
	if plant.Name != "My Special Plant" {
		t.Errorf("Expected name to remain 'My Special Plant', got '%s'", plant.Name)
	}

	// Test invalid timeout
	_, err = service.UpdatePlantSettings("", -1)
	if err == nil {
		t.Error("Expected error for negative timeout")
	}
}

func TestPlantService_ResetPlant(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	service := NewPlantService(store)

	// Water the plant first
	_, err := service.WaterPlant("test@example.com")
	if err != nil {
		t.Fatalf("Failed to water plant: %v", err)
	}

	// Verify plant was watered
	plant, err := service.GetPlant()
	if err != nil {
		t.Fatalf("Failed to get plant: %v", err)
	}

	if plant.LastWatered == nil {
		t.Fatal("Expected plant to be watered")
	}

	if plant.WateredBy == "" {
		t.Fatal("Expected watered by to be set")
	}

	// Reset the plant
	resetPlant, err := service.ResetPlant()
	if err != nil {
		t.Fatalf("Failed to reset plant: %v", err)
	}

	if resetPlant.LastWatered != nil {
		t.Error("Expected last watered to be nil after reset")
	}

	if resetPlant.WateredBy != "" {
		t.Error("Expected watered by to be empty after reset")
	}

	// Verify plant is critical after reset
	if resetPlant.GetHealthStatus() != models.HealthStatusCritical {
		t.Errorf("Expected critical status after reset, got %s", resetPlant.GetHealthStatus())
	}
}

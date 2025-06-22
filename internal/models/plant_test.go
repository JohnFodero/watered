package models

import (
	"testing"
	"time"
)

func TestPlantState_GetHealthStatus(t *testing.T) {
	tests := []struct {
		name         string
		lastWatered  *time.Time
		timeoutHours int
		expected     PlantHealthStatus
	}{
		{
			name:         "Never watered",
			lastWatered:  nil,
			timeoutHours: 24,
			expected:     HealthStatusCritical,
		},
		{
			name:         "Recently watered (healthy)",
			lastWatered:  timePtr(time.Now().Add(-1 * time.Hour)),
			timeoutHours: 24,
			expected:     HealthStatusHealthy,
		},
		{
			name:         "Half timeout (needs water)",
			lastWatered:  timePtr(time.Now().Add(-13 * time.Hour)),
			timeoutHours: 24,
			expected:     HealthStatusNeedsWater,
		},
		{
			name:         "Past timeout (critical)",
			lastWatered:  timePtr(time.Now().Add(-25 * time.Hour)),
			timeoutHours: 24,
			expected:     HealthStatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plant := &PlantState{
				LastWatered:  tt.lastWatered,
				TimeoutHours: tt.timeoutHours,
			}

			result := plant.GetHealthStatus()
			if result != tt.expected {
				t.Errorf("Expected health status %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPlantState_IsOverdue(t *testing.T) {
	tests := []struct {
		name         string
		lastWatered  *time.Time
		timeoutHours int
		expected     bool
	}{
		{
			name:         "Never watered",
			lastWatered:  nil,
			timeoutHours: 24,
			expected:     true,
		},
		{
			name:         "Recently watered",
			lastWatered:  timePtr(time.Now().Add(-1 * time.Hour)),
			timeoutHours: 24,
			expected:     false,
		},
		{
			name:         "Exactly at timeout",
			lastWatered:  timePtr(time.Now().Add(-24 * time.Hour)),
			timeoutHours: 24,
			expected:     true, // At exactly timeout, it should be overdue
		},
		{
			name:         "Past timeout",
			lastWatered:  timePtr(time.Now().Add(-25 * time.Hour)),
			timeoutHours: 24,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plant := &PlantState{
				LastWatered:  tt.lastWatered,
				TimeoutHours: tt.timeoutHours,
			}

			result := plant.IsOverdue()
			if result != tt.expected {
				t.Errorf("Expected overdue %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPlantState_GetFormattedTimeSinceWatering(t *testing.T) {
	tests := []struct {
		name        string
		lastWatered *time.Time
		expected    string
	}{
		{
			name:        "Never watered",
			lastWatered: nil,
			expected:    "Never watered",
		},
		{
			name:        "1 minute ago",
			lastWatered: timePtr(time.Now().Add(-1 * time.Minute)),
			expected:    "1 minute ago",
		},
		{
			name:        "30 minutes ago",
			lastWatered: timePtr(time.Now().Add(-30 * time.Minute)),
			expected:    "30 minutes ago",
		},
		{
			name:        "1 hour ago",
			lastWatered: timePtr(time.Now().Add(-1 * time.Hour)),
			expected:    "1 hour ago",
		},
		{
			name:        "5 hours ago",
			lastWatered: timePtr(time.Now().Add(-5 * time.Hour)),
			expected:    "5 hours ago",
		},
		{
			name:        "1 day ago",
			lastWatered: timePtr(time.Now().Add(-24 * time.Hour)),
			expected:    "1 day ago",
		},
		{
			name:        "3 days ago",
			lastWatered: timePtr(time.Now().Add(-72 * time.Hour)),
			expected:    "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plant := &PlantState{
				LastWatered: tt.lastWatered,
			}

			result := plant.GetFormattedTimeSinceWatering()
			if result != tt.expected {
				t.Errorf("Expected formatted time %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPlantState_Validate(t *testing.T) {
	tests := []struct {
		name      string
		plant     PlantState
		expectErr bool
	}{
		{
			name: "Valid plant",
			plant: PlantState{
				Name:         "Test Plant",
				TimeoutHours: 24,
			},
			expectErr: false,
		},
		{
			name: "Empty name",
			plant: PlantState{
				Name:         "",
				TimeoutHours: 24,
			},
			expectErr: true,
		},
		{
			name: "Zero timeout",
			plant: PlantState{
				Name:         "Test Plant",
				TimeoutHours: 0,
			},
			expectErr: true,
		},
		{
			name: "Negative timeout",
			plant: PlantState{
				Name:         "Test Plant",
				TimeoutHours: -1,
			},
			expectErr: true,
		},
		{
			name: "Timeout too large",
			plant: PlantState{
				Name:         "Test Plant",
				TimeoutHours: 10000,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plant.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error, got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestPlantState_GetHoursSinceWatering(t *testing.T) {
	plant := &PlantState{
		LastWatered: timePtr(time.Now().Add(-2 * time.Hour)),
	}

	hours := plant.GetHoursSinceWatering()
	if hours == nil {
		t.Error("Expected hours since watering, got nil")
	}

	if *hours < 1.9 || *hours > 2.1 {
		t.Errorf("Expected approximately 2 hours, got %f", *hours)
	}

	// Test nil case
	plant.LastWatered = nil
	hours = plant.GetHoursSinceWatering()
	if hours != nil {
		t.Error("Expected nil for hours since watering when never watered")
	}
}

func TestPlantState_GetTimeSinceWatering(t *testing.T) {
	plant := &PlantState{
		LastWatered: timePtr(time.Now().Add(-30 * time.Minute)),
	}

	duration := plant.GetTimeSinceWatering()
	if duration == nil {
		t.Error("Expected duration since watering, got nil")
	}

	if duration.Minutes() < 29 || duration.Minutes() > 31 {
		t.Errorf("Expected approximately 30 minutes, got %f", duration.Minutes())
	}

	// Test nil case
	plant.LastWatered = nil
	duration = plant.GetTimeSinceWatering()
	if duration != nil {
		t.Error("Expected nil for duration since watering when never watered")
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

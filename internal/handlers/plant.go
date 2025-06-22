package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"watered/internal/auth"
	"watered/internal/services"
)

// PlantHandlers contains all plant-related HTTP handlers
type PlantHandlers struct {
	plantService *services.PlantService
	authService  *auth.AuthService
}

// NewPlantHandlers creates a new plant handlers instance
func NewPlantHandlers(plantService *services.PlantService, authService *auth.AuthService) *PlantHandlers {
	return &PlantHandlers{
		plantService: plantService,
		authService:  authService,
	}
}

// GetPlantHandler returns the current plant state
// GET /api/plant
func (h *PlantHandlers) GetPlantHandler(w http.ResponseWriter, r *http.Request) {
	plant, err := h.plantService.GetPlant()
	if err != nil {
		log.Printf("Failed to get plant: %v", err)
		http.Error(w, "Failed to get plant state", http.StatusInternalServerError)
		return
	}

	// Create response with computed status
	response := map[string]interface{}{
		"id":                   plant.ID,
		"name":                 plant.Name,
		"last_watered":         plant.LastWatered,
		"timeout_hours":        plant.TimeoutHours,
		"watered_by":           plant.WateredBy,
		"created_at":           plant.CreatedAt,
		"updated_at":           plant.UpdatedAt,
		"health_status":        plant.GetHealthStatus(),
		"time_since_watering":  plant.GetFormattedTimeSinceWatering(),
		"hours_since_watering": plant.GetHoursSinceWatering(),
		"is_overdue":           plant.IsOverdue(),
		"time_until_due":       plant.GetTimeUntilDue(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WaterPlantHandler records a plant watering event
// POST /api/plant/water
func (h *PlantHandlers) WaterPlantHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current authenticated user
	user, err := h.authService.GetCurrentUser(r)
	if err != nil || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Water the plant
	plant, err := h.plantService.WaterPlant(user.Email)
	if err != nil {
		log.Printf("Failed to water plant: %v", err)
		http.Error(w, "Failed to water plant", http.StatusInternalServerError)
		return
	}

	// Return updated plant state
	response := map[string]interface{}{
		"success": true,
		"message": "Plant watered successfully! 🌱",
		"plant": map[string]interface{}{
			"id":                   plant.ID,
			"name":                 plant.Name,
			"last_watered":         plant.LastWatered,
			"timeout_hours":        plant.TimeoutHours,
			"watered_by":           plant.WateredBy,
			"updated_at":           plant.UpdatedAt,
			"health_status":        plant.GetHealthStatus(),
			"time_since_watering":  plant.GetFormattedTimeSinceWatering(),
			"hours_since_watering": plant.GetHoursSinceWatering(),
			"is_overdue":           plant.IsOverdue(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPlantStatusHandler returns just the plant health status
// GET /api/plant/status
func (h *PlantHandlers) GetPlantStatusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := h.plantService.GetPlantStatus()
	if err != nil {
		log.Printf("Failed to get plant status: %v", err)
		http.Error(w, "Failed to get plant status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetPlantTimerHandler returns plant timer information
// GET /api/plant/timer
func (h *PlantHandlers) GetPlantTimerHandler(w http.ResponseWriter, r *http.Request) {
	timer, err := h.plantService.GetPlantTimer()
	if err != nil {
		log.Printf("Failed to get plant timer: %v", err)
		http.Error(w, "Failed to get plant timer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timer)
}

// UpdatePlantSettingsHandler updates plant configuration (admin only)
// PUT /api/plant/settings
func (h *PlantHandlers) UpdatePlantSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Name         string `json:"name"`
		TimeoutHours int    `json:"timeout_hours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update plant settings
	plant, err := h.plantService.UpdatePlantSettings(req.Name, req.TimeoutHours)
	if err != nil {
		log.Printf("Failed to update plant settings: %v", err)
		http.Error(w, "Failed to update plant settings: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Plant settings updated successfully",
		"plant": map[string]interface{}{
			"id":                  plant.ID,
			"name":                plant.Name,
			"last_watered":        plant.LastWatered,
			"timeout_hours":       plant.TimeoutHours,
			"watered_by":          plant.WateredBy,
			"updated_at":          plant.UpdatedAt,
			"health_status":       plant.GetHealthStatus(),
			"time_since_watering": plant.GetFormattedTimeSinceWatering(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPlantHandler resets the plant to unwatered state (admin only)
// POST /api/plant/reset
func (h *PlantHandlers) ResetPlantHandler(w http.ResponseWriter, r *http.Request) {
	plant, err := h.plantService.ResetPlant()
	if err != nil {
		log.Printf("Failed to reset plant: %v", err)
		http.Error(w, "Failed to reset plant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Plant reset to unwatered state",
		"plant": map[string]interface{}{
			"id":                  plant.ID,
			"name":                plant.Name,
			"last_watered":        plant.LastWatered,
			"timeout_hours":       plant.TimeoutHours,
			"watered_by":          plant.WateredBy,
			"updated_at":          plant.UpdatedAt,
			"health_status":       plant.GetHealthStatus(),
			"time_since_watering": plant.GetFormattedTimeSinceWatering(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

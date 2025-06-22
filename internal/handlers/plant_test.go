package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"watered/internal/auth"
	"watered/internal/services"
	"watered/internal/storage"
)

func TestPlantHandlers_GetPlantHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	req := httptest.NewRequest("GET", "/api/plant", nil)
	w := httptest.NewRecorder()

	handlers.GetPlantHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["name"] != "Our Plant" {
		t.Errorf("Expected default plant name 'Our Plant', got %v", response["name"])
	}

	if response["health_status"] != "critical" {
		t.Errorf("Expected health status 'critical' for new plant, got %v", response["health_status"])
	}

	if response["is_overdue"] != true {
		t.Errorf("Expected is_overdue true for new plant, got %v", response["is_overdue"])
	}
}

func TestPlantHandlers_GetPlantStatusHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	req := httptest.NewRequest("GET", "/api/plant/status", nil)
	w := httptest.NewRecorder()

	handlers.GetPlantStatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "critical" {
		t.Errorf("Expected status 'critical', got %v", response["status"])
	}

	if response["is_overdue"] != true {
		t.Errorf("Expected is_overdue true, got %v", response["is_overdue"])
	}
}

func TestPlantHandlers_GetPlantTimerHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	req := httptest.NewRequest("GET", "/api/plant/timer", nil)
	w := httptest.NewRecorder()

	handlers.GetPlantTimerHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["last_watered"] != nil {
		t.Errorf("Expected nil last_watered for new plant, got %v", response["last_watered"])
	}

	if response["timeout_hours"] != 24.0 {
		t.Errorf("Expected timeout_hours 24, got %v", response["timeout_hours"])
	}

	if response["is_overdue"] != true {
		t.Errorf("Expected is_overdue true, got %v", response["is_overdue"])
	}
}

func TestPlantHandlers_WaterPlantHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	// Test without authentication
	req := httptest.NewRequest("POST", "/api/plant/water", nil)
	w := httptest.NewRecorder()

	handlers.WaterPlantHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthenticated request, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with authentication
	req = httptest.NewRequest("POST", "/api/plant/water", nil)
	w = httptest.NewRecorder()

	// Create demo session
	userInfo := &auth.GoogleUserInfo{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	authService.SetAllowedEmails(map[string]bool{"test@example.com": true})
	err := authService.CreateSession(w, req, userInfo)
	if err != nil {
		t.Fatalf("Failed to create demo session: %v", err)
	}

	// Create new request with session cookie
	cookies := w.Result().Cookies()
	req2 := httptest.NewRequest("POST", "/api/plant/water", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	handlers.WaterPlantHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d for authenticated request, got %d", http.StatusOK, w2.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}

	plant, ok := response["plant"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected plant object in response")
	}

	if plant["watered_by"] != "test@example.com" {
		t.Errorf("Expected watered_by 'test@example.com', got %v", plant["watered_by"])
	}

	if plant["health_status"] != "healthy" {
		t.Errorf("Expected health_status 'healthy' after watering, got %v", plant["health_status"])
	}
}

func TestPlantHandlers_UpdatePlantSettingsHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	reqBody := map[string]interface{}{
		"name":          "Updated Plant",
		"timeout_hours": 48,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/plant/settings", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.UpdatePlantSettingsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}

	plant, ok := response["plant"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected plant object in response")
	}

	if plant["name"] != "Updated Plant" {
		t.Errorf("Expected name 'Updated Plant', got %v", plant["name"])
	}

	if plant["timeout_hours"] != 48.0 {
		t.Errorf("Expected timeout_hours 48, got %v", plant["timeout_hours"])
	}
}

func TestPlantHandlers_ResetPlantHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()

	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	handlers := NewPlantHandlers(plantService, authService)

	// First water the plant
	_, err := plantService.WaterPlant("test@example.com")
	if err != nil {
		t.Fatalf("Failed to water plant: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/plant/reset", nil)
	w := httptest.NewRecorder()

	handlers.ResetPlantHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}

	plant, ok := response["plant"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected plant object in response")
	}

	if plant["last_watered"] != nil {
		t.Errorf("Expected last_watered nil after reset, got %v", plant["last_watered"])
	}

	if plant["watered_by"] != "" {
		t.Errorf("Expected watered_by empty after reset, got %v", plant["watered_by"])
	}

	if plant["health_status"] != "critical" {
		t.Errorf("Expected health_status 'critical' after reset, got %v", plant["health_status"])
	}
}

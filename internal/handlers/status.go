package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// StatusResponse represents the API status response
type StatusResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// GetStatus returns the current API status
func GetStatus(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		Status:    "ok",
		Service:   "watered-api",
		Version:   "1.0.0",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
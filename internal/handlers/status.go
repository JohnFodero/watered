package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var serverStartTime = time.Now()

// StatusResponse represents the API status response
type StatusResponse struct {
	Status         string    `json:"status"`
	Service        string    `json:"service"`
	Version        string    `json:"version"`
	Timestamp      time.Time `json:"timestamp"`
	UptimeSeconds  float64   `json:"uptime_seconds"`
	UptimeFormatted string   `json:"uptime_formatted"`
}

// formatUptime formats uptime duration into human-readable string
func formatUptime(duration time.Duration) string {
	totalMinutes := int(duration.Minutes())
	totalHours := int(duration.Hours())
	totalDays := int(duration.Hours() / 24)

	if totalDays >= 1 {
		if totalDays == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", totalDays)
	} else if totalHours >= 1 {
		if totalHours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", totalHours)
	} else {
		if totalMinutes <= 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", totalMinutes)
	}
}

// GetStatus returns the current API status
func GetStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	uptime := now.Sub(serverStartTime)
	
	response := StatusResponse{
		Status:          "ok",
		Service:         "watered-api",
		Version:         "1.0.0",
		Timestamp:       now,
		UptimeSeconds:   uptime.Seconds(),
		UptimeFormatted: formatUptime(uptime),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

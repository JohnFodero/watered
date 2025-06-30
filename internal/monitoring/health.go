package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"watered/internal/storage"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     time.Duration              `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
	System     SystemMetrics              `json:"system"`
	Details    map[string]interface{}     `json:"details,omitempty"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	MemoryUsage  MemoryMetrics `json:"memory"`
	GoRoutines   int           `json:"goroutines"`
	CGOCalls     int64         `json:"cgo_calls"`
	GCStats      GCMetrics     `json:"gc_stats"`
	OpenFileDesc int           `json:"open_file_descriptors,omitempty"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Alloc       uint64  `json:"alloc_bytes"`
	TotalAlloc  uint64  `json:"total_alloc_bytes"`
	Sys         uint64  `json:"sys_bytes"`
	NumGC       uint32  `json:"num_gc"`
	HeapAlloc   uint64  `json:"heap_alloc_bytes"`
	HeapInuse   uint64  `json:"heap_inuse_bytes"`
	StackInuse  uint64  `json:"stack_inuse_bytes"`
	MemoryUsage float64 `json:"memory_usage_percent"`
}

// GCMetrics represents garbage collection metrics
type GCMetrics struct {
	NumGC      uint32        `json:"num_gc"`
	PauseTotal time.Duration `json:"pause_total_ns"`
	LastPause  time.Duration `json:"last_pause_ns"`
	AverageGC  time.Duration `json:"average_gc_pause_ns"`
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	Check(ctx context.Context) ComponentHealth
	Name() string
}

// HealthMonitor manages health checks for the application
type HealthMonitor struct {
	checkers  map[string]HealthChecker
	startTime time.Time
	version   string
	mu        sync.RWMutex
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(version string) *HealthMonitor {
	return &HealthMonitor{
		checkers:  make(map[string]HealthChecker),
		startTime: time.Now(),
		version:   version,
	}
}

// RegisterChecker registers a health checker
func (hm *HealthMonitor) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[checker.Name()] = checker
}

// CheckHealth performs all health checks and returns a comprehensive report
func (hm *HealthMonitor) CheckHealth(ctx context.Context) *HealthReport {
	hm.mu.RLock()
	checkers := make(map[string]HealthChecker, len(hm.checkers))
	for k, v := range hm.checkers {
		checkers[k] = v
	}
	hm.mu.RUnlock()

	report := &HealthReport{
		Timestamp:  time.Now(),
		Version:    hm.version,
		Uptime:     time.Since(hm.startTime),
		Components: make(map[string]ComponentHealth),
		System:     hm.getSystemMetrics(),
	}

	// Check all components in parallel
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	overallStatus := HealthStatusHealthy

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			health := c.Check(checkCtx)

			mu.Lock()
			report.Components[n] = health

			// Determine overall status
			switch health.Status {
			case HealthStatusUnhealthy:
				overallStatus = HealthStatusUnhealthy
			case HealthStatusDegraded:
				if overallStatus == HealthStatusHealthy {
					overallStatus = HealthStatusDegraded
				}
			}
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	report.Status = overallStatus

	return report
}

// getSystemMetrics collects system-level metrics
func (hm *HealthMonitor) getSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	gcMetrics := GCMetrics{
		NumGC:      memStats.NumGC,
		PauseTotal: time.Duration(memStats.PauseTotalNs),
	}

	if memStats.NumGC > 0 {
		gcMetrics.LastPause = time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256])
		gcMetrics.AverageGC = time.Duration(memStats.PauseTotalNs / uint64(memStats.NumGC))
	}

	memoryUsagePercent := float64(memStats.Alloc) / float64(memStats.Sys) * 100

	return SystemMetrics{
		MemoryUsage: MemoryMetrics{
			Alloc:       memStats.Alloc,
			TotalAlloc:  memStats.TotalAlloc,
			Sys:         memStats.Sys,
			NumGC:       memStats.NumGC,
			HeapAlloc:   memStats.HeapAlloc,
			HeapInuse:   memStats.HeapInuse,
			StackInuse:  memStats.StackInuse,
			MemoryUsage: memoryUsagePercent,
		},
		GoRoutines: runtime.NumGoroutine(),
		CGOCalls:   runtime.NumCgoCall(),
		GCStats:    gcMetrics,
	}
}

// HTTPHandler returns an HTTP handler for health checks
func (hm *HealthMonitor) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		report := hm.CheckHealth(ctx)

		// Set appropriate HTTP status code
		switch report.Status {
		case HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case HealthStatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK, but with warnings
		case HealthStatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		if err := json.NewEncoder(w).Encode(report); err != nil {
			http.Error(w, "Failed to encode health report", http.StatusInternalServerError)
		}
	}
}

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	storage storage.Storage
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(storage storage.Storage) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{storage: storage}
}

// Name returns the name of this health checker
func (d *DatabaseHealthChecker) Name() string {
	return "database"
}

// Check performs the database health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        d.Name(),
		LastChecked: start,
	}

	// Test basic database operations
	if _, err := d.storage.GetPlantState(); err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Database query failed: %v", err)
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "Database connectivity verified"
	}

	health.Duration = time.Since(start)
	return health
}

// MemoryHealthChecker checks memory usage
type MemoryHealthChecker struct {
	maxMemoryMB float64
}

// NewMemoryHealthChecker creates a new memory health checker
func NewMemoryHealthChecker(maxMemoryMB float64) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		maxMemoryMB: maxMemoryMB,
	}
}

// Name returns the name of this health checker
func (m *MemoryHealthChecker) Name() string {
	return "memory"
}

// Check performs the memory health check
func (m *MemoryHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        m.Name(),
		LastChecked: start,
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMemoryMB := float64(memStats.Alloc) / 1024 / 1024
	usagePercent := (currentMemoryMB / m.maxMemoryMB) * 100

	health.Details = map[string]interface{}{
		"current_memory_mb": currentMemoryMB,
		"max_memory_mb":     m.maxMemoryMB,
		"usage_percent":     usagePercent,
		"heap_objects":      memStats.HeapObjects,
		"gc_cycles":         memStats.NumGC,
	}

	if usagePercent > 90 {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Memory usage critically high: %.1f%%", usagePercent)
	} else if usagePercent > 75 {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Memory usage high: %.1f%%", usagePercent)
	} else {
		health.Status = HealthStatusHealthy
		health.Message = fmt.Sprintf("Memory usage normal: %.1f%%", usagePercent)
	}

	health.Duration = time.Since(start)
	return health
}

// ApplicationHealthChecker checks application-specific health
type ApplicationHealthChecker struct {
	storage storage.Storage
}

// NewApplicationHealthChecker creates a new application health checker
func NewApplicationHealthChecker(storage storage.Storage) *ApplicationHealthChecker {
	return &ApplicationHealthChecker{storage: storage}
}

// Name returns the name of this health checker
func (a *ApplicationHealthChecker) Name() string {
	return "application"
}

// Check performs application-specific health checks
func (a *ApplicationHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:        a.Name(),
		LastChecked: start,
	}

	// Check if we can get admin config
	if _, err := a.storage.GetAdminConfig(); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = "Admin configuration not accessible"
		health.Details = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "Application components functional"
		health.Details = map[string]interface{}{
			"demo_mode": true, // Could check actual mode
		}
	}

	health.Duration = time.Since(start)
	return health
}

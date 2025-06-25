package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"watered/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestHealthMonitor(t *testing.T) {
	monitor := NewHealthMonitor("test-1.0.0")
	
	// Test initial state
	assert.NotNil(t, monitor)
	assert.Equal(t, "test-1.0.0", monitor.version)
	
	// Test empty health check
	report := monitor.CheckHealth(context.Background())
	assert.Equal(t, HealthStatusHealthy, report.Status)
	assert.Equal(t, "test-1.0.0", report.Version)
	assert.True(t, report.Uptime > 0)
}

func TestDatabaseHealthChecker(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	checker := NewDatabaseHealthChecker(store)
	assert.Equal(t, "database", checker.Name())
	
	health := checker.Check(context.Background())
	assert.Equal(t, "database", health.Name)
	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.Contains(t, health.Message, "connectivity verified")
	assert.True(t, health.Duration > 0)
}

func TestMemoryHealthChecker(t *testing.T) {
	checker := NewMemoryHealthChecker(512.0) // 512MB limit
	assert.Equal(t, "memory", checker.Name())
	
	health := checker.Check(context.Background())
	assert.Equal(t, "memory", health.Name)
	assert.Equal(t, HealthStatusHealthy, health.Status) // Should be healthy for test
	assert.Contains(t, health.Message, "Memory usage")
	assert.True(t, health.Duration > 0)
	
	// Check details
	assert.Contains(t, health.Details, "current_memory_mb")
	assert.Contains(t, health.Details, "max_memory_mb")
	assert.Contains(t, health.Details, "usage_percent")
}

func TestApplicationHealthChecker(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	checker := NewApplicationHealthChecker(store)
	assert.Equal(t, "application", checker.Name())
	
	health := checker.Check(context.Background())
	assert.Equal(t, "application", health.Name)
	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.Contains(t, health.Message, "functional")
	assert.True(t, health.Duration > 0)
}

func TestHealthMonitorWithCheckers(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	monitor := NewHealthMonitor("test-1.0.0")
	monitor.RegisterChecker(NewDatabaseHealthChecker(store))
	monitor.RegisterChecker(NewMemoryHealthChecker(512.0))
	monitor.RegisterChecker(NewApplicationHealthChecker(store))
	
	report := monitor.CheckHealth(context.Background())
	
	// Verify overall status
	assert.Equal(t, HealthStatusHealthy, report.Status)
	assert.Len(t, report.Components, 3)
	
	// Verify each component
	assert.Contains(t, report.Components, "database")
	assert.Contains(t, report.Components, "memory")
	assert.Contains(t, report.Components, "application")
	
	// Verify system metrics
	assert.True(t, report.System.GoRoutines > 0)
	assert.True(t, report.System.MemoryUsage.Alloc > 0)
}

func TestHealthHTTPHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	defer store.Close()
	
	monitor := NewHealthMonitor("test-1.0.0")
	monitor.RegisterChecker(NewDatabaseHealthChecker(store))
	
	handler := monitor.HTTPHandler()
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	
	// Verify response contains expected fields
	body := w.Body.String()
	assert.Contains(t, body, "\"status\":")
	assert.Contains(t, body, "\"version\":")
	assert.Contains(t, body, "\"components\":")
	assert.Contains(t, body, "\"system\":")
}

func TestHealthCheckTimeout(t *testing.T) {
	monitor := NewHealthMonitor("test-1.0.0")
	
	// Create a slow checker
	slowChecker := &slowHealthChecker{}
	monitor.RegisterChecker(slowChecker)
	
	// Health check should complete even with slow checker
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	start := time.Now()
	report := monitor.CheckHealth(ctx)
	duration := time.Since(start)
	
	// Should complete within reasonable time due to internal timeout
	assert.Less(t, duration, 10*time.Second)
	assert.Contains(t, report.Components, "slow")
}

// Mock slow health checker for testing
type slowHealthChecker struct{}

func (s *slowHealthChecker) Name() string {
	return "slow"
}

func (s *slowHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Simulate slow operation that respects context timeout
	select {
	case <-time.After(2 * time.Second):
		// Slow operation completed
	case <-ctx.Done():
		// Context cancelled/timed out
	}
	
	return ComponentHealth{
		Name:        s.Name(),
		Status:      HealthStatusHealthy,
		Message:     "Slow checker completed",
		LastChecked: start,
		Duration:    time.Since(start),
	}
}

func TestSystemMetrics(t *testing.T) {
	monitor := NewHealthMonitor("test-1.0.0")
	
	metrics := monitor.getSystemMetrics()
	
	// Verify system metrics are populated
	assert.True(t, metrics.MemoryUsage.Alloc > 0)
	assert.True(t, metrics.MemoryUsage.Sys > 0)
	assert.True(t, metrics.GoRoutines > 0)
	assert.True(t, metrics.MemoryUsage.MemoryUsage >= 0)
	assert.True(t, metrics.MemoryUsage.MemoryUsage <= 100)
}